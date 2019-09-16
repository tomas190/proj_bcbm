package internal

import (
	"github.com/shopspring/decimal"
	"math/rand"
	"proj_bcbm/src/server/constant"
	con "proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

// Mgr <--> Dealer <--> C2C
//            ^
//            |
//
//           Bots

type Dealer struct {
	*Room
	clock       *time.Ticker
	counter     uint32
	ddl         uint32
	bankerRound uint32 // 庄家做了多少轮

	Status      uint32     // 房间状态
	res         uint32     // 最新开奖结果
	bankerWin   float64    // 庄家输赢
	bankerMoney float64    // 庄家余额
	History     []uint32   // 房间开奖历史
	HRChan      chan HRMsg // 房间大厅通信

	Users          map[uint32]*User     // 房间用户-不包括机器人
	Bots           []*Bot               // 房间机器人
	Bankers        []Player             // 上庄玩家榜单
	UserBets       map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
	UserBetsDetail map[uint32][]msg.Bet // 用户具体投注
	UserAutoBet    map[uint32]bool      // 本局投注记录
	AutoBetRecord  map[uint32][]msg.Bet // 续投记录
	AreaBets       []float64            // 每个区域玩家投注总数
	AreaBotBets    []float64            // 每个区域机器人投注总数
}

func NewDealer(rID uint32, hr chan HRMsg) *Dealer {
	return &Dealer{
		Users:          make(map[uint32]*User),
		Bankers:        make([]Player, 0),
		Room:           NewRoom(rID, con.RL1MinBet, con.RL1MaxBet, con.RL1MinLimit),
		clock:          time.NewTicker(time.Second),
		HRChan:         hr,
		UserAutoBet:    map[uint32]bool{},
		UserBets:       map[uint32][]float64{},
		UserBetsDetail: map[uint32][]msg.Bet{},
		AutoBetRecord:  map[uint32][]msg.Bet{},
		AreaBets:       []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
		AreaBotBets:    []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

// 15-开始下注-0 停止下注  下注
// 23 22 21 - 跑马灯 - 随便什么时候开奖 显示 4 3 2 1 0  开奖
// 2 1 0 清空筹码
// 重置表
func (dl *Dealer) ClockReset(duration uint32, next func()) {
	defer func() { dl.counter = 0 }()
	// log.Debug("Deadline: %v, Event: %v, RoomID: %+v", duration, util.Function{}.GetFunctionName(next), dl.RoomID)
	go func() {
		for t := range dl.clock.C {
			// log.Debug("ticker：%v", t)
			_ = t
			dl.counter++
			if duration == dl.counter {
				next()
				break
			}
		}
	}()
}

func (dl *Dealer) StartGame() {
	dl.AddBots()
	dl.Status = constant.RSBetting
	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime
	dl.ClockReset(con.ClearTime, dl.Bet)
}

// 下注
func (dl *Dealer) Bet() {
	dl.Status = constant.RSBetting
	dl.HRChan <- HRMsg{
		RoomID:     dl.RoomID,
		RoomStatus: dl.Status,
		EndTime:    uint32(time.Now().Unix() + constant.BetTime),
	}
	// log.Debug("bet... %+v", dl.RoomID)

	dl.ddl = uint32(time.Now().Unix()) + con.BetTime
	converter := DTOConverter{}

	// fixme 其实这种消息分开比较好，不然每次会有很多无谓的计算
	for u := range dl.Users {
		user := dl.Users[u]
		var autoBetSum float64
		for _, b := range dl.AutoBetRecord[u] {
			bet := b
			autoBetSum += constant.ChipSize[bet.Chip]
		}

		resp := converter.RSBMsg(0, autoBetSum, 0, *dl)
		user.ConnAgent.WriteMsg(&resp)
	}
	// 开始下注广播完之后，机器人开始下注
	go dl.BotsBet()

	dl.ClockReset(constant.BetTime, dl.Settle)
}

// 结算 开奖
func (dl *Dealer) Settle() {
	res := dl.profitPoolLottery()
	dl.res = res
	dl.Status = constant.RSSettle
	dl.HRChan <- HRMsg{
		RoomID:        dl.RoomID,
		RoomStatus:    dl.Status,
		LotteryResult: res,
	}
	// 结算
	// 庄家赢数 = Sum(所有筹码数) - 中奖倍数*中奖筹码数
	// 玩家赢数 = 开奖区域投注金额*区域倍数-总投注金额

	math := util.Math{}
	// todo 庄家结算
	dl.bankerWin, _ = math.SumSliceFloat64(dl.AreaBets).Sub(math.MultiFloat64(con.AreaX[dl.res], dl.AreaBets[dl.res])).Float64()
	// fixme 庄家是玩家的情况
	dl.bankerMoney = dl.bankerMoney + dl.bankerWin

	// log.Debug("settle... %+v", dl.RoomID)

	dl.ddl = uint32(time.Now().Unix()) + con.SettleTime
	converter := DTOConverter{}

	// fixme 用户离开房间之后要删除掉

	for uID := range dl.Users {
		// 用户在开奖区域投注数*区域倍数-用户所有投注数
		// 要么加投注赢得数，要么不加，和用户总数，是分开的
		user := dl.Users[uID]
		// 中心服需要结算的输赢
		uWin := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
		// 前端显示的输赢 精度问题
		uDisplayWin, _ := math.MultiFloat64(dl.UserBets[user.UserID][dl.res], constant.AreaX[dl.res]).Sub(math.SumSliceFloat64(dl.UserBets[user.UserID])).Float64()
		beforeBalance := user.Balance
		uuid := util.UUID{}
		order := uuid.GenUUID()
		if uWin > 0 {
			c4c.UserWinScore(user.UserID, uWin, order, func(data *User) {
				win, _ := decimal.NewFromFloat(data.Balance).Sub(math.SumSliceFloat64(dl.UserBets[user.UserID])).Sub(decimal.NewFromFloat(beforeBalance)).Float64()
				user.BalanceLock.Lock()
				user.Balance = data.Balance
				user.BalanceLock.Unlock()
				resp := converter.RSBMsg(win, 0, data.Balance, *dl)
				user.ConnAgent.WriteMsg(&resp)
			})
		} else {
			resp := converter.RSBMsg(uDisplayWin, 0, user.Balance, *dl)
			user.ConnAgent.WriteMsg(&resp)
		}
	}

	dl.ClockReset(constant.SettleTime, dl.ClearChip)
}

// 清理筹码
func (dl *Dealer) ClearChip() {
	dl.Status = constant.RSClear
	// log.Debug("clear chip... %+v", dl.RoomID)

	// 清理
	dl.AreaBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i := range dl.UserBets {
		dl.UserBets[i] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	dl.AreaBotBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	for u := range dl.Users {
		// 续投 投注 累加到 auto bet record
		if dl.UserAutoBet[u] == true && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = append(dl.AutoBetRecord[u], dl.UserBetsDetail[u]...)
			// 投注 不续投 覆盖掉
		} else if dl.UserAutoBet[u] == false && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = dl.UserBetsDetail[u]
			// 续投 不投注 不变
			// 不续投 不投注 不变
		}

		dl.UserAutoBet[u] = false
	}

	// 清空投注详情记录
	dl.UserBetsDetail = map[uint32][]msg.Bet{}

	dl.res = 0
	dl.bankerWin = 0
	dl.bankerRound += 1

	converter := DTOConverter{}

	if dl.bankerRound >= constant.BankerMaxTimes || dl.Bankers[0].GetBalance() < constant.BankerMinBar {
		if len(dl.Bankers) > 1 {
			dl.Bankers = dl.Bankers[1:]
			dl.bankerMoney = dl.Bankers[0].GetBalance()
		}
		// 换一批机器人
		dl.Bots = nil
		dl.AddBots()

		for _, b := range dl.Bots {
			if b.botType == constant.BTNextBanker {
				dl.Bankers = append(dl.Bankers, b)
			}
		}

		bankerResp := converter.BBMsg(*dl)
		dl.Broadcast(&bankerResp)
		dl.bankerRound = 0
	}

	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime

	resp := converter.RSBMsg(0, 0, 0, *dl)
	dl.Broadcast(&resp)

	dl.ClockReset(constant.ClearTime, dl.Bet)
}

func (dl *Dealer) Broadcast(m interface{}) {
	// log.Debug("room %+v brd %+v, content: %+v", dl.RoomID, reflect.TypeOf(m), m)
	for _, u := range dl.Users {
		user := u
		if user.ConnAgent != nil {
			user.ConnAgent.WriteMsg(m)
		}
	}
}

// 根据盈余池开奖
func (dl *Dealer) profitPoolLottery() uint32 {
	// 盈余池 随机从10%到50%取一个值，算出一个预计赔付数
	randomUtil := util.Random{}
	profitPoolRatePercent := randomUtil.RandInRange(constant.ProfitPoolMinPercent, constant.ProfitPoolMaxPercent)
	profitPoolRate := float64(profitPoolRatePercent) / 100.0
	acceptableMaxLose := profitPool() * profitPoolRate

	var area uint32
	for i := 0; i < 100; i++ {
		preArea := dl.fairLottery()
		preLoseAmount := preUserWin(dl.UserBets, preArea)
		if preLoseAmount > acceptableMaxLose {
			area = preArea
			continue
		} else {
			area = preArea
			break
		}
	}

	return area
}

// 公平开奖
func (dl *Dealer) fairLottery() uint32 {
	rand.Seed(time.Now().UnixNano())
	x := time.Duration(rand.Intn(5))
	time.Sleep(x * time.Nanosecond)
	prob := rand.Intn(121) // [0, 121)
	var area uint32

	if prob >= 0 && prob <= 2 {
		area = constant.Area40x
	} else if prob <= 6 {
		area = constant.Area30x
	} else if prob <= 12 {
		area = constant.Area20x
	} else if prob <= 24 {
		area = constant.Area10x
	} else if prob <= 48 {
		area = constant.Area5x1
	} else if prob <= 72 {
		area = constant.Area5x2
	} else if prob <= 96 {
		area = constant.Area5x3
	} else if prob <= 120 {
		area = constant.Area5x4
	}

	return area
}

// 玩家赢 - 官方庄家和机器人赢
// todo
func preUserWin(userBets map[uint32][]float64, preArea uint32) float64 {
	return 5
}

// 盈余池 = 玩家总输 - 玩家总赢 * 杀数 - (玩家数量 * 6)
// todo 统计计算玩家总赢和玩家总输、玩家数量
func profitPool() float64 {
	// 需要数据库
	// return pTotalLose - pTotalWin * constant.HouseEdgePercent - pCount*constant.GiftAmount
	return 20.0
}

package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/patrickmn/go-cache"
	"proj_bcbm/src/server/constant"
	con "proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Mgr <--> Dealer <--> C2C
//            ^
//            |
//
//           Bots

type Dealer struct {
	bankerWin   float64 // 庄家输赢
	bankerMoney float64 // 庄家余额

	*Room
	clock       *time.Ticker
	counter     uint32
	ddl         uint32
	bankerRound uint32 // 庄家做了多少轮

	RoundID string     // 轮次
	Status  uint32     // 房间状态
	res     uint32     // 最新开奖结果
	History []uint32   // 房间开奖历史
	HRChan  chan HRMsg // 房间大厅通信

	Users          sync.Map             // 房间用户-不包括机器人
	UserLeave      []uint32             // 用户是否在房间
	Bots           []*Bot               // 房间机器人
	Bankers        []Player             // 上庄玩家榜单
	DownBanker     bool                 // 手动下庄
	UserBets       map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
	UserBetsDetail map[uint32][]msg.Bet // 用户具体投注
	UserAutoBet    map[uint32]bool      // 本局投注记录
	AutoBetRecord  map[uint32][]msg.Bet // 续投记录
	AreaBets       []float64            // 每个区域玩家投注总数
	AreaBotBets    []float64            // 每个区域机器人投注总数
}

func NewDealer(rID uint32, hr chan HRMsg) *Dealer {
	return &Dealer{
		Users:          sync.Map{},
		UserLeave:      make([]uint32, 0),
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
	// 时间戳+随机数，每局一个
	uid := util.UUID{}
	dl.RoundID = fmt.Sprintf("%+v-%+v", time.Now().Unix(), uid.GenUUID())

	dl.Status = constant.RSBetting
	dl.ddl = uint32(time.Now().Unix()) + con.BetTime

	dl.HRChan <- HRMsg{
		RoomID:     dl.RoomID,
		RoomStatus: dl.Status,
		EndTime:    uint32(time.Now().Unix() + constant.BetTime),
	}
	// log.Debug("bet... %+v", dl.RoomID)

	converter := DTOConverter{}

	// fixme 其实这种消息分开比较好，不然每次会有很多无谓的计算
	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		var autoBetSum float64
		for _, b := range dl.AutoBetRecord[key.(uint32)] {
			bet := b
			autoBetSum += constant.ChipSize[bet.Chip]
		}
		resp := converter.RSBMsg(0, autoBetSum, 0, *dl)
		user.ConnAgent.WriteMsg(&resp)
		return true
	})
	// 开始下注广播完之后，机器人开始下注
	go dl.BotsBet()

	dl.ClockReset(constant.BetTime, dl.Settle)
}

// 结算 开奖
func (dl *Dealer) Settle() {
	res := dl.profitPoolLottery()
	dl.res = res

	dl.Status = constant.RSSettle
	dl.ddl = uint32(time.Now().Unix()) + con.SettleTime

	dl.HRChan <- HRMsg{
		RoomID:        dl.RoomID,
		RoomStatus:    dl.Status,
		LotteryResult: res,
	}

	uuid := util.UUID{}

	// 结算
	// 庄家赢数 = Sum(所有筹码数) - 中奖倍数*中奖筹码数
	// 玩家赢数 = 开奖区域投注金额*区域倍数-总投注金额

	math := util.Math{}
	// 税前庄家输赢
	preBankerWin, _ := math.SumSliceFloat64(dl.AreaBets).Sub(math.MultiFloat64(con.AreaX[dl.res], dl.AreaBets[dl.res])).Float64()
	switch dl.Bankers[0].(type) {
	case User:
		{
			u := dl.Bankers[0].(User)
			v, _ := dl.Users.Load(u.UserID)
			up := v.(*User)
			preBalance := up.Balance
			order := uuid.GenUUID()

			if preBankerWin > 0 {
				c4c.UserWinScore(u.UserID, preBankerWin, 0, 0, order+"-banker-win", dl.RoundID, func(data *User) {
					up.BalanceLock.Lock()
					up.Balance = data.Balance
					up.BalanceLock.Unlock()

					win, _ := decimal.NewFromFloat(data.Balance).Sub(decimal.NewFromFloat(preBalance)).Float64()

					dl.bankerWin = win
					dl.bankerMoney = dl.bankerMoney + dl.bankerWin

					u.IncBankerBalance(dl.bankerWin)
					u.IncBalance(dl.bankerWin)
				})
			} else {
				c4c.UserLoseScore(u.UserID, preBankerWin, 0, 0, order+"-banker-lose", dl.RoundID, func(data *User) {
					up.BalanceLock.Lock()
					up.Balance = data.Balance
					up.BalanceLock.Unlock()

					win, _ := decimal.NewFromFloat(data.Balance).Sub(decimal.NewFromFloat(preBalance)).Float64()

					dl.bankerWin = win
					dl.bankerMoney = dl.bankerMoney + dl.bankerWin

					u.IncBankerBalance(dl.bankerWin)
					u.IncBalance(dl.bankerWin)
				})
			}
		}
	case Bot:
		dl.bankerWin = preBankerWin * 0.95
		dl.bankerMoney = dl.bankerMoney + dl.bankerWin
	}

	time.Sleep(500 * time.Millisecond)

	// log.Debug("settle... %+v", dl.RoomID)
	dl.playerSettle()

	// 处理离开房间的用户
	for _, uid := range dl.UserLeave {
		userID := uid
		_, ok := dl.Users.Load(userID)
		if ok {
			dl.Users.Delete(userID)
		}
	}

	dl.ClockReset(constant.SettleTime, dl.ClearChip)
}

func (dl *Dealer) playerSettle() {
	dtoC := DTOConverter{}
	daoC := DAOConverter{}
	math := util.Math{}
	uuid := util.UUID{}
	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		// 中心服需要结算的输赢
		uWin := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
		// 前端显示的输赢 精度问题
		uDisplayWin, _ := math.MultiFloat64(dl.UserBets[user.UserID][dl.res], constant.AreaX[dl.res]).Sub(math.SumSliceFloat64(dl.UserBets[user.UserID])).Float64()
		beforeBalance := user.Balance

		order := uuid.GenUUID()
		var winFlag bool
		if uWin > 0 {
			winFlag = true
			c4c.UserWinScore(user.UserID, uWin, 0, 0, order, dl.RoundID, func(data *User) {
				win, _ := decimal.NewFromFloat(data.Balance).Sub(math.SumSliceFloat64(dl.UserBets[user.UserID])).Sub(decimal.NewFromFloat(beforeBalance)).Float64()
				// 赢钱之后更新余额
				user.BalanceLock.Lock()
				user.Balance = data.Balance
				user.BalanceLock.Unlock()
				resp := dtoC.RSBMsg(win, 0, data.Balance, *dl)
				user.ConnAgent.WriteMsg(&resp)
			})
		} else {
			winFlag = false
			resp := dtoC.RSBMsg(uDisplayWin, 0, user.Balance, *dl)
			user.ConnAgent.WriteMsg(&resp)
		}

		// 玩家结算记录
		uBet, _ := math.SumSliceFloat64(dl.UserBets[user.UserID]).Float64()
		if uBet > 0 {
			sdb := daoC.Settle2DB(*user, order, dl.RoundID, winFlag, uBet, uWin)
			err := db.CUserSettle(sdb)
			if err != nil {
				log.Debug("保存用户结算数据错误 %+v", err)
			}

			err = db.UProfitPool(uBet, uWin)
			if err != nil {
				log.Debug("更新盈余池失败 %+v", err)
			}
		}

		return true
	})
}

// 清理筹码
func (dl *Dealer) ClearChip() {
	dl.Status = constant.RSClear
	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime

	// log.Debug("clear chip... %+v", dl.RoomID)

	// 更新玩家列表数据
	dl.UpdatePlayerList()
	// 清空数据
	dl.ClearData()

	converter := DTOConverter{}

	resp := converter.RSBMsg(0, 0, 0, *dl)
	dl.Broadcast(&resp)

	// 玩家金币总数不足上庄金币数 移除金币不足的列表中玩家
	for i, b := range dl.Bankers {
		banker := b
		if banker.GetBalance() < banker.GetBankerBalance() && i != 0 {
			dl.Bankers = append(dl.Bankers[:i], dl.Bankers[i+1:]...)
		}
	}

	// 庄家轮换
	fmt.Println("************ 上庄金币 ****************", dl.Bankers[0].GetBankerBalance())
	if dl.bankerRound >= constant.BankerMaxTimes || dl.Bankers[0].GetBankerBalance() < constant.BankerMinBar || dl.DownBanker == true {
		if len(dl.Bankers) > 1 {
			dl.Bankers = dl.Bankers[1:]
			dl.bankerMoney = dl.Bankers[0].GetBankerBalance()
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
		dl.DownBanker = false
	}

	// 下一阶段
	dl.ClockReset(constant.ClearTime, dl.Bet)
}

func (dl *Dealer) Broadcast(m interface{}) {
	// log.Debug("room %+v brd %+v, content: %+v", dl.RoomID, reflect.TypeOf(m), m)
	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		if user.ConnAgent != nil {
			user.ConnAgent.WriteMsg(m)
		}

		return true
	})
}

func (dl *Dealer) UpdatePlayerList() {
	//玩家列表数据统计
	// 玩家结算记录
	math := util.Math{}
	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		uBet, _ := math.SumSliceFloat64(dl.UserBets[user.UserID]).Float64()
		if uBet > 0 {
			if _, exist := ca.Get(fmt.Sprintf("%+v-betAmount", user.UserID)); !exist {
				var winCount int64
				winCount = 0
				ca.Set(fmt.Sprintf("%+v-betAmount", user.UserID), 0.0, cache.DefaultExpiration)
				ca.Set(fmt.Sprintf("%+v-winCount", user.UserID), winCount, cache.DefaultExpiration)
			} else {
				addBet, err := ca.IncrementFloat64(fmt.Sprintf("%+v-betAmount", user.UserID), uBet)
				if err != nil {
					log.Debug("累加用户投注数错误 %+v", err)
				}

				log.Debug("用户累计投注 %+v", addBet)

				uWin := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
				if uWin > 0 {
					addWin, err := ca.IncrementInt64(fmt.Sprintf("%+v-winCount", user.UserID), 1)
					if err != nil {
						log.Debug("累加用户赢数错误 %+v", err)
					}

					log.Debug("用户累计赢数 %+v", addWin)
				}
			}
		} else {
			if _, exist := ca.Get(fmt.Sprintf("%+v-betAmount", user.UserID)); !exist {
				var winCount int64
				winCount = 0
				ca.Set(fmt.Sprintf("%+v-betAmount", user.UserID), 0.0, cache.DefaultExpiration)
				ca.Set(fmt.Sprintf("%+v-winCount", user.UserID), winCount, cache.DefaultExpiration)
			}
		}
		return true
	})
}

func (dl *Dealer) ClearData() {
	// 清理
	dl.AreaBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i := range dl.UserBets {
		dl.UserBets[i] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	dl.AreaBotBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

	dl.Users.Range(func(key, value interface{}) bool {
		// 续投 投注 累加到 auto bet record
		u := key.(uint32)
		if dl.UserAutoBet[u] == true && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = append(dl.AutoBetRecord[u], dl.UserBetsDetail[u]...)
			// 投注 不续投 覆盖掉
		} else if dl.UserAutoBet[u] == false && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = dl.UserBetsDetail[u]
			// 续投 不投注 不变
			// 不续投 不投注 不变
		}

		dl.UserAutoBet[u] = false
		return true
	})

	// 清空投注详情记录
	dl.UserBetsDetail = map[uint32][]msg.Bet{}
	dl.UserLeave = []uint32{}

	dl.res = 0
	dl.bankerWin = 0
	dl.bankerRound += 1

}

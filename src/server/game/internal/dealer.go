package internal

import (
	"fmt"
	"github.com/shopspring/decimal"
	"proj_bcbm/src/server/constant"
	con "proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"strconv"
	"sync"
	"time"
)

// Mgr <--> Dealer <--> C2C
//            ^
//            |
//           Bots

type Dealer struct {
	bankerWin   float64 // 庄家输赢
	bankerMoney float64 // 庄家余额

	*Room
	clock       *time.Ticker
	counter     uint32
	ddl         uint32
	bankerRound uint32 // 庄家做了多少轮

	RoundID   string     // 轮次
	Status    uint32     // 房间状态
	res       uint32     // 最新开奖结果
	pos       uint32     // 开奖位置
	History   []uint32   // 房间开奖历史
	HRChan    chan HRMsg // 房间大厅通信
	IsDownBet bool       // 设置机器人下注状态

	Users          sync.Map             // 房间用户-不包括机器人
	UserLeave      []uint32             // 用户是否在房间
	Bots           []*Bot               // 房间机器人
	Bankers        []Player             // 上庄玩家榜单
	DownBanker     bool                 // 手动下
	UserBets       map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
	UserBetsDetail map[uint32][]msg.Bet // 用户具体投注
	UserAutoBet    map[uint32]bool      // 本局投注记录
	AutoBetRecord  map[uint32][]msg.Bet // 续投记录
	AreaBets       []float64            // 每个区域玩家投注总数
	AreaBotBets    []float64            // 每个区域机器人投注总数

	DownBetTotal float64 //玩家总下注
}

const taxRate = 0.05

func NewDealer(rID uint32, hr chan HRMsg) *Dealer {
	return &Dealer{
		Users:          sync.Map{},
		UserLeave:      make([]uint32, 0),
		Bankers:        make([]Player, 0),
		Room:           NewRoom(rID, con.RL1MinBet, con.RL1MaxBet, con.RL1MinLimit),
		clock:          time.NewTicker(time.Second),
		HRChan:         hr,
		IsDownBet:      false,
		UserAutoBet:    map[uint32]bool{},
		UserBets:       map[uint32][]float64{},
		UserBetsDetail: map[uint32][]msg.Bet{},
		AutoBetRecord:  map[uint32][]msg.Bet{},
		AreaBets:       []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
		AreaBotBets:    []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
		DownBetTotal:   0,
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
			if duration-1 == dl.counter {
				dl.IsDownBet = false
			}
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

	ru := util.Random{}
	dl.pos = uint32(ru.RandInRange(1, 5))

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
			preBankerBalance := dl.bankerMoney
			order := uuid.GenUUID()

			if preBankerWin > 0 {
				log.Debug("玩家的当局总下注1: %v", preBankerWin)
				c4c.BankerWinScore(u.UserID, preBankerWin, order+"-banker-win", dl.RoundID, func(data *User) {
					dl.bankerWin, _ = decimal.NewFromFloat(data.BankerBalance).Sub(decimal.NewFromFloat(preBankerBalance)).Float64()
					log.Debug("玩家的当局总下注2: %v", dl.bankerWin)
					//////庄家跑马灯
					//if dl.bankerWin > PaoMaDeng {
					//	c4c.NoticeWinMoreThan(u.UserID, u.NickName, dl.bankerWin)
					//}
					dl.bankerMoney = data.BankerBalance
					// 玩家坐庄盈余池更新
					err := db.UProfitPool(0, dl.bankerWin, dl.RoomID)
					if err != nil {
						log.Debug("更新盈余池失败 %+v", err)
					}
				})
			} else {
				c4c.BankerLoseScore(u.UserID, preBankerWin, order+"-banker-lose", dl.RoundID, func(data *User) {
					dl.bankerWin, _ = decimal.NewFromFloat(data.BankerBalance).Sub(decimal.NewFromFloat(preBankerBalance)).Float64()
					dl.bankerMoney = data.BankerBalance

					// 玩家坐庄盈余池更新
					err := db.UProfitPool(-dl.bankerWin, 0, dl.RoomID)
					if err != nil {
						log.Debug("更新盈余池失败 %+v", err)
					}
				})
			}

			time.Sleep(200 * time.Millisecond)
		}
	default:
		{
			if preBankerWin > 0 {
				dl.bankerWin = preBankerWin * 0.95
				////机器人跑马灯
				//u := dl.Bankers[0].(User)
				//if dl.bankerWin > PaoMaDeng {
				//	c4c.NoticeWinMoreThan(u.UserID, u.NickName, dl.bankerWin)
				//}
				dl.bankerMoney = dl.bankerMoney + dl.bankerWin
			} else {
				dl.bankerWin = preBankerWin
				dl.bankerMoney = dl.bankerMoney + dl.bankerWin
			}
		}
	}

	// log.Debug("settle... %+v", dl.RoomID)
	dl.playerSettle()

	// 处理离开房间的用户
	for _, uid := range dl.UserLeave {
		userID := uid
		_, ok := dl.Users.Load(userID)
		if ok {
			dl.Users.Delete(userID)
			c4c.UserLogoutCenter(userID, func(data *User) {
				log.Debug("投注后离开房间的玩家已登出")
			})
		}
	}

	dl.ClockReset(constant.SettleTime, dl.ClearChip)
}

func (dl *Dealer) playerSettle() {
	dtoC := DTOConverter{}
	daoC := DAOConverter{}
	math := util.Math{}

	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		// 中心服需要结算的输赢
		uWin := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]

		var ResultMoney float64

		var winFlag bool
		if uWin > 0 {
			winFlag = true
			uWin = uWin - dl.UserBets[user.UserID][dl.res]
			ResultMoney += uWin - (uWin * taxRate)

			winOrder := strconv.Itoa(int(user.UserID)) + "-" + time.Now().Format("2006-01-02 15:04:05") + "win"
			c4c.UserWinScore(user.UserID, uWin, winOrder, dl.RoundID, func(data *User) {
				// 赢钱之后更新余额
				user.BalanceLock.Lock()
				user.Balance = data.Balance
				user.BalanceLock.Unlock()
			})
		} else {
			winFlag = false
		}

		loseOrder := strconv.Itoa(int(user.UserID)) + "-" + time.Now().Format("2006-01-02 15:04:05") + "lose"
		if dl.DownBetTotal > 0 {
			if uWin > 0 {
				ResultMoney -= dl.DownBetTotal - dl.UserBets[user.UserID][dl.res]
				result := -dl.DownBetTotal + dl.UserBets[user.UserID][dl.res]
				c4c.UserLoseScore(user.UserID, result, loseOrder, "", func(data *User) {
					user.BalanceLock.Lock()
					user.Balance = data.Balance
					user.BalanceLock.Unlock()
				})
			} else {
				ResultMoney -= dl.DownBetTotal
				c4c.UserLoseScore(user.UserID, -dl.DownBetTotal, loseOrder, "", func(data *User) {
					user.BalanceLock.Lock()
					user.Balance = data.Balance
					user.BalanceLock.Unlock()
				})
			}
		}

		if ResultMoney > 0 {
			user.Balance += dl.DownBetTotal + ResultMoney
		}

		resp := dtoC.RSBMsg(ResultMoney, 0, user.Balance, *dl)
		user.ConnAgent.WriteMsg(&resp)

		if ResultMoney > PaoMaDeng {
			c4c.NoticeWinMoreThan(user.UserID, user.NickName, ResultMoney)
		}

		// 玩家结算记录
		uBet, _ := math.SumSliceFloat64(dl.UserBets[user.UserID]).Float64()
		if uBet > 0 && uWin >= 0 {
			order := strconv.Itoa(int(user.UserID)) + "-" + time.Now().Format("2006-01-02 15:04:05")
			sdb := daoC.Settle2DB(*user, order, dl.RoundID, winFlag, uBet, uWin)
			err := db.CUserSettle(sdb)
			if err != nil {
				log.Debug("保存用户结算数据错误 %+v", err)
			}

			err = db.UProfitPool(uBet, uWin, dl.RoomID)
			if err != nil {
				log.Debug("更新盈余池失败 %+v", err)
			}
		}

		if dl.DownBetTotal > 0 {
			timeNow := time.Now().Unix()
			data := &PlayerDownBetRecode{}
			data.Id = user.UserID
			data.RandId = dl.RoomID + - +uint32(timeNow)
			data.RoomId = dl.RoomID
			data.DownBetInfo = dl.UserBets[user.UserID]
			data.DownBetTime = timeNow
			data.CardResult = dl.res
			data.ResultMoney = ResultMoney
			data.TaxRate = taxRate

			err := db.InsertAccess(data)
			if err != nil {
				log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
			}
		}

		return true
	})
}

// 清理筹码
func (dl *Dealer) ClearChip() {
	dl.Status = constant.RSClear
	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime
	dl.DownBetTotal = 0

	// log.Debug("clear chip... %+v", dl.RoomID)

	// 更新玩家列表数据
	dl.UpdatePlayerList()
	// 清空数据
	dl.ClearData()

	converter := DTOConverter{}
	uuid := util.UUID{}

	resp := converter.RSBMsg(0, 0, 0, *dl)
	dl.Broadcast(&resp)

	// 庄家轮换
	if dl.bankerRound >= constant.BankerMaxTimes || dl.bankerMoney < constant.BankerMinBar || dl.DownBanker == true {
		// 加回玩家的钱
		switch dl.Bankers[0].(type) {
		case User:
			uid, _, _, _ := dl.Bankers[0].GetPlayerBasic()
			c4c.ChangeBankerStatus(uid, constant.BSNotBanker, -dl.bankerMoney, fmt.Sprintf("%+v-notBanker", uuid.GenUUID()), dl.RoundID, func(data *User) {
				data.Status = constant.BSNotBanker
				bankerStatus = constant.BSNotBanker
				log.Debug("玩家状态 :%v", data.Status)

				log.Debug("<--- 玩家下庄 --->")
				bankerResp := msg.BankersB{
					Banker: dl.getBankerInfoResp(),
					UpdateBanker: &msg.UserInfo{
						UserID: data.UserID,
						Money:  data.Balance,
					},
					ServerTime: uint32(time.Now().Unix()),
				}

				dl.Broadcast(&bankerResp)
			})

			// 如果玩家不在线，登出
			_, ok := dl.Users.Load(uid)
			if !ok {
				c4c.UserLogoutCenter(uid, func(data *User) {
					log.Debug("庄家已不在游戏中，下庄后自动登出 %+v", uid)
				})
			}
		}

		// 新庄家
		if len(dl.Bankers) > 1 {
			dl.Bankers = dl.Bankers[1:]
			dl.bankerMoney = dl.Bankers[0].GetBankerBalance()
			switch dl.Bankers[0].(type) {
			case User:
				uid, _, _, _ := dl.Bankers[0].GetPlayerBasic()
				c4c.ChangeBankerStatus(uid, constant.BSBeingBanker, 0, fmt.Sprintf("%+v-beBanker", uuid.GenUUID()), dl.RoundID, func(data *User) {
					data.Status = constant.BSBeingBanker
					bankerStatus = constant.BSBeingBanker
					log.Debug("玩家状态 :%v", data.Status)
					dec := util.Math{}
					var ok bool
					dl.bankerMoney, ok = dec.AddFloat64(data.BankerBalance, 0.0).Float64()
					log.Debug("<--- 玩家上庄，精度 %+v--->", ok)
				})
			}
		}

		// 换一批机器人
		dl.Bots = nil
		dl.AddBots()

		for _, b := range dl.Bots {
			if b.botType == constant.BTNextBanker {
				dl.Bankers = append(dl.Bankers, b)
			}
		}

		bankerResp := msg.BankersB{
			Banker:     dl.getBankerInfoResp(),
			ServerTime: uint32(time.Now().Unix()),
		}

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
		log.Debug("UpdatePlayerList  ~~: %v", uBet)
		winCount, bet := user.GetPlayerAccount()
		log.Debug("玩家win局数: %v 和 下注金额: %v", winCount, bet)

		if uBet > 0 {
			win := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
			result := win - dl.DownBetTotal
			if result > 0 {
				if _, exist := ca.Get(fmt.Sprintf("%+v-betAmount", user.UserID)); exist {
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
				} else {
					log.Debug("赢钱没有获取到用户 %+v", exist)
				}
			} else {
				if _, exist := ca.Get(fmt.Sprintf("%+v-betAmount", user.UserID)); exist {
					addBet, err := ca.IncrementFloat64(fmt.Sprintf("%+v-betAmount", user.UserID), uBet)
					if err != nil {
						log.Debug("累加用户投注数错误 %+v", err)
					}
					log.Debug("用户累计投注 %+v", addBet)
				} else {
					log.Debug("输钱没有获取到用户 %+v", exist)
				}
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

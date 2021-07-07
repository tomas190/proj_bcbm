package internal

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"proj_bcbm/src/server/conf"
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
	TotalDownMoney float64              // 当局所有总下注
	DownBetArea    []float64
	UserIsDownBet  map[uint32]bool // 玩家当局是否下注
}

const taxRate = 0.05

var packageTax map[uint16]float64

var downBankerChan chan bool

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
		UserIsDownBet:  map[uint32]bool{},
		AreaBets:       []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
		DownBetArea:    []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
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
	//log.Debug("Game 下注阶段~")
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
	//log.Debug("Game 结算阶段~")

	res := dl.profitPoolLottery()
	dl.res = res

	ru := util.Random{}
	dl.pos = uint32(ru.RandInRange(1, 5))

	dl.Status = constant.RSSettle
	dl.ddl = uint32(time.Now().Unix()) + con.SettleTime

	dl.HRChan <- HRMsg{
		RoomID:        dl.RoomID,
		RoomStatus:    dl.Status,
		LotteryResult: dl.res,
	}

	// 结算
	// 庄家赢数 = Sum(所有筹码数) - 中奖倍数*中奖筹码数
	// 玩家赢数 = 开奖区域投注金额*区域倍数-总投注金额

	math := util.Math{}
	// 税前庄家输赢
	preBankerWin, _ := math.SumSliceFloat64(dl.AreaBets).Sub(math.MultiFloat64(con.AreaX[dl.res], dl.AreaBets[dl.res])).Float64()

	var ResultMoney float64
	switch dl.Bankers[0].(type) {
	case User:
		{
			u := dl.Bankers[0].(User)
			//preBankerBalance := dl.bankerMoney
			order := bson.NewObjectId().Hex()

			if preBankerWin > 0 {
				log.Debug("玩家的当局总下注和庄家金额: %v,%v", preBankerWin, dl.bankerMoney)
				c4c.BankerWinScore(u.UserID, preBankerWin, order, dl.RoundID, func(data *User) {
					//dl.bankerWin, _ = decimal.NewFromFloat(data.BankerBalance).Sub(decimal.NewFromFloat(preBankerBalance)).Float64()
					//log.Debug("玩家的当局总下注2: %v", dl.bankerWin)
					dl.bankerMoney = data.BankerBalance
				})
				pac := packageTax[u.PackageId]
				taxR := float64(pac) / 100

				ResultMoney += preBankerWin - (preBankerWin * taxR)
				dl.bankerWin += preBankerWin - (preBankerWin * taxR)
				log.Debug("庄家金额为和庄家赢钱:%v，%v", dl.bankerMoney, dl.bankerWin)
				// 玩家坐庄盈余池更新
				err := db.UProfitPool(0, dl.bankerWin, dl.RoomID)
				if err != nil {
					log.Debug("更新盈余池失败 %+v", err)
				}
			} else {
				c4c.BankerLoseScore(u.UserID, preBankerWin, order, dl.RoundID, func(data *User) {
					//dl.bankerWin, _ = decimal.NewFromFloat(data.BankerBalance).Sub(decimal.NewFromFloat(preBankerBalance)).Float64()
					dl.bankerMoney = data.BankerBalance
				})
				ResultMoney = preBankerWin
				dl.bankerWin = preBankerWin
				log.Debug("庄家金额为和庄家赢钱:%v，%v", dl.bankerMoney, dl.bankerWin)
				// 玩家坐庄盈余池更新
				err := db.UProfitPool(-dl.bankerWin, 0, dl.RoomID)
				if err != nil {
					log.Debug("更新盈余池失败 %+v", err)
				}
			}

			log.Debug("庄家当局的输赢:%v", dl.bankerWin)
			if preBankerWin != 0 {
				timeNow := time.Now().Unix()
				data := &PlayerDownBetRecode{}
				data.Id = strconv.Itoa(int(u.UserID))
				data.GameId = conf.Server.GameID
				data.RoundId = dl.RoundID
				data.RoomId = dl.RoomID
				data.DownBetInfo = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
				for i := 0; i < len(data.DownBetInfo); i++ {
					data.DownBetInfo[i] += dl.AreaBets[i]
				}
				data.DownBetTime = timeNow
				data.StartTime = timeNow - 16
				data.EndTime = timeNow + 25
				data.CardResult = dl.res
				data.SettlementFunds = ResultMoney
				data.SpareCash = u.Balance
				data.TaxRate = taxRate
				err := db.InsertAccess(data)
				if err != nil {
					log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
				}
			}
			// todo 
			//time.Sleep(200 * time.Millisecond)
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

	r := util.Random{}
	for _, v := range dl.Bots {
		if v != nil {
			num := r.RandInRange(0, 100)
			if num <= 70 {
				v.TwentyData = append(v.TwentyData, 0)
			} else {
				v.TwentyData = append(v.TwentyData, 1)
			}
			if len(v.TwentyData) > 20 {
				v.TwentyData = append(v.TwentyData[:0], v.TwentyData[1:]...)
			}
			var n uint32
			for _, d := range v.TwentyData {
				if d == 1 {
					n++
				}
			}
			v.WinCount = n
		}
	}
	dl.ClockReset(constant.SettleTime, dl.ClearChip)
}

func (dl *Dealer) playerSettle() {
	dtoC := DTOConverter{}
	daoC := DAOConverter{}

	uid := util.UUID{}
	dl.RoundID = fmt.Sprintf("%+v-%+v", time.Now().Unix(), uid.GenUUID())
	dl.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		// 中心服需要结算的输赢
		uWin := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
		log.Debug("res:%v,AreaX:%v", dl.UserBets[user.UserID][dl.res], constant.AreaX[dl.res])
		timeNow := time.Now().Unix()

		var ResultMoney float64
		var uBet float64
		var data float64

		var winFlag bool

		//if dl.UserIsDownBet[user.UserID] == false {
		//	return true
		//}

		order := bson.NewObjectId().Hex()
		uid := util.UUID{}
		roundId := fmt.Sprintf("%+v-%+v", time.Now().Unix(), uid.GenUUID())
		c4c.UnlockSettlement(user, order, roundId)

		if uWin > 0 {
			winFlag = true
			uWin = uWin - dl.UserBets[user.UserID][dl.res]

			pac := packageTax[user.PackageId]
			taxR := float64(pac) / 100
			data += uWin - (uWin * taxR)
			log.Debug("uWin:%v,data:%v", uWin, data)
			user.Balance += dl.UserBets[user.UserID][dl.res] + data
			log.Debug("res:%v,Balance:%v", dl.UserBets[user.UserID][dl.res], user.Balance)

			ResultMoney += uWin - (uWin * taxR)
			winOrder := bson.NewObjectId().Hex()
			c4c.UserWinScore(uint32(timeNow), user.UserID, uWin, dl.UserBets[user.UserID][dl.res], winOrder, dl.RoundID, func(data *User) {
				// 赢钱之后更新余额
				user.BalanceLock.Lock()
				user.Balance = data.Balance
				user.BalanceLock.Unlock()
			})
			//select {
			//case t := <-winChan:
			//	if t == true {
			//		break
			//	}
			//}
		} else {
			winFlag = false
		}
		var uLose float64
		if user.DownBetTotal > 0 {
			loseOrder := bson.NewObjectId().Hex()
			if uWin > 0 {
				uBet = user.DownBetTotal - dl.UserBets[user.UserID][dl.res]
				ResultMoney -= user.DownBetTotal - dl.UserBets[user.UserID][dl.res]
				data -= user.DownBetTotal - uBet

				result := -user.DownBetTotal + dl.UserBets[user.UserID][dl.res]
				if result != 0 {
					c4c.UserLoseScore(uint32(timeNow), user.UserID, result, uBet, loseOrder, dl.RoundID, func(data *User) {
						user.BalanceLock.Lock()
						user.Balance = data.Balance
						user.BalanceLock.Unlock()
					})
					uLose = result
					//select {
					//case t := <-loseChan:
					//	if t == true {
					//		break
					//	}
					//}
				}
			} else {
				uBet = user.DownBetTotal
				ResultMoney -= user.DownBetTotal
				c4c.UserLoseScore(uint32(timeNow), user.UserID, -user.DownBetTotal, uBet, loseOrder, dl.RoundID, func(data *User) {
					user.BalanceLock.Lock()
					user.Balance = data.Balance
					user.BalanceLock.Unlock()
				})
				uLose = -user.DownBetTotal
				//select {
				//case t := <-loseChan:
				//	if t == true {
				//		break
				//	}
				//}
			}
		}

		resp := dtoC.RSBMsg(ResultMoney, 0, user.Balance, *dl)
		user.ConnAgent.WriteMsg(&resp)

		if ResultMoney > PaoMaDeng {
			c4c.NoticeWinMoreThan(user.UserID, user.NickName, ResultMoney)
		}

		// 玩家结算记录
		if uWin == 0 && uBet == 0 {
			//log.Debug("空数据,不插入")
		} else {
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

		if user.DownBetTotal > 0 {
			data := &PlayerDownBetRecode{}
			data.Id = strconv.Itoa(int(user.UserID))
			data.GameId = conf.Server.GameID
			data.RoundId = dl.RoundID
			data.RoomId = dl.RoomID
			data.DownBetInfo = dl.UserBets[user.UserID]
			data.DownBetTime = timeNow
			data.StartTime = timeNow - 16
			data.EndTime = timeNow + 25
			data.CardResult = dl.res
			data.SettlementFunds = ResultMoney
			data.SpareCash = user.Balance
			data.TaxRate = taxRate
			err := db.InsertAccess(data)
			if err != nil {
				log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
			}

			// 插入游戏统计数据
			sd := &StatementData{}
			sd.Id = strconv.Itoa(int(user.UserID))
			sd.GameId = conf.Server.GameID
			sd.GameName = "奔驰宝马"
			sd.StartTime = timeNow - 16
			sd.EndTime = timeNow + 25
			sd.DownBetTime = timeNow
			sd.PackageId = user.PackageId
			sd.WinStatementTotal = uWin
			sd.LoseStatementTotal = uLose
			sd.BetMoney = user.DownBetTotal
			db.InsertStatementDB(sd)
		}

		ResultMoney = 0
		return true
	})
}

// 清理筹码
func (dl *Dealer) ClearChip() {
	//log.Debug("Game 清除筹码~")

	dl.Status = constant.RSClear
	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime

	// log.Debug("clear chip... %+v", dl.RoomID)

	// 处理离开房间的用户
	for _, uid := range dl.UserLeave {
		userID := uid
		p, ok := dl.Users.Load(userID)
		if ok {
			player := p.(*User)
			dl.Users.Delete(userID)
			c4c.UserLogoutCenter(userID, func(data *User) {
				dl.AutoBetRecord[player.UserID] = nil
				Mgr.UserRecord.Delete(player.UserID)
				resp := &msg.LogoutR{}
				player.ConnAgent.WriteMsg(resp)
				player.ConnAgent.Close()
				log.Debug("投注后离开房间的玩家已登出")
			})
		}
	}

	// 更新玩家列表数据
	dl.UpdatePlayerList()

	// 清空数据
	dl.ClearData()

	converter := DTOConverter{}
	//uuid := util.UUID{}

	resp := converter.RSBMsg(0, 0, 0, *dl)
	dl.Broadcast(&resp)

	// 庄家轮换
	if dl.bankerRound >= constant.BankerMaxTimes || dl.bankerMoney < constant.BankerMinBar || dl.DownBanker == true {
		// 加回玩家的钱
		switch dl.Bankers[0].(type) {
		case User:
			uid, _, _, _ := dl.Bankers[0].GetPlayerBasic()
			order := bson.NewObjectId().Hex()
			var balance float64 = 0
			for _, v := range dl.Bots {
				if v.UserID == uid {
					v.Status = constant.BSNotBanker
				}
			}
			c4c.ChangeBankerStatus(uid, constant.BSNotBanker, -dl.bankerMoney, order, dl.RoundID, func(data *User) {

				// 更新庄家状态
				dl.Users.Range(func(key, value interface{}) bool {
					if key == uid {
						u := value.(*User)
						u.Status = constant.BSNotBanker
					}
					return true
				})
				log.Debug("<--- 玩家下庄 --->:%v", data.Balance)
				balance = data.Balance

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
			time.Sleep(10 * time.Millisecond)

			// 更新庄家状态
			dl.Users.Range(func(key, value interface{}) bool {
				if key == uid {
					u := value.(*User)
					u.Balance = balance
					resp := &msg.BetInfoB{
						PlayerID: u.UserID,
						Money:    u.Balance,
					}
					u.ConnAgent.WriteMsg(resp)
				}
				return true
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
				order := bson.NewObjectId().Hex()
				for _, v := range dl.Bots {
					if v.UserID == uid {
						v.Status = constant.BSBeingBanker
					}
				}
				c4c.ChangeBankerStatus(uid, constant.BSBeingBanker, 0, order, dl.RoundID, func(data *User) {

					// 更新庄家状态
					dl.Users.Range(func(key, value interface{}) bool {
						if key == uid {
							u := value.(*User)
							u.Status = constant.BSBeingBanker
						}
						return true
					})
					dec := util.Math{}
					var ok bool
					dl.bankerMoney, ok = dec.AddFloat64(data.BankerBalance, 0.0).Float64()
					log.Debug("<--- 玩家上庄，精度 %+v--->", ok)
				})
			}
		}

		// 换一批机器人
		//dl.Bots = nil
		dl.AddBots()

		// 隔几把才变动？
		if len(dl.Bankers) <= 1 {
			log.Debug("添加机器人庄家2")
			nextB := dl.NextBotBanker()
			dl.Bankers = append(dl.Bankers, nextB)
			dl.Bots = append(dl.Bots, &nextB)
		}

		//ru := util.Random{}
		//num := ru.RandInRange(0, 100)
		//if num >= 0 && num <= 50 {
		//	nextB := dl.NextBotBanker()
		//	dl.Bankers = append(dl.Bankers, nextB)
		//	dl.Bots = append(dl.Bots, &nextB)
		//} else if num > 50 && num <= 100 {
		//	var botId uint32
		//	for _, b := range dl.Bots {
		//		if b.botType == constant.BTNextBanker {
		//			botId = b.UserID
		//		}
		//	}
		//
		//	for i, b := range dl.Bankers {
		//		banker := b
		//		uID, _, _, _ := banker.GetPlayerBasic()
		//		if uID == botId {
		//			log.Debug("去掉庄家")
		//			dl.Bankers = append(dl.Bankers[:i], dl.Bankers[i+1:]...)
		//		}
		//	}
		//}
		//
		//if len(dl.Bankers) <= 1 {
		//	nextB := dl.NextBotBanker()
		//	dl.Bankers = append(dl.Bankers, nextB)
		//	dl.Bots = append(dl.Bots, &nextB)
		//} else if len(dl.Bankers) >= 4 {
		//	var botId uint32
		//	for _, b := range dl.Bots {
		//		if b.botType == constant.BTNextBanker {
		//			botId = b.UserID
		//			break
		//		}
		//	}
		//
		//	for i, b := range dl.Bankers {
		//		banker := b
		//		uID, _, _, _ := banker.GetPlayerBasic()
		//		if uID == botId {
		//			dl.Bankers = append(dl.Bankers[:i], dl.Bankers[i+1:]...)
		//		}
		//	}
		//}

		//for _, b := range dl.Bots {
		//	if b.botType == constant.BTNextBanker {
		//		dl.Bankers = append(dl.Bankers, b)
		//	}
		//}

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
		if uBet > 0 {
			win := dl.UserBets[user.UserID][dl.res] * constant.AreaX[dl.res]
			result := win - user.DownBetTotal
			if result > 0 {
				user.winCount++
				user.betAmount += uBet
			} else {
				user.betAmount += uBet
				if user.winCount > 10 {
					user.winCount--
				}
			}
		}
		return true
	})
}

func (dl *Dealer) ClearData() {
	// 清理
	dl.AreaBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	dl.DownBetArea = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i := range dl.UserBets {
		dl.UserBets[i] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	dl.AreaBotBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

	dl.Users.Range(func(key, value interface{}) bool {
		// 续投 投注 累加到 auto bet record
		user := value.(*User)
		u := key.(uint32)
		if dl.UserAutoBet[u] == true && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = append(dl.AutoBetRecord[u], dl.UserBetsDetail[u]...)
			// 投注 不续投 覆盖掉
		} else if dl.UserAutoBet[u] == false && len(dl.UserBetsDetail[u]) != 0 {
			dl.AutoBetRecord[u] = dl.UserBetsDetail[u]
			// 续投 不投注 不变
			// 不续投 不投注 不变
		}

		user.IsAction = false
		user.DownBetTotal = 0
		user.LockMoney = 0
		dl.TotalDownMoney = 0
		dl.UserAutoBet[u] = false
		dl.UserIsDownBet[u] = false
		return true
	})

	// 清空投注详情记录
	dl.UserBetsDetail = map[uint32][]msg.Bet{}
	dl.UserLeave = []uint32{}

	dl.res = 0
	dl.bankerWin = 0
	dl.bankerRound += 1

}

func SetPackageTaxM(packageT uint16, tax float64) {
	packageTax[packageT] = tax
}

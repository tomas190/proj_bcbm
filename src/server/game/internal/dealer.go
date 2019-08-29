package internal

import (
	"github.com/name5566/leaf/log"
	"math/rand"
	"proj_bcbm/src/server/constant"
	con "proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"reflect"
	"time"
)

// Mgr <--> Dealer <--> C2C

type Dealer struct {
	*Room
	clock   *time.Ticker
	counter uint32
	ddl     uint32

	Status    uint32        // 房间状态
	res       uint32        // 最新开奖结果
	bankerWin float64       // 庄家输赢
	History   []uint32      // 房间开奖历史
	HRChan    chan HRMsg    // 房间大厅通信
	BBChan    chan struct{} // 下注广播信号

	Users    map[uint32]*User     // 房间用户
	Bankers  []User               // 上庄玩家榜单 todo 玩家榜单
	UserBets map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
	AreaBets []float64            // 每个区域玩家投注总数
}

func NewDealer(rID uint32, hr chan HRMsg) *Dealer {
	return &Dealer{
		Users:    make(map[uint32]*User),
		Room:     NewRoom(rID, con.RL1MinBet, con.RL1MaxBet, con.RL1MinLimit),
		clock:    time.NewTicker(time.Second),
		HRChan:   hr,
		UserBets: map[uint32][]float64{},
		AreaBets: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

// 15-开始下注-0 停止下注  下注
// 23 22 21 - 跑马灯 - 随便什么时候开奖 显示 4 3 2 1 0  开奖
// 2 1 0 清空筹码
// 重置表
func (dl *Dealer) ClockReset(duration uint32, next func()) {
	defer func() { dl.counter = 0 }()
	log.Debug("Deadline: %v, Event: %v, RoomID: %+v", duration, util.Function{}.GetFunctionName(next), dl.RoomID)
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
	log.Debug("bet... %+v", dl.RoomID)

	dl.ddl = uint32(time.Now().Unix()) + con.BetTime
	converter := DTOConverter{}
	resp := converter.RSBMsg(0, 0, *dl)
	dl.Broadcast(&resp)
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
	dl.bankerWin = math.SumSliceFloat64(dl.AreaBets) - con.AreaX[dl.res]*dl.AreaBets[dl.res]

	log.Debug("settle... %+v", dl.RoomID)

	dl.ddl = uint32(time.Now().Unix()) + con.SettleTime
	converter := DTOConverter{}
	for _, u := range dl.Users {
		uWin := dl.UserBets[u.UserID][dl.res]*con.AreaX[dl.res] - math.SumSliceFloat64(dl.UserBets[u.UserID])
		if uWin > 0 {
			c4c.UserWinScore(u.UserID, uWin, func(data *User) {
				resp := converter.RSBMsg(uWin, data.Balance, *dl)
				u.ConnAgent.WriteMsg(&resp)
			})
		} else {
			resp := converter.RSBMsg(uWin, u.Balance, *dl)
			u.ConnAgent.WriteMsg(&resp)
		}
	}

	dl.ClockReset(constant.SettleTime, dl.ClearChip)
}

// 清理筹码
func (dl *Dealer) ClearChip() {
	dl.Status = constant.RSClear
	log.Debug("clear chip... %+v", dl.RoomID)

	// 清理
	dl.AreaBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i := range dl.UserBets {
		dl.UserBets[i] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	dl.res = 0
	dl.bankerWin = 0

	dl.ddl = uint32(time.Now().Unix()) + con.ClearTime

	converter := DTOConverter{}
	resp := converter.RSBMsg(0, 0, *dl)
	dl.Broadcast(&resp)

	dl.ClockReset(constant.ClearTime, dl.Bet)
}

func (dl *Dealer) Broadcast(m interface{}) {
	log.Debug("room brd %+v, content: %+v", reflect.TypeOf(m), m)
	for _, u := range dl.Users {
		u.ConnAgent.WriteMsg(m)
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

	//if prob >= 0 && prob <= 2 {
	//	area = constant.Area40x
	//} else if prob <= 6 {
	//	area = constant.Area30x
	//} else if prob <= 12 {
	//	area = constant.Area20x
	//} else if prob <= 24 {
	//	area = constant.Area10x
	//} else if prob <= 48 {
	//	area = constant.Area5x1
	//} else if prob <= 72 {
	//	area = constant.Area5x2
	//} else if prob <= 96 {
	//	area = constant.Area5x3
	//} else if prob <= 120 {
	//	area = constant.Area5x4
	//}

	if prob >= 0 && prob <= 20 {
		area = constant.Area40x
	} else if prob <= 40 {
		area = constant.Area30x
	} else if prob <= 60 {
		area = constant.Area20x
	} else if prob <= 80 {
		area = constant.Area10x
	} else if prob <= 100 {
		area = constant.Area5x1
	} else if prob <= 105 {
		area = constant.Area5x2
	} else if prob <= 110 {
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

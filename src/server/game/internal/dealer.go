package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"proj_bcbm/src/server/constant"
	con "proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"time"
)

// Mgr <--> Dealer <--> C2C

type Dealer struct {
	*Room

	clock   *time.Ticker
	counter uint32

	Status       uint32

	History      []uint32
	HisStatistic []uint32

	UserBets map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
}

func NewDealer(rID uint32) *Dealer {
	return &Dealer{
		Room: NewRoom(rID, con.RL1MinBet, con.RL1MaxBet, con.RL1MinLimit),
		clock:time.NewTicker(time.Second),
	}
}

// 15-开始下注-0 停止下注  下注
// 23 22 21 - 跑马灯 - 随便什么时候开奖 显示 4 3 2 1 0  开奖
// 2 1 0 清空筹码
// 重置表
func (dl *Dealer) ClockReset(duration uint32, next func()) {
	defer func() {dl.counter = 0}()

	log.Debug("Deadline: %v, Event: %v, RoomID: %+v", duration, util.Function{}.GetFunctionName(next), dl.RoomID)
	go func() {
		for t := range dl.clock.C {
			// log.Debug("时间滴答：%v", t)
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
	// 广播开始下注
	dl.Status = constant.RSBetting
	dl.ClockReset(constant.BetTime, dl.Lottery)
}


func (dl *Dealer) Lottery() {
	dl.ClockReset(constant.LotteryTime, dl.Settle)
	fmt.Printf("#################房间%+v 开奖结果 %+v \n", dl.RoomID, dl.profitPoolLottery())
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
	time.Sleep(x*time.Nanosecond)
	prob := rand.Intn(121) // [0, 121)
	var area uint32

	if prob >= 0 && prob <= 2 {
		area = constant.AreaBenzGolden
	} else if prob <= 6 {
		area = constant.AreaBMWGolden
	} else if prob <= 12 {
		area = constant.AreaAudiGolden
	} else if prob <= 24 {
		area = constant.AreaVWGolden
	} else if prob <= 48 {
		area = constant.AreaBenz
	} else if prob <= 72 {
		area = constant.AreaBMW
	} else if prob <= 96 {
		area = constant.AreaAudi
	} else if prob <= 120 {
		area = constant.AreaVW
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

// 开奖

// 有一个用于房间和大厅之间通信的接收通道，房间产生结果后发送给大厅，
// 大厅监听通道，如果有的话，然后大厅广播所有开奖结果

// 大厅保存历史数据、做历史统计 当发送来房间的数据的时候
// 当新用户来的时候把大厅数据给新用户
// 当新状态改变的时候广播消息给所有用户

// 房间怎么在goroutine上运行

// 发送状态改变

// 结算
func (dl *Dealer) Settle() {
	// 庄家赢数 = Sum(未中奖倍数*未中奖筹码数) - 中奖倍数*中奖筹码数
	log.Debug("结束开奖，结算 %+v", dl.RoomID)
	dl.ClockReset(constant.ClearTime, dl.Lottery)
}


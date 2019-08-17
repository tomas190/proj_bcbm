package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"time"
)

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`       // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"`   // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`         // 用户头像
	Balance   float64    `bson:"balance"json:"money"`          // 用户金额
	ConnAgent gate.Agent `bson:"conn_agent" json:"conn_agent"` // 网络连接代理
}

type Hall struct {
	Statistic []uint32 // 历史记录统计
	History   []uint32 // 历史记录
}

// 开赌场 初始化的时候直接开6个房间然后跑在不同的goroutine上
// 大厅和房间之间通过channel通信
func (h *Hall) OpenCasino() {
	for i := 0; i < constant.RoomCount; i++ {
		go h.openRoom()
	}
}

// 大厅开房
func (h *Hall) openRoom() {

}

// 大厅广播
func (h *Hall) BroadCast() {

}

// 大厅事件 进入房间
type Room struct {
	RoomID       uint32
	MinBet       float64
	MaxBet       float64
	MinLimit     float64
	Status       uint32
	EndTime      uint32 // fixme
	History      []uint32
	HisStatistic []uint32

	UserBets map[uint32][]float64 // 用户投注信息，在8个区域分别投了多少
}

type roomStatus struct {
	Status  uint32
	EndTime uint32
	Result  uint32
}

var roomStatusChan chan roomStatus

// 有一个用于房间和大厅之间通信的接收通道，房间产生结果后发送给大厅，
// 大厅监听通道，如果有的话，然后大厅广播所有开奖结果

// 大厅保存历史数据、做历史统计 当发送来房间的数据的时候
// 当新用户来的时候把大厅数据给新用户
// 当新状态改变的时候广播消息给所有用户

// 房间怎么在goroutine上运行

// 发送状态改变

// 结算
func (r *Room) Settle() {
	// 庄家赢数 = Sum(未中奖倍数*未中奖筹码数) - 中奖倍数*中奖筹码数
}

// 开奖

/*

(玩家赢 - 官方庄家和机器人赢)  小于或等于  从盈余池随机拿到的值，则定为本局开奖结果。
如果是 (玩家赢 - 官方庄家和机器人赢) > 从盈余池随机拿到的值，则重新获取开奖结果，直到 小于或等于

*/

func (r *Room) ProfitPoolLottery() uint32 {
	// 盈余池 随机从10%到50%取一个值，算出一个预计赔付数
	randomUtil := util.Random{}
	profitPoolRatePercent := randomUtil.RandInRange(constant.ProfitPoolMinPercent, constant.ProfitPoolMaxPercent)
	profitPoolRate := float64(profitPoolRatePercent) / 100.0
	acceptableMaxLose := profitPool() * profitPoolRate

	fmt.Println("最大可接受赔付", acceptableMaxLose)

	var area uint32
	for i := 0; i < 100; i++ {
		preArea := r.fairLottery()
		preLoseAmount := preUserWin(r.UserBets, preArea)
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

// todo 还需加入押注信息
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

// 公平开奖
func (r *Room) fairLottery() uint32 {
	rand.Seed(time.Now().UnixNano())
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

// 大厅监测各房间发来的消息，如果变化发出房间状态变化广播

// 房间

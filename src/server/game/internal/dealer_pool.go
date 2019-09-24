package internal

import (
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"time"
)

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

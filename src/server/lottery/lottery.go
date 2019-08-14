package lottery

import (
	"fmt"
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"time"
)

type Lottery struct {}
type RoomBetInfo struct {}

/*

(玩家赢 - 官方庄家和机器人赢)  小于或等于  从盈余池随机拿到的值，则定为本局开奖结果。
如果是 (玩家赢 - 官方庄家和机器人赢) > 从盈余池随机拿到的值，则重新获取开奖结果，直到 小于或等于

*/

func (l *Lottery) ProfitPoolLottery() uint32 {
	// 盈余池 随机从10%到50%取一个值，算出一个预计赔付数
	randomUtil := util.Random{}
	profitPoolRatePercent := randomUtil.RandInRange(constant.ProfitPoolMinPercent, constant.ProfitPoolMaxPercent)
	profitPoolRate := float64(profitPoolRatePercent)/100.0
	acceptableMaxLose := profitPool()*profitPoolRate
	fmt.Println("最大可接受赔付", acceptableMaxLose)

	var area uint32
	for i := 0; i < 100; i++ {
		preArea := l.fairLottery()
		preLoseAmount := preUserWin(RoomBetInfo{}, preArea)
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
func preUserWin(betInfo RoomBetInfo, preArea uint32) float64 {
	return 5
}

// 盈余池 = 玩家总输 - 玩家总赢 * 杀数 - (玩家数量 * 6)
// todo 统计计算玩家总赢和玩家总输、玩家数量
func profitPool() float64 {
	// return pTotalLose - pTotalWin * constant.HouseEdgePercent - pCount*constant.GiftAmount
	return 20.0
}

// 公平开奖
func (l *Lottery) fairLottery() uint32 {
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



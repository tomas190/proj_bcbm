package internal

import (
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
	"time"
)

// 根据盈余池开奖
func (dl *Dealer) profitPoolLottery() uint32 {

	sur, _ := db.GetSurPool()
	loseRate := sur.PlayerLoseRateAfterSurplusPool * 100
	percentageWin := sur.RandomPercentageAfterWin * 100
	countWin := sur.RandomCountAfterWin
	percentageLose := sur.RandomPercentageAfterLose * 100
	countLose := sur.RandomCountAfterLose
	surplusPool := sur.SurplusPool

	if dl.IsSpecial { // 特殊品牌赢率
		percentageWin = 95
		countWin = 4
		percentageLose = 0
		countLose = 0
	}

	r := util.Random{}
	preArea := dl.fairLottery()
	settle := dl.preUserWin(preArea)
	if settle >= 0 { // 玩家赢钱
		for {
			loseRateNum := r.RandInRange(1, 101)
			percentageWinNum := r.RandInRange(1, 101)
			if countWin > 0 {
				if percentageWinNum > int(percentageWin) { // 盈余池判定
					if surplusPool > settle { // 盈余池足够
						break
					} else {                             // 盈余池不足
						if loseRateNum > int(loseRate) { // 30%玩家赢钱
							break
						} else { // 70%玩家输钱
							for {
								preArea = dl.fairLottery()
								settle = dl.preUserWin(preArea)
								if settle <= 0 {
									break
								}
							}
							break
						}
					}
				} else { // 又随机生成牌型
					preArea = dl.fairLottery()
					settle = dl.preUserWin(preArea)
					if settle > 0 { // 玩家赢
						countWin--
					} else {
						break
					}
				}
			} else {
				// 盈余池判定
				if surplusPool > settle { // 盈余池足够
					break
				} else {                             // 盈余池不足
					if loseRateNum > int(loseRate) { // 30%玩家赢钱
						break
					} else { // 70%玩家输钱
						for {
							preArea = dl.fairLottery()
							settle = dl.preUserWin(preArea)
							if settle <= 0 {
								break
							}
						}
						break
					}
				}
			}
		}
	} else { // 玩家输钱
		for {
			loseRateNum := r.RandInRange(1, 101)
			percentageLoseNum := r.RandInRange(1, 101)
			if countLose > 0 {
				if percentageLoseNum > int(percentageLose) {
					break
				} else { // 又随机生成牌型
					preArea = dl.fairLottery()
					settle = dl.preUserWin(preArea)
					if settle > 0 { // 玩家赢
						// 盈余池判定
						if surplusPool > settle { // 盈余池足够
							break
						} else {                             // 盈余池不足
							if loseRateNum > int(loseRate) { // 30%玩家赢钱
								for {
									preArea = dl.fairLottery()
									settle = dl.preUserWin(preArea)
									if settle >= 0 {
										break
									}
								}
								break
							} else { // 70%玩家输钱
								for {
									preArea = dl.fairLottery()
									settle = dl.preUserWin(preArea)
									if settle <= 0 {
										break
									}
								}
								break
							}
						}
					} else {
						countLose--
					}
				}
			} else { // 玩家输钱
				for {
					preArea = dl.fairLottery()
					settle = dl.preUserWin(preArea)
					if settle <= 0 {
						break
					}
				}
				break
			}
		}
	}
	return preArea
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
// userBets 玩家投注
// preArea 预开奖区域
func (dl *Dealer) preUserWin(preArea uint32) float64 {
	// 玩家投注开奖
	userWin := dl.DownBetArea[preArea] * constant.AreaX[preArea]
	return userWin - dl.TotalDownMoney
}

// 盈余池 = 玩家总输 - 玩家总赢 * 杀数 - (玩家数量 * 6)
func (dl *Dealer) profitPool() float64 {
	// 需要数据库

	pp := db.RProfitPool()
	return pp.Profit
}

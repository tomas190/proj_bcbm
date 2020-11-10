package internal

import (
	"fmt"
	"github.com/name5566/leaf/module"
	"github.com/patrickmn/go-cache"
	_ "net/http/pprof"
	"proj_bcbm/src/server/base"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c *Client4Center // 连接中心服的客户端
	db  *MgoC          // 数据库客户端
	ca  *cache.Cache   // 内存缓存
	Mgr = NewHall()
)

type Module struct {
	*module.Skeleton
}

// 模块初始化
func (m *Module) OnInit() {
	m.Skeleton = skeleton


	var totalWinNum int
	var totalLoseNum int
	var totalBetWin float64
	var totalBetLose float64

	var surplusPool float64 = 10000000

	for i := 0; i < 10000; i++ {
		fmt.Println(i)

		r := util.Random{}
		dl := Dealer{}
		dl.AreaBets = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
		dl.DownBetArea = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

		loseRate := 70
		//1: "GoldenBenz-40X",
		//2: "GoldenBMW-30X",
		//3: "GoldenAudi-20X",
		//4: "GoldenVW-10X",
		//5: "Benz-5X",
		//6: "BMW-5X",
		//7: "Audi-5X",
		//8: "VW-5X",

		//机器人下注
		//for i := 0; i < 1; i++ {
		//	chip, area := dl.randBet()
		//	cs := con.ChipSize[chip]
		//	dl.TotalDownMoney = cs
		//	dl.DownBetArea[area] = cs
		//	dl.AreaBets[area] = cs
		//}

		//玩家下注
		var downPot int = 8
		dl.TotalDownMoney = 10
		dl.DownBetArea[downPot] = dl.TotalDownMoney

		percentageWin := 0
		countWin := 0
		percentageLose := 100
		countLose := 100

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

		settle = dl.preUserWin(preArea)
		fmt.Println("玩家当局输赢:", settle)
		if settle > 0 {
			totalWinNum += 1
			totalBetWin += settle
			surplusPool -= settle
		} else {
			totalLoseNum += 1
			totalBetLose -= settle
			surplusPool -= settle
		}

		//税前庄家输赢
		//math := util.Math{}
		//preBankerWin, _ := math.SumSliceFloat64(dl.AreaBets).Sub(math.MultiFloat64(con.AreaX[preArea], dl.AreaBets[preArea])).Float64()
		//fmt.Println("庄家当局输赢:", preBankerWin)
		//if preBankerWin > 0 {
		//	totalWinNum += 1
		//	totalBetWin += preBankerWin
		//	surplusPool -= settle
		//} else {
		//	totalLoseNum += 1
		//	totalBetLose -= preBankerWin
		//	surplusPool -= settle
		//}
	}

	taxWinMoney := totalBetWin - (totalBetWin * 0.05)
	fmt.Println("玩家盈余金额为:", int64(surplusPool))
	fmt.Println("玩家总赢局:", totalWinNum)
	fmt.Println("玩家总输局:", totalLoseNum)
	fmt.Println("玩家总输:", int64(totalBetLose))
	fmt.Println("玩家总赢:", int64(totalBetWin))
	fmt.Println("玩家总赢(税后):", int64(taxWinMoney))
	fmt.Println("总流水:", int64(totalBetWin+totalBetLose))


	// 中心服务器
	c4c = NewClient4Center()
	//c4c.ReqToken()
	c4c.HeartBeatAndListen()
	//c4c.CronUpdateToken()

	// 数据库
	db = NewMgoC(conf.Server.MongoDB)
	err := db.Init()
	if err != nil {
		log.Error("数据库初始化错误 %+v", err)
	}

	go StartHttpServer()

	winChan = make(chan bool)
	loseChan = make(chan bool)
	downBankerChan = make(chan bool)
	// 缓存
	ca = cache.New(5*time.Minute, 10*time.Minute)

	// 游戏大厅
	Mgr.OpenCasino()

	// net/http/pprof 已经在 init()函数中通过 import 副作用完成默认 Handler 的注册
	//go func() {
	//	err := http.ListenAndServe("localhost:6060", nil)
	//	if err != nil {
	//		log.Debug("性能分析服务启动错误...")
	//	}
	//	log.Debug("性能分析服务...")
	//}()
}

// 模块销毁
func (m *Module) OnDestroy() {
	log.Debug("game模块被销毁...")
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)
}

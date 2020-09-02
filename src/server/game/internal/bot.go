package internal

import (
	"fmt"
	"math"
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

func (dl *Dealer) AddBots() {
	betGod := dl.BetGod()
	dl.Bots = append(dl.Bots, &betGod)

	timeNow := time.Now().Hour()
	var handleNum int
	switch timeNow {
	case 1:
		handleNum = 75
		break
	case 2:
		handleNum = 68
		break
	case 3:
		handleNum = 60
		break
	case 4:
		handleNum = 51
		break
	case 5:
		handleNum = 41
		break
	case 6:
		handleNum = 30
		break
	case 7:
		handleNum = 17
		break
	case 8:
		handleNum = 15
		break
	case 9:
		handleNum = 17
		break
	case 10:
		handleNum = 30
		break
	case 11:
		handleNum = 41
		break
	case 12:
		handleNum = 51
		break
	case 13:
		handleNum = 60
		break
	case 14:
		handleNum = 68
		break
	case 15:
		handleNum = 75
		break
	case 16:
		handleNum = 80
		break
	case 17:
		handleNum = 84
		break
	case 18:
		handleNum = 87
		break
	case 19:
		handleNum = 89
		break
	case 20:
		handleNum = 90
		break
	case 21:
		handleNum = 89
		break
	case 22:
		handleNum = 87
		break
	case 23:
		handleNum = 84
		break
	case 24:
		handleNum = 80
		break
	case 0:
		handleNum = 80
		break
	}
	r := util.Random{}

	var randNum int

	var minP int
	var maxP int
	getNum := float64(handleNum) * 0.2
	maNum := math.Floor(getNum)
	minP = handleNum - int(maNum)
	maxP = handleNum + int(maNum)

	RNum := []float64{0.1, 0.11, 0.12, 0.13, 0.14, 0.15}
	rand.Seed(time.Now().UnixNano())
	rn := rand.Intn(len(RNum))
	rNNum := float64(handleNum) * RNum[rn]
	RNNNum := math.Floor(rNNum)
	randNum = int(RNNNum)

	var tn int
	for {
		n := r.RandInRange(0, len(dl.Bots))
		if dl.Bots[n] != nil && dl.Bots[n].botType != constant.BTNextBanker {
			dl.Bots[n].UserID = uint32(100000000 + r.RandInRange(0, 200000000))
			dl.Bots[n].NickName = fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))
			dl.Bots[n].Balance = float64(0+r.RandInRange(200, 4600)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
			dl.Bots[n].Avatar = fmt.Sprintf("%+v", r.RandInRange(1, 21)) + ".png"
			dl.Bots[n].WinCount = uint32(r.RandInRange(0, 3)) // 获胜局数
			dl.Bots[n].BetAmount = float64(r.RandInRange(20, 500))
		}
		time.Sleep(time.Millisecond)
		tn++
		if tn == randNum {
			log.Debug("修改%v:个机器人", randNum)
			break
		}
	}

	num := r.RandInRange(0, 100)
	num2 := r.RandInRange(3, 8)
	if num >= 0 && num < 50 {
		var n int
		for {
			richMan := dl.RichMan()
			dl.Bots = append(dl.Bots, &richMan)
			time.Sleep(time.Millisecond)
			n++
			if n == num2 {
				log.Debug("添加%v:个机器人", num2)
				break
			}
		}

	} else if num >= 50 && num < 100 {
		var n int
		for k, v := range dl.Bots {
			if v != nil && v.botType != constant.BTNextBanker {
				dl.Bots = append(dl.Bots[:k], dl.Bots[k+1:]...)
				time.Sleep(time.Millisecond)
				n++
				if n == num2 {
					log.Debug("减去%v:个机器人", num2)
					break
				}
			}
		}
	}

	robotNum := len(dl.Bots)
	log.Debug("机器人当前数量:%v,最小范围:%v.最大范围:%v", robotNum, minP, maxP)

	if robotNum < minP { // 加
		for {
			richMan := dl.RichMan()
			dl.Bots = append(dl.Bots, &richMan)
			time.Sleep(time.Millisecond)
			robotNum = len(dl.Bots)
			if robotNum == minP {
				log.Debug("房间:%v,加机器人数量:%v", dl.RoomID, len(dl.Bots))
				break
			}
		}
	} else if robotNum > maxP { // 减
		for k, v := range dl.Bots {
			if v != nil && v.botType != constant.BTNextBanker {
				dl.Bots = append(dl.Bots[:k], dl.Bots[k+1:]...)
				time.Sleep(time.Millisecond)
				robotNum =  len(dl.Bots)
				if robotNum == maxP {
					log.Debug("房间:%v,减机器人数量:%v", dl.RoomID, len(dl.Bots))
					break
				}
			}
		}
	}
}

// 机器人下注，随机下注后把结果赋值到下注结果列表中
func (dl *Dealer) BotsBet() {
	//ru := util.Random{}
	//chipCount := ru.RandInRange(55, 65)
	time.Sleep(time.Second * 1)
	//counter := 0
	dl.IsDownBet = true
	for i := 0; i < 100; i++ {
		if dl.IsDownBet == false {
			return
		}
		//counter++
		//delay := (30 - counter/2) * (30 - counter/2)
		//time.Sleep(time.Millisecond * time.Duration(rand.Intn(delay+5)))
		timerSlice := []int32{50, 150, 20, 300, 800, 30, 500}
		rand.Seed(time.Now().UnixNano())
		num2 := rand.Intn(len(timerSlice))
		time.Sleep(time.Millisecond * time.Duration(timerSlice[num2]))

		chip, area := dl.randBet()
		cs := constant.ChipSize[chip]

		// 限红
		if dl.roomBonusLimit(area) < cs || dl.dynamicBonusLimit(area) < cs {
			//log.Debug("<<===== 机器人下注结束 =====>>")
			continue
		}
		// 区域所有玩家投注总数
		dl.AreaBets[area] = dl.AreaBets[area] + cs
		// 区域机器人投注总数
		dl.AreaBotBets[area] = dl.AreaBotBets[area] + cs

		resp := &msg.BetInfoB{
			Area:      area,
			Chip:      chip,
			AreaTotal: dl.AreaBets[area],
		}

		dl.Broadcast(resp)
	}
}

func (dl *Dealer) randBet() (uint32, uint32) {
	var chip uint32
	var area uint32

	ru := util.Random{}

	areaProb := ru.RandInRange(0, 100)
	if areaProb >= 0 && areaProb <= 78 {
		area = uint32(ru.RandInRange(4, 8) + 1)
	} else if areaProb > 78 && areaProb <= 88 {
		area = 4
	} else if areaProb > 88 && areaProb <= 94 {
		area = 3
	} else if areaProb > 94 && areaProb <= 97 {
		area = 2
	} else if areaProb > 97 && areaProb <= 100 {
		area = 1
	}

	//获取一个随机数值，然后根据随机数值的区间来进行随机下注筹码
	chipProb := ru.RandInRange(0, 100)

	if chipProb >= 0 && chipProb <= 67 {
		chip = 1
	} else if chipProb > 67 && chipProb <= 88 {
		chip = 2
	} else if chipProb > 88 && chipProb <= 95 {
		chip = 3
	} else if chipProb > 95 && chipProb <= 98 {
		chip = 4
	} else if chipProb > 98 && chipProb <= 100 {
		chip = 5
	}

	return chip, area
}

func (dl *Dealer) BetGod() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(4, 5))                                                // 获胜局数
	BetAmount := float64(r.RandInRange(20, 500))                                           // 下注金额
	Balance := float64(0+r.RandInRange(200, 4600)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
	UserID := uint32(100000000 + r.RandInRange(0, 200000000))                              // 用户ID
	avatar := fmt.Sprintf("%+v", r.RandInRange(1, 21)) + ".png"
	randNum := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))

	betGod := Bot{
		UserID:    UserID,
		NickName:  randNum,
		Avatar:    avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTBetGod,
	}

	return betGod
}

func (dl *Dealer) RichMan() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))                                                // 获胜局数
	BetAmount := float64(r.RandInRange(20, 500))                                           // 下注金额
	Balance := float64(0+r.RandInRange(200, 4600)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
	UserID := uint32(100000000 + r.RandInRange(0, 200000000))                              // 用户ID
	avatar := fmt.Sprintf("%+v", r.RandInRange(1, 21)) + ".png"

	richMan := Bot{
		UserID: UserID,
		// NickName:  "richMan",
		Avatar:    avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTRichMan,
	}

	return richMan
}

func (dl *Dealer) NextBotBanker() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))                   // 获胜局数
	BetAmount := float64(r.RandInRange(20, 500))              // 下注金额
	Balance := float64(0 + r.RandInRange(20000, 40000))       // 金币数
	UserID := uint32(100000000 + r.RandInRange(0, 200000000)) // 用户ID
	avatar := fmt.Sprintf("%+v", r.RandInRange(1, 21)) + ".png"
	randNum := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))

	nextBanker := Bot{
		UserID:    UserID,
		NickName:  randNum,
		Avatar:    avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTNextBanker,
	}

	return nextBanker
}

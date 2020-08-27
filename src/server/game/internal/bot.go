package internal

import (
	"fmt"
	"math"
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

func (dl *Dealer) AddBots() {
	robotNum := len(dl.Bots)
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
	}

	var randNum int
	slice := []int32{1, 2, 1, 2} // 1为-,2为+
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(slice))
	var maNum float64
	if slice[num] == 1 {
		getNum := handleNum / 10
		maNum = math.Floor(float64(getNum))
		handleNum -= int(maNum)
		randNum = int(maNum)
	} else if slice[num] == 2 {
		getNum := handleNum / 10
		maNum = math.Floor(float64(getNum))
		RNum := float64(handleNum) * 0.25
		RNNum := math.Floor(RNum)
		handleNum += int(maNum)
		randNum = int(RNNum)
	}

	if robotNum < handleNum { // 加
		for {
			richMan := dl.RichMan()
			dl.Bots = append(dl.Bots, &richMan)
			robotNum = len(dl.Bots)
			if robotNum >= handleNum {
				break
			}
		}
	}
	if robotNum > handleNum { // 减
		for k, v := range dl.Bots {
			if v != nil {
				dl.Bots = append(dl.Bots[:k], dl.Bots[k+1:]...)
				robotNum = len(dl.Bots)
				if robotNum <= handleNum {
					break
				}
			}
		}
	}

	var num2 int
	for _, v := range dl.Bots {
		if v != nil {
			r := util.Random{}
			v.UserID = uint32(100000000 + r.RandInRange(0, 200000000))
			v.Balance = float64(0+r.RandInRange(200, 4600)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
			v.BetAmount = float64(r.RandInRange(20, 500))
			num2++
			if num2 >= randNum {
				return
			}
		}
	}

	betGod := dl.BetGod()
	nextBankerBot := dl.NextBotBanker()
	dl.Bots = append(dl.Bots, &betGod, &nextBankerBot)
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

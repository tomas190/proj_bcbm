package internal

import (
	"fmt"
	"math/rand"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

func (dl *Dealer) AddBots() {
	betGod := dl.BetGod()
	nextBankerBot := dl.NextBotBanker()
	dl.Bots = append(dl.Bots, &betGod, &nextBankerBot)

	for i := 0; i < 30; i++ {
		richMan := dl.RichMan()
		dl.Bots = append(dl.Bots, &richMan)
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
		rand.Seed(int64(time.Now().UnixNano()))
		num2 := rand.Intn(len(timerSlice))
		time.Sleep(time.Millisecond * time.Duration(timerSlice[num2]))

		chip, area := dl.randBet()
		cs := constant.ChipSize[chip]

		// 限红
		if dl.roomBonusLimit(area) < cs || dl.dynamicBonusLimit(area) < cs {
			log.Debug("<<===== 机器人下注结束 =====>>")
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
	if areaProb >= 0 && areaProb <= 80 {
		area = uint32(ru.RandInRange(4, 8) + 1)
	} else {
		area = uint32(ru.RandInRange(0, 4) + 1)
	}

	//获取一个随机数值，然后根据随机数值的区间来进行随机下注筹码
	chipProb := ru.RandInRange(0, 100)

	if chipProb >= 0 && chipProb <= 50 {
		chip = 1
	} else if chipProb > 50 && chipProb <= 70 {
		chip = 2
	} else if chipProb > 70 && chipProb <= 80 {
		chip = 3
	} else if chipProb > 80 && chipProb < 95 {
		chip = 2
	} else {
		chip = 3
	}

	return chip, area
}

func (dl *Dealer) BetGod() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(4, 5))                                                // 获胜局数
	BetAmount := float64(r.RandInRange(80, 450))                                           // 下注金额
	Balance := float64(0+r.RandInRange(200, 1888)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
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
	BetAmount := float64(r.RandInRange(80, 450))                                           // 下注金额
	Balance := float64(0+r.RandInRange(200, 1888)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
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
	BetAmount := float64(r.RandInRange(80, 450))              // 下注金额
	Balance := float64(0 + r.RandInRange(6000, 10000))        // 金币数
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

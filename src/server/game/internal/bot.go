package internal

import (
	"fmt"
	"math/rand"
	"proj_bcbm/src/server/constant"
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
	ru := util.Random{}
	chipCount := ru.RandInRange(55, 65)
	time.Sleep(time.Second * 1)
	counter := 0
	for i := 0; i < chipCount; i++ {
		counter++
		delay := (30 - counter/2) * (30 - counter/2)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(delay+5)))

		chip, area := dl.randBet()
		cs := constant.ChipSize[chip]

		// 限红
		if dl.roomBonusLimit(area) < cs || dl.dynamicBonusLimit(area) < cs {
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
	if areaProb >= 0 && areaProb <= 90 {
		area = uint32(ru.RandInRange(4, 8) + 1)
	} else {
		area = uint32(ru.RandInRange(0, 4) + 1)
	}

	chipProb := ru.RandInRange(0, 100)

	if chipProb >= 0 && chipProb <= 50 {
		chip = 1
	} else if chipProb > 50 && chipProb <= 70 {
		chip = 2
	} else if chipProb > 70 && chipProb <= 80 {
		chip = 3
	} else if chipProb > 80 && chipProb < 95 {
		chip = 4
	} else {
		chip = 5
	}

	return chip, area
}

func (dl *Dealer) BetGod() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(4, 5))                                                   // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))                                            // 下注金额
	Balance := float64(20000+r.RandInRange(0, 20000)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000))                                     // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	betGod := Bot{
		UserID:    UserID,
		NickName:  "betGod",
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTBetGod,
	}

	return betGod
}

func (dl *Dealer) RichMan() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))                                                   // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))                                            // 下注金额
	Balance := float64(20000+r.RandInRange(0, 20000)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000))                                     // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	richMan := Bot{
		UserID:    UserID,
		NickName:  "richMan",
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTRichMan,
	}

	return richMan
}

func (dl *Dealer) NextBotBanker() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))                                                   // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))                                            // 下注金额
	Balance := float64(50000+r.RandInRange(0, 20000)) + float64(r.RandInRange(50, 100))/100.0 // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000))                                     // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	nextBanker := Bot{
		UserID:    UserID,
		NickName:  "nextBanker" + fmt.Sprintf("%+v", UserID)[:2],
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
		botType:   constant.BTNextBanker,
	}

	return nextBanker
}

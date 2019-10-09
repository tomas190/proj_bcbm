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
	chipCount := ru.RandInRange(100, 150)
	time.Sleep(time.Second * 1)
	for i := 0; i < chipCount; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(150)))
		area := uint32(rand.Intn(8) + 1)
		var chip uint32

		prob := ru.RandInRange(0, 100)

		if prob >= 0 && prob <= 50 {
			chip = 1
		} else if prob > 50 && prob <= 70 {
			chip = 2
		} else if prob > 70 && prob <= 80 {
			chip = 3
		} else if prob > 80 && prob < 95 {
			chip = 4
		} else {
			chip = 5
		}

		cs := constant.ChipSize[chip]

		// 限红
		if dl.roomBonusLimit(area) < cs || dl.dynamicBonusLimit(area) < cs {
			return
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

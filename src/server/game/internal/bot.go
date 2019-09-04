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
	time.Sleep(time.Second * 1)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
		area := uint32(rand.Intn(8) + 1)
		chip := uint32(rand.Intn(5) + 1)

		cs := constant.ChipSize[chip]
		// 区域所有玩家投注总数
		dl.AreaBets[area] = dl.AreaBets[area] + cs
		// 区域机器人投注总数
		dl.AreaBotBets[area] = dl.AreaBotBets[area] + cs

		resp := &msg.BetInfoB{
			Area:        area,
			Chip:        chip,
			AreaTotal:   dl.AreaBets[area],
			PlayerTotal: 0,
			PlayerID:    0, // todo
			Money:       0,
		}

		dl.Broadcast(resp)
	}

}

// 若无真人玩家上庄随机一个机器人上
// 若有真人玩家上装机器人机器人靠后
func (dl *Dealer) BotsGrabBanker() {

}

// 庄家轮换时修改列表
// 若上庄列表只有一个
// 保留之前上庄的
// 玩家列表中只有一个机器人大于50000，作为上庄
// bot 玩家
// 玩家的近20局统计从数据库中找

func (dl *Dealer) BetGod() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(4, 5))               // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))        // 下注金额
	Balance := float64(20000 + r.RandInRange(0, 20000))   // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000)) // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	betGod := Bot{
		UserID:    UserID,
		NickName:  "betGod",
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
	}

	return betGod
}

func (dl *Dealer) RichMan() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))               // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))        // 下注金额
	Balance := float64(20000 + r.RandInRange(0, 20000))   // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000)) // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	richMan := Bot{
		UserID:    UserID,
		NickName:  "richMan",
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
	}

	return richMan
}

func (dl *Dealer) NextBotBanker() Bot {
	r := util.Random{}
	WinCount := uint32(r.RandInRange(0, 3))               // 获胜局数
	BetAmount := float64(r.RandInRange(800, 5000))        // 下注金额
	Balance := float64(50000 + r.RandInRange(0, 20000))   // 金币数
	UserID := uint32(1000000 + r.RandInRange(0, 2000000)) // 用户ID
	Avatar := "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png"

	nextBanker := Bot{
		UserID:    UserID,
		NickName:  "nextBanker" + fmt.Sprintf("%+v", UserID)[:2],
		Avatar:    Avatar,
		Balance:   Balance,
		WinCount:  WinCount,
		BetAmount: BetAmount,
	}

	return nextBanker
}

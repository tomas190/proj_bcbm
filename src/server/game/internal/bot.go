package internal

import (
	"fmt"
	"proj_bcbm/src/server/util"
)

// 机器人随机投注
// 若无真人玩家上庄随机一个机器人上
// 若有真人玩家上装机器人机器人靠后

// 结算 只需要返回玩家输赢
// 直接在内存里运算？

func (dl *Dealer) AddBots() {

}

func (dl *Dealer) RemoveBots() {

}

func (dl *Dealer) BotsBet() {

}

func (dl *Dealer) BotsSettle() {

}

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

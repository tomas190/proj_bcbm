package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"sync"
)

type User struct {
	Balance     float64      // 用户金额
	BalanceLock sync.RWMutex // 锁
	UserID      uint32       // 用户id
	NickName    string       // 用户昵称
	Avatar      string       // 用户头像
	ConnAgent   gate.Agent   // 网络连接代理
}

type Bot struct {
	UserID    uint32
	NickName  string
	Avatar    string
	Balance   float64
	WinCount  uint32
	BetAmount float64
	botType   uint32
}

type Player interface {
	GetBalance() float64
	GetPlayerBasic() (uint32, string, string, float64)
	GetPlayerAccount() (uint32, float64)
}

func (u User) GetBalance() float64 {
	return u.Balance
}
func (u User) GetPlayerBasic() (uint32, string, string, float64) {
	return u.UserID, u.NickName, u.Avatar, u.Balance
}

// 返回玩家投注了的近20局获胜局数和总下注数
func (u User) GetPlayerAccount() (uint32, float64) {
	// 只记录玩家进入房间之后的，不从库中读取
	//his, err := db.RUserSettle(u.UserID)
	//if err != nil {
	//	log.Debug("获取用户历史数据错误 %+v", err)
	//	return 10, 100
	//}
	//
	//var winCount uint32
	//var totalBet float64
	//for _, sRecord := range his {
	//	sr := sRecord
	//	if sr.IsWin == true {
	//		winCount++
	//	}
	//	totalBet += sr.BetAmount
	//}
	//
	//return winCount, totalBet
	var winCount uint32
	var betAmount float64

	if v1, exist1 := ca.Get(fmt.Sprintf("%+v-betAmount", u.UserID)); exist1 {
		if ba, ok := v1.(float64); ok {
			betAmount = ba
		} else {
			log.Debug("转换投注数错误 error")
		}
	} else {
		log.Debug("查找投注数错误 error")
	}

	if v2, exist2 := ca.Get(fmt.Sprintf("%+v-winCount", u.UserID)); exist2 {
		if wc, ok := v2.(int64); ok {
			winCount = uint32(wc)
		} else {
			log.Debug("转换胜场数错误 error")
		}
	} else {
		log.Debug("查找胜场数错误 error")
	}
	return winCount, betAmount
}

func (b Bot) GetBalance() float64 {
	return b.Balance
}

func (b Bot) GetPlayerBasic() (uint32, string, string, float64) {
	return b.UserID, b.NickName, b.Avatar, b.Balance
}

func (b Bot) GetPlayerAccount() (uint32, float64) {
	return b.WinCount, b.BetAmount
}

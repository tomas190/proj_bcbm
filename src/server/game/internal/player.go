package internal

import (
	"github.com/name5566/leaf/gate"
	"sync"
)

type User struct {
	Balance       float64      // 用户金额
	BalanceLock   sync.RWMutex // 锁
	BankerBalance float64      // 上庄金额
	UserID        uint32       // 用户id
	Status        int          // 上庄状态
	NickName      string       // 用户昵称
	Avatar        string       // 用户头像
	PackageId     uint16
	ConnAgent     gate.Agent // 网络连接代理
	DownBetTotal  float64    // 玩家总下注
	winCount      uint32     // 玩家赢的次数
	betAmount     float64    // 玩家总投注金额
	LockMoney     float64    // 下注锁定的钱
	IsAction      bool       // 玩家是否行动
	LockSucc      int        // 是否锁钱成功
}

func (u *User) Init() {
	u.Balance = 0
	u.BankerBalance = 0
	u.DownBetTotal = 0
	u.winCount = 0
	u.betAmount = 0
	u.IsAction = false
	u.LockSucc = 0
}

type Bot struct {
	UserID        uint32
	NickName      string
	Avatar        string
	Balance       float64
	BankerBalance float64
	WinCount      uint32
	BetAmount     float64
	botType       uint32
	TwentyData    []int
	Status        int
}

type Player interface {
	GetBalance() float64
	GetBankerBalance() float64

	GetPlayerBasic() (uint32, string, string, float64)
	GetPlayerAccount() (uint32, float64)
}

func (u User) GetBalance() float64 {
	return u.Balance
}

func (u User) GetBankerBalance() float64 {
	return u.BankerBalance
}

func (u User) GetPlayerBasic() (uint32, string, string, float64) {
	return u.UserID, u.NickName, u.Avatar, u.Balance
}

// 返回玩家投注了的近20局获胜局数和总下注数
func (u User) GetPlayerAccount() (uint32, float64) {
	return u.winCount, u.betAmount
}

func (b Bot) GetBalance() float64 {
	return b.Balance
}

func (b Bot) GetBankerBalance() float64 {
	return b.Balance
}

func (b Bot) GetPlayerBasic() (uint32, string, string, float64) {
	return b.UserID, b.NickName, b.Avatar, b.Balance
}

func (b Bot) GetPlayerAccount() (uint32, float64) {
	return b.WinCount, b.BetAmount
}

package internal

import "github.com/name5566/leaf/gate"

type Bet struct {
	Area       uint32
	AreaStr    string
	Chip       uint32
	ChipAmount uint32
}

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`       // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"`   // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`         // 用户头像
	Balance   float64    `bson:"balance"json:"money"`          // 用户金额
	ConnAgent gate.Agent `bson:"conn_agent" json:"conn_agent"` // 网络连接代理
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
	GetPlayerBasic() (uint32, string, string, float64)
	GetPlayerAccount() (uint32, float64)
}

func (u User) GetPlayerBasic() (uint32, string, string, float64) {
	return u.UserID, u.NickName, u.Avatar, u.Balance
}

func (u User) GetPlayerAccount() (uint32, float64) {
	// todo
	return 10, 100
}

func (b Bot) GetPlayerBasic() (uint32, string, string, float64) {
	return b.UserID, b.NickName, b.Avatar, b.Balance
}

func (b Bot) GetPlayerAccount() (uint32, float64) {
	return b.WinCount, b.BetAmount
}

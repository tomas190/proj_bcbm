package internal

import "time"

type SettleDB struct {
	User      UserDB  `bson:"User"`
	WinOrder  string  `bson:"WinOrder"`
	RoundID   string  `bson:"RoundID"`
	IsWin     bool    `bson:"IsWin"`
	BetAmount float64 `bson:"BetAmount"`
	WinAmount float64 `bson:"WinAmount"`
}

type UserDB struct {
	UserID   uint32  `bson:"UserID" json:"UserID"`     // 用户id
	NickName string  `bson:"NickName" json:"NickName"` // 用户昵称
	Avatar   string  `bson:"Avatar" json:"Avatar"`     // 用户头像
	Balance  float64 `bson:"Balance" json:"Balance"`   // 用户金额
}

type ProfitDB struct {
	UpdateTime    time.Time `bson:"UpdateTime"`
	UpdateTimeStr string    `bson:"UpdateTimeStr"`
	AllWin        float64   `bson:"AllWin"`
	AllLost       float64   `bson:"AllLost"`
	Profit        float64   `bson:"Profit"`
	PlayerNum     uint32    `bson:"PlayerNum"`
}

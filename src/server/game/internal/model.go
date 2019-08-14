package internal

import (
	"github.com/name5566/leaf/gate"
)

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`       // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"`   // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`         // 用户头像
	Balance   float64    `bson:"balance"json:"money"`          // 用户金额
	UserType  uint32     `bson:"user_type" json:"user_type"`   // 用户类型
	ConnAgent gate.Agent `bson:"conn_agent" json:"conn_agent"` // 网络连接代理
}

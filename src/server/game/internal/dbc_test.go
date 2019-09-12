package internal

import (
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/msg"
	"testing"
)

var dao = DAOConverter{}

func TestMgoC_CUserInfo(t *testing.T) {
	// 数据库
	db, err := NewMgoC(conf.Server.MongoDB)
	if err != nil {
		log.Error("创建数据库客户端错误", err)
	}
	err = db.Init()
	if err != nil {
		log.Error("数据库初始化错误", err)
	}

	udb := dao.U2DB(User{UserID: 12345, NickName: "test", Avatar: "test.png", Balance: 1000.000000001})
	db.CUserInfo(udb)
}

func TestMgoC_CUserBet(t *testing.T) {
	// 数据库
	db, err := NewMgoC(conf.Server.MongoDB)
	if err != nil {
		log.Error("创建数据库客户端错误", err)
	}
	err = db.Init()
	if err != nil {
		log.Error("数据库初始化错误", err)
	}

	u := User{UserID: 12345, NickName: "test", Avatar: "test.png", Balance: 1000.000000001}
	b := msg.Bet{Area: 1, Chip: 2}
	bdb := dao.Bet2DB(u, b)
	db.CUserBet(bdb)
}

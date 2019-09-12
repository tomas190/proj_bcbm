package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
	"proj_bcbm/src/server/base"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/msg"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c *Client4Center // 连接中心服的客户端
	db  *MgoC          // 数据库客户端
	Mgr = NewHall()
)

type Module struct {
	*module.Skeleton
}

// 模块初始化
func (m *Module) OnInit() {
	m.Skeleton = skeleton

	// 中心服务器
	c4c = NewClient4Center()
	c4c.ReqToken()
	c4c.HeartBeatAndListen()
	c4c.CronUpdateToken()

	// 数据库
	db, err := NewMgoC(conf.Server.MongoDB)
	if err != nil {
		log.Error("创建数据库客户端错误 %+v", err)
	}
	err = db.Init()
	if err != nil {
		log.Error("数据库初始化错误 %+v", err)
	}

	// 游戏大厅
	Mgr.OpenCasino()
}

// 模块销毁
func (m *Module) OnDestroy() {
	log.Debug("game模块被销毁...")
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)
}

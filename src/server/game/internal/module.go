package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
	"proj_bcbm/src/server/base"
	"proj_bcbm/src/server/msg"
)


var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c      *Client4Center // 连接中心服的客户端
)

type Module struct {
	*module.Skeleton
}

// 模块初始化
func (m *Module) OnInit() {
	m.Skeleton = skeleton

	//c4c = center.NewClient4Center()
	//c4c.ReqToken()
	//c4c.HeartBeatAndListen()
}

// 模块销毁
func (m *Module) OnDestroy() {
	log.Debug("game模块被销毁...")
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)
}


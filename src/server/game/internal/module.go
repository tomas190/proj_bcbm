package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
	"server/base"
	"server/msg"
)


var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
	manager  = NewHall()                    // 大厅管理
	c4c      *Client4Center                 // 连接中心服的客户端
)

type Module struct {
	*module.Skeleton // 继承自Skeleton
}

// 模块初始化
func (m *Module) OnInit() {
	m.Skeleton = skeleton

	// c4c = NewClient4Center()
	// c4c.ReqToken()
	// c4c.HeartBeatAndListen()
}

// 模块销毁
func (m *Module) OnDestroy() {
	// 服务器主动关闭连接
	log.Debug("game模块被销毁...")

	// 对所有客户端推送踢出消息
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)
	for _, u := range manager.users {
		u.ConnAgent.WriteMsg(data)
	}
}


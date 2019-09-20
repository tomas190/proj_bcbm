package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<----新连接---->")

	u := &User{}
	u.ConnAgent = a  // 保存连接到用户信息
	a.SetUserData(u) // 附加用户信息到连接
}

// todo 暂时删除用户
func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	u, ok := a.UserData().(*User)

	if ok {
		log.Debug("玩家 %+v 主动断开连接...", u.UserID)
		Mgr.UserRecord.Delete(u.UserID)
	}
}

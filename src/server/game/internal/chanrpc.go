package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)

	//skeleton.RegisterChanRPC("UserLogout", rpcUserLogout)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<----新连接---->")
	_ = a

	//u := &User{}
	//u.ConnAgent = a  // 保存连接到用户信息
	//a.SetUserData(u) // 附加用户信息到连接
}

// 心跳停止（被动断开）-掉线
// 关闭连接（主动断开）-断连
// 关服更新（被动断开）-掉线

// 总之，都有可能重连，主动断开不需要向用户推送消息（因为其实连接已经被销毁了）
// 被动断开需要向用户推送消息（网络不稳定，心跳无法检测，服务器重启之类）

// 处理用户主动断开连接的情况
func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	u, ok := a.UserData().(*User)
	// 要是用户没登录，断开就断开，不用做什么处理
	_, logged := manager.users[u.UserID]
	roomID, inRoom := manager.userRoom[u.UserID]
	dealer := manager.dealers[roomID]

	log.Debug("<----用户主动断开连接 %+v---->", u.UserID)
	log.Debug("大厅人数 %+v", len(manager.users))

	// 在大厅中-登出-从大厅中移除用户-重连时重新登录
	if ok && logged && (!inRoom || (inRoom && !dealer.isPlaying())) {
		// rpcUserLogout(args)
		log.Debug("已从大厅中移除用户，用户 %+v 已从中心服登出，当前大厅人数 %+v", u.UserID, len(manager.users))

		if inRoom {
			// 从房间中移除玩家
		}

		a.Close()
		return
	}

	// todo 玩家正在游戏的时候杀了进程
	if ok && logged && inRoom && dealer.isPlaying() {
		// 如果没托管，托管

		// 重连的时候，如果 是 isPlaying
	}

	// 其他情况不用处理
}

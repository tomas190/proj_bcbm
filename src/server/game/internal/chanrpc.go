package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"server/msg"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC("UserLoginTest", rpcTestLogin) // 与中心服对接后注释掉

	//skeleton.RegisterChanRPC("UserLogin", rpcUserLogin)
	//skeleton.RegisterChanRPC("UserLogout", rpcUserLogout)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<----新连接---->")

	u := &User{}
	u.ConnAgent = a  // 保存连接到用户信息
	a.SetUserData(u) // 附加用户信息到连接
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
		rpcUserLogout(args)
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

func rpcTestLogin(args []interface{}) {
	m := args[0].(*msg.LoginTest)
	a := args[1].(gate.Agent)
	userID := m.GetUserID()
	log.Debug("处理用户登录请求 %+v", userID)
	if u, ok := manager.users[userID]; ok && u.ConnAgent == a {
		log.Debug("rpcUserLogin 同一用户相同连接重复登录")
		resp := &msg.Error{
			Code: msg.ErrorCode_UserRepeatLogin,
		}
		log.Debug("<----login 重复登录 resp---->%+v", resp)
		a.WriteMsg(resp)
		return
	} else if _, ok := manager.users[userID]; ok {
		log.Debug("rpcUserLogin 异地登陆")
		err := manager.ReplaceUserAgent(userID, a)
		if err != nil {
			log.Error("用户连接替换错误", err)
		}

		u := manager.users[userID]
		resp := &msg.LoginR{
			User: &msg.UserInfo{
				UserID:   u.UserID,
				Avatar:   u.Avatar,
				Money:    u.Balance,
				NickName: u.NickName,
			},
		}

		if rID, ok := manager.userRoom[userID]; ok {
			resp.RoomID = rID // 如果用户之前在房间里后来退出，返回房间号
		}
		log.Debug("<----当前大厅人数---->%+v", len(manager.users))
		log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
		a.WriteMsg(resp)
	} else if !manager.agentExist(a) {
		u := mockUserInfo(userID)
		resp := &msg.LoginR{
			User: &msg.UserInfo{
				UserID:   u.UserID,
				Avatar:   u.Avatar,
				NickName: u.NickName,
				Money:    u.Balance,
			},
		}

		// 重新绑定信息
		u.ConnAgent = a
		a.SetUserData(u)

		err := manager.AddUser(u) // 添加用户进入大厅
		if err != nil {
			log.Error("添加用户进入大厅失败", err)
		}
		log.Debug("<----当前大厅人数---->%+v", len(manager.users))
		log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
		a.WriteMsg(resp)
	}
}

func rpcUserLogin(args []interface{}) {
	m := args[0].(*msg.Login)
	a := args[1].(gate.Agent)
	userID := m.GetUserID()
	log.Debug("处理用户登录请求 %+v", userID)
	if u, ok := manager.users[userID]; ok && u.ConnAgent == a { // 用户和连接都相同
		log.Debug("rpcUserLogin 同一用户相同连接重复登录")
		resp := &msg.Error{
			Code: msg.ErrorCode_UserRepeatLogin,
		}
		log.Debug("<----login 重复登录 resp---->%+v", resp)
		a.WriteMsg(resp)
		return
	} else if _, ok := manager.users[userID]; ok { // 用户存在，但连接不同
		err := manager.ReplaceUserAgent(userID, a)
		if err != nil {
			log.Error("用户连接替换错误", err)
		}

		u := manager.users[userID]
		resp := &msg.LoginR{
			User: &msg.UserInfo{
				UserID:   u.UserID,
				Avatar:   u.Avatar,
				Money:    u.Balance,
				NickName: u.NickName,
			},
		}

		if rID, ok := manager.userRoom[userID]; ok {
			resp.RoomID = rID // 如果用户之前在房间里后来退出，返回房间号
		}
		log.Debug("<----当前大厅人数---->%+v", len(manager.users))
		log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
		a.WriteMsg(resp)
	} else if !manager.agentExist(a) { // 正常大多数情况
		c4c.UserLoginCenter(userID, m.Password, func(u *User) {
			resp := &msg.LoginR{
				User: &msg.UserInfo{
					UserID:   u.UserID,
					Avatar:   u.Avatar,
					NickName: u.NickName,
					Money:    u.Balance,
				},
			}
			log.Debug("<----login 登录 resp---->%+v", resp)

			// 重新绑定信息
			u.ConnAgent = a
			a.SetUserData(u)

			err := manager.AddUser(u) // 添加用户进入大厅
			if err != nil {
				log.Error("添加用户进入大厅失败", err)
			}
			log.Debug("<----当前大厅人数---->%+v", len(manager.users))
			log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
			a.WriteMsg(resp)
		})
	} // 同一连接上不同用户的情况对第二个用户的请求不做处理
}

func rpcUserLogout(args []interface{}) {
	a := args[0].(gate.Agent)
	u, ok := a.UserData().(*User)

	if ok {
		log.Debug("用户 %+v 从中心服登出", u.UserID)

		//c4c.UserLogoutCenter(u.UserID, func(u *User) {
		//	resp := &msg.LogoutR{}
		//
		//	// 写入
		//	log.Debug("<----logout 登出 resp---->%+v", resp)
		//	a.WriteMsg(resp)
		//	a.SetUserData(nil)
		//	a.Close()
		//	return
		//})

		err := manager.RemoveUser(u.UserID) // 删除登记表中的用户
		if err != nil {
			log.Error("删除登记表中的玩家失败", err)
		}

		a.Close()
		a.Destroy()
	}
}

func mockUserInfo(userID uint32) *User {
	nickName := fmt.Sprintf("test%d", userID)
	avatar := "https://image.flaticon.com/icons/png/128/145/145842.png"
	u := &User{userID, nickName, avatar, 1000, nil}

	return u
}

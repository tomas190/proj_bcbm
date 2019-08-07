package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/patrickmn/go-cache"
	"reflect"
	"server/msg"
	"time"
)

func init()  {
	handlerReg(&msg.Ping{}, handlePing)
	handlerReg(&msg.LoginTest{}, handleTestLogin)

	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)
	handlerReg(&msg.JoinRoom{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom{}, handleLeaveRoom)

	handlerReg(&msg.Bet{}, handleBet)

	handlerReg(&msg.GrabDealer{}, handleGrabDealer)
	handlerReg(&msg.AutoBet{}, handleAutoBet)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

// 处理ping消息
func handlePing(args []interface{})  {
	c := cache.New(time.Second, time.Second)
	c.Set("foo", "bar", cache.DefaultExpiration)
	fmt.Println(c.Get("foo"))
}

func handleTestLogin(args []interface{})  {
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

// 处理登录
func handleLogin(args []interface{})  {
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

// 处理登出
func handleLogout(args []interface{})  {
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

// 加入房间
func handleJoinRoom(args []interface{})  {

}

// 离开房间
func handleLeaveRoom(args []interface{})  {

}

func handleBet(args []interface{})  {

}

// 上庄
func handleGrabDealer(args []interface{})  {

}

// 自动续注
func handleAutoBet(args []interface{})  {

}


func mockUserInfo(userID uint32) *User {
	nickName := fmt.Sprintf("test%d", userID)
	avatar := "https://image.flaticon.com/icons/png/128/145/145842.png"
	u := &User{userID, nickName, avatar, 1000, nil}

	return u
}

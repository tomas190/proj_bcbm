package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"reflect"
	"time"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.LoginTest{}, handleTestLogin)
	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)

	handlerReg(&msg.JoinRoom{}, handleJoinRoom)

	handlerReg(&msg.Bet{}, handleRoomEvent)
	handlerReg(&msg.LeaveRoom{}, handleRoomEvent)
	handlerReg(&msg.GrabBanker{}, handleRoomEvent)
	handlerReg(&msg.AutoBet{}, handleRoomEvent)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

/*************************************

	普通事件

 *************************************/

func handlePing(args []interface{}) {
	// m := args[0].(*msg.Ping)
	a := args[1].(gate.Agent)
	log.Debug("recv Ping %+v", a.RemoteAddr())
	a.WriteMsg(&msg.Pong{})
}

func handleTestLogin(args []interface{}) {
	m := args[0].(*msg.LoginTest)
	a := args[1].(gate.Agent)

	log.Debug("recv %+v, addr %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m)
	userID := m.GetUserID()
	u := mockUserInfo(userID) // 模拟用户

	resp := &msg.LoginR{
		User: &msg.UserInfo{
			UserID:   u.UserID,
			Avatar:   u.Avatar,
			NickName: u.NickName,
			Money:    u.Balance,
		},
		Rooms:      Mgr.GetRoomsInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	// 重新绑定信息
	u.ConnAgent = a
	a.SetUserData(u)

	Mgr.UserRecord[u.UserID] = u
	log.Debug("<---测试登入响应 %+v--->", resp.User)
	a.WriteMsg(resp)
}

func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login)
	a := args[1].(gate.Agent)

	// u := a.UserData().(*User)
	log.Debug("recv %+v, addr %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m)

	a.WriteMsg(&msg.LoginR{
		Rooms: Mgr.GetRoomsInfoResp(),
	})
}

func handleLogout(args []interface{}) {
	m := args[0].(*msg.Logout)
	a := args[1].(gate.Agent)

	log.Debug("recv %+v, addr %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m)
	for i := 0; i < len(args); i++ {
		fmt.Println(reflect.TypeOf(args[i]))
	}
	resp := &msg.LogoutR{}
	a.WriteMsg(resp)
}

/*************************************

	大厅事件-加入房间

 *************************************/

func handleJoinRoom(args []interface{}) {
	m := args[0].(*msg.JoinRoom)
	a := args[1].(gate.Agent)

	log.Debug("recv %+v, addr %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m)

	au := a.UserData().(*User)

	// 找到当前房间的玩家 dealer.getPlayerInfoResp()
	room, exist := Mgr.RoomRecord[m.RoomID]
	if !exist {
		errorResp(a, msg.ErrorCode_RoomNotExist, "")
		return
	}

	// fixme 最大人数
	if len(room.UserBets) == constant.MaxPlayerCount {
		errorResp(a, msg.ErrorCode_RoomFull, "")
		return
	}

	Mgr.AllocateUser(au, Mgr.RoomRecord[m.RoomID])
}

/*************************************

	房间事件-投注、续投、上庄、离开房间

 *************************************/

func handleRoomEvent(args []interface{}) {
	a := args[1].(gate.Agent)
	u, ok := a.UserData().(*User)
	_, logged := Mgr.UserRecord[u.UserID]
	// _, inRoom := Mgr.userRoom[u.UserID]
	inRoom := true
	log.Debug("<----game 房间事件 %v %v ---->", u.UserID, reflect.TypeOf(args[0]))

	if ok && logged && inRoom {
		// 找到玩家房间
		// roomID, ok := Mgr.userRoom[u.UserID]
		ok = true
		if ok {
			// dealer := Mgr.RoomRecord[roomID]
			//log.Debug("当前房间状态 %v", dealer.Status)
			switch t := args[0].(type) {
			//case *msg.Bet:
			//	dealer.handleBet(u)
			case *msg.GrabBanker:
			case *msg.AutoBet:
			default:
				log.Error("房间事件无法识别", t)
			}
		}
	} else {
		errorResp(a, msg.ErrorCode_UserNotInRoom, "")
	}
}

func mockUserInfo(userID uint32) *User {
	nickName := fmt.Sprintf("test%d", userID)
	avatar := "https://image.flaticon.com/icons/png/128/145/145842.png"
	u := &User{userID, nickName, avatar, 1000, nil}

	return u
}

func errorResp(a gate.Agent, err msg.ErrorCode, detail string) {
	log.Debug("<----game 错误resp---->", err)
	a.WriteMsg(&msg.Error{Code: err, Detail: detail})
}

package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/msg"
	"reflect"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)
	handlerReg(&msg.LoginTest{}, handleTestLogin)
	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)
	handlerReg(&msg.JoinRoom{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom{}, handleLeaveRoom)

	handlerReg(&msg.Bet{}, handleBet)
	handlerReg(&msg.GrabBanker{}, handleGrabBanker)
	handlerReg(&msg.AutoBet{}, handleAutoBet)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	// m := args[0].(*msg.Ping)
	a := args[1].(gate.Agent)
	log.Debug("recv Ping %+v", a.RemoteAddr())
	a.WriteMsg(&msg.Pong{})
}

func handleTestLogin(args []interface{}) {
	m := args[0].(*msg.LoginTest)
	a := args[1].(gate.Agent)

	log.Debug("recv LoginTest %+v", a.RemoteAddr())
	userID := m.GetUserID()
	u := mockUserInfo(userID) // 模拟用户

	resp := &msg.LoginR{
		User: &msg.UserInfo{
			UserID:   u.UserID,
			Avatar:   u.Avatar,
			NickName: u.NickName,
			Money:    u.Balance,
		},
		Rooms:getRoomsInfoResp(),
	}

	// 重新绑定信息
	u.ConnAgent = a
	a.SetUserData(u)

	log.Debug("<---登陆响应 %+v--->", resp.User)
	a.WriteMsg(resp)
}

func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login)
	a := args[1].(gate.Agent)

	u := a.UserData().(*User)
	log.Debug("recv Login %+v", a.RemoteAddr())
	a.WriteMsg(&msg.LoginR{
		Rooms:getRoomsInfoResp(),
	})
	fmt.Println(m.UserID, u.UserID)
}

func handleLogout(args []interface{}) {
	for i := 0; i < len(args); i++ {
		fmt.Println(reflect.TypeOf(args[0]))
	}
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*msg.JoinRoom)
	a := args[1].(gate.Agent)

	log.Debug("recv JoinRoom %+v", m.RoomID)
	resp := &msg.JoinRoomR{
		CurBankers:getPlayerInfoResp(),
		Amount:[]float64{21, 400, 325, 235, 109, 111, 345, 908},
		Players:getPlayerInfoResp(),
	}

	log.Debug("<---加入房间响应 %+v--->", resp.Players)
	a.WriteMsg(resp)
}

func handleBet(args []interface{}) {
	m := args[0].(*msg.Bet)
	a := args[1].(gate.Agent)
	au := a.UserData().(*User)

	log.Debug("recv Bet %+v", au.UserID, m.Amount)

	fmt.Println(m.Amount, au.Balance)

	resp := msg.BetR{}
	a.WriteMsg(resp)
}

func handleGrabBanker(args []interface{})  {
	m := args[0].(*msg.GrabBanker)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	fmt.Println(m, au.Balance)

	resp := &msg.BankersB{}
	a.WriteMsg(resp)
}

func handleAutoBet(args []interface{}) {
	m := args[0].(*msg.AutoBet)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	fmt.Println(m, au.Balance)

	resp := &msg.AutoBetR{}
	a.WriteMsg(resp)
}

func handleLeaveRoom(args []interface{}) {
	m := args[0].(*msg.LeaveRoom)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	fmt.Println(m, au.Balance)

	resp := &msg.LeaveRoomR{}
	a.WriteMsg(resp)
}

func getRoomsInfoResp() []*msg.RoomInfo {
	var testResp []*msg.RoomInfo
	room1Info := &msg.RoomInfo{RoomID:908, MinBet:50, History:[]uint32{1, 2, 3, 4, 5, 6, 7}}
	room2Info := &msg.RoomInfo{RoomID:909, MinBet:50, History:[]uint32{1, 2, 3, 4, 5, 6, 7}}

	testResp = append(testResp, room1Info, room2Info)
	return testResp
}

func getPlayerInfoResp() []*msg.UserInfo {
	u1 := mockUserInfo(8976784)
	u2 := mockUserInfo(7829401)

	converter := DTOConverter{}
	userInfo1 := converter.U2Msg(*u1)
	userInfo2 := converter.U2Msg(*u2)

	var testResp []*msg.UserInfo
	testResp = append(testResp, &userInfo1, &userInfo2)

	return testResp
}

func mockUserInfo(userID uint32) *User {
	nickName := fmt.Sprintf("test%d", userID)
	avatar := "https://image.flaticon.com/icons/png/128/145/145842.png"
	u := &User{userID, nickName, avatar, 1000, nil}

	return u
}

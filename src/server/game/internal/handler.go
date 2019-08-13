package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"reflect"
	"server/msg"
)

func init()  {
	handlerReg(&msg.Ping{}, handlePing)
	handlerReg(&msg.LoginTest{}, handleTestLogin)
	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)
	handlerReg(&msg.JoinRoom{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom{}, handleLeaveRoom)

	handlerReg(&msg.Bet{}, handleBet)
	handlerReg(&msg.GrabBanker{}, handleGrabDealer)
	handlerReg(&msg.AutoBet{}, handleAutoBet)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	// m := args[0].(*msg.Ping)
	a := args[1].(gate.Agent)
	a.WriteMsg(&msg.Pong{})
}

func handleTestLogin(args []interface{}) {
	m := args[0].(*msg.LoginTest)
	a := args[1].(gate.Agent)

	userID := m.GetUserID()

	a.WriteMsg(&msg.LoginR{
		Rooms:getRoomsInfoResp(),
	})
	fmt.Println(userID)
}

func handleLogin(args []interface{}) {
	for i := 0; i < len(args); i++ {
		fmt.Println(reflect.TypeOf(args[0]))
	}
}

func handleLogout(args []interface{}) {
	for i := 0; i < len(args); i++ {
		fmt.Println(reflect.TypeOf(args[0]))
	}
}

func handleJoinRoom(args []interface{}) {

}

func handleLeaveRoom(args []interface{}) {

}

func handleBet(args []interface{}) {

}

func handleGrabDealer(args []interface{})  {

}

func handleAutoBet(args []interface{})  {

}

func getRoomsInfoResp() []*msg.RoomInfo {

	var testResp []*msg.RoomInfo
	room1Info := &msg.RoomInfo{RoomID:908, MinBet:50, History:[]uint32{1, 2, 3, 4, 5, 6, 7}}
	room2Info := &msg.RoomInfo{RoomID:909, MinBet:50, History:[]uint32{1, 2, 3, 4, 5, 6, 7}}

	testResp = append(testResp, room1Info, room2Info)
	return testResp
}

func mockUserInfo(userID uint32) *User {
	nickName := fmt.Sprintf("test%d", userID)
	avatar := "https://image.flaticon.com/icons/png/128/145/145842.png"
	u := &User{userID, nickName, avatar, 1000, 1, nil}

	return u
}

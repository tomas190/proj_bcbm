package internal

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"reflect"
	"server/msg"
	"time"
)

func init()  {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)
	handlerReg(&msg.JoinRoom{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom{}, handleLeaveRoom)

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

// 处理登录
func handleLogin(args []interface{})  {

}

// 处理登出
func handleLogout(args []interface{})  {

}

// 加入房间
func handleJoinRoom(args []interface{})  {

}

// 离开房间
func handleLeaveRoom(args []interface{})  {

}

// 上庄
func handleGrabDealer(args []interface{})  {

}

// 自动续注
func handleAutoBet(args []interface{})  {

}


// 8个投注区域
//enum AreaCode {
//_AreaCode = 0;
//AreaCodeBenzGolden = 1; // *40
//AreaCodeBenz = 2;       // *5
//AreaCodeBMWGolden = 3;  // *30
//AreaCodeBMW = 4;        // *5
//AreaCodeAudiGolden = 5; // *20
//AreaCodeAudi = 6;       // *5
//AreaCodeVWGolden = 7;   // *10
//AreaCodeVW = 8;         // *5
//}
package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"server/game"
	"server/msg"
)

func init() {
	// 向当前模块（login 模块）注册 Hello 消息的消息处理函数 handleHello
	handler(&msg.LoginTest{}, handleLoginTest)
	handler(&msg.Login{}, handleLogin)
	handler(&msg.Logout{}, handleLogout)
}

// 异步处理
func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleLoginTest(args []interface{}) {
	log.Debug("<----login 测试登录---->")
	m := args[0].(*msg.LoginTest)
	a := args[1].(gate.Agent)

	log.Debug("Login Test %v %v", m, a)
	game.ChanRPC.Go("UserLoginTest", args[0], args[1])
}

func handleLogin(args []interface{}) {
	log.Debug("<----login 登录---->")

	m := args[0].(*msg.Login)
	a := args[1].(gate.Agent)

	log.Debug("用户登录: %v %v", m, a)
	game.ChanRPC.Go("UserLogin", args[0], args[1])
}

func handleLogout(args []interface{}) {
	log.Debug("<----logout 登出---->")
	m := args[0].(*msg.Logout)
	a := args[1].(gate.Agent)

	log.Debug("用户登出: %v %v", m, a)
	game.ChanRPC.Go("UserLogout", args[0], args[1])
}

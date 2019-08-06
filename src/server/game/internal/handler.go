package internal

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"reflect"
	"server/msg"
	"time"
)

func init()  {
	handler(&msg.Ping{}, handlePing)
}

// 异步处理
func handler(m interface{}, h interface{}) {
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
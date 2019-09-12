package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"testing"
	"time"
)

func TestClient4Center_ServerLoginCenter(t *testing.T) {
	c := NewClient4Center()
	c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(5 * time.Second)

	c.CronUpdateToken()

	for {
		// 在没有收到服务器登陆成功返回之前不应该执行后续操作
		userID := uint32(516499995)

		c.UserLoginCenter(userID, "123456", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

		fmt.Println("#####", c.userWaitEvent)

		time.Sleep(7 * time.Second)

		c.UserLoseScore(userID, -5, "",
			func(data *User) {
				log.Debug("<----用户减钱回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
			})
		time.Sleep(7 * time.Second)

		fmt.Println(c.userWaitEvent)

		c.UserLogoutCenter(userID, func(data *User) {
			log.Debug("<----用户登出回调---->%+v", data.UserID)
		})

		fmt.Println("#####", c.userWaitEvent)

		time.Sleep(7 * time.Second)

		fmt.Println("#####", c.userWaitEvent)

		c.UserLoginCenter(userID, "123456", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

		time.Sleep(7 * time.Second)

		fmt.Println("#####", c.userWaitEvent)
	}
}

// 减钱
func TestClient4Center_MinusMoney(t *testing.T) {
	userID := uint32(516499995)

	c := NewClient4Center()
	c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	c.UserLoginCenter(userID, "123456", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v %+V", data.UserID, data.NickName, data.Balance)
	})

	time.Sleep(2 * time.Second)

	c.UserLoseScore(userID, -1000, "",
		func(data *User) {
			log.Debug("<----用户减钱回调---->%+v %+v", data.UserID, data.Balance)
		})
	time.Sleep(2 * time.Second)

	fmt.Println(c.userWaitEvent)
}

// 加钱
func TestClient4Center_AddMoney(t *testing.T) {
	userID := uint32(789694945)

	c := NewClient4Center()
	c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	c.UserLoginCenter(userID, "123456", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v", data.UserID, data.Balance)
	})

	time.Sleep(2 * time.Second)

	c.UserWinScore(userID, 6000, "test-order-add",
		func(data *User) {
			log.Debug("<----用户加钱回调---->%+v %+v", data.UserID, data.Balance)
		})
	time.Sleep(5 * time.Second)

	fmt.Println(c.userWaitEvent)
}

// 测试中心服重连和断连
func TestClient4Center_ReconnectCenter(t *testing.T) {

}

// 测试定时更新token同时加钱减钱行为
func TestClient4Center_UpdateToken(t *testing.T) {

}

// 并发减钱
func TestClient4Center_ConcurrentLose(t *testing.T) {

}

// 并发加钱
func TestClient4Center_ConcurrentWin(t *testing.T) {

}

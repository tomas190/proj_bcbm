package internal

import (
	"testing"
	"time"

	"github.com/name5566/leaf/log"
)

func TestClient4Center_ServerLoginCenter(t *testing.T) {
	c := NewClient4Center()
	// c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(5 * time.Second)

	// c.CronUpdateToken()

	for {
		// 在没有收到服务器登陆成功返回之前不应该执行后续操作
		userID := uint32(516499995)

		c.UserLoginCenter(userID, "123456", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

		time.Sleep(7 * time.Second)

		c.UserLoseScore(userID, -5, 0, 0, "", "",
			func(data *User) {
				log.Debug("<----用户减钱回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
			})
		time.Sleep(7 * time.Second)

		c.UserLogoutCenter(userID, func(data *User) {
			log.Debug("<----用户登出回调---->%+v", data.UserID)
		})

		time.Sleep(7 * time.Second)

		c.UserLoginCenter(userID, "123456", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

		time.Sleep(7 * time.Second)
	}
}

// 减钱
func TestClient4Center_MinusMoney(t *testing.T) {
	userID := uint32(139366987)

	c := NewClient4Center()
	// c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	c.UserLoginCenter(userID, "e10adc3949ba59abbe56e057f20f883e", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v %+v", data.UserID, data.NickName, data.Balance)
	})

	time.Sleep(2 * time.Second)

	c.UserLoseScore(userID, 1, -1000, 0, "", "",
		func(data *User) {
			log.Debug("<----用户减钱回调---->%+v %+v", data.UserID, data.Balance)
		})
	time.Sleep(2 * time.Second)
}

// 加钱
func TestClient4Center_AddMoney(t *testing.T) {

	var userIDs = []uint32{194989239, 735835433, 990684188, 909098851, 612303604,
		100148012, 139366987, 303586538, 828606651, 984968541,
		678653255, 617222183, 415824137, 251735891, 243271456}

	c := NewClient4Center()
	// c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	for _, uid := range userIDs {
		userID := uid
		c.UserLoginCenter(userID, "e10adc3949ba59abbe56e057f20f883e", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v", data.UserID, data.Balance)
		})

		time.Sleep(2 * time.Second)

		c.UserWinScore(userID, 2000, 0, 0, "test-order-add", "",
			func(data *User) {
				log.Debug("<----用户加钱回调---->%+v %+v", data.UserID, data.Balance)
			})
		time.Sleep(5 * time.Second)
	}
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

package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"testing"
	"time"
)

/*

Dev环境奔驰宝马测试账号 密码全部为 123456
955509280 409972380 615426645 651488813 900948081 263936609 538509606 704898825 943979274 613251393

*/

func TestClient4Center_ServerLoginCenter(t *testing.T) {
	c := NewClient4Center()
	c.ReqToken()
	c.HeartBeatAndListen()
	// 在没有收到服务器登陆成功返回之前不应该执行后续操作
	time.Sleep(1 * time.Second)

	c.UserLoginCenter(955509280, "123456", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
	})

	fmt.Println("#####", c.userWaitEvent)

	time.Sleep(1 * time.Second)

	//c.UserWinScore(955509280, uint32(time.Now().Unix()),
	//	time.Now().Format("2006-01-02T15:04:05"), "test win score",
	//	func(data *User) {
	//		log.Debug("<----用户加钱回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
	//	})
	//
	//fmt.Println("#####", c.userWaitEvent)

	time.Sleep(1 * time.Second)

	c.UserLoseScore(955509280, uint32(time.Now().Unix()),
		time.Now().Format("2006-01-02T15:04:05"), "test win score",
		func(data *User) {
			log.Debug("<----用户减钱回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

	fmt.Println("#####", c.userWaitEvent)

	c.UserLoseScore(955509280, uint32(time.Now().Unix()),
		time.Now().Format("2006-01-02T15:04:05"), "test win score",
		func(data *User) {
			log.Debug("<----用户减钱回调2---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

	fmt.Println("#####", c.userWaitEvent)

	time.Sleep(1 * time.Second)

	c.UserLogoutCenter(955509280, func(data *User) {
		log.Debug("<----用户登出回调---->%+v", data.UserID)
	})

	fmt.Println("#####", c.userWaitEvent)

	time.Sleep(1 * time.Second)

	fmt.Println("#####", c.userWaitEvent)

}

// 投注减钱
func TestClient4Center_BetLoseMoney(t *testing.T) {
	// bm := 100.1

	c := NewClient4Center()
	c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	c.UserLoginCenter(955509280, "123456", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
	})

	time.Sleep(2 * time.Second)

	c.UserLoseScore(955509280, uint32(time.Now().Unix()), "", "aaa",
		func(data *User) {
			log.Debug("<----用户减钱回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})
	time.Sleep(2 * time.Second)

	fmt.Println(c.userWaitEvent)
}

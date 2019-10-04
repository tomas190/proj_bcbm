package internal

import (
	"fmt"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/util"
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

	uuid := util.UUID{}
	round := uuid.GenUUID()

	for {
		// 在没有收到服务器登陆成功返回之前不应该执行后续操作
		userID := uint32(516499995)

		c.UserLoginCenter(userID, "123456", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v %+v", data.Balance, data.NickName, data.Avatar)
		})

		time.Sleep(7 * time.Second)

		c.UserLoseScore(userID, -5, uuid.GenUUID(), round,
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

	uuid := util.UUID{}
	round := uuid.GenUUID()
	c.UserLoginCenter(userID, "e10adc3949ba59abbe56e057f20f883e", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v %+v", data.UserID, data.NickName, data.Balance)
	})

	time.Sleep(2 * time.Second)

	c.UserLoseScore(userID, 0, uuid.GenUUID(), round,
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

	uuid := util.UUID{}
	round := uuid.GenUUID()

	for _, uid := range userIDs {
		winOrder := uuid.GenUUID()
		userID := uid
		c.UserLoginCenter(userID, "e10adc3949ba59abbe56e057f20f883e", func(data *User) {
			log.Debug("<----用户登录回调---->%+v %+v", data.UserID, data.Balance)
		})

		time.Sleep(1 * time.Second)

		c.UserWinScore(userID, 20000, winOrder+"-add", round,
			func(data *User) {
				log.Debug("<----用户加钱回调---->%+v %+v", data.UserID, data.Balance)
			})
		time.Sleep(1 * time.Second)

		c.UserLogoutCenter(userID, func(data *User) {
			log.Debug("<----用户登出回调---->%+v %+v", data.UserID, data.Balance)
		})
	}
}

func TestClient4Center_ChangeBankerStatus(t *testing.T) {
	c := NewClient4Center()
	// c.ReqToken()
	c.HeartBeatAndListen()
	time.Sleep(1 * time.Second)

	uuid := util.UUID{}
	round := uuid.GenUUID()
	userID := uint32(194989239)
	// 登录
	c.UserLoginCenter(userID, "e10adc3949ba59abbe56e057f20f883e", func(data *User) {
		log.Debug("<----用户登录回调---->%+v %+v", data.UserID, data.Balance)
	})

	time.Sleep(1 * time.Second)
	//
	// 投注
	c.UserLoseScore(userID, -100, uuid.GenUUID(), round, func(data *User) {
		fmt.Println("减钱完成")
	})

	time.Sleep(1 * time.Second)

	// 申请上庄
	c.ChangeBankerStatus(userID, constant.BSGrabbingBanker, 5000, uuid.GenUUID(), round, func(data *User) {
		fmt.Println("申请上庄")
	})

	time.Sleep(1 * time.Second)

	// 坐庄

	// 庄家输
	c.BankerLoseScore(userID, -200, 0, 0, uuid.GenUUID(), round, func(data *User) {
		fmt.Println("庄家输")
	})

	time.Sleep(1 * time.Second)

	// 庄家赢
	c.BankerWinScore(userID, 400, 0, 0, uuid.GenUUID(), round, func(data *User) {
		fmt.Println("庄家赢")
	})

	time.Sleep(1 * time.Second)

	// 下庄
	c.ChangeBankerStatus(userID, constant.BSNotBanker, -5180, uuid.GenUUID(), round, func(data *User) {
		fmt.Println("庄家下庄")
	})

	time.Sleep(1 * time.Second)

	// 登出（如果不在游戏里面）
	c.UserLogoutCenter(userID, func(data *User) {
		fmt.Println("登出")
	})

	time.Sleep(1 * time.Second)
}

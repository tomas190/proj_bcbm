package internal

import (
	"github.com/name5566/leaf/gate"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"sync"
	"time"
)

type User struct {
	Balance       float64      // 用户金额
	BalanceLock   sync.RWMutex // 锁
	BankerBalance float64      // 上庄金额
	UserID        uint32       // 用户id
	Status        int          // 上庄状态
	NickName      string       // 用户昵称
	Avatar        string       // 用户头像
	PackageId     uint16
	ConnAgent     gate.Agent // 网络连接代理
	DownBetTotal  float64    // 玩家总下注
	winCount      uint32     // 玩家赢的次数
	betAmount     float64    // 玩家总投注金额
	LockMoney     float64    // 下注锁定的钱
	IsAction      bool       // 玩家是否行动
}

func (au *User) Init() {
	au.Balance = 0
	au.BankerBalance = 0
	au.DownBetTotal = 0
	au.winCount = 0
	au.betAmount = 0
	au.IsAction = false
}

type Bot struct {
	UserID        uint32
	NickName      string
	Avatar        string
	Balance       float64
	BankerBalance float64
	WinCount      uint32
	BetAmount     float64
	botType       uint32
	TwentyData    []int
	Status        int
}

type Player interface {
	GetBalance() float64
	GetBankerBalance() float64

	GetPlayerBasic() (uint32, string, string, float64)
	GetPlayerAccount() (uint32, float64)
}

func (au User) GetBalance() float64 {
	return au.Balance
}

func (au User) GetBankerBalance() float64 {
	return au.BankerBalance
}

func (au User) GetPlayerBasic() (uint32, string, string, float64) {
	return au.UserID, au.NickName, au.Avatar, au.Balance
}

// 返回玩家投注了的近20局获胜局数和总下注数
func (au User) GetPlayerAccount() (uint32, float64) {
	return au.winCount, au.betAmount
}

func (b Bot) GetBalance() float64 {
	return b.Balance
}

func (b Bot) GetBankerBalance() float64 {
	return b.Balance
}

func (b Bot) GetPlayerBasic() (uint32, string, string, float64) {
	return b.UserID, b.NickName, b.Avatar, b.Balance
}

func (b Bot) GetPlayerAccount() (uint32, float64) {
	return b.WinCount, b.BetAmount
}

func (au *User) PlayerReqExit(dl *Dealer) {
	math := util.Math{}
	uBets, _ := math.SumSliceFloat64(dl.UserBets[au.UserID]).Float64() // 获取下注金额
	if au.IsAction == false || uBets == 0 {
		au.winCount = 0
		au.betAmount = 0
		dl.UserIsDownBet[au.UserID] = false
		au.IsAction = false
		dl.UserBets[au.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
		dl.Users.Delete(au.UserID)
		delete(Mgr.UserRoom, au.UserID)
		dl.DeleteRoomRecord()
	} else {
		var exist bool
		for _, v := range dl.UserLeave {
			if v == au.UserID {
				exist = true
			}
		}
		if exist == false {
			dl.UserLeave = append(dl.UserLeave, au.UserID)
		}
	}

	dl.AutoBetRecord[au.UserID] = nil

	resp := &msg.LeaveRoomR{
		User: &msg.UserInfo{
			UserID:   au.UserID,
			Avatar:   au.Avatar,
			NickName: au.NickName,
			Money:    au.Balance,
		},
		Rooms:      Mgr.GetRoomsInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}
	au.ConnAgent.WriteMsg(resp)
}

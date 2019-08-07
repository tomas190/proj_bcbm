package internal

import (
	"errors"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"server/msg"
	"time"
)

type Hall struct {
	roomHead uint32
	dealers  map[uint32]*Dealer // 房间记录
	users    map[uint32]*User   // 用户记录
	userRoom map[uint32]uint32  // 用户在哪间房间记录
}


func NewHall() *Hall {
	return &Hall{
		roomHead: 100000,
		dealers:  make(map[uint32]*Dealer, 1),
		users:    make(map[uint32]*User, 1),
		userRoom: make(map[uint32]uint32, 1),
	}
}

/******************************

	用户类

 ******************************/

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`     // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"` // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`       // 用户头像
	Balance   float64    `bson:"balance"json:"money"`        // 用户金额
	ConnAgent gate.Agent // 网络连接代理
}

type Dealer struct {
	*Room
	status      int          // 房间状态
	multi       uint32       // 当前倍数
	clock       *time.Ticker // 计时器
	counter     uint32       // 已经过去多少秒
}

type Room struct {
	RoomID    uint32 // 房间基本信息
	IsPrivate bool   // 是否私有房
	RoomPWD   string // 房间密码

	BasePoint float64       // 底分
	MinMoney  float64       // 最小进入限制
}

func (*Dealer) isPlaying() bool {
	return false
}

// 替换用户连接
func (h *Hall) ReplaceUserAgent(userID uint32, agent gate.Agent) error {
	log.Debug("用户重连或顶替，正在替换agent", userID)
	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
	if _, ok := h.users[userID]; ok {
		resp := &msg.Error{
			Code:      msg.ErrorCode_UserRemoteLogin,
		}
		h.users[userID].ConnAgent.WriteMsg(resp)
		h.users[userID].ConnAgent.Destroy()
		h.users[userID].ConnAgent = agent
		h.users[userID].ConnAgent.SetUserData(h.users[userID])
		return nil
	} else {
		return errors.New("用户不在登记表中")
	}
}

// agent 是否已经存在
// 是否开销过大？后续可通过新增记录解决
func (h *Hall) agentExist(a gate.Agent) bool {
	for _, u := range h.users {
		if u.ConnAgent == a {
			return true
		}
	}

	return false
}

// fixme 两个goroutine同时写map会有问题
func (h *Hall) AddUser(u *User) error {
	h.users[u.UserID] = u
	return nil
}

func (h *Hall) RemoveUser(userID uint32) error {
	return nil
}




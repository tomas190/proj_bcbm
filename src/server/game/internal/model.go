package internal

import (
	"github.com/name5566/leaf/gate"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
)

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`       // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"`   // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`         // 用户头像
	Balance   float64    `bson:"balance"json:"money"`          // 用户金额
	ConnAgent gate.Agent `bson:"conn_agent" json:"conn_agent"` // 网络连接代理
}

type Hall struct {
	UserRecord map[uint32]User        // 用户记录
	RoomRecord map[uint32]Dealer        // 房间记录
	Statistic  map[uint32][]uint32    // 各房间历史记录统计
	History    map[uint32][]uint32    // 各房间历史记录
}

func NewHall() *Hall {
	return &Hall{

	}
}

// 开赌场 初始化的时候直接开6个房间然后跑在不同的goroutine上
// 大厅和房间之间通过channel通信
func (h *Hall) OpenCasino() {
	for i := 0; i < constant.RoomCount; i++ {
		go h.openRoom(uint32(i))
	}
}

// 大厅开房
func (h *Hall) openRoom(rID uint32) {
	dl := NewDealer(rID)
	dl.StartGame()

	// h.BroadCast(&msg.BetInfoB{})
	// 下注
	// 开奖
	// h.BroadCast(&msg.PlayersB{})
	// 结算
}

// 大厅广播
func (h *Hall) BroadCast(bMsg interface{}) {
	for _, u := range h.UserRecord {
		u.ConnAgent.WriteMsg(bMsg)
	}
}

// 当前房间信息
func (h *Hall) GetRoomsInfoResp() []*msg.RoomInfo {
	var roomsInfoResp []*msg.RoomInfo
	converter := DTOConverter{}

	for _, r := range h.RoomRecord {
		rMsg := converter.R2Msg(r)
		roomsInfoResp = append(roomsInfoResp, &rMsg)
	}

	return roomsInfoResp
}

type Room struct {
	RoomID       uint32
	MinBet       float64
	MaxBet       float64
	MinLimit     float64
}

func NewRoom(rID uint32, minB, maxB, minL float64) *Room  {
	return &Room{
		RoomID:rID,
		MinBet:minB,
		MaxBet:maxB,
		MinLimit:minL,
	}
}

type roomStatus struct {
	Status  uint32
	EndTime uint32
	Result  uint32
}

var roomStatusChan chan roomStatus

// 大厅监测各房间发来的消息，如果变化发出房间状态变化广播

// 房间

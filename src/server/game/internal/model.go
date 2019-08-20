package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"reflect"
)

// 房间状态改变时通知大厅
type HRMsg struct {
	RoomID        uint32
	RoomStatus    uint32
	LotteryResult uint32
	EndTime       uint32
}

type Hall struct {
	UserRecord map[uint32]*User    // 用户记录
	RoomRecord map[uint32]*Dealer  // 房间记录
	History    map[uint32][]uint32 // 各房间历史记录
	HRChan     chan HRMsg          // 房间大厅通信
}

func NewHall() *Hall {
	return &Hall{
		UserRecord: make(map[uint32]*User),
		RoomRecord: make(map[uint32]*Dealer),
		History:    make(map[uint32][]uint32),
		HRChan:     make(chan HRMsg, 6),
	}
}

// 开赌场 初始化的时候直接开6个房间然后跑在不同的goroutine上
// 大厅和房间之间通过channel通信
func (h *Hall) OpenCasino() {
	for i := 0; i < constant.RoomCount; i++ {
		go h.openRoom(uint32(i + 1))
	}

	// 收到房间channel消息后发广播
	go func() {
		for {
			select {
			case hrMsg := <-h.HRChan:
				h.ChangeRoomStatus(hrMsg)
			default:

			}
		}
	}()
}

// 大厅开房
func (h *Hall) openRoom(rID uint32) {
	dl := NewDealer(rID, h.HRChan)
	h.RoomRecord[rID] = dl
	dl.StartGame()
}

// 收到房间消息状态改变的消息后
// 修改大厅统计任务
// 发广播
func (h *Hall) ChangeRoomStatus(hrMsg HRMsg) {
	rID := hrMsg.RoomID
	log.Debug("roomStatus: %+v", hrMsg.RoomStatus)
	if hrMsg.RoomStatus == constant.RSSettle {
		h.History[rID] = append(h.History[rID], hrMsg.LotteryResult)
		h.RoomRecord[rID].History = append(h.RoomRecord[rID].History, hrMsg.LotteryResult)
		log.Debug("room: %+v, his: %+v", rID, h.History[rID])
		if len(h.History[rID]) > constant.HisCount {
			h.History[rID] = h.History[rID][1:]
		}
	}

	converter := &DTOConverter{}
	res := converter.RChangeHB(hrMsg, h.RoomRecord[rID].counter)
	h.BroadCast(&res)
}

// 大厅广播
func (h *Hall) BroadCast(bMsg interface{}) {
	log.Debug("brd msg %+v, content: %+v", reflect.TypeOf(bMsg), bMsg)
	for _, u := range h.UserRecord {
		u.ConnAgent.WriteMsg(bMsg)
	}
}

// 当前房间信息
func (h *Hall) GetRoomsInfoResp() []*msg.RoomInfo {
	var roomsInfoResp []*msg.RoomInfo
	converter := DTOConverter{}

	for _, r := range h.RoomRecord {
		rMsg := converter.R2Msg(*r)
		roomsInfoResp = append(roomsInfoResp, &rMsg)
	}

	return roomsInfoResp
}

type Room struct {
	RoomID   uint32
	MinBet   float64
	MaxBet   float64
	MinLimit float64
}

func NewRoom(rID uint32, minB, maxB, minL float64) *Room {
	return &Room{
		RoomID:   rID,
		MinBet:   minB,
		MaxBet:   maxB,
		MinLimit: minL,
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

type User struct {
	UserID    uint32     `bson:"user_id" json:"user_id"`       // 用户id
	NickName  string     `bson:"nick_name" json:"nick_name"`   // 用户昵称
	Avatar    string     `bson:"avatar" json:"avatar"`         // 用户头像
	Balance   float64    `bson:"balance"json:"money"`          // 用户金额
	ConnAgent gate.Agent `bson:"conn_agent" json:"conn_agent"` // 网络连接代理
}

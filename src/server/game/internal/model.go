package internal

import (
	"errors"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"reflect"
	"time"
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
	UserRoom   map[uint32]uint32   // 用户房间
	History    map[uint32][]uint32 // 各房间历史记录
	HRChan     chan HRMsg          // 房间大厅通信
}

func NewHall() *Hall {
	return &Hall{
		UserRecord: make(map[uint32]*User),
		RoomRecord: make(map[uint32]*Dealer),
		UserRoom:   make(map[uint32]uint32),
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
	dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
	h.RoomRecord[rID] = dl // fixme 会导致 fatal error: concurrent map writes
	dl.StartGame()
}

func (h *Hall) AllocateUser(u *User, dl *Dealer) {
	h.UserRoom[u.UserID] = dl.RoomID
	dl.Users[u.UserID] = u
	dl.UserBets[u.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

	converter := DTOConverter{}
	r := converter.R2Msg(*dl)

	resp := &msg.JoinRoomR{
		CurBankers: dl.getBankerInfoResp(),
		Amount:     dl.AreaBets,
		PAmount:    dl.UserBets[u.UserID],
		Room:       &r,
		ServerTime: uint32(time.Now().Unix()),
	}

	log.Debug("<---加入房间响应 %+v--->", resp.Room)
	u.ConnAgent.WriteMsg(resp)
}

// 收到房间消息状态改变的消息后 修改大厅统计任务 发广播
func (h *Hall) ChangeRoomStatus(hrMsg HRMsg) {
	rID := hrMsg.RoomID
	log.Debug("roomStatus: %+v", hrMsg.RoomStatus)
	if hrMsg.RoomStatus == constant.RSSettle {
		h.History[rID] = append(h.History[rID], hrMsg.LotteryResult)
		log.Debug("room: %+v, his: %+v", rID, h.History[rID])
		if len(h.History[rID]) > constant.HisCount {
			h.History[rID] = h.History[rID][1:]
		}
		h.RoomRecord[rID].History = h.History[rID]
	}

	converter := &DTOConverter{}
	res := converter.RChangeHB(hrMsg, *h.RoomRecord[rID])
	h.BroadCast(&res)
}

// 大厅广播
func (h *Hall) BroadCast(bMsg interface{}) {
	log.Debug("brd msg %+v, content: %+v", reflect.TypeOf(bMsg), bMsg)
	for _, u := range h.UserRecord {
		if u.ConnAgent != nil {
			u.ConnAgent.WriteMsg(bMsg)
		}
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

// 替换用户连接
func (h *Hall) ReplaceUserAgent(userID uint32, agent gate.Agent) error {
	log.Debug("用户重连或顶替，正在替换agent %+v", userID)
	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
	if _, ok := h.UserRecord[userID]; ok {
		errorResp(agent, msg.ErrorCode_UserRemoteLogin, "异地登录")
		h.UserRecord[userID].ConnAgent.Destroy()
		h.UserRecord[userID].ConnAgent = agent
		h.UserRecord[userID].ConnAgent.SetUserData(h.UserRecord[userID])
		return nil
	} else {
		return errors.New("用户不在登记表中")
	}
}

// agent 是否已经存在
// 是否开销过大？后续可通过新增记录解决
func (h *Hall) agentExist(a gate.Agent) bool {
	for _, u := range h.UserRecord {
		if u.ConnAgent == a {
			return true
		}
	}

	return false
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

package internal

import (
	"errors"
	"fmt"
	"github.com/name5566/leaf/gate"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"strconv"
	"sync"
	"time"
)

// 房间状态改变时通知大厅
type HRMsg struct {
	RoomID        string
	RoomStatus    uint32
	LotteryResult uint32
	EndTime       uint32
}

type Hall struct {
	UserRecord sync.Map            // 用户记录
	RoomRecord sync.Map            // 房间记录
	UserRoom   map[uint32]string   // 用户房间
	History    map[string][]uint32 // 各房间历史记录
	HRChan     chan HRMsg          // 房间大厅通信

	OrderIDRecord sync.Map // orderID对应user
}

func NewHall() *Hall {
	return &Hall{
		UserRecord:    sync.Map{},
		RoomRecord:    sync.Map{},
		UserRoom:      make(map[uint32]string),
		History:       make(map[string][]uint32),
		HRChan:        make(chan HRMsg, 6),
		OrderIDRecord: sync.Map{},
	}
}

// 开赌场 初始化的时候直接开6个房间然后跑在不同的goroutine上
// 大厅和房间之间通过channel通信
func (h *Hall) OpenCasino() {
	for i := 0; i < constant.RoomCount; i++ {
		go h.openRoom(strconv.Itoa(i + 1))
	}

	// 收到房间channel消息后发广播
	go func() {
		for {
			select {
			case hrMsg := <-h.HRChan:
				h.ChangeRoomStatus(hrMsg)
			}
		}
	}()
}

// 大厅开房
func (h *Hall) openRoom(rID string) *Dealer {

	dl := NewDealer(rID, h.HRChan)

	ru := util.Random{}
	num := ru.RandInRange(0, 100)
	if num >= 0 && num <= 33 {
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
	} else if num > 33 && num <= 66 {
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
	} else if num > 66 && num <= 100 {
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
		dl.Bankers = append(dl.Bankers, dl.NextBotBanker())
	}

	dl.bankerMoney = dl.Bankers[0].(Bot).Balance

	h.RoomRecord.Store(rID, dl)
	dl.StartGame()

	return dl
}

func (h *Hall) CreatePackageRoom() {

}

func (h *Hall) PlayerJoinRoom(roomId string, au *User) {

	rid := Mgr.UserRoom[au.UserID]
	v, _ := Mgr.RoomRecord.Load(rid)
	if v != nil {
		dl := v.(*Dealer)
		for i, lu := range dl.UserLeave {
			user := lu
			// 把玩家从掉线列表中移除
			if user == au.UserID {
				dl.UserLeave = append(dl.UserLeave[:i], dl.UserLeave[i+1:]...)
				log.Debug("AllocateUser 清除玩家记录~")
				break
			}
		}
		h.AllocateUser(au, dl, true)
		return
	}

	ok, dl := h.GetPackageIdRoom(au)
	if ok {
		h.AllocateUser(au, dl, false)
	} else {
		h.CreatJoinPackageIdRoom(roomId, au)
	}
}

func (h *Hall) AllocateUser(u *User, dl *Dealer, again bool) {

	if !again {
		h.UserRoom[u.UserID] = dl.RoomID
		dl.Users.Store(u.UserID, u)
		if _, ok := dl.UserBets[u.UserID]; !ok {
			dl.UserBets[u.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
		}
	}

	converter := DTOConverter{}
	r := converter.R2Msg(*dl)
	mu := converter.U2Msg(*u)

	data := &msg.RespRoomStatus{
		InGame: true,
		RoomID: dl.RoomID,
	}
	u.ConnAgent.WriteMsg(data)

	resp := &msg.JoinRoomR{
		User:       &mu,
		CurBankers: dl.getBankerInfoResp(),
		Amount:     dl.AreaBets,
		PAmount:    dl.UserBets[u.UserID],
		Room:       &r,
		ServerTime: uint32(time.Now().Unix()),
	}

	u.ConnAgent.WriteMsg(resp)
}

func (h *Hall) GetPackageIdRoom(p *User) (bool, *Dealer) {
	dealer := &Dealer{}
	var IsExist = false
	h.RoomRecord.Range(func(key, value interface{}) bool {
		dl := value.(*Dealer)
		if dl.PackageId == p.PackageId {
			dealer = dl
			IsExist = true
		}
		return true
	})
	return IsExist, dealer
}

func (h *Hall) CreatJoinPackageIdRoom(roomId string, au *User) {

	rid := fmt.Sprintf(roomId + "-" + strconv.Itoa(int(au.PackageId)))

	go func() {
		dl := h.openRoom(rid)
		dl.PackageId = au.PackageId
		if dl.PackageId == 8 || dl.PackageId == 11 {
			dl.IsSpecial = true
		}
		// 加入房间
		h.AllocateUser(au, dl, false)
		log.Debug("品牌房间首次进入玩家:%v, %v", dl.PackageId, au.Id)
	}()
}

// 收到房间消息状态改变的消息后 修改大厅统计任务 发广播
func (h *Hall) ChangeRoomStatus(hrMsg HRMsg) {
	rID := hrMsg.RoomID
	if hrMsg.RoomStatus == constant.RSSettle {
		h.History[rID] = append(h.History[rID], hrMsg.LotteryResult)
		if len(h.History[rID]) > constant.HisCount {
			h.History[rID] = h.History[rID][1:]
		}
		v, _ := h.RoomRecord.Load(rID)
		v.(*Dealer).History = h.History[rID]
	}

	converter := &DTOConverter{}
	v, _ := h.RoomRecord.Load(rID)
	res := converter.RChangeHB(hrMsg, *(v.(*Dealer)))
	h.BroadCast(&res)
}

// 大厅广播
func (h *Hall) BroadCast(bMsg interface{}) {
	// log.Debug("hall brd msg %+v, content: %+v", reflect.TypeOf(bMsg), bMsg)
	h.UserRecord.Range(func(key, value interface{}) bool {
		user := value.(*User)
		if user.ConnAgent != nil {
			user.ConnAgent.WriteMsg(bMsg)
		}
		return true
	})
}

// 当前房间信息
func (h *Hall) GetRoomsInfoResp() []*msg.RoomInfo {
	var roomsInfoResp []*msg.RoomInfo
	converter := DTOConverter{}

	// 因为前端需要排序
	var sortedKeys []uint32
	for i := 0; i < constant.RoomCount; i++ {
		sortedKeys = append(sortedKeys, uint32(i)+1)
	}

	for _, k := range sortedKeys {
		kx := k
		if v, ok := h.RoomRecord.Load(kx); ok {
			rMsg := converter.R2Msg(*v.(*Dealer))
			roomsInfoResp = append(roomsInfoResp, &rMsg)
		} else {
			log.Debug("GetRoomsInfoResp 找不到房间id")
		}
	}

	return roomsInfoResp
}

// 替换用户连接
func (h *Hall) ReplaceUserAgent(userID uint32, agent gate.Agent) error {
	log.Debug("用户重连或顶替，正在替换agent %+v", userID)
	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
	if v, ok := h.UserRecord.Load(userID); ok {
		errorResp(agent, msg.ErrorCode_UserRemoteLogin, "异地登录")
		user := v.(*User)
		user.ConnAgent = agent
		user.ConnAgent.SetUserData(v)
		return nil
	} else {
		return errors.New("用户不在登记表中")
	}
}

// agent 是否已经存在
// 是否开销过大？后续可通过新增记录解决
func (h *Hall) agentExist(a gate.Agent) bool {
	var exist bool
	h.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*User)
		if u.ConnAgent == a {
			exist = true
		}
		return true
	})

	return exist
}

type Room struct {
	RoomID    string
	IsSpecial bool
	PackageId uint16
	MinBet    float64
	MaxBet    float64
	MinLimit  float64
}

func NewRoom(rID string, minB, maxB, minL float64) *Room {
	return &Room{
		RoomID:    rID,
		MinBet:    minB,
		PackageId: 0,
		IsSpecial: false,
		MaxBet:    maxB,
		MinLimit:  minL,
	}
}

package internal

import (
	"proj_bcbm/src/server/msg"
	"time"
)

type DTOConverter struct{}

func (c *DTOConverter) U2Msg(u User) msg.UserInfo {
	uMsg := msg.UserInfo{
		UserID:   u.UserID,
		NickName: u.NickName,
		Avatar:   u.Avatar,
		Money:    u.Balance,
	}

	return uMsg
}

func (c *DTOConverter) R2Msg(dl Dealer) msg.RoomInfo {
	stat := make([]uint32, 8)
	for _, his := range dl.History {
		stat[his-1]++
	}
	rMsg := msg.RoomInfo{
		RoomID:     dl.RoomID,
		MinBet:     dl.MinBet,
		MaxBet:     dl.MaxBet,
		MinLimit:   dl.MinLimit,
		Counter:    dl.counter,
		Status:     dl.Status,
		History:    dl.History,
		Statistics: stat,
	}

	return rMsg
}

func (c *DTOConverter) RChangeHB(m HRMsg, dl Dealer) msg.RoomChangeHB {
	stat := make([]uint32, 8)
	for _, his := range dl.History {
		stat[his-1]++
	}
	bMsg := msg.RoomChangeHB{
		RoomID:     m.RoomID,
		Result:     m.LotteryResult,
		EndTime:    m.EndTime,
		ServerTime: uint32(time.Now().Unix()),
		Status:     m.RoomStatus,
		Counter:    dl.counter,
		Statistics: stat,
	}

	return bMsg
}

type DAOConverter struct{}

func (c *DAOConverter) U2Bson() {

}

func (c *DAOConverter) R2Bson() {

}

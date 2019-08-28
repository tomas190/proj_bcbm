package internal

import (
	"proj_bcbm/src/server/msg"
	"time"
)

type DTOConverter struct{}

func (c *DTOConverter) U2Msg(u User) msg.UserInfo {
	uMsg := msg.UserInfo{
		UserID:    u.UserID,
		NickName:  u.NickName,
		Avatar:    u.Avatar,
		Money:     u.Balance,
		WinCount:  10, // fixme
		BetAmount: 100,
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
		EndTime:    dl.ddl,
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
		EndTime:    dl.ddl,
		ServerTime: uint32(time.Now().Unix()),
		Status:     m.RoomStatus,
		Counter:    dl.counter,
		Statistics: stat,
	}

	return bMsg
}

func (c *DTOConverter) RSBMsg(res uint32, win float64, money float64, dl Dealer) msg.RoomStatusB {
	bMsg := msg.RoomStatusB{
		Status:      dl.Status,
		Counter:     0,
		EndTime:     dl.ddl,
		Result:      res,
		WinMoney:    win,
		PlayerMoney: money,
		ServerTime:  uint32(time.Now().Unix()),
	}

	return bMsg
}

type DAOConverter struct{}

func (c *DAOConverter) U2Bson() {

}

func (c *DAOConverter) R2Bson() {

}

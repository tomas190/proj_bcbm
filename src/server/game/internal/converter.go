package internal

import (
	"proj_bcbm/src/server/msg"
	"time"
)

type DTOConverter struct{}

func (c *DTOConverter) U2Msg(p Player) msg.UserInfo {
	id, name, img, score := p.GetPlayerBasic()
	win, bet := p.GetPlayerAccount()
	uMsg := msg.UserInfo{
		UserID:    id,
		NickName:  name,
		Avatar:    img,
		Money:     score,
		WinCount:  win,
		BetAmount: bet,
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

func (c *DTOConverter) RSBMsg(userWin float64, money float64, dl Dealer) msg.RoomStatusB {
	bMsg := msg.RoomStatusB{
		Status:      dl.Status,
		Counter:     0, // fixme
		EndTime:     dl.ddl,
		Result:      dl.res,
		BankerWin:   dl.bankerWin,
		WinMoney:    userWin,
		PlayerMoney: money,
		ServerTime:  uint32(time.Now().Unix()),
	}

	return bMsg
}

func (c *DTOConverter) BBMsg(dealer Dealer) msg.BankersB {
	bMsg := msg.BankersB{
		Banker:     dealer.getBankerInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	return bMsg
}

type DAOConverter struct{}

func (c *DAOConverter) U2Bson() {

}

func (c *DAOConverter) R2Bson() {

}

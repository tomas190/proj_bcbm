package internal

import (
	"proj_bcbm/src/server/msg"
	"time"
	"go.mongodb.org/mongo-driver/bson"
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

func (c *DTOConverter) Banker2Msg(p Player) msg.UserInfo {
	id, name, img, money := p.GetPlayerBasic()
	bankerMoney := p.GetBankerBalance()
	win, bet := p.GetPlayerAccount()
	uMsg := msg.UserInfo{
		UserID:      id,
		NickName:    name,
		Avatar:      img,
		Money:       money,
		BankerMoney: bankerMoney,
		WinCount:    win,
		BetAmount:   bet,
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
		// Counter:    0,
		Statistics: stat,
	}

	return bMsg
}

func (c *DTOConverter) RSBMsg(userWin float64, autoBetAmount, userBalance float64, dl Dealer) msg.RoomStatusB {
	bMsg := msg.RoomStatusB{
		Status: dl.Status,
		// Counter:       0,
		EndTime:       dl.ddl,
		Result:        dl.res,
		BankerWin:     dl.bankerWin,
		WinMoney:      userWin,
		BankerMoney:   dl.bankerMoney,
		AutoBetAmount: autoBetAmount, // 若不可续投则为0
		PlayerMoney:   userBalance,
		ServerTime:    uint32(time.Now().Unix()),
		Position:      dl.pos,
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

func (c *DAOConverter) U2DB(u User) UserDB {
	udb := UserDB{u.UserID, u.NickName, u.Avatar, u.Balance}
	return udb
}

// 玩家结算记录
func (c *DAOConverter) Settle2DB(u User, winOrder, rID string, isWin bool, betAmount, winAmount float64) SettleDB {
	user := c.U2DB(u)
	sdb := SettleDB{User: user, WinOrder: winOrder, RoundID: rID, IsWin: isWin, BetAmount: betAmount, WinAmount: winAmount}
	return sdb
}

func (c *DAOConverter) R2DB() {

}

func ToDoc(v interface{}) (doc interface{}, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

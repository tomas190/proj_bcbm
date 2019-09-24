package internal

import (
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type DTOConverter struct{}

func (c *DTOConverter) U2Msg(p Player) msg.UserInfo {
	// id, name, img, score := p.GetPlayerBasic()
	id, name, _, score := p.GetPlayerBasic()

	win, bet := p.GetPlayerAccount()
	uMsg := msg.UserInfo{
		UserID:    id,
		NickName:  name,
		Avatar:    "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png",
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
		Counter:    0, // fixme
		Statistics: stat,
	}

	return bMsg
}

func (c *DTOConverter) RSBMsg(userWin float64, autoBetAmount, userBalance float64, dl Dealer) msg.RoomStatusB {
	dl.Bankers[0].GetPlayerBasic()
	bMsg := msg.RoomStatusB{
		Status:        dl.Status,
		Counter:       0, // fixme
		EndTime:       dl.ddl,
		Result:        dl.res,
		BankerWin:     dl.bankerWin,
		WinMoney:      userWin,
		BankerMoney:   dl.bankerMoney,
		AutoBetAmount: autoBetAmount, // 若不可续投则为0
		PlayerMoney:   userBalance,
		ServerTime:    uint32(time.Now().Unix()),
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

// 玩家下注
func (c *DAOConverter) Bet2DB(u User, betM msg.Bet) BetDB {
	user := c.U2DB(u)
	aStr := constant.AreaName[betM.Area]
	cAmount := constant.ChipSize[betM.Chip]
	bdb := BetDB{User: user, Area: betM.Area, AreaStr: aStr, Chip: betM.Chip, ChipAmount: cAmount}
	return bdb
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

package internal

import "proj_bcbm/src/server/msg"

type DTOConverter struct {}

func (c *DTOConverter) U2Msg(u User) msg.UserInfo {
	uMsg := msg.UserInfo{
		UserID:u.UserID,
		NickName:u.NickName,
		Avatar:u.Avatar,
		Money:u.Balance,
	}

	return uMsg
}

func (c *DTOConverter) R2Msg(dl Dealer) msg.RoomInfo {
	rMsg := msg.RoomInfo{
		RoomID:dl.RoomID,
		MinBet:dl.MinBet,
		MaxBet:dl.MaxBet,
		MinLimit:dl.MinLimit,
		Status: dl.Status,
		EndTime:0,
		History:dl.History,
		HisStatistic:dl.HisStatistic,
	}

	return rMsg
}

type DAOConverter struct {}

func (c *DAOConverter) U2Bson() {

}

func (c *DAOConverter) R2Bson() {

}

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

func (c *DTOConverter) R2Msg(r Room) msg.RoomInfo {
	rMsg := msg.RoomInfo{
		RoomID:r.RoomID,
		MinBet:r.MinBet,
		MaxBet:r.MaxBet,
		MinLimit:r.MinLimit,
		Status: r.Status,
		EndTime:r.EndTime,
		History:r.History,
		HisStatistic:r.HisStatistic,
	}

	return rMsg
}

type DAOConverter struct {}

func (c *DAOConverter) U2Bson() {

}

func (c *DAOConverter) R2Bson() {

}

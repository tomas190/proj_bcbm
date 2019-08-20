package common

import (
	"github.com/golang/protobuf/proto"
	"proj_bcbm/src/server/msg"
)

func TransIDToMsg(id uint16) proto.Message {
	var resp proto.Message
	switch id {
	case 0: resp = &msg.Error{}
	case 1: resp = &msg.Ping{}
	case 2: resp = &msg.Pong{}
	case 3: resp = &msg.LoginTest{}
	case 4: resp = &msg.Login{}
	case 5: resp = &msg.LoginR{}
	case 6: resp = &msg.Logout{}
	case 7: resp = &msg.LogoutR{}
	case 8: resp = &msg.JoinRoom{}
	case 9: resp = &msg.JoinRoomR{}
	case 10: resp = &msg.LeaveRoom{}
	case 11: resp = &msg.LeaveRoomR{}
	case 12: resp = &msg.GrabBanker{}
	case 13: resp = &msg.AutoBet{}
	case 14: resp = &msg.AutoBetR{}
	case 15: resp = &msg.Bet{}
	case 16: resp = &msg.BetR{}
	case 17: resp = &msg.BetInfoB{}
	case 18: resp = &msg.BankersB{}
	case 19: resp = &msg.PlayersB{}
	}

	return resp
}

var msgType2ID = map[string]uint16{
	"*msg.Error" : 0,
	"*msg.Ping" : 1,
	"*msg.Pong" : 2,
	"*msg.LoginTest" : 3,
	"*msg.Login" : 4,
	"*msg.LoginR" : 5,
	"*msg.Logout" : 6,
	"*msg.LogoutR" : 7,
	"*msg.JoinRoom" : 8,
	"*msg.JoinRoomR" : 9,
	"*msg.LeaveRoom" : 10,
	"*msg.LeaveRoomR" : 11,
	"*msg.GrabBanker" : 12,
	"*msg.AutoBet" : 13,
	"*msg.AutoBetR" : 14,
	"*msg.Bet" : 15,
	"*msg.BetR" : 16,
	"*msg.BetInfoB" : 17,
	"*msg.BankersB" : 18,
	"*msg.PlayersB" : 19,
}

func transMsgToID(t string) uint16 {
	if id, ok := msgType2ID[t]; ok {
		return id
	}
	return 1024
}



package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"reflect"
)

func (dl *Dealer) handleBet(args []interface{}) {
	m := args[0].(*msg.Bet)
	a := args[1].(gate.Agent)
	au := a.UserData().(*User)

	log.Debug("筹码信息 %+v", m)

	cs := constant.ChipSize[m.Chip]
	if au.Balance < cs {
		errorResp(a, msg.ErrorCode_InsufficientBalanceBet, "没钱玩啥")
	}

	// 够 记录
	// 在中心服务器减钱，拿返回的余额
	dl.AreaBets[m.Area] = dl.AreaBets[m.Area] + cs
	dl.UserBets[au.UserID][m.Area] = dl.UserBets[au.UserID][m.Area] + cs
	if dl.Status == constant.RSBetting {
		resp := &msg.BetInfoB{
			Area:        m.Area,
			Chip:        m.Chip,
			AreaTotal:   dl.AreaBets[m.Area],
			PlayerTotal: dl.UserBets[au.UserID][m.Area],
			PlayerID:    au.UserID,
			Money:       au.Balance - dl.UserBets[au.UserID][m.Area], // fixme
		}

		dl.Broadcast(resp)
	} else {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
	}
}

func (dl *Dealer) handleGrabBanker(args []interface{}) {
	m := args[0].(*msg.GrabBanker)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	fmt.Println("上庄", m, au.Balance)

	resp := &msg.BankersB{}
	a.WriteMsg(resp)
}

func (dl *Dealer) handleAutoBet(args []interface{}) {
	m := args[0].(*msg.AutoBet)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	fmt.Println("续投", m, au.Balance)

	resp := &msg.AutoBetR{}
	a.WriteMsg(resp)
}

func (dl *Dealer) handleLeaveRoom(args []interface{}) {
	m := args[0].(*msg.LeaveRoom)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	fmt.Println("离房", m, au.Balance)

	resp := &msg.LeaveRoomR{}
	a.WriteMsg(resp)
}

// 玩家列表
func (dl *Dealer) getPlayerInfoResp() []*msg.UserInfo {
	u1 := mockUserInfo(8976784)
	u2 := mockUserInfo(7829401)

	converter := DTOConverter{}
	userInfo1 := converter.U2Msg(*u1)
	userInfo2 := converter.U2Msg(*u2)

	var playerInfoResp []*msg.UserInfo
	playerInfoResp = append(playerInfoResp, &userInfo1, &userInfo2)

	return playerInfoResp
}

// 庄家列表
func (dl *Dealer) getBankerInfoResp() []*msg.UserInfo {
	return nil
}

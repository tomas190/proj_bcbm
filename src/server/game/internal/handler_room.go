package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"reflect"
	"time"
)

func (dl *Dealer) handleBet(args []interface{}) {
	m := args[0].(*msg.Bet)
	a := args[1].(gate.Agent)
	au := a.UserData().(*User)

	if dl.Status == constant.RSBetting {
		log.Debug("筹码信息 %+v", m)

		cs := constant.ChipSize[m.Chip]
		if au.Balance < cs {
			errorResp(a, msg.ErrorCode_InsufficientBalanceBet, "没钱玩啥")
		}

		dl.AreaBets[m.Area] = dl.AreaBets[m.Area] + cs
		dl.UserBets[au.UserID][m.Area] = dl.UserBets[au.UserID][m.Area] + cs

		c4c.UserLoseScore(au.UserID, -cs, func(data *User) {
			log.Debug("用户 %+v 下注后余额 %+v", data.UserID, data.Balance)
			fmt.Println("#########区域总和", au.NickName, dl.AreaBets)
			fmt.Println("@@@@@@@@@玩家各区域投注", au.NickName, dl.UserBets[au.UserID])

			au.Balance = data.Balance

			resp := &msg.BetInfoB{
				Area:        m.Area,
				Chip:        m.Chip,
				AreaTotal:   dl.AreaBets[m.Area],
				PlayerTotal: dl.UserBets[au.UserID][m.Area],
				PlayerID:    au.UserID,
				Money:       au.Balance,
			}
			dl.Broadcast(resp)
		})
	} else {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
	}
}

func (dl *Dealer) handlePlayers(args []interface{}) {
	m := args[0].(*msg.Players)
	a := args[1].(gate.Agent)
	au := a.UserData().(*User)

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	resp := &msg.PlayersR{
		Players:    dl.getPlayerInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	a.WriteMsg(resp)
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

	resp := &msg.LeaveRoomR{
		User: &msg.UserInfo{
			UserID:   au.UserID,
			Avatar:   au.Avatar,
			NickName: au.NickName,
			Money:    au.Balance,
		},
		Rooms:      Mgr.GetRoomsInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	a.WriteMsg(resp)
}

// 玩家列表
func (dl *Dealer) getPlayerInfoResp() []*msg.UserInfo {
	var playerInfoResp []*msg.UserInfo
	for _, u := range dl.Users {
		converter := DTOConverter{}
		uInfo := converter.U2Msg(*u)
		playerInfoResp = append(playerInfoResp, &uInfo)

	}

	return playerInfoResp
}

func (dl *Dealer) getBankerInfoResp() []*msg.UserInfo {
	var bankerInfoResp []*msg.UserInfo
	for _, b := range dl.Bankers {
		converter := DTOConverter{}
		buInfo := converter.U2Msg(b)
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
		fmt.Println(buInfo)
		bankerInfoResp = append(bankerInfoResp, &buInfo)
	}

	fmt.Println(bankerInfoResp)
	return bankerInfoResp
}

package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/patrickmn/go-cache"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"reflect"
	"sort"
	"time"
)

func (dl *Dealer) handleBet(args []interface{}) {
	m := args[0].(*msg.Bet)
	a := args[1].(gate.Agent)
	au := a.UserData().(*User)

	if m.Chip == 0 {
		errorResp(a, msg.ErrorCode_InsufficientBalanceBet, "余额不足")
		return
	}

	if dl.Status != constant.RSBetting {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	// 最后一秒
	if dl.Status == constant.RSBetting && dl.counter > 14 {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	_, found := ca.Get(fmt.Sprintf("%+v-bet", au.UserID))
	if found {
		errorResp(a, msg.ErrorCode_ServerBusy, "服务器忙")
		return
	} else {
		cs := constant.ChipSize[m.Chip]
		if au.Balance < cs {
			errorResp(a, msg.ErrorCode_InsufficientBalanceBet, "余额不足")
			return
		}

		uuid := util.UUID{}
		order := uuid.GenUUID()

		ca.Set(fmt.Sprintf("%+v-bet", au.UserID), order, cache.DefaultExpiration)
		c4c.UserLoseScore(au.UserID, -cs, order, "", func(data *User) {
			// log.Debug("用户 %+v 下注后余额 %+v", data.UserID, data.Balance)
			au.BalanceLock.Lock()
			au.Balance = data.Balance
			au.BalanceLock.Unlock()

			// 所有用户在该区域历史投注+机器人在该区域历史投注+当前用户投注
			dl.AreaBets[m.Area] = dl.AreaBets[m.Area] + cs
			// 当前用户在该区域的历史投注+当前用户投注
			dl.UserBets[au.UserID][m.Area] = dl.UserBets[au.UserID][m.Area] + cs
			// 用户具体投注信息
			dl.UserBetsDetail[au.UserID] = append(dl.UserBetsDetail[au.UserID], *m)

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
		ca.Delete(fmt.Sprintf("%+v-bet", au.UserID))
		// 记录玩家投注信息
		return
	}
}

func (dl *Dealer) handleAutoBet(args []interface{}) {
	m := args[0].(*msg.AutoBet)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)
	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	if dl.Status != constant.RSBetting {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	if dl.Status == constant.RSBetting && dl.counter > 14 {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	for _, b := range dl.AutoBetRecord[au.UserID] {
		bet := b
		cs := constant.ChipSize[bet.Chip]

		// 所有用户在该区域历史投注+机器人在该区域历史投注+当前用户投注
		dl.AreaBets[bet.Area] = dl.AreaBets[bet.Area] + cs
		// 当前用户在该区域的历史投注+当前用户投注
		dl.UserBets[au.UserID][bet.Area] = dl.UserBets[au.UserID][bet.Area] + cs
		// 用户具体投注信息
		// dl.UserBetsDetail[au.UserID] = append(dl.UserBetsDetail[au.UserID], bet)

		// fixme 为了前端动画代价有点大
		uuid := util.UUID{}
		order := uuid.GenUUID()
		c4c.UserLoseScore(au.UserID, -cs, order, dl.RoundID, func(data *User) {
			// log.Debug("用户 %+v 下注后余额 %+v", data.UserID, data.Balance)
			au.Balance = data.Balance

			resp := &msg.BetInfoB{
				Area:        bet.Area,
				Chip:        bet.Chip,
				AreaTotal:   dl.AreaBets[bet.Area],
				PlayerTotal: dl.UserBets[au.UserID][bet.Area],
				PlayerID:    au.UserID,
				Money:       au.Balance,
			}
			dl.Broadcast(resp)
		})

		// 暂时延迟处理
		time.Sleep(time.Millisecond * 3)
	}
	dl.UserAutoBet[au.UserID] = true
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

	if au.Balance < constant.BankerMinBar {
		errorResp(a, msg.ErrorCode_InsufficientBalanceGrabBanker, "金币未达到50000")
		return
	}

	// 当前庄家不变，其他清空
	curBanker := dl.Bankers[0]
	dl.Bankers = []Player{}
	dl.Bankers = append(dl.Bankers, curBanker)

	// 如果玩家不在列表中，将玩家排在最前面
	flag := false
	for _, b := range dl.Bankers {
		uID, _, _, _ := b.GetPlayerBasic()
		if uID == au.UserID {
			flag = true
		}
	}

	if flag == false {
		dl.Bankers = append(dl.Bankers, au)
	}

	resp := &msg.BankersB{
		Banker:     dl.getBankerInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	dl.Broadcast(resp)
}

func (dl *Dealer) handleLeaveRoom(args []interface{}) {
	m := args[0].(*msg.LeaveRoom)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	// fixme 不能直接删除
	dl.Users.Delete(au.UserID)

	// todo 玩家离开房间后 清空续投 需要结算
	//dl.Bankers
	//dl.UserAutoBet
	//dl.UserBetsDetail
	dl.AutoBetRecord[au.UserID] = nil

	resp := &msg.LeaveRoomR{
		User: &msg.UserInfo{
			UserID: au.UserID,
			// Avatar:   au.Avatar,
			Avatar:   "https://cdn1.iconfinder.com/data/icons/avatars-1-5/136/81-512.png",
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
	converter := DTOConverter{}

	dl.Users.Range(func(key, value interface{}) bool {
		u := value.(*User)
		uInfo := converter.U2Msg(*u)
		playerInfoResp = append(playerInfoResp, &uInfo)
		return true
	})

	for _, b := range dl.Bots {
		pInfo := converter.U2Msg(*b)
		playerInfoResp = append(playerInfoResp, &pInfo)
	}

	// 先按照获胜局数排序
	sort.Slice(playerInfoResp, func(i, j int) bool {
		return playerInfoResp[i].WinCount > playerInfoResp[j].WinCount
	})

	// 拿到赌神
	betGod := playerInfoResp[0]

	// 再把其余人按照投注数排序
	playerInfoResp = playerInfoResp[1:]
	sort.Slice(playerInfoResp, func(i, j int) bool {
		return playerInfoResp[i].BetAmount > playerInfoResp[j].BetAmount
	})

	// 组合在一起
	playerInfoResp = append([]*msg.UserInfo{betGod}, playerInfoResp...)

	return playerInfoResp
}

func (dl *Dealer) getBankerInfoResp() []*msg.UserInfo {
	var bankerInfoResp []*msg.UserInfo
	for _, b := range dl.Bankers {
		converter := DTOConverter{}
		buInfo := converter.U2Msg(b)
		bankerInfoResp = append(bankerInfoResp, &buInfo)
	}

	return bankerInfoResp
}

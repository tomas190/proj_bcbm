package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/patrickmn/go-cache"
	"gopkg.in/mgo.v2/bson"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
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
		if dl.roomBonusLimit(m.Area) < cs || dl.dynamicBonusLimit(m.Area) < cs {
			errorResp(a, msg.ErrorCode_ReachTableLimit, "到达限红")
			return
		}

		uuid := util.UUID{}
		order := uuid.GenUUID()

		// fixme 暂时延迟处理
		rd := util.Random{}
		delay := rd.RandInRange(0, 100)
		time.Sleep(time.Millisecond * time.Duration(delay))
		ca.Set(fmt.Sprintf("%+v-bet", au.UserID), order, cache.DefaultExpiration)

		// 所有用户在该区域历史投注+机器人在该区域历史投注+当前用户投注
		dl.AreaBets[m.Area] = dl.AreaBets[m.Area] + cs
		dl.DownBetArea[m.Area] = dl.DownBetArea[m.Area] + cs
		// 当前用户在该区域的历史投注+当前用户投注
		dl.UserBets[au.UserID][m.Area] = dl.UserBets[au.UserID][m.Area] + cs
		// 用户具体投注信息
		dl.UserBetsDetail[au.UserID] = append(dl.UserBetsDetail[au.UserID], *m)

		au.DownBetTotal += constant.ChipSize[m.Chip]
		dl.TotalDownMoney += constant.ChipSize[m.Chip]
		au.Balance -= constant.ChipSize[m.Chip]

		dl.UserIsDownBet[au.UserID] = true
		au.IsAction = true

		resp := &msg.BetInfoB{
			Area:        m.Area,
			Chip:        m.Chip,
			AreaTotal:   dl.AreaBets[m.Area],
			PlayerTotal: dl.UserBets[au.UserID][m.Area],
			PlayerID:    au.UserID,
			Money:       au.Balance,
		}
		dl.Broadcast(resp)

		//log.Debug("<<=====>>玩家金额: %v", au.Balance)
		// fixme 暂时延迟处理
		time.Sleep(6 * time.Millisecond)
		ca.Delete(fmt.Sprintf("%+v-bet", au.UserID))
		// 记录玩家投注信息
		return
	}
}

func (dl *Dealer) handleAutoBet(args []interface{}) {
	m := args[0].(*msg.AutoBet)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)
	log.Debug("======>> recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)
	var LimitRed bool

	if dl.Status != constant.RSBetting {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	if dl.Status == constant.RSBetting && dl.counter > 14 {
		errorResp(au.ConnAgent, msg.ErrorCode_NotInBetting, "当前不是下注状态")
		return
	}

	var autoBetAmounts = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

	var cs float64
	for _, b := range dl.AutoBetRecord[au.UserID] {
		cs += constant.ChipSize[b.Chip]
		//log.Debug("总投注:%v", cs)

		if dl.roomBonusLimit(b.Area) < cs || dl.dynamicBonusLimit(b.Area) < cs {
			LimitRed = true
			errorResp(a, msg.ErrorCode_ContinueBetError, "续投失败")
			return
		}
		if LimitRed == true {
			return
		}
	}

	for _, b := range dl.AutoBetRecord[au.UserID] {
		bet := b
		cs := constant.ChipSize[bet.Chip]

		// 所有用户在该区域历史投注+机器人在该区域历史投注+当前用户投注
		dl.AreaBets[bet.Area] = dl.AreaBets[bet.Area] + cs
		dl.DownBetArea[bet.Area] = dl.DownBetArea[bet.Area] + cs
		// 当前用户在该区域的历史投注+当前用户投注
		dl.UserBets[au.UserID][bet.Area] = dl.UserBets[au.UserID][bet.Area] + cs

		autoBetAmounts[bet.Area] += cs

		au.DownBetTotal += cs
		dl.TotalDownMoney += cs
		au.Balance -= cs

		//log.Debug("续投成功 ~~~~: %v", au.DownBetTotal)
	}

	resp := &msg.AutoBetB{
		UserID:      au.UserID,
		Amounts:     autoBetAmounts,
		AreaTotal:   dl.AreaBets,
		PlayerTotal: dl.UserBets[au.UserID],
		Money:       au.Balance,
	}

	dl.Broadcast(resp)

	dl.UserIsDownBet[au.UserID] = true
	au.IsAction = true
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

	// 取消上庄申请
	if m.LockMoney == constant.CancelGrab {
		dl.cancelGrabBanker(au.UserID)
		return
	}

	// 申请下庄，先标记，一局结束之后轮换
	if m.LockMoney == constant.DownBanker {
		dl.DownBanker = true
		return
	}

	if m.LockMoney < constant.BankerMinBar || m.LockMoney > au.Balance {
		errorResp(a, msg.ErrorCode_InsufficientBalanceGrabBanker, "上庄金币不足")
		return
	}

	var newBankers []Player
	newBankers = append(newBankers, dl.Bankers[0])

	//只清理机器人 不清理真实玩家
	for i := range dl.Bankers {
		if i != 0 {
			switch dl.Bankers[i].(type) {
			case User:
				newBankers = append(newBankers, dl.Bankers[i])
			default:
			}
		}
	}

	dl.Bankers = newBankers

	// 如果玩家已经在列表中，直接返回
	for _, b := range dl.Bankers {
		uID, _, _, _ := b.GetPlayerBasic()
		if uID == au.UserID {
			return
		}
	}

	//uuid := util.UUID{}
	// 上庄
	//log.Debug("<<===== 上庄金额: %v =====>>", m.LockMoney)
	order := bson.NewObjectId().Hex()
	c4c.ChangeBankerStatus(au.UserID, constant.BSGrabbingBanker, m.LockMoney, order, dl.RoundID, func(data *User) {
		// 更新房间玩家列表中的玩家余额
		dl.Users.Range(func(key, value interface{}) bool {
			if key == au.UserID {
				u := value.(*User)
				u.Status = constant.BSGrabbingBanker
				u.BankerBalance = data.BankerBalance
				u.Balance = data.Balance
			}
			return true
		})
	})

	au.BankerBalance = 5000
	bUser := User{
		UserID:        au.UserID,
		Balance:       au.Balance - au.BankerBalance,
		BankerBalance: au.BankerBalance,
		Avatar:        au.Avatar,
		NickName:      au.NickName,
	}

	dl.Bankers = append(dl.Bankers, bUser)

	resp := &msg.BankersB{
		Banker:     dl.getBankerInfoResp(),
		ServerTime: uint32(time.Now().Unix()),
	}

	log.Debug("<--- 庄家列表更新 --->")
	dl.Broadcast(resp)
}

func (dl *Dealer) handleLeaveRoom(args []interface{}) {
	m := args[0].(*msg.LeaveRoom)
	a := args[1].(gate.Agent)

	au := a.UserData().(*User)

	//if dl.DownBetTotal > 0 {
	//	resp := &msg.LeaveRoomR{
	//		User: &msg.UserInfo{
	//			UserID:   au.UserID,
	//			Avatar:   au.Avatar,
	//			NickName: au.NickName,
	//			Money:    au.Balance,
	//		},
	//		Rooms:      Mgr.GetRoomsInfoResp(),
	//		ServerTime: uint32(time.Now().Unix()),
	//	}
	//
	//	a.WriteMsg(resp)
	//	return
	//}

	log.Debug("recv %+v, addr %+v, %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m, au.UserID)

	math := util.Math{}
	uBets, _ := math.SumSliceFloat64(dl.UserBets[au.UserID]).Float64()
	if uBets == 0 {
		au.winCount = 0
		au.betAmount = 0
		dl.UserIsDownBet[au.UserID] = false
		au.IsAction = false
		dl.UserBets[au.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
		dl.Users.Delete(au.UserID)
	} else {
		var exist bool
		for _, v := range dl.UserLeave {
			if v == au.UserID {
				exist = true
			}
		}
		if exist == false {
			dl.UserLeave = append(dl.UserLeave, au.UserID)
		}
	}

	dl.AutoBetRecord[au.UserID] = nil

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

func (dl *Dealer) cancelGrabBanker(userID uint32) {
	for i, b := range dl.Bankers {
		banker := b
		uID, _, _, _ := banker.GetPlayerBasic()
		bankerBalance := banker.GetBankerBalance()
		if userID == uID {
			// 移除玩家
			dl.Bankers = append(dl.Bankers[:i], dl.Bankers[i+1:]...)

			if len(dl.Bankers) < 2 {
				nextB := dl.NextBotBanker()
				dl.Bankers = append(dl.Bankers, nextB)
				dl.Bots = append(dl.Bots, &nextB)
			}

			//uuid := util.UUID{}
			order := bson.NewObjectId().Hex()
			// 玩家取消申请上庄
			c4c.ChangeBankerStatus(userID, constant.BSNotBanker, -bankerBalance, order, dl.RoundID, func(data *User) {
				// 更新房间玩家列表中的玩家余额
				dl.Users.Range(func(key, value interface{}) bool {
					if key == uID {
						u := value.(*User)
						u.Status = constant.BSNotBanker
						u.BankerBalance = data.BankerBalance
						u.Balance = data.Balance
					}
					return true
				})
				log.Debug("<---玩家取消申请上庄--->")
			})

			// 更新房间玩家列表中的玩家余额
			dl.Users.Range(func(key, value interface{}) bool {
				if key == uID {
					u := value.(*User)
					resp := &msg.BankersB{
						Banker: dl.getBankerInfoResp(),
						UpdateBanker: &msg.UserInfo{
							Money:  u.Balance,
							UserID: u.UserID,
						},
						ServerTime: uint32(time.Now().Unix()),
					}
					log.Debug("<--- 庄家列表更新 --->")
					dl.Broadcast(resp)
				}
				return true
			})
		}
	}
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

// 房间剩余限红
func (dl *Dealer) roomBonusLimit(area uint32) float64 {
	return constant.RoomMaxBonus/constant.AreaX[area] - dl.AreaBets[area]
}

// 区域剩余限红
func (dl *Dealer) dynamicBonusLimit(area uint32) float64 {
	var sum float64
	for i, v := range dl.AreaBets {
		if uint32(i) == area {
		} else {
			sum += v
		}
	}
	return (dl.bankerMoney+sum)/(constant.AreaX[area]-1) - dl.AreaBets[area]
}

// 其他区域投注数总和

func (dl *Dealer) getBankerInfoResp() []*msg.UserInfo {
	var bankerInfoResp []*msg.UserInfo
	for i, b := range dl.Bankers {
		converter := DTOConverter{}
		buInfo := converter.Banker2Msg(b)
		if i == 0 {
			buInfo.BankerMoney = dl.bankerMoney
		}
		bankerInfoResp = append(bankerInfoResp, &buInfo)
	}

	return bankerInfoResp
}

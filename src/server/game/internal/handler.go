package internal

import (
	"github.com/name5566/leaf/gate"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"reflect"
	"time"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.Login{}, handleLogin)
	handlerReg(&msg.Logout{}, handleLogout)

	handlerReg(&msg.JoinRoom{}, handleJoinRoom)

	handlerReg(&msg.Bet{}, handleRoomEvent)
	handlerReg(&msg.Players{}, handleRoomEvent)
	handlerReg(&msg.LeaveRoom{}, handleRoomEvent)
	handlerReg(&msg.GrabBanker{}, handleRoomEvent)
	handlerReg(&msg.AutoBet{}, handleRoomEvent)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

/*************************************

	普通事件

 *************************************/

func handlePing(args []interface{}) {
	// m := args[0].(*msg.Ping)
	a := args[1].(gate.Agent)
	// log.Debug("recv Ping %+v", a.RemoteAddr())
	a.WriteMsg(&msg.Pong{})
}

func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login)
	// m := randomLoginMsg()
	a := args[1].(gate.Agent)
	userID := m.GetUserID()
	log.Debug("处理用户登录请求 %+v", userID)
	v, ok := Mgr.UserRecord.Load(userID)
	if ok {
		u := v.(*User)
		if u.ConnAgent == a { // 用户和连接都相同
			log.Debug("rpcUserLogin 同一用户相同连接重复登录")
			errorResp(a, msg.ErrorCode_UserRepeatLogin, "重复登录")
			return
		} else { // 用户存在，但连接不同
			rid := Mgr.UserRoom[u.UserID]
			r, _ := Mgr.RoomRecord.Load(rid)
			if r != nil {
				dl := r.(*Dealer)
				for i, lu := range dl.UserLeave {
					user := lu
					// 把玩家从掉线列表中移除
					if user == u.UserID {
						dl.UserLeave = append(dl.UserLeave[:i], dl.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家掉线记录~")
						break
					}
				}
			}

			err := Mgr.ReplaceUserAgent(userID, a)
			if err != nil {
				log.Error("用户连接替换错误", err)
			}

			c4c.UserLoginCenter(userID, m.Password, m.Token, func(u *User) {})

			v, _ := Mgr.UserRecord.Load(userID)
			u := v.(*User)

			resp := &msg.LoginR{
				User: &msg.UserInfo{
					UserID:   u.UserID,
					Avatar:   u.Avatar,
					Money:    u.Balance,
					NickName: u.NickName,
				},
				Rooms:      Mgr.GetRoomsInfoResp(),
				ServerTime: uint32(time.Now().Unix()),
			}

			if rID, ok := Mgr.UserRoom[userID]; ok {
				resp.RoomID = rID
			}
			log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
			a.WriteMsg(resp)
		}
	} else if !Mgr.agentExist(a) {
		c4c.UserLoginCenter(userID, m.Password, m.Token, func(u *User) {
			resp := &msg.LoginR{
				User: &msg.UserInfo{
					UserID:   u.UserID,
					Avatar:   u.Avatar,
					NickName: u.NickName,
					Money:    u.Balance,
				},
				Rooms:      Mgr.GetRoomsInfoResp(),
				ServerTime: uint32(time.Now().Unix()),
			}

			// 重新绑定信息
			u.ConnAgent = a
			a.SetUserData(u)

			go func() {
				err := db.RUserInfo(u.UserID)
				if err != nil {
					converter := DAOConverter{}
					udb := converter.U2DB(*u)
					errC := db.CUserInfo(udb)
					if errC != nil {
						log.Error("玩家信息保存错误 %+v", errC)
					}
				}
			}()

			Mgr.UserRecord.Store(u.UserID, u)
			log.Debug("<----login 登录 resp---->%+v", resp.User.UserID)
			a.WriteMsg(resp)
		})
	}
}

func handleLogout(args []interface{}) {
	// m := args[0].(*msg.Logout)
	a := args[1].(gate.Agent)

	au, ok := a.UserData().(*User)

	if ok {
		rid := Mgr.UserRoom[au.UserID]
		v, _ := Mgr.RoomRecord.Load(rid)
		if v != nil {
			dl := v.(*Dealer)
			math := util.Math{}
			uBets, _ := math.SumSliceFloat64(dl.UserBets[au.UserID]).Float64() // 获取下注金额
			log.Debug("玩家下注金额为:%v,是否行動:%v", uBets, au.IsAction)
			if au.IsAction == false || uBets == 0 {
				c4c.UserLogoutCenter(au.UserID, func(data *User) {
					Mgr.UserRecord.Delete(au.UserID)
					resp := &msg.LogoutR{}
					a.WriteMsg(resp)
					a.Close()
				})
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
				resp := &msg.LogoutR{}
				a.WriteMsg(resp)
			}
		}
	}
}

/*************************************

	大厅事件-加入房间

 *************************************/

func handleJoinRoom(args []interface{}) {
	m := args[0].(*msg.JoinRoom)
	a := args[1].(gate.Agent)

	log.Debug("recv %+v, addr %+v, %+v", reflect.TypeOf(m), a.RemoteAddr(), m)

	au, ok := a.UserData().(*User)
	if ok {
		Mgr.PlayerJoinRoom(m.RoomID, au)
	}
}

/*************************************

	房间事件-投注、续投、上庄、离开房间

 *************************************/

func handleRoomEvent(args []interface{}) {
	a := args[1].(gate.Agent)
	u, ok := a.UserData().(*User)
	_, logged := Mgr.UserRecord.Load(u.UserID)
	_, inRoom := Mgr.UserRoom[u.UserID]
	//log.Debug("<----game 房间事件 %v %v %v---->", u.UserID, reflect.TypeOf(args[0]), args[0])

	if ok && logged && inRoom {
		// 找到玩家房间
		roomID, ok := Mgr.UserRoom[u.UserID]
		ok = true
		if ok {
			v, _ := Mgr.RoomRecord.Load(roomID)
			dealer := v.(*Dealer)
			//log.Debug("当前房间状态 %v", dealer.Status)
			switch t := args[0].(type) {
			case *msg.Bet:
				dealer.handleBet(args)
			case *msg.Players:
				dealer.handlePlayers(args)
			case *msg.GrabBanker:
				dealer.handleGrabBanker(args)
			case *msg.AutoBet:
				dealer.handleAutoBet(args)
			case *msg.LeaveRoom:
				dealer.handleLeaveRoom(args)
			default:
				log.Error("房间事件无法识别", t)
			}
		}
	} else {
		errorResp(a, msg.ErrorCode_UserNotInRoom, "")
	}
}

func errorResp(a gate.Agent, err msg.ErrorCode, detail string) {
	log.Debug("<----game 错误resp %+v---->", err)
	a.WriteMsg(&msg.Error{Code: err, Detail: detail})
}

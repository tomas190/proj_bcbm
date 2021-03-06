package internal

import (
	"github.com/name5566/leaf/gate"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<----新连接---->")

	u := &User{}
	u.Init()
	u.ConnAgent = a  // 保存连接到用户信息
	a.SetUserData(u) // 附加用户信息到连接
}

// 玩家掉线
func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	au, ok := a.UserData().(*User)

	if ok && au.ConnAgent == a {
		log.Debug("玩家 %+v 主动断开连接...", au.UserID)
		au.betAmount = 0
		au.winCount = 0

		rid := Mgr.UserRoom[au.UserID]
		v, _ := Mgr.RoomRecord.Load(rid)
		if v != nil {
			dl := v.(*Dealer)
			math := util.Math{}
			uBets, _ := math.SumSliceFloat64(dl.UserBets[au.UserID]).Float64() // 获取下注金额
			log.Debug("rpcCloseAgent 玩家下注金额:%v", uBets)
			log.Debug("rpcCloseAgent au.IsAction:%v", au.IsAction)
			if au.IsAction == false || uBets == 0 {
				dl.Users.Delete(au.UserID)
				delete(Mgr.UserRoom, au.UserID)
				dl.DeleteRoomRecord()
				c4c.UserLogoutCenter(au.UserID, func(data *User) {
					dl.AutoBetRecord[au.UserID] = nil
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
						log.Debug("rpcCloseAgent 玩家已存在UserLeave:%v", au.UserID)
					}
				}
				if exist == false {
					dl.UserLeave = append(dl.UserLeave, au.UserID)
					log.Debug("rpcCloseAgent 添加离线UserLeave:%v,%v", au.UserID, dl.UserLeave)
				}
			}
		} else {
			c4c.UserLogoutCenter(au.UserID, func(data *User) {
				Mgr.UserRecord.Delete(au.UserID)
				resp := &msg.LogoutR{}
				a.WriteMsg(resp)
				a.Close()
			})
		}
	}
}

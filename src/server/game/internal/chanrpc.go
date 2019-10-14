package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
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
	u.ConnAgent = a  // 保存连接到用户信息
	a.SetUserData(u) // 附加用户信息到连接
}

// 玩家掉线
func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	au, ok := a.UserData().(*User)

	if ok {
		log.Debug("玩家 %+v 主动断开连接...", au.UserID)
		ca.Delete(fmt.Sprintf("%+v-betAmount", au.UserID))
		ca.Delete(fmt.Sprintf("%+v-winCount", au.UserID))

		rid := Mgr.UserRoom[au.UserID]
		v, _ := Mgr.RoomRecord.Load(rid)
		dl := v.(*Dealer)

		math := util.Math{}
		uBets, _ := math.SumSliceFloat64(dl.UserBets[au.UserID]).Float64()
		if uBets == 0 {
			dl.Users.Delete(au.UserID)
		} else {
			dl.UserLeave = append(dl.UserLeave, au.UserID)
		}

		dl.AutoBetRecord[au.UserID] = nil
		ca.Delete(fmt.Sprintf("%+v-betAmount", au.UserID))
		ca.Delete(fmt.Sprintf("%+v-winCount", au.UserID))

		c4c.UserLogoutCenter(au.UserID, func(data *User) {
			Mgr.UserRecord.Delete(au.UserID)
			resp := &msg.LogoutR{}
			a.WriteMsg(resp)
			a.Close()
		})
	}
}

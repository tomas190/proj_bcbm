package internal

import (
	"fmt"
	"github.com/name5566/leaf/module"
	"github.com/patrickmn/go-cache"
	"gopkg.in/mgo.v2/bson"
	_ "net/http/pprof"
	"proj_bcbm/src/server/base"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"time"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c *Client4Center // 连接中心服的客户端
	db  *MgoC          // 数据库客户端
	ca  *cache.Cache   // 内存缓存
	Mgr = NewHall()
)

type Module struct {
	*module.Skeleton
}

// 模块初始化
func (m *Module) OnInit() {
	m.Skeleton = skeleton

	packageTax = make(map[uint16]float64)

	// 中心服务器
	c4c = NewClient4Center()
	c4c.HeartBeatAndListen()

	// 数据库
	db = NewMgoC(conf.Server.MongoDB)
	err := db.Init()
	if err != nil {
		log.Error("数据库初始化错误 %+v", err)
	}

	go StartHttpServer()

	// 缓存
	ca = cache.New(5*time.Minute, 10*time.Minute)

	// 游戏大厅
	Mgr.OpenCasino()


}

// 模块销毁
func (m *Module) OnDestroy() {
	log.Debug("game模块被销毁...")
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)

	Mgr.UserRecord.Range(func(key, value interface{}) bool {
		p := value.(*User)
		if p.LockMoney > 0 {
			order := bson.NewObjectId().Hex()
			uid := util.UUID{}
			roundId := fmt.Sprintf("%+v-%+v", time.Now().Unix(), uid.GenUUID())
			c4c.UnlockSettlement(p.UserID, p.LockMoney, order, roundId)
		}
		c4c.UserLogoutCenter(p.UserID, func(data *User) {})
		return true
	})
}

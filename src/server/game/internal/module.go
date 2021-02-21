package internal

import (
	"github.com/name5566/leaf/module"
	"github.com/patrickmn/go-cache"
	_ "net/http/pprof"
	"proj_bcbm/src/server/base"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
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

	packageTax = make(map[uint16]uint8)

	// 中心服务器
	c4c = NewClient4Center()
	//c4c.ReqToken()
	c4c.HeartBeatAndListen()
	//c4c.CronUpdateToken()

	// 数据库
	db = NewMgoC(conf.Server.MongoDB)
	err := db.Init()
	if err != nil {
		log.Error("数据库初始化错误 %+v", err)
	}

	go StartHttpServer()

	winChan = make(chan bool)
	loseChan = make(chan bool)
	downBankerChan = make(chan bool)
	// 缓存
	ca = cache.New(5*time.Minute, 10*time.Minute)

	// 游戏大厅
	Mgr.OpenCasino()

	// net/http/pprof 已经在 init()函数中通过 import 副作用完成默认 Handler 的注册
	//go func() {
	//	err := http.ListenAndServe("localhost:6060", nil)
	//	if err != nil {
	//		log.Debug("性能分析服务启动错误...")
	//	}
	//	log.Debug("性能分析服务...")
	//}()
}

// 模块销毁
func (m *Module) OnDestroy() {
	log.Debug("game模块被销毁...")
	data := &msg.Error{
		Code: msg.ErrorCode_ServerClosed,
	}
	log.Debug("踢出所有客户端 %+v...", data)
}

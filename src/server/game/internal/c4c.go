package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/msg"
	"proj_bcbm/src/server/util"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserCallback func(data *User)

//UserCallback 用户登录回调函数保存
type UserBack struct {
	Data     User
	Callback func(data *User)
}

type Client4Center struct {
	conn          *websocket.Conn
	isServerLogin bool
	userWaitEvent sync.Map

	waitUser map[uint32]*UserBack
}


// 添加互斥锁，防止websocket写并发
var writeMutex sync.Mutex

func NewClient4Center() *Client4Center {
	wsURL := "ws" + strings.TrimPrefix(conf.Server.CenterServer, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	log.Debug("连接中心服 %+v", wsURL)
	if err != nil {
		log.Error("dial error %v", err)
	}

	return &Client4Center{
		//token:         conf.Server.DevName,
		isServerLogin: false,
		conn:          c,
		userWaitEvent: sync.Map{},
		waitUser:      make(map[uint32]*UserBack),
	}
}

/*****************************************

	监听中心服返回数据并处理

******************************************/

func (c4c *Client4Center) HeartBeatAndListen() {
	ticker := time.NewTicker(time.Second * 3)
	log.Debug("发送心跳!")
	go func() {
		for {
			<-ticker.C
			c4c.heartBeat()
		}
	}()

	go func() {
		for {
			msgType, message, err := c4c.conn.ReadMessage()
			if err != nil {
				c4c.conn.Close()

				log.Error("Read msg error %+v", err.Error())

				time.Sleep(10 * time.Second)
				wsURL := "ws" + strings.TrimPrefix(conf.Server.CenterServer, "http")
				c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				log.Debug("重新连接中心服 %+v", wsURL)
				if err != nil {
					log.Error("dial error %v", err)
				} else {
					// 替换连接并重新登录服务器
					c4c.conn = c
					c4c.ServerLoginCenter()
				}
			}

			if msgType == websocket.TextMessage {
				//log.Debug("Msg from center %v", string(message))

				var msgData Server2CenterMsg
				decoder := json.NewDecoder(strings.NewReader(string(message)))
				decoder.UseNumber()

				err := decoder.Decode(&msgData)
				if err != nil {
					log.Error(err.Error())
				}

				var msg Server2CenterMsg
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Error("Json Unmarshal error", err.Error())
				}
				switch msg.Event {
				case constant.CEventServerLogin:
					c4c.onServerLogin(message)
					c4c.onPackageTax(msgData.Data)
				case constant.CEventUserLogin:
					c4c.onUserLogin(message)
					c4c.onUserLoginPac(msgData.Data)
				case constant.CEventUserLogout:
					c4c.onUserLogout(message)
				case constant.CEventUserLoseScore:
					c4c.onUserLoseScore(message)
				case constant.CEventUserWinScore:
					c4c.onUserWinScore(message)
				case constant.CEventChangeBankerStatus:
					c4c.onChangeBankerStatus(message)
				case constant.CEventBankerLoseScore:
					c4c.onBankerLoseScore(message)
				case constant.CEventBankerWinScore:
					c4c.onBankerWinScore(message)
				case constant.CEventNotice:
					c4c.onNotice(message)
				case constant.CEventError:
					c4c.onError(message)
				case constant.MsgLockSettlement:
					c4c.onLockSettlement(message)
				case constant.MsgUnlockSettlement:
					c4c.onUnlockSettlement(message)
				default:
					c4c.onDefault(message)
				}
			}
		}
	}()

	c4c.ServerLoginCenter()
}

func (c4c *Client4Center) onServerLogin(msg []byte) {
	// log.Debug("收到了中心服确认服务器登陆消息 %v", string(msg))
	sLogin := ServerLoginResp{}

	err := json.Unmarshal(msg, &sLogin)
	if err != nil {
		log.Error("解析服务器登录返回数据错误:%v", err)
	}

	data := sLogin.Data
	status := data.Status
	// code := data.Code
	taxPercent := data.Msg.PlatformTaxPercent

	c4c.isServerLogin = true

	SendTgMessage("启动成功")

	log.Debug("服务器登陆 %+v 税率 %%%+v ...", status, taxPercent)
}

// 收到用户登录返回之后
func (c4c *Client4Center) onPackageTax(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if ok {
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {

			msginfo := data["msg"].(map[string]interface{})
			fmt.Println("globals:", msginfo["globals"], reflect.TypeOf(msginfo["globals"]))

			globals := msginfo["globals"].([]interface{})
			fmt.Println("allList", globals)
			for k, v := range globals {
				fmt.Println(k, v)
				info := v.(map[string]interface{})
				fmt.Println("package_id", info["package_id"])

				var nPackage uint16
				var nTax float64

				jsonPackageId, err := info["package_id"].(json.Number).Int64()
				if err != nil {
					log.Fatal("onPackageTax jsonPackageId:%v", err.Error())
				} else {
					fmt.Println("nPackage", uint16(jsonPackageId))
					nPackage = uint16(jsonPackageId)
				}
				jsonTax, err := info["platform_tax_percent"].(json.Number).Float64()
				if err != nil {
					log.Fatal("onPackageTax jsonTax:%v", err.Error())
				} else {
					fmt.Println("tax", jsonTax)
					nTax = jsonTax
				}

				SetPackageTaxM(nPackage, nTax)

				log.Debug("packageId:%v,tax:%v", nPackage, nTax)
			}
		}
	} else {
		log.Debug("onPackageTax error!!!")
	}
}

// 收到用户登录返回之后
func (c4c *Client4Center) onUserLogin(msg []byte) {
	loginResp := UserLoginResp{}
	err := json.Unmarshal(msg, &loginResp)
	if err != nil {
		log.Error("解析中心服返回数据出错")
	}

	userData := loginResp.Data

	code := userData.Code
	if code == constant.CRespStatusSuccess {
		log.Debug("onUserLogin SUCCESS :%v", loginResp)

		gameUser := userData.Msg.GameUser
		gameAccount := userData.Msg.GameAccount
		lockBalance := userData.Msg.GameAccount.LockBalance

		// 登入存在锁钱将解锁金额
		user, _ := Mgr.UserRecord.Load(gameUser.UserID)
		order := bson.NewObjectId().Hex()
		uid := util.UUID{}
		roundId := fmt.Sprintf("%+v-%+v", time.Now().Unix(), uid.GenUUID())
		if user != nil {
			u := user.(*User)
			rid := Mgr.UserRoom[u.UserID]
			v, _ := Mgr.RoomRecord.Load(rid)
			if v != nil {
				dl := v.(*Dealer)
				math := util.Math{}
				uBets, _ := math.SumSliceFloat64(dl.UserBets[u.UserID]).Float64() // 获取下注金额
				if u.IsAction == false || uBets == 0 {
					if lockBalance > 0 {
						c4c.UnlockSettlement(gameUser.UserID, lockBalance, order, roundId)
						log.Debug("玩家登入时锁资金:%v", lockBalance)
					}
				}
			}
		} else {
			if lockBalance > 0 {
				c4c.UnlockSettlement(gameUser.UserID, lockBalance, order, roundId)
				log.Debug("玩家登入时锁资金:%v", lockBalance)
			}
		}

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-login", gameUser.UserID)); ok {
			loginCallBack.(UserCallback)(&User{
				UserID:        gameUser.UserID,
				NickName:      gameUser.GameNick,
				Avatar:        gameUser.GameIMG,
				PackageId:     gameUser.PackageId,
				Balance:       gameAccount.Balance,
				BankerBalance: gameAccount.BankerBalance,
				Status:        gameAccount.Status,
			})

			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-login", gameUser.UserID))
		} else {
			log.Error("找不到用户回调")
		}

	} else {
		log.Error("中心服务器状态码", code)
	}
}

func (c4c *Client4Center) onUserLoginPac(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if !ok {
		log.Debug("onUserLoginPac Error")
		return
	}

	code, err := data["code"].(json.Number).Int64()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("data:%v,ok:%v", data, ok)

		userInfo, ok := data["msg"].(map[string]interface{})
		var userData *UserBack
		if ok {
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				userId := gameUser["id"]
				packageId := gameUser["package_id"]

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}

				pckId, err2 := packageId.(json.Number).Int64()
				if err2 != nil {
					log.Fatal(err2.Error())
				}

				log.Debug("packageID :%v", pckId)
				log.Debug("获取玩家的税率 :%v", packageTax[uint16(pckId)])
				//找到等待登录玩家
				userData, ok = c4c.waitUser[uint32(intID)]
				if ok {
					userData.Data.PackageId = uint16(pckId)
				}
			}
		}
	}
}

func (c4c *Client4Center) onUserLogout(msg []byte) {
	logoutResp := UserLogoutResp{}
	err := json.Unmarshal(msg, &logoutResp)
	if err != nil {
		log.Error("解析中心服返回数据出错")
	}

	userData := logoutResp.Data

	code := userData.Code
	if code == constant.CRespStatusSuccess {
		log.Debug("onUserLogout SUCCESS :%v", logoutResp)

		gameUser := userData.Msg.GameUser
		gameAccount := userData.Msg.GameAccount

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-logout", gameUser.UserID)); ok {
			loginCallBack.(UserCallback)(&User{
				UserID:        gameUser.UserID,
				NickName:      gameUser.GameNick,
				Avatar:        gameUser.GameIMG,
				Balance:       gameAccount.Balance,
				BankerBalance: gameAccount.BankerBalance,
				Status:        gameAccount.Status,
			})

			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-logout", gameUser.UserID))

		} else {
			log.Error("找不到用户回调")
		}
	} else {
		log.Error("中心服务器状态码", code)
	}
}

func (c4c *Client4Center) onUserWinScore(msg []byte) {
	winResp := SyncScoreResp{}
	err := json.Unmarshal(msg, &winResp)
	if err != nil {
		log.Error("解析加钱返回错误:%v", err)
	}

	syncData := winResp.Data
	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("onUserWinScore SUCCESS :%v", winResp)
		//winChan <- true

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-win-%+v", syncData.Msg.ID, syncData.Msg.Order)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, Balance: syncData.Msg.FinalBalance})
			// 回调成功之后要删除
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-win-%+v", syncData.Msg.ID, syncData.Msg.Order))
			// log.Debug("用户回调已删除: %+v, 回调队列 %+v", fmt.Sprintf("%+v-win-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}
	} else {
		log.Error("中心服务器状态码 %+v", syncData.Code)
	}
}

func (c4c *Client4Center) onUserLoseScore(msgData []byte) {
	loseResp := SyncScoreResp{}
	err := json.Unmarshal(msgData, &loseResp)
	if err != nil {
		log.Error("解析减钱返回错误:%v", err)
	}

	syncData := loseResp.Data
	order := syncData.Msg.Order
	if syncData.Code != constant.CRespStatusSuccess {
		id, _ := Mgr.OrderIDRecord.Load(order)
		v, ok := Mgr.UserRecord.Load(id)
		if ok {
			u := v.(*User)
			c4c.UserLogoutCenter(u.UserID, func(data *User) {
				Mgr.UserRecord.Delete(u.UserID)
				resp := &msg.LogoutR{}
				u.ConnAgent.WriteMsg(resp)
				u.ConnAgent.Close()
				Mgr.OrderIDRecord.Delete(order)
				id := strconv.FormatUint(uint64(u.UserID), 10)
				message := fmt.Sprintf("玩家" + id + "输钱失败并登出")
				SendTgMessage(message)
			})
		}
	}

	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("onUserLoseScore SUCCESS :%v", loseResp)
		//loseChan <- true
		Mgr.OrderIDRecord.Delete(order)
		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, Balance: syncData.Msg.FinalBalance})
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order))
			// log.Debug("用户回调已删除: %+v 回调队列 %+v", fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}

	} else {
		log.Error("中心服务器状态码 %+v %+v", syncData.Code, syncData.Msg)
	}
}

func (c4c *Client4Center) onChangeBankerStatus(msg []byte) {
	bankerResp := BankerResp{}
	err := json.Unmarshal(msg, &bankerResp)
	if err != nil {
		log.Error("解析庄家状态返回错误:%v", err)
	}

	syncData := bankerResp.Data
	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("onChangeBankerStatus SUCCESS :%v", bankerResp)

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-banker-status-%+v", syncData.Msg.ID, syncData.Msg.Status)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, BankerBalance: syncData.Msg.BankerBalance, Balance: syncData.Msg.Balance})
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-banker-status-%+v", syncData.Msg.ID, syncData.Msg.Status))
			// log.Debug("用户回调已删除: %+v 回调队列 %+v", fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}

	} else {
		log.Error("中心服务器状态码 %+v", syncData.Code)
	}
}

func (c4c *Client4Center) onBankerLoseScore(msgData []byte) {
	loseResp := SyncScoreResp{}
	err := json.Unmarshal(msgData, &loseResp)
	if err != nil {
		log.Error("解析减钱返回错误:%v", err)
	}

	syncData := loseResp.Data
	order := syncData.Msg.Order
	if syncData.Code != constant.CRespStatusSuccess {
		id, _ := Mgr.OrderIDRecord.Load(order)
		v, ok := Mgr.UserRecord.Load(id)
		if ok {
			u := v.(*User)
			c4c.UserLogoutCenter(u.UserID, func(data *User) {
				Mgr.UserRecord.Delete(u.UserID)
				resp := &msg.LogoutR{}
				u.ConnAgent.WriteMsg(resp)
				u.ConnAgent.Close()
				Mgr.OrderIDRecord.Delete(order)
				id := strconv.FormatUint(uint64(u.UserID), 10)
				message := fmt.Sprintf("庄家" + id + "输钱失败并登出")
				SendTgMessage(message)
			})
		}
	}

	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("onBankerLoseScore SUCCESS :%v", loseResp)

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-banker-lose-%+v", syncData.Msg.ID, syncData.Msg.Order)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, BankerBalance: syncData.Msg.FinalBankerBalance})
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-banker-lose-%+v", syncData.Msg.ID, syncData.Msg.Order))
			// log.Debug("用户回调已删除: %+v 回调队列 %+v", fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}

	} else {
		log.Error("中心服务器状态码 %+v", syncData.Code)
	}
}

func (c4c *Client4Center) onBankerWinScore(msg []byte) {
	winResp := SyncScoreResp{}
	err := json.Unmarshal(msg, &winResp)
	if err != nil {
		log.Error("解析加钱返回错误", err)
	}

	syncData := winResp.Data
	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("onBankerWinScore SUCCESS :%v", winResp)

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-banker-win-%+v", syncData.Msg.ID, syncData.Msg.Order)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, BankerBalance: syncData.Msg.FinalBankerBalance})
			// 回调成功之后要删除
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-banker-win-%+v", syncData.Msg.ID, syncData.Msg.Order))
			// log.Debug("用户回调已删除: %+v, 回调队列 %+v", fmt.Sprintf("%+v-win-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}
	} else {
		log.Error("中心服务器状态码 %+v", syncData.Code)
	}
}

func (c4c *Client4Center) onNotice(msg []byte) {
	log.Debug("<-------- onWinMoreThanNotice SUCCESS~!!! -------->")
}

func (c4c *Client4Center) onError(msg []byte) {
	centerErr := CenterErrorResp{}
	err := json.Unmarshal(msg, &centerErr)
	if err != nil {
		log.Error("中心服错误事件解析错误", err)
	}

	errData := centerErr.Data

	if errData.Code == constant.CRespTokenError {
		time.Sleep(30 * time.Second)
		//c4c.ReqToken()
		c4c.HeartBeatAndListen()
	}
}

//onWinMoreThanNotice 加锁金额
func (c4c *Client4Center) onLockSettlement(msgData []byte) {
	loseResp := LockSettleResp{}
	err := json.Unmarshal(msgData, &loseResp)
	if err != nil {
		log.Error("解析锁钱返回错误:%v", err)
	}

	syncData := loseResp.Data
	order := syncData.Msg.Order
	if syncData.Code != constant.CRespStatusSuccess {
		log.Debug("锁钱失败:%v", syncData)
		id, _ := Mgr.OrderIDRecord.Load(order)
		v, ok := Mgr.UserRecord.Load(id)
		if ok {
			u := v.(*User)
			c4c.UserLogoutCenter(u.UserID, func(data *User) {
				rid := Mgr.UserRoom[u.UserID]
				v, _ := Mgr.RoomRecord.Load(rid)
				if v != nil {
					dl := v.(*Dealer)
					u.winCount = 0
					u.betAmount = 0
					dl.UserIsDownBet[u.UserID] = false
					u.IsAction = false
					dl.UserBets[u.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
					dl.Users.Delete(u.UserID)
					delete(Mgr.UserRoom, u.UserID)
				}
				Mgr.UserRecord.Delete(u.UserID)
				resp := &msg.LogoutR{}
				u.ConnAgent.WriteMsg(resp)
				u.ConnAgent.Close()
				Mgr.OrderIDRecord.Delete(order)
				id := strconv.FormatUint(uint64(u.UserID), 10)
				message := fmt.Sprintf("玩家" + id + "锁钱失败")
				SendTgMessage(message)
			})
		}
		return
	}
	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("<-------- onLockSettlement SUCCESS~!!! -------->")
		id, _ := Mgr.OrderIDRecord.Load(order)
		v, ok := Mgr.UserRecord.Load(id)
		if ok {
			u := v.(*User)
			u.LockMoney += syncData.Msg.LockMoney
			Mgr.OrderIDRecord.Delete(order)
		}
		return
	}
}

//onWinMoreThanNotice 解锁金额
func (c4c *Client4Center) onUnlockSettlement(msgData []byte) {
	onUnlockResp := LockSettleResp{}
	err := json.Unmarshal(msgData, &onUnlockResp)
	if err != nil {
		log.Error("解析减钱返回错误:%v", err)
	}
	syncData := onUnlockResp.Data
	if syncData.Code == constant.CRespStatusSuccess {
		log.Debug("<-------- onUnlockSettlement SUCCESS~!!! -------->")
		return
	} else {
		log.Debug("解锁金额失败:%v", syncData)
	}
}

func (c4c *Client4Center) onDefault(msg []byte) {
	log.Error("中心服务器事件无法识别 %+v", string(msg))
}

/*****************************************************

	向中心服发送事件

******************************************************/

// 服务器登录中心服
func (c4c *Client4Center) ServerLoginCenter() {
	port, _ := strconv.Atoi(conf.Server.CenterServerPort)
	serverLoginMsg := ServerLoginReq{
		constant.CEventServerLogin,
		ServerLoginReqData{
			Host:    conf.Server.CenterServer,
			Port:    port,
			GameID:  conf.Server.GameID,
			DevName: conf.Server.DevName,
			DevKey:  conf.Server.DevKey,
		},
	}

	c4c.sendMsg2Center(serverLoginMsg)
}

func (c4c *Client4Center) heartBeat() {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	err := c4c.conn.WriteMessage(websocket.PingMessage, nil)

	if err != nil {
		log.Error(err.Error())
	}
}

// 操作用户数据一定要等中心服确认消息返回之后再进行展示或其他操作

// UserLoginCenter 用户登录
func (c4c *Client4Center) UserLoginCenter(userID uint32, pass, token string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserLoginCenter c4c.Token- %+v", c4c.token)

	userLoginMsg := UserLoginReq{
		Event: constant.CEventUserLogin,
		Data: UserLoginReqData{
			UserID:   userID,
			Password: pass,
			Token:    token,
			DevName:  conf.Server.DevName,
			GameID:   conf.Server.GameID,
			DevKey:   conf.Server.DevKey,
		},
	}

	c4c.sendMsg2Center(userLoginMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-login", userID), callback)

	//加入待处理map，等待处理
	c4c.waitUser[userID] = &UserBack{}
	c4c.waitUser[userID].Data.UserID = userID
	c4c.waitUser[userID].Callback = callback
}

// UserLogoutCenter 用户登出
func (c4c *Client4Center) UserLogoutCenter(userID uint32, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserLogoutCenter c4c.Token- %+v", c4c.token)

	logoutMsg := UserLogoutReq{
		Event: constant.CEventUserLogout,
		Data: UserLogoutReqData{
			UserID: userID,
			//Token:  c4c.token,
			DevName: conf.Server.DevName,
			GameID:  conf.Server.GameID,
			DevKey:  conf.Server.DevKey,
		},
	}

	c4c.sendMsg2Center(logoutMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-logout", userID), callback)
}

func (c4c *Client4Center) UserWinScore(timeNow, userID uint32, money float64, DownBetTotal float64, order, roundID string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserWinScore c4c.Token- %+v", c4c.token)

	winSettleMsg := SyncScoreReq{
		Event: constant.CEventUserWinScore,
		Data: SyncScoreReqData{
			Auth: ServerAuth{
				//Token:  c4c.token,
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     userID,
				CreateTime: timeNow,
				PayReason:  "玩家赢钱",
				Money:      money,
				BetMoney:   DownBetTotal,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}

	c4c.sendMsg2Center(winSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-win-%+v", userID, order), callback)
}

func (c4c *Client4Center) UserLoseScore(timeNow, userID uint32, money float64, DownBetTotal float64, order, roundID string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserLoseScore c4c.Token- %+v", c4c.token)

	loseSettleMsg := SyncScoreReq{
		Event: constant.CEventUserLoseScore,
		Data: SyncScoreReqData{
			Auth: ServerAuth{
				//Token:  c4c.token,
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     userID,
				CreateTime: timeNow,
				PayReason:  "玩家输钱",
				Money:      money,
				BetMoney:   DownBetTotal,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}
	Mgr.OrderIDRecord.Store(order, userID)
	c4c.sendMsg2Center(loseSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-lose-%+v", userID, order), callback)
}

func (c4c *Client4Center) ChangeBankerStatus(userID uint32, status int, money float64, order, round string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	bankerMsg := BankerReq{
		Event: constant.CEventChangeBankerStatus,
		Data: BankerReqData{
			Auth: ServerAuth{
				//Token:  c4c.token,
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: BankerReqDataInfo{
				UserID:     userID,
				Status:     status,
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "玩家上下庄",
				Order:      order,
				RoundID:    round,
				Money:      money,
				GameID:     conf.Server.GameID,
			},
		},
	}

	c4c.sendMsg2Center(bankerMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-banker-status-%+v", userID, status), callback)
}

func (c4c *Client4Center) BankerWinScore(userID uint32, money float64, order, roundID string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserWinScore c4c.Token- %+v", c4c.token)

	winSettleMsg := SyncScoreReq{
		Event: constant.CEventBankerWinScore,
		Data: SyncScoreReqData{
			Auth: ServerAuth{
				//Token:  c4c.token,
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     userID,
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "庄家赢钱",
				Money:      money,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}

	c4c.sendMsg2Center(winSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-banker-win-%+v", userID, order), callback)
}

func (c4c *Client4Center) BankerLoseScore(userID uint32, money float64, order, roundID string, callback UserCallback) {
	if !c4c.isServerLogin {
		log.Debug("Game Server NOT Ready! Need login to Center Server!")
		return
	}

	//log.Debug("UserLoseScore c4c.Token- %+v", c4c.token)

	loseSettleMsg := SyncScoreReq{
		Event: constant.CEventBankerLoseScore,
		Data: SyncScoreReqData{
			Auth: ServerAuth{
				//Token:  c4c.token,
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     userID,
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "庄家输钱",
				Money:      money,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}
	Mgr.OrderIDRecord.Store(order, userID)
	c4c.sendMsg2Center(loseSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-banker-lose-%+v", userID, order), callback)
}

//锁钱
func (c4c *Client4Center) LockSettlement(au *User, lockAccount float64, order, roundID string) {
	lockSettle := LockSettle{
		Event: constant.MsgLockSettlement,
		Data: LockChangeSettle{
			Auth: ServerAuth{
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     au.UserID,
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "LockMoney",
				LockMoney:  lockAccount,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}
	c4c.sendMsg2Center(lockSettle)
	Mgr.OrderIDRecord.Store(order, au.UserID)
}

//解锁
func (c4c *Client4Center) UnlockSettlement(UserId uint32, LockMoney float64, order, roundID string) {
	unLockSettle := UnLockSettle{
		Event: constant.MsgUnlockSettlement,
		Data: LockChangeSettle{
			Auth: ServerAuth{
				DevName: conf.Server.DevName,
				DevKey:  conf.Server.DevKey,
			},

			Info: SyncScoreReqDataInfo{
				UserID:     UserId,
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "UnLockMoney",
				LockMoney:  LockMoney,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}
	c4c.sendMsg2Center(unLockSettle)
}

// 向中心服发送消息的基础函数
func (c4c *Client4Center) sendMsg2Center(data interface{}) {
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("解析失败", err)
	}
	log.Debug("Msg to center %v", string(bs))

	writeMutex.Lock()
	defer writeMutex.Unlock()
	err = c4c.conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		log.Fatal("发送数据失败", err)
	}
}

func (c4c *Client4Center) NoticeWinMoreThan(playerId uint32, playerName string, winGold float64) {
	log.Debug("<-------- NoticeWinMoreThan  -------->")
	msg := fmt.Sprintf("<size=20><color=yellow>恭喜!</color><color=orange>%v</color><color=yellow>在</color></><color=orange><size=25>奔驰宝马</color></><color=yellow><size=20>中一把赢了</color></><color=yellow><size=30>%.2f</color></><color=yellow><size=25>金币！</color></>", playerName, winGold)

	base := &NoticeReq{}
	base.Event = constant.CEventNotice
	base.Data = NoticeReqData{
		DevName: conf.Server.DevName,
		DevKey:  conf.Server.DevKey,
		ID:      playerId,
		GameId:  conf.Server.GameID,
		Type:    2000,
		Message: msg,
		Topic:   "系统提示",
	}
	c4c.sendMsg2Center(base)
}

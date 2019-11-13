package internal

import (
	"encoding/json"
	"fmt"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/constant"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"proj_bcbm/src/server/log"
)

type UserCallback func(data *User)

type Client4Center struct {
	//token         string
	//tokenLock     sync.RWMutex
	conn          *websocket.Conn
	isServerLogin bool
	userWaitEvent sync.Map
}

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
	}
}

/***********************************************************

	请求服务器token并连接中心服

************************************************************/

//// 从中心服请求token
//func (c4c *Client4Center) ReqToken() {
//	req, err := http.NewRequest("GET", conf.Server.TokenServer, nil)
//	if err != nil {
//		log.Fatal("生成请求失败")
//		panic(err)
//	}
//	params := req.URL.Query()
//	params.Add("dev_key", conf.Server.DevKey)
//	params.Add("dev_name", conf.Server.DevName)
//	req.URL.RawQuery = params.Encode()
//
//	log.Debug("请求Token %+v", req.URL.String())
//
//	client := http.Client{}
//	resp, err := client.Do(req)
//	if err != nil || resp.StatusCode != 200 {
//		log.Debug("请求中心服token失败 %+v", err)
//	}
//
//	bs, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		log.Fatal("响应体读取失败", err)
//	}
//
//	// log.Debug(string(bs))
//	tokenResp := TokenResp{}
//
//	err = json.Unmarshal(bs, &tokenResp)
//
//	if err != nil {
//		log.Fatal("Token响应解析失败", err)
//	}
//
//	if tokenResp.StatusCode != 200 {
//		log.Fatal("Token响应码不是200", tokenResp.StatusCode)
//	}
//
//	c4c.tokenLock.Lock()
//	c4c.token = tokenResp.TokenMsg.Token
//	c4c.tokenLock.Unlock()
//
//	log.Debug("Token更新完成 %+v", c4c.token)
//}
//
//func (c4c *Client4Center) CronUpdateToken() {
//	// ticker := time.NewTicker(time.Second * 7200)
//	ticker := time.NewTicker(time.Second * 600)
//	go func() {
//		for {
//			<-ticker.C
//			c4c.ReqToken()
//		}
//	}()
//}

/*****************************************

	监听中心服返回数据并处理

******************************************/

func (c4c *Client4Center) HeartBeatAndListen() {
	ticker := time.NewTicker(time.Second * 3)
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
				log.Debug("Msg from center %v", string(message))

				var msg Server2CenterMsg
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Error("Json Unmarshal error", err.Error())
				}
				switch msg.Event {
				case constant.CEventServerLogin:
					c4c.onServerLogin(message)
				case constant.CEventUserLogin:
					c4c.onUserLogin(message)
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
		log.Error("解析服务器登录返回数据错误", err)
	}

	data := sLogin.Data
	status := data.Status
	// code := data.Code
	taxPercent := data.Msg.PlatformTaxPercent

	c4c.isServerLogin = true
	log.Debug("服务器登陆 %+v 税率 %%%+v ...", status, taxPercent)
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
		gameUser := userData.Msg.GameUser
		gameAccount := userData.Msg.GameAccount

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-login", gameUser.UserID)); ok {
			loginCallBack.(UserCallback)(&User{
				UserID:        gameUser.UserID,
				NickName:      gameUser.GameNick,
				Avatar:        gameUser.GameIMG,
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

func (c4c *Client4Center) onUserLogout(msg []byte) {
	logoutResp := UserLogoutResp{}
	err := json.Unmarshal(msg, &logoutResp)
	if err != nil {
		log.Error("解析中心服返回数据出错")
	}

	userData := logoutResp.Data

	code := userData.Code
	if code == constant.CRespStatusSuccess {
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
		log.Error("解析加钱返回错误", err)
	}

	syncData := winResp.Data
	if syncData.Code == constant.CRespStatusSuccess {

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

func (c4c *Client4Center) onUserLoseScore(msg []byte) {
	loseResp := SyncScoreResp{}
	err := json.Unmarshal(msg, &loseResp)
	if err != nil {
		log.Error("解析减钱返回错误", err)
	}

	syncData := loseResp.Data
	if syncData.Code == constant.CRespStatusSuccess {

		if loginCallBack, ok := c4c.userWaitEvent.Load(fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order)); ok {
			loginCallBack.(UserCallback)(&User{UserID: syncData.Msg.ID, Balance: syncData.Msg.FinalBalance})
			c4c.userWaitEvent.Delete(fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order))
			// log.Debug("用户回调已删除: %+v 回调队列 %+v", fmt.Sprintf("%+v-lose-%+v", syncData.Msg.ID, syncData.Msg.Order), c4c.userWaitEvent)
		} else {
			log.Error("找不到用户回调")
		}

	} else {
		log.Error("中心服务器状态码 %+v", syncData.Code)
	}
}

func (c4c *Client4Center) onChangeBankerStatus(msg []byte) {
	bankerResp := BankerResp{}
	err := json.Unmarshal(msg, &bankerResp)
	if err != nil {
		log.Error("解析庄家状态返回错误", err)
	}

	syncData := bankerResp.Data
	if syncData.Code == constant.CRespStatusSuccess {
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

func (c4c *Client4Center) onBankerLoseScore(msg []byte) {
	loseResp := SyncScoreResp{}
	err := json.Unmarshal(msg, &loseResp)
	if err != nil {
		log.Error("解析减钱返回错误", err)
	}

	syncData := loseResp.Data
	if syncData.Code == constant.CRespStatusSuccess {

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

// todo
func (c4c *Client4Center) onNotice(msg []byte) {

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

func (c4c *Client4Center) onDefault(msg []byte) {
	log.Error("中心服务器事件无法识别 %+v", string(msg))
}

/*****************************************************

	向中心服发送事件

******************************************************/

// 服务器登录中心服
func (c4c *Client4Center) ServerLoginCenter() {
	serverLoginMsg := ServerLoginReq{
		constant.CEventServerLogin,
		ServerLoginReqData{
			Host:    conf.Server.CenterServer,
			Port:    conf.Server.CenterServerPort,
			GameID:  conf.Server.GameID,
			DevName: conf.Server.DevName,
			DevKey:  conf.Server.DevKey,
		},
	}

	c4c.sendMsg2Center(serverLoginMsg)
}

func (c4c *Client4Center) heartBeat() {
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

func (c4c *Client4Center) UserWinScore(userID uint32, money float64, order, roundID string, callback UserCallback) {
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
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "玩家赢钱",
				Money:      money,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}

	c4c.sendMsg2Center(winSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-win-%+v", userID, order), callback)
}

func (c4c *Client4Center) UserLoseScore(userID uint32, money float64, order, roundID string, callback UserCallback) {
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
				CreateTime: uint32(time.Now().Unix()),
				PayReason:  "玩家输钱",
				Money:      money,
				Order:      order,
				GameID:     conf.Server.GameID,
				RoundID:    roundID,
			},
		},
	}

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
				PayReason:  "玩家上下庄", // todo 上下庄分开？
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

	c4c.sendMsg2Center(loseSettleMsg)
	c4c.userWaitEvent.Store(fmt.Sprintf("%+v-banker-lose-%+v", userID, order), callback)
}

// todo
func (c4c *Client4Center) PromoNotice() {

}

// 向中心服发送消息的基础函数
func (c4c *Client4Center) sendMsg2Center(data interface{}) {
	// todo 写入数据库
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("解析失败", err)
	}
	log.Debug("Msg to center %v", string(bs))

	err = c4c.conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		log.Fatal("发送数据失败", err)
	}
}

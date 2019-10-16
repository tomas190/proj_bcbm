package internal

// 游戏服务器<--->中心服 消息基本格式
type Server2CenterMsg struct {
	Event string `json:"event"`
	// Data  interface{} `json:"data"`
}

/****************************************

	ServerLogin 服务器登录

 ****************************************/

type ServerLoginReq struct {
	Event string             `json:"event"`
	Data  ServerLoginReqData `json:"data"`
}

type ServerLoginReqData struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	GameID  string `json:"game_id"`
	DevName string `json:"dev_name"`
	DevKey  string `json:"dev_key"`
}

type ServerLoginResp struct {
	Event string              `json:"event"`
	Data  ServerLoginRespData `json:"data"`
}

type ServerLoginRespData struct {
	Status string                 `json:"status"`
	Code   int                    `json:"code"`
	Msg    ServerLoginRespDataMsg `json:"msg"`
}

type ServerLoginRespDataMsg struct {
	PlatformTaxPercent int `json:"platform_tax_percent"`
}

/****************************************

	UserLogin 用户登录

 ****************************************/

type UserLoginReq struct {
	Event string           `json:"event"`
	Data  UserLoginReqData `json:"data"`
}

type UserLoginReqData struct {
	UserID   uint32 `json:"id"`
	Password string `json:"password"`
	Token    string `json:"token"`
	DevName  string `json:"dev_name"`
	GameID   string `json:"game_id"`
	DevKey   string `json:"dev_key"`
}

type UserLoginResp struct {
	Event string            `json:"event"`
	Data  UserLoginRespData `json:"data"`
}

type UserLoginRespData struct {
	Code   int        `json:"code"`
	Status string     `json:"status"`
	Msg    UserLogMsg `json:"msg"`
}

type UserLogMsg struct {
	GameUser    GameUser    `json:"game_user"`
	GameAccount GameAccount `json:"game_account"`
}

type GameUser struct {
	UserID       uint32 `json:"id"`
	UUID         string `json:"uuid"`
	GameNick     string `json:"game_nick"`
	GameIMG      string `json:"game_img"`
	GameUserType uint32 `json:"game_user_type"`
}

type GameAccount struct {
	Balance       float64 `json:"balance"`
	BankerBalance float64 `json:"banker_balance"`
	Status        int     `json:"status"`
	GameName      string  `json:"game_name"`
}

/****************************************

	UserLogout 用户登出

 ****************************************/

type UserLogoutReq struct {
	Event string            `json:"event"`
	Data  UserLogoutReqData `json:"data"`
}

type UserLogoutReqData struct {
	UserID  uint32 `json:"id"`
	DevName string `json:"dev_name"`
	GameID  string `json:"game_id"`
	DevKey  string `json:"dev_key"`
}

type UserLogoutResp struct {
	Event string             `json:"event"`
	Data  UserLogoutRespData `json:"data"`
}

type UserLogoutRespData struct {
	Code   int        `json:"code"`
	Status string     `json:"status"`
	Msg    UserLogMsg `json:"msg"`
}

/********************************************************

	错误

********************************************************/

type CenterErrorResp struct {
	Event string              `json:"event"`
	Data  CenterErrorRespData `json:"data"`
}

type CenterErrorRespData struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"error"`
}

/********************************************************

	用户加钱减钱-中心服API垃圾！！！

********************************************************/
type SyncScoreReq struct {
	Event string           `json:"event"`
	Data  SyncScoreReqData `json:"data"`
}

type SyncScoreReqData struct {
	Auth ServerAuth           `json:"auth"`
	Info SyncScoreReqDataInfo `json:"info"`
}

// 服务器验证信息
type ServerAuth struct {
	DevName string `json:"dev_name"`
	DevKey  string `json:"dev_key"`
}

// 请求信息
type SyncScoreReqDataInfo struct {
	UserID     uint32  `json:"id"`
	CreateTime uint32  `json:"create_time"`
	PayReason  string  `json:"pay_reason"`
	Money      float64 `json:"money"`
	LockMoney  float64 `json:"lock_money"`
	PreMoney   float64 `json:"pre_money"`
	Order      string  `json:"order"`
	GameID     string  `json:"game_id"`
	RoundID    string  `json:"round_id"`
}

type SyncScoreResp struct {
	Event string            `json:"event"`
	Data  SyncScoreRespData `json:"data"`
}

type SyncScoreRespData struct {
	Code   int          `json:"code"`
	Status string       `json:"status"`
	Msg    SyncScoreMsg `json:"msg"`
}

type SyncScoreMsg struct {
	ID                 uint32  `json:"id"`
	Balance            float64 `json:"balance"`
	FinalBalance       float64 `json:"final_balance"`
	BankerBalance      float64 `json:"banker_balance"`
	FinalBankerBalance float64 `json:"final_banker_balance"`
	Income             float64 `json:"income"`
	Order              string  `json:"order"`
	RoundID            string  `json:"round_id"`
}

/***************************************************

	庄家上庄下庄

****************************************************/

type BankerReq struct {
	Event string        `json:"event"`
	Data  BankerReqData `json:"data"`
}

type BankerReqData struct {
	Auth ServerAuth        `json:"auth"`
	Info BankerReqDataInfo `json:"info"`
}

// 请求信息
type BankerReqDataInfo struct {
	UserID     uint32  `json:"id"`
	Status     int     `json:"status"`
	CreateTime uint32  `json:"create_time"`
	PayReason  string  `json:"pay_reason"`
	Money      float64 `json:"money"`
	LockMoney  float64 `json:"lock_money"`
	PreMoney   float64 `json:"pre_money"`
	Order      string  `json:"order"`
	GameID     string  `json:"game_id"`
	RoundID    string  `json:"round_id"`
}

type BankerResp struct {
	Event string         `json:"event"`
	Data  BankerRespData `json:"data"`
}

type BankerRespData struct {
	Code   int       `json:"code"`
	Status string    `json:"status"`
	Msg    BankerMsg `json:"msg"`
}

type BankerMsg struct {
	ID            uint32  `json:"id"`
	BankerBalance float64 `json:"banker_balance"`
	Balance       float64 `json:"balance"`
	Status        int     `json:"status"`
	Reason        string  `json:"reason"`
}

/***************************************************

	请求Token

****************************************************/

//type TokenResp struct {
//	StatusCode int      `json:"code"`
//	TokenMsg   tokenMsg `json:"msg"`
//}
//
//type tokenMsg struct {
//	Token string `json:"token"`
//}

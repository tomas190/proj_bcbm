package constant

const (
	CEventServerLogin        string = "/GameServer/Login/login"
	CEventUserLogin          string = "/GameServer/GameUser/login"
	CEventUserLogout         string = "/GameServer/GameUser/loginout"
	CEventUserWinScore       string = "/GameServer/GameUser/winSettlement"
	CEventUserLoseScore      string = "/GameServer/GameUser/loseSettlement"
	CEventChangeBankerStatus string = "/GameServer/GameUser/changeAccountBankerAndStatus"
	CEventBankerWinScore     string = "/GameServer/GameUser/bankerWinSettlement"
	CEventBankerLoseScore    string = "/GameServer/GameUser/bankerLoseSettlement"
	CEventNotice             string = "/GameServer/Notice/notice"
	MsgLockSettlement        string = "/GameServer/GameUser/lockSettlement"   //锁钱
	MsgUnlockSettlement      string = "/GameServer/GameUser/unlockSettlement" //解锁

	CEventError string = "error"

	CRespStatusSuccess int = 200
	CRespTokenError    int = 501
)

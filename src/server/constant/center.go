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

	CEventError string = "error"

	CRespStatusSuccess int = 200
	CRespTokenError    int = 501
)

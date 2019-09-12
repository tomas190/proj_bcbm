package constant

const (
	CEventServerLogin   string = "/GameServer/Login/login"
	CEventUserLogin     string = "/GameServer/GameUser/login"
	CEventUserLogout    string = "/GameServer/GameUser/loginout"
	CEventUserWinScore  string = "/GameServer/GameUser/winSettlement"
	CEventUserLoseScore string = "/GameServer/GameUser/loseSettlement"
	CEventError         string = "error"

	CRespStatusSuccess int = 200
	CRespTokenError    int = 501
)

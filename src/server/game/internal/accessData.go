package internal

import (
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/msg"
	"strconv"
	"time"
)

type GameDataReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	RoundId   string `form:"round_id" json:"round_id"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Skip      int    `form:"skip" json:"skip"`
	Limit     int    `form:"limit" json:"limit"`
}

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GameData struct {
	Time       int64       `json:"time"`
	TimeFmt    string      `json:"time_fmt"`
	StartTime  int64       `json:"start_time"`
	EndTime    int64       `json:"end_time"`
	PlayerId   string      `json:"player_id"`
	RoundId    string      `json:"round_id"`
	RoomId     uint32      `json:"room_id"`
	TaxRate    float64     `json:"tax_rate"`
	Card       interface{} `json:"card"`       // 开牌信息
	BetInfo    interface{} `json:"bet_info"`   // 玩家下注信息
	Settlement interface{} `json:"settlement"` // 结算信息 输赢结果
}

type pageData struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

const (
	SuccCode = 0
	ErrCode  = -1
)

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	http.HandleFunc("/api/getGameData", getAccessData)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq

	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.RoundId = r.FormValue("round_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	skip := r.FormValue("skip")
	limit := r.FormValue("limit")

	selector := bson.M{}

	if req.Id != "" {
		selector["id"] = req.Id
	}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoundId != "" {
		selector["round_id"] = req.RoundId
	}

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime != 0 && eTime == 0 {
		selector["start_time"] = bson.M{"$gte": sTime}
	}

	if eTime != 0 && sTime == 0 {
		selector["end_time"] = bson.M{"$lte": eTime}
	}

	skips, _ := strconv.Atoi(skip)
	if skips != 0 {
		selector["skip"] = skips
	}

	limits, _ := strconv.Atoi(limit)
	if limits != 0 {
		selector["limit"] = limits
	}

	recodes, count, err := db.GetDownRecodeList(skips, limits, selector, "down_bet_time")
	if err != nil {
		return
	}

	var gameData []GameData
	for i := 0; i < len(recodes); i++ {
		var gd GameData
		pr := recodes[i]
		gd.Time = pr.DownBetTime
		gd.TimeFmt = FormatTime(pr.DownBetTime, "2006-01-02 15:04:05")
		gd.StartTime = pr.StartTime
		gd.EndTime = pr.EndTime
		gd.PlayerId = pr.Id
		gd.RoomId = pr.RoomId
		gd.RoundId = pr.RoundId
		gd.BetInfo = pr.DownBetInfo
		gd.Card = pr.CardResult
		gd.Settlement = pr.ResultMoney
		gd.TaxRate = pr.TaxRate
		gameData = append(gameData, gd)
	}

	var result pageData
	result.Total = count
	result.List = gameData

	//fmt.Fprintf(w, "%+v", ApiResp{Code: SuccCode, Msg: "", Data: result})
	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func FormatTime(timeUnix int64, layout string) string {
	if timeUnix == 0 {
		return ""
	}
	format := time.Unix(timeUnix, 0).Format(layout)
	return format
}

func NewResp(code int, msg string, data interface{}) ApiResp {
	return ApiResp{Code: code, Msg: msg, Data: data}
}

func reqPlayerLeave(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("id")
	userId, _ := strconv.Atoi(Id)
	log.Debug("玩家id为:%v,%v", Id, uint32(userId))
	u, _ := Mgr.UserRecord.Load(uint32(userId))
	if u != nil {
		au := u.(*User)
		log.Debug("玩家信息:%v", au)
		rid := Mgr.UserRoom[au.UserID]
		v, _ := Mgr.RoomRecord.Load(rid)
		if v != nil {
			dl := v.(*Dealer)
			if au.IsAction == false {
				log.Debug("进来了111")
				dl.Users.Delete(au.UserID)
				c4c.UserLogoutCenter(au.UserID, func(data *User) {
					dl.AutoBetRecord[au.UserID] = nil
					Mgr.UserRecord.Delete(au.UserID)
					resp := &msg.LogoutR{}
					au.ConnAgent.WriteMsg(resp)
					au.ConnAgent.Close()
				})
			} else {
				log.Debug("进来了222")
				var exist bool
				for _, v := range dl.UserLeave {
					if v == au.UserID {
						exist = true
						log.Debug("rpcCloseAgent 玩家已存在UserLeave:%v", au.UserID)
					}
				}
				if exist == false {
					log.Debug("rpcCloseAgent 添加离线UserLeave:%v", au.UserID)
					dl.UserLeave = append(dl.UserLeave, au.UserID)
				}
			}
		} else {
			log.Debug("进来了333")
			c4c.UserLogoutCenter(au.UserID, func(data *User) {
				resp := &msg.LogoutR{}
				au.ConnAgent.WriteMsg(resp)
				au.ConnAgent.Close()
			})
		}
	}
}

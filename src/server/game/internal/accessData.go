package internal

import (
	//"fmt"
	//"strconv"
	//"gopkg.in/mgo.v2/bson"
	"proj_bcbm/src/server/conf"
	"github.com/name5566/leaf/log"
	"net/http"
	"time"
)

type GameDataReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	RoundId   string `form:"round_id" json:"round_id"`
	StartTime int64  `form:"start_time" json:"start_time"`
	EndTime   int64  `form:"end_time" json:"end_time"`
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
	PlayerId   string      `json:"player_id"`
	RoundId    string      `json:"round_id"`
	RoomId     string      `json:"room_id"`
	TaxRate    float64     `json:"tax_rate"`
	Card       interface{} `json:"card"`       // 开牌信息
	BetInfo    interface{} `json:"bet_info"`   // 玩家下注信息  //todo  betinfo
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

func StartHttpServer() {
	http.HandleFunc("/api/accessData", getAccessData)

	err := http.ListenAndServe(":"+ conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	//var req GameDataReq
	//
	//req.Id = r.FormValue("id")
	//req.GameId = r.FormValue("game_id")
	//req.RoundId = r.FormValue("round_id")
	//startTime := r.FormValue("start_time")
	//endTime := r.FormValue("end_time")
	//skip := r.FormValue("skip")
	//limit := r.FormValue("limit")
	//
	////if req.Id == "" || req.GameId == "" || req.RoundId == "" || startTime == "" ||
	////	endTime == "" || skip == "" || limit == "" {
	////	fmt.Fprint(w, NewResp(ErrCode, "", nil))
	////	return
	////}
	//
	//selector := bson.M{}
	//
	//if req.Id != "" {
	//	selector["id"] = req.Id
	//}
	//
	//if req.GameId != "" {
	//	selector["game_id"] = req.GameId
	//}
	//
	//if req.RoundId != "" {
	//	selector["rand_id"] = req.RoundId
	//}
	//
	//sTime, _ := strconv.Atoi(startTime)
	//
	//eTime, _ := strconv.Atoi(endTime)
	//
	//if sTime != 0 && eTime != 0 {
	//	selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	//}
	//
	//if sTime != 0 && eTime == 0 {
	//	selector["down_bet_time"] = bson.M{"$gt": sTime}
	//}
	//
	//if eTime != 0 && sTime == 0 {
	//	selector["down_bet_time"] = bson.M{"$lt": eTime}
	//}
	//
	//skips, _ := strconv.Atoi(skip)
	//if skips != 0 {
	//	selector["skip"] = skips
	//}
	//
	//limits, _ := strconv.Atoi(limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}
	//
	//recodes, count, err := db.GetDownRecodeList(skips, limits, selector, "down_bet_time")
	//if err != nil {
	//	return
	//}
	//
	//var gameData []GameData
	//for i := 0; i < len(recodes); i++ {
	//	var gd GameData
	//	pr := recodes[i]
	//	gd.Time = pr.DownBetTime * 1000
	//	gd.TimeFmt = FormatTime(pr.DownBetTime, "2006-01-02 15:04:05")
	//	gd.PlayerId = pr.Id
	//	gd.RoomId = pr.RoomId
	//	gd.RoundId = pr.RandId
	//	gd.BetInfo = pr.DownBetInfo
	//	gd.Card = pr.CardResult
	//	gd.Settlement = pr.ResultMoney
	//	gd.TaxRate = pr.TaxRate
	//	gameData = append(gameData, gd)
	//}
	//
	//var result pageData
	//result.Total = count
	//result.List = gameData
	//
	//fmt.Fprint(w, NewResp(SuccCode, "Success", result))
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
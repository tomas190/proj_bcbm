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
	Skip      string `form:"skip" json:"skip"`
	Limit     string `form:"limit" json:"limit"`
}

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GameData struct {
	Time            int64       `json:"time"`
	TimeFmt         string      `json:"time_fmt"`
	StartTime       int64       `json:"start_time"`
	EndTime         int64       `json:"end_time"`
	Id              string      `json:"id"`
	RoundId         string      `json:"round_id"`
	RoomId          uint32      `json:"room_id"`
	TaxRate         float64     `json:"tax_rate"`
	Card            interface{} `json:"card"`             // 开牌信息
	BetInfo         interface{} `json:"bet_info"`         // 玩家下注信息
	SettlementFunds interface{} `json:"settlement_funds"` // 结算信息 输赢结果
	SpareCash       interface{} `json:"spare_cash"`       // 剩余金额

}

type pageData struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

const (
	SuccCode = 0
	ErrCode  = -1
)

type GetSurPool struct {
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int64   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int64   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
}

type UpSurPool struct {
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	CoefficientToTotalPlayer       int64   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
}

type GRobotData struct {
	RoomId   uint32       `json:"room_id" bson:"room_id"`
	RoomTime int64        `json:"room_time" bson:"room_time"`
	RobotNum int          `json:"robot_num" bson:"robot_num"`
	AreaX1   *ChipDownBet `json:"area_x_1" bson:"area_x_1"`
	AreaX2   *ChipDownBet `json:"area_x_2" bson:"area_x_2"`
	AreaX3   *ChipDownBet `json:"area_x_3" bson:"area_x_3"`
	AreaX4   *ChipDownBet `json:"area_x_4" bson:"area_x_4"`
	AreaX5   *ChipDownBet `json:"area_x_5" bson:"area_x_5"`
	AreaX6   *ChipDownBet `json:"area_x_6" bson:"area_x_6"`
	AreaX7   *ChipDownBet `json:"area_x_7" bson:"area_x_7"`
	AreaX8   *ChipDownBet `json:"area_x_8" bson:"area_x_8"`
}

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	http.HandleFunc("/api/getGameData", getAccessData)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)
	// 查询子游戏盈余池数据
	http.HandleFunc("/api/getSurplusOne", getSurplusOne)
	// 修改盈余池数据
	http.HandleFunc("/api/uptSurplusConf", uptSurplusOne)
	// 获取机器人数据
	http.HandleFunc("/api/getRobotData", getRobotData)

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
	req.Skip = r.FormValue("skip")
	req.Limit = r.FormValue("limit")

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

	skips, _ := strconv.Atoi(req.Skip)
	if skips != 0 {
		selector["skip"] = skips
	}

	limits, _ := strconv.Atoi(req.Limit)
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
		gd.Id = pr.Id
		gd.RoomId = pr.RoomId
		gd.RoundId = pr.RoundId
		gd.BetInfo = pr.DownBetInfo
		gd.Card = pr.CardResult
		gd.SettlementFunds = pr.SettlementFunds
		gd.SpareCash = pr.SpareCash
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
	Mgr.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*User)
		if u.UserID == uint32(userId) {
			rid := Mgr.UserRoom[u.UserID]
			v, _ := Mgr.RoomRecord.Load(rid)
			if v != nil {
				dl := v.(*Dealer)
				u.winCount = 0
				u.betAmount = 0
				u.DownBetTotal = 0
				u.IsAction = false
				dl.UserIsDownBet[u.UserID] = false
				dl.UserBets[u.UserID] = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
				dl.Users.Delete(u.UserID)
				resp := &msg.LeaveRoomR{
					User: &msg.UserInfo{
						UserID:   u.UserID,
						Avatar:   u.Avatar,
						NickName: u.NickName,
						Money:    u.Balance,
					},
					Rooms:      Mgr.GetRoomsInfoResp(),
					ServerTime: uint32(time.Now().Unix()),
				}
				if u.ConnAgent != nil {
					log.Debug("玩家退出房间信息:%v", u)
				}
				u.ConnAgent.WriteMsg(resp)

				js, err := json.Marshal(NewResp(SuccCode, "", "玩家退出房间成功"))
				if err != nil {
					fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
					//return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			}
			//else {
			//	js, err := json.Marshal(NewResp(ErrCode, "", "玩家退出房间失败"))
			//	if err != nil {
			//		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			//		//return
			//	}
			//	w.Header().Set("Content-Type", "application/json")
			//	w.Write(js)
			//}
		}
		return true
	})
}

// 查询子游戏盈余池数据
func getSurplusOne(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq
	req.GameId = r.FormValue("game_id")
	log.Debug("game_id :%v", req.GameId)

	selector := bson.M{}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	result, err := db.GetSurPoolData(selector)
	if err != nil {
		return
	}

	var getSur GetSurPool
	getSur.PlayerTotalLose = result.PlayerTotalLose
	getSur.PlayerTotalWin = result.PlayerTotalWin
	getSur.PercentageToTotalWin = result.PercentageToTotalWin
	getSur.TotalPlayer = result.TotalPlayer
	getSur.CoefficientToTotalPlayer = result.CoefficientToTotalPlayer
	getSur.FinalPercentage = result.FinalPercentage
	getSur.PlayerTotalLoseWin = result.PlayerTotalLoseWin
	getSur.SurplusPool = result.SurplusPool
	getSur.PlayerLoseRateAfterSurplusPool = result.PlayerLoseRateAfterSurplusPool
	getSur.DataCorrection = result.DataCorrection

	js, err := json.Marshal(NewResp(SuccCode, "", getSur))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func uptSurplusOne(w http.ResponseWriter, r *http.Request) {

	rateSur := r.PostFormValue("player_lose_rate_after_surplus_pool")
	percentage := r.PostFormValue("percentage_to_total_win")
	coefficient := r.PostFormValue("coefficient_to_total_player")
	final := r.PostFormValue("final_percentage")
	correction := r.PostFormValue("data_correction")

	var req GameDataReq
	req.GameId = r.FormValue("game_id")

	selector := bson.M{}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}
	sur, err := db.GetSurPoolData(selector)
	if err != nil {
		return
	}

	var upt UpSurPool
	upt.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
	upt.PercentageToTotalWin = sur.PercentageToTotalWin
	upt.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
	upt.FinalPercentage = sur.FinalPercentage
	upt.DataCorrection = sur.DataCorrection

	if rateSur != "" {
		upt.PlayerLoseRateAfterSurplusPool, _ = strconv.ParseFloat(rateSur, 64)
		sur.PlayerLoseRateAfterSurplusPool = upt.PlayerLoseRateAfterSurplusPool
	}
	if percentage != "" {
		upt.PercentageToTotalWin, _ = strconv.ParseFloat(percentage, 64)
		sur.PercentageToTotalWin = upt.PercentageToTotalWin
	}
	if coefficient != "" {
		upt.CoefficientToTotalPlayer, _ = strconv.ParseInt(coefficient, 10, 64)
		sur.CoefficientToTotalPlayer = upt.CoefficientToTotalPlayer
	}
	if final != "" {
		upt.FinalPercentage, _ = strconv.ParseFloat(final, 64)
		sur.FinalPercentage = upt.FinalPercentage
	}
	if correction != "" {
		upt.DataCorrection, _ = strconv.ParseFloat(correction, 64)
		sur.DataCorrection = upt.DataCorrection
	}

	sur.SurplusPool = (sur.PlayerTotalLose - (sur.PlayerTotalWin * sur.PercentageToTotalWin) - float64(sur.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
	// 更新盈余池数据
	_ = db.UpdateSurPool(&sur)

	js, err := json.Marshal(NewResp(SuccCode, "", upt))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getRobotData(w http.ResponseWriter, r *http.Request) {
	recodes, err := db.GetRobotData()
	if err != nil {
		return
	}

	var rData []GRobotData
	for i := 0; i < len(recodes); i++ {
		var rd GRobotData
		rd.AreaX1 = new(ChipDownBet)
		rd.AreaX2 = new(ChipDownBet)
		rd.AreaX3 = new(ChipDownBet)
		rd.AreaX4 = new(ChipDownBet)
		rd.AreaX5 = new(ChipDownBet)
		rd.AreaX6 = new(ChipDownBet)
		rd.AreaX7 = new(ChipDownBet)
		rd.AreaX8 = new(ChipDownBet)
		pr := recodes[i]
		log.Debug("获取机器数据:%v", pr)
		rd.RoomId = pr.RoomId
		rd.RoomTime = pr.RoomTime
		rd.RobotNum = pr.RobotNum
		rd.AreaX1 = pr.AreaX1
		rd.AreaX2 = pr.AreaX2
		rd.AreaX3 = pr.AreaX3
		rd.AreaX4 = pr.AreaX4
		rd.AreaX5 = pr.AreaX5
		rd.AreaX6 = pr.AreaX6
		rd.AreaX7 = pr.AreaX7
		rd.AreaX8 = pr.AreaX8
		rData = append(rData, rd)
	}

	js, err := json.Marshal(NewResp(SuccCode, "", rData))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

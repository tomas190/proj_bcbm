package internal

import "time"

type SettleDB struct {
	User      UserDB  `bson:"User"`
	WinOrder  string  `bson:"WinOrder"`
	RoundID   string  `bson:"RoundID"`
	IsWin     bool    `bson:"IsWin"`
	BetAmount float64 `bson:"BetAmount"`
	WinAmount float64 `bson:"WinAmount"`
}

type UserDB struct {
	UserID   uint32  `bson:"UserID" json:"UserID"`     // 用户id
	NickName string  `bson:"NickName" json:"NickName"` // 用户昵称
	Avatar   string  `bson:"Avatar" json:"Avatar"`     // 用户头像
	Balance  float64 `bson:"Balance" json:"Balance"`   // 用户金额
}

type ProfitDB struct {
	UpdateTime     time.Time `bson:"UpdateTime"`
	UpdateTimeStr  string    `bson:"UpdateTimeStr"`
	RoomID         uint32    `bson:"RoomID"`
	PlayerThisWin  float64   `bson:"PlayerThisWin"`
	PlayerThisLost float64   `bson:"PlayerThisLost"`
	PlayerAllWin   float64   `bson:"PlayerAllWin"`
	PlayerAllLost  float64   `bson:"PlayerAllLost"`
	Profit         float64   `bson:"Profit"`
	PlayerNum      uint32    `bson:"PlayerNum"`
}

type SurPool struct {
	GameId                         string  `json:"game_id" bson:"game_id"`
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
	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
	RandomCountAfterWin            float64 `json:"random_count_after_win" bson:"random_count_after_win"`
	RandomCountAfterLose           float64 `json:"random_count_after_lose" bson:"random_count_after_lose"`
	RandomPercentageAfterWin       float64 `json:"random_percentage_after_win" bson:"random_percentage_after_win"`
	RandomPercentageAfterLose      float64 `json:"random_percentage_after_lose" bson:"random_percentage_after_lose"`
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id              string    `json:"id" bson:"id"`                             // 玩家Id
	GameId          string    `json:"game_id" bson:"game_id"`                   // GameId
	RoundId         string    `json:"round_id" bson:"round_id"`                 // 随机Id
	RoomId          uint32    `json:"room_id" bson:"room_id"`                   // 所在房间
	DownBetInfo     []float64 `json:"down_bet_info" bson:"down_bet_info"`       // 8个注池个下注金额
	DownBetTime     int64     `json:"down_bet_time" bson:"down_bet_time"`       // 下注时间
	StartTime       int64     `json:"start_time" bson:"start_time"`             // 开始时间
	EndTime         int64     `json:"end_time" bson:"end_time"`                 // 结束时间
	CardResult      uint32    `json:"card_result" bson:"card_result"`           // 当局开牌结果
	SettlementFunds float64   `json:"settlement_funds" bson:"settlement_funds"` // 当局输赢结果(税后)
	SpareCash       float64   `json:"spare_cash" bson:"spare_cash"`             // 剩余金额
	TaxRate         float64   `json:"tax_rate" bson:"tax_rate"`                 // 税率
}

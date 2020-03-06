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
	player_total_lose                   float64
	player_total_win                    float64
	percentage_to_total_win             float64
	total_player                        int64
	coefficient_to_total_player         int64
	final_percentage                    float64
	player_lose_rate_after_surplus_pool float64
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id          uint32    `json:"id" bson:"id"`                       // 玩家Id
	RandId      uint32    `json:"rand_id" bson:"rand_id"`             // 随机Id
	RoomId      uint32    `json:"room_id" bson:"room_id"`             // 所在房间
	DownBetInfo []float64 `json:"down_bet_info" bson:"down_bet_info"` // 8个注池个下注金额
	DownBetTime int64     `json:"down_bet_time" bson:"down_bet_time"` // 下注时间
	CardResult  uint32    `json:"card_result" bson:"card_result"`     // 当局开牌结果
	ResultMoney float64   `json:"result_money" bson:"result_money"`   // 当局输赢结果(税后)
	TaxRate     float64   `json:"tax_rate" bson:"tax_rate"`           // 税率
}

package constant

// 房间状态 room status
const (
	_         = iota
	RSBetting // 下注
	RSSettle  // 结算
	RSClear   // 清理筹码
)

// 时间间隔
const (
	BetTime    = 16 //16
	SettleTime = 25 //24
	ClearTime  = 3
)

// 最大容量
const MaxPlayerCount = 100
const RoomMaxBonus = 100000

// 房间等级 room level
const (
	RL1MinBet   = 1
	RL1MaxBet   = 10000
	RL1MinLimit = 50
)


var ChipSize = map[uint32]float64{
	1: 1,
	2: 10,
	3: 100,
	4: 500,
	5: 1000,
}

var AreaX = map[uint32]float64{
	0: 0,
	1: 40,
	2: 30,
	3: 20,
	4: 10,
	5: 5,
	6: 5,
	7: 5,
	8: 5,
}

var AreaName = map[uint32]string{
	0: "PlaceHolder",
	1: "GoldenBenz-40X",
	2: "GoldenBMW-30X",
	3: "GoldenAudi-20X",
	4: "GoldenVW-10X",
	5: "Benz-5X",
	6: "BMW-5X",
	7: "Audi-5X",
	8: "VW-5X",
}

package constant

// 房间状态
const (
	_         = iota
	RSBetting // 下注
	RSSettle  // 结算
	RSClear   // 清理筹码
)

const (
	BetTime    = 16 //16
	SettleTime = 24 //24
	ClearTime  = 3
)

const MaxPlayerCount = 100

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

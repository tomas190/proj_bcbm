package constant

// 房间状态
const (
	_         = iota
	RSBetting // 下注
	RSSettle  // 结算
	RSClear   // 清理筹码
)

const (
	BetTime    = 3 //15
	SettleTime = 3 //23
	ClearTime  = 3
)

const (
	RL1MinBet   = 1
	RL1MaxBet   = 10000
	RL1MinLimit = 50
)

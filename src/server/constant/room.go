package constant

// 房间状态
const (
	_ = iota
	RSBetting
	RSLottery
	RSClear
)

const (
	BetTime     = 5
	LotteryTime = 6
	ClearTime   = 3
)

const (
	RL1MinBet   = 1
	RL1MaxBet   = 10000
	RL1MinLimit = 50
)

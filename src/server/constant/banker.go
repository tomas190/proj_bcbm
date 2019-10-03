package constant

// 庄家配置
const (
	BankerMinBar   = 50000
	BankerMaxTimes = 5
	CancelGrab     = -1 // 取消申请上庄
	DownBanker     = -2 // 申请下庄

	BSNotBanker      = 0 // 非庄家
	BSGrabbingBanker = 1 // 正在申请上庄
	BSBeingBanker    = 2 // 正在坐庄
)

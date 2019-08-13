package lottery

import (
	"fmt"
	"testing"
)

// 测试公平开奖
func TestOpenAward(t *testing.T) {
	count := map[uint32]int{
		1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,
	}
	l := &Lottery{}
	for i := 0; i < 121000; i++ {
		count[l.fairLottery()]++
	}

	fmt.Println(count)
}

func TestLottery_ProfitPoolLottery(t *testing.T) {
	count := map[uint32]int{
		1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,
	}
	l := &Lottery{}
	for i := 0; i < 20; i++ {
		count[l.ProfitPoolLottery()]++
	}

	fmt.Println(count)
}


package internal

import (
	"fmt"
	"testing"
	"time"
)

// 测试公平开奖
func TestRoom_fairLottery(t *testing.T) {
	count := map[uint32]int{
		1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,
	}
	r := &Room{}
	for i := 0; i < 121000; i++ {
		count[r.fairLottery()]++
	}

	fmt.Println(count)
}

func TestLottery_ProfitPoolLottery(t *testing.T) {
	count := map[uint32]int{
		1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,
	}
	r := &Room{}
	for i := 0; i < 20; i++ {
		count[r.ProfitPoolLottery()]++
	}

	fmt.Println(count)
}

func TestRoom_ProfitPoolLottery(t *testing.T) {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		c1 <- "one"
	}()

	go func() {
		time.Sleep(2 * time.Second)
		c2 <- "two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}

// 15-开始下注-0 停止下注  下注
// 23 22 21 - 跑马灯 - 随便什么时候开奖 显示 4 3 2 1 0  开奖
// 2 1 0 清空筹码




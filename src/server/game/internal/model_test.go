package internal

import (
	"fmt"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/util"
	"testing"
	"time"
)

// 测试公平开奖
func TestRoom_fairLottery(t *testing.T) {
	count := map[uint32]int{
		1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0, 8: 0,
	}
	dl := &Dealer{}
	for i := 0; i < 121000; i++ {
		count[dl.fairLottery()]++
	}

	fmt.Println(count)
}

func TestLottery_ProfitPoolLottery(t *testing.T) {
	count := map[uint32]int{
		1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0, 8: 0,
	}
	dl := &Dealer{}
	for i := 0; i < 20; i++ {
		count[dl.profitPoolLottery()]++
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

func TestDealer_TimeMachine(t *testing.T) {
	hr := make(chan HRMsg)
	dl := NewDealer(123, hr)
	dl.StartGame()
	time.Sleep(20 * time.Second)
}

var interval = time.Second

// 调度实验
func TestNewDealer(t *testing.T) {

	schedule := map[int]interface{}{
		5:  printNotToday,
		10: printNotToday,
		15: printNotToday,
	}
	var ahead int
	var counter int

	ticker := time.NewTicker(interval)
	go func() {
		for t := range ticker.C {
			fmt.Println(t)
			counter++
			fmt.Println(counter)
			// 需要移动时间轴
			if next, ok := schedule[counter+ahead]; ok {
				next.(func(uint32))(123)
			}
		}
	}()

	go func() {
		time.Sleep(2 * interval)
		// 事件到了
		printNotToday(123)
		latestEventTime := min(schedule)
		// 验证过数据之后，删除
		delete(schedule, latestEventTime)
		// 找到最小的key和当前count
		ahead = ahead + latestEventTime - counter // 5 - 2 // 提前到达时差 会出现数据冲突的问题

	}()

	time.Sleep(20 * time.Second)
}

func printNotToday(num uint32) {
	fmt.Println("Not Today!", num)
}

func min(numbers map[int]interface{}) int {
	var minNumber int
	for minNumber = range numbers {
		break
	}
	for n := range numbers {
		if n < minNumber {
			minNumber = n
		}
	}
	return minNumber
}

func TestDealer_ClockReset(t *testing.T) {
	ClockReset(5, func() {
		fmt.Println("fuck")
	})
}

// 重置表
func ClockReset(duration uint32, next func()) {
	ticker := time.NewTicker(interval)
	var counter uint32

	log.Debug("clock重置 deadline: %v event: %v", duration, util.Function{}.GetFunctionName(next))
	go func() {
		for t := range ticker.C {
			log.Debug("时间滴答：%v", t)
			counter++
			if duration == counter {
				next()
				break
			}
		}
	}()
}

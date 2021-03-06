package util

import (
	"math/rand"
	"time"
)

type Random struct{}

// return random int between [min, max)
func (r *Random) RandInRange(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}

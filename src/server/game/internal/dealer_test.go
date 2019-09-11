package internal

import (
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)

func TestDealer_Settle(t *testing.T) {
	// 精度问题
	// case1: 135.90*100 ====
	// float32
	var f1 float32 = 135.90
	fmt.Println(f1 * 100) // output:13589.999
	// float64
	var f2 float64 = 135.90
	fmt.Println(f2 * 100) // output:13590

	// case2: 0.1 add 10 times ===
	// float32
	var f3 float32 = 0
	for i := 0; i < 10; i++ {
		f3 += 0.1
	}
	fmt.Println(f3) //output:1.0000001
	// float64
	var f4 float64 = 0
	for i := 0; i < 10; i++ {
		f4 += 0.1
	}
	fmt.Println(f4) //output:0.9999999999999999

	// case: float32==>float64
	// 从数据库中取出80.45, 历史代码用float32接收
	var a float32 = 80.45
	var b float64
	// 有些函数只能接收float64, 只能强转
	b = float64(a)
	// 打印出值, 强转后出现偏差
	fmt.Println(a) //output:80.45
	fmt.Println(b) //output:80.44999694824219
	// ... 四舍五入保留小数点后1位, 期望80.5, 结果是80.4
	// case: int64==>float64
	var c int64 = 987654321098765432
	fmt.Printf("%.f\n", float64(c)) //output:987654321098765440
	// case: int(float64(xx.xx*100))
	var d float64 = 1129.6
	var e int64 = int64(d * 100)
	fmt.Println(e) //output:112959

	price, err := decimal.NewFromString("136.02")
	if err != nil {
		panic(err)
	}

	quantity := decimal.NewFromFloat(3)
	fee, _ := decimal.NewFromString(".035")
	taxRate, _ := decimal.NewFromString(".08875")
	subtotal := price.Mul(quantity)
	preTax := subtotal.Mul(fee.Add(decimal.NewFromFloat(1)))
	total := preTax.Mul(taxRate.Add(decimal.NewFromFloat(1)))
	fmt.Println("Subtotal:", subtotal)                      // Subtotal: 408.06
	fmt.Println("Pre-tax:", preTax)                         // Pre-tax: 422.3421
	fmt.Println("Taxes:", total.Sub(preTax))                // Taxes: 37.482861375
	fmt.Println("Total:", total)                            // Total: 459.824961375
	fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax)) // Tax rate: 0.08875
}

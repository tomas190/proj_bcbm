package util

import (
	"github.com/shopspring/decimal"
)

type Math struct{}

func (m *Math) SumSliceFloat64(s []float64) decimal.Decimal {
	sum := decimal.NewFromFloat(0)
	for i := range s {
		fd := decimal.NewFromFloat(s[i])
		sum = sum.Add(fd)
	}
	return sum
}

func (m *Math) AddFloat64(a float64, b float64) decimal.Decimal {
	ad := decimal.NewFromFloat(a)
	bd := decimal.NewFromFloat(b)

	return ad.Add(bd)
}

func (m *Math) MultiFloat64(a float64, b float64) decimal.Decimal {
	ad := decimal.NewFromFloat(a)
	bd := decimal.NewFromFloat(b)

	return ad.Mul(bd)
}

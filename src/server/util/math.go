package util

type Math struct{}

func (m *Math) SumSliceFloat64(s []float64) float64 {
	var sum float64
	for _, f := range s {
		sum += f
	}

	return sum
}

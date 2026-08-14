package splitter

import "math"

func round(data float64) int64 {
	return int64(math.Round(data))
}

func minusPercent(value int64, mdr float64) int64 {
	result := float64(value) - (float64(value) * mdr / 100)

	return round(result)
}

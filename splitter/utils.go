package splitter

import "math"

// sharePercentEpsilon absorbs float64 representation error when summing
// percentages. Sums that are exact in decimal are not always exact in binary:
// 0.1 + 64.1 + 35.8 evaluates to 99.99999999999998.
const sharePercentEpsilon = 1e-9

func round(data float64) int64 {
	return int64(math.Round(data))
}

func minusPercent(value int64, mdr float64) int64 {
	result := float64(value) - (float64(value) * mdr / 100)

	return round(result)
}

// totalSharePercent sums the percentages of a set of split rules.
func totalSharePercent(rules []*SplitRule) float64 {
	total := float64(0)
	for _, rule := range rules {
		total += rule.Amount
	}

	return total
}

// isFullShare reports whether the percentages allocate the item exactly once.
func isFullShare(percent float64) bool {
	return math.Abs(percent-100) < sharePercentEpsilon
}

// reconcileIndex returns the index of the result that should absorb a rounding
// remainder: the highest-cost item with at least one share. Charging the
// remainder to a zero-cost item would hand revenue to a giveaway line.
func reconcileIndex[T any](results []*SplitResult[T]) int {
	target := -1
	for i, result := range results {
		if len(result.Shares) == 0 {
			continue
		}

		if target == -1 || result.Cost > results[target].Cost {
			target = i
		}
	}

	return target
}

package splitter

import (
	"errors"
)

func BreakdownCash[T any](params SplitParams[T]) ([]*SplitResult[T], error) {
	results := make([]*SplitResult[T], 0)

	sharedTotal := int64(0)
	sharedNetTotal := int64(0)

	for _, split := range params.Splits {
		cost := split.Price * split.Quantity
		currentResult := &SplitResult[T]{
			ItemID:  split.ItemID,
			Cost:    cost,
			NetCost: cost,
		}

		sharedTotal += currentResult.Cost
		sharedNetTotal += currentResult.NetCost

		currentSharePercent := float64(0)
		currentRevenue := int64(0)

		for _, share := range split.SplitRules {
			currentSharePercent += share.Amount
			revenue := round(float64(cost) * share.Amount / float64(100))
			currentRevenue += revenue

			currentShare := &Share{
				DepartmentID: share.DepartmentID,
				GrossRevenue: revenue,
				NetRevenue:   revenue,
			}

			currentResult.Shares = append(currentResult.Shares, currentShare)
		}

		if currentRevenue != currentResult.NetCost {
			diff := currentResult.NetCost - currentRevenue
			if len(currentResult.Shares) > 0 {
				currentResult.Shares[0].NetRevenue = currentResult.Shares[0].NetRevenue + diff
				currentResult.Shares[0].GrossRevenue = currentResult.Shares[0].GrossRevenue + diff
			}
		}

		if currentSharePercent != 100 {
			return nil, errors.New("total share must be 100%")
		}

		results = append(results, currentResult)
	}

	if sharedTotal != params.GrossRevenue || sharedNetTotal != params.NetRevenue {
		return nil, errors.New("items total must equal to purchase total")
	}

	return results, nil
}

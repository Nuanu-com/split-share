package splitter

import (
	"errors"
	"fmt"
)

type Cash[T any] struct{}

// Breakdown implements [defs.Splitter].
func (c *Cash[T]) Breakdown(params SplitParams[T]) ([]*SplitResult[T], error) {
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
			fmt.Printf("revenue: %d \n", revenue)
			fmt.Printf("cost: %d \n", cost)
			fmt.Printf("percent amount: %f \n", share.Amount)

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

func NewCash[T any]() Splitter[T] {
	return &Cash[T]{}
}

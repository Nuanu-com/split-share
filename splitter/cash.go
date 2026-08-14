package splitter

func BreakdownCash[T any](params SplitParams[T]) ([]*SplitResult[T], error) {
	results := make([]*SplitResult[T], 0, len(params.Splits))

	sharedTotal := int64(0)
	sharedNetTotal := int64(0)

	for _, split := range params.Splits {
		if !isFullShare(totalSharePercent(split.SplitRules)) {
			return nil, ErrSharePercentNot100
		}

		cost := split.Price * split.Quantity
		currentResult := &SplitResult[T]{
			ItemID:  split.ItemID,
			Cost:    cost,
			NetCost: cost,
		}

		sharedTotal += currentResult.Cost
		sharedNetTotal += currentResult.NetCost

		currentRevenue := int64(0)

		for _, share := range split.SplitRules {
			revenue := round(float64(cost) * share.Amount / float64(100))
			currentRevenue += revenue

			currentShare := &Share{
				DepartmentID: share.DepartmentID,
				GrossRevenue: revenue,
				NetRevenue:   revenue,
			}

			currentResult.Shares = append(currentResult.Shares, currentShare)
		}

		// Per-share rounding can drift from the item total; charge the
		// difference to the first share so the shares always add back up.
		if currentRevenue != currentResult.NetCost && len(currentResult.Shares) > 0 {
			diff := currentResult.NetCost - currentRevenue
			currentResult.Shares[0].NetRevenue += diff
			currentResult.Shares[0].GrossRevenue += diff
		}

		results = append(results, currentResult)
	}

	if sharedTotal != params.GrossRevenue || sharedNetTotal != params.NetRevenue {
		return nil, ErrTotalMismatch
	}

	return results, nil
}

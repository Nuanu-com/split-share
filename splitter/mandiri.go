package splitter

type MandiriParams[T any] struct {
	GrossAmount int64
	NetAmount   int64
	MDR         float64
	MDRAmount   int64
	DCC         float64
	DCCAmount   int64
	Splits      []*ItemSplit[T] `json:"splits"`
}

func BreakdownMandiri[T any](params MandiriParams[T]) ([]*SplitResult[T], error) {
	results := make([]*SplitResult[T], 0, len(params.Splits))

	totalCost := int64(0)
	totalNetCost := int64(0)

	for _, item := range params.Splits {
		if !isFullShare(totalSharePercent(item.SplitRules)) {
			return nil, ErrSharePercentNot100
		}

		cost := item.Price * item.Quantity
		netCost := minusPercent(cost, params.MDR) + round(float64(cost)*params.DCC/100)

		totalCost += cost
		totalNetCost += netCost

		currentResult := &SplitResult[T]{
			ItemID:  item.ItemID,
			Cost:    cost,
			NetCost: netCost,
		}

		totalGrossShare := int64(0)
		totalNetShare := int64(0)

		for _, splitRule := range item.SplitRules {
			grossRevenue := round(float64(cost) * splitRule.Amount / 100)
			netRevenue := round(float64(netCost) * splitRule.Amount / 100)

			totalGrossShare += grossRevenue
			totalNetShare += netRevenue

			share := &Share{
				DepartmentID: splitRule.DepartmentID,
				GrossRevenue: grossRevenue,
				NetRevenue:   netRevenue,
			}

			currentResult.Shares = append(currentResult.Shares, share)
		}

		if cost != totalGrossShare && len(currentResult.Shares) > 0 {
			diff := cost - totalGrossShare
			currentResult.Shares[0].GrossRevenue += diff
		}

		if netCost != totalNetShare && len(currentResult.Shares) > 0 {
			diff := netCost - totalNetShare
			currentResult.Shares[0].NetRevenue += diff
		}

		results = append(results, currentResult)
	}

	if totalCost != params.GrossAmount {
		return nil, ErrTotalMismatch
	}

	// Per-item MDR/DCC rounding drifts from the amount the vendor actually
	// settled; charge the difference to the largest item that has shares.
	expectedNet := params.NetAmount + params.DCCAmount
	if totalNetCost != expectedNet {
		if target := reconcileIndex(results); target != -1 {
			diff := expectedNet - totalNetCost
			results[target].NetCost += diff
			results[target].Shares[0].NetRevenue += diff
		}
	}

	return results, nil
}

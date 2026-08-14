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
	results := make([]*SplitResult[T], 0)

	totalCost := int64(0)
	totalNetCost := int64(0)

	for _, item := range params.Splits {
		cost := item.Price * item.Quantity
		netCost := minusPercent(item.Price*item.Quantity, params.MDR) + round(float64(cost)*params.DCC/100)

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

	if totalNetCost != (params.NetAmount+params.DCCAmount) && len(results) > 0 {
		diff := (params.NetAmount + params.DCCAmount) - totalNetCost
		results[0].NetCost += diff

		if len(results[0].Shares) > 0 {
			results[0].Shares[0].NetRevenue += diff
		}
	}

	return results, nil
}

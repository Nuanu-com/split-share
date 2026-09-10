package splitter

import "github.com/google/uuid"

type MandiriParams[T any] struct {
	GrossAmount int64
	NetAmount   int64
	MDR         float64
	MDRAmount   int64
	DCC         float64
	DCCAmount   int64
	Splits      []*ItemSplit[T] `json:"splits"`
	// TransactionOwner, when set, is the department that carries the whole MDR
	// instead of it being spread over every share of the item.
	TransactionOwner *uuid.UUID
}

func BreakdownMandiri[T any](params MandiriParams[T]) ([]*SplitResult[T], error) {
	if params.TransactionOwner != nil && *params.TransactionOwner == uuid.Nil {
		return nil, ErrNoTransactionOwner
	}

	results := make([]*SplitResult[T], 0, len(params.Splits))

	totalCost := int64(0)
	totalNetCost := int64(0)

	for _, item := range params.Splits {
		if !isFullShare(totalSharePercent(item.SplitRules)) {
			return nil, ErrSharePercentNot100
		}

		cost := item.Price * item.Quantity
		mdrPart := cost - minusPercent(cost, params.MDR)
		dccPart := round(float64(cost) * params.DCC / 100)
		netCost := cost - mdrPart + dccPart

		totalCost += cost
		totalNetCost += netCost

		currentResult := &SplitResult[T]{
			ItemID:  item.ItemID,
			Cost:    cost,
			NetCost: netCost,
		}

		// With an owner the MDR leaves the prorated base: every share keeps its
		// gross plus its part of the DCC, and the owner alone is charged the
		// item's MDR below.
		netShareBase := netCost
		if params.TransactionOwner != nil {
			netShareBase = cost + dccPart
		}

		totalGrossShare := int64(0)
		totalNetShare := int64(0)

		var ownerShare *Share

		for _, splitRule := range item.SplitRules {
			grossRevenue := round(float64(cost) * splitRule.Amount / 100)
			netRevenue := round(float64(netShareBase) * splitRule.Amount / 100)

			totalGrossShare += grossRevenue
			totalNetShare += netRevenue

			share := &Share{
				DepartmentID: splitRule.DepartmentID,
				GrossRevenue: grossRevenue,
				NetRevenue:   netRevenue,
			}

			// The owner may legitimately hold several rules on one item; the MDR
			// is a per-item cost and must only be charged once.
			if params.TransactionOwner != nil && ownerShare == nil && splitRule.DepartmentID == *params.TransactionOwner {
				ownerShare = share
			}

			currentResult.Shares = append(currentResult.Shares, share)
		}

		if cost != totalGrossShare && len(currentResult.Shares) > 0 {
			diff := cost - totalGrossShare
			currentResult.Shares[0].GrossRevenue += diff
		}

		if netShareBase != totalNetShare && len(currentResult.Shares) > 0 {
			diff := netShareBase - totalNetShare
			currentResult.Shares[0].NetRevenue += diff
		}

		if params.TransactionOwner != nil {
			if ownerShare == nil {
				ownerShare = &Share{DepartmentID: *params.TransactionOwner}
				currentResult.Shares = append(currentResult.Shares, ownerShare)
			}

			// The transaction owner carries the whole MDR for the item.
			ownerShare.NetRevenue -= mdrPart
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

			// The drift is an MDR difference, so it belongs to whoever carries
			// the MDR: the owner when there is one, the first share otherwise.
			absorber := results[target].Shares[0]
			if params.TransactionOwner != nil {
				absorber = shareOfOwner(results[target], *params.TransactionOwner)
			}

			absorber.NetRevenue += diff
		}
	}

	return results, nil
}

// shareOfOwner returns the share holding the transaction owner's revenue, which
// is the one charged the item's MDR. Owner mode always adds it, so the fallback
// to the first share only guards against that changing.
func shareOfOwner[T any](result *SplitResult[T], owner uuid.UUID) *Share {
	for _, share := range result.Shares {
		if share.DepartmentID == owner {
			return share
		}
	}

	return result.Shares[0]
}

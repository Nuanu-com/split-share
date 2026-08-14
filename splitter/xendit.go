package splitter

import (
	"github.com/google/uuid"
)

type XenditParams[T any] struct {
	GrossAmount      int64
	NetAmount        int64
	MDRAmount        int64
	TransactionOwner uuid.UUID
	Splits           []*ItemSplit[T] `json:"splits"`
}

func BreakdownXendit[T any](params XenditParams[T]) ([]*SplitResult[T], error) {
	results := make([]*SplitResult[T], 0, len(params.Splits))

	if params.TransactionOwner == uuid.Nil {
		return nil, ErrNoTransactionOwner
	}

	if len(params.Splits) == 0 {
		return results, nil
	}

	// MDRPart prorates against the transaction total, so it has to be usable as
	// a divisor.
	if params.GrossAmount <= 0 {
		return nil, ErrInvalidGrossAmount
	}

	costs := make([]int64, len(params.Splits))
	totalCost := int64(0)

	for i, item := range params.Splits {
		if !isFullShare(totalSharePercent(item.SplitRules)) {
			return nil, ErrSharePercentNot100
		}

		costs[i] = item.Price * item.Quantity
		totalCost += costs[i]
	}

	if totalCost != params.GrossAmount {
		return nil, ErrTotalMismatch
	}

	mdrParts := distributeMDR(costs, params.GrossAmount, params.MDRAmount)

	for i, item := range params.Splits {
		cost := costs[i]
		mdrPart := mdrParts[i]

		currentResult := &SplitResult[T]{
			ItemID:  item.ItemID,
			Cost:    cost,
			NetCost: cost - mdrPart,
		}

		var ownerShare *Share
		totalGrossShare := int64(0)

		for _, splitRule := range item.SplitRules {
			splitCost := round(float64(cost) * splitRule.Amount / 100)
			totalGrossShare += splitCost

			share := &Share{
				DepartmentID: splitRule.DepartmentID,
				GrossRevenue: splitCost,
				NetRevenue:   splitCost,
			}

			// The owner may legitimately hold several rules on one item; the MDR
			// is a per-item cost and must only be charged once.
			if ownerShare == nil && splitRule.DepartmentID == params.TransactionOwner {
				ownerShare = share
			}

			currentResult.Shares = append(currentResult.Shares, share)
		}

		if totalGrossShare != cost && len(currentResult.Shares) > 0 {
			diff := cost - totalGrossShare
			currentResult.Shares[0].GrossRevenue += diff
			currentResult.Shares[0].NetRevenue += diff
		}

		if ownerShare == nil {
			ownerShare = &Share{DepartmentID: params.TransactionOwner}
			currentResult.Shares = append(currentResult.Shares, ownerShare)
		}

		// The transaction owner carries the whole MDR for the item.
		ownerShare.NetRevenue -= mdrPart

		results = append(results, currentResult)
	}

	return results, nil
}

// distributeMDR splits mdrAmount across items proportionally to their cost.
// Rounding each item independently loses or invents cents, so the remainder is
// charged to the largest item and the parts always add back up to mdrAmount.
func distributeMDR(costs []int64, totalTrx int64, mdrAmount int64) []int64 {
	parts := make([]int64, len(costs))

	allocated := int64(0)
	largest := -1

	for i, cost := range costs {
		parts[i] = MDRPart(totalTrx, cost, mdrAmount)
		allocated += parts[i]

		if cost > 0 && (largest == -1 || cost > costs[largest]) {
			largest = i
		}
	}

	if largest != -1 && allocated != mdrAmount {
		parts[largest] += mdrAmount - allocated
	}

	return parts
}

// Need to distribute the MDR cost fairly across the items
func MDRPart(totalTrx int64, itemCost int64, mdrAmount int64) int64 {
	if totalTrx == 0 {
		return 0
	}

	return round(float64(mdrAmount) * float64(itemCost) / float64(totalTrx))
}

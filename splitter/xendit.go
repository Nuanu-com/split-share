package splitter

import (
	"errors"

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
	results := make([]*SplitResult[T], 0)

	if params.TransactionOwner == uuid.Nil {
		return nil, errors.New("transaction owner need to be defined")
	}

	for _, item := range params.Splits {
		cost := item.Price * item.Quantity
		mdrPart := MDRPart(params.GrossAmount, cost, params.MDRAmount)
		netCost := cost - mdrPart

		currentResult := &SplitResult[T]{
			ItemID:  item.ItemID,
			Cost:    cost,
			NetCost: netCost,
		}

		var ownerShare *Share

		for _, splitRule := range item.SplitRules {
			splitCost := round(float64(cost) * splitRule.Amount / 100)
			share := &Share{
				DepartmentID: splitRule.DepartmentID,
				GrossRevenue: splitCost,
				NetRevenue:   splitCost,
			}

			if splitRule.DepartmentID == params.TransactionOwner {
				ownerShare = share
				share.NetRevenue -= mdrPart
			}

			currentResult.Shares = append(currentResult.Shares, share)
		}

		if ownerShare == nil {
			currentResult.Shares = append(currentResult.Shares, &Share{
				DepartmentID: params.TransactionOwner,
				GrossRevenue: 0,
				NetRevenue:   -mdrPart,
			})
		}

		results = append(results, currentResult)
	}

	return results, nil
}

// Need to distribute the MDR cost fairly across the items
func MDRPart(totalTrx int64, itemCost int64, mdrAmount int64) int64 {
	return round(float64(mdrAmount) * float64(itemCost) / float64(totalTrx))
}

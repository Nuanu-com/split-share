package splitter_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/Nuanu-com/split-share/splitter"
)

func TestCashBreakdownPercentNot100(t *testing.T) {
	c := splitter.NewCash[uuid.UUID]()

	itemID1 := uuid.New()

	result, err := c.Breakdown(splitter.SplitParams[uuid.UUID]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[uuid.UUID]{
			{
				ItemID:   itemID1,
				Price:    40_000,
				Quantity: 2,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: uuid.New(),
						Amount:       60,
					},
					{
						DepartmentID: uuid.New(),
						Amount:       30,
					},
				},
			},
		},
	})

	if result != nil {
		t.Error("result should be nil")
	}

	if err.Error() != "total share must be 100%" {
		t.Errorf(`got: %s, wants: %s`, err.Error(), "error total share must be 100%")
	}
}

func TestCashBreakdownNotEqualToTotal(t *testing.T) {
	c := splitter.NewCash[uuid.UUID]()

	itemID1 := uuid.New()

	result, err := c.Breakdown(splitter.SplitParams[uuid.UUID]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[uuid.UUID]{
			{
				ItemID:   itemID1,
				Price:    40_000,
				Quantity: 2,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: uuid.New(),
						Amount:       60,
					},
					{
						DepartmentID: uuid.New(),
						Amount:       40,
					},
				},
			},
		},
	})

	if err == nil {
		t.Error("error expected to occur")
		return
	}

	if result != nil {
		t.Error("result should be nil")
		return
	}

	if err.Error() != "items total must equal to purchase total" {
		t.Errorf(`got: %s, wants: %s`, err.Error(), "items total must equal to purchase total")
		return
	}
}

func TestCashHandleRounding(t *testing.T) {
	c := splitter.NewCash[uuid.UUID]()
	itemID1 := uuid.New()

	result, err := c.Breakdown(splitter.SplitParams[uuid.UUID]{
		GrossRevenue:  75,
		NetRevenue:    75,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[uuid.UUID]{
			{
				ItemID:   itemID1,
				Price:    75,
				Quantity: 1,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: uuid.New(),
						Amount:       50,
					},
					{
						DepartmentID: uuid.New(),
						Amount:       50,
					},
				},
			},
		},
	})

	if err != nil {
		t.Errorf("got: %s, wants: nil", err.Error())
	}

	if len(result) != 1 {
		t.Errorf("got: %d, wants: 1", len(result))
	}

	totalShare := int64(0)
	netTotal := int64(0)

	for _, r := range result {
		netTotal += r.NetCost
		for _, s := range r.Shares {
			totalShare += s.NetRevenue
		}
	}

	if totalShare != netTotal {
		t.Errorf("got: %d, wants: %d", totalShare, netTotal)
	}
}

func TestCashBreakdownSuccess(t *testing.T) {
	c := splitter.NewCash[uuid.UUID]()

	itemID1 := uuid.New()
	itemID2 := uuid.New()

	result, err := c.Breakdown(splitter.SplitParams[uuid.UUID]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[uuid.UUID]{
			{
				ItemID:   itemID1,
				Price:    40_000,
				Quantity: 2,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: uuid.New(),
						Amount:       60,
					},
					{
						DepartmentID: uuid.New(),
						Amount:       40,
					},
				},
			},

			{
				ItemID:   itemID2,
				Price:    10_000,
				Quantity: 2,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: uuid.New(),
						Amount:       50,
					},
					{
						DepartmentID: uuid.New(),
						Amount:       50,
					},
				},
			},
		},
	})

	if err != nil {
		t.Errorf("got: %s, wants: nil", err.Error())
	}

	if len(result) != 2 {
		t.Errorf("got: %d, wants: 2", len(result))
	}
}

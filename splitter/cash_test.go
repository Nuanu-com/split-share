package splitter_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Nuanu-com/split-share/splitter"
)

func TestCashBreakdownPercentNot100(t *testing.T) {

	itemID1 := 1

	result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[int]{
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

	itemID1 := 1

	result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[int]{
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
	itemID1 := 1

	result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue:  75,
		NetRevenue:    75,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[int]{
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

	itemID1 := 1
	itemID2 := 2

	result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[int]{
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

// 0.1 + 64.1 + 35.8 is exactly 100 in decimal but 99.99999999999998 in float64,
// so an exact comparison rejects a legitimate split.
func TestCashPercentWithFloatRepresentationError(t *testing.T) {
	result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue:  100_000,
		NetRevenue:    100_000,
		PaymentVendor: splitter.VendorCash,
		Splits: []*splitter.ItemSplit[int]{
			{
				ItemID:   1,
				Price:    100_000,
				Quantity: 1,
				SplitRules: []*splitter.SplitRule{
					{DepartmentID: uuid.New(), Amount: 0.1},
					{DepartmentID: uuid.New(), Amount: 64.1},
					{DepartmentID: uuid.New(), Amount: 35.8},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("got: %s, wants: nil", err.Error())
	}

	total := int64(0)
	for _, s := range result[0].Shares {
		total += s.GrossRevenue
	}

	if total != 100_000 {
		t.Errorf("got: %d, wants: 100000", total)
	}
}

func TestCashSharesAlwaysAddUpToCost(t *testing.T) {
	// Odd totals across three-way splits are the worst case for rounding.
	for price := int64(1); price <= 500; price++ {
		result, err := splitter.BreakdownCash(splitter.SplitParams[int]{
			GrossRevenue:  price,
			NetRevenue:    price,
			PaymentVendor: splitter.VendorCash,
			Splits: []*splitter.ItemSplit[int]{
				{
					ItemID:   1,
					Price:    price,
					Quantity: 1,
					SplitRules: []*splitter.SplitRule{
						{DepartmentID: uuid.New(), Amount: 33.33},
						{DepartmentID: uuid.New(), Amount: 33.33},
						{DepartmentID: uuid.New(), Amount: 33.34},
					},
				},
			},
		})

		if err != nil {
			t.Fatalf("price %d: got: %s, wants: nil", price, err.Error())
		}

		gross, net := int64(0), int64(0)
		for _, s := range result[0].Shares {
			gross += s.GrossRevenue
			net += s.NetRevenue
		}

		if gross != price || net != price {
			t.Errorf("price %d: gross=%d net=%d, wants both %d", price, gross, net, price)
		}
	}
}

func TestCashErrorsAreComparable(t *testing.T) {
	_, err := splitter.BreakdownCash(splitter.SplitParams[int]{
		GrossRevenue: 100,
		NetRevenue:   100,
		Splits: []*splitter.ItemSplit[int]{
			{
				ItemID:     1,
				Price:      100,
				Quantity:   1,
				SplitRules: []*splitter.SplitRule{{DepartmentID: uuid.New(), Amount: 50}},
			},
		},
	})

	if !errors.Is(err, splitter.ErrSharePercentNot100) {
		t.Errorf("got: %v, wants: ErrSharePercentNot100", err)
	}
}

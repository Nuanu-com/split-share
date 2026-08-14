package splitter_test

import (
	"testing"

	"github.com/Nuanu-com/split-share/splitter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMDRPart(t *testing.T) {
	assert.Equal(t, splitter.MDRPart(100_000, 40_000, 4_000), int64(1_600))
	assert.Equal(t, splitter.MDRPart(100_000, 60_000, 4_000), int64(2_400))
}

func TestBreakdownXenditWithoutOwner(t *testing.T) {
	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount: 100_000,
		NetAmount:   86_000,
		MDRAmount:   4_000,
	})

	assert.Nil(t, res)
	assert.ErrorContains(t, err, "transaction owner need to be defined")
}

func TestXenditBreakdownSuccess(t *testing.T) {
	department1 := uuid.New()
	department2 := uuid.New()
	department3 := uuid.New()

	item1 := 1
	item2 := 2

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      100_000,
		NetAmount:        86_000,
		MDRAmount:        4_000,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{
				ItemID:   item1,
				Price:    30000,
				Quantity: 2,
				SplitRules: []*splitter.SplitRule{
					{
						Amount:       20,
						DepartmentID: department1,
					},
					{
						Amount:       80,
						DepartmentID: department2,
					},
				},
			},
			{
				ItemID:   item2,
				Price:    40000,
				Quantity: 1,
				SplitRules: []*splitter.SplitRule{
					{
						Amount:       100,
						DepartmentID: department3,
					},
				},
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)

	assert.Equal(t, res[0].Cost, int64(60_000))
	assert.Equal(t, res[0].NetCost, int64(60_000-2_400))
	assert.Equal(t, res[0].ItemID, item1)
	assert.Len(t, res[0].Shares, 2)

	assert.Equal(t, res[0].Shares[0].DepartmentID, department1)
	assert.Equal(t, res[0].Shares[0].GrossRevenue, int64(12_000))
	assert.Equal(t, res[0].Shares[0].NetRevenue, int64(9_600))

	assert.Equal(t, res[0].Shares[1].DepartmentID, department2)
	assert.Equal(t, res[0].Shares[1].GrossRevenue, int64(48_000))
	assert.Equal(t, res[0].Shares[1].NetRevenue, int64(48_000))

	assert.Equal(t, res[1].Cost, int64(40_000))
	assert.Equal(t, res[1].NetCost, int64(40_000-1_600))
	assert.Equal(t, res[1].ItemID, item2)
	assert.Len(t, res[1].Shares, 2)

	assert.Equal(t, res[1].Shares[0].DepartmentID, department3)
	assert.Equal(t, res[1].Shares[0].GrossRevenue, int64(40_000))
	assert.Equal(t, res[1].Shares[0].NetRevenue, int64(40_000))

	assert.Equal(t, res[1].Shares[1].DepartmentID, department1)
	assert.Equal(t, res[1].Shares[1].GrossRevenue, int64(0))
	assert.Equal(t, res[1].Shares[1].NetRevenue, int64(-1_600))
}

func TestMDRPartZeroTransactionTotal(t *testing.T) {
	// Without a guard this divides by zero and overflows to math.MaxInt64.
	assert.Equal(t, int64(0), splitter.MDRPart(0, 10_000, 4_000))
}

func TestXenditRejectsZeroGrossAmount(t *testing.T) {
	department1 := uuid.New()

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      0,
		MDRAmount:        4_000,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{ItemID: 1, Price: 10_000, Quantity: 1, SplitRules: []*splitter.SplitRule{
				{DepartmentID: department1, Amount: 100},
			}},
		},
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, splitter.ErrInvalidGrossAmount)
}

// The owner can hold more than one rule on an item. The MDR is a per-item cost
// and must be charged exactly once, not once per matching rule.
func TestXenditOwnerWithMultipleRulesChargedMDROnce(t *testing.T) {
	department1 := uuid.New()
	department2 := uuid.New()

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      100_000,
		MDRAmount:        4_000,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{ItemID: 1, Price: 100_000, Quantity: 1, SplitRules: []*splitter.SplitRule{
				{DepartmentID: department1, Amount: 30},
				{DepartmentID: department1, Amount: 20},
				{DepartmentID: department2, Amount: 50},
			}},
		},
	})

	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, int64(96_000), res[0].NetCost)

	totalNet := int64(0)
	for _, s := range res[0].Shares {
		totalNet += s.NetRevenue
	}

	assert.Equal(t, res[0].NetCost, totalNet, "shares must add up to the item net cost")
}

func TestXenditSharesAddUpDespiteRounding(t *testing.T) {
	department1 := uuid.New()
	department2 := uuid.New()

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      75,
		MDRAmount:        0,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{ItemID: 1, Price: 75, Quantity: 1, SplitRules: []*splitter.SplitRule{
				{DepartmentID: department1, Amount: 50},
				{DepartmentID: department2, Amount: 50},
			}},
		},
	})

	assert.Nil(t, err)

	totalGross := int64(0)
	for _, s := range res[0].Shares {
		totalGross += s.GrossRevenue
	}

	assert.Equal(t, int64(75), totalGross, "50/50 of 75 must not invent a unit")
}

// Prorating the MDR per item rounds each part independently, so the parts have
// to be reconciled or the platform over/under-charges the fee.
func TestXenditMDRPartsSumToMDRAmount(t *testing.T) {
	department1 := uuid.New()

	splits := make([]*splitter.ItemSplit[int], 0, 3)
	for i := range 3 {
		splits = append(splits, &splitter.ItemSplit[int]{
			ItemID: i, Price: 10_000, Quantity: 1,
			SplitRules: []*splitter.SplitRule{{DepartmentID: department1, Amount: 100}},
		})
	}

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      30_000,
		MDRAmount:        1_000,
		TransactionOwner: department1,
		Splits:           splits,
	})

	assert.Nil(t, err)

	totalNet := int64(0)
	for _, r := range res {
		totalNet += r.NetCost
	}

	assert.Equal(t, int64(29_000), totalNet, "gross minus MDR")
}

func TestXenditRejectsIncompleteShare(t *testing.T) {
	department1 := uuid.New()

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      10_000,
		MDRAmount:        400,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{ItemID: 1, Price: 10_000, Quantity: 1, SplitRules: nil},
		},
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, splitter.ErrSharePercentNot100)
}

func TestXenditRejectsTotalMismatch(t *testing.T) {
	department1 := uuid.New()

	res, err := splitter.BreakdownXendit(splitter.XenditParams[int]{
		GrossAmount:      999_999,
		MDRAmount:        400,
		TransactionOwner: department1,
		Splits: []*splitter.ItemSplit[int]{
			{ItemID: 1, Price: 10_000, Quantity: 1, SplitRules: []*splitter.SplitRule{
				{DepartmentID: department1, Amount: 100},
			}},
		},
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, splitter.ErrTotalMismatch)
}

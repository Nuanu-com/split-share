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

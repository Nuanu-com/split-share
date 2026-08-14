package splitter_test

import (
	"testing"

	"github.com/Nuanu-com/split-share/splitter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBreakdownMandiriSuccess(t *testing.T) {
	department1 := uuid.New()
	department2 := uuid.New()
	department3 := uuid.New()

	item1 := 1
	item2 := 2
	item3 := 3

	result, err := splitter.BreakdownMandiri(splitter.MandiriParams[int]{
		GrossAmount: int64(530_000),
		NetAmount:   int64(523_110),
		MDR:         1.3,
		MDRAmount:   6_890,
		DCC:         0,
		DCCAmount:   0,
		Splits: []*splitter.ItemSplit[int]{
			{
				ItemID:   item1,
				Price:    75_000,
				Quantity: 6,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       20,
					},
					{
						DepartmentID: department2,
						Amount:       80,
					},
				},
			},
			{
				ItemID:   item2,
				Price:    20_000,
				Quantity: 4,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       30,
					},
					{
						DepartmentID: department3,
						Amount:       70,
					},
				},
			},
			{
				ItemID:   item3,
				Price:    0,
				Quantity: 3,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       100,
					},
				},
			},
		},
	})

	assert.Nil(t, err, "should not occur")
	assert.Len(t, result, 3)

	assert.Equal(t, result[0].Cost, int64(450_000))
	assert.Equal(t, result[0].NetCost, int64(444_150))
	assert.Equal(t, result[0].ItemID, item1)
	assert.Len(t, result[0].Shares, 2)

	assert.Equal(t, result[0].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[0].Shares[0].GrossRevenue, int64(90_000))
	assert.Equal(t, result[0].Shares[0].NetRevenue, int64(88_830))

	assert.Equal(t, result[0].Shares[1].DepartmentID, department2)
	assert.Equal(t, result[0].Shares[1].GrossRevenue, int64(360_000))
	assert.Equal(t, result[0].Shares[1].NetRevenue, int64(355_320))

	assert.Equal(t, result[1].Cost, int64(80_000))
	assert.Equal(t, result[1].NetCost, int64(78_960))
	assert.Equal(t, result[1].ItemID, item2)
	assert.Len(t, result[1].Shares, 2)

	assert.Equal(t, result[1].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[1].Shares[0].GrossRevenue, int64(24_000))
	assert.Equal(t, result[1].Shares[0].NetRevenue, int64(23_688))

	assert.Equal(t, result[1].Shares[1].DepartmentID, department3)
	assert.Equal(t, result[1].Shares[1].GrossRevenue, int64(56_000))
	assert.Equal(t, result[1].Shares[1].NetRevenue, int64(55_272))

	assert.Equal(t, result[2].Cost, int64(0))
	assert.Equal(t, result[2].NetCost, int64(0))
	assert.Equal(t, result[2].ItemID, item3)
	assert.Len(t, result[2].Shares, 1)

	assert.Equal(t, result[2].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[2].Shares[0].GrossRevenue, int64(0))
	assert.Equal(t, result[2].Shares[0].NetRevenue, int64(0))
}

func TestMandiriSuccessWithDCC(t *testing.T) {
	department1 := uuid.New()
	department2 := uuid.New()
	department3 := uuid.New()

	item1 := 1
	item2 := 2
	item3 := 3

	result, err := splitter.BreakdownMandiri(splitter.MandiriParams[int]{
		GrossAmount: int64(530_000),
		NetAmount:   int64(523_110),
		MDR:         1.3,
		MDRAmount:   6_890,
		DCC:         0.5,
		DCCAmount:   2_650,
		Splits: []*splitter.ItemSplit[int]{
			{
				ItemID:   item1,
				Price:    75000,
				Quantity: 6,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       20,
					},
					{
						DepartmentID: department2,
						Amount:       80,
					},
				},
			},
			{
				ItemID:   item2,
				Price:    20000,
				Quantity: 4,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       30,
					},
					{
						DepartmentID: department3,
						Amount:       70,
					},
				},
			},
			{
				ItemID:   item3,
				Price:    0,
				Quantity: 3,
				SplitRules: []*splitter.SplitRule{
					{
						DepartmentID: department1,
						Amount:       100,
					},
				},
			},
		},
	})

	assert.Nil(t, err, "should not occur")
	assert.Len(t, result, 3)

	assert.Equal(t, result[0].Cost, int64(450_000))
	assert.Equal(t, result[0].NetCost, int64(444_150+2_250))
	assert.Len(t, result[0].Shares, 2)

	assert.Equal(t, result[0].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[0].Shares[0].GrossRevenue, int64(90_000))
	assert.Equal(t, result[0].Shares[0].NetRevenue, int64(88_830+450))

	assert.Equal(t, result[0].Shares[1].DepartmentID, department2)
	assert.Equal(t, result[0].Shares[1].GrossRevenue, int64(360_000))
	assert.Equal(t, result[0].Shares[1].NetRevenue, int64(355_320+1_800))

	assert.Equal(t, result[1].Cost, int64(80_000))
	assert.Equal(t, result[1].NetCost, int64(78_960+400))
	assert.Len(t, result[1].Shares, 2)

	assert.Equal(t, result[1].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[1].Shares[0].GrossRevenue, int64(24_000))
	assert.Equal(t, result[1].Shares[0].NetRevenue, int64(23_688+120))

	assert.Equal(t, result[1].Shares[1].DepartmentID, department3)
	assert.Equal(t, result[1].Shares[1].GrossRevenue, int64(56_000))
	assert.Equal(t, result[1].Shares[1].NetRevenue, int64(55_272+280))

	assert.Equal(t, result[2].Cost, int64(0))
	assert.Equal(t, result[2].NetCost, int64(0))
	assert.Len(t, result[2].Shares, 1)

	assert.Equal(t, result[2].Shares[0].DepartmentID, department1)
	assert.Equal(t, result[2].Shares[0].GrossRevenue, int64(0))
	assert.Equal(t, result[2].Shares[0].NetRevenue, int64(0))
}

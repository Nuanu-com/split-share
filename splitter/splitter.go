package splitter

import "github.com/google/uuid"

type SplitRule struct {
	// Percent Amount of the split
	Amount       float64   `json:"amount"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type ItemSplit[T any] struct {
	ItemID     T            `json:"item_id"`
	Price      int64        `json:"price"`
	Quantity   int64        `json:"quantity"`
	SplitRules []*SplitRule `json:"split_rules"`
}

type Share struct {
	DepartmentID uuid.UUID `json:"department_id"`
	GrossRevenue int64     `json:"gross_revenue"`
	NetRevenue   int64     `json:"net_revenue"`
}

type SplitResult[T any] struct {
	ItemID  T
	Cost    int64
	NetCost int64
	Shares  []*Share
}

type PaymentVendor string

const (
	VendorCash    PaymentVendor = "CASH"
	VendorXendit  PaymentVendor = "XENDIT"
	VendorMandiri PaymentVendor = "MANDIRI"
)

type SplitParams[ItemT any] struct {
	GrossRevenue  int64               `json:"gross_revenue"`
	NetRevenue    int64               `json:"net_revenue"`
	PaymentVendor PaymentVendor       `json:"payment_vendor"`
	Splits        []*ItemSplit[ItemT] `json:"splits"`
}

type Splitter[ItemT any] interface {
	Breakdown(params SplitParams[ItemT]) ([]*SplitResult[ItemT], error)
}

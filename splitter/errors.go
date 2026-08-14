package splitter

import "errors"

var (
	// ErrSharePercentNot100 is returned when an item's split rules do not add up
	// to exactly 100%, which would leave revenue unallocated or over-allocated.
	ErrSharePercentNot100 = errors.New("total share must be 100%")

	// ErrTotalMismatch is returned when the sum of the items does not match the
	// declared transaction total.
	ErrTotalMismatch = errors.New("items total must equal to purchase total")

	// ErrNoTransactionOwner is returned when a vendor that charges the fee to a
	// single department is called without that department.
	ErrNoTransactionOwner = errors.New("transaction owner need to be defined")

	// ErrInvalidGrossAmount is returned when a gross amount is needed to prorate
	// fees across items but is missing or non-positive.
	ErrInvalidGrossAmount = errors.New("gross amount must be greater than zero")
)

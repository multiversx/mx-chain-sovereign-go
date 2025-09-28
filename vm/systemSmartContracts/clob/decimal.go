package clob

import (
	"github.com/nikolaydubina/fpdecimal"
)

// Decimal is a wrapper around fpdecimal.Decimal to provide additional functionality.
type Decimal struct {
	fpdecimal.Decimal
}

// NewDecimal creates a new Decimal from a string.
func NewDecimal(s string) (Decimal, error) {
	d, err := fpdecimal.NewFromString(s)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{d}, nil
}

// MustNewDecimal creates a new Decimal from a string, panicking on error.
func MustNewDecimal(s string) Decimal {
	d, err := NewDecimal(s)
	if err != nil {
		panic(err)
	}
	return d
}
package utils

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func ParseDecimal(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("failed to parse decimal string '%s': %w", s, err)
	}
	return d, nil
}

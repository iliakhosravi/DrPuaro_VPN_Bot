package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/shopspring/decimal"
)

// Decimal is a custom type to handle shopspring/decimal in GORM
type Decimal struct {
	decimal.Decimal
}

// Scan implements the Scanner interface.
func (d *Decimal) Scan(value interface{}) error {
	if value == nil {
		d.Decimal = decimal.NewFromInt(0) // Or handle nil as you see fit
		return nil
	}

	switch v := value.(type) {
	case []byte:
		dec, err := decimal.NewFromString(string(v))
		if err != nil {
			return err
		}
		d.Decimal = dec
	case string:
		dec, err := decimal.NewFromString(v)
		if err != nil {
			return err
		}
		d.Decimal = dec
	case float64: // Important for some drivers like SQLite
		d.Decimal = decimal.NewFromFloat(v)
	case int64: // For integer types from the database
		d.Decimal = decimal.NewFromInt(v)
	default:
		return fmt.Errorf("unsupported type %T for Decimal Scan", value)
	}
	return nil
}

// Value implements the Valuer interface.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

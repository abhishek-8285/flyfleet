package models

import (
	"time"

	"github.com/google/uuid"
)

type DriverFuelLimit struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	OrganizationID     uuid.UUID `json:"organization_id" db:"organization_id"`
	DriverID           uuid.UUID `json:"driver_id" db:"driver_id"`
	DailyLitersLimit   *float64  `json:"daily_liters_limit,omitempty" db:"daily_liters_limit"`
	MonthlyLitersLimit *float64  `json:"monthly_liters_limit,omitempty" db:"monthly_liters_limit"`
	DailyAmountLimit   *float64  `json:"daily_amount_limit,omitempty" db:"daily_amount_limit"`
	MonthlyAmountLimit *float64  `json:"monthly_amount_limit,omitempty" db:"monthly_amount_limit"`
	Currency           string    `json:"currency" db:"currency"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type OrgFuelSettings struct {
	OrganizationID            uuid.UUID `json:"organization_id" db:"organization_id"`
	DefaultDailyLitersLimit   *float64  `json:"default_daily_liters_limit,omitempty" db:"default_daily_liters_limit"`
	DefaultMonthlyLitersLimit *float64  `json:"default_monthly_liters_limit,omitempty" db:"default_monthly_liters_limit"`
	DefaultDailyAmountLimit   *float64  `json:"default_daily_amount_limit,omitempty" db:"default_daily_amount_limit"`
	DefaultMonthlyAmountLimit *float64  `json:"default_monthly_amount_limit,omitempty" db:"default_monthly_amount_limit"`
	Currency                  string    `json:"currency" db:"currency"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

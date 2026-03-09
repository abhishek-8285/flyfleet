package models

import (
	"time"

	"github.com/google/uuid"
)

// Discount types
const (
	CouponTypePercentage  = "PERCENTAGE"
	CouponTypeFixedAmount = "FIXED_AMOUNT"
)

// BillingCycle represents a distinct time period over which usage is aggregated and billed.
type BillingCycle struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrgID          uuid.UUID  `json:"org_id" db:"org_id"`
	SubscriptionID uuid.UUID  `json:"subscription_id" db:"subscription_id"`
	CycleStart     time.Time  `json:"cycle_start" db:"cycle_start"`
	CycleEnd       time.Time  `json:"cycle_end" db:"cycle_end"`
	Status         string     `json:"status" db:"status"` // ACTIVE, CLOSED, CANCELLED
	ClosedAt       *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	InvoiceID      *uuid.UUID `json:"invoice_id,omitempty" db:"invoice_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// Coupon represents an administrative created logic to provide discounts.
type Coupon struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Code          string     `json:"code" db:"code"`
	Name          string     `json:"name" db:"name"`
	DiscountType  string     `json:"discount_type" db:"discount_type"`
	DiscountValue float64    `json:"discount_value" db:"discount_value"`
	Currency      *string    `json:"currency,omitempty" db:"currency"`
	MaxRedeems    *int       `json:"max_redemptions,omitempty" db:"max_redemptions"`
	RedeemedCount int        `json:"redeemed_count" db:"redeemed_count"`
	ValidFrom     *time.Time `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	Status        string     `json:"status" db:"status"` // ACTIVE, EXPIRED, DISABLED
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
}

// CouponRedemption represents the application of a coupon by an organization.
type CouponRedemption struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	CouponID       uuid.UUID  `json:"coupon_id" db:"coupon_id"`
	OrgID          uuid.UUID  `json:"org_id" db:"org_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty" db:"subscription_id"`
	InvoiceID      *uuid.UUID `json:"invoice_id,omitempty" db:"invoice_id"`
	RedeemedAt     time.Time  `json:"redeemed_at" db:"redeemed_at"`
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdvancedSubRepository struct {
	db *pgxpool.Pool
}

func NewAdvancedSubRepository(db *pgxpool.Pool) *AdvancedSubRepository {
	return &AdvancedSubRepository{db: db}
}

// CreateCycle starts a new billing cycle for a subscription
func (r *AdvancedSubRepository) CreateCycle(ctx context.Context, cycle *models.BillingCycle) error {
	if cycle.ID == uuid.Nil {
		cycle.ID = uuid.New()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_cycles (id, tenant_id, subscription_id, cycle_start, cycle_end, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`, cycle.ID, cycle.OrgID, cycle.SubscriptionID, cycle.CycleStart, cycle.CycleEnd, cycle.Status)

	return err
}

// GetActiveCycle returns the currently active cycle for a subscription
func (r *AdvancedSubRepository) GetActiveCycle(ctx context.Context, subID uuid.UUID) (*models.BillingCycle, error) {
	var c models.BillingCycle
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, subscription_id, cycle_start, cycle_end, status, created_at
		FROM billing_cycles
		WHERE subscription_id = $1 AND status = 'ACTIVE'
		ORDER BY cycle_end DESC LIMIT 1
	`, subID).Scan(&c.ID, &c.OrgID, &c.SubscriptionID, &c.CycleStart, &c.CycleEnd, &c.Status, &c.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil // No active cycle
	} else if err != nil {
		return nil, err
	}

	return &c, nil
}

// CloseCycle archives a billing cycle and associates it with an invoice
func (r *AdvancedSubRepository) CloseCycle(ctx context.Context, cycleID uuid.UUID, invoiceID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE billing_cycles
		SET status = 'CLOSED', closed_at = CURRENT_TIMESTAMP, invoice_id = $1
		WHERE id = $2
	`, invoiceID, cycleID)

	return err
}

// GetCouponByCode fetches a discount coupon
func (r *AdvancedSubRepository) GetCouponByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var c models.Coupon
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name, discount_type, discount_value, currency, max_redemptions, redeemed_count, valid_from, valid_until, status, created_at
		FROM billing_coupons
		WHERE code = $1
	`, code).Scan(&c.ID, &c.Code, &c.Name, &c.DiscountType, &c.DiscountValue, &c.Currency, &c.MaxRedeems, &c.RedeemedCount, &c.ValidFrom, &c.ValidUntil, &c.Status, &c.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &c, nil
}

// RedeemCoupon creates a redemption record and increments the global redeemed_count atomically
func (r *AdvancedSubRepository) RedeemCoupon(ctx context.Context, cr *models.CouponRedemption) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check eligibility
	var maxRedeems *int
	var redeemedCount int
	var status string
	var validFrom, validUntil *time.Time

	err = tx.QueryRow(ctx, `
		SELECT max_redemptions, redeemed_count, status, valid_from, valid_until
		FROM billing_coupons WHERE id = $1 FOR UPDATE
	`, cr.CouponID).Scan(&maxRedeems, &redeemedCount, &status, &validFrom, &validUntil)
	if err != nil {
		return err
	}

	now := time.Now()
	if status != "ACTIVE" {
		return errors.New("coupon is not active")
	}
	if validFrom != nil && now.Before(*validFrom) {
		return errors.New("coupon is not yet valid")
	}
	if validUntil != nil && now.After(*validUntil) {
		return errors.New("coupon has expired")
	}
	if maxRedeems != nil && redeemedCount >= *maxRedeems {
		return errors.New("coupon redemption limit reached")
	}

	if cr.ID == uuid.Nil {
		cr.ID = uuid.New()
	}

	// Insert Redemption
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_coupon_redemptions (id, coupon_id, tenant_id, subscription_id, invoice_id)
		VALUES ($1, $2, $3, $4, $5)
	`, cr.ID, cr.CouponID, cr.OrgID, cr.SubscriptionID, cr.InvoiceID)
	if err != nil {
		return err
	}

	// Increment code counter
	_, err = tx.Exec(ctx, `
		UPDATE billing_coupons SET redeemed_count = redeemed_count + 1 WHERE id = $1
	`, cr.CouponID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

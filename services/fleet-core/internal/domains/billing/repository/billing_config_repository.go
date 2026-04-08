package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingConfigRepository handles billing configuration persistence
type BillingConfigRepository struct {
	db *pgxpool.Pool
}

// NewBillingConfigRepository creates a new repository
func NewBillingConfigRepository(db *pgxpool.Pool) *BillingConfigRepository {
	return &BillingConfigRepository{db: db}
}

// GetByOrgID retrieves billing config for an organization
func (r *BillingConfigRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*models.BillingConfiguration, error) {
	query := `
		SELECT id, org_id, company_name, company_address, company_gstin, company_pan,
		       company_email, company_phone, company_website, company_logo_url,
		       primary_color, secondary_color, theme,
		       bank_name, bank_account_name, bank_account_number, bank_ifsc, bank_branch, bank_upi,
		       invoice_prefix, invoice_terms, invoice_notes, show_hsn, show_qr_code, show_payment_info,
		       default_currency, default_tax_calculation, default_payment_terms_days,
		       is_active, created_at, created_by, updated_at, updated_by
		FROM billing_configurations
		WHERE org_id = $1 AND is_active = true`

	var config models.BillingConfiguration
	var termsJSON []byte

	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&config.ID, &config.OrgID, &config.CompanyName, &config.CompanyAddress, &config.CompanyGSTIN,
		&config.CompanyPAN, &config.CompanyEmail, &config.CompanyPhone, &config.CompanyWebsite,
		&config.CompanyLogoURL, &config.PrimaryColor, &config.SecondaryColor, &config.Theme,
		&config.BankName, &config.BankAccountName, &config.BankAccountNumber, &config.BankIFSC,
		&config.BankBranch, &config.BankUPI, &config.InvoicePrefix, &termsJSON, &config.InvoiceNotes,
		&config.ShowHSN, &config.ShowQRCode, &config.ShowPaymentInfo, &config.DefaultCurrency,
		&config.DefaultTaxCalculation, &config.DefaultPaymentTermsDays, &config.IsActive,
		&config.CreatedAt, &config.CreatedBy, &config.UpdatedAt, &config.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get billing config: %w", err)
	}

	// Parse terms JSON
	if len(termsJSON) > 0 {
		if err := json.Unmarshal(termsJSON, &config.InvoiceTerms); err != nil {
			config.InvoiceTerms = []string{}
		}
	}

	return &config, nil
}

// Create creates a new billing configuration
func (r *BillingConfigRepository) Create(ctx context.Context, config *models.BillingConfiguration) error {
	termsJSON, err := json.Marshal(config.InvoiceTerms)
	if err != nil {
		return fmt.Errorf("failed to marshal terms: %w", err)
	}

	query := `
		INSERT INTO billing_configurations (
			id, org_id, company_name, company_address, company_gstin, company_pan,
			company_email, company_phone, company_website, company_logo_url,
			primary_color, secondary_color, theme,
			bank_name, bank_account_name, bank_account_number, bank_ifsc, bank_branch, bank_upi,
			invoice_prefix, invoice_terms, invoice_notes, show_hsn, show_qr_code, show_payment_info,
			default_currency, default_tax_calculation, default_payment_terms_days,
			is_active, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31
		)
		ON CONFLICT (org_id) DO UPDATE SET
			company_name = EXCLUDED.company_name,
			company_address = EXCLUDED.company_address,
			company_gstin = EXCLUDED.company_gstin,
			company_pan = EXCLUDED.company_pan,
			company_email = EXCLUDED.company_email,
			company_phone = EXCLUDED.company_phone,
			company_website = EXCLUDED.company_website,
			company_logo_url = EXCLUDED.company_logo_url,
			primary_color = EXCLUDED.primary_color,
			secondary_color = EXCLUDED.secondary_color,
			theme = EXCLUDED.theme,
			bank_name = EXCLUDED.bank_name,
			bank_account_name = EXCLUDED.bank_account_name,
			bank_account_number = EXCLUDED.bank_account_number,
			bank_ifsc = EXCLUDED.bank_ifsc,
			bank_branch = EXCLUDED.bank_branch,
			bank_upi = EXCLUDED.bank_upi,
			invoice_prefix = EXCLUDED.invoice_prefix,
			invoice_terms = EXCLUDED.invoice_terms,
			invoice_notes = EXCLUDED.invoice_notes,
			show_hsn = EXCLUDED.show_hsn,
			show_qr_code = EXCLUDED.show_qr_code,
			show_payment_info = EXCLUDED.show_payment_info,
			default_currency = EXCLUDED.default_currency,
			default_tax_calculation = EXCLUDED.default_tax_calculation,
			default_payment_terms_days = EXCLUDED.default_payment_terms_days,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = EXCLUDED.updated_by`

	_, err = r.db.Exec(ctx, query,
		config.ID, config.OrgID, config.CompanyName, config.CompanyAddress, config.CompanyGSTIN,
		config.CompanyPAN, config.CompanyEmail, config.CompanyPhone, config.CompanyWebsite,
		config.CompanyLogoURL, config.PrimaryColor, config.SecondaryColor, config.Theme,
		config.BankName, config.BankAccountName, config.BankAccountNumber, config.BankIFSC,
		config.BankBranch, config.BankUPI, config.InvoicePrefix, termsJSON, config.InvoiceNotes,
		config.ShowHSN, config.ShowQRCode, config.ShowPaymentInfo, config.DefaultCurrency,
		config.DefaultTaxCalculation, config.DefaultPaymentTermsDays, config.IsActive,
		config.CreatedBy, config.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create/update billing config: %w", err)
	}

	return nil
}

// Update updates an existing billing configuration
func (r *BillingConfigRepository) Update(ctx context.Context, config *models.BillingConfiguration) error {
	return r.Create(ctx, config) // Uses upsert
}

// Delete soft-deletes a billing configuration
func (r *BillingConfigRepository) Delete(ctx context.Context, orgID uuid.UUID) error {
	query := `UPDATE billing_configurations SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE org_id = $1`
	_, err := r.db.Exec(ctx, query, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete billing config: %w", err)
	}
	return nil
}

// GetOrCreate gets existing config or creates default one
func (r *BillingConfigRepository) GetOrCreate(ctx context.Context, orgID uuid.UUID, createdBy uuid.UUID) (*models.BillingConfiguration, error) {
	// Try to get existing
	config, err := r.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// If exists, return it
	if config != nil {
		return config, nil
	}

	// Create default
	config = models.DefaultBillingConfig(orgID)
	config.CreatedBy = createdBy
	config.UpdatedBy = createdBy

	if err := r.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

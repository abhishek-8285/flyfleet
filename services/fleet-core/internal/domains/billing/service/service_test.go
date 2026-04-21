package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Mock Repository ───────────────────────────────────────────────────────

type mockBillingRepo struct {
	invoices       map[uuid.UUID]*models.Invoice
	settings       map[uuid.UUID]*models.BillingSettings
	metrics        *models.BillingMetrics
	listTotal      int
	listErr        error
	getErr         error
	createErr      error
	updateErr      error
	deleteErr      error
	recordPayErr   error
	settingsErr    error
	metricsErr     error
}

func newMockRepo() *mockBillingRepo {
	return &mockBillingRepo{
		invoices: make(map[uuid.UUID]*models.Invoice),
		settings: make(map[uuid.UUID]*models.BillingSettings),
	}
}

func (m *mockBillingRepo) ListInvoices(_ context.Context, tenantID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []models.Invoice
	for _, inv := range m.invoices {
		if inv.TenantID == tenantID {
			result = append(result, *inv)
		}
	}
	return result, m.listTotal, nil
}

func (m *mockBillingRepo) CreateInvoice(_ context.Context, inv *models.Invoice) error {
	if m.createErr != nil {
		return m.createErr
	}
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	m.invoices[inv.ID] = inv
	return nil
}

func (m *mockBillingRepo) GetInvoice(_ context.Context, id uuid.UUID) (*models.Invoice, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	inv, ok := m.invoices[id]
	if !ok {
		return nil, nil
	}
	return inv, nil
}

func (m *mockBillingRepo) UpdateInvoice(_ context.Context, inv *models.Invoice) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.invoices[inv.ID] = inv
	return nil
}

func (m *mockBillingRepo) DeleteInvoice(_ context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.invoices, id)
	return nil
}

func (m *mockBillingRepo) GetSettings(_ context.Context, tenantID uuid.UUID) (*models.BillingSettings, error) {
	if m.settingsErr != nil {
		return nil, m.settingsErr
	}
	s, ok := m.settings[tenantID]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockBillingRepo) UpdateSettings(_ context.Context, s *models.BillingSettings) error {
	m.settings[s.TenantID] = s
	return nil
}

func (m *mockBillingRepo) GetMetrics(_ context.Context, _ uuid.UUID) (*models.BillingMetrics, error) {
	if m.metricsErr != nil {
		return nil, m.metricsErr
	}
	return m.metrics, nil
}

func (m *mockBillingRepo) RecordPayment(_ context.Context, id uuid.UUID, amount float64) error {
	if m.recordPayErr != nil {
		return m.recordPayErr
	}
	inv, ok := m.invoices[id]
	if !ok {
		return ErrInvoiceNotFound
	}
	inv.AmountPaid += amount
	inv.BalanceDue = inv.TotalAmount - inv.AmountPaid
	// Don't overwrite status — the service handles that
	if inv.BalanceDue < 0 {
		inv.BalanceDue = 0
	}
	return nil
}

func (m *mockBillingRepo) CreateTaxLines(_ context.Context, _ []models.InvoiceTaxLine) error {
	return nil
}

// ─── Test Helpers ───────────────────────────────────────────────────────────

func newTestService(repo *mockBillingRepo) *BillingService {
	return NewBillingService(
		repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		zap.NewNop(),
	)
}

func mustParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func newTestInvoice(tenantID uuid.UUID, status string) *models.Invoice {
	return &models.Invoice{
		ID:            uuid.New(),
		TenantID:         tenantID,
		CustomerName:  "Test Customer",
		InvoiceDate:   mustParseDate("2026-04-01"),
		DueDate:       mustParseDate("2026-05-01"),
		Status:        models.InvoiceStatus(status),
		Currency:      "INR",
		TotalAmount:   10000.0,
		AmountPaid:    0,
		BalanceDue:    10000.0,
		Subtotal:      10000.0,
		Items: []models.InvoiceItem{
			{ID: uuid.New(), Description: "Service", Quantity: 1, UnitPrice: 10000.0, TaxPercent: 0, Total: 10000.0},
		},
	}
}

// ─── TestCreateInvoice ──────────────────────────────────────────────────────

func TestCreateInvoice(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name    string
		invoice *models.Invoice
		wantErr error
	}{
		{
			name: "valid invoice with items",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items: []models.InvoiceItem{
					{Description: "Fleet Management", Quantity: 1, UnitPrice: 5000.0, TaxPercent: 18, Total: 0},
					{Description: "GPS Tracking", Quantity: 10, UnitPrice: 500.0, TaxPercent: 18, Total: 0},
				},
			},
			wantErr: nil,
		},
		{
			name: "due date before invoice date",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-03-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: ErrDueDateBeforeInvoiceDate,
		},
		{
			name: "due date equals invoice date",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-04-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: ErrDueDateBeforeInvoiceDate,
		},
		{
			name: "item with zero quantity",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 0, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: errors.New("item quantity must be positive"),
		},
		{
			name: "item with negative unit price",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: -50.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: errors.New("item unit price cannot be negative"),
		},
		{
			name: "invalid status",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Status:        "INVALID_STATUS",
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: ErrInvalidInvoiceStatus,
		},
		{
			name: "negative discount value",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountPercentage),
				DiscountValue: -10,
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: errors.New("discount value cannot be negative"),
		},
		{
			name: "percentage discount exceeds 100",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountPercentage),
				DiscountValue: 150,
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: errors.New("percentage discount cannot exceed 100%"),
		},
		{
			name: "invalid discount type",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  "bogus",
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: errors.New("invalid discount type"),
		},
		{
			name: "default status is PENDING when empty",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: nil,
		},
		{
			name: "valid DRAFT status",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Status:        "DRAFT",
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: nil,
		},
		{
			name: "valid PAID status",
			invoice: &models.Invoice{
				TenantID:         tenantID,
				CustomerName:  "Acme Corp",
				InvoiceDate:   mustParseDate("2026-04-01"),
				DueDate:       mustParseDate("2026-05-01"),
				Status:        "PAID",
				Currency:      "INR",
				DiscountType:  string(models.DiscountNone),
				Items:         []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			svc := newTestService(repo)
			err := svc.CreateInvoice(context.Background(), tt.invoice)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("CreateInvoice() unexpected error = %v", err)
				}
				// Verify invoice was stored
				if len(repo.invoices) == 0 {
					t.Error("CreateInvoice() did not store invoice")
				}
			} else {
				if err == nil {
					t.Errorf("CreateInvoice() expected error, got nil")
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("CreateInvoice() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

// ─── TestRecordPayment ──────────────────────────────────────────────────────

func TestRecordPayment(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name         string
		setupInvoice func() *models.Invoice
		amount       float64
		wantErr      error
		wantStatus   models.InvoiceStatus
	}{
		{
			name: "full payment on pending invoice",
			setupInvoice: func() *models.Invoice {
				return newTestInvoice(tenantID, "PENDING")
			},
			amount:     10000.0,
			wantErr:    nil,
			wantStatus: models.StatusPaid,
		},
		{
			name: "partial payment on pending invoice",
			setupInvoice: func() *models.Invoice {
				return newTestInvoice(tenantID, "PENDING")
			},
			amount:     5000.0,
			wantErr:    nil,
			wantStatus: models.StatusPartial,
		},
		{
			name: "complete partial payment",
			setupInvoice: func() *models.Invoice {
				inv := newTestInvoice(tenantID, "PARTIAL")
				inv.AmountPaid = 5000.0
				inv.BalanceDue = 5000.0
				return inv
			},
			amount:     5000.0,
			wantErr:    nil,
			wantStatus: models.StatusPaid,
		},
		{
			name: "payment on already paid invoice",
			setupInvoice: func() *models.Invoice {
				inv := newTestInvoice(tenantID, "PAID")
				inv.AmountPaid = 10000.0
				inv.BalanceDue = 0
				return inv
			},
			amount:  1000.0,
			wantErr: ErrInvoiceAlreadyPaid,
		},
		{
			name: "payment exceeds outstanding balance",
			setupInvoice: func() *models.Invoice {
				return newTestInvoice(tenantID, "PENDING")
			},
			amount:  15000.0,
			wantErr: ErrPaymentExceedsOutstanding,
		},
		{
			name: "zero payment amount",
			setupInvoice: func() *models.Invoice {
				return newTestInvoice(tenantID, "PENDING")
			},
			amount:  0,
			wantErr: ErrNegativeAmount,
		},
		{
			name: "negative payment amount",
			setupInvoice: func() *models.Invoice {
				return newTestInvoice(tenantID, "PENDING")
			},
			amount:  -500.0,
			wantErr: ErrNegativeAmount,
		},
		{
			name: "payment for non-existent invoice",
			setupInvoice: func() *models.Invoice {
				return nil // no invoice seeded
			},
			amount:  1000.0,
			wantErr: ErrInvoiceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			inv := tt.setupInvoice()
			var invoiceID uuid.UUID
			if inv != nil {
				invoiceID = inv.ID
				repo.invoices[invoiceID] = inv
			} else {
				invoiceID = uuid.New()
			}
			svc := newTestService(repo)

			err := svc.RecordPayment(context.Background(), invoiceID, tt.amount)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("RecordPayment() unexpected error = %v", err)
					return
				}
				storedInv := repo.invoices[invoiceID]
				if storedInv.Status != tt.wantStatus {
					t.Errorf("RecordPayment() status = %v, want %v", storedInv.Status, tt.wantStatus)
				}
			} else {
				if err == nil {
					t.Errorf("RecordPayment() expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("RecordPayment() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

// ─── TestUpdateInvoice ──────────────────────────────────────────────────────

func TestUpdateInvoice(t *testing.T) {
	tenantID := uuid.New()

	draftInv := newTestInvoice(tenantID, "DRAFT")
	draftInv.Status = "draft" // IsInvoiceEditable checks lowercase
	issuedInv := newTestInvoice(tenantID, "issued")
	paidInv := newTestInvoice(tenantID, "paid")

	tests := []struct {
		name    string
		update  *models.Invoice
		preseed map[uuid.UUID]*models.Invoice
		wantErr error
	}{
		{
			name: "update draft invoice",
			update: &models.Invoice{
				ID:           draftInv.ID,
				CustomerName: "Updated Name",
				Status:       "draft", // keep same status
				Items:        []models.InvoiceItem{{Description: "New Service", Quantity: 2, UnitPrice: 5000.0, TaxPercent: 18, Total: 0}},
			},
			preseed: map[uuid.UUID]*models.Invoice{draftInv.ID: draftInv},
			wantErr: nil,
		},
		{
			name: "update paid invoice — immutable",
			update: &models.Invoice{
				ID:           paidInv.ID,
				CustomerName: "Updated Name",
				Items:        []models.InvoiceItem{{Description: "New Service", Quantity: 1, UnitPrice: 10000.0, TaxPercent: 0, Total: 0}},
			},
			preseed: map[uuid.UUID]*models.Invoice{paidInv.ID: paidInv},
			wantErr: ErrInvoiceImmutable,
		},
		{
			name: "update issued invoice — immutable",
			update: &models.Invoice{
				ID:           issuedInv.ID,
				CustomerName: "Updated Name",
				Items:        []models.InvoiceItem{{Description: "New Service", Quantity: 1, UnitPrice: 10000.0, TaxPercent: 0, Total: 0}},
			},
			preseed: map[uuid.UUID]*models.Invoice{issuedInv.ID: issuedInv},
			wantErr: ErrInvoiceImmutable,
		},
		{
			name: "update non-existent invoice",
			update: &models.Invoice{
				ID:           uuid.New(),
				CustomerName: "Ghost",
				Items:        []models.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100.0, TaxPercent: 0, Total: 0}},
			},
			preseed: map[uuid.UUID]*models.Invoice{},
			wantErr: ErrInvoiceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			for id, inv := range tt.preseed {
				repo.invoices[id] = inv
			}
			svc := newTestService(repo)

			err := svc.UpdateInvoice(context.Background(), tt.update)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("UpdateInvoice() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("UpdateInvoice() expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("UpdateInvoice() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

// ─── TestGetSettings ────────────────────────────────────────────────────────

func TestGetSettings(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name           string
		preseed        *models.BillingSettings
		wantTaxRate    float64
		wantCurrency   string
		wantGraceDays  int
	}{
		{
			name:          "returns defaults when no settings exist",
			preseed:       nil,
			wantTaxRate:   18.0,
			wantCurrency:  "USD",
			wantGraceDays: 14,
		},
		{
			name: "returns stored settings",
			preseed: &models.BillingSettings{
				TenantID:                tenantID,
				DefaultTaxRate:       28.0,
				BaseCurrency:         "INR",
				GracePeriodDays:      30,
				AutoInvoicingEnabled: true,
			},
			wantTaxRate:   28.0,
			wantCurrency:  "INR",
			wantGraceDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			if tt.preseed != nil {
				repo.settings[tenantID] = tt.preseed
			}
			svc := newTestService(repo)

			settings, err := svc.GetSettings(context.Background(), tenantID)
			if err != nil {
				t.Fatalf("GetSettings() error = %v", err)
			}
			if settings.DefaultTaxRate != tt.wantTaxRate {
				t.Errorf("DefaultTaxRate = %v, want %v", settings.DefaultTaxRate, tt.wantTaxRate)
			}
			if settings.BaseCurrency != tt.wantCurrency {
				t.Errorf("BaseCurrency = %v, want %v", settings.BaseCurrency, tt.wantCurrency)
			}
			if settings.GracePeriodDays != tt.wantGraceDays {
				t.Errorf("GracePeriodDays = %v, want %v", settings.GracePeriodDays, tt.wantGraceDays)
			}
		})
	}
}

// ─── TestIsValidStatusTransition ────────────────────────────────────────────

func TestIsValidStatusTransition(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Valid transitions
		{"draft → issued", "draft", "issued", true},
		{"draft → cancelled", "draft", "cancelled", true},
		{"issued → paid", "issued", "paid", true},
		{"issued → partial", "issued", "partial", true},
		{"issued → overdue", "issued", "overdue", true},
		{"issued → cancelled", "issued", "cancelled", true},
		{"partial → paid", "partial", "paid", true},
		{"partial → overdue", "partial", "overdue", true},
		{"partial → cancelled", "partial", "cancelled", true},
		{"overdue → paid", "overdue", "paid", true},
		{"overdue → partial", "overdue", "partial", true},
		{"overdue → cancelled", "overdue", "cancelled", true},

		// Invalid transitions
		{"paid → anything", "paid", "draft", false},
		{"cancelled → anything", "cancelled", "draft", false},
		{"draft → paid (skip)", "draft", "paid", false},
		{"issued → draft (rollback)", "issued", "draft", false},
		{"unknown → issued", "unknown", "issued", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidStatusTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("IsValidStatusTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

// ─── TestIsInvoiceEditable ──────────────────────────────────────────────────

func TestIsInvoiceEditable(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"draft is editable", "draft", true},
		{"issued not editable", "issued", false},
		{"paid not editable", "paid", false},
		{"partial not editable", "partial", false},
		{"overdue not editable", "overdue", false},
		{"cancelled not editable", "cancelled", false},
		{"empty not editable", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInvoiceEditable(tt.status)
			if got != tt.want {
				t.Errorf("IsInvoiceEditable(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// ─── TestIsInvoiceTerminal ──────────────────────────────────────────────────

func TestIsInvoiceTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"paid is terminal", "paid", true},
		{"cancelled is terminal", "cancelled", true},
		{"draft not terminal", "draft", false},
		{"issued not terminal", "issued", false},
		{"partial not terminal", "partial", false},
		{"overdue not terminal", "overdue", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInvoiceTerminal(tt.status)
			if got != tt.want {
				t.Errorf("IsInvoiceTerminal(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// ─── TestDeleteInvoice ──────────────────────────────────────────────────────

func TestDeleteInvoice(t *testing.T) {
	tenantID := uuid.New()
	inv := newTestInvoice(tenantID, "draft")

	repo := newMockRepo()
	repo.invoices[inv.ID] = inv
	svc := newTestService(repo)

	err := svc.DeleteInvoice(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("DeleteInvoice() error = %v", err)
	}
	if _, exists := repo.invoices[inv.ID]; exists {
		t.Error("DeleteInvoice() did not remove invoice")
	}
}

// ─── TestListInvoices ───────────────────────────────────────────────────────

func TestListInvoices(t *testing.T) {
	tenantID := uuid.New()
	repo := newMockRepo()
	repo.invoices[uuid.New()] = newTestInvoice(tenantID, "PENDING")
	repo.invoices[uuid.New()] = newTestInvoice(tenantID, "PAID")
	repo.listTotal = 2

	svc := newTestService(repo)

	invoices, total, err := svc.ListInvoices(context.Background(), tenantID, 1, 10)
	if err != nil {
		t.Fatalf("ListInvoices() error = %v", err)
	}
	if total != 2 {
		t.Errorf("ListInvoices() total = %d, want 2", total)
	}
	if len(invoices) != 2 {
		t.Errorf("ListInvoices() count = %d, want 2", len(invoices))
	}
}

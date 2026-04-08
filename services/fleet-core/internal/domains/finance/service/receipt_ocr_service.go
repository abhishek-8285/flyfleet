package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"go.uber.org/zap"
)

// OCRExtractResult holds the parsed fields from a receipt image.
type OCRExtractResult struct {
	VendorName  string   `json:"vendor_name"`
	VendorGSTIN string   `json:"vendor_gstin"`
	TotalAmount float64  `json:"total_amount"`
	GSTAmount   float64  `json:"gst_amount"`
	InvoiceNo   string   `json:"invoice_number"`
	Confidence  float64  `json:"confidence"`
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
}

// ReceiptOCRService processes receipt images via OCR.
// In production this would call Google Vision API, AWS Textract, or Azure Form Recognizer.
type ReceiptOCRService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewReceiptOCRService(repo *repository.FinanceRepository, logger *zap.Logger) *ReceiptOCRService {
	return &ReceiptOCRService{repo: repo, logger: logger}
}

// ExtractReceipt parses a receipt URL and returns extracted fields.
func (s *ReceiptOCRService) ExtractReceipt(ctx context.Context, receiptURL string) (*OCRExtractResult, error) {
	if receiptURL == "" {
		return nil, fmt.Errorf("receipt URL cannot be empty")
	}

	decodedURL, _ := url.QueryUnescape(receiptURL)
	upper := strings.ToUpper(decodedURL)

	result := &OCRExtractResult{Errors: []string{}}
	if idx := strings.Index(upper, "VENDOR="); idx >= 0 {
		segment := upper[idx+7:]
		if end := strings.Index(segment, "&"); end >= 0 {
			segment = segment[:end]
		}
		result.VendorName = strings.TrimSpace(strings.ReplaceAll(segment, "+", " "))
	}

	gstinRe := regexp.MustCompile(`[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][A-Z0-9]Z[A-Z0-9]`)
	if gstin := gstinRe.FindString(upper); gstin != "" {
		result.VendorGSTIN = gstin
	}

	amountRe := regexp.MustCompile(`(?i)(total|amount)=([0-9]+(?:\.[0-9]{1,2})?)`)
	if m := amountRe.FindStringSubmatch(decodedURL); len(m) == 3 {
		if amt, err := strconv.ParseFloat(m[2], 64); err == nil {
			result.TotalAmount = amt
		}
	}

	gstRe := regexp.MustCompile(`(?i)(gst|tax)=([0-9]+(?:\.[0-9]{1,2})?)`)
	if m := gstRe.FindStringSubmatch(decodedURL); len(m) == 3 {
		if gst, err := strconv.ParseFloat(m[2], 64); err == nil {
			result.GSTAmount = gst
		}
	}
	invoiceRe := regexp.MustCompile(`(?i)(invoice|inv|invoice_number)=([A-Za-z0-9\\-_/]+)`)
	if m := invoiceRe.FindStringSubmatch(decodedURL); len(m) == 3 {
		result.InvoiceNo = strings.TrimSpace(m[2])
	}

	score := 0.25
	if result.VendorName != "" {
		score += 0.25
	}
	if result.VendorGSTIN != "" {
		score += 0.25
	}
	if result.TotalAmount > 0 {
		score += 0.25
	}
	result.Confidence = score
	result.Valid = result.TotalAmount > 0
	if !result.Valid {
		result.Errors = append(result.Errors, "unable to determine receipt total")
	}

	s.logger.Info("OCR extraction completed", zap.String("url", receiptURL), zap.Float64("confidence", result.Confidence))
	return result, nil
}

// ValidateTotals cross-checks extracted totals against declared amounts.
func (s *ReceiptOCRService) ValidateTotals(extracted *OCRExtractResult, declared float64) bool {
	if extracted.TotalAmount <= 0 {
		return false
	}
	tolerance := 0.01 // 1% tolerance
	diff := extracted.TotalAmount - declared
	if diff < 0 {
		diff = -diff
	}
	return diff/declared <= tolerance
}

// ProcessExpenseItem runs OCR on an expense item's receipt and updates confidence.
func (s *ReceiptOCRService) ProcessExpenseItem(ctx context.Context, item *models.ExpenseItem) error {
	if item.ReceiptURL == nil || *item.ReceiptURL == "" {
		return nil
	}

	result, err := s.ExtractReceipt(ctx, *item.ReceiptURL)
	if err != nil {
		return err
	}

	_ = strings.ToUpper(result.VendorGSTIN) // normalize

	// Update item with OCR data
	item.OCRProcessed = true
	if result.Confidence > 0 {
		item.OCRConfidence = &result.Confidence
	}
	if result.VendorName != "" && item.VendorName == nil {
		item.VendorName = &result.VendorName
	}

	if result.VendorGSTIN != "" {
		item.VendorGSTIN = &result.VendorGSTIN
	}
	if result.GSTAmount > 0 {
		item.GSTAmount = result.GSTAmount
	}
	if result.InvoiceNo != "" {
		item.InvoiceNumber = &result.InvoiceNo
	}
	if err := s.repo.UpdateExpenseItemOCRData(ctx, item); err != nil {
		return err
	}
	return s.repo.MarkExpenseItemOCRProcessed(ctx, item.ID, result.Confidence)
}

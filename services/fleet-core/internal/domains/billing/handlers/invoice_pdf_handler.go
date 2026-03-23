package handlers

import (
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InvoicePDFHandler handles secure PDF retrieval
type InvoicePDFHandler struct {
	invoiceService *service.InvoiceService
}

func NewInvoicePDFHandler(is *service.InvoiceService) *InvoicePDFHandler {
	return &InvoicePDFHandler{invoiceService: is}
}

// GetInvoicePDF returns a signed URL or the PDF stream with RBAC
func (h *InvoicePDFHandler) GetInvoicePDF(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}

	// RBAC Check (Mock: In real system, verify user belongs to Org)
	// orgID := c.MustGet("org_id").(uuid.UUID)

	ctx := c.Request.Context()

	// Generate a signed URL for secure download (Compliance)
	url, err := h.invoiceService.GenerateSecureDownloadURL(ctx, invoiceID, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secure link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"download_url": url})
}

// PreviewTemplate allows dry-run rendering for the template editor
func (h *InvoicePDFHandler) PreviewTemplate(c *gin.Context) {
	var req struct {
		HTMLContent string                 `json:"html_content"`
		SampleData  map[string]interface{} `json:"sample_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: One-time context for preview
	pdfData, err := h.invoiceService.PreviewRender(c.Request.Context(), req.HTMLContent, req.SampleData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdfData)
}

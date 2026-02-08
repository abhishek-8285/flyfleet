package handlers

import (
	"net/http"
	"strconv"

	"github.com/flyfleet/billing-service/internal/models"
	"github.com/flyfleet/billing-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BillingHandler struct {
	service *service.BillingService
	logger  *zap.Logger
}

func NewBillingHandler(s *service.BillingService, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		service: s,
		logger:  logger,
	}
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	orgIDHeader := c.GetHeader("X-Org-ID")
	orgID, err := uuid.Parse(orgIDHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Organization ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	invoices, total, err := h.service.ListInvoices(c.Request.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  invoices,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	orgIDHeader := c.GetHeader("X-Org-ID")
	orgID, err := uuid.Parse(orgIDHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Organization ID"})
		return
	}

	var inv models.Invoice
	if err := c.ShouldBindJSON(&inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}
	inv.OrgID = orgID

	if err := h.service.CreateInvoice(c.Request.Context(), &inv); err != nil {
		h.logger.Error("Failed to create invoice", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}

	invoice, err := h.service.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if invoice == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice Not Found"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) UpdateInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}

	var inv models.Invoice
	if err := c.ShouldBindJSON(&inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}
	inv.ID = id

	if err := h.service.UpdateInvoice(c.Request.Context(), &inv); err != nil {
		h.logger.Error("Failed to update invoice", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *BillingHandler) DeleteInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}

	if err := h.service.DeleteInvoice(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete invoice", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BillingHandler) RecordPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}

	if err := h.service.RecordPayment(c.Request.Context(), id, req.Amount); err != nil {
		h.logger.Error("Failed to record payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment recorded successfully"})
}

func (h *BillingHandler) GetSettings(c *gin.Context) {
	orgIDHeader := c.GetHeader("X-Org-ID")
	orgID, err := uuid.Parse(orgIDHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Organization ID"})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *BillingHandler) UpdateSettings(c *gin.Context) {
	orgIDHeader := c.GetHeader("X-Org-ID")
	orgID, err := uuid.Parse(orgIDHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Organization ID"})
		return
	}

	var settings models.BillingSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}
	settings.OrgID = orgID

	if err := h.service.UpdateSettings(c.Request.Context(), &settings); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *BillingHandler) GetMetrics(c *gin.Context) {
	orgIDHeader := c.GetHeader("X-Org-ID")
	orgID, err := uuid.Parse(orgIDHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Organization ID"})
		return
	}

	metrics, err := h.service.GetMetrics(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

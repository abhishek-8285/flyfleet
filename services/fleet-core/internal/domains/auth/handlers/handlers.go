package handlers

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handlers struct {
	svc    *service.AuthService
	logger *zap.Logger
}

func NewHandlers(svc *service.AuthService, logger *zap.Logger) *Handlers {
	return &Handlers{
		svc:    svc,
		logger: logger,
	}
}

func (h *Handlers) SwitchContext(ctx context.Context, userID, targetTenantID uuid.UUID, ip, ua string) (*models.TenantSwitchResponse, error) {
	return h.svc.SwitchTenant(ctx, userID, targetTenantID, ip, ua)
}

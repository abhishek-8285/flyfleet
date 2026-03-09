package reports

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TenantQuota defines limits per tenant to prevent resource abuse.
type TenantQuota struct {
	MaxReportsPerHour   int // Max new reports a tenant can create within an hour
	MaxConcurrentJobs   int // Max simultaneously RUNNING report jobs per tenant
}

// DefaultQuota applies to all tenants unless overridden.
var DefaultQuota = TenantQuota{
	MaxReportsPerHour: 20,
	MaxConcurrentJobs: 5,
}

// QuotaEnforcer checks tenant report limits before allowing creation.
type QuotaEnforcer struct {
	repo   Repository
	logger *zap.Logger
}

func NewQuotaEnforcer(repo Repository, logger *zap.Logger) *QuotaEnforcer {
	return &QuotaEnforcer{repo: repo, logger: logger}
}

// CheckAndEnforce validates whether a tenant is within quota limits.
// Returns an error if the tenant has exceeded either limit.
func (q *QuotaEnforcer) CheckAndEnforce(ctx context.Context, tenantID string, quota TenantQuota) error {
	hourlyCount, err := q.repo.CountTenantReportsInWindow(ctx, tenantID, time.Now().Add(-1*time.Hour))
	if err != nil {
		q.logger.Warn("Failed to check hourly quota, allowing request", zap.Error(err))
		return nil // Fail open for availability
	}
	if hourlyCount >= hourlyCount && hourlyCount >= int64(quota.MaxReportsPerHour) {
		return fmt.Errorf("tenant quota exceeded: max %d reports per hour (current: %d)", quota.MaxReportsPerHour, hourlyCount)
	}

	concurrentCount, err := q.repo.CountTenantRunningJobs(ctx, tenantID)
	if err != nil {
		q.logger.Warn("Failed to check concurrent quota, allowing request", zap.Error(err))
		return nil
	}
	if concurrentCount >= int64(quota.MaxConcurrentJobs) {
		return fmt.Errorf("tenant quota exceeded: max %d concurrent reports (current: %d)", quota.MaxConcurrentJobs, concurrentCount)
	}

	return nil
}

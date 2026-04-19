package models

import (
	"time"

	"github.com/google/uuid"
)

// FinancialMetrics represents the core SaaS metrics for an organization or system-wide
type FinancialMetrics struct {
	OrgID         *uuid.UUID         `json:"tenant_id,omitempty"`
	Timestamp     time.Time          `json:"timestamp"`
	MRR           float64            `json:"mrr"`
	ARR           float64            `json:"arr"`
	NetRevenue    float64            `json:"net_revenue"`
	ChurnRate     float64            `json:"churn_rate"`
	LTV           float64            `json:"ltv"`
	ActiveSubs    int                `json:"active_subs"`
	RevenueByPlan map[string]float64 `json:"revenue_by_plan"`
}

// ChurnMetric tracks a single churn event or period
type ChurnMetric struct {
	OrgID     uuid.UUID `json:"tenant_id"`
	Reason    string    `json:"reason"`
	PlanID    uuid.UUID `json:"plan_id"`
	Value     float64   `json:"value"` // MRR lost
	Timestamp time.Time `json:"timestamp"`
}

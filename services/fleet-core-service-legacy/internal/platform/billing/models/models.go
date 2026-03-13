package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	MonthlyPrice float64                `json:"monthly_price"`
	MaxVehicles  int                    `json:"max_vehicles"`
	MaxDrivers   int                    `json:"max_drivers"`
	Features     map[string]interface{} `json:"features"`
}

type Subscription struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	PlanID         uuid.UUID `json:"plan_id"`
	Status         string    `json:"status"`
	StartDate      time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

type UsageRecord struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Metric         string    `json:"metric"`
	Value          int       `json:"value"`
	RecordedAt     time.Time `json:"recorded_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Domain       string    `json:"domain" db:"domain"`
	Status       string    `json:"status" db:"status"`
	Plan         string    `json:"plan" db:"plan"`
	ContactEmail string    `json:"contactEmail" db:"contact_email"`
	DefaultTheme string    `json:"defaultTheme" db:"default_theme"`
	RenewalDate  time.Time `json:"renewalDate" db:"renewal_date"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	CreatedBy    uuid.UUID `json:"createdBy" db:"created_by"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	UpdatedBy    uuid.UUID `json:"updatedBy" db:"updated_by"`
}

type TenantStats struct {
	TotalTenants  int `json:"total_tenants"`
	ActiveTenants int `json:"active_tenants"`
	PendingTenants int `json:"pending_tenants"`
	InactiveTenants int `json:"inactive_tenants"`
	TotalUsers    int `json:"total_users"`
	TotalVehicles int `json:"total_vehicles"`
	TotalDevices  int `json:"total_devices"`
	ActiveAlerts  int `json:"active_alerts"`
}

type DashboardStats struct {
	TotalVehicles      int     `json:"total_vehicles"`
	ActiveVehicles     int     `json:"active_vehicles"`
	TotalDrivers       int     `json:"total_drivers"`
	ActiveTrips        int     `json:"active_trips"`
	TotalDistanceToday float64 `json:"total_distance_today"`
	FuelConsumedToday  float64 `json:"fuel_consumed_today"`
	AlertsCount        int     `json:"alerts_count"`
}

type CreateTenantRequest struct {
	Name             string `json:"name" binding:"required"`
	Domain           string `json:"domain"`
	DefaultTheme     string `json:"default_theme"`
	ContactEmail     string `json:"contact_email"`
	SubscriptionTier string `json:"subscription_tier"`
}

type CreateTenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Branch struct {
	ID       uuid.UUID `json:"id"   db:"id"`
	Name     string    `json:"name" db:"name"`
	Code     string    `json:"code" db:"code"`
	Location string    `json:"location" db:"location"`
}

type CostCenter struct {
	ID          uuid.UUID `json:"id"   db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
}

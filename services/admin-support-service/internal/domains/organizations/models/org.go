package models

import (
	"time"
)

type Organization struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Domain      string    `json:"domain" db:"domain"`
	Status      string    `json:"status" db:"status"`
	Plan        string    `json:"plan" db:"plan"`
	RenewalDate time.Time `json:"renewalDate" db:"renewal_date"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type OrganizationStats struct {
	TotalOrgs      int `json:"total_orgs"`
	ActiveOrgs     int `json:"active_orgs"`
	PendingOrgs    int `json:"pending_orgs"`
	InactiveOrgs   int `json:"inactive_orgs"`
	TotalUsers     int `json:"total_users"`
	TotalVehicles  int `json:"total_vehicles"`
	TotalDevices   int `json:"total_devices"`
	ActiveAlerts   int `json:"active_alerts"`
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

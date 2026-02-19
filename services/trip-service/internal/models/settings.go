package models

import "time"

type TripSettings struct {
	AutoStartEnabled           bool    `json:"auto_start_enabled"`
	IdleThresholdMinutes       int     `json:"idle_threshold_minutes"`
	SpeedLimitKph              *int    `json:"speed_limit_kph,omitempty"`
	GeofenceAlertsEnabled      bool    `json:"geofence_alerts_enabled"`
	MaintenanceReminderDays    int     `json:"maintenance_reminder_days"`
	DriverBreakReminderMinutes int     `json:"driver_break_reminder_minutes"`
	FuelEfficiencyAlerts       bool    `json:"fuel_efficiency_alerts"`
	HardBrakingThreshold       float64 `json:"hard_braking_threshold"`
	RapidAccelerationThreshold float64 `json:"rapid_acceleration_threshold"`
	NightDrivingRestricted     bool    `json:"night_driving_restricted"`
	NightDrivingStart          string  `json:"night_driving_start"`
	NightDrivingEnd            string  `json:"night_driving_end"`
}

type SettingsMeta struct {
	Version      string     `json:"version"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	ModifiedBy   *string    `json:"modified_by,omitempty"`
}

type SettingsResponse struct {
	Data TripSettings `json:"data"`
	Meta SettingsMeta    `json:"meta"`
}

type UpdateSettingsRequest struct {
	Settings TripSettings `json:"settings"`
}

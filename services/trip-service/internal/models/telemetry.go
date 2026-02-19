package models

import "time"

type TelemetryPoint struct {
	Timestamp         time.Time `json:"timestamp" db:"time"`
	Latitude          float64   `json:"latitude" db:"latitude"`
	Longitude         float64   `json:"longitude" db:"longitude"`
	SpeedKph          float64   `json:"speed_kph" db:"speed_kph"`
	Heading           int16     `json:"heading" db:"heading"`
	OdometerKm        float64   `json:"odometer_km" db:"odometer_km"`
	FuelLevelPercent  int16     `json:"fuel_level_percent" db:"fuel_level_percent"`
	EngineTempCelsius int16     `json:"engine_temp_celsius" db:"engine_temp_celsius"`
	Alerts            []string  `json:"alerts" db:"alerts"`
}

type TelemetryMeta struct {
	TripID      string     `json:"trip_id"`
	VehicleID   string     `json:"vehicle_id"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	TotalPoints int        `json:"total_points"`
	Aggregation string     `json:"aggregation"`
}

type TelemetryResponse struct {
	Data []TelemetryPoint `json:"data"`
	Meta TelemetryMeta    `json:"meta"`
}

type TelemetryFilter struct {
	Aggregation string     `form:"aggregation" binding:"omitempty,oneof=raw hourly daily"`
	Limit       int        `form:"limit" binding:"omitempty,min=1,max=1000"`
	StartTime   *time.Time `form:"start_time" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime     *time.Time `form:"end_time" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
}

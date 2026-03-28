package dto

import (
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/models"
)

// DRTestSettingsDTO represents DR test configuration settings
// This DTO decouples the API from the database model
type DRTestSettingsDTO struct {
	ScheduleEnabled    bool     `json:"schedule_enabled"`
	Frequency          string   `json:"frequency"`
	ScheduleDay        string   `json:"schedule_day"`
	ScheduleTime       string   `json:"schedule_time"`
	Intensity          string   `json:"intensity"`
	NotificationEmails []string `json:"notification_emails"`
}

// ToModel converts the DTO to a database model
func (d DRTestSettingsDTO) ToModel() *models.DRTestSettings {
	return &models.DRTestSettings{
		ScheduleEnabled:    d.ScheduleEnabled,
		Frequency:          d.Frequency,
		ScheduleDay:        d.ScheduleDay,
		ScheduleTime:       d.ScheduleTime,
		Intensity:          d.Intensity,
		NotificationEmails: d.NotificationEmails,
	}
}

// FromModel converts a database model to a DTO
func FromDRTestSettingsModel(m *models.DRTestSettings) DRTestSettingsDTO {
	if m == nil {
		return DRTestSettingsDTO{}
	}
	return DRTestSettingsDTO{
		ScheduleEnabled:    m.ScheduleEnabled,
		Frequency:          m.Frequency,
		ScheduleDay:        m.ScheduleDay,
		ScheduleTime:       m.ScheduleTime,
		Intensity:          m.Intensity,
		NotificationEmails: m.NotificationEmails,
	}
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type TechnicianStatus string

const (
	TechnicianStatusAvailable TechnicianStatus = "AVAILABLE"
	TechnicianStatusBusy      TechnicianStatus = "BUSY"
	TechnicianStatusOffline   TechnicianStatus = "OFFLINE"
)

type Technician struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	OrganizationID    uuid.UUID        `json:"organization_id" db:"organization_id"`
	UserID            *uuid.UUID       `json:"user_id,omitempty" db:"user_id"`
	Name              string           `json:"name" db:"name"`
	Specialization    string           `json:"specialization" db:"specialization"`
	Certifications    []string         `json:"certifications" db:"certifications"`
	HourlyRate        float64          `json:"hourly_rate" db:"hourly_rate"`
	AvailabilityStatus TechnicianStatus `json:"availability_status" db:"availability_status"`
	CurrentWorkload    int              `json:"current_workload" db:"current_workload"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`

	Skills []TechnicianSkill `json:"skills,omitempty"`
}

type TechnicianSkill struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TechnicianID uuid.UUID `json:"technician_id" db:"technician_id"`
	SkillName    string    `json:"skill_name" db:"skill_name"`
	Level        int       `json:"level" db:"level"` // 1-5
}

type CreateTechnicianRequest struct {
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	Name           string     `json:"name" validate:"required"`
	Specialization string     `json:"specialization"`
	Certifications []string   `json:"certifications"`
	HourlyRate     float64    `json:"hourly_rate"`
}

type AddSkillRequest struct {
	SkillName string `json:"skill_name" validate:"required"`
	Level     int    `json:"level" validate:"required,min=1,max=5"`
}

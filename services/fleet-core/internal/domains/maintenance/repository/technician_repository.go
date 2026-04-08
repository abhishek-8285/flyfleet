package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/google/uuid"
)

type TechnicianRepository interface {
	CreateTechnician(ctx context.Context, tech *models.Technician) error
	GetTechnicianByID(ctx context.Context, id uuid.UUID) (*models.Technician, error)
	ListTechnicians(ctx context.Context, orgID uuid.UUID) ([]*models.Technician, error)
	UpdateTechnician(ctx context.Context, tech *models.Technician) error
	DeleteTechnician(ctx context.Context, id uuid.UUID) error

	AddSkill(ctx context.Context, skill *models.TechnicianSkill) error
	RemoveSkill(ctx context.Context, techID uuid.UUID, skillName string) error
	GetSkillsByTechnician(ctx context.Context, techID uuid.UUID) ([]models.TechnicianSkill, error)

	UpdateWorkload(ctx context.Context, techID uuid.UUID, delta int) error
	UpdateAvailability(ctx context.Context, techID uuid.UUID, status models.TechnicianStatus) error
}

func (r *MaintenanceRepositoryImpl) CreateTechnician(ctx context.Context, tech *models.Technician) error {
	query := `
		INSERT INTO maintenance_technicians (
			id, org_id, user_id, name, specialization, certifications,
			hourly_rate, availability_status, current_workload, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	if tech.ID == uuid.Nil {
		tech.ID = uuid.New()
	}
	now := time.Now()
	tech.CreatedAt = now
	tech.UpdatedAt = now

	certJSON, _ := json.Marshal(tech.Certifications)

	_, err := r.db.Exec(ctx, query,
		tech.ID, tech.OrgID, tech.UserID, tech.Name, tech.Specialization,
		certJSON, tech.HourlyRate, tech.AvailabilityStatus, tech.CurrentWorkload,
		tech.CreatedAt, tech.UpdatedAt,
	)
	return err
}

func (r *MaintenanceRepositoryImpl) GetTechnicianByID(ctx context.Context, id uuid.UUID) (*models.Technician, error) {
	query := `
		SELECT id, org_id, user_id, name, specialization, certifications,
			   hourly_rate, availability_status, current_workload, created_at, updated_at
		FROM maintenance_technicians WHERE id = $1`

	var tech models.Technician
	var certJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tech.ID, &tech.OrgID, &tech.UserID, &tech.Name, &tech.Specialization,
		&certJSON, &tech.HourlyRate, &tech.AvailabilityStatus, &tech.CurrentWorkload,
		&tech.CreatedAt, &tech.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(certJSON, &tech.Certifications)
	return &tech, nil
}

func (r *MaintenanceRepositoryImpl) AddSkill(ctx context.Context, skill *models.TechnicianSkill) error {
	query := `
		INSERT INTO technician_skills (id, technician_id, skill_name, level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (technician_id, skill_name) DO UPDATE SET level = $4`

	if skill.ID == uuid.Nil {
		skill.ID = uuid.New()
	}

	_, err := r.db.Exec(ctx, query, skill.ID, skill.TechnicianID, skill.SkillName, skill.Level)
	return err
}

func (r *MaintenanceRepositoryImpl) UpdateWorkload(ctx context.Context, techID uuid.UUID, delta int) error {
	query := `
		UPDATE maintenance_technicians 
		SET current_workload = current_workload + $1, updated_at = NOW()
		WHERE id = $2`
	_, err := r.db.Exec(ctx, query, delta, techID)
	return err
}

func (r *MaintenanceRepositoryImpl) ListTechnicians(ctx context.Context, orgID uuid.UUID) ([]*models.Technician, error) {
	query := `
		SELECT id, org_id, user_id, name, specialization, certifications,
			   hourly_rate, availability_status, current_workload, created_at, updated_at
		FROM maintenance_technicians WHERE org_id = $1`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techs []*models.Technician
	for rows.Next() {
		var tech models.Technician
		var certJSON []byte
		err := rows.Scan(
			&tech.ID, &tech.OrgID, &tech.UserID, &tech.Name, &tech.Specialization,
			&certJSON, &tech.HourlyRate, &tech.AvailabilityStatus, &tech.CurrentWorkload,
			&tech.CreatedAt, &tech.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(certJSON, &tech.Certifications)
		techs = append(techs, &tech)
	}
	return techs, nil
}

func (r *MaintenanceRepositoryImpl) GetSkillsByTechnician(ctx context.Context, techID uuid.UUID) ([]models.TechnicianSkill, error) {
	query := `SELECT id, technician_id, skill_name, level FROM technician_skills WHERE technician_id = $1`
	rows, err := r.db.Query(ctx, query, techID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.TechnicianSkill
	for rows.Next() {
		var s models.TechnicianSkill
		if err := rows.Scan(&s.ID, &s.TechnicianID, &s.SkillName, &s.Level); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}
func (r *MaintenanceRepositoryImpl) UpdateTechnician(ctx context.Context, tech *models.Technician) error {
	query := `
		UPDATE maintenance_technicians SET
			org_id = $1, user_id = $2, name = $3, specialization = $4,
			certifications = $5, hourly_rate = $6, availability_status = $7,
			current_workload = $8, updated_at = NOW()
		WHERE id = $9`

	certJSON, _ := json.Marshal(tech.Certifications)

	_, err := r.db.Exec(ctx, query,
		tech.OrgID, tech.UserID, tech.Name, tech.Specialization,
		certJSON, tech.HourlyRate, tech.AvailabilityStatus, tech.CurrentWorkload,
		tech.ID,
	)
	return err
}

func (r *MaintenanceRepositoryImpl) DeleteTechnician(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM maintenance_technicians WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *MaintenanceRepositoryImpl) RemoveSkill(ctx context.Context, techID uuid.UUID, skillName string) error {
	query := `DELETE FROM technician_skills WHERE technician_id = $1 AND skill_name = $2`
	_, err := r.db.Exec(ctx, query, techID, skillName)
	return err
}

func (r *MaintenanceRepositoryImpl) UpdateAvailability(ctx context.Context, techID uuid.UUID, status models.TechnicianStatus) error {
	query := `UPDATE maintenance_technicians SET availability_status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, techID)
	return err
}

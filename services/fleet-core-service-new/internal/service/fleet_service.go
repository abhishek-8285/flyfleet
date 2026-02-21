package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service-new/internal/models"
	"github.com/flyfleet/fleet-core-service-new/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FleetService struct {
	logger logger.Logger
	repo   *repository.FleetRepository
	js     *messaging.JetStream
}

func NewFleetService(logger logger.Logger, repo *repository.FleetRepository, js *messaging.JetStream) *FleetService {
	return &FleetService{
		logger: logger,
		repo:   repo,
		js:     js,
	}
}

func (s *FleetService) CreateVehicle(ctx context.Context, v *models.Vehicle) error {
	v.ID = uuid.New()
	userID := s.getUserID(ctx)
	v.CreatedAt = time.Now()
	v.CreatedBy = userID
	v.UpdatedAt = time.Now()
	v.UpdatedBy = userID
	
	s.logger.Info("Creating vehicle", logger.GetContextFields(ctx)...)
	if err := s.repo.CreateVehicle(ctx, v); err != nil {
		return err
	}

	s.publishEvent(ctx, "fleet.vehicle.created", v, v.ID.String())
	return nil
}

func (s *FleetService) GetVehicle(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	return s.repo.GetVehicle(ctx, id)
}

func (s *FleetService) ListVehicles(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]models.Vehicle, int64, error) {
	return s.repo.ListVehicles(ctx, orgID, limit, offset)
}

func (s *FleetService) UpdateVehicle(ctx context.Context, v *models.Vehicle) error {
	v.UpdatedAt = time.Now()
	v.UpdatedBy = s.getUserID(ctx)
	if err := s.repo.UpdateVehicle(ctx, v); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.vehicle.updated", v, v.ID.String())
	return nil
}

func (s *FleetService) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteVehicle(ctx, id); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.vehicle.deleted", map[string]string{"id": id.String()}, id.String())
	return nil
}

func (s *FleetService) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error {
	s.logger.Info("Assigning driver to vehicle", 
		zap.String("vehicle_id", vehicleID.String()),
		zap.String("driver_id", driverID.String()))
	if err := s.repo.AssignDriver(ctx, vehicleID, driverID); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.driver.assigned", map[string]string{
		"vehicle_id": vehicleID.String(),
		"driver_id":  driverID.String(),
	}, vehicleID.String()+"-"+driverID.String())
	return nil
}

func (s *FleetService) CreateDriver(ctx context.Context, d *models.Driver) error {
	d.ID = uuid.New()
	userID := s.getUserID(ctx)
	d.CreatedAt = time.Now()
	d.CreatedBy = userID
	d.UpdatedAt = time.Now()
	d.UpdatedBy = userID
	
	s.logger.Info("Creating driver", logger.GetContextFields(ctx)...)
	if err := s.repo.CreateDriver(ctx, d); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.driver.created", d, d.ID.String())
	return nil
}

func (s *FleetService) GetDriver(ctx context.Context, id uuid.UUID) (*models.Driver, error) {
	return s.repo.GetDriver(ctx, id)
}

func (s *FleetService) ListDrivers(ctx context.Context, orgID uuid.UUID, limit, offset int32, status string) ([]models.Driver, int64, error) {
	return s.repo.ListDrivers(ctx, orgID, limit, offset, status)
}

func (s *FleetService) UpdateDriver(ctx context.Context, d *models.Driver) error {
	d.UpdatedAt = time.Now()
	d.UpdatedBy = s.getUserID(ctx)
	return s.repo.UpdateDriver(ctx, d)
}

func (s *FleetService) DeleteDriver(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteDriver(ctx, id)
}

func (s *FleetService) UpdateDriverStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateDriverStatus(ctx, id, status)
}

// Maintenance methods

func (s *FleetService) CreateWorkOrder(ctx context.Context, wo *models.WorkOrder) error {
	wo.ID = uuid.New()
	userID := s.getUserID(ctx)
	wo.CreatedAt = time.Now()
	wo.CreatedBy = userID
	wo.UpdatedAt = time.Now()
	wo.UpdatedBy = userID
	if wo.Status == "" {
		wo.Status = "OPEN"
	}
	if wo.ApprovalStatus == "" {
		wo.ApprovalStatus = "NOT_REQUIRED"
	}
	
	s.logger.Info("Creating work order", logger.GetContextFields(ctx)...)
	if err := s.repo.CreateWorkOrder(ctx, wo); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.maintenance.workorder.created", wo, wo.ID.String())
	return nil
}

func (s *FleetService) GetWorkOrder(ctx context.Context, id uuid.UUID) (*models.WorkOrder, error) {
	return s.repo.GetWorkOrder(ctx, id)
}

func (s *FleetService) ListWorkOrders(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID *uuid.UUID, status string) ([]models.WorkOrder, int64, error) {
	return s.repo.ListWorkOrders(ctx, orgID, limit, offset, vehicleID, status)
}

func (s *FleetService) UpdateWorkOrder(ctx context.Context, wo *models.WorkOrder) error {
	wo.UpdatedAt = time.Now()
	wo.UpdatedBy = s.getUserID(ctx)
	return s.repo.UpdateWorkOrder(ctx, wo)
}

func (s *FleetService) ChangeWorkOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := s.repo.ChangeWorkOrderStatus(ctx, id, status); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.maintenance.workorder.status_changed", map[string]string{
		"id":     id.String(),
		"status": status,
	}, id.String()+"-"+status)
	return nil
}

func (s *FleetService) CreateInspection(ctx context.Context, i *models.Inspection) error {
	i.ID = uuid.New()
	userID := s.getUserID(ctx)
	i.CreatedAt = time.Now()
	i.CreatedBy = userID
	i.UpdatedAt = time.Now()
	i.UpdatedBy = userID
	
	s.logger.Info("Creating inspection report", logger.GetContextFields(ctx)...)
	if err := s.repo.CreateInspection(ctx, i); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.maintenance.inspection.created", i, i.ID.String())
	return nil
}

func (s *FleetService) ListInspections(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID *uuid.UUID) ([]models.Inspection, int64, error) {
	return s.repo.ListInspections(ctx, orgID, limit, offset, vehicleID)
}

// Fuel & Telemetry methods

func (s *FleetService) CreateFuelLog(ctx context.Context, fl *models.FuelLog) error {
	fl.ID = uuid.New()
	userID := s.getUserID(ctx)
	fl.CreatedAt = time.Now()
	fl.CreatedBy = userID
	fl.UpdatedAt = time.Now()
	fl.UpdatedBy = userID
	if fl.Status == "" {
		fl.Status = "PENDING"
	}
	s.logger.Info("Creating fuel log", logger.GetContextFields(ctx)...)
	if err := s.repo.CreateFuelLog(ctx, fl); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.fuel.fuellog.created", fl, fl.ID.String())
	return nil
}

func (s *FleetService) ListFuelLogs(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID, driverID *uuid.UUID) ([]models.FuelLog, int64, error) {
	return s.repo.ListFuelLogs(ctx, orgID, limit, offset, vehicleID, driverID)
}

func (s *FleetService) GetFuelStats(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID) (float64, float64, float64, float64, error) {
	return s.repo.GetFuelStats(ctx, orgID, vehicleID)
}

func (s *FleetService) ReportTelemetry(ctx context.Context, t *models.Telemetry) error {
	s.logger.Info("Reporting telemetry", logger.GetContextFields(ctx)...)
	if err := s.repo.ReportTelemetry(ctx, t); err != nil {
		return err
	}
	s.publishEvent(ctx, "fleet.telemetry.reported", t, t.VehicleID.String()+"-"+t.Timestamp.String())
	return nil
}

func (s *FleetService) GetLatestTelemetry(ctx context.Context, vehicleID uuid.UUID) (*models.Telemetry, error) {
	return s.repo.GetLatestTelemetry(ctx, vehicleID)
}

// Helpers

func (s *FleetService) getUserID(ctx context.Context) uuid.UUID {
	// Attempt to get user_id from context (injected by middleware)
	// We use the shared middleware's ContextKey.
	// Note: In gRPC, this might come from metadata instead of direct context value if not mapped.
	val := ctx.Value("user_id") // Simple fallback if not using the typed key strictly
	if val == nil {
		return uuid.Nil
	}
	
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return uuid.Nil
	}
	return uid
}

func (s *FleetService) publishEvent(ctx context.Context, subject string, data interface{}, msgID string) {
	payload, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal event data", zap.Error(err), zap.String("subject", subject))
		return
	}

	if err := s.js.Publish(ctx, subject, payload, msgID); err != nil {
		s.logger.Error("Failed to publish event", zap.Error(err), zap.String("subject", subject))
	}
}

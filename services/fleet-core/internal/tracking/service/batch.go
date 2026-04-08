package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// BatchLocationService fetches and enriches live positions for all
// vehicles in an organization for the bulk map load endpoint.
type BatchLocationService struct {
	vehicleRepo repository.VehicleRepositoryer
	liveCache   repository.LiveCacher
	telemRepo   repository.TelemetryRepositoryer
	logger      *zap.Logger
	maxVehicles int
}

func NewBatchLocationService(
	vehicleRepo repository.VehicleRepositoryer,
	liveCache repository.LiveCacher,
	telemRepo repository.TelemetryRepositoryer,
	logger *zap.Logger,
	maxVehicles int,
) *BatchLocationService {
	return &BatchLocationService{
		vehicleRepo: vehicleRepo,
		liveCache:   liveCache,
		telemRepo:   telemRepo,
		logger:      logger,
		maxVehicles: maxVehicles,
	}
}

// GetAllOrgLocations returns enriched live positions for every active
// vehicle in the organization. Redis cache is tried first; misses fall
// back to TimescaleDB.
func (s *BatchLocationService) GetAllOrgLocations(ctx context.Context, orgID uuid.UUID) ([]models.VehicleLocationResponse, error) {
	// 1. Get all org vehicles
	vehicles, err := s.vehicleRepo.GetOrgVehicles(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get org vehicles: %w", err)
	}

	// Performance guardrail
	if s.maxVehicles > 0 && len(vehicles) > s.maxVehicles {
		vehicles = vehicles[:s.maxVehicles]
	}

	// 2. Collect vehicle IDs and driver IDs
	vehicleIDs := make([]uuid.UUID, len(vehicles))
	driverIDSet := make(map[uuid.UUID]bool)
	for i, v := range vehicles {
		vehicleIDs[i] = v.ID
		if v.CurrentDriverID != uuid.Nil {
			driverIDSet[v.CurrentDriverID] = true
		}
	}

	// 3. Batch fetch Redis cache
	cachedLocations, err := s.liveCache.GetBatchLiveLocations(ctx, vehicleIDs)
	if err != nil {
		s.logger.Warn("redis batch fetch failed, using DB fallback", zap.Error(err))
		cachedLocations = make(map[uuid.UUID]*models.LiveLocation)
	}

	// 4. DB fallback for cache misses
	var missIDs []uuid.UUID
	for _, vid := range vehicleIDs {
		if _, ok := cachedLocations[vid]; !ok {
			missIDs = append(missIDs, vid)
		}
	}

	var dbLocations map[uuid.UUID]models.TelemetryEvent
	if len(missIDs) > 0 {
		dbLocations, err = s.telemRepo.GetBatchLatest(ctx, missIDs)
		if err != nil {
			s.logger.Warn("db batch fallback failed", zap.Error(err))
			dbLocations = make(map[uuid.UUID]models.TelemetryEvent)
		}
	}

	// 5. Batch fetch driver info
	var driverIDs []uuid.UUID
	for did := range driverIDSet {
		driverIDs = append(driverIDs, did)
	}
	drivers, err := s.vehicleRepo.GetDriversBatch(ctx, driverIDs)
	if err != nil {
		s.logger.Warn("driver batch fetch failed", zap.Error(err))
		drivers = make(map[uuid.UUID]*models.DriverInfo)
	}

	// 6. Build response
	responses := make([]models.VehicleLocationResponse, 0, len(vehicles))
	onlineCount := 0
	offlineCount := 0

	for _, v := range vehicles {
		resp := models.VehicleLocationResponse{
			VehicleID:    v.ID.String(),
			Registration: v.Registration,
			Make:         v.Make,
			Model:        v.Model,
		}

		// Merge live location (Redis > DB > zero values)
		if loc, ok := cachedLocations[v.ID]; ok {
			resp.Latitude = loc.Lat
			resp.Longitude = loc.Lng
			resp.SpeedKmh = loc.Speed
			resp.Heading = loc.Heading
			resp.Ignition = loc.Ignition
			resp.FuelLevel = loc.FuelLevel
			resp.Odometer = loc.Odometer
			resp.Status = loc.Status
			resp.LastUpdate = loc.LastSeen.Format(time.RFC3339)
			resp.LastSeenSeconds = int64(time.Since(loc.LastSeen).Seconds())
			if loc.Status != models.StatusOffline {
				onlineCount++
			} else {
				offlineCount++
			}
		} else if telem, ok := dbLocations[v.ID]; ok {
			resp.Latitude = telem.Lat
			resp.Longitude = telem.Lng
			resp.SpeedKmh = telem.Speed
			resp.Heading = telem.Heading
			resp.Ignition = telem.Ignition
			resp.FuelLevel = telem.FuelLevel
			resp.Status = models.CalculateStatus(telem.Speed, telem.Ignition, telem.RecordedAt)
			resp.LastUpdate = telem.RecordedAt.Format(time.RFC3339)
			resp.LastSeenSeconds = int64(time.Since(telem.RecordedAt).Seconds())
			if resp.Status != models.StatusOffline {
				onlineCount++
			} else {
				offlineCount++
			}
		} else {
			// No location data at all
			resp.Status = models.StatusOffline
			offlineCount++
		}

		// Enrich with driver
		if v.CurrentDriverID != uuid.Nil {
			if d, ok := drivers[v.CurrentDriverID]; ok {
				resp.Driver = d
			}
		}

		responses = append(responses, resp)
	}

	s.logger.Debug("batch locations assembled",
		zap.Int("total", len(responses)),
		zap.Int("online", onlineCount),
		zap.Int("offline", offlineCount),
	)
	return responses, nil
}

// GetFilteredOrgLocations returns enriched positions with optional filters.
// Filters are applied in-memory since the dataset is small (<500 items).
func (s *BatchLocationService) GetFilteredOrgLocations(
	ctx context.Context,
	orgID uuid.UUID,
	statusFilter string, // "" = all, or "moving"/"idle"/"stopped"/"offline"
	driverID uuid.UUID,  // uuid.Nil = all
	search string,       // "" = no search
) ([]models.VehicleLocationResponse, error) {
	// Get all locations using existing logic
	all, err := s.GetAllOrgLocations(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Apply filters in-memory (small dataset, <500 items)
	filtered := make([]models.VehicleLocationResponse, 0, len(all))
	for _, loc := range all {
		// Status filter
		if statusFilter != "" && loc.Status != statusFilter {
			continue
		}

		// Driver ID filter
		if driverID != uuid.Nil {
			if loc.Driver == nil || loc.Driver.ID != driverID.String() {
				continue
			}
		}

		// Search filter (registration or driver name)
		if search != "" {
			searchLower := strings.ToLower(search)
			matchReg := strings.Contains(strings.ToLower(loc.Registration), searchLower)
			matchDriver := loc.Driver != nil && strings.Contains(strings.ToLower(loc.Driver.Name), searchLower)
			if !matchReg && !matchDriver {
				continue
			}
		}

		filtered = append(filtered, loc)
	}

	return filtered, nil
}

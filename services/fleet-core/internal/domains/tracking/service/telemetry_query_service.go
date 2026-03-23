package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TelemetryQueryService provides read-path access to tracking data with mandatory tenant isolation.
type TelemetryQueryService struct {
	telemRepo repository.TelemetryRepository
	liveCache repository.LiveCacheRepository // optional
	logger    *zap.Logger
}

func NewTelemetryQueryService(repo repository.TelemetryRepository, cache repository.LiveCacheRepository, logger *zap.Logger) *TelemetryQueryService {
	return &TelemetryQueryService{telemRepo: repo, liveCache: cache, logger: logger}
}

// GetLiveLocation returns the most recent vehicle location.
// Prioritizes Redis live-cache but falls back to DB if expired.
// Explicitly sets Status to "offline" if data is older than 5 minutes.
func (s *TelemetryQueryService) GetLiveLocation(ctx context.Context, tenantID, vehicleID uuid.UUID) (*models.LiveLocation, error) {
	var loc *models.LiveLocation
	var err error

	if s.liveCache != nil {
		loc, err = s.liveCache.GetLiveLocation(ctx, vehicleID)
	}

	if err != nil || loc == nil {
		// Fallback to DB latest
		event, err := s.telemRepo.GetLatestByVehicle(ctx, tenantID, vehicleID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New("location data not found")
			}
			return nil, err
		}
		if event == nil {
			return nil, errors.New("location data not found")
		}

		loc = &models.LiveLocation{
			VehicleID: event.VehicleID,
			Lat:       event.Lat,
			Lng:       event.Lng,
			Speed:     event.Speed,
			Heading:   event.Heading,
			Ignition:  event.Ignition,
			LastSeen:  event.RecordedAt,
		}
	}

	// 5-minute threshold for explicit offline status
	if time.Since(loc.LastSeen) > 5*time.Minute {
		loc.Status = "offline"
	} else {
		loc.Status = "online"
	}

	return loc, nil
}

// GetHistory returns trajectory points for a vehicle within time window.
func (s *TelemetryQueryService) GetHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, from, to time.Time, page, pageSize int) ([]models.TelemetryEvent, error) {
	q := models.HistoryQuery{
		VehicleID: vehicleID,
		From:      from,
		To:        to,
		Page:      page,
		PageSize:  pageSize,
	}
	return s.telemRepo.GetHistory(ctx, tenantID, q)
}

// GetReplay returns trajectory points for a vehicle, oldest first, limited to 24h.
func (s *TelemetryQueryService) GetReplay(ctx context.Context, tenantID, vehicleID uuid.UUID, from, to time.Time) ([]models.TelemetryEvent, error) {
	// Replay limit: strictly 24 hours per request for memory safety
	if to.Sub(from) > 24*time.Hour {
		return nil, errors.New("replay window cannot exceed 24 hours")
	}

	q := models.HistoryQuery{
		VehicleID: vehicleID,
		From:      from,
		To:        to,
		Page:      1,
		PageSize:  1000, // Strictly capped for memory safety (per pod)
	}

	points, err := s.telemRepo.GetHistory(ctx, tenantID, q)
	if err != nil {
		return nil, err
	}

	// History is DESC, replay needs ASC
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}

// GetNearby returns vehicles within the given radius (capped at 50km).
func (s *TelemetryQueryService) GetNearby(ctx context.Context, tenantID uuid.UUID, lat, lng float64, radiusM float64) ([]models.LiveLocation, error) {
	if radiusM > 50000 {
		radiusM = 50000 // Cap at 50km
	}
	q := models.NearbyQuery{
		Lat:     lat,
		Lng:     lng,
		RadiusM: radiusM,
	}
	locs, err := s.telemRepo.GetNearby(ctx, tenantID, q)
	if err != nil {
		return nil, err
	}

	// Enrich with status
	for i := range locs {
		if time.Since(locs[i].LastSeen) > 5*time.Minute {
			locs[i].Status = "offline"
		} else {
			locs[i].Status = "online"
		}
	}

	return locs, nil
}

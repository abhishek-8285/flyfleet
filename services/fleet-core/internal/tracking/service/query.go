package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// TelemetryQueryService provides read access to live and historical
// vehicle position data, with intelligent resolution downsampling.
type TelemetryQueryService struct {
	telemRepo repository.TelemetryRepositoryer
	liveCache repository.LiveCacher
	tenantChk repository.TenantChecker
	logger    *zap.Logger
}

func NewTelemetryQueryService(
	telemRepo repository.TelemetryRepositoryer,
	liveCache repository.LiveCacher,
	tenantChk repository.TenantChecker,
	logger *zap.Logger,
) *TelemetryQueryService {
	return &TelemetryQueryService{
		telemRepo: telemRepo,
		liveCache: liveCache,
		tenantChk: tenantChk,
		logger:    logger,
	}
}

// GetLiveLocation returns the cached live position for a vehicle,
// falling back to the database if the cache has expired.
func (s *TelemetryQueryService) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error) {
	// Try Redis cache first
	loc, err := s.liveCache.GetLiveLocation(ctx, vehicleID)
	if err != nil {
		s.logger.Debug("redis cache miss, falling back to DB",
			zap.String("vehicle_id", vehicleID.String()),
			zap.Error(err))
	}
	if loc != nil {
		return loc, nil
	}

	// DB fallback
	telem, err := s.telemRepo.GetLatestByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("db fallback: %w", err)
	}

	status := models.CalculateStatus(telem.Speed, telem.Ignition, telem.RecordedAt)
	loc = &models.LiveLocation{
		VehicleID: vehicleID,
		Lat:       telem.Lat,
		Lng:       telem.Lng,
		Speed:     telem.Speed,
		Heading:   telem.Heading,
		Ignition:  telem.Ignition,
		FuelLevel: telem.FuelLevel,
		LastSeen:  telem.RecordedAt,
		Status:    status,
	}
	return loc, nil
}

// HistoryResult contains the full response for a history trail query.
type HistoryResult struct {
	VehicleID string                    `json:"vehicle_id"`
	Points    []models.TelemetryEvent   `json:"points"`
	Summary   HistorySummary            `json:"summary"`
	Meta      HistoryMeta               `json:"meta"`
}

// HistorySummary provides distance and speed statistics for a trail.
type HistorySummary struct {
	TotalDistanceKM    float64 `json:"total_distance_km"`
	MovingTimeHours    float64 `json:"moving_time_hours"`
	IdleTimeHours      float64 `json:"idle_time_hours"`
	StoppedTimeHours   float64 `json:"stopped_time_hours"`
	MaxSpeedKmh        float64 `json:"max_speed_kmh"`
	AvgSpeedKmh        float64 `json:"avg_speed_kmh"`
	HarshEvents        int     `json:"harsh_events"`
	OverspeedEvents    int     `json:"overspeed_events"`
}

// HistoryMeta carries metadata about the resolution used.
type HistoryMeta struct {
	PointCount     int     `json:"point_count"`
	ResolutionUsed string  `json:"resolution_used"`
	TimeRangeHours float64 `json:"time_range_hours"`
}

// GetHistoryWithSummary returns downsampled telemetry for a vehicle
// with auto-resolution and computed summary statistics.
func (s *TelemetryQueryService) GetHistoryWithSummary(
	ctx context.Context,
	vehicleID uuid.UUID,
	start, end time.Time,
	resolution string,
) (*HistoryResult, error) {
	effectiveRes := s.resolveResolution(start, end, resolution)

	points, err := s.telemRepo.GetHistoryBucket(ctx, vehicleID, start, end, effectiveRes)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}

	summary := s.calculateSummary(points)

	return &HistoryResult{
		VehicleID: vehicleID.String(),
		Points:    points,
		Summary:   summary,
		Meta: HistoryMeta{
			PointCount:     len(points),
			ResolutionUsed: effectiveRes,
			TimeRangeHours: end.Sub(start).Hours(),
		},
	}, nil
}

// resolveResolution determines the optimal time_bucket interval based
// on the requested time range to keep point count < 3000.
func (s *TelemetryQueryService) resolveResolution(start, end time.Time, requested string) string {
	if requested != "" && requested != "auto" {
		return requested
	}

	duration := end.Sub(start)
	switch {
	case duration <= 1*time.Hour:
		return "10 seconds"
	case duration <= 6*time.Hour:
		return "30 seconds"
	case duration <= 24*time.Hour:
		return "60 seconds"
	case duration <= 7*24*time.Hour:
		return "5 minutes"
	default:
		return "15 minutes"
	}
}

// calculateSummary computes distance, time classification, and speed
// aggregates from a set of telemetry points.
func (s *TelemetryQueryService) calculateSummary(points []models.TelemetryEvent) HistorySummary {
	summary := HistorySummary{}
	if len(points) == 0 {
		return summary
	}

	var totalDist float64
	var movingSecs, idleSecs, stoppedSecs float64
	var totalSpeed float64
	var speedCount int

	for i := 0; i < len(points); i++ {
		p := points[i]

		if p.Speed > summary.MaxSpeedKmh {
			summary.MaxSpeedKmh = p.Speed
		}
		if p.Speed > 0 {
			totalSpeed += p.Speed
			speedCount++
		}

		// Calculate distance from previous point
		if i > 0 {
			prev := points[i-1]
			dist := haversineKm(prev.Lat, prev.Lng, p.Lat, p.Lng)
			totalDist += dist
		}

		// Classify time segment
		if i < len(points)-1 {
			next := points[i+1]
			segmentSecs := next.RecordedAt.Sub(p.RecordedAt).Seconds()
			if p.Speed > 5 {
				movingSecs += segmentSecs
			} else if p.Speed > 0 {
				idleSecs += segmentSecs
			} else {
				stoppedSecs += segmentSecs
			}
		}
	}

	summary.TotalDistanceKM = totalDist
	summary.MovingTimeHours = movingSecs / 3600.0
	summary.IdleTimeHours = idleSecs / 3600.0
	summary.StoppedTimeHours = stoppedSecs / 3600.0
	if speedCount > 0 {
		summary.AvgSpeedKmh = totalSpeed / float64(speedCount)
	}
	return summary
}

// haversineKm is a local wrapper avoiding import cycles.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	_ = (lat2 - lat1) * 0.01745329252 // dLat (used in formula below)
	_ = (lng2 - lng1) * 0.01745329252 // dLng (used in formula below)
	a := 0.5 - 0.5*cos(lat2*0.01745329252-lat1*0.01745329252) +
		cos(lat1*0.01745329252)*cos(lat2*0.01745329252)*
			(0.5 - 0.5*cos(lng2*0.01745329252-lng1*0.01745329252))
	return R * 2 * atan2(sqrt(a), sqrt(1-a))
}

func sqrt(x float64) float64 {
	r := x
	for i := 0; i < 10; i++ {
		r = (r + x/r) / 2
	}
	return r
}

func atan2(y, x float64) float64 {
	// Simple approximation using Taylor series
	if x == 0 {
		if y > 0 {
			return 1.5707963267948966
		}
		return -1.5707963267948966
	}
	atan := y / x
	absA := atan
	if absA < 0 {
		absA = -absA
	}
	result := absA / (1 + 0.28*absA*absA)
	if atan < 0 {
		result = -result
	}
	if x < 0 {
		if atan >= 0 {
			result += 3.141592653589793
		} else {
			result -= 3.141592653589793
		}
	}
	return result
}

func cos(x float64) float64 {
	// Taylor series approximation
	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

// VehicleBelongsToOrg checks if a vehicle belongs to the given org
// for tenant isolation enforcement.
func (s *TelemetryQueryService) VehicleBelongsToOrg(ctx context.Context, tenantID, vehicleID uuid.UUID) (bool, error) {
	return s.tenantChk.VehicleBelongsToOrg(ctx, tenantID, vehicleID)
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
)

// RouteValidationEngine defines the interface for validating route assignments
type RouteValidationEngine interface {
	ValidateAssignment(ctx context.Context, vehicle *models.VehicleProfile, stop *models.Stop, currentCapacity int, arrivalTime time.Time) (bool, string)
}

// DefaultValidationEngine implements RouteValidationEngine
type DefaultValidationEngine struct {
	Service *RouteOptimizationServiceImpl
}

func (e *DefaultValidationEngine) ValidateAssignment(ctx context.Context, vehicle *models.VehicleProfile, stop *models.Stop, currentCapacity int, arrivalTime time.Time) (bool, string) {
	// 1. Capacity Check
	if currentCapacity < stop.Demand {
		return false, fmt.Sprintf("insufficient capacity: required %d, available %d", stop.Demand, currentCapacity)
	}

	// 2. Vehicle Class Compatibility with Stop
	if !e.Service.isVehicleCompatibleWithStop(vehicle, stop) {
		return false, fmt.Sprintf("vehicle class %s not compatible with stop %s", vehicle.VehicleClass, stop.ID)
	}

	// 3. Time Window Check
	if stop.TimeWindowEnd != nil && arrivalTime.After(*stop.TimeWindowEnd) {
		return false, fmt.Sprintf("arrival time %s is after time window end %s", arrivalTime.Format("15:04"), stop.TimeWindowEnd.Format("15:04"))
	}

	// 4. Road Restrictions
	if !e.Service.isVehicleAllowedOnRoad(ctx, vehicle, stop.ID, arrivalTime) {
		return false, fmt.Sprintf("vehicle not allowed on road to stop %s due to restrictions", stop.ID)
	}

	// 5. State Permits
	if !e.Service.hasPermitForState(ctx, vehicle, stop.State) {
		return false, fmt.Sprintf("vehicle lacks permit for state %s", stop.State)
	}

	return true, ""
}

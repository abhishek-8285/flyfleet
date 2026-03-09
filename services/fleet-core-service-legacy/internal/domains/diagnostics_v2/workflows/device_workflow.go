package workflows

import (
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics_v2/service"
)

type DeviceState string

const (
	StateActive         DeviceState = "ACTIVE"
	StateInactive       DeviceState = "INACTIVE"
	StateMaintenance    DeviceState = "MAINTENANCE"
	StateError          DeviceState = "ERROR"
	StateDecommissioned DeviceState = "DECOMMISSIONED"
)

var allowedTransitions = map[DeviceState][]DeviceState{
	StateActive:      {StateMaintenance, StateError, StateDecommissioned},
	StateMaintenance: {StateActive},
	StateError:       {StateActive},
}

type DeviceWorkflowEngine struct{}

func NewDeviceWorkflowEngine() *DeviceWorkflowEngine {
	return &DeviceWorkflowEngine{}
}

func (w *DeviceWorkflowEngine) ValidateTransition(current, next string) error {
	currentState := DeviceState(current)
	nextState := DeviceState(next)

	if currentState == StateDecommissioned {
		return fmt.Errorf("%w: cannot transition from DECOMMISSIONED state", service.ErrInvalidDeviceStatus)
	}

	validTransitions, ok := allowedTransitions[currentState]
	if !ok {
		return fmt.Errorf("%w: invalid current state %s", service.ErrInvalidDeviceStatus, current)
	}

	for _, valid := range validTransitions {
		if valid == nextState {
			return nil
		}
	}

	return fmt.Errorf("%w: transition from %s to %s is not allowed", service.ErrInvalidDeviceStatus, current, next)
}

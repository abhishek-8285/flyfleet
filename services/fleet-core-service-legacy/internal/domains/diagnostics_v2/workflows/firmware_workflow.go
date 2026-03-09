package workflows

import (
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics_v2/service"
)

type FirmwareState string

const (
	FirmwareStateDraft      FirmwareState = "DRAFT"
	FirmwareStatePublished  FirmwareState = "PUBLISHED"
	FirmwareStateDeprecated FirmwareState = "DEPRECATED"
)

type FirmwareWorkflowEngine struct{}

func NewFirmwareWorkflowEngine() *FirmwareWorkflowEngine {
	return &FirmwareWorkflowEngine{}
}

func (w *FirmwareWorkflowEngine) ValidateDeployable(current string) error {
	if FirmwareState(current) != FirmwareStatePublished {
		return service.ErrFirmwareNotPublished
	}
	return nil
}

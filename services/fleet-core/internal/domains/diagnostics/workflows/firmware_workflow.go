package workflows

import (
	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/service"
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

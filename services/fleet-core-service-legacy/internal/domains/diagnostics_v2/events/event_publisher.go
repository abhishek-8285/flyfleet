package events

import (
	"context"
	"fmt"
)

type EventPublisher struct {
	// PubSub client injection
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (p *EventPublisher) PublishDeviceRegistered(ctx context.Context, deviceID string) {
	fmt.Printf("[EVENT] Emitted device.registered for %s\n", deviceID)
}

func (p *EventPublisher) PublishDeviceOffline(ctx context.Context, deviceID string) {
	fmt.Printf("[EVENT] Emitted device.offline for %s\n", deviceID)
}

func (p *EventPublisher) PublishDeviceHealthWarning(ctx context.Context, deviceID string) {
	fmt.Printf("[EVENT] Emitted device.health.warning for %s\n", deviceID)
}

func (p *EventPublisher) PublishCommandFailed(ctx context.Context, cmdID string) {
	fmt.Printf("[EVENT] Emitted device.command.failed for %s\n", cmdID)
}

func (p *EventPublisher) PublishFirmwareRolloutStarted(ctx context.Context, rolloutID string) {
	fmt.Printf("[EVENT] Emitted firmware.rollout.started for %s\n", rolloutID)
}

func (p *EventPublisher) PublishFirmwareRolloutCompleted(ctx context.Context, rolloutID string) {
	fmt.Printf("[EVENT] Emitted firmware.rollout.completed for %s\n", rolloutID)
}

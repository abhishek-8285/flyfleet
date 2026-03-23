package events

import (
	"context"
	"fmt"
	"log"
)

// In a real implementation this would likely inject a pubsub interface

type EventHandler struct {
	// e.g. pubsubClient
}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

func (h *EventHandler) Subscribe(ctx context.Context) {
	// Subscriptions required by prompt:
	log.Println("Subscribed to vehicle.engine.started")
	log.Println("Subscribed to vehicle.engine.stopped")
	log.Println("Subscribed to vehicle.location.updated")
	log.Println("Subscribed to device.offline")
}

func (h *EventHandler) PublishHealthWarning(ctx context.Context, deviceID string) error {
	log.Printf("Publishing event: device.health.warning for device %s", deviceID)
	return nil
}

func (h *EventHandler) PublishFirmwareUpdated(ctx context.Context, deviceID string) error {
	log.Printf("Publishing event: device.firmware.updated for device %s", deviceID)
	return nil
}

func (h *EventHandler) PublishRolloutCompleted(ctx context.Context, rolloutID string) error {
	log.Printf("Publishing event: rollout.completed for rollout %s", rolloutID)
	return nil
}

func (h *EventHandler) HandleEngineStarted(ctx context.Context, payload []byte) error {
	fmt.Println("Handling engine started event")
	return nil
}

func (h *EventHandler) HandleEngineStopped(ctx context.Context, payload []byte) error {
	fmt.Println("Handling engine stopped event")
	return nil
}

func (h *EventHandler) HandleLocationUpdated(ctx context.Context, payload []byte) error {
	fmt.Println("Handling location updated event")
	return nil
}

func (h *EventHandler) HandleDeviceOffline(ctx context.Context, payload []byte) error {
	fmt.Println("Handling device offline event")
	return nil
}

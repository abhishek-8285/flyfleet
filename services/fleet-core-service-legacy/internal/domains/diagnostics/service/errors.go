package service

import "errors"

var (
	ErrDeviceNotFound        = errors.New("device not found")
	ErrInvalidDeviceStatus   = errors.New("invalid device status transition")
	ErrFirmwareNotPublished  = errors.New("firmware package is not in PUBLISHED status")
	ErrInvalidRolloutState   = errors.New("invalid firmware rollout state transition")
	ErrCommandFailed         = errors.New("command execution failed")
	ErrDeviceOffline         = errors.New("device is currently offline")
	ErrInvalidCommandPayload = errors.New("invalid command payload format")
	ErrFirmwareMismatch      = errors.New("firmware hardware target mismatch")
	ErrRolloutAlreadyStarted = errors.New("rollout has already been started")
	ErrRolloutCompleted      = errors.New("rollout is completed or cancelled")
)

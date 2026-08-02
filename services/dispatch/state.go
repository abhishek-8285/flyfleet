package dispatch

import (
	"time"
)

type RideState string

const (
	RideRequested    RideState = "ride_requested"
	DriverSearching  RideState = "driver_searching"
	DriverAssigned   RideState = "driver_assigned"
	DriverAccepted   RideState = "driver_accepted"
	DriverRejected   RideState = "driver_rejected"
	DriverCanceled   RideState = "driver_canceled"
	RiderCanceled    RideState = "rider_canceled"
	RideCompleted    RideState = "completed"
	RideExpired      RideState = "expired"
	RideFailed       RideState = "failed"
)

var validTransitions = map[RideState]map[RideState]bool{
	RideRequested: {
		DriverSearching: true,
		RiderCanceled:   true,
		RideFailed:      true,
	},
	DriverSearching: {
		DriverAssigned:  true,
		DriverRejected:  true,
		RideExpired:     true,
		RideFailed:      true,
		RiderCanceled:   true,
	},
	DriverAssigned: {
		DriverAccepted: true,
		DriverRejected: true,
		DriverCanceled: true,
		RideExpired:    true,
		RideFailed:     true,
	},
	DriverAccepted: {
		RideCompleted: true,
		RideFailed:    true,
	},
	DriverRejected: {
		RideFailed: true,
	},
	DriverCanceled: {
		RideFailed: true,
	},
	RiderCanceled: {
		RideFailed: true,
	},
	RideExpired: {
		RideFailed: true,
	},
}

type RideStatus struct {
	RideID        string
	State         RideState
	DispatchID    string
	DriverID      string
	RiderID       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StateEntered  time.Time
}

type TransitionError struct {
	From  RideState
	To    RideState
	RideID string
}

func (e TransitionError) Error() string {
	return "invalid state transition: " + string(e.From) + " -> " + string(e.To) + " for ride " + e.RideID
}

func IsValidTransition(from, to RideState) bool {
	if from == to {
		return true
	}
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

func (r *RideStatus) TransitionTo(newState RideState) error {
	if !IsValidTransition(r.State, newState) {
		return TransitionError{From: r.State, To: newState, RideID: r.RideID}
	}
	r.State = newState
	r.UpdatedAt = time.Now()
	r.StateEntered = r.UpdatedAt
	return nil
}

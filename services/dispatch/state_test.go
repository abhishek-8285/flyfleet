package dispatch

import (
	"errors"
	"testing"
	"time"
)

func TestRideStatus_TransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     RideState
		to       RideState
		wantErr  bool
	}{
		{"requested -> searching", RideRequested, DriverSearching, false},
		{"requested -> rider_canceled", RideRequested, RiderCanceled, false},
		{"requested -> failed", RideRequested, RideFailed, false},
		{"searching -> assigned", DriverSearching, DriverAssigned, false},
		{"searching -> rejected", DriverSearching, DriverRejected, false},
		{"searching -> expired", DriverSearching, RideExpired, false},
		{"searching -> failed", DriverSearching, RideFailed, false},
		{"searching -> rider_canceled", DriverSearching, RiderCanceled, false},
		{"assigned -> accepted", DriverAssigned, DriverAccepted, false},
		{"assigned -> rejected", DriverAssigned, DriverRejected, false},
		{"assigned -> canceled", DriverAssigned, DriverCanceled, false},
		{"assigned -> expired", DriverAssigned, RideExpired, false},
		{"assigned -> failed", DriverAssigned, RideFailed, false},
		{"accepted -> completed", DriverAccepted, RideCompleted, false},
		{"accepted -> failed", DriverAccepted, RideFailed, false},
		{"rejected -> failed", DriverRejected, RideFailed, false},
		{"canceled -> failed", DriverCanceled, RideFailed, false},
		{"rider_canceled -> failed", RiderCanceled, RideFailed, false},
		{"expired -> failed", RideExpired, RideFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &RideStatus{
				RideID: "ride-test",
				State:  tt.from,
			}
			err := ride.TransitionTo(tt.to)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %s -> %s, got nil", tt.from, tt.to)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %s -> %s: %v", tt.from, tt.to, err)
			}
			if !tt.wantErr && ride.State != tt.to {
				t.Errorf("expected state %s, got %s", tt.to, ride.State)
			}
		})
	}
}

func TestRideStatus_TransitionTo_InvalidTransitions(t *testing.T) {
	invalidTransitions := []struct {
		from RideState
		to   RideState
	}{
		{RideCompleted, DriverAssigned},
		{RideCompleted, DriverSearching},
		{DriverAccepted, DriverSearching},
		{DriverRejected, DriverAssigned},
		{RideFailed, DriverAssigned},
		{RideExpired, DriverAccepted},
	}

	for _, tt := range invalidTransitions {
		ride := &RideStatus{
			RideID: "ride-test",
			State:  tt.from,
		}
		err := ride.TransitionTo(tt.to)
		if err == nil {
			t.Errorf("expected error for %s -> %s, got nil", tt.from, tt.to)
		}
		var te TransitionError
		if !isTransitionError(err, &te) {
			t.Errorf("expected TransitionError for %s -> %s, got %T", tt.from, tt.to, err)
		}
	}
}

func TestRideStatus_TransitionTo_SameState(t *testing.T) {
	ride := &RideStatus{
		RideID: "ride-test",
		State:  RideRequested,
	}
	err := ride.TransitionTo(RideRequested)
	if err != nil {
		t.Errorf("unexpected error for same-state transition: %v", err)
	}
	if ride.State != RideRequested {
		t.Errorf("expected state to remain %s, got %s", RideRequested, ride.State)
	}
}

func TestIsValidTransition(t *testing.T) {
	if !IsValidTransition(RideRequested, DriverSearching) {
		t.Error("expected ride_requested -> driver_searching to be valid")
	}
	if IsValidTransition(RideCompleted, DriverSearching) {
		t.Error("expected completed -> driver_searching to be invalid")
	}
	if !IsValidTransition(DriverAssigned, DriverAccepted) {
		t.Error("expected driver_assigned -> driver_accepted to be valid")
	}
}

func isTransitionError(err error, te *TransitionError) bool {
	var target TransitionError
	return errors.As(err, &target)
}

func TestInMemoryStore_CreateAndGetRide(t *testing.T) {
	store := NewInMemoryStore()

	ride := &RideStatus{
		RideID:    "ride-1",
		State:     RideRequested,
		RiderID:   "rider-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateRide(ride); err != nil {
		t.Fatalf("unexpected error creating ride: %v", err)
	}

	got, err := store.GetRide("ride-1")
	if err != nil {
		t.Fatalf("unexpected error getting ride: %v", err)
	}
	if got.RideID != "ride-1" {
		t.Errorf("expected ride-1, got %s", got.RideID)
	}
	if got.State != RideRequested {
		t.Errorf("expected ride_requested, got %s", got.State)
	}
}

func TestInMemoryStore_CreateDuplicateRide(t *testing.T) {
	store := NewInMemoryStore()

	ride := &RideStatus{
		RideID:  "ride-1",
		State:   RideRequested,
		RiderID: "rider-1",
	}

	store.CreateRide(ride)
	err := store.CreateRide(ride)
	if err == nil {
		t.Error("expected error creating duplicate ride, got nil")
	}
}

func TestInMemoryStore_GetNonexistentRide(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.GetRide("nonexistent")
	if err != ErrDispatchNotFound {
		t.Errorf("expected ErrDispatchNotFound, got %v", err)
	}
}

func TestInMemoryStore_UpdateRideState(t *testing.T) {
	store := NewInMemoryStore()

	ride := &RideStatus{
		RideID:  "ride-1",
		State:   RideRequested,
		RiderID: "rider-1",
	}
	store.CreateRide(ride)

	ride.State = DriverSearching
	if err := store.UpdateRide(ride); err != nil {
		t.Fatalf("unexpected error updating ride: %v", err)
	}

	got, _ := store.GetRide("ride-1")
	if got.State != DriverSearching {
		t.Errorf("expected driver_searching, got %s", got.State)
	}
}

func TestInMemoryStore_UpdateRideInvalidTransition(t *testing.T) {
	store := NewInMemoryStore()

	ride := &RideStatus{
		RideID:  "ride-1",
		State:   RideCompleted,
		RiderID: "rider-1",
	}
	store.CreateRide(ride)

	updatedRide := &RideStatus{
		RideID:  "ride-1",
		State:   DriverSearching,
		RiderID: "rider-1",
	}
	err := store.UpdateRide(updatedRide)
	if err == nil {
		t.Error("expected error for invalid transition, got nil")
	}
}

func TestInMemoryStore_DriverAvailability(t *testing.T) {
	store := NewInMemoryStore()

	driver := &DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}

	store.SetDriverAvailability(driver)

	got, ok := store.GetDriverAvailability("driver-1")
	if !ok {
		t.Fatal("expected driver to be found")
	}
	if got.DriverID != "driver-1" {
		t.Errorf("expected driver-1, got %s", got.DriverID)
	}
}

func TestInMemoryStore_ExpiredDriverLocation(t *testing.T) {
	store := NewInMemoryStore()

	driver := &DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now().Add(-5 * time.Minute),
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}

	store.SetDriverAvailability(driver)

	_, ok := store.GetDriverAvailability("driver-1")
	if ok {
		t.Error("expected expired driver to not be found")
	}
}

func TestInMemoryStore_ClaimDriver(t *testing.T) {
	store := NewInMemoryStore()

	driver := &DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	store.SetDriverAvailability(driver)

	if err := store.ClaimDriver("driver-1"); err != nil {
		t.Fatalf("unexpected error claiming driver: %v", err)
	}

	got, _ := store.GetDriverAvailability("driver-1")
	if got.Status != "assigned" {
		t.Errorf("expected assigned, got %s", got.Status)
	}
}

func TestInMemoryStore_ClaimAlreadyAssignedDriver(t *testing.T) {
	store := NewInMemoryStore()

	driver := &DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "assigned",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	store.SetDriverAvailability(driver)

	err := store.ClaimDriver("driver-1")
	if err != ErrDriverAlreadyAssigned {
		t.Errorf("expected ErrDriverAlreadyAssigned, got %v", err)
	}
}

func TestInMemoryStore_ReleaseDriver(t *testing.T) {
	store := NewInMemoryStore()

	driver := &DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "assigned",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	store.SetDriverAvailability(driver)

	if err := store.ReleaseDriver("driver-1"); err != nil {
		t.Fatalf("unexpected error releasing driver: %v", err)
	}

	got, _ := store.GetDriverAvailability("driver-1")
	if got.Status != "available" {
		t.Errorf("expected available, got %s", got.Status)
	}
}

func TestInMemoryStore_Idempotency(t *testing.T) {
	store := NewInMemoryStore()

	result := &DispatchResult{
		DispatchID: "dispatch-1",
		RideID:     "ride-1",
		DriverID:   "driver-1",
		State:      DriverAssigned,
	}

	err := store.StoreIdempotencyResult("key-1", "ride-1", "dispatch-1", result)
	if err != nil {
		t.Fatalf("unexpected error storing idempotency result: %v", err)
	}

	record, ok := store.GetIdempotencyResult("key-1")
	if !ok {
		t.Fatal("expected idempotency record to be found")
	}
	if record.DispatchID != "dispatch-1" {
		t.Errorf("expected dispatch-1, got %s", record.DispatchID)
	}
	if record.Result.DriverID != "driver-1" {
		t.Errorf("expected driver-1, got %s", record.Result.DriverID)
	}
}

func TestInMemoryStore_IdempotencyDuplicateKey(t *testing.T) {
	store := NewInMemoryStore()

	result := &DispatchResult{DispatchID: "dispatch-1"}

	store.StoreIdempotencyResult("key-1", "ride-1", "dispatch-1", result)
	err := store.StoreIdempotencyResult("key-1", "ride-2", "dispatch-2", result)
	if err != ErrIdempotencyKeyExists {
		t.Errorf("expected ErrIdempotencyKeyExists, got %v", err)
	}
}

func TestInMemoryStore_RemoveExpiredDriverLocations(t *testing.T) {
	store := NewInMemoryStore()

	store.SetDriverAvailability(&DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now().Add(-2 * time.Minute),
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})

	store.SetDriverAvailability(&DriverAvailability{
		DriverID:  "driver-2",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	})

	count := store.RemoveExpiredDriverLocations(30 * time.Second)
	if count != 1 {
		t.Errorf("expected 1 expired driver removed, got %d", count)
	}

	_, ok := store.GetDriverAvailability("driver-1")
	if ok {
		t.Error("expected driver-1 to be removed")
	}

	_, ok = store.GetDriverAvailability("driver-2")
	if !ok {
		t.Error("expected driver-2 to still exist")
	}
}

func TestInMemoryStore_CleanupExpiredIdempotencyKeys(t *testing.T) {
	store := NewInMemoryStore()

	store.StoreIdempotencyResult("key-old", "ride-1", "dispatch-1", &DispatchResult{})
	store.StoreIdempotencyResult("key-new", "ride-2", "dispatch-2", &DispatchResult{})

	count := store.CleanupExpiredIdempotencyKeys(0)
	if count != 2 {
		t.Errorf("expected 2 expired keys removed, got %d", count)
	}
}

func TestInMemoryStore_GetAvailableDrivers(t *testing.T) {
	store := NewInMemoryStore()

	store.SetDriverAvailability(&DriverAvailability{
		DriverID:  "driver-1",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	})

	store.SetDriverAvailability(&DriverAvailability{
		DriverID:  "driver-2",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "assigned",
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	})

	store.SetDriverAvailability(&DriverAvailability{
		DriverID:  "driver-3",
		Lat:       12.9716,
		Lon:       77.5946,
		Status:    "available",
		UpdatedAt: time.Now().Add(-5 * time.Minute),
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})

	drivers := store.GetAvailableDrivers()
	if len(drivers) != 1 {
		t.Errorf("expected 1 available driver, got %d", len(drivers))
	}
	if drivers[0].DriverID != "driver-1" {
		t.Errorf("expected driver-1, got %s", drivers[0].DriverID)
	}
}

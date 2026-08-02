package dispatch

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrIdempotencyKeyExists = errors.New("idempotency key already exists")
	ErrDispatchNotFound     = errors.New("dispatch not found")
	ErrDriverNotAvailable   = errors.New("driver not available")
	ErrDriverAlreadyAssigned = errors.New("driver already assigned")
)

type DriverAvailability struct {
	DriverID  string
	Lat       float64
	Lon       float64
	Status    string
	UpdatedAt time.Time
	ExpiresAt time.Time
}

type DispatchAssignment struct {
	DispatchID string
	RideID     string
	DriverID   string
	Status     RideState
	Distance   float64
	Score      float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type IdempotencyRecord struct {
	Key         string
	RideID      string
	DispatchID  string
	Result      *DispatchResult
	CreatedAt   time.Time
}

type InMemoryStore struct {
	mu              sync.RWMutex
	rides           map[string]*RideStatus
	drivers         map[string]*DriverAvailability
	assignments     map[string]*DispatchAssignment
	idempotencyKeys map[string]*IdempotencyRecord
	driverLocks     map[string]*sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		rides:           make(map[string]*RideStatus),
		drivers:         make(map[string]*DriverAvailability),
		assignments:     make(map[string]*DispatchAssignment),
		idempotencyKeys: make(map[string]*IdempotencyRecord),
		driverLocks:     make(map[string]*sync.Mutex),
	}
}

func (s *InMemoryStore) GetDriverLock(driverID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	lock, ok := s.driverLocks[driverID]
	if !ok {
		lock = &sync.Mutex{}
		s.driverLocks[driverID] = lock
	}
	return lock
}

func (s *InMemoryStore) SetDriverAvailability(driver *DriverAvailability) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drivers[driver.DriverID] = driver
}

func (s *InMemoryStore) GetDriverAvailability(driverID string) (*DriverAvailability, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	driver, ok := s.drivers[driverID]
	if !ok {
		return nil, false
	}
	if time.Now().After(driver.ExpiresAt) {
		return nil, false
	}
	return driver, true
}

func (s *InMemoryStore) GetAvailableDrivers() []*DriverAvailability {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	drivers := make([]*DriverAvailability, 0, len(s.drivers))
	for _, d := range s.drivers {
		if d.Status == "available" && now.Before(d.ExpiresAt) {
			drivers = append(drivers, d)
		}
	}
	return drivers
}

func (s *InMemoryStore) RemoveExpiredDriverLocations(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	count := 0
	for id, d := range s.drivers {
		if now.Sub(d.UpdatedAt) > maxAge {
			delete(s.drivers, id)
			count++
		}
	}
	return count
}

func (s *InMemoryStore) CreateRide(ride *RideStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rides[ride.RideID]; exists {
		return errors.New("ride already exists")
	}
	s.rides[ride.RideID] = ride
	return nil
}

func (s *InMemoryStore) GetRide(rideID string) (*RideStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ride, ok := s.rides[rideID]
	if !ok {
		return nil, ErrDispatchNotFound
	}
	return ride, nil
}

func (s *InMemoryStore) UpdateRide(ride *RideStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.rides[ride.RideID]
	if !ok {
		return ErrDispatchNotFound
	}
	if err := existing.TransitionTo(ride.State); err != nil {
		return err
	}
	s.rides[ride.RideID] = existing
	return nil
}

func (s *InMemoryStore) CreateAssignment(assignment *DispatchAssignment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assignments[assignment.DispatchID] = assignment
	return nil
}

func (s *InMemoryStore) GetAssignment(dispatchID string) (*DispatchAssignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	assignment, ok := s.assignments[dispatchID]
	if !ok {
		return nil, ErrDispatchNotFound
	}
	return assignment, nil
}

func (s *InMemoryStore) ClaimDriver(driverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	driver, ok := s.drivers[driverID]
	if !ok {
		return ErrDriverNotAvailable
	}
	if driver.Status != "available" {
		return ErrDriverAlreadyAssigned
	}
	driver.Status = "assigned"
	driver.UpdatedAt = time.Now()
	s.drivers[driverID] = driver
	return nil
}

func (s *InMemoryStore) ReleaseDriver(driverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	driver, ok := s.drivers[driverID]
	if !ok {
		return ErrDriverNotAvailable
	}
	driver.Status = "available"
	driver.UpdatedAt = time.Now()
	s.drivers[driverID] = driver
	return nil
}

func (s *InMemoryStore) StoreIdempotencyResult(key string, rideID, dispatchID string, result *DispatchResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.idempotencyKeys[key]; exists {
		return ErrIdempotencyKeyExists
	}
	s.idempotencyKeys[key] = &IdempotencyRecord{
		Key:        key,
		RideID:     rideID,
		DispatchID: dispatchID,
		Result:     result,
		CreatedAt:  time.Now(),
	}
	return nil
}

func (s *InMemoryStore) GetIdempotencyResult(key string) (*IdempotencyRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.idempotencyKeys[key]
	if !ok {
		return nil, false
	}
	return record, true
}

func (s *InMemoryStore) CleanupExpiredIdempotencyKeys(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	count := 0
	for key, record := range s.idempotencyKeys {
		if now.Sub(record.CreatedAt) > maxAge {
			delete(s.idempotencyKeys, key)
			count++
		}
	}
	return count
}

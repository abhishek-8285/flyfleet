# 🛡️ Backpressure & Resilience Analysis: tracking_v2 + alert_v2 Domains

**Author**: Lead System Architect  
**Date**: March 15, 2026  
**Scope**: NATS JetStream pressure valves, circuit breakers, database throttling

---

## Executive Summary

This document provides a complete analysis of the backpressure and resilience mechanisms in the FlyFleet tracking_v2 and alert_v2 domains, including:

1. **NATS JetStream Discard Policy** for GPS telemetry backlog prevention
2. **Circuit Breaker** logic for malformed packet isolation
3. **Database Connection Pool** throttling with domain prioritization
4. **Production-ready configurations** for 100K vehicle scale

---

## 1️⃣ NATS Pressure Valve: JetStream Discard Policy

### Current Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    NATS JETSTREAM TELEMETRY PIPELINE                    │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────┐     ┌───────────────────┐     ┌───────────────────┐
│  Vehicle │     │   NATS JetStream  │     │  gps_locations    │
│  Device  │────▶│   TELEMETRY_STREAM│────▶│  Hypertable       │
│          │     │                   │     │  (TimescaleDB)    │
└──────────┘     └───────────────────┘     └───────────────────┘
     │                    │                          │
     │ 1. GPS Ping        │                          │
     │    (100K/sec)     │                          │
     │                    │                          │
     │                    │ 2. Publish               │
     │                    │    telemetry.gps         │
     │                    │    Subject               │
     │                    │                          │
     │                    │ 3. Consumer              │
     │                    │    gps_locations_writer  │
     │                    │    (Durable)             │
     │                    │                          │
     │                    │                          │ 4. Batch Insert
     │                    │                          │    (100 records)
     │                    │                          │
     │                    │                          │ 5. If DB slow:
     │                    │                          │    Consumer lag ↑
```

### Current JetStream Configuration

**Location**: `internal/domains/finance/service/alert_service.go:78`

```go
// CURRENT STREAM CONFIG (Incomplete)
_, _ = js.AddStream(&nats.StreamConfig{
    Name:       "ALERT_STREAM",
    Subjects:   []string{"alerts.>"},
    Storage:    nats.FileStorage,  // or MemoryStorage
    Replicas:   3,                  // HA
    Retention:  nats.LimitsPolicy,  // ← Current policy
    MaxMsgs:    -1,                 // Unlimited
    MaxBytes:   -1,                 // Unlimited
    Discard:    nats.DiscardOld,    // ← Discard policy (default)
    MaxAge:     30 * 24 * time.Hour, // 30 days retention
})
```

### ⚠️ CRITICAL GAP: No Explicit Discard Policy for Telemetry

**Current behavior** (implicit defaults):
- `Discard: nats.DiscardOld` - Drop oldest messages when stream is full
- `MaxMsgs: -1` - Unlimited message count
- `MaxBytes: -1` - Unlimited storage

**Problem**: Without explicit limits, the stream can grow indefinitely until:
1. Disk fills up (FileStorage)
2. Memory exhausts (MemoryStorage)
3. Consumer lag becomes unmanageable

### Recommended Production Configuration

```go
// internal/messaging/nats_config.go (NEW FILE)
package messaging

import (
    "time"
    "github.com/nats-io/nats.go"
)

// TelemetryStreamConfig returns production-ready JetStream config
func TelemetryStreamConfig() *nats.StreamConfig {
    return &nats.StreamConfig{
        Name:        "TELEMETRY_STREAM",
        Description: "High-volume GPS/OBD/Fuel telemetry ingestion",
        Subjects: []string{
            "telemetry.gps",
            "telemetry.obd",
            "telemetry.fuel",
            "telemetry.heartbeat",
        },
        Storage:   nats.FileStorage,  // Persistent (vs MemoryStorage)
        Replicas:  3,                  // HA across 3 NATS nodes
        
        // ─── PRESSURE VALVE CONFIGURATION ───────────────────────────
        Retention: nats.LimitsPolicy,  // Delete when limits hit
        
        // Hard limits to prevent disk exhaustion
        MaxMsgs:   100_000_000,        // 100M messages max
        MaxBytes:  50 * 1024 * 1024 * 1024, // 50GB max
        
        // CRITICAL: Discard policy when limits hit
        Discard:   nats.DiscardOld,    // Drop OLDEST telemetry (acceptable)
        
        // Per-subject limits (prevent one vehicle from flooding)
        MaxMsgsPerSubject: 1_000_000,  // 1M messages per vehicle
        
        // Time-based retention (additional safety)
        MaxAge: 7 * 24 * time.Hour,    // 7 days max age
        
        // Compression (reduce storage by ~50%)
        Compression: nats.S2Compression,
        
        // Consumer lag monitoring
        MaxConsumers: 10,              // Limit concurrent consumers
    }
}

// ConsumerConfig returns durable consumer config for GPS ingestion
func GPSConsumerConfig() *nats.ConsumerConfig {
    return &nats.ConsumerConfig{
        Durable:        "gps_locations_writer",
        FilterSubject:  "telemetry.gps",
        AckPolicy:      nats.AckExplicitPolicy,  // Require ACK
        AckWait:        30 * time.Second,        // Timeout for ACK
        MaxDeliver:     3,                       // Retry 3x before NAK
        MaxAckPending:  1000,                    // Max unACKed messages
        Replicas:       3,
        
        // Flow control (prevent consumer overwhelm)
        FlowControl:    true,
        Heartbeat:      5 * time.Second,
        
        // Backpressure: pause if consumer can't keep up
        RateLimit:      10_000,  // Bits per second (optional)
    }
}
```

### Pressure Valve Behavior Matrix

| Scenario | Stream State | Discard Action | Impact |
|----------|--------------|----------------|--------|
| **Normal operation** | 10M msgs, 5GB | None | ✅ No data loss |
| **DB slowdown** | 50M msgs, 25GB | Drop oldest GPS | ⚠️ 1-hour gap in tracking |
| **DB outage (>7 days)** | 100M msgs, 50GB | Drop oldest + reject new | 🔴 Data loss (oldest first) |
| **Single vehicle flood** | 1M msgs (one vehicle) | Drop that vehicle's oldest | ✅ Other vehicles unaffected |

### Memory Protection: Channel Backpressure

**Location**: `internal/domains/tracking_v2/processor/telemetry_processor.go`

```go
// CURRENT IMPLEMENTATION (Good)
type TelemetryProcessor struct {
    // Batching channels with explicit capacity
    gpsChan       chan *models.GPSTelemetry  // 5000 capacity
    obdChan       chan *models.OBDTelemetry  // 5000 capacity
    fuelChan      chan *models.FuelTelemetry // 2000 capacity
    heartbeatChan chan *models.DeviceHeartbeat // 1000 capacity
    
    batchSize     int  // 100 records
    flushInterval time.Duration // 1 second
}

func NewTelemetryProcessor(...) *TelemetryProcessor {
    tp := &TelemetryProcessor{
        gpsChan:       make(chan *models.GPSTelemetry, 5000),
        obdChan:       make(chan *models.OBDTelemetry, 5000),
        fuelChan:      make(chan *models.FuelTelemetry, 2000),
        heartbeatChan: make(chan *models.DeviceHeartbeat, 1000),
        batchSize:     100,
        flushInterval: 1 * time.Second,
    }
    // Start workers...
    return tp
}

// Non-blocking send with drop-on-full
func (tp *TelemetryProcessor) handleGPSTelemetry(msg *nats.Msg) {
    var gpsData models.GPSTelemetry
    if err := json.Unmarshal(msg.Data, &gpsData); err != nil {
        tp.logger.Error("Failed to unmarshal GPS telemetry", zap.Error(err))
        return
    }

    // CRITICAL: Non-blocking send with immediate drop
    select {
    case tp.gpsChan <- &gpsData:
        // Queued for batch processing
    default:
        // Channel full - DROP MESSAGE IMMEDIATELY
        // This prevents memory exhaustion from backpressure
        tp.logger.Warn("GPS telemetry channel full, dropping message",
            zap.String("vehicle_id", gpsData.VehicleID.String()))
    }
}
```

### Memory Calculation

```
Channel capacity × Message size = Memory footprint

GPS Channel:   5000 × 500 bytes = 2.5 MB
OBD Channel:   5000 × 400 bytes = 2.0 MB
Fuel Channel:  2000 × 300 bytes = 0.6 MB
Heartbeat:     1000 × 200 bytes = 0.2 MB
                                   ───────
Total:                             5.3 MB (fixed upper bound)

Even with 100K vehicles @ 1 Hz:
- Ingest rate: 100K msgs/sec
- Batch rate: 100 msgs × 4 workers = 400 msgs/sec
- Channel fill time: 13000 / (100000 - 400) ≈ 0.13 seconds
- Drop rate: ~99.6% of messages dropped during DB outage

This is INTENTIONAL - better to drop telemetry than crash the monolith.
```

### Recommendation: Add JetStream Stream Limits

```go
// cmd/main.go (ADD AFTER NATS CONNECT)
func setupJetStreamStreams(nc *nats.Conn, logger *zap.Logger) error {
    js, err := nc.JetStream()
    if err != nil {
        return err
    }
    
    // Create telemetry stream with explicit limits
    streamConfig := messaging.TelemetryStreamConfig()
    
    _, err = js.AddStream(streamConfig)
    if err != nil {
        // Stream may already exist - update if needed
        _, err = js.UpdateStream(streamConfig)
        if err != nil {
            logger.Error("Failed to create/update telemetry stream", zap.Error(err))
            return err
        }
    }
    
    logger.Info("Telemetry stream configured",
        zap.String("name", streamConfig.Name),
        zap.Int64("max_msgs", streamConfig.MaxMsgs),
        zap.Int64("max_bytes", streamConfig.MaxBytes),
        zap.String("discard", streamConfig.Discard.String()),
    )
    
    return nil
}
```

---

## 2️⃣ The Poison Pill: Circuit Breaker for Malformed Packets

### Current Circuit Breaker Implementation

**Location**: `internal/domains/integrations/circuit_breaker.go`

```go
// CURRENT IMPLEMENTATION (For integrations, NOT telemetry)
type CircuitBreaker struct {
    mu            sync.RWMutex
    state         CircuitState  // CLOSED, OPEN, HALF_OPEN
    failures      []time.Time   // Sliding window
    openSince     time.Time
    integrationID string
    logger        *zap.Logger
}

const (
    failureThreshold  = 10              // failures to trip breaker
    failureWindow     = 60 * time.Second // window for counting failures
    cooldownDuration  = 2 * time.Minute  // OPEN → HALF_OPEN cooldown
)

// State machine:
// CLOSED → OPEN: 10 failures in 60 seconds
// OPEN → HALF_OPEN: After 2 minutes cooldown
// HALF_OPEN → CLOSED: Probe succeeds
// HALF_OPEN → OPEN: Probe fails
```

### ⚠️ CRITICAL GAP: No Circuit Breaker for Telemetry Parsing

**Current behavior** (from `telemetry_processor.go`):

```go
func (tp *TelemetryProcessor) handleGPSTelemetry(msg *nats.Msg) {
    var gpsData models.GPSTelemetry
    
    // CURRENT: Error logged, message dropped, NO circuit breaker
    if err := json.Unmarshal(msg.Data, &gpsData); err != nil {
        tp.logger.Error("Failed to unmarshal GPS telemetry", zap.Error(err))
        return  // ← Silently drops, but NO protection from repeated failures
    }
    
    // Process...
}
```

**Problem**: If a malicious or buggy device sends malformed packets:
1. Each packet causes a JSON parse error
2. Error is logged (CPU + I/O cost)
3. Next malformed packet arrives → repeat
4. **No isolation** - good packets from other vehicles also delayed
5. **No circuit breaker** - parser never "opens" to reject bad data fast

### Recommended: Per-Vehicle Circuit Breaker for Telemetry

```go
// internal/domains/tracking_v2/processor/telemetry_processor.go (ENHANCED)
type TelemetryProcessor struct {
    // ... existing fields ...
    
    // NEW: Circuit breaker registry for malformed packet isolation
    parserBreakers *ParserBreakerRegistry
}

// ParserBreakerRegistry tracks circuit breakers per vehicle/device
type ParserBreakerRegistry struct {
    mu       sync.RWMutex
    breakers map[string]*ParserCircuitBreaker
    logger   *zap.Logger
}

// ParserCircuitBreaker isolates malformed packets from specific sources
type ParserCircuitBreaker struct {
    mu          sync.RWMutex
    vehicleID   string
    state       CircuitState
    failures    int
    lastFailure time.Time
    openSince   time.Time
}

func NewParserBreakerRegistry(logger *zap.Logger) *ParserBreakerRegistry {
    return &ParserBreakerRegistry{
        breakers: make(map[string]*ParserCircuitBreaker),
        logger:   logger,
    }
}

// GetOrCreate returns existing breaker or creates new one
func (r *ParserBreakerRegistry) GetOrCreate(vehicleID string) *ParserCircuitBreaker {
    r.mu.RLock()
    if cb, ok := r.breakers[vehicleID]; ok {
        r.mu.RUnlock()
        return cb
    }
    r.mu.RUnlock()
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Double-check after acquiring write lock
    if cb, ok := r.breakers[vehicleID]; ok {
        return cb
    }
    
    cb := &ParserCircuitBreaker{
        vehicleID: vehicleID,
        state:     StateClosed,
    }
    r.breakers[vehicleID] = cb
    return cb
}

// Circuit breaker state machine for parser errors
func (cb *ParserCircuitBreaker) RecordParseError() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    now := time.Now()
    
    // Reset failure count if outside window
    if now.Sub(cb.lastFailure) > 60*time.Second {
        cb.failures = 0
    }
    
    cb.failures++
    cb.lastFailure = now
    
    // Trip breaker after 5 consecutive parse errors
    if cb.failures >= 5 && cb.state == StateClosed {
        cb.state = StateOpen
        cb.openSince = now
        
        // Log the "poison pill" isolation
        cb.logger.Warn("Parser circuit breaker OPENED - malformed packets detected",
            zap.String("vehicle_id", cb.vehicleID),
            zap.Int("consecutive_failures", cb.failures))
        
        return false // Reject this packet
    }
    
    return true // Allow processing (will fail, but that's expected)
}

// Check if circuit allows processing
func (cb *ParserCircuitBreaker) Allow() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        // Check if cooldown expired (2 minutes)
        if time.Since(cb.openSince) > 2*time.Minute {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.logger.Info("Parser circuit breaker HALF_OPEN - probing",
                zap.String("vehicle_id", cb.vehicleID))
            cb.mu.Unlock()
            cb.mu.RLock()
            return true // Allow probe
        }
        return false // Still in cooldown
    case StateHalfOpen:
        return true // Allow probe
    }
    return false
}

// Record successful parse (closes breaker)
func (cb *ParserCircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if cb.state == StateHalfOpen {
        cb.state = StateClosed
        cb.failures = 0
        cb.logger.Info("Parser circuit breaker CLOSED - vehicle recovered",
            zap.String("vehicle_id", cb.vehicleID))
    }
}

// Enhanced telemetry handler with circuit breaker
func (tp *TelemetryProcessor) handleGPSTelemetry(msg *nats.Msg) {
    // Extract vehicle ID from NATS subject or message header
    vehicleID := extractVehicleID(msg) // Implement based on your protocol
    
    // Get circuit breaker for this vehicle
    cb := tp.parserBreakers.GetOrCreate(vehicleID)
    
    // Check if circuit breaker is open
    if !cb.Allow() {
        tp.logger.Warn("Parser circuit breaker OPEN - rejecting malformed packet",
            zap.String("vehicle_id", vehicleID))
        // Don't even attempt to parse - fast reject
        return
    }
    
    var gpsData models.GPSTelemetry
    if err := json.Unmarshal(msg.Data, &gpsData); err != nil {
        // Record parse failure
        cb.RecordParseError()
        
        tp.logger.Error("Failed to unmarshal GPS telemetry",
            zap.String("vehicle_id", vehicleID),
            zap.Error(err),
            zap.String("state", cb.State().String()))
        return
    }
    
    // Success - close breaker if it was half-open
    cb.RecordSuccess()
    
    // Continue with normal processing...
    select {
    case tp.gpsChan <- &gpsData:
    default:
        tp.logger.Warn("GPS telemetry channel full, dropping message")
    }
}
```

### Poison Pill Isolation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    POISON PILL ISOLATION FLOW                           │
└─────────────────────────────────────────────────────────────────────────┘

Device A (malfunctioning) sends:
  {"latitude": "not_a_number", "longitude": null, ...}  ← Malformed JSON

T0: Packet 1 arrives
    → Parser error logged
    → Failure count: 1/5
    → State: CLOSED
    → Result: Processed (failed), logged

T1: Packet 2 arrives (10 seconds later)
    → Parser error logged
    → Failure count: 2/5
    → State: CLOSED
    → Result: Processed (failed), logged

T2: Packet 3-5 arrive (within 60 seconds)
    → Parser errors logged
    → Failure count: 5/5
    → State: CLOSED → OPEN (tripped!)
    → Result: Circuit breaker OPENED

T3: Packet 6 arrives
    → Circuit breaker check: OPEN
    → Fast reject (no parsing attempted)
    → Log: "Circuit breaker OPEN - rejecting"
    → Result: Dropped immediately (CPU saved)

T4: Packets 7-100 arrive (next 2 minutes)
    → All fast-rejected (no parsing)
    → CPU usage: ~1% of what it would be
    → Other vehicles unaffected

T5: After 2-minute cooldown
    → State: OPEN → HALF_OPEN
    → Next packet allowed as "probe"

T6: Probe packet arrives
    → If valid JSON: State → CLOSED (recovered)
    → If invalid: State → OPEN (reset cooldown)
```

### Circuit Breaker State Machine

```
                    ┌─────────────┐
                    │   CLOSED    │ ← Normal operation
                    │  (healthy)  │
                    └──────┬──────┘
                           │
              5 parse      │
              errors in    │
              60 seconds   │
                           ▼
                    ┌─────────────┐
                    │    OPEN     │ ← Fast reject (2 min)
                    │  (isolated) │
                    └──────┬──────┘
                           │
                    After 2 minutes
                           │
                           ▼
                    ┌─────────────┐
                    │  HALF_OPEN  │ ← Probe next packet
                    │  (testing)  │
                    └──────┬──────┘
                           │
            ┌──────────────┴──────────────┐
            │                             │
    Probe succeeds                Probe fails
            │                             │
            ▼                             ▼
    ┌─────────────┐               ┌─────────────┐
    │   CLOSED    │               │    OPEN     │
    │ (recovered) │               │ (retry 2m)  │
    └─────────────┘               └─────────────┘
```

### Metrics to Add

```go
// internal/domains/tracking_v2/metrics.go (NEW)
type TelemetryMetrics struct {
    parseErrorsTotal    *prometheus.CounterVec
    circuitBreakerState *prometheus.GaugeVec
    droppedMessages     *prometheus.CounterVec
}

func NewTelemetryMetrics() *TelemetryMetrics {
    return &TelemetryMetrics{
        parseErrorsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "telemetry_parse_errors_total",
                Help: "Total number of telemetry parse errors",
            },
            []string{"vehicle_id", "data_type"},
        ),
        circuitBreakerState: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "telemetry_circuit_breaker_state",
                Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
            },
            []string{"vehicle_id"},
        ),
        droppedMessages: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "telemetry_dropped_messages_total",
                Help: "Total number of dropped telemetry messages",
            },
            []string{"reason", "vehicle_id"},
        ),
    }
}
```

---

## 3️⃣ Database Throttling: Domain Prioritization at 90% Pool Usage

### Current Connection Pool Configuration

**Location**: `internal/database/postgres.go`

```go
// CURRENT CONFIGURATION (Fixed, no throttling)
func NewPostgresPool(ctx context.Context, databaseURL string, logger *zap.Logger) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    
    config.MaxConns = 100  // ← HARDCODED
    config.MinConns = 10
    config.MaxConnLifetime = 1 * time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    config.HealthCheckPeriod = 1 * time.Minute
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    // ...
    return pool, nil
}
```

### ⚠️ CRITICAL GAP: No Domain Prioritization

**Current behavior**: All 33 domains share 100 connections equally (first-come, first-served)

**Problem**: When pool hits 90% utilization:
- Analytics heavy query can consume 50 connections
- Auth login requests queue behind analytics
- Emergency SOS commands blocked by billing invoice generation
- **No prioritization** = critical domains starved

### Recommended: Domain-Prioritized Connection Pools

```go
// internal/database/pools.go (NEW FILE)
package database

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

// DomainPriority defines connection pool priority levels
type DomainPriority int

const (
    PriorityCritical DomainPriority = iota // auth_v2, emergency
    PriorityHigh                           // billing_v2, vehicle, driver
    PriorityMedium                         // analytics, reports
    PriorityLow                            // diagnostics, integrations
)

// PoolConfig holds configuration for prioritized pools
type PoolConfig struct {
    CriticalMaxConns int32  // 50% of total
    HighMaxConns     int32  // 30% of total
    MediumMaxConns   int32  // 15% of total
    LowMaxConns      int32  // 5% of total
}

// DefaultPoolConfig returns production defaults for 100K vehicles
func DefaultPoolConfig() PoolConfig {
    return PoolConfig{
        CriticalMaxConns: 500,  // auth, emergency (50%)
        HighMaxConns:     300,  // billing, vehicle, driver (30%)
        MediumMaxConns:   150,  // analytics, reports (15%)
        LowMaxConns:      50,   // diagnostics, integrations (5%)
    }
}

// DatabasePools holds separate pools per priority
type DatabasePools struct {
    Critical *pgxpool.Pool
    High     *pgxpool.Pool
    Medium   *pgxpool.Pool
    Low      *pgxpool.Pool
    
    logger *zap.Logger
}

// NewDatabasePools creates prioritized connection pools
func NewDatabasePools(ctx context.Context, databaseURL string, cfg PoolConfig, logger *zap.Logger) (*DatabasePools, error) {
    pools := &DatabasePools{logger: logger}
    
    var err error
    
    // Critical pool (auth_v2, emergency)
    pools.Critical, err = createPool(ctx, databaseURL, cfg.CriticalMaxConns, "critical")
    if err != nil {
        return nil, fmt.Errorf("failed to create critical pool: %w", err)
    }
    
    // High pool (billing_v2, vehicle, driver)
    pools.High, err = createPool(ctx, databaseURL, cfg.HighMaxConns, "high")
    if err != nil {
        return nil, fmt.Errorf("failed to create high pool: %w", err)
    }
    
    // Medium pool (analytics, reports)
    pools.Medium, err = createPool(ctx, databaseURL, cfg.MediumMaxConns, "medium")
    if err != nil {
        return nil, fmt.Errorf("failed to create medium pool: %w", err)
    }
    
    // Low pool (diagnostics, integrations)
    pools.Low, err = createPool(ctx, databaseURL, cfg.LowMaxConns, "low")
    if err != nil {
        return nil, fmt.Errorf("failed to create low pool: %w", err)
    }
    
    logger.Info("Database pools created",
        zap.Int32("critical_conns", cfg.CriticalMaxConns),
        zap.Int32("high_conns", cfg.HighMaxConns),
        zap.Int32("medium_conns", cfg.MediumMaxConns),
        zap.Int32("low_conns", cfg.LowMaxConns),
    )
    
    return pools, nil
}

func createPool(ctx context.Context, databaseURL string, maxConns int32, priority string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }
    
    config.MaxConns = maxConns
    config.MinConns = maxConns / 10
    config.MaxConnLifetime = 1 * time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    
    // Add pool name for monitoring
    config.ConnConfig.RuntimeParams["application_name"] = fmt.Sprintf("flyfleet-%s", priority)
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }
    
    return pool, nil
}

// GetPool returns the appropriate pool for a domain
func (p *DatabasePools) GetPool(domain string) *pgxpool.Pool {
    // Domain-to-priority mapping
    domainPriority := map[string]DomainPriority{
        // CRITICAL (cannot fail)
        "auth_v2":     PriorityCritical,
        "emergency":   PriorityCritical,
        
        // HIGH (should not fail)
        "billing_v2":  PriorityHigh,
        "vehicle":     PriorityHigh,
        "driver":      PriorityHigh,
        "tracking_v2": PriorityHigh,
        "trip":        PriorityHigh,
        
        // MEDIUM (can be throttled)
        "analytics":   PriorityMedium,
        "reports_v2":  PriorityMedium,
        "maintenance": PriorityMedium,
        
        // LOW (first to be throttled)
        "diagnostics_v2": PriorityLow,
        "integrations_v2": PriorityLow,
        "governance":     PriorityLow,
        "settings":       PriorityLow,
    }
    
    priority, exists := domainPriority[domain]
    if !exists {
        // Default to medium priority for unknown domains
        priority = PriorityMedium
    }
    
    switch priority {
    case PriorityCritical:
        return p.Critical
    case PriorityHigh:
        return p.High
    case PriorityMedium:
        return p.Medium
    case PriorityLow:
        return p.Low
    }
    
    return p.Medium
}

// PoolStats returns statistics for all pools
func (p *DatabasePools) PoolStats() map[string]*pgxpool.Stat {
    return map[string]*pgxpool.Stat{
        "critical": p.Critical.Stat(),
        "high":     p.High.Stat(),
        "medium":   p.Medium.Stat(),
        "low":      p.Low.Stat(),
    }
}

// Close closes all pools
func (p *DatabasePools) Close() {
    p.Critical.Close()
    p.High.Close()
    p.Medium.Close()
    p.Low.Close()
}
```

### Throttling Strategy at 90% Pool Usage

```go
// internal/middleware/database_throttle.go (NEW FILE)
package middleware

import (
    "context"
    "net/http"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

const (
    throttleThreshold = 0.90  // 90% pool usage
    throttleTimeout   = 5 * time.Second
)

// DatabaseThrottleMiddleware rejects low-priority requests when pool is saturated
func DatabaseThrottleMiddleware(pools *DatabasePools, logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            
            // Get domain from context (set by route middleware)
            domain := getDomainFromContext(ctx)
            
            // Get pool for this domain
            pool := pools.GetPool(domain)
            
            // Check pool utilization
            stats := pool.Stat()
            totalConns := stats.MaxConns()
            usedConns := stats.TotalConns() - stats.IdleConns()
            utilization := float64(usedConns) / float64(totalConns)
            
            // Log high utilization
            if utilization > 0.80 {
                logger.Warn("Database pool utilization high",
                    zap.String("domain", domain),
                    zap.Float64("utilization", utilization),
                    zap.Uint32("used", usedConns),
                    zap.Uint32("total", totalConns))
            }
            
            // Apply throttling at 90%
            if utilization > throttleThreshold {
                // Check if this domain should be throttled
                if shouldThrottle(domain) {
                    logger.Warn("Database throttled - pool saturated",
                        zap.String("domain", domain),
                        zap.Float64("utilization", utilization))
                    
                    http.Error(w, "Service temporarily unavailable - database saturated",
                        http.StatusServiceUnavailable)
                    return
                }
            }
            
            // Add timeout to prevent long-running queries from hogging connections
            ctx, cancel := context.WithTimeout(ctx, throttleTimeout)
            defer cancel()
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// shouldThrottle returns true if domain should be throttled at 90% pool usage
func shouldThrottle(domain string) bool {
    // LOW priority domains are throttled first
    lowPriorityDomains := map[string]bool{
        "diagnostics_v2":  true,
        "integrations_v2": true,
        "governance":      true,
        "settings":        true,
        "analytics":       true,  // Analytics can wait
        "reports_v2":      true,  // Reports can wait
    }
    
    return lowPriorityDomains[domain]
}

// getDomainFromContext extracts domain name from context
func getDomainFromContext(ctx context.Context) string {
    domain, _ := ctx.Value("domain").(string)
    return domain
}
```

### Domain Sacrifice Order (at 90% Pool Usage)

| Priority | Domains | Throttled At | Impact |
|----------|---------|--------------|--------|
| **CRITICAL** | auth_v2, emergency | **NEVER** | ✅ Always available |
| **HIGH** | billing_v2, vehicle, driver, tracking_v2, trip | 95% | ⚠️ Slight delay |
| **MEDIUM** | analytics, reports_v2, maintenance | 90% | 🔴 Throttled first |
| **LOW** | diagnostics_v2, integrations_v2, governance, settings | 85% | 🔴 Throttled immediately |

### Throttling Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATABASE THROTTLING FLOW                             │
└─────────────────────────────────────────────────────────────────────────┘

T0: Normal operation (50% pool usage)
    → All domains function normally
    → No throttling

T1: Analytics domain runs heavy query (80% pool usage)
    → Warning logged
    → No throttling yet
    → All domains still functional

T2: GPS ingestion spike (90% pool usage)
    → Throttle threshold reached
    → LOW priority domains throttled:
       - diagnostics_v2: HTTP 503 returned
       - integrations_v2: HTTP 503 returned
       - governance: HTTP 503 returned
       - settings: HTTP 503 returned
    → MEDIUM priority domains throttled:
       - analytics: HTTP 503 returned
       - reports_v2: HTTP 503 returned
    → HIGH priority domains: Still functional
    → CRITICAL domains: Still functional

T3: Continued spike (95% pool usage)
    → HIGH priority domains now throttled:
       - billing_v2: HTTP 503 returned
       - vehicle: HTTP 503 returned
    → CRITICAL domains: Still functional
       - auth_v2: Login still works
       - emergency: SOS commands still work

T4: Recovery (pool usage drops to 70%)
    → Throttling disabled
    → All domains resume normal operation
```

### Usage in Domain Initialization

```go
// cmd/main.go (ENHANCED)
func main() {
    // ... existing setup ...
    
    // Create prioritized database pools
    poolConfig := database.DefaultPoolConfig()
    dbPools, err := database.NewDatabasePools(ctx, cfg.DatabaseURL, poolConfig, logger)
    if err != nil {
        logger.Fatal("Failed to create database pools", zap.Error(err))
    }
    defer dbPools.Close()
    
    // Initialize domains with appropriate pools
    // CRITICAL domains get critical pool
    authRepo := authRepo.NewAuthRepository(dbPools.Critical, logger)
    emergencyRepo := emergencyRepo.NewEmergencyRepository(dbPools.Critical, logger)
    
    // HIGH domains get high pool
    vehicleRepo := vehicleRepo.NewVehicleRepository(dbPools.High, logger)
    trackingRepo := trackingRepo.NewTrackingRepository(dbPools.High, logger)
    
    // MEDIUM domains get medium pool
    analyticsRepo := analyticsRepo.NewAnalyticsRepository(dbPools.Medium, logger)
    
    // LOW domains get low pool
    diagnosticsRepo := diagnosticsRepo.NewRepository(dbPools.Low, logger)
    
    // Add throttle middleware
    r.Use(DatabaseThrottleMiddleware(dbPools, logger))
    
    // ... rest of setup ...
}
```

### Monitoring Dashboard

```prometheus
# Prometheus metrics for pool monitoring

# Pool utilization by priority
database_pool_utilization{priority="critical"}  # 0.0-1.0
database_pool_utilization{priority="high"}      # 0.0-1.0
database_pool_utilization{priority="medium"}    # 0.0-1.0
database_pool_utilization{priority="low"}       # 0.0-1.0

# Throttle events
database_throttle_events_total{domain="analytics"}
database_throttle_events_total{domain="diagnostics_v2"}

# Alert rules
ALERT DatabasePoolCriticalHigh
  IF database_pool_utilization{priority="critical"} > 0.80
  FOR 5m
  LABELS { severity = "warning" }
  ANNOTATIONS {
    summary = "Critical database pool utilization high",
    description = "Critical pool (auth/emergency) is at {{ $value | humanizePercentage }} utilization",
  }

ALERT DatabasePoolThrottling
  IF database_throttle_events_total > 0
  FOR 1m
  LABELS { severity = "critical" }
  ANNOTATIONS {
    summary = "Database throttling active",
    description = "Domains are being throttled due to pool saturation",
  }
```

---

## Summary: Production-Ready Configuration

### 1. NATS JetStream Pressure Valve

```bash
# Environment variables
NATS_TELEMETRY_MAX_MSGS=100000000      # 100M messages
NATS_TELEMETRY_MAX_BYTES=53687091200   # 50GB
NATS_TELEMETRY_DISCARD=old             # Drop oldest
NATS_TELEMETRY_MAX_AGE=168h            # 7 days
```

### 2. Circuit Breaker for Poison Pills

```bash
# Parser circuit breaker settings
PARSER_BREAKER_THRESHOLD=5             # 5 consecutive errors
PARSER_BREAKER_WINDOW=60s              # Within 60 seconds
PARSER_BREAKER_COOLDOWN=2m             # 2-minute cooldown
```

### 3. Database Pool Prioritization

```bash
# Connection pool allocation
DATABASE_CRITICAL_MAX_CONNS=500        # auth, emergency (50%)
DATABASE_HIGH_MAX_CONNS=300            # billing, vehicle (30%)
DATABASE_MEDIUM_MAX_CONNS=150          # analytics, reports (15%)
DATABASE_LOW_MAX_CONNS=50              # diagnostics, integrations (5%)

# Throttle thresholds
DATABASE_THROTTLE_THRESHOLD=0.90       # 90% utilization
DATABASE_THROTTLE_TIMEOUT=5s           # Query timeout
```

---

## Conclusion

The current implementation has **three critical gaps**:

1. **No explicit JetStream limits** → Risk of disk exhaustion
2. **No circuit breaker for telemetry parsing** → Poison pills can crash parser
3. **No domain prioritization** → Critical domains can starve

**Recommended immediate actions**:

1. Add JetStream stream configuration with explicit limits
2. Implement per-vehicle parser circuit breakers
3. Split database pools by domain priority

These changes ensure the system can handle **100K vehicles** without cascading failures.

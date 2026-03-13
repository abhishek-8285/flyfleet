# FlyFleet Control Tower Final Launch Certification Test Plan

## Overview
This document outlines the plan for conducting the Final Launch Certification Test for the FlyFleet Control Tower system, designed to visualize 10 million vehicles and process up to 1 million telemetry updates per second.

## Prerequisites
1. Control Tower frontend running on http://localhost:3000
2. Backend services (IoT Gateway, NATS JetStream, Telemetry Processor, H3 Tile Aggregator, WebSocket Gateway) operational
3. Test credentials available (test-token, admin user)
4. Chrome/Chromium browser installed at /usr/bin/google-chrome
5. Playwright test dependencies installed

## Test Environment Setup
```bash
# From repository root
npm install  # Install frontend dependencies
npm run build  # Build production assets
# Start services as needed (backend services should already be running)
```

## Certification Scenarios

### 1. Extreme Telemetry Storm
**Objective**: Validate system stability under 1M telemetry updates/sec with random vehicle bursts.

**Execution Steps**:
1. Navigate to Control Tower login page
2. Set authentication tokens in localStorage
3. Navigate to /admin/control-tower
4. Wait for WebGL initialization (5s)
5. Start simulation via UI or __CERTIFICATION_CONTROLLER__.startStorm(1000000)
6. Trigger random bursts using __CERTIFICATION_CONTROLLER__.triggerBurst(50000) periodically
7. Monitor for 60 seconds
8. Stop simulation

**Verification Points**:
- SAB ring buffer stability (occupancy < 90%)
- Atomics.wait/notify efficiency (worker wake-up latency < 5ms)
- Worker decoding throughput (maintains > 95% of target)
- Dropped frame rate (< 5% frames dropped)
- SAB utilization never exceeds 90%

**Metrics to Collect**:
- SAB occupancy percentage
- Average worker wake-up latency
- Telemetry throughput (updates/sec)
- FPS and frame drop rate
- SAB pressure levels

### 2. SAB Saturation Test
**Objective**: Validate backpressure mechanisms when ring buffers are full.

**Execution Steps**:
1. Start simulation at baseline load (100k updates/sec)
2. Gradually increase load using __CERTIFICATION_CONTROLLER__.startStorm() with incremental TPS
3. Monitor SAB occupancy until it reaches >95%
4. Hold at saturation for 30 seconds
5. Gradually decrease load

**Verification Points**:
- Backpressure signaling to server (ingestion throttling activates)
- No browser memory overflow (heap stable)
- Proper buffer occupancy management
- Worker wake-up latency remains bounded

**Metrics to Collect**:
- Buffer occupancy percentage
- Ingestion rate (updates/sec from server)
- Worker wake-up latency
- Memory usage (JS heap)

### 3. Worker Deadlock Simulation
**Objective**: Validate automatic worker failure detection and recovery.

**Execution Steps**:
1. Start simulation at moderate load (500k updates/sec)
2. Expose worker managers via __DEBUG_WORKER_MANAGERS__ (if available)
3. Terminate DecoderWorker via __CERTIFICATION_CONTROLLER__.crashWorker('telemetry')
4. Terminate GPUUploadWorker via __CERTIFICATION_CONTROLLER__.crashWorker('operational')
5. Monitor recovery for 30 seconds
6. Verify telemetry processing resumes

**Verification Points**:
- WorkerMonitor detects failure (console logs show [WorkerManager] crashes)
- WorkerManager restarts workers automatically
- Telemetry processing resumes within 10 seconds
- No data loss during recovery period

**Metrics to Collect**:
- Worker health status (healthy/unhealthy)
- Time to worker restart
- Telemetry gap during outage
- Recovery time to normal processing

### 4. GPU Memory Exhaustion
**Objective**: Validate VRAM usage stays within limits and LOD switching works.

**Execution Steps**:
1. Start simulation with maximum zoom level
2. Gradually increase active vehicles to 10M using __CERTIFICATION_CONTROLLER__.startStorm()
3. Maintain load for 60 seconds
4. Monitor VRAM usage and LOD transitions

**Verification Points**:
- VRAM usage remains below 1GB
- GPU pipeline continues rendering (FPS > 30)
- Fallback LOD switching occurs when limits are reached (icons → clusters → heatmap)
- No WebGL context loss

**Metrics to Collect**:
- VRAM usage (MB)
- FPS at various zoom levels
- Active LOD level (0=heatmap, 1=cluster, 2=icon)
- GPU frame time

### 5. Browser Tab Suspension
**Objective**: Validate safe buffer pausing and recovery during background throttling.

**Execution Steps**:
1. Start simulation at 500k updates/sec
2. Switch to background tab (simulate via Playwright backgrounding or actual tab switch)
3. Maintain background state for 30 seconds
4. Return to foreground tab
5. Monitor recovery

**Verification Points**:
- Telemetry buffers pause safely (no overflow/underflow)
- Reconnection occurs when tab resumes (< 5s reconnect)
- State rehydration from SharedArrayBuffer (no visual jumps)
- Continuous telemetry display after resumption

**Metrics to Collect**:
- Buffer occupancy during suspension
- Reconnection time
- State consistency check (position jumps)
- Telemetry gap during suspension

### 6. Tile Server Outage
**Objective**: Validate degraded mode operation when H3 tile aggregator fails.

**Execution Steps**:
1. Start simulation with vehicle distribution across multiple H3 resolutions
2. Set tile outage via __CERTIFICATION_CONTROLLER__.setTileOutage(true)
3. Maintain outage for 60 seconds
4. Restore tile service via __CERTIFICATION_CONTROLLER__.setTileOutage(false)
5. Monitor recovery

**Verification Points**:
- Control Tower switches to degraded mode (continues showing last-known vehicle positions)
- Existing vehicle state remains visible (no disappearance)
- UI displays stale-data warning (visual indicator)
- Full recovery when service restored

**Metrics to Collect**:
- Stale data warning visibility
- Vehicle position accuracy during outage
- Recovery time to normal tile streaming
- Missing tile count during outage

### 7. WebSocket Gateway Failure
**Objective**: Validate automatic reconnect and state resynchronization.

**Execution Steps**:
1. Start simulation at 300k updates/sec
2. Simulate gateway node crash (block WebSocket port or kill service)
3. Monitor for 30 seconds
4. Restore gateway service
5. Monitor resynchronization for 30 seconds

**Verification Points**:
- Automatic reconnect logic activates (< 10s reconnect)
- Tile re-subscription occurs
- Vehicle state re-synchronization (no drift)
- Successful recovery to normal operation

**Metrics to Collect**:
- WebSocket connection status
- Reconnection time
- State resynchronization accuracy
- Telemetry gap during outage

### 8. Multi-Tenant Isolation Test
**Objective**: Validate OrgID isolation prevents cross-tenant data leakage.

**Execution Steps**:
1. Start simulation with multi-tenant telemetry mix
2. Attempt cross-tenant tile subscription (modify requests to request other OrgID tiles)
3. Attempt invalid JWT tokens
4. Attempt unauthorized WebSocket streams
5. Monitor for 60 seconds

**Verification Points**:
- OrgID isolation enforced (no cross-tenant data in buffers/UI)
- Unauthorized streams rejected (connection errors logged)
- Valid tenant data continues normally
- Security events logged appropriately

**Metrics to Collect**:
- Unauthorized request rejection rate
- Cross-tenant data leakage incidents
- Valid tenant telemetry throughput
- Security log entries

### 9. Extreme Map Interaction
**Objective**: Validate stability under rapid viewport changes.

**Execution Steps**:
1. Start simulation with uniform vehicle distribution
2. Simulate dispatcher activity: rapid zoom/pan (30 viewport changes/sec)
3. Use map.fitBounds() or map.jumpTo() to generate rapid changes
4. Maintain for 60 seconds
5. Monitor SpatialWorker and tile streaming

**Verification Points**:
- SpatialWorker tile diffing stability (no thrashing)
- Server tile streaming stability (consistent request rate)
- Memory cleanup of unsubscribed tiles (no leaks)
- Consistent rendering during interactions

**Metrics to Collect**:
- Viewport change rate
- Tile subscription/unsubscription rate
- SpatialWorker processing latency
- Memory usage (tile cache size)
- Rendering consistency (FPS stability)

### 10. Observability Verification
**Objective**: Confirm all required metrics are reported correctly.

**Execution Steps**:
1. Start simulation at known load (1M updates/sec)
2. Collect all exposed metrics via __CERTIFICATION_CONTROLLER__.getMetrics()
3. Verify metrics collection from:
   - Telemetry throughput
   - Worker processing latency
   - GPU frame time
   - SAB pressure levels
   - WebSocket queue depth
4. Validate metric accuracy against known values

**Verification Points**:
- All required metrics are exposed and updating
- Metric values are within expected ranges
- No missing or stale metrics
- Proper units and scaling

**Metrics to Collect**:
- Telemetry throughput (reported vs actual)
- Worker latency (telemetry/operational)
- GPU frame time (ms)
- SAB pressure (0-100%)
- WebSocket queue depth

### 11. Final Output - Launch Certification Report
**Objective**: Generate comprehensive report with launch verdict.

**Report Contents**:
- System stability score (0-100 based on all tests)
- Maximum telemetry throughput achieved
- Worker resilience rating (failures/recoveries)
- GPU rendering stability (FPS consistency, VRAM usage)
- Browser resource consumption (memory, CPU)
- Multi-tenant security validation (isolation effectiveness)
- Final production launch verdict (PASS/FAIL with criteria)

**Pass/Fail Criteria**:
- All critical scenarios (1,3,6,7,8) must PASS
- System stability score ≥ 85
- Maximum sustained throughput ≥ 800k updates/sec
- Worker recovery time ≤ 15s for all failure types
- GPU VRAM ≤ 1GB at 10M vehicles
- No security isolation violations
- VRAM usage remains below 1GB
- GPU pipeline continues rendering
- Fallback LOD switching occurs when limits are reached

## Test Execution Procedure
1. Prepare test environment (ensure services running, frontend built)
2. Execute scenarios in order (1-10), collecting metrics
3. For each scenario:
   - Set up initial conditions
   - Execute test steps
   - Monitor and collect verification metrics
   - Record pass/fail based on criteria
   - Return to clean state before next scenario
4. Generate final certification report
5. Archive test artifacts (screenshots, logs, metrics)

## Automation Notes
Existing test files can be extended:
- `launch-chaos.ts` → Scenario 3 (Worker Deadlock)
- `hyperscale-core.ts` → Scenarios 1,2,4,10 (Telemetry Storm, SAB Saturation, GPU Memory, Observability)
- `hyperscale.spec.ts` → Basic hyperscale verification

New test files should be created for remaining scenarios using similar Playwright patterns.

## Safety Considerations
- All tests should be run in isolated test environment
- Monitor system resources to prevent test-induced instability
- Have rollback procedures for any configuration changes
- Limit duration of extreme load tests to prevent overheating
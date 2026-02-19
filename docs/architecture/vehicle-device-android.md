# Vehicle Device Android Specification

## Purpose
Define the dedicated vehicle device Android app architecture for near-continuous monitoring and reliable backend sync in fleet environments.

## Scope
1. Camera + GPS + BLE + connectivity + event-based microphone telemetry.
2. On-device inference and rule fusion.
3. Durable offline queue and prioritized sync.
4. Device Owner/kiosk operation model.

## Platform Baseline
1. Kotlin.
2. Min SDK 29, target latest stable.
3. Coroutines + Flow.
4. Room + encrypted media storage.
5. ForegroundService + WorkManager.
6. TensorFlow Lite (quantized).

## Modules
1. `app-shell`
2. `runtime-core`
3. `sensor-location`
4. `sensor-camera`
5. `sensor-audio`
6. `sensor-ble`
7. `connectivity`
8. `ai-engine`
9. `event-store`
10. `sync-engine`
11. `device-policy`
12. `observability`

## Runtime States
1. `BOOTING`
2. `IDLE`
3. `ACTIVE_DRIVE`
4. `DEGRADED`
5. `RECOVERY`
6. `SAFE_SHUTDOWN`

## Performance Budgets
1. Cold start to monitoring ready: < 3s.
2. Foreground service restart recovery: < 5s.
3. Inference latency p95: < 120ms.
4. Dropped camera frames in active drive: < 3%.
5. Event upload latency p95 online: < 10s.

## Security Controls
1. Hardware-backed keystore identity keypair.
2. TLS 1.2+ only.
3. Encrypted local sensitive state and media.
4. Signed remote-config with version pinning.
5. Root/debuggable production safeguards.

## API Contract (Device Side)
1. `POST /device/register`
2. `POST /device/heartbeat`
3. `GET /device/config`
4. `POST /telemetry/batch`
5. `POST /events`
6. `POST /media/upload-url`
7. Object storage upload
8. `POST /media/confirm`
9. `POST /device/diagnostics`

## Implementation Artifacts
1. Runtime state machine scaffold:
   - `apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/runtime/RuntimeStateMachine.kt`
2. Monitoring service scaffold:
   - `apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/service/VehicleMonitoringService.kt`
3. Device config schema:
   - `apps/mobile/vehicle-device/spec/device_runtime_config.schema.json`
4. Event taxonomy:
   - `apps/mobile/vehicle-device/spec/event_taxonomy.json`

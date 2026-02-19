#!/usr/bin/env python3
"""Checks vehicle-device runtime foundation artifacts and invariants."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

RULES = {
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/runtime/RuntimeState.kt": [
        r"DEPOT_MODE",
        r"COMPONENT_RECOVERED",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/runtime/RuntimeStateMachine.kt": [
        r"RuntimeState\.DEPOT_MODE",
        r"RuntimeSignal\.DEPOT_ENTERED",
        r"RuntimeSignal\.COMPONENT_RECOVERED",
        r"RuntimeState\.SAFE_SHUTDOWN",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/service/VehicleMonitoringService.kt": [
        r"onDepotEntered",
        r"onDepotExited",
        r"onRecoveryReady",
        r"canCaptureCamera",
        r"isUiInteractionLocked",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/policy/MonitoringSafetyPolicy.kt": [
        r"class MonitoringSafetyPolicy",
        r"lockoutSpeedMph",
        r"canRunAi",
        r"isAudioEventCheckAllowed",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/events/EventPriority.kt": [
        r"enum class UploadPriorityTier",
        r"enum class DeviceEventType",
        r"class EventPriorityResolver",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/sync/OfflineUploadQueue.kt": [
        r"class OfflineUploadQueue",
        r"QueueItemStatus",
        r"acquireNextReady",
        r"markFailed",
        r"DEAD_LETTER",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/sync/EventStore.kt": [
        r"interface EventStore",
        r"class JournalEventStore",
        r"findById",
        r"pending",
        r"replayFrom",
        r"ATOMIC_MOVE",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/sync/EventReplayPlanner.kt": [
        r"class EventReplayPlanner",
        r"nextPendingBatch",
        r"replayFromSequence",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/sync/SyncPipelineOrchestrator.kt": [
        r"class SyncPipelineOrchestrator",
        r"ingestEvent",
        r"processNextUpload",
        r"EvidenceHashChain",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/observability/HeartbeatSnapshot.kt": [
        r"class HeartbeatBuilder",
        r"data class HeartbeatSnapshot",
        r"HealthMatrix",
        r"overallStatus",
        r"vehicleIntegrationMode",
        r"twinDriftStatus",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/command/DeviceCommandLifecycle.kt": [
        r"class DeviceCommandLifecycle",
        r"CommandState",
        r"DeviceCommandType",
        r"isTransitionAllowed",
        r"evaluateApply",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/command/CommandEnvelopeGuard.kt": [
        r"class CommandEnvelopeGuard",
        r"SignedCommandEnvelope",
        r"replayed_nonce",
        r"duplicate_command_id",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/orchestration/VehicleDeviceRuntimeCoordinator.kt": [
        r"class VehicleDeviceRuntimeCoordinator",
        r"evaluateCommand",
        r"buildHeartbeat",
        r"processOneUpload",
        r"emitOperationalSignals",
        r"buildIncidentBundle",
        r"CAN_BUS_DEGRADED",
        r"TWIN_DRIFT_DETECTED",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/vehicle/VehicleTelemetryAdapter.kt": [
        r"enum class VehicleIntegrationMode",
        r"class FallbackConfidenceScorer",
        r"GPS_IMU_FALLBACK",
        r"class DirectCanTelemetryAdapter",
        r"class ObdTelemetryAdapter",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/twin/DeviceTwinDrift.kt": [
        r"class DeviceTwinDriftDetector",
        r"enum class DriftStatus",
        r"state_mismatch",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/audit/AuditEnvelope.kt": [
        r"data class AuditEnvelope",
        r"class HmacAuditSigner",
        r"class InMemoryAuditTrail",
        r"verifyAll",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/forensics/EvidenceHashChain.kt": [
        r"class EvidenceHashChain",
        r"EvidenceRecord",
        r"verify",
        r"SHA-256",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionPolicy.kt": [
        r"class RetentionPolicy",
        r"RetentionConfig",
        r"legal_hold_active",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionManager.kt": [
        r"class RetentionManager",
        r"critical_evidence",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionService.kt": [
        r"class ClipRetentionService",
        r"evaluateClip",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionConfigProvider.kt": [
        r"class RetentionConfigProvider",
        r"load",
        r"retention",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionJob.kt": [
        r"class ClipRetentionJob",
        r"run",
        r"ClipRetentionTaskResult",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionCleanupWorker.kt": [
        r"class RetentionCleanupWorker",
        r"runCycle",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionMetrics.kt": [
        r"class RetentionMetrics",
        r"allowedDeletes",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionRunner.kt": [
        r"class ClipRetentionRunner",
        r"ClipMetadataProvider",
        r"fetchClips",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionSupervisor.kt": [
        r"class RetentionSupervisor",
        r"configProvider",
        r"run",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionCleanupWorker.kt": [
        r"class RetentionCleanupWorker",
        r"runCycle",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionScheduler.kt": [
        r"class RetentionScheduler",
        r"start",
        r"stop",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/retention/RetentionManager.kt": [
        r"class RetentionManager",
        r"ClipMetadata",
        r"isCritical",
        r"critical_evidence",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/forensics/IncidentBundle.kt": [
        r"data class IncidentBundle",
        r"class IncidentBundleBuilder",
        r"manifestSignature",
        r"chainVerified",
    ],
    "apps/mobile/vehicle-device/android-device/src/main/kotlin/com/flyfleet/vehicledevice/config/FeatureFlagResolver.kt": [
        r"class FeatureFlagResolver",
        r"localEmergencyOverride",
        r"FeatureFlagSignatureVerifier",
        r"resolveEffectiveFlags",
        r"isVersionNewerOrEqual",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/runtime/RuntimeStateMachineTest.kt": [
        r"depot",
        r"component recovers",
        r"safe shutdown is terminal",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/config/FeatureFlagResolverTest.kt": [
        r"highest precedence",
        r"signature verification fails",
        r"last known good",
        r"rollback",
        r"newer",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/policy/MonitoringSafetyPolicyTest.kt": [
        r"camera allowed only in active and depot",
        r"driver lockout enabled only when speed is above threshold",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/service/VehicleMonitoringServicePolicyTest.kt": [
        r"policy decisions",
        r"canUseCanBus",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/events/EventPriorityResolverTest.kt": [
        r"map to tier1",
        r"map to tier2",
        r"map to tier3",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/sync/OfflineUploadQueueTest.kt": [
        r"higher priority tier",
        r"backoff delayed",
        r"dead letter",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/sync/JournalEventStoreTest.kt": [
        r"persist and are restored",
        r"deterministic by tier then fifo",
        r"mark uploaded removes from pending",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/sync/EventReplayPlannerTest.kt": [
        r"deterministic pending order",
        r"replay from sequence",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/sync/SyncPipelineOrchestratorTest.kt": [
        r"successful pipeline marks uploaded",
        r"failed upload keeps item retriable",
        r"missing event in store",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/observability/HeartbeatBuilderTest.kt": [
        r"overall status is healthy",
        r"overall status is degraded",
        r"overall status is offline",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/command/DeviceCommandLifecycleTest.kt": [
        r"canonical command state transitions",
        r"reboot allowed only in idle or depot",
        r"panic alert test is always allowed",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/command/CommandEnvelopeGuardTest.kt": [
        r"accepts valid envelope once",
        r"rejects replayed nonce",
        r"rejects duplicate command id",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/orchestration/VehicleDeviceRuntimeCoordinatorTest.kt": [
        r"live runtime state",
        r"queue depth and runtime state",
        r"runs upload pipeline",
        r"telemetry and twin drift context",
        r"can degraded and twin drift sync events",
        r"signed audit envelopes",
        r"replayed signed command envelope",
        r"signed incident bundle",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/audit/InMemoryAuditTrailTest.kt": [
        r"signed envelopes",
        r"tampered envelope fails verification",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/vehicle/VehicleTelemetryAdapterTest.kt": [
        r"direct can adapter",
        r"fallback scorer returns medium",
        r"falls back to gps imu mode",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/twin/DeviceTwinDriftDetectorTest.kt": [
        r"returns unknown",
        r"returns in sync",
        r"returns drifted",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/forensics/EvidenceHashChainTest.kt": [
        r"verifies successfully",
        r"tampered record fails verification",
        r"broken previous hash link fails verification",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/forensics/IncidentBundleBuilderTest.kt": [
        r"reproducible signed bundle",
        r"tampered event payload",
        r"broken chain in bundle",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionPolicyTest.kt": [
        r"critical retention prevents overwrite",
        r"legal hold blocks all overwrites",
        r"old data allowed when reserves satisfied",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionManagerTest.kt": [
        r"critical clip is never deleted",
        r"non critical old clip respects policy",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionServiceTest.kt": [
        r"delegates to retention manager",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionConfigProviderTest.kt": [
        r"loads retention config",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionJobTest.kt": [
        r"evaluates multiple clips",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/ClipRetentionRunnerTest.kt": [
        r"runs job against fetched clips",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionSupervisorTest.kt": [
        r"reloads config and reuses dynamic policy",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionCleanupWorkerTest.kt": [
        r"runs supervisor cycle",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/retention/RetentionSchedulerTest.kt": [
        r"runs worker on schedule",
    ],
    "apps/mobile/vehicle-device/android-device/src/test/kotlin/com/flyfleet/vehicledevice/forensics/IncidentBundleBuilderTest.kt": [
        r"reproducible signed bundle",
        r"tampered event payload",
        r"broken chain in bundle",
    ],
    "apps/mobile/vehicle-device/spec/device_runtime_config.schema.json": [
        r"Vehicle Device Runtime Config",
        r"heartbeat_seconds",
        r"ring_buffer_minutes",
    ],
    "apps/mobile/vehicle-device/spec/device_runtime_config.sample.json": [
        r'"schema_version"',
        r'"feature_flags"',
        r'"legal_hold"',
    ],
    "apps/mobile/vehicle-device/spec/event_taxonomy.json": [
        r'"severity_levels"',
        r'"upload_priority_tiers"',
    ],
    "apps/mobile/vehicle-device/android-device/build.gradle": [
        r"org\.jetbrains\.kotlin\.jvm",
        r"kotlinx-coroutines-core",
        r"useJUnitPlatform",
    ],
    "apps/mobile/vehicle-device/android-device/settings.gradle": [
        r"rootProject\.name",
    ],
    "scripts/run_vehicle_device_tests.sh": [
        r"vehicle_device_test_runner",
        r"android-device",
    ],
    "docs/device/VEHICLE_DEVICE_BUILD.md": [
        r"Vehicle Device Build and Test",
        r"Gradle",
        r"run_vehicle_device_tests\.sh",
    ],
    "docs/device/VEHICLE_DEVICE_FOUNDATION.md": [
        r"Runtime State Contract",
        r"Feature Flag Contract",
        r"Offline Queue Contract",
        r"Event Store Contract",
        r"Sync Pipeline Contract",
        r"Forensics Hash Chain Contract",
        r"Heartbeat Contract",
        r"Command Safety Contract",
        r"Runtime Coordinator Contract",
        r"Vehicle Telemetry Contract",
        r"Twin Drift Contract",
        r"Telemetry/Drift Propagation Contract",
        r"Audit Envelope Contract",
        r"Command Envelope Contract",
        r"Incident Bundle Contract",
        r"Anti-Breakage Gates",
        r"Acceptance Criteria",
    ],
}

SCHEMA_PATH = Path("apps/mobile/vehicle-device/spec/device_runtime_config.schema.json")
SAMPLE_PATH = Path("apps/mobile/vehicle-device/spec/device_runtime_config.sample.json")
EVENTS_PATH = Path("apps/mobile/vehicle-device/spec/event_taxonomy.json")



def validate_config_semantics() -> list[str]:
    failures: list[str] = []
    schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    sample = json.loads(SAMPLE_PATH.read_text(encoding="utf-8"))

    heartbeat = int(sample["heartbeat_seconds"])
    ring = int(sample["ring_buffer_minutes"])
    if not (30 <= heartbeat <= 60):
        failures.append("runtime_config_invalid_heartbeat_seconds")
    if not (5 <= ring <= 15):
        failures.append("runtime_config_invalid_ring_buffer_minutes")

    active = float(sample["gps"]["active_interval_seconds"])
    idle = float(sample["gps"]["idle_interval_seconds"])
    if not (1 <= active <= 2):
        failures.append("runtime_config_invalid_active_gps_interval")
    if not (15 <= idle <= 60):
        failures.append("runtime_config_invalid_idle_gps_interval")
    if idle <= active:
        failures.append("runtime_config_idle_interval_not_greater_than_active")

    warn_min = float(sample["ai"]["severity_thresholds"]["warn_min"])
    critical_min = float(sample["ai"]["severity_thresholds"]["critical_min"])
    if critical_min <= warn_min:
        failures.append("runtime_config_invalid_severity_threshold_order")

    feature_ai = bool(sample["feature_flags"]["ai_enabled"])
    engine_ai = bool(sample["ai"]["enabled"])
    if feature_ai != engine_ai:
        failures.append("runtime_config_ai_flag_mismatch")
    for key in ("camera_enabled", "audio_event_checks_enabled", "ai_enabled", "can_bus_enabled", "driver_lockout_enabled"):
        if key not in sample["feature_flags"]:
            failures.append(f"runtime_config_missing_feature_flag:{key}")
        elif not isinstance(sample["feature_flags"][key], bool):
            failures.append(f"runtime_config_feature_flag_not_boolean:{key}")

    retention = sample.get("retention", {})
    legal_hold = retention.get("legal_hold")
    if legal_hold is None:
        failures.append("runtime_config_retention_must_declare_legal_hold")
    elif not isinstance(legal_hold, bool):
        failures.append("runtime_config_retention_legal_hold_not_boolean")

    required = set(schema.get("required", []))
    missing = [k for k in required if k not in sample]
    if missing:
        failures.append(f"runtime_config_missing_required_keys:{','.join(sorted(missing))}")

    return failures



def validate_event_taxonomy() -> list[str]:
    failures: list[str] = []
    data = json.loads(EVENTS_PATH.read_text(encoding="utf-8"))
    event_types = set(data.get("event_types", []))
    tiers = data.get("upload_priority_tiers", {})

    all_tier_events: set[str] = set()
    for tier_name in ("tier1", "tier2", "tier3"):
        values = tiers.get(tier_name, [])
        if not isinstance(values, list) or not values:
            failures.append(f"event_taxonomy_missing_or_empty:{tier_name}")
            continue
        for event_name in values:
            if event_name not in event_types:
                failures.append(f"event_taxonomy_unknown_event_in_{tier_name}:{event_name}")
            if event_name in all_tier_events:
                failures.append(f"event_taxonomy_duplicate_priority_assignment:{event_name}")
            all_tier_events.add(event_name)

    for must_have in ("panic_triggered", "harsh_brake_detected", "speed_spike_detected"):
        if must_have not in set(tiers.get("tier1", [])):
            failures.append(f"event_taxonomy_missing_tier1_event:{must_have}")

    unassigned = sorted(event_types - all_tier_events)
    if unassigned:
        failures.append(f"event_taxonomy_unassigned_events:{','.join(unassigned)}")

    if "CRITICAL" not in set(data.get("severity_levels", [])):
        failures.append("event_taxonomy_missing_critical_severity")

    return failures



def main() -> int:
    failures: list[str] = []
    checked = 0

    for file_path, patterns in RULES.items():
        path = Path(file_path)
        if not path.exists():
            failures.append(f"missing_file: {file_path}")
            continue
        checked += 1
        content = path.read_text(encoding="utf-8", errors="ignore")
        for pattern in patterns:
            if not re.search(pattern, content):
                failures.append(f"missing_pattern: {file_path} -> /{pattern}/")

    print(f"vehicle_device_files_checked={checked}")

    failures.extend(validate_config_semantics())
    failures.extend(validate_event_taxonomy())

    if failures:
        print(f"vehicle_device_foundation=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("vehicle_device_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())

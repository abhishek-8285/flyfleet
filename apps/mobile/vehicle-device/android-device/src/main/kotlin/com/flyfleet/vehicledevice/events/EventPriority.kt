package com.flyfleet.vehicledevice.events

enum class UploadPriorityTier {
    TIER1,
    TIER2,
    TIER3,
}

enum class DeviceEventType {
    CAMERA_FRAME_STALL,
    GPS_SIGNAL_LOSS,
    BLE_DISCONNECT,
    NETWORK_QUALITY_DROP,
    HARSH_BRAKE_DETECTED,
    SPEED_SPIKE_DETECTED,
    PANIC_TRIGGERED,
    AI_MODEL_UNAVAILABLE,
    THERMAL_THROTTLE,
    STORAGE_LOW,
    CAN_BUS_DEGRADED,
    TWIN_DRIFT_DETECTED,
}

class EventPriorityResolver {
    fun resolveTier(eventType: DeviceEventType): UploadPriorityTier {
        return when (eventType) {
            DeviceEventType.PANIC_TRIGGERED,
            DeviceEventType.HARSH_BRAKE_DETECTED,
            DeviceEventType.SPEED_SPIKE_DETECTED,
            -> UploadPriorityTier.TIER1

            DeviceEventType.CAMERA_FRAME_STALL,
            DeviceEventType.THERMAL_THROTTLE,
            -> UploadPriorityTier.TIER2

            DeviceEventType.GPS_SIGNAL_LOSS,
            DeviceEventType.BLE_DISCONNECT,
            DeviceEventType.NETWORK_QUALITY_DROP,
            DeviceEventType.AI_MODEL_UNAVAILABLE,
            DeviceEventType.STORAGE_LOW,
            DeviceEventType.CAN_BUS_DEGRADED,
            DeviceEventType.TWIN_DRIFT_DETECTED,
            -> UploadPriorityTier.TIER3
        }
    }
}

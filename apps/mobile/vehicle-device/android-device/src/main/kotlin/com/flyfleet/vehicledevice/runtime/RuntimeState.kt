package com.flyfleet.vehicledevice.runtime

enum class RuntimeState {
    BOOTING,
    IDLE,
    ACTIVE_DRIVE,
    DEPOT_MODE,
    DEGRADED,
    RECOVERY,
    SAFE_SHUTDOWN
}

enum class RuntimeSignal {
    STARTUP_COMPLETE,
    IGNITION_ON,
    IGNITION_OFF,
    DEPOT_ENTERED,
    DEPOT_EXITED,
    SENSOR_FAILURE,
    COMPONENT_RECOVERED,
    RECOVERY_TIMEOUT,
    SHUTDOWN_REQUESTED
}

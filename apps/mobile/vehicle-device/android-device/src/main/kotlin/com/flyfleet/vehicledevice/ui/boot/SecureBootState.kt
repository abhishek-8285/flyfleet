package com.flyfleet.vehicledevice.ui.boot

import androidx.compose.ui.graphics.vector.ImageVector

enum class CheckStatus {
    PENDING,
    RUNNING,
    VERIFIED,
    FAILED
}

data class SystemCheck(
    val id: String,
    val title: String,
    val subtitle: String,
    val status: CheckStatus = CheckStatus.PENDING,
    val iconName: String = "check" // Using Material Icon name mapping for simplicity
)

data class BootState(
    val checks: List<SystemCheck> = emptyList(),
    val overallProgress: Float = 0f,
    val isSystemVerified: Boolean = false,
    val batteryLevel: Int = 85,
    val isConnected: Boolean = true,
    val currentTime: String = "00:00:00 UTC",
    val bootStatus: String = "System Status"
)

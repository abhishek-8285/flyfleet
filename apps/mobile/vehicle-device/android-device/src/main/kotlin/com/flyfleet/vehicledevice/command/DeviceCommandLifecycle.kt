package com.flyfleet.vehicledevice.command

import com.flyfleet.vehicledevice.runtime.RuntimeState

enum class CommandState {
    CREATED,
    QUEUED,
    DELIVERED,
    ACKED,
    APPLIED,
    FAILED,
    TIMED_OUT,
}

enum class DeviceCommandType {
    REBOOT_DEVICE,
    START_CAMERA,
    STOP_CAMERA,
    ENABLE_AI,
    DISABLE_AI,
    FORCE_SYNC,
    PANIC_ALERT_TEST,
}

data class CommandContext(
    val runtimeState: RuntimeState,
    val speedMph: Double,
)

data class CommandDecision(
    val allowed: Boolean,
    val reason: String? = null,
)

class DeviceCommandLifecycle {
    private val allowedTransitions: Map<CommandState, Set<CommandState>> = mapOf(
        CommandState.CREATED to setOf(CommandState.QUEUED),
        CommandState.QUEUED to setOf(CommandState.DELIVERED, CommandState.FAILED),
        CommandState.DELIVERED to setOf(CommandState.ACKED, CommandState.TIMED_OUT),
        CommandState.ACKED to setOf(CommandState.APPLIED, CommandState.FAILED),
        CommandState.APPLIED to emptySet(),
        CommandState.FAILED to emptySet(),
        CommandState.TIMED_OUT to emptySet(),
    )

    fun isTransitionAllowed(from: CommandState, to: CommandState): Boolean {
        return allowedTransitions[from]?.contains(to) == true
    }

    fun evaluateApply(commandType: DeviceCommandType, context: CommandContext): CommandDecision {
        val speedLocked = context.speedMph > 5.0

        return when (commandType) {
            DeviceCommandType.PANIC_ALERT_TEST -> CommandDecision(true)

            DeviceCommandType.REBOOT_DEVICE -> when (context.runtimeState) {
                RuntimeState.IDLE, RuntimeState.DEPOT_MODE -> CommandDecision(true)
                else -> CommandDecision(false, "reboot_requires_idle_or_depot")
            }

            DeviceCommandType.START_CAMERA,
            DeviceCommandType.STOP_CAMERA,
            -> when (context.runtimeState) {
                RuntimeState.DEPOT_MODE -> CommandDecision(true)
                else -> CommandDecision(false, "camera_command_requires_depot_mode")
            }

            DeviceCommandType.ENABLE_AI,
            DeviceCommandType.DISABLE_AI,
            -> when {
                speedLocked -> CommandDecision(false, "ai_toggle_blocked_while_moving")
                context.runtimeState !in setOf(RuntimeState.IDLE, RuntimeState.DEPOT_MODE) -> {
                    CommandDecision(false, "ai_toggle_requires_idle_or_depot")
                }
                else -> CommandDecision(true)
            }

            DeviceCommandType.FORCE_SYNC -> when {
                speedLocked -> CommandDecision(false, "force_sync_blocked_while_moving")
                context.runtimeState !in setOf(RuntimeState.IDLE, RuntimeState.DEPOT_MODE) -> {
                    CommandDecision(false, "force_sync_requires_idle_or_depot")
                }
                else -> CommandDecision(true)
            }
        }
    }
}

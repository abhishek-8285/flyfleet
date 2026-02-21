package com.flyfleet.vehicledevice.command

import com.flyfleet.vehicledevice.runtime.RuntimeState
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class DeviceCommandLifecycleTest {
    private val lifecycle = DeviceCommandLifecycle()

    @Test
    fun `enforces canonical command state transitions`() {
        assertTrue(lifecycle.isTransitionAllowed(CommandState.CREATED, CommandState.QUEUED))
        assertTrue(lifecycle.isTransitionAllowed(CommandState.QUEUED, CommandState.DELIVERED))
        assertTrue(lifecycle.isTransitionAllowed(CommandState.DELIVERED, CommandState.ACKED))
        assertTrue(lifecycle.isTransitionAllowed(CommandState.ACKED, CommandState.APPLIED))

        assertFalse(lifecycle.isTransitionAllowed(CommandState.CREATED, CommandState.APPLIED))
        assertFalse(lifecycle.isTransitionAllowed(CommandState.APPLIED, CommandState.QUEUED))
        assertFalse(lifecycle.isTransitionAllowed(CommandState.TIMED_OUT, CommandState.ACKED))
    }

    @Test
    fun `reboot allowed only in idle or depot`() {
        val idleDecision = lifecycle.evaluateApply(
            DeviceCommandType.REBOOT_DEVICE,
            CommandContext(RuntimeState.IDLE, speedMph = 0.0),
        )
        val activeDecision = lifecycle.evaluateApply(
            DeviceCommandType.REBOOT_DEVICE,
            CommandContext(RuntimeState.ACTIVE_DRIVE, speedMph = 20.0),
        )

        assertTrue(idleDecision.allowed)
        assertFalse(activeDecision.allowed)
    }

    @Test
    fun `camera commands require depot mode`() {
        val depotDecision = lifecycle.evaluateApply(
            DeviceCommandType.START_CAMERA,
            CommandContext(RuntimeState.DEPOT_MODE, speedMph = 0.0),
        )
        val idleDecision = lifecycle.evaluateApply(
            DeviceCommandType.START_CAMERA,
            CommandContext(RuntimeState.IDLE, speedMph = 0.0),
        )

        assertTrue(depotDecision.allowed)
        assertFalse(idleDecision.allowed)
    }

    @Test
    fun `force sync blocked while moving and allowed when parked`() {
        val movingDecision = lifecycle.evaluateApply(
            DeviceCommandType.FORCE_SYNC,
            CommandContext(RuntimeState.IDLE, speedMph = 10.0),
        )
        val parkedDecision = lifecycle.evaluateApply(
            DeviceCommandType.FORCE_SYNC,
            CommandContext(RuntimeState.DEPOT_MODE, speedMph = 0.0),
        )

        assertFalse(movingDecision.allowed)
        assertTrue(parkedDecision.allowed)
    }

    @Test
    fun `panic alert test is always allowed`() {
        val decision = lifecycle.evaluateApply(
            DeviceCommandType.PANIC_ALERT_TEST,
            CommandContext(RuntimeState.ACTIVE_DRIVE, speedMph = 55.0),
        )

        assertTrue(decision.allowed)
    }
}

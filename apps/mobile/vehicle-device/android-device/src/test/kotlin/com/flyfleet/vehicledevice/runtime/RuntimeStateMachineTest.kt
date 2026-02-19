package com.flyfleet.vehicledevice.runtime

import kotlin.test.Test
import kotlin.test.assertEquals

class RuntimeStateMachineTest {
    @Test
    fun `active drive transitions to degraded on sensor failure`() {
        val machine = RuntimeStateMachine()
        machine.transition(RuntimeSignal.STARTUP_COMPLETE)
        machine.transition(RuntimeSignal.IGNITION_ON)

        val next = machine.transition(RuntimeSignal.SENSOR_FAILURE)
        assertEquals(RuntimeState.DEGRADED, next)
    }

    @Test
    fun `idle transitions to depot mode when depot entered`() {
        val machine = RuntimeStateMachine()
        machine.transition(RuntimeSignal.STARTUP_COMPLETE)

        val next = machine.transition(RuntimeSignal.DEPOT_ENTERED)
        assertEquals(RuntimeState.DEPOT_MODE, next)
    }

    @Test
    fun `depot mode transitions to active drive on ignition on`() {
        val machine = RuntimeStateMachine()
        machine.transition(RuntimeSignal.STARTUP_COMPLETE)
        machine.transition(RuntimeSignal.DEPOT_ENTERED)

        val next = machine.transition(RuntimeSignal.IGNITION_ON)
        assertEquals(RuntimeState.ACTIVE_DRIVE, next)
    }

    @Test
    fun `degraded transitions to recovery when component recovers`() {
        val machine = RuntimeStateMachine()
        machine.transition(RuntimeSignal.STARTUP_COMPLETE)
        machine.transition(RuntimeSignal.IGNITION_ON)
        machine.transition(RuntimeSignal.SENSOR_FAILURE)

        val next = machine.transition(RuntimeSignal.COMPONENT_RECOVERED)
        assertEquals(RuntimeState.RECOVERY, next)
    }

    @Test
    fun `safe shutdown is terminal`() {
        val machine = RuntimeStateMachine()
        machine.transition(RuntimeSignal.SHUTDOWN_REQUESTED)

        val next = machine.transition(RuntimeSignal.STARTUP_COMPLETE)
        assertEquals(RuntimeState.SAFE_SHUTDOWN, next)
    }
}

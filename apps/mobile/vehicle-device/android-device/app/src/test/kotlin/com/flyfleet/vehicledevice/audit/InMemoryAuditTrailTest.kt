package com.flyfleet.vehicledevice.audit

import com.flyfleet.vehicledevice.command.CommandDecision
import com.flyfleet.vehicledevice.command.DeviceCommandType
import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.sync.Clock
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class InMemoryAuditTrailTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    @Test
    fun `records signed envelopes for command and operational signals`() {
        val clock = FakeClock(now = 1_000)
        val trail = InMemoryAuditTrail(
            deviceId = "dev-1",
            clock = clock,
            signer = HmacAuditSigner("secret".toByteArray()),
        )

        trail.recordCommandDecision(
            commandType = DeviceCommandType.FORCE_SYNC,
            speedMph = 0.0,
            decision = CommandDecision(allowed = true),
        )
        clock.now += 5
        trail.recordOperationalSignal(DeviceEventType.CAN_BUS_DEGRADED, payloadRef = "reason=low")

        val all = trail.snapshot()
        assertEquals(2, all.size)
        assertEquals(AuditKind.COMMAND_DECISION, all[0].kind)
        assertEquals(AuditKind.OPERATIONAL_SIGNAL, all[1].kind)
        assertTrue(trail.verifyAll())
    }

    @Test
    fun `tampered envelope fails verification`() {
        val clock = FakeClock(now = 1_000)
        val signer = HmacAuditSigner("secret".toByteArray())
        val trail = InMemoryAuditTrail(deviceId = "dev-1", clock = clock, signer = signer)

        trail.recordOperationalSignal(DeviceEventType.TWIN_DRIFT_DETECTED, payloadRef = "mismatch=a")
        val original = trail.snapshot().first()

        val tamperedCanonical = "${original.id}|${original.deviceId}|${original.kind.name}|${original.timestampMs}|payload=tampered"
        assertFalse(signer.verify(tamperedCanonical, original.signature))
    }
}

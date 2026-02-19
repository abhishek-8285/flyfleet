package com.flyfleet.vehicledevice.forensics

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class EvidenceHashChainTest {
    @Test
    fun `valid hash chain verifies successfully`() {
        val chain = EvidenceHashChain()
        chain.append(EvidenceRecord("ev-1", 1000, "panic_triggered", "a1"))
        chain.append(EvidenceRecord("ev-2", 1010, "harsh_brake_detected", "b2"))
        chain.append(EvidenceRecord("ev-3", 1020, "camera_frame_stall", "c3"))

        assertTrue(chain.verify())
    }

    @Test
    fun `tampered record fails verification`() {
        val chain = EvidenceHashChain()
        chain.append(EvidenceRecord("ev-1", 1000, "panic_triggered", "a1"))
        chain.append(EvidenceRecord("ev-2", 1010, "harsh_brake_detected", "b2"))

        val tampered = chain.snapshot().toMutableList()
        val second = tampered[1]
        tampered[1] = second.copy(
            record = second.record.copy(eventType = "speed_spike_detected"),
        )

        assertFalse(chain.verify(tampered))
    }

    @Test
    fun `broken previous hash link fails verification`() {
        val chain = EvidenceHashChain()
        chain.append(EvidenceRecord("ev-1", 1000, "panic_triggered", "a1"))
        chain.append(EvidenceRecord("ev-2", 1010, "harsh_brake_detected", "b2"))

        val broken = chain.snapshot().toMutableList()
        val second = broken[1]
        broken[1] = second.copy(previousHashHex = "deadbeef")

        assertFalse(chain.verify(broken))
    }
}

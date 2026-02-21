package com.flyfleet.vehicledevice.forensics

import com.flyfleet.vehicledevice.audit.AuditKind
import com.flyfleet.vehicledevice.audit.AuditEnvelope
import com.flyfleet.vehicledevice.audit.HmacAuditSigner
import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.events.UploadPriorityTier
import com.flyfleet.vehicledevice.sync.StoredEvent
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class IncidentBundleBuilderTest {
    private val signer = HmacAuditSigner("bundle-secret".toByteArray())

    @Test
    fun `builds reproducible signed bundle and verifies`() {
        val chain = EvidenceHashChain()
        val events = listOf(
            StoredEvent(1, "ev-1", DeviceEventType.PANIC_TRIGGERED, UploadPriorityTier.TIER1, 1000, "blob://1", true),
            StoredEvent(2, "ev-2", DeviceEventType.CAN_BUS_DEGRADED, UploadPriorityTier.TIER3, 1010, "blob://2", true),
        )
        val chainRecords = listOf(
            chain.append(EvidenceRecord("ev-1", 1000, "PANIC_TRIGGERED", "a1")),
            chain.append(EvidenceRecord("ev-2", 1010, "CAN_BUS_DEGRADED", "b2")),
        )
        val audits = listOf(
            AuditEnvelope(
                id = "au-1",
                deviceId = "dev-1",
                kind = AuditKind.COMMAND_DECISION,
                timestampMs = 1000,
                payload = "command=FORCE_SYNC|allowed=true",
                signature = "sig1",
            ),
        )

        val builder = IncidentBundleBuilder(signer)
        val bundle = builder.build("dev-1", "inc-1", events, chainRecords, audits)

        assertTrue(builder.verify(bundle))
        assertEquals(2, bundle.manifest.eventCount)
        assertEquals(2, bundle.manifest.chainCount)
        assertTrue(bundle.manifest.chainVerified)
    }

    @Test
    fun `tampered event payload breaks verification`() {
        val chain = EvidenceHashChain()
        val events = listOf(
            StoredEvent(1, "ev-1", DeviceEventType.PANIC_TRIGGERED, UploadPriorityTier.TIER1, 1000, "blob://1", true),
        )
        val chainRecords = listOf(
            chain.append(EvidenceRecord("ev-1", 1000, "PANIC_TRIGGERED", "a1")),
        )

        val builder = IncidentBundleBuilder(signer)
        val bundle = builder.build("dev-1", "inc-1", events, chainRecords, emptyList())
        val tampered = bundle.copy(
            events = listOf(bundle.events.first().copy(payloadRef = "blob://tampered")),
        )

        assertFalse(builder.verify(tampered))
    }

    @Test
    fun `broken chain in bundle fails verification`() {
        val chain = EvidenceHashChain()
        val events = listOf(
            StoredEvent(1, "ev-1", DeviceEventType.PANIC_TRIGGERED, UploadPriorityTier.TIER1, 1000, "blob://1", true),
        )
        val chainRecords = listOf(
            chain.append(EvidenceRecord("ev-1", 1000, "PANIC_TRIGGERED", "a1")),
        )

        val builder = IncidentBundleBuilder(signer)
        val bundle = builder.build("dev-1", "inc-1", events, chainRecords, emptyList())
        val broken = bundle.copy(
            chainRecords = listOf(bundle.chainRecords.first().copy(previousHashHex = "bad-link")),
        )

        assertFalse(builder.verify(broken))
    }
}

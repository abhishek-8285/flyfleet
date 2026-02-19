package com.flyfleet.vehicledevice.forensics

import com.flyfleet.vehicledevice.audit.AuditEnvelope
import com.flyfleet.vehicledevice.audit.AuditSigner
import com.flyfleet.vehicledevice.sync.StoredEvent
import java.security.MessageDigest

data class IncidentBundleManifest(
    val deviceId: String,
    val incidentId: String,
    val eventCount: Int,
    val chainCount: Int,
    val auditEnvelopeCount: Int,
    val chainVerified: Boolean,
    val eventsDigestHex: String,
    val chainDigestHex: String,
    val auditDigestHex: String,
)

data class IncidentBundle(
    val manifest: IncidentBundleManifest,
    val events: List<StoredEvent>,
    val chainRecords: List<ChainedEvidenceRecord>,
    val auditEnvelopes: List<AuditEnvelope>,
    val manifestSignature: String,
)

class IncidentBundleBuilder(
    private val signer: AuditSigner,
) {
    fun build(
        deviceId: String,
        incidentId: String,
        events: List<StoredEvent>,
        chainRecords: List<ChainedEvidenceRecord>,
        auditEnvelopes: List<AuditEnvelope>,
    ): IncidentBundle {
        val canonicalEvents = events
            .sortedBy { it.sequence }
            .joinToString("\n") {
                listOf(
                    it.sequence.toString(),
                    it.id,
                    it.eventType.name,
                    it.tier.name,
                    it.createdAtMs.toString(),
                    it.payloadRef,
                    it.uploaded.toString(),
                ).joinToString("|")
            }
        val canonicalChain = chainRecords.joinToString("\n") {
            listOf(
                it.record.evidenceId,
                it.record.timestampMs.toString(),
                it.record.eventType,
                it.record.payloadDigestHex,
                it.previousHashHex,
                it.chainHashHex,
            ).joinToString("|")
        }
        val canonicalAudit = auditEnvelopes.joinToString("\n") {
            listOf(
                it.id,
                it.deviceId,
                it.kind.name,
                it.timestampMs.toString(),
                it.payload,
                it.signature,
            ).joinToString("|")
        }

        val hashChain = EvidenceHashChain()
        val chainVerified = hashChain.verify(chainRecords)

        val manifest = IncidentBundleManifest(
            deviceId = deviceId,
            incidentId = incidentId,
            eventCount = events.size,
            chainCount = chainRecords.size,
            auditEnvelopeCount = auditEnvelopes.size,
            chainVerified = chainVerified,
            eventsDigestHex = sha256(canonicalEvents),
            chainDigestHex = sha256(canonicalChain),
            auditDigestHex = sha256(canonicalAudit),
        )

        val signature = signer.sign(canonicalManifest(manifest))
        return IncidentBundle(
            manifest = manifest,
            events = events,
            chainRecords = chainRecords,
            auditEnvelopes = auditEnvelopes,
            manifestSignature = signature,
        )
    }

    fun verify(bundle: IncidentBundle): Boolean {
        val rebuilt = build(
            deviceId = bundle.manifest.deviceId,
            incidentId = bundle.manifest.incidentId,
            events = bundle.events,
            chainRecords = bundle.chainRecords,
            auditEnvelopes = bundle.auditEnvelopes,
        )

        val manifestMatches = rebuilt.manifest == bundle.manifest
        val signatureMatches = signer.verify(canonicalManifest(bundle.manifest), bundle.manifestSignature)
        return manifestMatches && signatureMatches && bundle.manifest.chainVerified
    }

    private fun canonicalManifest(m: IncidentBundleManifest): String {
        return listOf(
            m.deviceId,
            m.incidentId,
            m.eventCount.toString(),
            m.chainCount.toString(),
            m.auditEnvelopeCount.toString(),
            m.chainVerified.toString(),
            m.eventsDigestHex,
            m.chainDigestHex,
            m.auditDigestHex,
        ).joinToString("|")
    }

    private fun sha256(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(input.toByteArray())
        return digest.joinToString(separator = "") { "%02x".format(it) }
    }
}

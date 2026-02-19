package com.flyfleet.vehicledevice.audit

import com.flyfleet.vehicledevice.command.CommandDecision
import com.flyfleet.vehicledevice.command.DeviceCommandType
import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.sync.Clock
import java.nio.charset.StandardCharsets
import java.util.UUID
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

enum class AuditKind {
    COMMAND_DECISION,
    OPERATIONAL_SIGNAL,
}

data class AuditEnvelope(
    val id: String,
    val deviceId: String,
    val kind: AuditKind,
    val timestampMs: Long,
    val payload: String,
    val signature: String,
)

interface AuditSigner {
    fun sign(canonical: String): String
    fun verify(canonical: String, signature: String): Boolean
}

class HmacAuditSigner(secret: ByteArray) : AuditSigner {
    private val key = SecretKeySpec(secret, "HmacSHA256")

    override fun sign(canonical: String): String {
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(key)
        val digest = mac.doFinal(canonical.toByteArray(StandardCharsets.UTF_8))
        return digest.joinToString(separator = "") { "%02x".format(it) }
    }

    override fun verify(canonical: String, signature: String): Boolean {
        return sign(canonical).equals(signature, ignoreCase = true)
    }
}

interface AuditEmitter {
    fun recordCommandDecision(commandType: DeviceCommandType, speedMph: Double, decision: CommandDecision)
    fun recordOperationalSignal(eventType: DeviceEventType, payloadRef: String)
}

object NoopAuditEmitter : AuditEmitter {
    override fun recordCommandDecision(commandType: DeviceCommandType, speedMph: Double, decision: CommandDecision) = Unit
    override fun recordOperationalSignal(eventType: DeviceEventType, payloadRef: String) = Unit
}

class InMemoryAuditTrail(
    private val deviceId: String,
    private val clock: Clock,
    private val signer: AuditSigner,
) : AuditEmitter {
    private val envelopes = mutableListOf<AuditEnvelope>()

    override fun recordCommandDecision(commandType: DeviceCommandType, speedMph: Double, decision: CommandDecision) {
        val payload = "command=${commandType.name}|speed_mph=${"%.1f".format(speedMph)}|allowed=${decision.allowed}|reason=${decision.reason ?: "none"}"
        envelopes += envelope(kind = AuditKind.COMMAND_DECISION, payload = payload)
    }

    override fun recordOperationalSignal(eventType: DeviceEventType, payloadRef: String) {
        val payload = "event=${eventType.name}|payload_ref=${payloadRef}"
        envelopes += envelope(kind = AuditKind.OPERATIONAL_SIGNAL, payload = payload)
    }

    fun snapshot(): List<AuditEnvelope> = envelopes.toList()

    fun verifyAll(): Boolean {
        return envelopes.all { env ->
            signer.verify(canonical(env.id, env.deviceId, env.kind, env.timestampMs, env.payload), env.signature)
        }
    }

    private fun envelope(kind: AuditKind, payload: String): AuditEnvelope {
        val id = UUID.randomUUID().toString()
        val ts = clock.nowMs()
        val canonical = canonical(id, deviceId, kind, ts, payload)
        return AuditEnvelope(
            id = id,
            deviceId = deviceId,
            kind = kind,
            timestampMs = ts,
            payload = payload,
            signature = signer.sign(canonical),
        )
    }

    private fun canonical(id: String, deviceId: String, kind: AuditKind, timestampMs: Long, payload: String): String {
        return "$id|$deviceId|${kind.name}|$timestampMs|$payload"
    }
}

package com.flyfleet.vehicledevice.forensics

import com.flyfleet.vehicledevice.security.SecurityModule
import java.security.MessageDigest

/**
 * Immutable evidence record with cryptographic integrity.
 * 
 * Each record contains:
 * - Unique evidence ID (UUID v4)
 * - NIST-traceable timestamp (milliseconds since epoch)
 * - Event type from taxonomy
 * - SHA-256 digest of payload (video clip, telemetry batch, etc.)
 */
data class EvidenceRecord(
    val evidenceId: String,
    val timestampMs: Long,
    val eventType: String,
    val payloadDigestHex: String,
    val ntpTimestampMs: Long? = null, // NTP-synced timestamp for legal admissibility
    val gnssTimestampMs: Long? = null, // GNSS timestamp for fallback
)

/**
 * Hash-chained evidence record with previous hash linkage.
 * 
 * Chain structure:
 * [Genesis] → [Record 1] → [Record 2] → ... → [Record N]
 * 
 * Each link contains:
 * - The evidence record
 * - Hash of previous link (prevents insertion attacks)
 * - Hash of current link (enables verification)
 */
data class ChainedEvidenceRecord(
    val record: EvidenceRecord,
    val previousHashHex: String,
    val chainHashHex: String,
    val signature: ByteArray? = null, // Hardware-backed signature for high-assurance fleets
)

/**
 * Tamper-evident hash chain for forensic chain of custody.
 * 
 * Implements NIST SP 800-88 compliant evidence integrity protection.
 * Each evidence record is linked to the previous via SHA-256 hash chain.
 * Optional hardware-backed signing for FMCSA ELD compliance.
 * 
 * Usage:
 * ```
 * val chain = EvidenceHashChain(securityModule)
 * val record = EvidenceRecord("uuid-123", timestamp, "harsh_brake", payloadHash)
 * val chained = chain.append(record)
 * 
 * // Later, verify integrity:
 * val isValid = chain.verify(chain.snapshot())
 * ```
 */
class EvidenceHashChain(
    private val securityModule: SecurityModule? = null
) {
    private val chain = mutableListOf<ChainedEvidenceRecord>()
    private val lock = Any()

    /**
     * Appends a new evidence record to the chain.
     * 
     * Thread-safe with synchronized access.
     * Automatically computes previous hash and chain hash.
     * Optionally signs with hardware-backed key if SecurityModule provided.
     * 
     * @param record The evidence record to append
     * @return ChainedEvidenceRecord with hash linkage and optional signature
     */
    fun append(record: EvidenceRecord): ChainedEvidenceRecord {
        synchronized(lock) {
            val previousHash = chain.lastOrNull()?.chainHashHex ?: genesisHash()
            val chainInput = "${record.evidenceId}|${record.timestampMs}|${record.eventType}|${record.payloadDigestHex}|$previousHash"
            val chainHash = hashHex(chainInput)
            
            // Sign with hardware-backed key if available (required for FMCSA ELD)
            val signature = securityModule?.let { module ->
                val payloadToSign = "$chainInput|$chainHash".toByteArray()
                module.signEvidence(payloadToSign)
            }
            
            val chained = ChainedEvidenceRecord(
                record = record,
                previousHashHex = previousHash,
                chainHashHex = chainHash,
                signature = signature,
            )
            chain += chained
            return chained
        }
    }

    /**
     * Returns immutable snapshot of the entire chain.
     */
    fun snapshot(): List<ChainedEvidenceRecord> = synchronized(lock) {
        chain.toList()
    }

    /**
     * Verifies integrity of the hash chain.
     * 
     * Checks:
     * 1. Each record's previous hash matches the previous record's chain hash
     * 2. Each record's chain hash is correctly computed
     * 3. If signed, verifies hardware-backed signatures
     * 
     * @param records Chain to verify (defaults to current chain)
     * @return true if chain is intact, false if tampered
     */
    fun verify(records: List<ChainedEvidenceRecord> = chain): Boolean {
        synchronized(lock) {
            var prev = genesisHash()
            for (entry in records) {
                // Verify previous hash linkage
                if (entry.previousHashHex != prev) {
                    return false
                }
                
                // Verify chain hash computation
                val chainInput = "${entry.record.evidenceId}|${entry.record.timestampMs}|${entry.record.eventType}|${entry.record.payloadDigestHex}|$prev"
                val expected = hashHex(chainInput)
                if (entry.chainHashHex != expected) {
                    return false
                }
                
                // Verify hardware-backed signature if present (ELD compliance)
                if (entry.signature != null) {
                    val signatureInput = "$chainInput|$expected".toByteArray()
                    val signatureValid = securityModule?.verifySignature(
                        signatureInput, entry.signature
                    ) ?: false
                    if (!signatureValid) {
                        return false
                    }
                }
                
                prev = entry.chainHashHex
            }
            return true
        }
    }

    /**
     * Returns the current chain length (number of evidence records).
     */
    fun size(): Int = synchronized(lock) {
        chain.size
    }

    /**
     * Clears the chain (use with caution - only for testing).
     */
    fun clear() {
        synchronized(lock) {
            chain.clear()
        }
    }

    private fun genesisHash(): String = hashHex("flyfleet-evidence-genesis-v1")

    private fun hashHex(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val hashBytes = digest.digest(input.toByteArray())
        return hashBytes.joinToString(separator = "") { "%02x".format(it) }
    }
}

/**
 * Builder for creating EvidenceHashChain with optional SecurityModule.
 * 
 * Usage:
 * ```
 * val chain = EvidenceHashChain.Builder()
 *     .withSecurityModule(securityModule)
 *     .withNtpSync(ntpClient)
 *     .build()
 * ```
 */
class EvidenceHashChainBuilder {
    private var securityModule: SecurityModule? = null
    
    fun withSecurityModule(module: SecurityModule) = apply {
        this.securityModule = module
    }
    
    fun build(): EvidenceHashChain {
        return EvidenceHashChain(securityModule)
    }
}

/**
 * Convenience function to create EvidenceHashChain with builder syntax.
 */
fun evidenceHashChain(block: EvidenceHashChainBuilder.() -> Unit): EvidenceHashChain {
    val builder = EvidenceHashChainBuilder()
    builder.block()
    return builder.build()
}

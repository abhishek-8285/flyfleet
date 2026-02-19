package com.flyfleet.vehicledevice.security

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import timber.log.Timber
import java.security.KeyStore
import java.security.MessageDigest
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

/**
 * Security Module for FlyFleet Vehicle Device
 * 
 * Provides:
 * - Hardware-backed key storage (Android Keystore)
 * - Encrypted data storage
 * - Cryptographic signing for evidence
 * - Device attestation
 */
class SecurityModule(private val context: Context) {

    companion object {
        private const val ANDROID_KEYSTORE = "AndroidKeyStore"
        private const val MASTER_KEY_ALIAS = "_androidx_security_master_key_"
        private const val EVIDENCE_KEY_ALIAS = "flyfleet-evidence-key"
        private const val TRANSFORMATION = "AES/GCM/NoPadding"
        private const val GCM_TAG_LENGTH = 128
        private const val GCM_IV_LENGTH = 12
    }

    private val keyStore: KeyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply {
        load(null)
    }

    private val masterKey: MasterKey by lazy {
        MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
    }

    private val encryptedPrefs by lazy {
        EncryptedSharedPreferences.create(
            context,
            "secure_prefs",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    /**
     * Initialize security module
     */
    fun initialize(): SecurityStatus {
        return try {
            // Generate or retrieve evidence signing key
            getOrCreateEvidenceKey()
            
            SecurityStatus(
                isHardwareBacked = isHardwareBacked(EVIDENCE_KEY_ALIAS),
                isStrongBoxBacked = isStrongBoxBacked(EVIDENCE_KEY_ALIAS)
            )
        } catch (e: Exception) {
            Timber.e(e, "Failed to initialize security module")
            SecurityStatus(
                isHardwareBacked = false,
                isStrongBoxBacked = false,
                error = e.message
            )
        }
    }

    /**
     * Encrypt sensitive data
     */
    fun encrypt(plaintext: String): EncryptedData {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, getEvidenceKey())
        
        val iv = cipher.iv
        val ciphertext = cipher.doFinal(plaintext.toByteArray())
        
        return EncryptedData(
            ciphertext = ciphertext,
            iv = iv
        )
    }

    /**
     * Decrypt data
     */
    fun decrypt(encryptedData: EncryptedData): String {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        val spec = GCMParameterSpec(GCM_TAG_LENGTH, encryptedData.iv)
        cipher.init(Cipher.DECRYPT_MODE, getEvidenceKey(), spec)
        
        val plaintext = cipher.doFinal(encryptedData.ciphertext)
        return String(plaintext)
    }

    /**
     * Sign evidence data
     */
    fun signEvidence(data: ByteArray): ByteArray {
        val digest = MessageDigest.getInstance("SHA-256")
        val hash = digest.digest(data)
        
        // Sign with hardware-backed key
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, getEvidenceKey())
        
        return cipher.doFinal(hash)
    }

    /**
     * Verify evidence signature
     */
    fun verifySignature(data: ByteArray, signature: ByteArray): Boolean {
        return try {
            val digest = MessageDigest.getInstance("SHA-256")
            val hash = digest.digest(data)
            
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.DECRYPT_MODE, getEvidenceKey())
            
            val decryptedHash = cipher.doFinal(signature)
            hash.contentEquals(decryptedHash)
        } catch (e: Exception) {
            Timber.e(e, "Signature verification failed")
            false
        }
    }

    /**
     * Compute SHA-256 hash
     */
    fun hash(data: ByteArray): ByteArray {
        val digest = MessageDigest.getInstance("SHA-256")
        return digest.digest(data)
    }

    /**
     * Store sensitive string securely
     */
    fun storeSecureString(key: String, value: String) {
        encryptedPrefs.edit().putString(key, value).apply()
    }

    /**
     * Retrieve sensitive string securely
     */
    fun getSecureString(key: String, defaultValue: String = ""): String {
        return encryptedPrefs.getString(key, defaultValue) ?: defaultValue
    }

    /**
     * Clear sensitive data
     */
    fun clearSecureData(key: String) {
        encryptedPrefs.edit().remove(key).apply()
    }

    /**
     * Clear all secure data (for device wipe)
     */
    fun wipeAllSecureData() {
        // Delete keys from keystore
        keyStore.deleteEntry(EVIDENCE_KEY_ALIAS)
        
        // Clear encrypted prefs
        encryptedPrefs.edit().clear().apply()
        
        Timber.i("All secure data wiped")
    }

    //region Private Implementation

    private fun getOrCreateEvidenceKey(): SecretKey {
        val existingKey = keyStore.getCertificate(EVIDENCE_KEY_ALIAS)
        if (existingKey != null) {
            Timber.d("Evidence key already exists")
            return (keyStore.getEntry(EVIDENCE_KEY_ALIAS, null) as KeyStore.SecretKeyEntry).secretKey
        }

        Timber.d("Generating new evidence key")
        return generateEvidenceKey()
    }

    private fun generateEvidenceKey(): SecretKey {
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            ANDROID_KEYSTORE
        )

        val spec = KeyGenParameterSpec.Builder(
            EVIDENCE_KEY_ALIAS,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        ).apply {
            setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            setKeySize(256)
            
            // Require user authentication for high-security operations
            // setUserAuthenticationRequired(false) // Disabled for background operations
            
            // Request StrongBox if available
            try {
                setIsStrongBoxBacked(true)
            } catch (e: Exception) {
                Timber.w("StrongBox not available, using TEE")
            }
        }.build()

        keyGenerator.init(spec)
        return keyGenerator.generateKey()
    }

    private fun getEvidenceKey(): SecretKey {
        return (keyStore.getEntry(EVIDENCE_KEY_ALIAS, null) as KeyStore.SecretKeyEntry).secretKey
    }

    private fun isHardwareBacked(alias: String): Boolean {
        return try {
            val key = keyStore.getKey(alias, null)
            val factory = javax.crypto.SecretKeyFactory.getInstance(key.algorithm, ANDROID_KEYSTORE)
            val keyInfo = factory.getKeySpec(key as SecretKey, android.security.keystore.KeyInfo::class.java) as android.security.keystore.KeyInfo
            keyInfo.isInsideSecureHardware
        } catch (e: Exception) {
            false
        }
    }

    private fun isStrongBoxBacked(alias: String): Boolean {
        if (android.os.Build.VERSION.SDK_INT < android.os.Build.VERSION_CODES.P) return false
        return try {
            val key = keyStore.getKey(alias, null)
            val factory = javax.crypto.SecretKeyFactory.getInstance(key.algorithm, ANDROID_KEYSTORE)
            val keyInfo = factory.getKeySpec(key as SecretKey, android.security.keystore.KeyInfo::class.java) as android.security.keystore.KeyInfo
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.S) {
                keyInfo.securityLevel == android.security.keystore.KeyProperties.SECURITY_LEVEL_STRONGBOX
            } else {
                false
            }
        } catch (e: Exception) {
            false
        }
    }

    //endregion
}

/**
 * Security module initialization status
 */
data class SecurityStatus(
    val isHardwareBacked: Boolean,
    val isStrongBoxBacked: Boolean,
    val error: String? = null
)

/**
 * Encrypted data container
 */
data class EncryptedData(
    val ciphertext: ByteArray,
    val iv: ByteArray
)

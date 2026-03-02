package com.fleet.driver.data.remote.api

import com.fleet.driver.data.remote.model.*
import com.google.gson.annotations.SerializedName
import retrofit2.Response
import retrofit2.http.*

/**
 * Authentication API Service Interface
 */
interface AuthService {
    
    /**
     * User login with email and password
     */
    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>
    
    /**
     * Refresh access token using refresh token
     */
    @POST("auth/refresh")
    suspend fun refreshToken(@Body request: RefreshTokenRequest): Response<LoginResponse>
    
    /**
     * Logout current user session
     */
    @POST("auth/logout")
    suspend fun logout(): Response<Unit>
    
    /**
     * Enroll driver using enrollment code
     */
    @POST("auth/enroll")
    suspend fun enrollDriver(@Body request: EnrollDriverRequest): Response<EnrollDriverResponse>
    
    /**
     * Register device for push notifications
     */
    @POST("auth/device/register")
    suspend fun registerDevice(@Body request: RegisterDeviceRequest): Response<RegisterDeviceResponse>
    
    /**
     * Enable MFA for user
     */
    @POST("auth/mfa/enable")
    suspend fun enableMFA(@Body request: MFAEnableRequest): Response<MFAEnableResponse>
    
    /**
     * Verify MFA code
     */
    @POST("auth/mfa/verify")
    suspend fun verifyMFA(@Body request: MFAVerifyRequest): Response<LoginResponse>
}

// MFA Models
data class MFAEnableRequest(
    @SerializedName("method")
    val method: String // "totp" or "sms"
)

data class MFAEnableResponse(
    @SerializedName("secret")
    val secret: String, // QR code data for TOTP
    
    @SerializedName("backup_codes")
    val backupCodes: List<String>
)

data class MFAVerifyRequest(
    @SerializedName("code")
    val code: String,
    
    @SerializedName("method")
    val method: String
)
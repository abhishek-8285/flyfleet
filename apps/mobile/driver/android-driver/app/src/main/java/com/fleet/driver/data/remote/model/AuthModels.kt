package com.fleet.driver.data.remote.model

import com.google.gson.annotations.SerializedName

/**
 * Authentication API Models
 */

data class LoginRequest(
    @SerializedName("email")
    val email: String,
    
    @SerializedName("password")
    val password: String,
    
    @SerializedName("device_id")
    val deviceId: String? = null
)

data class LoginResponse(
    @SerializedName("access_token")
    val accessToken: String,
    
    @SerializedName("refresh_token")
    val refreshToken: String,
    
    @SerializedName("user")
    val user: UserResponse
)

data class UserResponse(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("email")
    val email: String,
    
    @SerializedName("role")
    val role: String,
    
    @SerializedName("permissions")
    val permissions: List<String>,
    
    @SerializedName("org_id")
    val orgId: String,
    
    @SerializedName("first_name")
    val firstName: String? = null,
    
    @SerializedName("last_name")
    val lastName: String? = null
)

data class RefreshTokenRequest(
    @SerializedName("refresh_token")
    val refreshToken: String
)

data class EnrollDriverRequest(
    @SerializedName("enrollment_code")
    val enrollmentCode: String,
    
    @SerializedName("device_id")
    val deviceId: String,
    
    @SerializedName("fcm_token")
    val fcmToken: String? = null
)

data class EnrollDriverResponse(
    @SerializedName("success")
    val success: Boolean,
    
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("access_token")
    val accessToken: String,
    
    @SerializedName("refresh_token")
    val refreshToken: String
)

data class RegisterDeviceRequest(
    @SerializedName("device_id")
    val deviceId: String,
    
    @SerializedName("fcm_token")
    val fcmToken: String? = null,
    
    @SerializedName("platform")
    val platform: String = "android"
)

data class RegisterDeviceResponse(
    @SerializedName("device_registered")
    val deviceRegistered: Boolean
)

// Error response models
data class ErrorResponse(
    @SerializedName("error")
    val error: String,
    
    @SerializedName("message")
    val message: String,
    
    @SerializedName("attempts_remaining")
    val attemptsRemaining: Int? = null
)

// Authentication state
sealed class AuthState {
    object Unauthenticated : AuthState()
    object Authenticating : AuthState()
    data class Authenticated(val user: UserResponse) : AuthState()
    data class Error(val message: String) : AuthState()
}

// Enrollment state
sealed class EnrollmentState {
    object Idle : EnrollmentState()
    object Enrolling : EnrollmentState()
    data class Success(val driverId: String) : EnrollmentState()
    data class Error(val message: String) : EnrollmentState()
}
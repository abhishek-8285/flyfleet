package com.fleet.driver.data.repository

import android.content.SharedPreferences
import com.fleet.driver.data.remote.api.AuthService
import com.fleet.driver.data.remote.model.*
import com.fleet.driver.util.DeviceUtils
import com.fleet.driver.util.Resource
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Authentication Repository
 * Handles all authentication-related data operations
 */
@Singleton
class AuthRepository @Inject constructor(
    private val authService: AuthService,
    private val sharedPreferences: SharedPreferences,
    private val deviceUtils: DeviceUtils
) {
    companion object {
        private const val PREF_ACCESS_TOKEN = "access_token"
        private const val PREF_REFRESH_TOKEN = "refresh_token"
        private const val PREF_USER_ID = "user_id"
        private const val PREF_USER_EMAIL = "user_email"
        private const val PREF_ORG_ID = "org_id"
        private const val PREF_DEVICE_REGISTERED = "device_registered"
    }

    /**
     * Get stored access token
     */
    fun getAccessToken(): String? {
        return sharedPreferences.getString(PREF_ACCESS_TOKEN, null)
    }

    /**
     * Get stored refresh token
     */
    fun getRefreshToken(): String? {
        return sharedPreferences.getString(PREF_REFRESH_TOKEN, null)
    }

    /**
     * Check if user is authenticated
     */
    fun isAuthenticated(): Boolean {
        return !getAccessToken().isNullOrEmpty()
    }

    /**
     * Get current user information
     */
    fun getCurrentUser(): UserResponse? {
        val userId = sharedPreferences.getString(PREF_USER_ID, null) ?: return null
        val email = sharedPreferences.getString(PREF_USER_EMAIL, null) ?: return null
        val orgId = sharedPreferences.getString(PREF_ORG_ID, null) ?: return null
        
        return UserResponse(
            id = userId,
            email = email,
            role = "driver", // Default role
            permissions = emptyList(),
            orgId = orgId
        )
    }

    /**
     * User login
     */
    suspend fun login(email: String, password: String): Resource<LoginResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val deviceId = deviceUtils.getDeviceId()
                val request = LoginRequest(email, password, deviceId)
                val response = authService.login(request)
                
                if (response.isSuccessful) {
                    response.body()?.let { loginResponse ->
                        saveTokens(loginResponse)
                        saveUserInfo(loginResponse.user)
                        Resource.Success(loginResponse)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Login failed")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }

    /**
     * Driver enrollment using enrollment code
     */
    suspend fun enrollDriver(enrollmentCode: String): Resource<EnrollDriverResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val deviceId = deviceUtils.getDeviceId()
                val fcmToken = deviceUtils.getFCMToken()
                val request = EnrollDriverRequest(enrollmentCode, deviceId, fcmToken)
                val response = authService.enrollDriver(request)
                
                if (response.isSuccessful) {
                    response.body()?.let { enrollResponse ->
                        if (enrollResponse.success) {
                            saveTokensFromEnrollment(enrollResponse)
                            sharedPreferences.edit().putString(PREF_USER_ID, enrollResponse.driverId).apply()
                            Resource.Success(enrollResponse)
                        } else {
                            Resource.Error("Enrollment failed")
                        }
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Enrollment failed")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }

    /**
     * Refresh access token
     */
    suspend fun refreshToken(): Resource<LoginResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val refreshToken = getRefreshToken() ?: return@withContext Resource.Error("No refresh token")
                val request = RefreshTokenRequest(refreshToken)
                val response = authService.refreshToken(request)
                
                if (response.isSuccessful) {
                    response.body()?.let { loginResponse ->
                        saveTokens(loginResponse)
                        Resource.Success(loginResponse)
                    } ?: Resource.Error("Empty response body")
                } else {
                    // Clear tokens on refresh failure
                    clearTokens()
                    Resource.Error("Session expired")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }

    /**
     * Logout user
     */
    suspend fun logout(): Resource<Unit> {
        return withContext(Dispatchers.IO) {
            try {
                val response = authService.logout()
                
                // Clear local tokens regardless of API response
                clearTokens()
                
                if (response.isSuccessful) {
                    Resource.Success(Unit)
                } else {
                    Resource.Success(Unit) // Still consider logout successful locally
                }
            } catch (e: Exception) {
                // Even if API call fails, clear local tokens
                clearTokens()
                Resource.Success(Unit)
            }
        }
    }

    /**
     * Register device for push notifications
     */
    suspend fun registerDevice(): Resource<RegisterDeviceResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val deviceId = deviceUtils.getDeviceId()
                val fcmToken = deviceUtils.getFCMToken()
                val request = RegisterDeviceRequest(deviceId, fcmToken)
                val response = authService.registerDevice(request)
                
                if (response.isSuccessful) {
                    response.body()?.let { registerResponse ->
                        if (registerResponse.deviceRegistered) {
                            sharedPreferences.edit().putBoolean(PREF_DEVICE_REGISTERED, true).apply()
                        }
                        Resource.Success(registerResponse)
                    } ?: Resource.Error("Empty response body")
                } else {
                    val errorBody = response.errorBody()?.string()
                    Resource.Error(errorBody ?: "Device registration failed")
                }
            } catch (e: Exception) {
                Resource.Error(e.message ?: "Network error")
            }
        }
    }

    /**
     * Save authentication tokens
     */
    private fun saveTokens(response: LoginResponse) {
        with(sharedPreferences.edit()) {
            putString(PREF_ACCESS_TOKEN, response.accessToken)
            putString(PREF_REFRESH_TOKEN, response.refreshToken)
            apply()
        }
    }

    /**
     * Save tokens from enrollment response
     */
    private fun saveTokensFromEnrollment(response: EnrollDriverResponse) {
        with(sharedPreferences.edit()) {
            putString(PREF_ACCESS_TOKEN, response.accessToken)
            putString(PREF_REFRESH_TOKEN, response.refreshToken)
            apply()
        }
    }

    /**
     * Save user information
     */
    private fun saveUserInfo(user: UserResponse) {
        with(sharedPreferences.edit()) {
            putString(PREF_USER_ID, user.id)
            putString(PREF_USER_EMAIL, user.email)
            putString(PREF_ORG_ID, user.orgId)
            apply()
        }
    }

    /**
     * Clear all authentication tokens and user info
     */
    private fun clearTokens() {
        with(sharedPreferences.edit()) {
            remove(PREF_ACCESS_TOKEN)
            remove(PREF_REFRESH_TOKEN)
            remove(PREF_USER_ID)
            remove(PREF_USER_EMAIL)
            remove(PREF_ORG_ID)
            apply()
        }
    }

    /**
     * Check if device is registered for push notifications
     */
    fun isDeviceRegistered(): Boolean {
        return sharedPreferences.getBoolean(PREF_DEVICE_REGISTERED, false)
    }
}
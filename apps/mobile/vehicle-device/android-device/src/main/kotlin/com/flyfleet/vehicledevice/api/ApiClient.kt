package com.flyfleet.vehicledevice.api

import com.flyfleet.vehicledevice.BuildConfig
import com.flyfleet.vehicledevice.models.*
import com.flyfleet.vehicledevice.security.SecurityModule
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.serialization.json.Json
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.HttpException
import retrofit2.Retrofit
import retrofit2.http.*
import timber.log.Timber
import java.util.concurrent.TimeUnit

/**
 * API Client for FlyFleet Backend Services
 * 
 * Handles:
 * - Device registration and authentication
 * - Telemetry batch uploads
 * - Media upload management
 * - Driver HOS tracking
 * - Remote command retrieval
 */
class ApiClient(
    baseUrl: String,
    jwtSecret: String,
    private val securityModule: SecurityModule
) {
    private val retrofit: Retrofit
    private val apiService: DeviceApiService
    internal var authToken: String? = null

    init {
        // Create HTTP client with logging and timeouts
        val httpClient = OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .writeTimeout(120, TimeUnit.SECONDS) // Longer for uploads
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = if (BuildConfig.DEBUG) {
                    HttpLoggingInterceptor.Level.BODY
                } else {
                    HttpLoggingInterceptor.Level.NONE
                }
            })
            .addInterceptor(AuthInterceptor(this))
            .build()

        // Create Retrofit instance
        val contentType = "application/json".toMediaType()
        val json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
        }
        
        retrofit = Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(httpClient)
            .addConverterFactory(
                @Suppress("DEPRECATION")
                json.asConverterFactory(contentType)
            )
            .build()

        apiService = retrofit.create(DeviceApiService::class.java)
    }

    /**
     * Set authentication token (called after device registration)
     */
    fun setAuthToken(token: String) {
        authToken = token
    }

    /**
     * Clear authentication token (called on logout/wipe)
     */
    fun clearAuthToken() {
        authToken = null
    }

    //region Device Management

    /**
     * Register device with backend
     */
    suspend fun registerDevice(request: DeviceRegisterRequest): Result<DeviceRegisterResponse> {
        return try {
            val response = apiService.registerDevice(request)
            if (response.isSuccessful && response.body() != null) {
                // Store auth token for future requests
                setAuthToken(response.body()!!.jwtToken)
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Device registration failed")
            Result.failure(e)
        }
    }

    /**
     * Send device heartbeat
     */
    suspend fun sendHeartbeat(request: HeartbeatRequest): Result<HeartbeatResponse> {
        return try {
            val response = apiService.sendHeartbeat(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Heartbeat failed")
            Result.failure(e)
        }
    }

    /**
     * Fetch device configuration
     */
    suspend fun getConfig(deviceId: String): Result<ConfigResponse> {
        return try {
            val response = apiService.getConfig(deviceId)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Get config failed")
            Result.failure(e)
        }
    }

    /**
     * Get pending remote commands
     */
    suspend fun getPendingCommands(deviceId: String, lastCommandId: String? = null): Result<List<Command>> {
        return try {
            val response = apiService.getPendingCommands(deviceId, lastCommandId)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.success(emptyList())
            }
        } catch (e: Exception) {
            Timber.e(e, "Get commands failed")
            Result.failure(e)
        }
    }

    //endregion

    //region Telemetry

    /**
     * Upload batch of telemetry data
     */
    suspend fun uploadTelemetryBatch(request: TelemetryBatchRequest): Result<TelemetryResponse> {
        return try {
            val response = apiService.uploadTelemetryBatch(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Telemetry upload failed")
            Result.failure(e)
        }
    }

    //endregion

    //region Media Management

    /**
     * Get storage location for media clip
     */
    suspend fun getMediaStorageLocation(request: UploadUrlRequest): Result<UploadUrlResponse> {
        return try {
            val response = apiService.getMediaStorageLocation(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Get media storage failed")
            Result.failure(e)
        }
    }

    /**
     * Confirm media upload completion
     */
    suspend fun confirmMediaUpload(request: ConfirmUploadRequest): Result<ConfirmUploadResponse> {
        return try {
            val response = apiService.confirmMediaUpload(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Confirm media upload failed")
            Result.failure(e)
        }
    }

    //endregion

    //region Forensics

    /**
     * Attest evidence integrity
     */
    suspend fun attestEvidence(request: AttestationRequest): Result<AttestationResponse> {
        return try {
            val response = apiService.attestEvidence(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Evidence attestation failed")
            Result.failure(e)
        }
    }

    //endregion

    //region Driver Management

    /**
     * Driver login with HOS tracking
     */
    suspend fun driverLogin(request: DriverLoginRequest): Result<DriverLoginResponse> {
        return try {
            val response = apiService.driverLogin(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Driver login failed")
            Result.failure(e)
        }
    }

    /**
     * Driver logout
     */
    suspend fun driverLogout(request: DriverLogoutRequest): Result<DriverLogoutResponse> {
        return try {
            val response = apiService.driverLogout(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "Driver logout failed")
            Result.failure(e)
        }
    }

    /**
     * Update driver HOS duty status
     */
    suspend fun updateDriverHOS(request: HOSUpdateRequest): Result<HOSUpdateResponse> {
        return try {
            val response = apiService.updateDriverHOS(request)
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                Result.failure(HttpException(response))
            }
        } catch (e: Exception) {
            Timber.e(e, "HOS update failed")
            Result.failure(e)
        }
    }

    //endregion
}

/**
 * Retrofit API service interface
 */
interface DeviceApiService {
    
    @POST("/device/register")
    suspend fun registerDevice(@Body request: DeviceRegisterRequest): retrofit2.Response<DeviceRegisterResponse>
    
    @POST("/device/heartbeat")
    suspend fun sendHeartbeat(@Body request: HeartbeatRequest): retrofit2.Response<HeartbeatResponse>
    
    @GET("/device/config")
    suspend fun getConfig(@Query("device_id") deviceId: String): retrofit2.Response<ConfigResponse>
    
    @GET("/device/commands")
    suspend fun getPendingCommands(
        @Query("device_id") deviceId: String,
        @Query("last_command_id") lastCommandId: String? = null
    ): retrofit2.Response<List<Command>>
    
    @POST("/telemetry/batch")
    suspend fun uploadTelemetryBatch(@Body request: TelemetryBatchRequest): retrofit2.Response<TelemetryResponse>
    
    @POST("/media/upload-url")
    suspend fun getMediaStorageLocation(@Body request: UploadUrlRequest): retrofit2.Response<UploadUrlResponse>
    
    @POST("/media/confirm")
    suspend fun confirmMediaUpload(@Body request: ConfirmUploadRequest): retrofit2.Response<ConfirmUploadResponse>
    
    @POST("/forensics/attest")
    suspend fun attestEvidence(@Body request: AttestationRequest): retrofit2.Response<AttestationResponse>
    
    @POST("/driver/login")
    suspend fun driverLogin(@Body request: DriverLoginRequest): retrofit2.Response<DriverLoginResponse>
    
    @POST("/driver/logout")
    suspend fun driverLogout(@Body request: DriverLogoutRequest): retrofit2.Response<DriverLogoutResponse>
    
    @POST("/driver/hos")
    suspend fun updateDriverHOS(@Body request: HOSUpdateRequest): retrofit2.Response<HOSUpdateResponse>
}

/**
 * Authentication interceptor - adds JWT token to requests
 */
class AuthInterceptor(private val apiClient: ApiClient) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): okhttp3.Response {
        val originalRequest = chain.request()
        
        // Skip auth for registration endpoint
        if (originalRequest.url.encodedPath == "/device/register") {
            return chain.proceed(originalRequest)
        }
        
        // Add auth token if available
        val authToken = apiClient.authToken
        val newRequest = if (authToken != null) {
            originalRequest.newBuilder()
                .header("Authorization", "Bearer $authToken")
                .build()
        } else {
            originalRequest
        }
        
        return chain.proceed(newRequest)
    }
}

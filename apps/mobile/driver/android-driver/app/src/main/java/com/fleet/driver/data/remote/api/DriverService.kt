package com.fleet.driver.data.remote.api

import com.fleet.driver.data.remote.model.*
import com.google.gson.annotations.SerializedName
import retrofit2.Response
import retrofit2.http.*

/**
 * Driver API Service Interface
 */
interface DriverService {
    
    /**
     * Get current driver profile
     */
    @GET("drivers/profile")
    suspend fun getDriverProfile(): Response<DriverProfile>
    
    /**
     * Update driver profile
     */
    @PUT("drivers/profile")
    suspend fun updateDriverProfile(@Body request: UpdateDriverProfileRequest): Response<DriverProfile>
    
    /**
     * Get driver analytics/performance metrics
     */
    @GET("drivers/analytics")
    suspend fun getDriverAnalytics(): Response<DriverAnalytics>
    
    /**
     * Get driver behavior events
     */
    @GET("drivers/behavior/events")
    suspend fun getBehaviorEvents(
        @Query("from_date") fromDate: String? = null,
        @Query("to_date") toDate: String? = null,
        @Query("limit") limit: Int = 50
    ): Response<List<BehaviorEvent>>
    
    /**
     * Get attendance records
     */
    @GET("drivers/attendance")
    suspend fun getAttendanceRecords(
        @Query("from_date") fromDate: String? = null,
        @Query("to_date") toDate: String? = null
    ): Response<List<AttendanceRecord>>
    
    /**
     * Mark attendance
     */
    @POST("drivers/attendance")
    suspend fun markAttendance(@Body request: UpdateAttendanceRequest): Response<AttendanceRecord>
    
    /**
     * Get work schedules
     */
    @GET("drivers/schedules")
    suspend fun getWorkSchedules(): Response<List<WorkSchedule>>
    
    /**
     * Set work schedule
     */
    @POST("drivers/schedules")
    suspend fun setWorkSchedule(@Body request: SetScheduleRequest): Response<WorkSchedule>
    
    /**
     * Update work schedule
     */
    @PUT("drivers/schedules/{id}")
    suspend fun updateWorkSchedule(
        @Path("id") scheduleId: String,
        @Body request: SetScheduleRequest
    ): Response<WorkSchedule>
    
    /**
     * Delete work schedule
     */
    @DELETE("drivers/schedules/{id}")
    suspend fun deleteWorkSchedule(@Path("id") scheduleId: String): Response<Unit>
    
    /**
     * Upload driver document
     */
    @POST("drivers/documents")
    suspend fun uploadDocument(@Body request: UploadDriverDocumentRequest): Response<DriverDocument>
    
    /**
     * Get driver documents
     */
    @GET("drivers/documents")
    suspend fun getDocuments(): Response<List<DriverDocument>>
    
    /**
     * Verify driver document
     */
    @POST("drivers/documents/{id}/verify")
    suspend fun verifyDocument(@Path("id") documentId: String): Response<DriverDocument>
    
    /**
     * Get driver performance summary
     */
    @GET("drivers/performance")
    suspend fun getPerformanceSummary(): Response<DriverPerformanceSummary>
    
    /**
     * Get nearby facilities (fuel stations, rest areas, etc.)
     */
    @GET("drivers/nearby-facilities")
    suspend fun getNearbyFacilities(
        @Query("latitude") latitude: Double,
        @Query("longitude") longitude: Double,
        @Query("radius_km") radiusKm: Double = 10.0,
        @Query("types") types: List<String>? = null // fuel_station, rest_area, hospital, restaurant
    ): Response<List<NearbyFacility>>
    
    /**
     * Report driver incident
     */
    @POST("drivers/incidents")
    suspend fun reportIncident(@Body request: DriverIncidentRequest): Response<DriverIncidentResponse>
}

// Document Upload Model
data class UploadDriverDocumentRequest(
    @SerializedName("document_type")
    val documentType: String,
    
    @SerializedName("document_number")
    val documentNumber: String,
    
    @SerializedName("issuing_authority")
    val issuingAuthority: String,
    
    @SerializedName("issue_date")
    val issueDate: String,
    
    @SerializedName("expiry_date")
    val expiryDate: String,
    
    @SerializedName("file_url")
    val fileUrl: String
)

// Nearby Facility Model
data class NearbyFacility(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("name")
    val name: String,
    
    @SerializedName("type")
    val type: String, // fuel_station, rest_area, hospital, restaurant
    
    @SerializedName("location")
    val location: Location,
    
    @SerializedName("distance_km")
    val distanceKm: Double,
    
    @SerializedName("rating")
    val rating: Double? = null,
    
    @SerializedName("phone")
    val phone: String? = null,
    
    @SerializedName("opening_hours")
    val openingHours: String? = null,
    
    @SerializedName("services")
    val services: List<String> = emptyList()
)

// Driver Incident Models
data class DriverIncidentRequest(
    @SerializedName("type")
    val type: String, // accident, violation, vehicle_damage, health_issue
    
    @SerializedName("description")
    val description: String,
    
    @SerializedName("severity")
    val severity: String, // low, medium, high, critical
    
    @SerializedName("location")
    val location: Location? = null,
    
    @SerializedName("photos")
    val photos: List<String>? = null,
    
    @SerializedName("vehicle_id")
    val vehicleId: String? = null
)

data class DriverIncidentResponse(
    @SerializedName("incident_id")
    val incidentId: String,
    
    @SerializedName("status")
    val status: String, // reported, under_review, resolved, rejected
    
    @SerializedName("reported_at")
    val reportedAt: String,
    
    @SerializedName("assigned_to")
    val assignedTo: String? = null
)

// Performance Summary
data class DriverPerformanceSummary(
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("name")
    val name: String,
    
    @SerializedName("current_rating")
    val currentRating: Double,
    
    @SerializedName("behavior_score")
    val behaviorScore: Int,
    
    @SerializedName("total_trips")
    val totalTrips: Int,
    
    @SerializedName("completion_rate")
    val completionRate: Double,
    
    @SerializedName("average_trip_distance")
    val averageTripDistance: Double,
    
    @SerializedName("average_trip_duration")
    val averageTripDuration: Double,
    
    @SerializedName("last_active")
    val lastActive: String,
    
    @SerializedName("status")
    val status: String,
    
    @SerializedName("vehicle_assigned")
    val vehicleAssigned: Boolean,
    
    @SerializedName("license_valid")
    val licenseValid: Boolean
)
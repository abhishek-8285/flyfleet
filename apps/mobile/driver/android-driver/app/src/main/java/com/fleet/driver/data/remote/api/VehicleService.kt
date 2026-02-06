package com.fleet.driver.data.remote.api

import com.fleet.driver.data.remote.model.*
import retrofit2.Response
import retrofit2.http.*

/**
 * Vehicle API Service Interface
 */
interface VehicleService {
    
    /**
     * Get list of vehicles with filtering
     */
    @GET("vehicles")
    suspend fun getVehicles(
        @Query("status") status: String? = null,
        @Query("type") type: String? = null,
        @Query("search") search: String? = null,
        @Query("page") page: Int = 1,
        @Query("page_size") pageSize: Int = 20
    ): Response<VehicleListResponse>
    
    /**
     * Get specific vehicle by ID
     */
    @GET("vehicles/{id}")
    suspend fun getVehicle(@Path("id") vehicleId: String): Response<Vehicle>
    
    /**
     * Get vehicles assigned to current driver
     */
    @GET("vehicles/my")
    suspend fun getMyVehicles(): Response<List<Vehicle>>
    
    /**
     * Assign vehicle to driver
     */
    @POST("vehicles/{id}/assign")
    suspend fun assignVehicle(
        @Path("id") vehicleId: String,
        @Body request: AssignVehicleRequest
    ): Response<VehicleAssignmentResponse>
    
    /**
     * Unassign vehicle from driver
     */
    @POST("vehicles/{id}/unassign")
    suspend fun unassignVehicle(@Path("id") vehicleId: String): Response<Unit>
    
    /**
     * Update vehicle status
     */
    @PUT("vehicles/{id}/status")
    suspend fun updateVehicleStatus(
        @Path("id") vehicleId: String,
        @Body request: VehicleStatusUpdateRequest
    ): Response<Vehicle>
    
    /**
     * Get vehicle telemetry data
     */
    @GET("vehicles/{id}/telemetry")
    suspend fun getVehicleTelemetry(
        @Path("id") vehicleId: String,
        @Query("from") from: String? = null,
        @Query("to") to: String? = null
    ): Response<List<VehicleTelemetry>>
    
    /**
     * Get vehicle assignment history
     */
    @GET("vehicles/{id}/assignments")
    suspend fun getVehicleAssignments(@Path("id") vehicleId: String): Response<List<VehicleAssignment>>
    
    /**
     * Report vehicle issue/maintenance need
     */
    @POST("vehicles/{id}/report-issue")
    suspend fun reportVehicleIssue(
        @Path("id") vehicleId: String,
        @Body request: VehicleIssueRequest
    ): Response<VehicleIssueResponse>
    
    /**
     * Get vehicle documents
     */
    @GET("vehicles/{id}/documents")
    suspend fun getVehicleDocuments(@Path("id") vehicleId: String): Response<List<VehicleDocument>>
    
    /**
     * Upload vehicle document
     */
    @POST("vehicles/{id}/documents")
    suspend fun uploadVehicleDocument(
        @Path("id") vehicleId: String,
        @Body request: UploadDocumentRequest
    ): Response<VehicleDocument>
}

// Vehicle Issue Models
data class VehicleIssueRequest(
    @SerializedName("title")
    val title: String,
    
    @SerializedName("description")
    val description: String,
    
    @SerializedName("priority")
    val priority: String, // low, medium, high, critical
    
    @SerializedName("category")
    val category: String, // mechanical, electrical, body, tires, other
    
    @SerializedName("images")
    val images: List<String>? = null // image URLs
)

data class VehicleIssueResponse(
    @SerializedName("issue_id")
    val issueId: String,
    
    @SerializedName("status")
    val status: String, // reported, in_progress, resolved, rejected
    
    @SerializedName("reported_at")
    val reportedAt: String,
    
    @SerializedName("assigned_to")
    val assignedTo: String? = null
)

// Vehicle Document Models
data class VehicleDocument(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("document_type")
    val documentType: String, // rc, insurance, fitness, permit, pollution
    
    @SerializedName("document_number")
    val documentNumber: String,
    
    @SerializedName("issue_date")
    val issueDate: String,
    
    @SerializedName("expiry_date")
    val expiryDate: String,
    
    @SerializedName("file_url")
    val fileUrl: String,
    
    @SerializedName("verified")
    val verified: Boolean,
    
    @SerializedName("created_at")
    val createdAt: String
)

data class UploadDocumentRequest(
    @SerializedName("document_type")
    val documentType: String,
    
    @SerializedName("document_number")
    val documentNumber: String,
    
    @SerializedName("issue_date")
    val issueDate: String,
    
    @SerializedName("expiry_date")
    val expiryDate: String,
    
    @SerializedName("file_url")
    val fileUrl: String
)
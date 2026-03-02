package com.fleet.driver.data.remote.model

import com.google.gson.annotations.SerializedName

/**
 * Vehicle API Models
 */

data class Vehicle(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("org_id")
    val orgId: String,
    
    @SerializedName("registration_number")
    val registrationNumber: String,
    
    @SerializedName("make")
    val make: String,
    
    @SerializedName("model")
    val model: String,
    
    @SerializedName("year")
    val year: Int,
    
    @SerializedName("type")
    val type: String, // truck, van, car, motorcycle
    
    @SerializedName("status")
    val status: String, // available, in_use, maintenance, retired
    
    @SerializedName("current_driver_id")
    val currentDriverId: String? = null,
    
    @SerializedName("current_location")
    val currentLocation: Location? = null,
    
    @SerializedName("fuel_type")
    val fuelType: String, // petrol, diesel, electric, hybrid
    
    @SerializedName("fuel_capacity")
    val fuelCapacity: Double, // liters
    
    @SerializedName("mileage")
    val mileage: Double, // km
    
    @SerializedName("last_service_date")
    val lastServiceDate: String? = null,
    
    @SerializedName("next_service_due")
    val nextServiceDue: String? = null,
    
    @SerializedName("insurance_expiry")
    val insuranceExpiry: String? = null,
    
    @SerializedName("fitness_expiry")
    val fitnessExpiry: String? = null,
    
    @SerializedName("metadata")
    val metadata: Map<String, Any>? = null,
    
    @SerializedName("created_at")
    val createdAt: String,
    
    @SerializedName("updated_at")
    val updatedAt: String
)

data class Location(
    @SerializedName("latitude")
    val latitude: Double,
    
    @SerializedName("longitude")
    val longitude: Double,
    
    @SerializedName("accuracy")
    val accuracy: Float? = null,
    
    @SerializedName("timestamp")
    val timestamp: String? = null,
    
    @SerializedName("address")
    val address: String? = null
)

data class VehicleListResponse(
    @SerializedName("vehicles")
    val vehicles: List<Vehicle>,
    
    @SerializedName("total")
    val total: Int,
    
    @SerializedName("page")
    val page: Int,
    
    @SerializedName("page_size")
    val pageSize: Int
)

data class AssignVehicleRequest(
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("driver_id")
    val driverId: String
)

data class VehicleStatusUpdateRequest(
    @SerializedName("status")
    val status: String // available, in_use, maintenance, retired
)

data class VehicleTelemetry(
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("timestamp")
    val timestamp: String,
    
    @SerializedName("location")
    val location: Location,
    
    @SerializedName("speed")
    val speed: Double, // km/h
    
    @SerializedName("heading")
    val heading: Double? = null, // degrees
    
    @SerializedName("fuel_level")
    val fuelLevel: Double? = null, // percentage
    
    @SerializedName("odometer")
    val odometer: Double? = null, // km
    
    @SerializedName("engine_status")
    val engineStatus: Boolean, // running/stopped
    
    @SerializedName("battery_level")
    val batteryLevel: Double? = null, // percentage (for electric vehicles)
    
    @SerializedName("temperature")
    val temperature: Double? = null, // engine temperature
    
    @SerializedName("metadata")
    val metadata: Map<String, Any>? = null
)

// Vehicle assignment response
data class VehicleAssignmentResponse(
    @SerializedName("success")
    val success: Boolean,
    
    @SerializedName("assignment_id")
    val assignmentId: String,
    
    @SerializedName("vehicle")
    val vehicle: Vehicle,
    
    @SerializedName("assigned_at")
    val assignedAt: String
)

// Vehicle filters
data class VehicleFilter(
    val status: String? = null,
    val type: String? = null,
    val search: String? = null,
    val page: Int = 1,
    val pageSize: Int = 20
)

// Vehicle assignment history
data class VehicleAssignment(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("assigned_by")
    val assignedBy: String,
    
    @SerializedName("assigned_at")
    val assignedAt: String,
    
    @SerializedName("unassigned_at")
    val unassignedAt: String? = null,
    
    @SerializedName("notes")
    val notes: String? = null
)

// Vehicle issue reporting
data class VehicleIssueRequest(
    @SerializedName("title") val title: String,
    @SerializedName("description") val description: String,
    @SerializedName("priority") val priority: String, // low, medium, high, critical
    @SerializedName("category") val category: String, // mechanical, electrical, body, tires, other
    @SerializedName("images") val images: List<String>? = null // image URLs
)

data class VehicleIssueResponse(
    @SerializedName("issue_id") val issueId: String,
    @SerializedName("status") val status: String, // reported, in_progress, resolved, rejected
    @SerializedName("reported_at") val reportedAt: String,
    @SerializedName("assigned_to") val assignedTo: String? = null
)

// Vehicle documentation
data class VehicleDocument(
    @SerializedName("id") val id: String,
    @SerializedName("vehicle_id") val vehicleId: String,
    @SerializedName("document_type") val documentType: String, // rc, insurance, fitness, permit, pollution
    @SerializedName("document_number") val documentNumber: String,
    @SerializedName("issue_date") val issueDate: String,
    @SerializedName("expiry_date") val expiryDate: String,
    @SerializedName("file_url") val fileUrl: String,
    @SerializedName("verified") val verified: Boolean,
    @SerializedName("created_at") val createdAt: String
)

data class UploadDocumentRequest(
    @SerializedName("document_type") val documentType: String,
    @SerializedName("document_number") val documentNumber: String,
    @SerializedName("issue_date") val issueDate: String,
    @SerializedName("expiry_date") val expiryDate: String,
    @SerializedName("file_url") val fileUrl: String
)
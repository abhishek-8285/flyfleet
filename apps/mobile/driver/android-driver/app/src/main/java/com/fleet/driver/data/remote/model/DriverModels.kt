package com.fleet.driver.data.remote.model

import com.google.gson.annotations.SerializedName

/**
 * Driver Profile API Models
 */

data class DriverProfile(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("org_id")
    val orgId: String,
    
    @SerializedName("first_name")
    val firstName: String,
    
    @SerializedName("last_name")
    val lastName: String,
    
    @SerializedName("full_name")
    val fullName: String,
    
    @SerializedName("email")
    val email: String,
    
    @SerializedName("phone")
    val phone: String,
    
    @SerializedName("license_number")
    val licenseNumber: String,
    
    @SerializedName("license_type")
    val licenseType: String, // LMV, HMV, Transport
    
    @SerializedName("license_expiry")
    val licenseExpiry: String,
    
    @SerializedName("license_authority")
    val licenseAuthority: String,
    
    @SerializedName("date_of_birth")
    val dateOfBirth: String,
    
    @SerializedName("address")
    val address: Address,
    
    @SerializedName("emergency_contact")
    val emergencyContact: EmergencyContact,
    
    @SerializedName("status")
    val status: String, // active, inactive, suspended, resigned
    
    @SerializedName("current_vehicle_id")
    val currentVehicleId: String? = null,
    
    @SerializedName("current_trip_id")
    val currentTripId: String? = null,
    
    @SerializedName("joining_date")
    val joiningDate: String,
    
    @SerializedName("experience_years")
    val experienceYears: Int,
    
    @SerializedName("rating")
    val rating: Double, // 0-5 scale
    
    @SerializedName("total_trips")
    val totalTrips: Int,
    
    @SerializedName("total_distance_km")
    val totalDistanceKm: Double,
    
    @SerializedName("total_hours_driven")
    val totalHoursDriven: Double,
    
    @SerializedName("behavior_score")
    val behaviorScore: Int, // 0-100 scale
    
    @SerializedName("documents")
    val documents: List<DriverDocument> = emptyList(),
    
    @SerializedName("attendance_records")
    val attendanceRecords: List<AttendanceRecord> = emptyList(),
    
    @SerializedName("schedules")
    val schedules: List<WorkSchedule> = emptyList(),
    
    @SerializedName("metadata")
    val metadata: Map<String, Any>? = null,
    
    @SerializedName("created_at")
    val createdAt: String,
    
    @SerializedName("updated_at")
    val updatedAt: String
)

data class Address(
    @SerializedName("street")
    val street: String,
    
    @SerializedName("city")
    val city: String,
    
    @SerializedName("state")
    val state: String,
    
    @SerializedName("pincode")
    val pincode: String,
    
    @SerializedName("country")
    val country: String,
    
    @SerializedName("latitude")
    val latitude: Double? = null,
    
    @SerializedName("longitude")
    val longitude: Double? = null
)

data class EmergencyContact(
    @SerializedName("name")
    val name: String,
    
    @SerializedName("phone")
    val phone: String,
    
    @SerializedName("relation")
    val relation: String
)

data class DriverDocument(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("document_type")
    val documentType: String, // license, aadhaar, pan, medical, training
    
    @SerializedName("document_number")
    val documentNumber: String,
    
    @SerializedName("issuing_authority")
    val issuingAuthority: String,
    
    @SerializedName("issue_date")
    val issueDate: String,
    
    @SerializedName("expiry_date")
    val expiryDate: String,
    
    @SerializedName("file_url")
    val fileUrl: String,
    
    @SerializedName("verified")
    val verified: Boolean,
    
    @SerializedName("verified_by")
    val verifiedBy: String? = null,
    
    @SerializedName("verified_at")
    val verifiedAt: String? = null,
    
    @SerializedName("created_at")
    val createdAt: String
)

data class AttendanceRecord(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("date")
    val date: String,
    
    @SerializedName("status")
    val status: String, // present, absent, leave, holiday
    
    @SerializedName("check_in_time")
    val checkInTime: String? = null,
    
    @SerializedName("check_out_time")
    val checkOutTime: String? = null,
    
    @SerializedName("remarks")
    val remarks: String? = null,
    
    @SerializedName("created_at")
    val createdAt: String
)

data class WorkSchedule(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("day_of_week")
    val dayOfWeek: Int, // 0-6 (Sunday-Saturday)
    
    @SerializedName("start_time")
    val startTime: String, // HH:MM format
    
    @SerializedName("end_time")
    val endTime: String, // HH:MM format
    
    @SerializedName("is_working")
    val isWorking: Boolean,
    
    @SerializedName("created_at")
    val createdAt: String,
    
    @SerializedName("updated_at")
    val updatedAt: String
)

// Driver Analytics
data class DriverAnalytics(
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("total_trips")
    val totalTrips: Int,
    
    @SerializedName("completed_trips")
    val completedTrips: Int,
    
    @SerializedName("cancelled_trips")
    val cancelledTrips: Int,
    
    @SerializedName("total_distance_km")
    val totalDistanceKm: Double,
    
    @SerializedName("total_hours_driven")
    val totalHoursDriven: Double,
    
    @SerializedName("average_rating")
    val averageRating: Double,
    
    @SerializedName("current_behavior_score")
    val currentBehaviorScore: Int,
    
    @SerializedName("attendance_rate")
    val attendanceRate: Double,
    
    @SerializedName("on_time_rate")
    val onTimeRate: Double,
    
    @SerializedName("fuel_efficiency")
    val fuelEfficiency: Double, // km/liter
    
    @SerializedName("speeding_incidents")
    val speedingIncidents: Int,
    
    @SerializedName("harsh_braking_count")
    val harshBrakingCount: Int,
    
    @SerializedName("idling_time_hours")
    val idlingTimeHours: Double,
    
    @SerializedName("completion_rate")
    val completionRate: Double
)

// Behavior Event
data class BehaviorEvent(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("event_type")
    val eventType: String, // speeding, harsh_braking, harsh_acceleration, phone_usage, seatbelt_violation
    
    @SerializedName("timestamp")
    val timestamp: String,
    
    @SerializedName("severity")
    val severity: String, // low, medium, high, critical
    
    @SerializedName("value")
    val value: Double, // speed, g-force, etc.
    
    @SerializedName("location")
    val location: Location? = null,
    
    @SerializedName("trip_id")
    val tripId: String? = null,
    
    @SerializedName("created_at")
    val createdAt: String
)

// Update Requests
data class UpdateDriverProfileRequest(
    @SerializedName("first_name")
    val firstName: String? = null,
    
    @SerializedName("last_name")
    val lastName: String? = null,
    
    @SerializedName("phone")
    val phone: String? = null,
    
    @SerializedName("address")
    val address: Address? = null,
    
    @SerializedName("emergency_contact")
    val emergencyContact: EmergencyContact? = null,
    
    @SerializedName("metadata")
    val metadata: Map<String, Any>? = null
)

data class UpdateAttendanceRequest(
    @SerializedName("date")
    val date: String,
    
    @SerializedName("status")
    val status: String, // present, absent, leave, holiday
    
    @SerializedName("check_in_time")
    val checkInTime: String? = null,
    
    @SerializedName("check_out_time")
    val checkOutTime: String? = null,
    
    @SerializedName("remarks")
    val remarks: String? = null
)

data class SetScheduleRequest(
    @SerializedName("day_of_week")
    val dayOfWeek: Int, // 0-6
    
    @SerializedName("start_time")
    val startTime: String, // HH:MM
    
    @SerializedName("end_time")
    val endTime: String, // HH:MM
    
    @SerializedName("is_working")
    val isWorking: Boolean
)
package com.fleet.driver.data.remote.model

import com.google.gson.annotations.SerializedName

/**
 * Trip API Models
 */

data class Trip(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("org_id")
    val orgId: String,
    
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("driver_id")
    val driverId: String,
    
    @SerializedName("status")
    val status: String, // scheduled, in_progress, completed, cancelled
    
    @SerializedName("origin")
    val origin: TripLocation,
    
    @SerializedName("destination")
    val destination: TripLocation,
    
    @SerializedName("waypoints")
    val waypoints: List<TripLocation> = emptyList(),
    
    @SerializedName("route")
    val route: Route? = null,
    
    @SerializedName("scheduled_start")
    val scheduledStart: String,
    
    @SerializedName("scheduled_end")
    val scheduledEnd: String? = null,
    
    @SerializedName("actual_start")
    val actualStart: String? = null,
    
    @SerializedName("actual_end")
    val actualEnd: String? = null,
    
    @SerializedName("distance_km")
    val distanceKm: Double? = null,
    
    @SerializedName("duration_minutes")
    val durationMinutes: Int? = null,
    
    @SerializedName("fuel_consumed")
    val fuelConsumed: Double? = null,
    
    @SerializedName("cargo_details")
    val cargoDetails: CargoDetails? = null,
    
    @SerializedName("notes")
    val notes: String? = null,
    
    @SerializedName("created_at")
    val createdAt: String,
    
    @SerializedName("updated_at")
    val updatedAt: String
)

data class TripLocation(
    @SerializedName("name")
    val name: String,
    
    @SerializedName("address")
    val address: String,
    
    @SerializedName("latitude")
    val latitude: Double,
    
    @SerializedName("longitude")
    val longitude: Double,
    
    @SerializedName("contact_name")
    val contactName: String? = null,
    
    @SerializedName("contact_phone")
    val contactPhone: String? = null,
    
    @SerializedName("arrival_time")
    val arrivalTime: String? = null,
    
    @SerializedName("departure_time")
    val departureTime: String? = null
)

data class Route(
    @SerializedName("coordinates")
    val coordinates: List<List<Double>>, // [[lng, lat], [lng, lat], ...]
    
    @SerializedName("distance_km")
    val distanceKm: Double,
    
    @SerializedName("duration_minutes")
    val durationMinutes: Int,
    
    @SerializedName("polyline")
    val polyline: String? = null
)

data class CargoDetails(
    @SerializedName("type")
    val type: String,
    
    @SerializedName("weight_kg")
    val weightKg: Double? = null,
    
    @SerializedName("quantity")
    val quantity: Int? = null,
    
    @SerializedName("description")
    val description: String? = null,
    
    @SerializedName("value")
    val value: Double? = null,
    
    @SerializedName("special_instructions")
    val specialInstructions: String? = null
)

data class TripListResponse(
    @SerializedName("trips")
    val trips: List<Trip>,
    
    @SerializedName("total")
    val total: Int,
    
    @SerializedName("page")
    val page: Int,
    
    @SerializedName("page_size")
    val pageSize: Int
)

data class StartTripRequest(
    @SerializedName("trip_id")
    val tripId: String,
    
    @SerializedName("location")
    val location: Location,
    
    @SerializedName("odometer_reading")
    val odometerReading: Double? = null
)

data class CompleteTripRequest(
    @SerializedName("trip_id")
    val tripId: String,
    
    @SerializedName("location")
    val location: Location,
    
    @SerializedName("odometer_reading")
    val odometerReading: Double? = null,
    
    @SerializedName("notes")
    val notes: String? = null
)

data class UpdateTripStatusRequest(
    @SerializedName("status")
    val status: String,
    
    @SerializedName("notes")
    val notes: String? = null
)

data class ReachWaypointRequest(
    @SerializedName("waypoint_index")
    val waypointIndex: Int,
    
    @SerializedName("location")
    val location: Location,
    
    @SerializedName("notes")
    val notes: String? = null
)

data class TripSummary(
    @SerializedName("trip_id")
    val tripId: String,
    
    @SerializedName("total_distance_km")
    val totalDistanceKm: Double,
    
    @SerializedName("total_duration_minutes")
    val totalDurationMinutes: Int,
    
    @SerializedName("fuel_consumed")
    val fuelConsumed: Double,
    
    @SerializedName("average_speed")
    val averageSpeed: Double,
    
    @SerializedName("idle_time_minutes")
    val idleTimeMinutes: Int,
    
    @SerializedName("stops_count")
    val stopsCount: Int,
    
    @SerializedName("harsh_events")
    val harshEvents: HarshEvents,
    
    @SerializedName("route_deviation_count")
    val routeDeviationCount: Int
)

data class HarshEvents(
    @SerializedName("accelerations")
    val accelerations: Int,
    
    @SerializedName("brakings")
    val brakings: Int,
    
    @SerializedName("cornerings")
    val cornerings: Int,
    
    @SerializedName("speeding_incidents")
    val speedingIncidents: Int
)

data class CreateTripRequest(
    @SerializedName("vehicle_id")
    val vehicleId: String,
    
    @SerializedName("origin")
    val origin: TripLocation,
    
    @SerializedName("destination")
    val destination: TripLocation,
    
    @SerializedName("waypoints")
    val waypoints: List<TripLocation> = emptyList(),
    
    @SerializedName("scheduled_start")
    val scheduledStart: String,
    
    @SerializedName("cargo_details")
    val cargoDetails: CargoDetails? = null,
    
    @SerializedName("notes")
    val notes: String? = null
)
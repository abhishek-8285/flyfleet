package com.fleet.driver.data.remote.api

import com.fleet.driver.data.remote.model.*
import retrofit2.Response
import retrofit2.http.*

interface TripService {
    
    @GET("trips")
    suspend fun getTrips(
        @Query("status") status: String? = null,
        @Query("from_date") fromDate: String? = null,
        @Query("to_date") toDate: String? = null,
        @Query("page") page: Int = 1,
        @Query("page_size") pageSize: Int = 20
    ): Response<TripListResponse>
    
    @GET("trips/{id}")
    suspend fun getTrip(@Path("id") tripId: String): Response<Trip>
    
    @GET("trips/my/active")
    suspend fun getActiveTrip(): Response<Trip>
    
    @POST("trips")
    suspend fun createTrip(@Body request: CreateTripRequest): Response<Trip>
    
    @POST("trips/{id}/start")
    suspend fun startTrip(
        @Path("id") tripId: String,
        @Body request: StartTripRequest
    ): Response<Trip>
    
    @POST("trips/{id}/complete")
    suspend fun completeTrip(
        @Path("id") tripId: String,
        @Body request: CompleteTripRequest
    ): Response<Trip>
    
    @PUT("trips/{id}/status")
    suspend fun updateTripStatus(
        @Path("id") tripId: String,
        @Body request: UpdateTripStatusRequest
    ): Response<Trip>
    
    @POST("trips/{id}/waypoint")
    suspend fun reachWaypoint(
        @Path("id") tripId: String,
        @Body request: ReachWaypointRequest
    ): Response<Trip>
    
    @GET("trips/{id}/summary")
    suspend fun getTripSummary(@Path("id") tripId: String): Response<TripSummary>
    
    @GET("trips/{id}/route")
    suspend fun getTripRoute(@Path("id") tripId: String): Response<Route>
}
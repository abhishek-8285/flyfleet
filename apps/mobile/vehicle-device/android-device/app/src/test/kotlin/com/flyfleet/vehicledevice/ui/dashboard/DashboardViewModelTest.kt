package com.flyfleet.vehicledevice.ui.dashboard

import com.flyfleet.vehicledevice.domain.telemetry.TelemetryStats
import com.flyfleet.vehicledevice.domain.telemetry.TrackingState
import com.flyfleet.vehicledevice.domain.telemetry.VehicleTelemetryService
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class DashboardViewModelTest {

    private lateinit var viewModel: DashboardViewModel
    private val mockTelemetryService = mockk<VehicleTelemetryService>(relaxed = true)
    
    private val trackingStateFlow = MutableStateFlow<TrackingState>(TrackingState.Idle)
    private val telemetryStatsFlow = MutableStateFlow(TelemetryStats())

    @Before
    fun setup() {
        every { mockTelemetryService.trackingState } returns trackingStateFlow
        every { mockTelemetryService.telemetryStats } returns telemetryStatsFlow
        
        viewModel = DashboardViewModel(mockTelemetryService)
    }

    @Test
    fun `startTracking calls service`() = runTest {
        viewModel.startTracking()
        verify { mockTelemetryService.startTracking() }
    }

    @Test
    fun `stopTracking calls service`() = runTest {
        viewModel.stopTracking()
        verify { mockTelemetryService.stopTracking() }
    }

    @Test
    fun `states are correctly exposed from service`() = runTest {
        val stats = TelemetryStats(pointsCollected = 100)
        telemetryStatsFlow.value = stats
        trackingStateFlow.value = TrackingState.Tracking
        
        assertEquals(stats, viewModel.telemetryStats.value)
        assertEquals(TrackingState.Tracking, viewModel.trackingState.value)
    }
}

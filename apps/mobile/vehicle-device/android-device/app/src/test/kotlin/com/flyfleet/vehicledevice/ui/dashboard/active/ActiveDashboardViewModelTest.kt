package com.flyfleet.vehicledevice.ui.dashboard.active

import com.flyfleet.vehicledevice.config.DeviceConfig
import com.flyfleet.vehicledevice.data.api.DeviceApiService
import com.flyfleet.vehicledevice.data.db.DashboardDao
import com.flyfleet.vehicledevice.data.security.SecurityModule
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import com.flyfleet.vehicledevice.domain.model.DriverStatus
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test

@ExperimentalCoroutinesApi
class ActiveDashboardViewModelTest {

    private lateinit var viewModel: ActiveDashboardViewModel
    private lateinit var fakeTelemetry: FakeTelemetryService
    private lateinit var mockDeviceApi: DeviceApiService
    private lateinit var mockDeviceConfig: DeviceConfig
    private lateinit var mockSecurityModule: SecurityModule
    private lateinit var mockDashboardDao: DashboardDao

    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        fakeTelemetry = FakeTelemetryService()
        mockDeviceApi = mockk(relaxed = true)
        mockDeviceConfig = mockk(relaxed = true)
        mockSecurityModule = mockk(relaxed = true)
        mockDashboardDao = mockk(relaxed = true)

        viewModel =
                ActiveDashboardViewModel(
                        fakeTelemetry,
                        mockDeviceApi,
                        mockDeviceConfig,
                        mockSecurityModule,
                        mockDashboardDao
                ).apply { isTesting = true }
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial mode is Normal`() = runTest {
        advanceUntilIdle()
        assertEquals(DashboardMode.Normal, viewModel.state.value.mode)
    }

    @Test
    fun `test speed guard - speed greater than 5 locks features`() = runTest {
        fakeTelemetry.simulateSpeed(10.0)
        advanceUntilIdle()
        assertEquals(DashboardMode.FeatureLocked, viewModel.state.value.mode)
    }

    @Test
    fun `test storage guard - storage less than 5 percent triggers failure`() = runTest {
        fakeTelemetry.simulateStorage(4)
        advanceUntilIdle()
        assertEquals(DashboardMode.StorageFailure, viewModel.state.value.mode)
    }

    @Test
    fun `test time drift guard - drift greater than 10 mins triggers failure`() = runTest {
        fakeTelemetry.simulateTimeDrift(11 * 60 * 1000L)
        advanceUntilIdle()
        assertEquals(DashboardMode.TimeDrift, viewModel.state.value.mode)
    }

    @Test
    fun `test panic - panic always allowed and overrides other modes`() = runTest {
        // First enter FeatureLocked
        fakeTelemetry.simulateSpeed(60.0)
        advanceUntilIdle()
        assertEquals(DashboardMode.FeatureLocked, viewModel.state.value.mode)

        // Trigger Panic
        viewModel.onEvent(DashboardEvent.TriggerPanic)
        advanceUntilIdle()
        assertEquals(DashboardMode.PanicSent, viewModel.state.value.mode)
    }

    @Test
    fun `test ignition off - ignition off triggers mode`() = runTest {
        fakeTelemetry.simulateIgnition(false)
        advanceUntilIdle()
        assertEquals(DashboardMode.IgnitionOff, viewModel.state.value.mode)
    }

    @Test
    fun `test tamper detected - highest priority`() = runTest {
        fakeTelemetry.simulateTamper(true)
        fakeTelemetry.simulateSpeed(10.0)
        fakeTelemetry.simulateStorage(4)
        fakeTelemetry.simulateTimeDrift(20 * 60 * 1000L)
        advanceUntilIdle()
        assertEquals(DashboardMode.TamperDetected, viewModel.state.value.mode)
    }

    @Test
    fun `test session invalidated - flags cleared`() = runTest {
        viewModel.onEvent(DashboardEvent.OnSessionInvalidated)
        advanceUntilIdle()
        assertEquals(DashboardMode.SessionInvalidated, viewModel.state.value.mode)
    }

    @Test
    fun `test obd disconnected - mode change`() = runTest {
        fakeTelemetry.simulateObd(false)
        advanceUntilIdle()
        assertEquals(DashboardMode.ObdDisconnected, viewModel.state.value.mode)
    }

    @Test
    fun `test change status blocked while driving`() = runTest {
        // Set speed > 5
        fakeTelemetry.simulateSpeed(10.0)
        advanceUntilIdle()

        val initialStatus = viewModel.state.value.currentStatus
        viewModel.onEvent(DashboardEvent.ChangeStatus(DriverStatus.ON_DUTY))
        advanceUntilIdle()

        // Should NOT change
        assertEquals(initialStatus, viewModel.state.value.currentStatus)
    }
}

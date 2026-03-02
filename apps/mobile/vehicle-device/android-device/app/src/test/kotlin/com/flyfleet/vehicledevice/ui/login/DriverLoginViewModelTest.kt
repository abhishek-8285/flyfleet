package com.flyfleet.vehicledevice.ui.login

import android.app.Application
import com.flyfleet.vehicledevice.core.analytics.PostHogAnalytics
import com.flyfleet.vehicledevice.domain.model.DriverInfo
import com.flyfleet.vehicledevice.domain.model.RosterDriver
import com.flyfleet.vehicledevice.domain.repository.DriverAuthRepository
import com.flyfleet.vehicledevice.domain.service.DeviceStatusMonitor
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import app.cash.turbine.test
import kotlinx.coroutines.test.*
import org.junit.Before
import org.junit.After
import org.junit.Assert.*
import org.junit.runner.RunWith
import org.junit.Test
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.shadows.ShadowLog

@ExperimentalCoroutinesApi
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [31])
class DriverLoginViewModelTest {

    private lateinit var viewModel: DriverLoginViewModel
    private val mockApplication = mockk<Application>(relaxed = true)
    // private val mockAuthRepository = mockk<DriverAuthRepository>(relaxed = true) // Original mock
    private val mockAnalytics = mockk<PostHogAnalytics>(relaxed = true)
    private val mockDeviceMonitor = mockk<DeviceStatusMonitor>(relaxed = true)
    private val mockSecurityRepository = mockk<com.flyfleet.vehicledevice.domain.repository.SecurityRepository>(relaxed = true)
    private val mockErrorMapper = mockk<com.flyfleet.vehicledevice.data.util.ErrorMapper>(relaxed = true)
    private val fakeTelemetry = FakeTelemetryService()
    private val testDispatcher = UnconfinedTestDispatcher()

    private val onlineFlow = MutableStateFlow(true)
    private val batteryFlow = MutableStateFlow(80)

    private class FakeDriverAuthRepository : com.flyfleet.vehicledevice.domain.repository.DriverAuthRepository {
        var rosterResult: Result<List<RosterDriver>> = Result.success(emptyList())
        var loginResult: Result<DriverInfo> = Result.failure(Exception("Not set"))
        var conflictResult: Result<Boolean> = Result.success(false)
        var sessionResult: Result<String> = Result.success("MOCK-SESSION")

        override suspend fun getDriverRoster(vehicleId: String) = rosterResult
        override suspend fun pinLogin(deviceId: String, vehicleId: String, driverId: String, pin: String, gpsLat: Double?, gpsLon: Double?) = loginResult
        override suspend fun rfidLogin(deviceId: String, vehicleId: String, rfidTag: String, gpsLat: Double?, gpsLon: Double?) = loginResult
        override suspend fun checkActiveSession(driverId: String) = conflictResult
        override suspend fun createSession(driverId: String, vehicleId: String, deviceId: String) = sessionResult
        override suspend fun logout(sessionId: String) = Result.success(Unit)
    }

    private class FakePreferenceRepository : com.flyfleet.vehicledevice.domain.repository.PreferenceRepository {
        override fun getPreferences(driverId: String) = MutableStateFlow(null)
        override suspend fun savePreferences(preferences: com.flyfleet.vehicledevice.domain.model.UserPreference) {}
        override suspend fun fetchAndSyncPreferences(driverId: String) = Result.success(com.flyfleet.vehicledevice.domain.model.UserPreference(driverId, "en", "light", true))
    }

    private val fakeAuthRepository = FakeDriverAuthRepository()
    private val fakePreferenceRepository = FakePreferenceRepository()
    
    // We can use these interfaces to match what the ViewModel expects
    private val mockAuthRepository: DriverAuthRepository get() = fakeAuthRepository
    private val mockPreferenceRepository: com.flyfleet.vehicledevice.domain.repository.PreferenceRepository get() = fakePreferenceRepository

    @Before
    fun setup() {
        ShadowLog.stream = System.out
        Dispatchers.setMain(testDispatcher)
        
        // Use fake result in setup
        fakeAuthRepository.rosterResult = Result.success(listOf<RosterDriver>())
        
        every { mockDeviceMonitor.isOnline } returns onlineFlow
        every { mockDeviceMonitor.batteryLevel } returns batteryFlow
        every { mockDeviceMonitor.isLte } returns MutableStateFlow(true)
        every { mockDeviceMonitor.currentTimeUtc } returns MutableStateFlow("10:00 UTC")
        every { mockSecurityRepository.getSecureString(any(), any()) } returns "TEST-DEVICE"
        // coEvery { mockAuthRepository.getDriverRoster(any()) } returns Result.success(listOf<RosterDriver>()) // Replaced by fakeAuthRepository

        viewModel = DriverLoginViewModel(
            mockApplication,
            mockAuthRepository,
            mockAnalytics,
            mockDeviceMonitor,
            fakeTelemetry,
            mockPreferenceRepository,
            mockSecurityRepository,
            mockErrorMapper,
            testDispatcher
        )
        
        // Ensure init block launches run
        testDispatcher.scheduler.advanceUntilIdle()
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state is IDLE_RFID_WAIT`() = runTest {
        viewModel.uiState.test {
            assertTrue(awaitItem() is DriverLoginUiState.Idle)
        }
    }

    @Test
    fun `onSwitchToPin changes screen state`() = runTest {
        viewModel.uiState.test {
            assertTrue(awaitItem() is DriverLoginUiState.Idle) 
            viewModel.onSwitchToPin()
            assertTrue(awaitItem() is DriverLoginUiState.DriverSelection)
        }
    }

    @Test
    fun `onDriverSelected transitions to PIN entry`() = runTest {
        viewModel.uiState.test {
            assertTrue(awaitItem() is DriverLoginUiState.Idle)
            val driver = RosterDriver("DRV-001", "John Doe", "JD")
            viewModel.onDriverSelected(driver)
            val state = awaitItem()
            assertTrue(state is DriverLoginUiState.PinEntry)
            assertEquals(driver, (state as DriverLoginUiState.PinEntry).selectedDriver)
        }
    }

    @Test
    fun testOnDigitPressed() = runTest {
        try {
            val driver = RosterDriver("DRV-001", "John Doe", "JD")
            val driverInfo = DriverInfo(driverId = "DRV-001", name = "John Doe", org = "CMP-1", hosStatus = "V-101")
            fakeAuthRepository.loginResult = Result.success(driverInfo)
            fakeAuthRepository.conflictResult = Result.success(false)
            fakeAuthRepository.sessionResult = Result.success("SESS-123")

            viewModel.uiState.test {
                val s1 = awaitItem()
                assertTrue("Initial state should be Idle, was $s1", s1 is DriverLoginUiState.Idle)
                
                viewModel.onDriverSelected(driver)
                runCurrent()
                val s2 = awaitItem()
                assertTrue("State after driver selection should be PinEntry, was $s2", s2 is DriverLoginUiState.PinEntry)
                
                // Press 1-5
                repeat(5) { i ->
                    viewModel.onDigitPressed(i + 1)
                    val s = awaitItem()
                    assertTrue("State after ${i+1} digits should be PinEntry, was $s", s is DriverLoginUiState.PinEntry)
                }
                
                // 6th digit starts login
                viewModel.onDigitPressed(6)
                
                // 1st item after 6th digit is the updated PinEntry (6 digits)
                val s6 = awaitItem()
                assertTrue("State after 6th digit (1st emiss) should be PinEntry, was $s6", s6 is DriverLoginUiState.PinEntry)

                // 2nd item could be Loading or jump directly to Success due to StateFlow conflation
                val sNext = awaitItem()
                if (sNext is DriverLoginUiState.Loading) {
                    val sFinal = awaitItem()
                    assertTrue("Final state should be Success, was $sFinal", sFinal is DriverLoginUiState.Success)
                } else {
                    assertTrue("Final state should be Success, was $sNext", sNext is DriverLoginUiState.Success)
                }
            }
        } catch (e: Throwable) {
            println("CAPTURED EXCEPTION in testOnDigitPressed: ${e.javaClass.canonicalName}: ${e.message}")
            e.printStackTrace()
            throw e
        }
    }

    @Test
    fun `speed above threshold blocks login`() = runTest {
        viewModel.uiState.test {
            assertTrue(awaitItem() is DriverLoginUiState.Idle)
            viewModel.onVehicleSpeedChanged(10.0)
            assertTrue(awaitItem() is DriverLoginUiState.Idle) // From updateStatus
            assertTrue(awaitItem() is DriverLoginUiState.Blocked)
            
            viewModel.onVehicleSpeedChanged(2.0)
            assertTrue(awaitItem() is DriverLoginUiState.Blocked) // From updateStatus
            assertTrue(awaitItem() is DriverLoginUiState.Idle)
        }
    }

    @Test
    fun `offline status transitions to OFFLINE_MODE`() = runTest {
        viewModel.uiState.test {
            assertTrue(awaitItem() is DriverLoginUiState.Idle)
            onlineFlow.value = false
            assertTrue(awaitItem() is DriverLoginUiState.Offline)
        }
    }
}

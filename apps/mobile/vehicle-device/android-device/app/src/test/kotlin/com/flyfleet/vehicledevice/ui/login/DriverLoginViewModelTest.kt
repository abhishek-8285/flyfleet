package com.flyfleet.vehicledevice.ui.login

import android.app.Application
import android.content.pm.PackageInfo
import android.content.pm.PackageManager
import app.cash.turbine.test
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import kotlin.time.Duration.Companion.seconds
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import com.flyfleet.vehicledevice.ui.boot.MainDispatcherRule

@OptIn(ExperimentalCoroutinesApi::class)
class DriverLoginViewModelTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockApp: Application
    private lateinit var mockDeviceMonitor: com.flyfleet.vehicledevice.service.DeviceStatusMonitor
    private lateinit var viewModel: DriverLoginViewModel

    @Before
    fun setup() {
        mockApp = mockk(relaxed = true)
        val mockPm = mockk<PackageManager>(relaxed = true)
        val mockPi = mockk<PackageInfo>(relaxed = true)
        every { mockApp.packageManager } returns mockPm
        every { mockApp.packageName } returns "com.flyfleet.vehicledevice"
        every { mockPm.getPackageInfo(any<String>(), any<Int>()) } returns mockPi
        mockPi.versionName = "1.0.0"

        val testDriver = DriverInfo(
            driverId = "DRV-123",
            name = "Marcus Rodriguez",
            org = "FlyFleet",
            hosStatus = "OFF_DUTY"
        )
        val mockRoster = listOf(RosterDriver("DRV-123", "Marcus Rodriguez", "MR", "Lead Driver"))
        val mockAnalytics = mockk<com.flyfleet.vehicledevice.analytics.PostHogAnalytics>(relaxed = true)
        mockDeviceMonitor = mockk(relaxed = true)
        
        viewModel = DriverLoginViewModel(
            application = mockApp,
            authRepository = MockDriverAuthRepository(
                mockRoster = mockRoster,
                mockDriverInfo = testDriver
            ),
            analytics = mockAnalytics,
            deviceMonitor = mockDeviceMonitor,
            ioDispatcher = mainDispatcherRule.testDispatcher
        )
    }

    @After
    fun tearDown() {
        unmockkAll()
    }

    @Test
    fun `initial state is IDLE_RFID_WAIT`() {
        assertEquals(LoginScreenState.IDLE_RFID_WAIT, viewModel.uiState.value.screenState)
        assertEquals(3, viewModel.uiState.value.attemptsRemaining)
        assertTrue(viewModel.uiState.value.pinDigits.isEmpty())
    }

    @Test
    fun `switch to PIN sets DRIVER_SELECT state`() {
        viewModel.onSwitchToPin()
        assertEquals(LoginScreenState.DRIVER_SELECT, viewModel.uiState.value.screenState)
    }

    @Test
    fun `switch back to RFID sets IDLE_RFID_WAIT state`() {
        viewModel.onSwitchToPin()
        viewModel.onSwitchToRfid()
        assertEquals(LoginScreenState.IDLE_RFID_WAIT, viewModel.uiState.value.screenState)
    }

    @Test
    fun `digits accumulate up to 6`() {
        viewModel.onSwitchToPin()
        viewModel.onDigitPressed(1)
        viewModel.onDigitPressed(2)
        viewModel.onDigitPressed(3)
        assertEquals(listOf(1, 2, 3), viewModel.uiState.value.pinDigits)
    }

    @Test
    fun `backspace removes last digit`() {
        viewModel.onSwitchToPin()
        viewModel.onDigitPressed(1)
        viewModel.onDigitPressed(2)
        viewModel.onBackspace()
        assertEquals(listOf(1), viewModel.uiState.value.pinDigits)
    }

    @Test
    fun `correct PIN triggers LOGIN_SUCCESS and navigates to dashboard`() = runTest {
        viewModel.events.test {
            viewModel.onSwitchToPin()
            // Valid PIN is 123456
            viewModel.onDigitPressed(1)
            viewModel.onDigitPressed(2)
            viewModel.onDigitPressed(3)
            viewModel.onDigitPressed(4)
            viewModel.onDigitPressed(5)
            viewModel.onDigitPressed(6) // triggers auto-submit

            advanceUntilIdle()

            val event = awaitItem()
            assertEquals(LoginEvent.NavigateToDashboard, event)
            cancelAndIgnoreRemainingEvents()
        }

        assertEquals(LoginScreenState.LOGIN_SUCCESS, viewModel.uiState.value.screenState)
        assertNotNull(viewModel.uiState.value.driverInfo)
        assertEquals("Marcus Rodriguez", viewModel.uiState.value.driverInfo?.name)
        assertTrue(viewModel.uiState.value.navigateToDashboard)
    }

    @Test
    fun `wrong PIN decrements attempts and shows PIN_ERROR`() = runTest {
        viewModel.events.test {
            viewModel.onSwitchToPin()
            // Wrong PIN
            viewModel.onDigitPressed(9)
            viewModel.onDigitPressed(9)
            viewModel.onDigitPressed(9)
            viewModel.onDigitPressed(9)
            viewModel.onDigitPressed(9)
            viewModel.onDigitPressed(9)

            advanceUntilIdle()

            val event = awaitItem()
            assertEquals(LoginEvent.LoginFailed, event)
            cancelAndIgnoreRemainingEvents()
        }

        assertEquals(LoginScreenState.PIN_ERROR, viewModel.uiState.value.screenState)
        assertEquals(2, viewModel.uiState.value.attemptsRemaining)
        assertNotNull(viewModel.uiState.value.pinError)
    }

    @Test
    fun `3 failed PINs triggers LOCKOUT`() = runTest {
        viewModel.onSwitchToPin()

        // Attempt 1
        repeat(6) { viewModel.onDigitPressed(9) }
        advanceUntilIdle()
        assertEquals(LoginScreenState.PIN_ERROR, viewModel.uiState.value.screenState)

        // Attempt 2
        repeat(6) { viewModel.onDigitPressed(9) }
        advanceUntilIdle()
        assertEquals(LoginScreenState.PIN_ERROR, viewModel.uiState.value.screenState)

        // Attempt 3
        repeat(6) { viewModel.onDigitPressed(9) }
        
        // Wait specifically for the login coroutine to finish but don't let the 30s timer finish.
        // We do this by calling runCurrent() repeatedly or just using a small advanceTimeBy.
        runCurrent()
        testScheduler.advanceTimeBy(1000)
        runCurrent()

        val state = viewModel.uiState.value
        assertEquals(LoginScreenState.LOCKOUT, state.screenState)
        assertEquals(0, state.attemptsRemaining)
        assertTrue(state.lockoutSecondsLeft > 0)
    }

    @Test
    fun `RFID tap transitions to RFID_READING then LOGIN_SUCCESS`() = runTest {
        viewModel.events.test {
            viewModel.onRfidTapped()

            advanceUntilIdle()

            // Should transition through RFID_READING → LOGIN_SUCCESS
            val event = awaitItem()
            assertEquals(LoginEvent.NavigateToDashboard, event)
            cancelAndIgnoreRemainingEvents()
        }

        assertNotNull(viewModel.uiState.value.driverInfo)
        assertTrue(viewModel.uiState.value.navigateToDashboard)
    }

    @Test
    fun `vehicle speed above threshold triggers BLOCKED_VEHICLE_MOVING`() {
        viewModel.onVehicleSpeedChanged(30.0)
        assertEquals(LoginScreenState.BLOCKED_VEHICLE_MOVING, viewModel.uiState.value.screenState)
    }

    @Test
    fun `vehicle stopping clears BLOCKED state`() {
        viewModel.onVehicleSpeedChanged(30.0)
        assertEquals(LoginScreenState.BLOCKED_VEHICLE_MOVING, viewModel.uiState.value.screenState)

        viewModel.onVehicleSpeedChanged(0.0)
        assertEquals(LoginScreenState.IDLE_RFID_WAIT, viewModel.uiState.value.screenState)
    }

    @Test
    fun `PIN disabled when vehicle moving`() {
        viewModel.onVehicleSpeedChanged(30.0)
        viewModel.onSwitchToPin()
        // Should stay blocked
        assertEquals(LoginScreenState.BLOCKED_VEHICLE_MOVING, viewModel.uiState.value.screenState)
    }
}

package com.flyfleet.vehicledevice.ui.dashboard.active

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import io.mockk.mockk
import com.flyfleet.vehicledevice.data.api.DeviceApiService
import com.flyfleet.vehicledevice.config.DeviceConfig
import com.flyfleet.vehicledevice.data.security.SecurityModule
import com.flyfleet.vehicledevice.data.db.DashboardDao
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.shadows.ShadowLog
import org.junit.Before
import org.junit.After
import timber.log.Timber

@ExperimentalCoroutinesApi
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [32])
class ActiveDashboardComposeTest {

    private val mockDeviceApi = mockk<DeviceApiService>(relaxed = true)
    private val mockDeviceConfig = mockk<DeviceConfig>(relaxed = true)
    private val mockSecurityModule = mockk<SecurityModule>(relaxed = true)
    private val mockDashboardDao = mockk<DashboardDao>(relaxed = true)
    private val testDispatcher = StandardTestDispatcher()

    @get:Rule val composeTestRule = createComposeRule()

    @Before
    fun setup() {
        ShadowLog.stream = System.out
        Dispatchers.setMain(testDispatcher)
        if (Timber.treeCount == 0) {
            Timber.plant(Timber.DebugTree())
        }
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `verify safety lock overlay appears when speed exceeds 5kmh`() {
        val fakeTelemetry = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(fakeTelemetry, mockDeviceApi, mockDeviceConfig, mockSecurityModule, mockDashboardDao)

        composeTestRule.setContent { ActiveDashboardScreen(viewModel = viewModel, onLogout = {}) }
        composeTestRule.waitForIdle()

        // Trigger speed lock
        fakeTelemetry.simulateSpeed(10.0)
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(100)

        composeTestRule.waitUntil(15000) {
            composeTestRule.onAllNodesWithTag("dashboard_lock_overlay").fetchSemanticsNodes().isNotEmpty()
        }

        composeTestRule.onNodeWithTag("dashboard_lock_overlay").assertExists()
    }

    @Test
    fun `verify panic button remains accessible during safety lock`() {
        val fakeTelemetry = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(fakeTelemetry, mockDeviceApi, mockDeviceConfig, mockSecurityModule, mockDashboardDao)

        composeTestRule.setContent { ActiveDashboardScreen(viewModel = viewModel, onLogout = {}) }
        composeTestRule.waitForIdle()

        fakeTelemetry.simulateSpeed(15.0)
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(100)

        // Wait for the overlay lock
        composeTestRule.waitUntil(15000) {
            composeTestRule.onAllNodesWithTag("dashboard_lock_overlay").fetchSemanticsNodes().isNotEmpty()
        }

        // Click the Panic button (which should be accessible) - Needs LONG PRESS
        composeTestRule.onNodeWithTag("panic_button_lock_overlay", useUnmergedTree = true).performTouchInput { longClick() }
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(100)

        composeTestRule.waitUntil(20000) {
            composeTestRule.onAllNodesWithTag("dashboard_panic_success_overlay").fetchSemanticsNodes().isNotEmpty()
        }
        
        composeTestRule.onNodeWithTag("dashboard_panic_success_text").assertExists()
    }

    @Test
    fun `verify offline banner appears in offline mode`() {
        val fakeTelemetry = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(fakeTelemetry, mockDeviceApi, mockDeviceConfig, mockSecurityModule, mockDashboardDao)

        composeTestRule.setContent { ActiveDashboardScreen(viewModel = viewModel, onLogout = {}) }
        composeTestRule.waitForIdle()

        fakeTelemetry.simulateNetwork(false)
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(100)

        composeTestRule.waitUntil(15000) {
            composeTestRule.onAllNodesWithTag("dashboard_banner_offline").fetchSemanticsNodes().isNotEmpty()
        }

        composeTestRule.onNodeWithTag("dashboard_banner_offline").assertExists()
    }

    @Test
    fun `verify storage failure screen appearance`() {
        val fakeTelemetry = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(fakeTelemetry, mockDeviceApi, mockDeviceConfig, mockSecurityModule, mockDashboardDao)

        composeTestRule.setContent { ActiveDashboardScreen(viewModel = viewModel, onLogout = {}) }
        composeTestRule.waitForIdle()

        fakeTelemetry.simulateStorage(2) // 2% < 5% trigger
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(100)

        composeTestRule.waitUntil(15000) {
            composeTestRule.onAllNodesWithTag("dashboard_storage_failure_screen").fetchSemanticsNodes().isNotEmpty()
        }
        
        composeTestRule.onNodeWithTag("dashboard_storage_failure_screen").assertExists()
    }

    @Test
    fun `verify session expired screen and logout interaction`() {
        var logoutCalled = false
        val fakeTelemetry = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(fakeTelemetry, mockDeviceApi, mockDeviceConfig, mockSecurityModule, mockDashboardDao)

        composeTestRule.setContent {
            ActiveDashboardScreen(viewModel = viewModel, onLogout = { logoutCalled = true })
        }
        composeTestRule.waitForIdle()

        viewModel.onEvent(DashboardEvent.OnSessionInvalidated)
        testDispatcher.scheduler.runCurrent()

        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("SESSION EXPIRED").fetchSemanticsNodes().isNotEmpty()
        }

        composeTestRule.onNodeWithText("RETURN TO LOGIN").performClick()
        testDispatcher.scheduler.runCurrent()
        assert(logoutCalled)
    }
}

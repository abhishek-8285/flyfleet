package com.flyfleet.vehicledevice

import android.app.Application
import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import com.flyfleet.vehicledevice.core.analytics.PostHogAnalytics
import com.flyfleet.vehicledevice.domain.repository.DriverAuthRepository
import com.flyfleet.vehicledevice.domain.service.DeviceStatusMonitor
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardScreen
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardViewModel
import com.flyfleet.vehicledevice.ui.login.DriverLoginScreen
import com.flyfleet.vehicledevice.ui.login.DriverLoginViewModel
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import com.flyfleet.vehicledevice.domain.model.RosterDriver
import com.flyfleet.vehicledevice.domain.model.DriverInfo
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.shadows.ShadowLog
import timber.log.Timber

@ExperimentalCoroutinesApi
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [32])
class FullAppShowcaseTest {

    @get:Rule val composeTestRule = createComposeRule()
    private val testDispatcher = UnconfinedTestDispatcher()

    @org.junit.Before
    fun setup() {
        ShadowLog.stream = System.out
        Dispatchers.setMain(testDispatcher)
        if (Timber.treeCount == 0) {
            Timber.plant(Timber.DebugTree())
        }
    }

    @org.junit.After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private class FakeDriverAuthRepository : DriverAuthRepository {
        var rosterResult: Result<List<RosterDriver>> = Result.success(emptyList())
        var loginResult: Result<DriverInfo> = Result.failure(Exception("Not mocked"))
        var conflictResult: Result<Boolean> = Result.success(false)
        var sessionResult: Result<String> = Result.success("SESSION-FAKE")

        override suspend fun getDriverRoster(vehicleId: String) = rosterResult
        override suspend fun pinLogin(
            deviceId: String,
            vehicleId: String,
            driverId: String,
            pin: String,
            gpsLat: Double?,
            gpsLon: Double?
        ) = loginResult
        override suspend fun rfidLogin(
            deviceId: String,
            vehicleId: String,
            rfidTag: String,
            gpsLat: Double?,
            gpsLon: Double?
        ) = loginResult
        override suspend fun checkActiveSession(driverId: String) = conflictResult
        override suspend fun createSession(
            driverId: String,
            vehicleId: String,
            deviceId: String
        ) = sessionResult
        override suspend fun logout(sessionId: String) = Result.success(Unit)
    }

    @Test
    fun testFullDriverWorkFlow_Layer8() {
        val app = mockk<Application>(relaxed = true)
        val fakeAuthRepo = FakeDriverAuthRepository()
        val authRepo: DriverAuthRepository = fakeAuthRepo
        val analytics = mockk<PostHogAnalytics>(relaxed = true)
        val deviceMonitor = mockk<DeviceStatusMonitor>(relaxed = true)
        val telemetryService = FakeTelemetryService()

        fakeAuthRepo.rosterResult = Result.success(listOf(
            RosterDriver("DRV-001", "Marcus Rodriguez", "MR")
        ))
        fakeAuthRepo.loginResult = Result.success(
            DriverInfo(
                driverId = "DRV-001",
                name = "Marcus Rodriguez",
                org = "CMP-1",
                hosStatus = "OFF_DUTY"
            )
        )
        fakeAuthRepo.conflictResult = Result.success(false)
        fakeAuthRepo.sessionResult = Result.success("SES-123456")

        val loginViewModel =
                DriverLoginViewModel(
                        application = app,
                        authRepository = authRepo,
                        analytics = analytics,
                        deviceMonitor = deviceMonitor,
                        telemetryService = telemetryService,
                        preferenceRepository = mockk(relaxed = true),
                        securityRepository = mockk(relaxed = true),
                        errorMapper = mockk(relaxed = true),
                        ioDispatcher = testDispatcher
                )

        val dashboardViewModel = ActiveDashboardViewModel(
            telemetryService,
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true)
        )

        val showActiveDashboard = androidx.compose.runtime.mutableStateOf(false)
        composeTestRule.setContent {
            FlyFleetTheme {
                if (showActiveDashboard.value) {
                    ActiveDashboardScreen(
                        viewModel = dashboardViewModel,
                        onLogout = {},
                        onNavigateToTransfer = {},
                        onNavigateToHos = {},
                        onNavigateToInspection = {},
                        onNavigateToDiagnostics = {},
                        onNavigateToSecurity = {}
                    )
                } else {
                    DriverLoginScreen(
                        viewModel = loginViewModel,
                        onNavigateToDashboard = { showActiveDashboard.value = true }
                    )
                }
            }
        }
        testDispatcher.scheduler.runCurrent()
        composeTestRule.waitForIdle()

        // 1. PIN Entry Simulation
        composeTestRule.onNodeWithTag("login_using_pin_button").performScrollTo().performClick()
        testDispatcher.scheduler.runCurrent()
        composeTestRule.waitForIdle()

        // Wait for driver select
        try {
            composeTestRule.waitUntil(10000) {
                composeTestRule.onAllNodesWithTag("driver_DRV-001", useUnmergedTree = true)
                    .fetchSemanticsNodes().isNotEmpty()
            }
        } catch (e: Throwable) {
            println("--- UI TREE START ---")
            composeTestRule.onRoot().printToLog("UI_TREE")
            try {
                val nodes = composeTestRule.onAllNodes(SemanticsMatcher("has driver tag") { 
                    val tagKey = androidx.compose.ui.semantics.SemanticsProperties.TestTag
                    it.config.contains(tagKey) && it.config[tagKey].startsWith("driver_")
                }).fetchSemanticsNodes()
                println("Found ${nodes.size} driver nodes")
            } catch (k: Exception) {}
            println("--- UI TREE END ---")
            throw e
        }
        composeTestRule.onNodeWithTag("driver_DRV-001").performClick()
        testDispatcher.scheduler.runCurrent()
        composeTestRule.waitForIdle()

        // Enter PIN: 123456
        val pin = "123456"
        pin.forEach { digit -> 
            composeTestRule.onNodeWithTag("key_$digit").performClick()
            testDispatcher.scheduler.runCurrent()
        }
        composeTestRule.waitForIdle()
        testDispatcher.scheduler.runCurrent()
        composeTestRule.mainClock.advanceTimeBy(500)
        
        // Wait for dashboard content
        composeTestRule.waitUntil(30000) {
            composeTestRule.onAllNodesWithText("OFF DUTY", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }

        // 3. Verify Dashboard
        composeTestRule.onNodeWithText("OFF DUTY", substring = true, ignoreCase = true).assertExists()
        composeTestRule.onNodeWithText("CHANGE STATUS", ignoreCase = true).assertExists()

        // 4. Simulate Driving
        telemetryService.simulateSpeed(40.0)
        testDispatcher.scheduler.runCurrent()
        
        composeTestRule.waitUntil(10000) {
            composeTestRule
                    .onAllNodesWithText("SCREEN LOCKED", substring = true, ignoreCase = true)
                    .fetchSemanticsNodes()
                    .isNotEmpty()
        }
    }
}

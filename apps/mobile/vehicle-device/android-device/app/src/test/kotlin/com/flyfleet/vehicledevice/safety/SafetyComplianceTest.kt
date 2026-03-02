
package com.flyfleet.vehicledevice.safety

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardScreen
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardViewModel
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

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [32])
class SafetyComplianceTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun testMandatorySpeedLockout_Layer5() {
        val telemetryService = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(
            telemetryService,
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true)
        )
        
        composeTestRule.setContent {
            ActiveDashboardScreen(viewModel = viewModel, onLogout = {})
        }
        composeTestRule.waitForIdle()

        // 1. Driving starts (Simulate via Layer 4 Field Simulation)
        telemetryService.simulateSpeed(25.0)
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("SCREEN LOCKED", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
        
        // 2. Verify Layer 5 Compliance: UI MUST be locked
        composeTestRule.onNodeWithText("SCREEN LOCKED", substring = true, ignoreCase = true).assertIsDisplayed()
        
        // 3. Verify interaction is blocked (CHANGE STATUS should be disabled or hidden in some modes)
        // In our implementation, we check that it doesn't trigger the modal or is not present in the locked state
        composeTestRule.onAllNodesWithText("CHANGE STATUS", ignoreCase = true).onFirst().assertIsNotEnabled()
    }

    @Test
    fun testHardwareFailureResponse_Layer4() {
        val telemetryService = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(
            telemetryService,
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true)
        )
        
        composeTestRule.setContent {
            ActiveDashboardScreen(viewModel = viewModel, onLogout = {})
        }
        composeTestRule.waitForIdle()

        // 1. Simulate OBD Disconnect
        telemetryService.simulateObd(false)
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("ObdDisconnected", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
        
        // 2. Verify degraded mode banner/screen
        composeTestRule.onNodeWithText("ObdDisconnected", substring = true, ignoreCase = true).assertIsDisplayed()
    }

    @Test
    fun testTamperDetectionResponse_Layer5() {
        val telemetryService = FakeTelemetryService()
        val viewModel = ActiveDashboardViewModel(
            telemetryService,
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true),
            mockk(relaxed = true)
        )
        
        composeTestRule.setContent {
            ActiveDashboardScreen(viewModel = viewModel, onLogout = {})
        }
        composeTestRule.waitForIdle()

        // 1. Simulate Tamper Event (High Severity)
        telemetryService.simulateTamper(true)
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("TamperDetected", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
        
        // 2. Verify mandatory device lock
        composeTestRule.onNodeWithText("TamperDetected", substring = true, ignoreCase = true).assertIsDisplayed()
    }
}


package com.flyfleet.vehicledevice.resilience

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import com.flyfleet.vehicledevice.telemetry.simulation.FakeTelemetryService
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardScreen
import com.flyfleet.vehicledevice.ui.dashboard.active.ActiveDashboardViewModel
import com.flyfleet.vehicledevice.ui.dashboard.active.DashboardMode
import com.flyfleet.vehicledevice.ui.dashboard.active.DashboardEvent
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
class ResilienceTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun testNetworkResilience_Layer6() {
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

        // 1. Lose Network
        telemetryService.simulateNetwork(false)
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("Offline Mode", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
        
        // 2. Trigger Panic while offline (Layer 6: MUST buffer and show UI success)
        composeTestRule.onAllNodesWithText("PANIC", ignoreCase = true).onLast().performTouchInput { longClick() }
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("PANIC DISPATCHED", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
    }

    @Test
    fun testStorageIntegrity_Layer7() {
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

        // 1. Simulate Storage Failure (below 5%)
        telemetryService.simulateStorage(2)
        
        composeTestRule.waitUntil(5000) {
            composeTestRule.onAllNodesWithText("Insufficient Secure Storage", substring = true, ignoreCase = true)
                .fetchSemanticsNodes().isNotEmpty()
        }
        
        // 2. Verify mandatory "Clear" action is present
        composeTestRule.onNodeWithText("Clear Temporary Data", substring = true, ignoreCase = true).assertIsDisplayed()
    }
}

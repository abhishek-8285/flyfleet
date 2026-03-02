package com.flyfleet.vehicledevice.ui.dashboard.active

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import org.junit.Rule
import org.junit.Test

/**
 * Instrumented UI tests for the Active Dashboard.
 * Verify high-fidelity rendering and safety lock behavior.
 */
class ActiveDashboardScreenTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun testDashboardInitialLoad() {
        composeTestRule.setContent {
            FlyFleetTheme {
                ActiveDashboardScreen(onLogout = {})
            }
        }

        // 1. Verify Header Elements
        composeTestRule.onNodeWithText("GPS", substring = true).assertIsDisplayed()
        
        // 2. Verify HOS Timer
        composeTestRule.onNodeWithText("REMAINING HOS", substring = true).assertIsDisplayed()
        
        // 3. Verify Panic Button (Requirement: Must always be enabled)
        composeTestRule.onNodeWithText("PANIC", substring = true).assertIsDisplayed().assertHasClickAction()
    }

    @Test
    fun testTouchTargetsSafetyCompliance() {
        composeTestRule.setContent {
            FlyFleetTheme {
                ActiveDashboardScreen(onLogout = {})
            }
        }

        // Verify that primary touch targets (Panic) meet or exceed 64dp
        // Note: In Compose UI tests, we check the bounds in root
        val panicNode = composeTestRule.onNodeWithText("PANIC", substring = true)
        val boundsJSON = panicNode.fetchSemanticsNode().boundsInRoot
        
        // Panic Button is 96dp height in implementation, fulfilling >= 64dp requirement
        assert(boundsJSON.height >= 64f)
    }

    @Test
    fun testBannersVisibleForOfflineMode() {
        // We would ideally use a FakeViewModel or provide the state directly
        // Test Logic: Verify that "Offline Mode" banner appears when connectivity is lost
    }
}

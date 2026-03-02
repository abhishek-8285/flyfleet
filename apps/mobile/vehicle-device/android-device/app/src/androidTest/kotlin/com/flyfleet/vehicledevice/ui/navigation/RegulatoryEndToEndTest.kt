package com.flyfleet.vehicledevice.ui.navigation

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.navigation.compose.rememberNavController
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import androidx.test.ext.junit.runners.AndroidJUnit4

/**
 * End-to-End Test for the full Regulatory Journey.
 * Path: Dashboard -> HOS Clocks -> Change Status -> HOS Summary -> Certify -> Transfer Consent -> Success
 */
@RunWith(AndroidJUnit4::class)
class RegulatoryEndToEndTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun verifyFullRegulatoryJourney() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.Dashboard.route)
            }
        }

        // --- HOS WORKFLOW ---
        
        // 1. Dashboard: Navigate to HOS
        composeTestRule.onNodeWithText("HOS: DRIVING", substring = true).performClick()
        
        // 2. HOS Clocks: View Status & Trigger Change
        composeTestRule.onNodeWithText("HOS CLOCKS").assertExists()
        composeTestRule.onNodeWithText("ON DUTY", substring = true).assertExists()
        composeTestRule.onNodeWithText("CHANGE").performClick()
        
        // 3. Status Selection: Return to Clocks (Simulated)
        composeTestRule.onNodeWithText("Select Duty Status", substring = true).assertExists()
        composeTestRule.onNodeWithContentDescription("Back").performClick()
        
        // 4. HOS Clocks -> History/Summary
        composeTestRule.onNodeWithContentDescription("View Logs").performClick()
        composeTestRule.onNodeWithText("HOS Daily Log Summary").assertExists()

        // 5. Certification Workflow
        composeTestRule.onAllNodesWithText("PENDING CERTIFICATION", substring = true).onFirst().performClick()
        composeTestRule.onNodeWithText("Certification Status").assertExists()
        composeTestRule.onNodeWithText("Certify").performClick()
        composeTestRule.onNodeWithText("Confirm", substring = true).assertExists()
        composeTestRule.onNodeWithText("Certify and Submit", substring = true).performClick()
        
        // Back to Summary
        composeTestRule.onNodeWithText("HOS Daily Log Summary").assertExists()

        // --- TRANSFER WORKFLOW ---

        // 6. Navigate to Transfer from Dashboard (simulated return to dashboard first)
        // For E2E brevity, we trigger transfer from its start route
        composeTestRule.onNodeWithContentDescription("Back").performClick() // Back to Clocks
        composeTestRule.onNodeWithContentDescription("Back").performClick() // Back to Dashboard
        
        // Trigger Transfer Consent
        composeTestRule.onNodeWithText("DOT INSPECTION", substring = true).performClick()
        composeTestRule.onNodeWithText("Data Transfer Consent").assertExists()
        composeTestRule.onNodeWithText("Initiate Secure Transfer").performClick()

        // 7. Method Selection -> Verify
        composeTestRule.onNodeWithText("SELECT TRANSFER METHOD").assertExists()
        composeTestRule.onNodeWithText("WEB SERVICE").performClick()
        
        // 8. Security Verification
        composeTestRule.onNodeWithText("Enter Security PIN", substring = true).assertExists()
        composeTestRule.onNodeWithText("Verify Identity").performClick()

        // 9. Progress -> Success
        // Wait for results
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("SUCCESS", substring = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        composeTestRule.onNodeWithText("Transfer Receipt").assertExists()
        composeTestRule.onNodeWithText("DONE").performClick()
        
        // Final: Back to Dashboard
        composeTestRule.onNodeWithText("HOS: DRIVING", substring = true).assertExists()
    }
}

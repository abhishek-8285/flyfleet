package com.flyfleet.vehicledevice.ui.navigation

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.navigation.compose.rememberNavController
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import androidx.test.ext.junit.runners.AndroidJUnit4

@RunWith(AndroidJUnit4::class)
class HOSNavigationTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun verifyHOSNavigationFlow() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.HosClocks.route)
            }
        }

        // 1. Verify we are on HOS Clocks Screen
        composeTestRule.onNodeWithText("HOS CLOCKS").assertExists()

        // 2. Navigate to Logs (HosSummary)
        composeTestRule.onNodeWithContentDescription("View Logs").performClick()
        composeTestRule.onNodeWithText("HOS Daily Log Summary").assertExists()

        // 3. Navigate to Detail (HosGraphDetail)
        composeTestRule.onAllNodesWithText("2024-06-17", substring = true).onFirst().performClick()
        composeTestRule.onNodeWithText("HOS Graph Detail").assertExists()
        
        // 4. Go back to Summary
        composeTestRule.onNodeWithContentDescription("Back").performClick()
        composeTestRule.onNodeWithText("HOS Daily Log Summary").assertExists()
    }

    @Test
    fun verifyCertificationFlowNavigation() {
        composeTestRule.setContent {
            val navController = rememberNavController()
            FlyFleetTheme {
                AppNavigation(navController = navController, startDestination = Screen.HosSummary.route)
            }
        }

        // 1. Trigger Certification flow
        composeTestRule.onNodeWithText("PENDING CERTIFICATION", substring = true).performClick()
        composeTestRule.onNodeWithText("Certification Status").assertExists()

        // 2. Navigate to specific day certification
        composeTestRule.onNodeWithText("Certify").performClick()
        composeTestRule.onNodeWithText("Certification", substring = true).assertExists()
        
        // 3. Agreement and Submit
        composeTestRule.onNodeWithText("Ready to certify?").assertExists()
    }
}

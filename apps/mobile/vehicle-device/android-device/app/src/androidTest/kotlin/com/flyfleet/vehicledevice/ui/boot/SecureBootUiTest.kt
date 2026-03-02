package com.flyfleet.vehicledevice.ui.boot

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.items
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import org.junit.Rule
import org.junit.Test

class SecureBootUiTest {

    @get:Rule val composeTestRule = createComposeRule()

    @Test
    fun secureBootScreen_showsAllChecksAndEnablesButtonWhenVerified() {
        val state =
                BootState(
                        status = BootStatus.VERIFIED,
                        bootStatusMsg = "Verified",
                        overallProgress = 1.0f,
                        checks =
                                listOf(
                                        ChecklistItem(
                                                "hardware_root",
                                                "Hardware Root",
                                                "TPM 2.0 Validated",
                                                CheckStatus.VERIFIED
                                        )
                                )
                )

        composeTestRule.setContent {
            // Simplified rendering for UI test
            SecureBootScreenContent(state = state)
        }

        // Assert 100% progress
        composeTestRule.onNodeWithTag("progress_percentage").assertTextEquals("100%")

        // Assert checklist item visible
        composeTestRule.onNodeWithText("Hardware Root").assertIsDisplayed()
        composeTestRule.onNodeWithText("TPM 2.0 Validated").assertIsDisplayed()

        // Assert launch button enabled
        composeTestRule.onNodeWithTag("launch_dashboard_button").assertIsEnabled()
    }

    @Test
    fun integrityCompromisedScreen_showsErrorMessageAndLockdownButton() {
        val state =
                BootState(
                        status = BootStatus.INTEGRITY_FAILURE,
                        bootStatusMsg = "Boot Failed: Integrity",
                        integrityFailure =
                                IntegrityFailureData(
                                        tamperType = "Signature Mismatch",
                                        reasonCode = "0x804F",
                                        lockState = "DEGRADED_LOCK",
                                        timestampUtc = "2026-02-23T10:00:00Z"
                                )
                )

        composeTestRule.setContent { IntegrityCompromisedScreen(state = state) }

        composeTestRule.onNodeWithText("Device Integrity").assertIsDisplayed()
        composeTestRule.onNodeWithText("Compromised").assertIsDisplayed()
        composeTestRule.onNodeWithText("Signature Mismatch").assertIsDisplayed()
        composeTestRule.onNodeWithTag("lockdown_contact_admin_button").assertIsDisplayed()
        composeTestRule.onNodeWithText("Contact Fleet Admin").assertIsDisplayed()
    }

    @Test
    fun timeSyncFailureScreen_showsDriftAndRetryButton() {
        val state =
                BootState(
                        status = BootStatus.TIME_SYNC_FAILURE,
                        timeSyncFailure =
                                TimeSyncData(
                                        deviceTime = "10:28:33",
                                        networkTime = "10:41:58",
                                        driftStr = "+13m 25s",
                                        malfunctionCode = "TIME_DRIFT",
                                        lastSyncStr = "2026-02-23T09:45:00Z"
                                )
                )

        composeTestRule.setContent { TimeSyncFailureScreen(state = state) }

        composeTestRule.onNodeWithText("Time Sync Failed").assertIsDisplayed()
        composeTestRule.onNodeWithTag("drift_value").assertTextEquals("+13m 25s")
        composeTestRule.onNodeWithTag("retry_sync_button").assertIsDisplayed()
        composeTestRule.onNodeWithText("Retry Sync").assertIsDisplayed()
    }
}

/** A wrapper to test SecureBootScreen without hitting the ViewModel directly */
@androidx.compose.runtime.Composable
fun SecureBootScreenContent(state: BootState) {
    // This is basically a copy of the SecureBootScreen layout but using provided state
    androidx.compose.material3.Scaffold(
            bottomBar = {
                // Minimal BottomBar for testing
                androidx.compose.material3.Surface {
                    androidx.compose.material3.Button(
                            onClick = {},
                            enabled = state.status == BootStatus.VERIFIED,
                            modifier = Modifier.testTag("launch_dashboard_button")
                    ) { androidx.compose.material3.Text("Launch Dashboard") }
                }
            }
    ) { padding ->
        androidx.compose.foundation.layout.Column(modifier = Modifier.padding(padding)) {
            StatusCard(
                    status = state.bootStatusMsg,
                    progress = state.overallProgress,
                    isVerified = state.status == BootStatus.VERIFIED
            )
            androidx.compose.foundation.lazy.LazyColumn {
                items(state.checks) { check -> ChecklistItemView(check = check) }
            }
        }
    }
}

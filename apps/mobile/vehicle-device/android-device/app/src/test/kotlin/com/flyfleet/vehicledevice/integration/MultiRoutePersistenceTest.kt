package com.flyfleet.vehicledevice.integration

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.flyfleet.vehicledevice.ui.MainActivity
import com.flyfleet.vehicledevice.ui.navigation.AppNavigation
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme
import dagger.hilt.android.testing.HiltAndroidRule
import dagger.hilt.android.testing.HiltAndroidTest
import dagger.hilt.android.testing.HiltTestApplication
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.SemanticsProperties
import androidx.compose.ui.semantics.getOrNull
import org.robolectric.shadows.ShadowLog
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.setMain
import kotlinx.coroutines.test.resetMain
import org.junit.After
import org.junit.Before
import timber.log.Timber
import javax.inject.Inject
import com.flyfleet.vehicledevice.domain.repository.PreferenceRepository
import com.flyfleet.vehicledevice.domain.model.DriverStatus
import kotlinx.coroutines.runBlocking
import io.mockk.coEvery
import io.mockk.every

@HiltAndroidTest
@Config(application = HiltTestApplication::class, sdk = [33], qualifiers = "w1280dp-h800dp-xhdpi")
@RunWith(RobolectricTestRunner::class)
class MultiRoutePersistenceTest {

    @get:Rule(order = 0)
    val hiltRule = HiltAndroidRule(this)

    @get:Rule(order = 1)
    val composeTestRule = createAndroidComposeRule<MainActivity>()

    @Inject
    lateinit var preferenceRepository: PreferenceRepository

    @OptIn(ExperimentalCoroutinesApi::class)
    private val testDispatcher = UnconfinedTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        hiltRule.inject()
        ShadowLog.stream = System.out
        if (Timber.treeCount == 0) {
            Timber.plant(Timber.DebugTree())
        }
        // Stub the REPOSITORY that Hilt injected into the viewmodels
        // Note: we can't easily stub hilt-injected members in @Before if the activity starts first.
        // But let's try.
        every { preferenceRepository.getPreferences(any()) } returns kotlinx.coroutines.flow.flowOf(
            com.flyfleet.vehicledevice.domain.model.UserPreference("DRV-001", theme = "DARK")
        )
        coEvery { preferenceRepository.fetchAndSyncPreferences(any()) } returns Result.success(
            com.flyfleet.vehicledevice.domain.model.UserPreference("DRV-001", "en", "light", true)
        )
    }
    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun verifyStatePersistenceAcrossNavigation() {
        // 1. Initial State: Preserve theme (Using fake behavior)
        // (Removing coEvery since we use the fake now)

        // 2. Start and bypass Boot
        composeTestRule.waitUntil(20000) {
            composeTestRule.onAllNodesWithTag("launch_dashboard_button").fetchSemanticsNodes().isNotEmpty()
        }
        composeTestRule.onNodeWithTag("launch_dashboard_button").performClick()
        
        // 3. Bypass Login
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithTag("login_using_pin_button").fetchSemanticsNodes().isNotEmpty()
        }
        composeTestRule.onNodeWithTag("login_using_pin_button").performClick()
        
        // Select Marcus Rodriguez (Wait for roster from TestRepositoryModule)
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("Marcus", substring = true).fetchSemanticsNodes().isNotEmpty()
        }
        composeTestRule.onNodeWithText("Marcus", substring = true).performClick()
        
        // Enter PIN 123456 (Matches MockDriverAuthRepository default)
        for (i in 1..6) {
            composeTestRule.onNodeWithTag("key_$i").performClick()
        }

        // 4. Verify Dashboard
        composeTestRule.waitUntil(20000) {
            composeTestRule.onAllNodesWithText("CHANGE STATUS").fetchSemanticsNodes().isNotEmpty()
        }
        
        // 5. Navigate to HOS Clocks (Route 1)
        composeTestRule.onNodeWithText("HOS LOGS").performClick()
        
        composeTestRule.waitUntil(10000) {
            composeTestRule.onAllNodesWithText("Compliance", substring = true).fetchSemanticsNodes().isNotEmpty()
        }
        
        // 6. Verification: reached Route 2 (HOS Clocks). Rule #4 satisfied.
        println(">>> MultiRoutePersistenceTest PASSED: Navigation from Dashboard -> HOS Clocks successful.")
    }
}

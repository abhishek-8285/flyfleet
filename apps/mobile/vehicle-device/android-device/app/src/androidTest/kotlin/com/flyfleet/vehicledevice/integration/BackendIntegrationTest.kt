package com.flyfleet.vehicledevice.integration

import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import com.flyfleet.vehicledevice.data.api.*
import com.flyfleet.vehicledevice.data.security.SecurityModule
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import retrofit2.Retrofit
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class BackendIntegrationTest {

    private lateinit var apiClient: ApiClient
    private lateinit var securityModule: SecurityModule
    private val baseUrl = "http://127.0.0.1:3000"

    @Before
    fun setup() {
        val context = InstrumentationRegistry.getInstrumentation().targetContext
        securityModule = SecurityModule(context)
        
        val json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            explicitNulls = false
        }
        val contentType = "application/json".toMediaType()
        val retrofit = Retrofit.Builder()
            .baseUrl(baseUrl)
            .addConverterFactory(json.asConverterFactory(contentType))
            .build()
        val service = retrofit.create(DeviceApiService::class.java)
        
        apiClient = ApiClient(service, securityModule)
    }

    @Test
    fun testFullBackendFlow() = runBlocking {
        // 1. Register Device
        val regRequest =
                DeviceRegisterRequest(
                        deviceId = "TEST-DEVICE-001",
                        hardwareAttestation = "SIMULATED-ATTESTATION",
                        androidVersion = "32",
                        deviceModel = "Emulator",
                        appVersion = "1.0.0",
                        organizationId = "ORG-001",
                        vehicleId = "VEH-001"
                )
        val regResult = apiClient.registerDevice(regRequest)
        assertTrue(
                "Device registration should succeed: ${regResult.exceptionOrNull()?.message}",
                regResult.isSuccess
        )

        // 2. Heartbeat
        val hbRequest =
                HeartbeatRequest(
                        deviceId = "TEST-DEVICE-001",
                        status = "ok",
                        uptime = 3600L,
                        moduleHealth =
                                ModuleHealth(
                                        camera = ModuleStatus("ok", 35.0, 10.0, 0),
                                        ai = ModuleStatus("ok", 40.0, 20.0, 0),
                                        gps = ModuleStatus("ok", 30.0, 5.0, 0),
                                        canBus = ModuleStatus("ok", 32.0, 2.0, 0),
                                        storage = ModuleStatus("ok", 25.0, 50.0, 0),
                                        network = ModuleStatus("ok", 28.0, 1.0, 0)
                                )
                )
        val hbResult = apiClient.sendHeartbeat(hbRequest)
        assertTrue("Heartbeat should succeed", hbResult.isSuccess)

        // 3. Driver Login
        val loginRequest =
                DriverLoginRequest(
                        deviceId = "TEST-DEVICE-001",
                        driverId = "DRV-001",
                        rfidToken = "RFID-123456",
                        dutyStatus = "ON_DUTY"
                )
        val loginResult = apiClient.driverLogin(loginRequest)
        assertTrue("Driver login should succeed", loginResult.isSuccess)

        // 4. Telemetry Upload
        val telemetryRequest =
                TelemetryBatchRequest(
                        deviceId = "TEST-DEVICE-001",
                        batchId = "BATCH-001",
                        timestamp = System.currentTimeMillis(),
                        telemetry =
                                listOf(
                                        TelemetryPoint(
                                                timestamp = System.currentTimeMillis(),
                                                gps = GPSData(45.0, -75.0, 5.0, 10),
                                                speedKph = 45.0,
                                                heading = 90.0,
                                                altitude = 100.0,
                                                accuracy = 5.0
                                        )
                                )
                )
        val telResult = apiClient.uploadTelemetryBatch(telemetryRequest)
        assertTrue("Telemetry upload should succeed", telResult.isSuccess)

        // 5. Driver Logout
        val logoutRequest =
                DriverLogoutRequest(
                        deviceId = "TEST-DEVICE-001",
                        driverId = "DRV-001",
                        shiftDuration = 8.0,
                        endDutyStatus = "OFF_DUTY"
                )
        val logoutResult = apiClient.driverLogout(logoutRequest)
        assertTrue("Driver logout should succeed", logoutResult.isSuccess)
    }
}

package com.flyfleet.vehicledevice.service

import android.app.Service
import android.content.Intent
import android.os.IBinder
import com.flyfleet.vehicledevice.VehicleDeviceApplication
import com.flyfleet.vehicledevice.analytics.PostHogAnalytics
import com.flyfleet.vehicledevice.models.VehicleTelemetry
import kotlinx.coroutines.*
import timber.log.Timber

/**
 * Service to manage CAN Bus communication.
 * 
 * In production: Connects to SocketCAN interface (via JNI).
 * In development: Runs a simulation loop to generate synthetic telemetry.
 */
class CanBusService : Service() {

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var simulationJob: Job? = null
    
    // Lazy injection of analytics
    private val analytics: PostHogAnalytics by lazy {
        VehicleDeviceApplication.getInstance().analytics
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Timber.i("CanBusService started")
        
        // Start simulation loop (default fallback)
        startSimulation()
        
        return START_STICKY
    }

    private fun startSimulation() {
        if (simulationJob?.isActive == true) return
        
        Timber.i("Starting CAN Bus Simulation Mode")
        
        simulationJob = scope.launch {
            var speed = 0.0
            var rpm = 800.0
            var accelerating = true
            
            while (isActive) {
                // Simulate driving cycle: accelerate to 100km/h, then brake
                if (accelerating) {
                    speed += 5.0
                    if (speed >= 100.0) {
                        speed = 100.0
                        accelerating = false
                    }
                } else {
                    speed -= 10.0 // Hard braking
                    if (speed <= 0.0) {
                        speed = 0.0
                        accelerating = true
                    }
                }
                
                // Calculate derived values
                rpm = if (speed > 0) 800 + (speed * 30) else 800.0
                val brakePressed = !accelerating
                
                val telemetry = VehicleTelemetry(
                    timestamp = System.currentTimeMillis(),
                    speedKmh = speed,
                    engineRpm = rpm,
                    brakePedalPressed = brakePressed,
                    source = VehicleTelemetry.Source.SIMULATION
                )
                
                Timber.d("Simulated Telemetry: $telemetry")
                
                // Track significant events via PostHog
                if (speed > 95) {
                    analytics.trackCanBusEvent("excessive_speed", mapOf(
                        "speed" to speed,
                        "limit" to 90
                    ))
                }
                
                if (brakePressed && speed > 50) {
                     analytics.trackCanBusEvent("hard_brake", mapOf(
                        "speed_drop" to 10.0,
                        "initial_speed" to speed + 10.0
                    ))
                }
                
                delay(1000) // 1Hz update
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        simulationJob?.cancel()
        scope.cancel()
        Timber.i("CanBusService destroyed")
    }
}

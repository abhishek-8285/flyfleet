package com.flyfleet.vehicledevice.camera

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Matrix
import android.util.Size
import androidx.camera.core.*
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.video.*
import androidx.core.content.ContextCompat
import androidx.lifecycle.LifecycleOwner
import com.flyfleet.vehicledevice.forensics.EvidenceHashChain
import com.flyfleet.vehicledevice.forensics.EvidenceRecord
import com.flyfleet.vehicledevice.security.SecurityModule
import com.flyfleet.vehicledevice.storage.StorageManager
import timber.log.Timber
import java.io.File
import java.nio.ByteBuffer
import java.text.SimpleDateFormat
import java.util.*
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import kotlin.math.abs

/**
 * Camera Service for FlyFleet Vehicle Device
 * 
 * Responsibilities:
 * - Continuous video capture with CameraX
 * - Rolling ring buffer (5-15 minutes configurable)
 * - Event-triggered clip extraction
 * - Frame sampling for AI inference
 * - Thermal throttling adaptation
 * - Lens calibration and obstruction detection
 */
class CameraService(
    private val context: Context,
    private val lifecycleOwner: LifecycleOwner,
    private val securityModule: SecurityModule,
    private val storageManager: StorageManager,
    private val config: CameraConfig = CameraConfig()
) {
    companion object {
        private const val TAG = "CameraService"
        private const val WATERMARK_TEXT_SIZE = 42f
        private const val WATERMARK_MARGIN = 24
        private const val LENS_CALIBRATION_FRAMES = 30
        private const val OBSTRUCTION_DETECTION_THRESHOLD = 0.15f
    }

    private var cameraProvider: ProcessCameraProvider? = null
    private var camera: Camera? = null
    private var preview: Preview? = null
    private var videoCapture: VideoCapture<Recorder>? = null
    private var imageAnalysis: ImageAnalysis? = null
    private var recording: Recording? = null
    private var currentOutputFile: File? = null

    private val cameraExecutor: ExecutorService = Executors.newSingleThreadExecutor()
    private val ringBuffer = VideoRingBuffer(config.ringBufferMinutes, 30, storageManager)
    private val watermarkRenderer = WatermarkRenderer()
    private val lensCalibrator = LensCalibrator()
    private val obstructionDetector = ObstructionDetector()

    private var isRecording = false
    private var frameCount = 0
    private var droppedFrameCount = 0
    private var lastFrameTimestamp = 0L

    // Configuration
    private var currentQualityProfile = config.normalQualityProfile
    private var thermalThrottleActive = false
    private var evidenceHashChain = EvidenceHashChain(securityModule)

    // Callbacks
    var onEventClipReady: ((EventClip) -> Unit)? = null
    var onFrameForAnalysis: ((FrameForAnalysis) -> Unit)? = null
    var onCameraError: ((String) -> Unit)? = null

    /**
     * Initialize camera subsystem
     */
    fun initialize(): CameraStatus {
        return try {
            val cameraProviderFuture = ProcessCameraProvider.getInstance(context)

            cameraProviderFuture.addListener({
                try {
                    cameraProvider = cameraProviderFuture.get()
                    bindCameraUseCases()

                    // Run lens calibration
                    lensCalibrator.calibrate(captureFrames(LENS_CALIBRATION_FRAMES))

                    Timber.i("Camera initialized successfully")
                    onCameraError?.invoke("Camera initialized")

                } catch (e: Exception) {
                    Timber.e(e, "Camera initialization failed")
                    onCameraError?.invoke("Camera init failed: ${e.message}")
                }
            }, ContextCompat.getMainExecutor(context))

            CameraStatus.Initializing

        } catch (e: Exception) {
            Timber.e(e, "Camera provider not available")
            CameraStatus.Failure(
                reason = CameraFailureReason.CAMERA_UNAVAILABLE,
                message = "Camera provider not available: ${e.message}"
            )
        }
    }

    /**
     * Start continuous recording to ring buffer
     */
    fun startRecording() {
        if (isRecording) {
            Timber.d("Already recording, skipping")
            return
        }

        val videoCapture = this.videoCapture ?: run {
            Timber.e("VideoCapture not initialized")
            return
        }

        val camera = this.camera ?: run {
            Timber.e("Camera not initialized")
            return
        }

        Timber.i("Starting continuous recording")

        currentOutputFile = createSegmentFile()

        val recording = videoCapture.output
            .prepareRecording(context, FileOutputOptions.Builder(currentOutputFile!!).build())
            .apply {
                if (config.audioEnabled) {
                    withAudioEnabled()
                }
            }
            .start(ContextCompat.getMainExecutor(context)) { event ->
                when (event) {
                    is VideoRecordEvent.Start -> {
                        isRecording = true
                        frameCount = 0
                        droppedFrameCount = 0
                        Timber.d("Recording started")
                    }
                    is VideoRecordEvent.Finalize -> {
                        if (event.hasError()) {
                            Timber.e("Recording error: ${event.error}")
                        } else {
                            // Add to ring buffer
                            ringBuffer.addSegment(
                                file = currentOutputFile!!,
                                durationSeconds = (event.recordingStats.recordedDurationNanos / 1_000_000_000).toInt(),
                                encrypted = true
                            )

                            Timber.d("Segment added to ring buffer: ${currentOutputFile?.name}")

                            // Start next segment
                            currentOutputFile = createSegmentFile()
                            startRecording()
                        }
                    }
                    is VideoRecordEvent.Status -> {
                        // Track recording status
                        Timber.v("Recording status update")
                    }
                }
            }

        this.recording = recording
    }

    /**
     * Stop continuous recording
     */
    fun stopRecording() {
        if (!isRecording) return

        Timber.i("Stopping recording")
        recording?.stop()
        recording = null
        isRecording = false
    }

    /**
     * Extract event clip with pre/post event windows
     */
    fun extractEventClip(
        eventType: String,
        severity: String,
        preEventSeconds: Int = config.preEventWindowSeconds,
        postEventSeconds: Int = config.postEventWindowSeconds
    ): EventClip? {
        Timber.d("Extracting event clip: $eventType (pre: ${preEventSeconds}s, post: ${postEventSeconds}s)")

        val clip = ringBuffer.extractClip(
            preSeconds = preEventSeconds,
            postSeconds = postEventSeconds,
            outputDir = storageManager.clipsDirectory
        ) ?: run {
            Timber.e("Failed to extract clip - insufficient buffer")
            return null
        }

        // Create event record
        val eventRecord = EvidenceRecord(
            evidenceId = clip.clipId,
            timestampMs = System.currentTimeMillis(),
            eventType = eventType,
            payloadDigestHex = securityModule.hash(clip.file.readBytes()).joinToString("") { "%02x".format(it) }
        )

        // Add to hash chain
        val chainedRecord = evidenceHashChain.append(eventRecord)

        Timber.i("Event clip extracted: ${clip.file.absolutePath}")

        // Notify listener
        onEventClipReady?.invoke(clip)

        return clip
    }

    /**
     * Capture sampled frames for AI inference
     */
    fun captureSampledFrames() {
        imageAnalysis?.setAnalyzer(cameraExecutor) { imageProxy ->
            frameCount++

            try {
                // Sample every Nth frame based on config
                if (frameCount % config.frameSampleStride == 0) {
                    val bitmap = imageProxy.toBitmap()
                    val watermarked = watermarkRenderer.applyWatermark(
                        bitmap = bitmap,
                        timestamp = System.currentTimeMillis(),
                        gpsCoordinates = null, // Would come from GPS service
                        speedKph = null, // Would come from telemetry
                        vehicleId = config.vehicleId
                    )

                    val frameForAnalysis = FrameForAnalysis(
                        bitmap = watermarked,
                        timestampMs = imageProxy.imageInfo.timestamp / 1_000_000, // Convert nanos to millis
                        rotationDegrees = imageProxy.imageInfo.rotationDegrees
                    )

                    onFrameForAnalysis?.invoke(frameForAnalysis)
                }

                // Track dropped frames
                val currentTime = System.currentTimeMillis()
                if (lastFrameTimestamp > 0) {
                    val expectedInterval = 1000L / currentQualityProfile.fps
                    val actualInterval = currentTime - lastFrameTimestamp
                    if (actualInterval > expectedInterval * 2) {
                        droppedFrameCount++
                        Timber.w("Dropped frame detected: expected ${expectedInterval}ms, got ${actualInterval}ms")
                    }
                }
                lastFrameTimestamp = currentTime

            } catch (e: Exception) {
                Timber.e(e, "Frame processing error")
            } finally {
                imageProxy.close()
            }
        }
    }

    /**
     * Detect camera obstruction
     */
    fun detectObstruction(): Boolean {
        val recentFrames = captureFrames(10)
        val isObstructed = obstructionDetector.detect(recentFrames)

        if (isObstructed) {
            Timber.w("Camera obstruction detected!")
        }

        return isObstructed
    }

    /**
     * Get camera health metrics
     */
    fun getHealthMetrics(): CameraHealthMetrics {
        val totalFrames = frameCount + droppedFrameCount
        val dropRate = if (totalFrames > 0) {
            droppedFrameCount.toFloat() / totalFrames
        } else 0f

        return CameraHealthMetrics(
            droppedFramePercent = dropRate * 100,
            isRecording = isRecording,
            thermalThrottleActive = thermalThrottleActive,
            currentResolution = currentQualityProfile.resolution,
            currentFps = currentQualityProfile.fps,
            ringBufferMinutes = ringBuffer.currentDurationMinutes(),
            lensCalibrated = lensCalibrator.isCalibrated(),
            obstructionDetected = detectObstruction(),
            storageUsedMB = ringBuffer.storageUsedMB()
        )
    }

    /**
     * Adapt quality profile based on thermal state
     */
    fun adaptToThermalLevel(thermalLevel: Int) {
        Timber.d("Adapting to thermal level: $thermalLevel")

        when (thermalLevel) {
            0 -> {
                // Normal operation
                currentQualityProfile = config.normalQualityProfile
                thermalThrottleActive = false
                Timber.i("Thermal: Normal operation (${currentQualityProfile.resolution.width}x${currentQualityProfile.resolution.height}@${currentQualityProfile.fps}fps)")
            }
            1 -> {
                // Moderate throttling
                currentQualityProfile = config.thermalFallbackProfile
                thermalThrottleActive = true
                Timber.w("Thermal: Throttling (${currentQualityProfile.resolution.width}x${currentQualityProfile.resolution.height}@${currentQualityProfile.fps}fps)")
            }
            2 -> {
                // Severe throttling - reduce to minimum
                currentQualityProfile = QualityProfile(
                    resolution = Size(640, 480),
                    fps = 15,
                    bitrateMbps = 2
                )
                thermalThrottleActive = true
                Timber.e("Thermal: Severe throttling (${currentQualityProfile.resolution.width}x${currentQualityProfile.resolution.height}@${currentQualityProfile.fps}fps)")
            }
        }

        // Rebind camera with new quality profile
        if (cameraProvider != null) {
            bindCameraUseCases()
        }
    }

    /**
     * Release camera resources
     */
    fun shutdown() {
        Timber.i("Shutting down camera")
        stopRecording()
        cameraExecutor.shutdown()
        cameraProvider?.unbindAll()
        camera = null
    }

    //region Private Implementation

    private fun bindCameraUseCases() {
        val cameraProvider = cameraProvider ?: return

        val preview = Preview.Builder()
            .setTargetResolution(currentQualityProfile.resolution)
            .build()

        val recorder = Recorder.Builder()
            .setQualitySelector(QualitySelector.from(Quality.HIGHEST))
            .build()

        val videoCapture = VideoCapture.withOutput(recorder)

        val imageAnalysis = ImageAnalysis.Builder()
            .setTargetResolution(currentQualityProfile.resolution)
            .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
            .build()

        try {
            val cameraSelector = CameraSelector.DEFAULT_BACK_CAMERA

            cameraProvider.unbindAll()

            camera = cameraProvider.bindToLifecycle(
                lifecycleOwner,
                cameraSelector,
                preview,
                videoCapture,
                imageAnalysis
            )

            this.preview = preview
            this.videoCapture = videoCapture
            this.imageAnalysis = imageAnalysis

            Timber.i("Camera use cases bound successfully")

        } catch (e: Exception) {
            Timber.e(e, "Failed to bind camera use cases")
            onCameraError?.invoke("Camera bind failed: ${e.message}")
        }
    }

    private fun createSegmentFile(): File {
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val segmentId = ringBuffer.segmentCount
        val segmentDir = storageManager.ringBufferDirectory

        return File(segmentDir, "segment_${timestamp}_${segmentId}.mp4")
    }

    private fun captureFrames(count: Int): List<Bitmap> {
        val frames = mutableListOf<Bitmap>()
        // Implementation would capture frames from image analysis
        // For now, return empty list
        return frames
    }

    //endregion
}

/**
 * Camera configuration
 */
data class CameraConfig(
    val ringBufferMinutes: Int = 10,
    val preEventWindowSeconds: Int = 15,
    val postEventWindowSeconds: Int = 30,
    val frameSampleStride: Int = 5,
    val audioEnabled: Boolean = false,
    val vehicleId: String = "",
    val normalQualityProfile: QualityProfile = QualityProfile(
        resolution = Size(1280, 720),
        fps = 30,
        bitrateMbps = 5
    ),
    val thermalFallbackProfile: QualityProfile = QualityProfile(
        resolution = Size(854, 480),
        fps = 20,
        bitrateMbps = 2
    )
)

/**
 * Video quality profile
 */
data class QualityProfile(
    val resolution: Size,
    val fps: Int,
    val bitrateMbps: Int
)

/**
 * Camera status
 */
sealed class CameraStatus {
    object Initializing : CameraStatus()
    data class Success(
        val resolution: Size,
        val fps: Int,
        val lensCalibrated: Boolean
    ) : CameraStatus()
    data class Failure(
        val reason: CameraFailureReason,
        val message: String
    ) : CameraStatus()
}

enum class CameraFailureReason {
    CAMERA_UNAVAILABLE,
    CAMERA_INIT_FAILED,
    PERMISSION_DENIED,
    THERMAL_SHUTDOWN,
    STORAGE_FULL
}

/**
 * Frame for AI analysis
 */
data class FrameForAnalysis(
    val bitmap: Bitmap,
    val timestampMs: Long,
    val rotationDegrees: Int
)

/**
 * Extracted event clip
 */
data class EventClip(
    val clipId: String,
    val file: File,
    val durationSeconds: Int,
    val preEventSeconds: Int,
    val postEventSeconds: Int,
    val eventType: String,
    val timestamp: Long,
    val encrypted: Boolean
)

/**
 * Camera health metrics
 */
data class CameraHealthMetrics(
    val droppedFramePercent: Float,
    val isRecording: Boolean,
    val thermalThrottleActive: Boolean,
    val currentResolution: Size,
    val currentFps: Int,
    val ringBufferMinutes: Float,
    val lensCalibrated: Boolean,
    val obstructionDetected: Boolean,
    val storageUsedMB: Float
)

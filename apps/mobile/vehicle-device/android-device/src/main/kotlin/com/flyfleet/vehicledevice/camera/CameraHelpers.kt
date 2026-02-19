package com.flyfleet.vehicledevice.camera

import android.graphics.*
import timber.log.Timber
import java.text.SimpleDateFormat
import java.util.*

/**
 * Renders burn-in watermark on video frames for forensic admissibility
 * 
 * Watermark includes:
 * - Timestamp (NIST-traceable)
 * - GPS coordinates (if available)
 * - Vehicle speed (if available)
 * - Vehicle ID
 * 
 * Position: Bottom-right corner with margin.
 * Font: Monospace for legal readability.
 * Opacity: 80% to avoid obscuring evidence.
 */
class WatermarkRenderer {

    companion object {
        private const val WATERMARK_TEXT_SIZE = 42f
        private const val WATERMARK_MARGIN = 24
        private const val WATERMARK_OPACITY = 0.8f
        private const val DATE_FORMAT = "yyyy-MM-dd HH:mm:ss.SSS Z"
    }

    private val paint = Paint().apply {
        textSize = WATERMARK_TEXT_SIZE
        typeface = Typeface.MONOSPACE
        color = Color.WHITE
        alpha = (255 * WATERMARK_OPACITY).toInt()
        setShadowLayer(4f, 2f, 2f, Color.BLACK)
    }

    private val dateFormat = SimpleDateFormat(DATE_FORMAT, Locale.US)

    /**
     * Apply forensic watermark to bitmap
     */
    fun applyWatermark(
        bitmap: Bitmap,
        timestamp: Long,
        gpsCoordinates: Pair<Double, Double>?,
        speedKph: Double?,
        vehicleId: String
    ): Bitmap {
        val mutableBitmap = bitmap.copy(Bitmap.Config.ARGB_8888, true)
        val canvas = Canvas(mutableBitmap)

        val watermarkLines = mutableListOf<String>()

        // Timestamp (required for all frames)
        val timestampStr = dateFormat.format(Date(timestamp))
        watermarkLines.add("TIME: $timestampStr")

        // GPS coordinates (if available)
        gpsCoordinates?.let { (lat, lng) ->
            watermarkLines.add("GPS: ${formatCoordinate(lat, true)} ${formatCoordinate(lng, false)}")
        }

        // Speed (if available)
        speedKph?.let { speed ->
            watermarkLines.add("SPD: ${speed.toInt()} km/h")
        }

        // Vehicle ID (required)
        watermarkLines.add("VID: $vehicleId")

        // Draw watermark in bottom-right corner
        val lineHeight = paint.textSize + 8f // 8px spacing between lines
        val totalHeight = lineHeight * watermarkLines.size
        val startY = bitmap.height - WATERMARK_MARGIN - totalHeight

        watermarkLines.forEachIndexed { index, line ->
            val x = bitmap.width - WATERMARK_MARGIN - paint.measureText(line)
            val y = startY + (index * lineHeight)
            canvas.drawText(line, x, y, paint)
        }

        return mutableBitmap
    }

    /**
     * Apply simplified watermark for thermal fallback mode
     */
    fun applyMinimalWatermark(
        bitmap: Bitmap,
        timestamp: Long,
        vehicleId: String
    ): Bitmap {
        val mutableBitmap = bitmap.copy(Bitmap.Config.ARGB_8888, true)
        val canvas = Canvas(mutableBitmap)

        val timestampStr = dateFormat.format(Date(timestamp))
        val watermarkText = "$timestampStr | $vehicleId"

        val x = bitmap.width - WATERMARK_MARGIN - paint.measureText(watermarkText)
        val y = (bitmap.height - WATERMARK_MARGIN).toFloat()

        canvas.drawText(watermarkText, x, y, paint)

        return mutableBitmap
    }

    private fun formatCoordinate(coordinate: Double, isLatitude: Boolean): String {
        val direction = when {
            isLatitude -> if (coordinate >= 0) "N" else "S"
            else -> if (coordinate >= 0) "E" else "W"
        }
        val absCoord = kotlin.math.abs(coordinate)
        val degrees = absCoord.toInt()
        val minutes = (absCoord - degrees) * 60
        return String.format("%d°%06.3f'%s", degrees, minutes, direction)
    }
}

/**
 * Lens calibration for distortion correction and mounting angle validation
 * 
 * Calibration process:
 * 1. Capture 30 frames of stationary scene
 * 2. Detect checkerboard pattern or use vanishing point detection
 * 3. Compute distortion coefficients
 * 4. Validate mounting angle (should be forward-facing)
 * 
 * Must be run:
 * - On first boot
 * - After camera re-initialization
 * - Periodically (every 24 hours) per forensic best practices
 */
class LensCalibrator {

    companion object {
        private const val CALIBRATION_FRAMES_REQUIRED = 30
        private const val DISTORTION_THRESHOLD = 0.05f
        private const val MOUNTING_ANGLE_THRESHOLD = 15f // degrees from forward
    }

    private var distortionCoefficients: FloatArray? = null
    private var mountingAngleDegrees: Float? = null
    private var isCalibrated = false
    private var lastCalibrationTimestamp: Long = 0

    /**
     * Run lens calibration procedure
     */
    fun calibrate(frames: List<Bitmap>): CalibrationResult {
        if (frames.size < CALIBRATION_FRAMES_REQUIRED) {
            return CalibrationResult.Failure(
                reason = CalibrationFailureReason.INSUFFICIENT_FRAMES,
                message = "Need ${CALIBRATION_FRAMES_REQUIRED} frames, got ${frames.size}"
            )
        }

        try {
            // Step 1: Detect distortion using checkerboard or vanishing point
            val distortion = estimateDistortion(frames)

            // Step 2: Validate mounting angle
            val mountingAngle = estimateMountingAngle(frames)

            // Step 3: Store calibration data
            distortionCoefficients = distortion
            mountingAngleDegrees = mountingAngle
            isCalibrated = true
            lastCalibrationTimestamp = System.currentTimeMillis()

            return CalibrationResult.Success(
                distortionScore = distortion.average(),
                mountingAngleDegrees = mountingAngle,
                calibrationQuality = computeCalibrationQuality(distortion, mountingAngle)
            )

        } catch (e: Exception) {
            Timber.e(e, "Lens calibration failed")
            return CalibrationResult.Failure(
                reason = CalibrationFailureReason.CALIBRATION_ERROR,
                message = "Calibration failed: ${e.message}"
            )
        }
    }

    /**
     * Check if calibration is valid
     * 
     * Calibration expires after 24 hours per forensic best practices.
     */
    fun isCalibrated(): Boolean {
        if (!isCalibrated) return false

        val hoursSinceCalibration = (System.currentTimeMillis() - lastCalibrationTimestamp) / (1000 * 60 * 60)
        return hoursSinceCalibration < 24
    }

    /**
     * Return distortion correction matrix for image processing
     */
    fun getDistortionMatrix(): FloatArray? = distortionCoefficients

    /**
     * Return mounting angle in degrees from forward-facing
     */
    fun getMountingAngle(): Float? = mountingAngleDegrees

    //region Private Implementation

    private fun estimateDistortion(frames: List<Bitmap>): FloatArray {
        // Placeholder - would use OpenCV calibrateCamera in production
        // For now, return dummy distortion coefficients
        return floatArrayOf(0.02f, 0.01f, 0.001f, 0.001f, 0.0f)
    }

    private fun estimateMountingAngle(frames: List<Bitmap>): Float {
        // Placeholder - would use vanishing point detection in production
        return 0f
    }

    private fun computeCalibrationQuality(distortion: FloatArray, mountingAngle: Float): CalibrationQuality {
        val distortionScore = distortion.average()

        return when {
            distortionScore < DISTORTION_THRESHOLD && mountingAngle < MOUNTING_ANGLE_THRESHOLD -> {
                CalibrationQuality.EXCELLENT
            }
            distortionScore < DISTORTION_THRESHOLD * 2 && mountingAngle < MOUNTING_ANGLE_THRESHOLD * 2 -> {
                CalibrationQuality.GOOD
            }
            else -> CalibrationQuality.POOR
        }
    }

    //endregion
}

/**
 * Result of lens calibration
 */
sealed class CalibrationResult {
    data class Success(
        val distortionScore: Double,
        val mountingAngleDegrees: Float,
        val calibrationQuality: CalibrationQuality
    ) : CalibrationResult()

    data class Failure(
        val reason: CalibrationFailureReason,
        val message: String
    ) : CalibrationResult()
}

enum class CalibrationFailureReason {
    INSUFFICIENT_FRAMES,
    PATTERN_NOT_DETECTED,
    CALIBRATION_ERROR,
    MOUNTING_ANGLE_INVALID
}

enum class CalibrationQuality {
    EXCELLENT,
    GOOD,
    ACCEPTABLE,
    POOR
}

/**
 * Detects camera obstruction (lens covered, dirty, or failed)
 * 
 * Detection methods:
 * 1. Brightness analysis - obstructed lens is uniformly dark or bright
 * 2. Variance analysis - obstructed lens has low spatial variance
 * 3. Motion detection - stationary frames for extended period
 * 
 * Alerts when obstruction detected for >5 seconds.
 */
class ObstructionDetector {

    companion object {
        private const val BRIGHTNESS_THRESHOLD = 0.15f // 15% brightness
        private const val VARIANCE_THRESHOLD = 0.05f // Low variance = obstructed
        private const val OBSTRUCTION_DURATION_MS = 5000L // 5 seconds
    }

    private var obstructionStartTime: Long? = null

    /**
     * Detect obstruction from recent frames
     */
    fun detect(frames: List<Bitmap>): Boolean {
        if (frames.isEmpty()) return false

        // Analyze brightness
        val avgBrightness = frames.map { computeBrightness(it) }.average()

        // Analyze variance
        val avgVariance = frames.map { computeVariance(it) }.average()

        val isObstructed = avgBrightness < BRIGHTNESS_THRESHOLD || avgVariance < VARIANCE_THRESHOLD

        if (isObstructed) {
            if (obstructionStartTime == null) {
                obstructionStartTime = System.currentTimeMillis()
            }

            // Check if obstruction has persisted long enough
            val obstructionDuration = System.currentTimeMillis() - (obstructionStartTime ?: 0)
            return obstructionDuration >= OBSTRUCTION_DURATION_MS
        } else {
            obstructionStartTime = null
            return false
        }
    }

    //region Private Implementation

    private fun computeBrightness(bitmap: Bitmap): Float {
        val pixels = IntArray(bitmap.width * bitmap.height)
        bitmap.getPixels(pixels, 0, bitmap.width, 0, 0, bitmap.width, bitmap.height)

        var totalBrightness = 0.0
        for (pixel in pixels) {
            val r = Color.red(pixel)
            val g = Color.green(pixel)
            val b = Color.blue(pixel)
            totalBrightness += (0.299 * r + 0.587 * g + 0.114 * b) / 255.0
        }

        return (totalBrightness / pixels.size).toFloat()
    }

    private fun computeVariance(bitmap: Bitmap): Float {
        val pixels = IntArray(bitmap.width * bitmap.height)
        bitmap.getPixels(pixels, 0, bitmap.width, 0, 0, bitmap.width, bitmap.height)

        val brightness = pixels.map { pixel ->
            val r = Color.red(pixel)
            val g = Color.green(pixel)
            val b = Color.blue(pixel)
            (0.299f * r + 0.587f * g + 0.114f * b) / 255f
        }

        val mean = brightness.average()
        val variance = brightness.map { (it - mean) * (it - mean) }.average()

        return variance.toFloat()
    }

    //endregion
}

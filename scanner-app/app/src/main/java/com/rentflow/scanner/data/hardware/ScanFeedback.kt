package com.rentflow.scanner.data.hardware

import android.content.Context
import android.media.AudioManager
import android.media.ToneGenerator
import android.os.VibrationEffect
import android.os.Vibrator
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Provides audio and vibration feedback for scan events.
 * Wraps all calls in try-catch since some devices may lack a vibrator
 * or have restricted audio access.
 */
@Singleton
class ScanFeedback @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val vibrator: Vibrator? = try {
        context.getSystemService(Context.VIBRATOR_SERVICE) as? Vibrator
    } catch (_: Exception) {
        null
    }

    private val toneGenerator: ToneGenerator? = try {
        // 80% volume → 80 on the 0-100 scale
        ToneGenerator(AudioManager.STREAM_NOTIFICATION, 80)
    } catch (_: Exception) {
        null
    }

    /**
     * Short vibration (100 ms) + success tone.
     */
    fun onScanSuccess() {
        try {
            vibrator?.vibrate(VibrationEffect.createOneShot(100, VibrationEffect.DEFAULT_AMPLITUDE))
        } catch (_: Exception) { }

        try {
            toneGenerator?.startTone(ToneGenerator.TONE_PROP_ACK, 150)
        } catch (_: Exception) { }
    }

    /**
     * Longer vibration pattern (100 ms pause, 200 ms, 100 ms, 200 ms) + error tone.
     */
    fun onScanError() {
        try {
            // Pattern: wait 100ms, vibrate 200ms, wait 100ms, vibrate 200ms
            val pattern = longArrayOf(100, 200, 100, 200)
            vibrator?.vibrate(VibrationEffect.createWaveform(pattern, -1))
        } catch (_: Exception) { }

        try {
            toneGenerator?.startTone(ToneGenerator.TONE_PROP_NACK, 200)
        } catch (_: Exception) { }
    }

    /**
     * Very short vibration (50 ms) + subtle click tone for each RFID tag detected.
     */
    fun onRfidTagFound() {
        try {
            vibrator?.vibrate(VibrationEffect.createOneShot(50, VibrationEffect.DEFAULT_AMPLITUDE))
        } catch (_: Exception) { }

        try {
            toneGenerator?.startTone(ToneGenerator.TONE_CDMA_PIP, 100)
        } catch (_: Exception) { }
    }

    /**
     * Double vibration (100 ms on, 100 ms pause, 100 ms on) + success tone for RFID write confirmation.
     */
    fun onRfidWriteSuccess() {
        try {
            // Pattern: no initial delay, vibrate 100ms, pause 100ms, vibrate 100ms
            val pattern = longArrayOf(0, 100, 100, 100)
            vibrator?.vibrate(VibrationEffect.createWaveform(pattern, -1))
        } catch (_: Exception) { }

        try {
            toneGenerator?.startTone(ToneGenerator.TONE_PROP_ACK, 150)
        } catch (_: Exception) { }
    }

    /**
     * Release the [ToneGenerator] resources. Call when the feedback instance is no longer needed.
     */
    fun destroy() {
        try {
            toneGenerator?.release()
        } catch (_: Exception) { }
    }
}

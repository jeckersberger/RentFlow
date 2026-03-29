package com.rentflow.scanner.data.hardware

import android.device.DeviceManager
import android.util.Log

/**
 * POGO pin power control for the CF-H906 UHF module.
 * Simplified from the SDK's OtgUtils.java — only the node paths
 * relevant to the CF-H906 are included.
 */
object OtgUtils {
    private const val TAG = "OtgUtils"

    fun setPOGOPINEnable(enable: Boolean): Boolean {
        return try {
            val deviceManager = DeviceManager()
            val node5v = deviceManager.getSettingProperty("persist.sys.pogopin.otg5v.en")

            if (!node5v.isNullOrEmpty()) {
                val bytes = if (enable) byteArrayOf(0x31) else byteArrayOf(0x30)
                java.io.FileOutputStream(node5v).use { it.write(bytes) }
                Log.d(TAG, "POGO PIN ${if (enable) "enabled" else "disabled"} via $node5v")
            } else {
                Log.w(TAG, "No POGO PIN node found, skipping power control")
            }
            true
        } catch (e: Exception) {
            Log.e(TAG, "POGO PIN control failed: ${e.message}")
            false
        }
    }
}

package de.cratedesk.scanner.data.hardware

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.util.Log

/**
 * Receives barcode/RFID scan results from the Urovo DataWedge service.
 * The hardware scan button triggers DataWedge, which sends a broadcast
 * with the scanned data.
 *
 * Supports multiple Urovo/UBX intent formats:
 * - android.intent.action.DECODE_DATA (standard Urovo)
 * - com.ubx.datawedge.RESULT_ACTION
 * - android.intent.ACTION_DECODE_DATA
 */
class DataWedgeReceiver(
    private val onScan: (barcode: String, symbology: String) -> Unit,
) : BroadcastReceiver() {

    companion object {
        private const val TAG = "DataWedgeReceiver"

        // Urovo / UBX standard intent actions
        private val SCAN_ACTIONS = listOf(
            "android.intent.action.DECODE_DATA",
            "android.intent.ACTION_DECODE_DATA",
            "com.ubx.datawedge.RESULT_ACTION",
            "com.ubx.datawedge.DATA_RESULT",
        )

        // Data keys used by various Urovo firmware versions
        private val BARCODE_KEYS = listOf(
            "barcode_string",       // Urovo standard
            "decode_data_string",   // Alternative
            "com.ubx.datawedge.data_string", // DataWedge
            "data",                 // Generic
            "SCAN_BARCODE1",        // Some models
        )

        private val SYMBOLOGY_KEYS = listOf(
            "barcode_type",
            "decode_type",
            "com.ubx.datawedge.label_type",
            "SCAN_BARCODE_TYPE",
        )

        fun getIntentFilter(): IntentFilter {
            return IntentFilter().apply {
                SCAN_ACTIONS.forEach { addAction(it) }
            }
        }
    }

    override fun onReceive(context: Context?, intent: Intent?) {
        if (intent == null) return

        val action = intent.action ?: return
        Log.d(TAG, "Received intent: $action")

        // Try to extract barcode from various keys
        var barcode: String? = null
        for (key in BARCODE_KEYS) {
            barcode = intent.getStringExtra(key)
            if (!barcode.isNullOrBlank()) {
                Log.d(TAG, "Found barcode in key '$key': $barcode")
                break
            }
        }

        // Also try byte array (some models send raw bytes)
        if (barcode.isNullOrBlank()) {
            val bytes = intent.getByteArrayExtra("decode_data")
                ?: intent.getByteArrayExtra("barcode")
            if (bytes != null) {
                barcode = String(bytes, Charsets.UTF_8).trim()
                Log.d(TAG, "Found barcode from bytes: $barcode")
            }
        }

        if (barcode.isNullOrBlank()) {
            Log.w(TAG, "No barcode data found in intent extras: ${intent.extras?.keySet()}")
            return
        }

        // Try to get symbology
        var symbology = "UNKNOWN"
        for (key in SYMBOLOGY_KEYS) {
            val sym = intent.getStringExtra(key) ?: intent.getIntExtra(key, -1).let {
                if (it >= 0) it.toString() else null
            }
            if (sym != null) {
                symbology = sym.toString()
                break
            }
        }

        Log.i(TAG, "Scan result: $barcode (type: $symbology)")
        onScan(barcode, symbology)
    }
}

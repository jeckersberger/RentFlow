package com.rentflow.scanner.data.hardware

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.device.ScanManager
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow

/**
 * Receives barcode scan results from the CF-H906 built-in ScanManager.
 * The ScanManager sends broadcasts with action ACTION_DECODE when in intent output mode.
 */
class BroadcastScannerReceiver : BroadcastReceiver() {
    private val _scanEvents = MutableSharedFlow<BarcodeScanEvent>(extraBufferCapacity = 10)
    val scanEvents: Flow<BarcodeScanEvent> = _scanEvents

    override fun onReceive(context: Context?, intent: Intent?) {
        intent ?: return
        // Try SDK constant first, then fallback to common extras
        val barcode = intent.getStringExtra(ScanManager.BARCODE_STRING_TAG)
            ?: intent.getStringExtra("barcode_string")
            ?: return
        val barcodeType = intent.getByteExtra(ScanManager.BARCODE_TYPE_TAG, 0.toByte())
        _scanEvents.tryEmit(BarcodeScanEvent(barcode.trim(), barcodeType.toString()))
    }

    fun getIntentFilter(): IntentFilter {
        val filter = IntentFilter()
        filter.addAction(ScanManager.ACTION_DECODE)
        filter.addAction("android.intent.ACTION_DECODE_DATA")
        return filter
    }
}

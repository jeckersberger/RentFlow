package com.rentflow.scanner.data.hardware

import kotlinx.coroutines.flow.Flow

data class BarcodeScanEvent(val barcode: String, val format: String)
data class RfidReadEvent(val epc: String, val tid: String = "", val rssi: Int)
data class RfidWriteResult(val success: Boolean, val error: String? = null)

interface HardwareScanner {
    val barcodeScanEvents: Flow<BarcodeScanEvent>
    val rfidReadEvents: Flow<RfidReadEvent>
    /** Emits true when hardware trigger is pressed, false when released */
    val triggerState: Flow<Boolean>

    fun initBarcodeScan()
    fun startBarcodeScan()
    fun stopBarcodeScan()
    fun closeBarcodeScan()
    fun startRfidRead()
    fun startRfidBulkRead()
    fun stopRfid()
    fun closeRfid()
    suspend fun readTid(epc: String): String?
    fun isRfidAvailable(): Boolean
    fun destroy()
}

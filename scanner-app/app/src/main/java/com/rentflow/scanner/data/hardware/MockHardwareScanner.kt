package com.rentflow.scanner.data.hardware

import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import javax.inject.Inject

class MockHardwareScanner @Inject constructor() : HardwareScanner {
    private val _barcodeScanEvents = MutableSharedFlow<BarcodeScanEvent>()
    override val barcodeScanEvents: Flow<BarcodeScanEvent> = _barcodeScanEvents

    private val _rfidReadEvents = MutableSharedFlow<RfidReadEvent>()
    override val rfidReadEvents: Flow<RfidReadEvent> = _rfidReadEvents

    private val _triggerState = MutableSharedFlow<Boolean>()
    override val triggerState: Flow<Boolean> = _triggerState

    private val mockBarcodes = listOf("EQ-001", "EQ-002", "EQ-003", "EQ-004", "EQ-005")
    private val mockTids = listOf("E2801160200000", "E2801160200001", "E2801160200002")
    private var scanIndex = 0

    override fun initBarcodeScan() {}
    override fun startBarcodeScan() {}
    override fun stopBarcodeScan() {}
    override fun closeBarcodeScan() {}
    override fun startRfidRead() {}
    override fun startRfidBulkRead() {}
    override fun stopRfid() {}
    override fun closeRfid() {}

    override suspend fun readTid(epc: String): String? {
        delay(200)
        return "E2801160MOCK${epc.hashCode().toString(16).uppercase().takeLast(8)}"
    }

    override fun isRfidAvailable(): Boolean = true
    override fun destroy() {}

    suspend fun simulateBarcodeScan(barcode: String? = null) {
        val code = barcode ?: mockBarcodes[scanIndex++ % mockBarcodes.size]
        _barcodeScanEvents.emit(BarcodeScanEvent(code, "QR_CODE"))
    }

    suspend fun simulateRfidRead(epc: String? = null, rssi: Int = -45) {
        val tag = epc ?: "E200${mockTids.random()}"
        val tid = "E2801160MOCK${tag.hashCode().toString(16).uppercase().takeLast(8)}"
        _rfidReadEvents.emit(RfidReadEvent(tag, tid, rssi))
    }
}

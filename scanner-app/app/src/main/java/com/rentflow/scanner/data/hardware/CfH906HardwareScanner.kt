package com.rentflow.scanner.data.hardware

import android.content.Context
import android.device.ScanManager
import android.device.scanner.configuration.PropertyID
import android.device.scanner.configuration.Triggering
import android.util.Log
import com.rfid.trans.ReadTag
import com.rfid.trans.ReaderHelp
import com.rfid.trans.TagCallback
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Real hardware scanner implementation for the CF-H906 UHF PDA.
 *
 * Barcode: Uses android.device.ScanManager (built-in barcode engine).
 *          Receives results via BroadcastReceiver in intent output mode.
 *
 * UHF RFID: Uses com.rfid.trans.ReaderHelp (rfiddrive-release.aar).
 *           Connects via serial port /dev/ttyHSL0 at 115200 baud.
 *           Tags arrive via TagCallback on background thread.
 */
class CfH906HardwareScanner(
    private val context: Context,
) : HardwareScanner {

    companion object {
        private const val TAG = "CfH906Scanner"
        private const val RFID_SERIAL_PORT = "/dev/ttyHSL0"
        private const val RFID_BAUD_RATE = 115200
        private const val RFID_DEFAULT_PASSWORD = "00000000"
    }

    // --- Barcode ---
    private var scanManager: ScanManager? = null
    private val broadcastReceiver = BroadcastScannerReceiver()
    private var isBarcodeOpen = false

    override val barcodeScanEvents: Flow<BarcodeScanEvent> = broadcastReceiver.scanEvents

    // --- Trigger State ---
    private val _triggerState = MutableSharedFlow<Boolean>(extraBufferCapacity = 10)
    override val triggerState: Flow<Boolean> = _triggerState

    // --- UHF RFID ---
    private val rfidReader by lazy { ReaderHelp() }
    private var isRfidConnected = false

    private val _rfidReadEvents = MutableSharedFlow<RfidReadEvent>(extraBufferCapacity = 64)
    override val rfidReadEvents: Flow<RfidReadEvent> = _rfidReadEvents

    private var readTidInBulk = false

    private val rfidCallback = object : TagCallback {
        override fun tagCallback(tag: ReadTag?) {
            tag ?: return
            val epc = tag.epcId?.uppercase() ?: return
            val rssi = tag.rssi
            // In bulk mode: skip TID read for speed (TID is read separately during pairing)
            val tid = if (readTidInBulk) {
                try {
                    rfidReader.ReadData_G2(epc, 2.toByte(), 0, 6.toByte(), RFID_DEFAULT_PASSWORD)?.uppercase() ?: ""
                } catch (e: Exception) {
                    Log.w(TAG, "Could not read TID for EPC $epc: ${e.message}")
                    ""
                }
            } else ""
            _rfidReadEvents.tryEmit(RfidReadEvent(epc, tid, rssi))
        }

        override fun StopReadCallBack() {
            // Bulk read cycle ended — restart if still scanning
            if (isRfidReading) {
                try {
                    rfidReader.StartRead()
                    Log.d(TAG, "RFID bulk read restarted (continuous)")
                } catch (e: Exception) {
                    Log.w(TAG, "Could not restart RFID read: ${e.message}")
                    isRfidReading = false
                }
            } else {
                Log.d(TAG, "RFID continuous read stopped")
            }
        }
    }

    // ==================== Barcode Scanner ====================

    override fun initBarcodeScan() {
        try {
            if (scanManager == null) {
                scanManager = ScanManager()
            }
            if (!isBarcodeOpen) {
                val opened = scanManager?.openScanner() ?: false
                if (opened) {
                    isBarcodeOpen = true
                    // Switch to intent output mode (broadcasts result)
                    scanManager?.switchOutputMode(0)
                    // Trigger mode: HOST = we control start/stop via key events
                    try { scanManager?.setTriggerMode(Triggering.HOST) } catch (_: Exception) {}
                    // Disable keyboard wedge so we get broadcast instead
                    scanManager?.setParameterInts(
                        intArrayOf(PropertyID.WEDGE_KEYBOARD_ENABLE),
                        intArrayOf(0)
                    )
                    // Register receiver
                    context.registerReceiver(broadcastReceiver, broadcastReceiver.getIntentFilter())
                    Log.d(TAG, "Barcode scanner initialized in HOST mode (trigger keys pass through)")
                } else {
                    Log.e(TAG, "Failed to open barcode scanner")
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error initializing barcode scanner", e)
        }
    }

    override fun startBarcodeScan() {
        initBarcodeScan()
        try {
            scanManager?.startDecode()
        } catch (e: Exception) {
            Log.e(TAG, "Error starting barcode decode", e)
        }
    }

    override fun stopBarcodeScan() {
        try {
            scanManager?.stopDecode()
        } catch (e: Exception) {
            Log.e(TAG, "Error stopping barcode scan", e)
        }
    }

    override fun closeBarcodeScan() {
        try {
            if (isBarcodeOpen) {
                scanManager?.stopDecode()
                scanManager?.closeScanner()
                try { context.unregisterReceiver(broadcastReceiver) } catch (_: Exception) {}
                isBarcodeOpen = false
                scanManager = null
                Log.d(TAG, "Barcode scanner closed")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error closing barcode scanner", e)
        }
    }


    // ==================== UHF RFID ====================

    private fun connectRfid(): Boolean {
        if (isRfidConnected) return true
        try {
            // Power on POGO pin for RFID module
            OtgUtils.setPOGOPINEnable(true)
            Thread.sleep(300) // Allow module to power up

            var result = rfidReader.Connect(RFID_SERIAL_PORT, RFID_BAUD_RATE, 1)
            if (result != 0) {
                // Fallback to 57600 baud
                result = rfidReader.Connect(RFID_SERIAL_PORT, 57600, 1)
            }
            if (result == 0) {
                isRfidConnected = true
                rfidReader.SetCallBack(rfidCallback)

                // Set maximum RF power for read + write (30 dBm)
                try {
                    rfidReader.SetRfPower(30)
                    rfidReader.SetWritePower(30)
                    Log.d(TAG, "RFID read/write power set to 30 dBm (max)")
                } catch (e: Exception) {
                    Log.w(TAG, "Could not set RF power: ${e.message}")
                }

                // Configure inventory session — Session 0 for best weak-tag sensitivity
                try {
                    val param = rfidReader.GetInventoryPatameter()
                    param.Session = 0 // Session 0: re-reads tags every round (better for weak tags)
                    rfidReader.SetInventoryPatameter(param)
                } catch (e: Exception) {
                    Log.w(TAG, "Could not set inventory params: ${e.message}")
                }

                Log.d(TAG, "RFID connected, reader type: ${rfidReader.GetReaderType()}")
                return true
            } else {
                Log.e(TAG, "RFID connect failed with code: $result")
                return false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error connecting RFID", e)
            return false
        }
    }

    override fun startRfidRead() {
        if (!connectRfid()) return
        try {
            rfidReader.ScanRfid()
        } catch (e: Exception) {
            Log.e(TAG, "Error starting single RFID read", e)
        }
    }

    private var isRfidReading = false

    override fun startRfidBulkRead() {
        if (isRfidReading) return
        if (!connectRfid()) return
        try {
            rfidReader.StartRead()
            isRfidReading = true
            Log.d(TAG, "RFID bulk read started")
        } catch (e: Exception) {
            Log.e(TAG, "Error starting bulk RFID read", e)
        }
    }

    override fun stopRfid() {
        try {
            if (isRfidConnected && isRfidReading) {
                rfidReader.StopRead()
                isRfidReading = false
                Log.d(TAG, "RFID bulk read stopped")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error stopping RFID", e)
        }
    }

    override fun closeRfid() {
        try {
            if (isRfidConnected) {
                if (isRfidReading) {
                    rfidReader.StopRead()
                    isRfidReading = false
                }
                rfidReader.DisConnect()
                isRfidConnected = false
            }
            OtgUtils.setPOGOPINEnable(false)
            Log.d(TAG, "RFID closed")
        } catch (e: Exception) {
            Log.e(TAG, "Error closing RFID", e)
        }
    }

    override suspend fun readTid(epc: String): String? {
        if (!connectRfid()) return null
        return try {
            // Memory Bank 2 = TID, WordPtr 0, read 6 words (12 bytes)
            val tid = rfidReader.ReadData_G2(epc, 2.toByte(), 0, 6.toByte(), RFID_DEFAULT_PASSWORD)
            tid?.uppercase()
        } catch (e: Exception) {
            Log.e(TAG, "Error reading TID for EPC $epc", e)
            null
        }
    }

    override fun isRfidAvailable(): Boolean {
        return try {
            connectRfid()
        } catch (e: Exception) {
            false
        }
    }

    // ==================== Lifecycle ====================

    override fun destroy() {
        try {
            // Barcode
            if (isBarcodeOpen) {
                scanManager?.stopDecode()
                scanManager?.closeScanner()
                try {
                    context.unregisterReceiver(broadcastReceiver)
                } catch (_: Exception) {}
                isBarcodeOpen = false
            }
            // RFID
            if (isRfidConnected) {
                rfidReader.StopRead()
                rfidReader.DisConnect()
                isRfidConnected = false
            }
            OtgUtils.setPOGOPINEnable(false)
            Log.d(TAG, "Scanner destroyed, resources released")
        } catch (e: Exception) {
            Log.e(TAG, "Error during destroy", e)
        }
    }
}

package com.rentflow.scanner

import android.os.Bundle
import android.view.KeyEvent
import androidx.appcompat.app.AppCompatActivity
import androidx.activity.compose.setContent
import androidx.compose.runtime.*
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import com.rentflow.scanner.data.hardware.HardwareScanner
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.repository.AuthRepository
import com.rentflow.scanner.data.service.FindMyScannerService
import com.rentflow.scanner.data.service.FindMyScannerWorker
import com.rentflow.scanner.data.service.LockState
import com.rentflow.scanner.data.service.SessionTimeoutManager
import com.rentflow.scanner.ui.findme.FindMeOverlay
import com.rentflow.scanner.ui.navigation.AppNavigation
import com.rentflow.scanner.ui.navigation.Routes
import com.rentflow.scanner.ui.theme.RentFlowScannerTheme
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.flow.first
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : AppCompatActivity() {
    @Inject lateinit var authRepository: AuthRepository
    @Inject lateinit var findMyScannerService: FindMyScannerService
    @Inject lateinit var sessionTimeoutManager: SessionTimeoutManager
    @Inject lateinit var hardwareScanner: HardwareScanner
    @Inject lateinit var settingsDataStore: SettingsDataStore

    private var isRfidScanning = false
    private var isTriggerHeld = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val startDest = if (authRepository.isLoggedIn()) Routes.HOME else Routes.LOGIN

        // Start polling for ring commands
        FindMyScannerWorker.startPolling(this)

        // Track lifecycle for session timeout
        lifecycle.addObserver(LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_RESUME -> sessionTimeoutManager.onAppForeground()
                Lifecycle.Event.ON_PAUSE -> sessionTimeoutManager.onAppBackground()
                else -> {}
            }
        })

        setContent {
            RentFlowScannerTheme {
                var isRinging by remember { mutableStateOf(false) }
                val lockState by sessionTimeoutManager.lockState.collectAsState()

                // Check ring state periodically
                LaunchedEffect(Unit) {
                    while (true) {
                        isRinging = findMyScannerService.isRinging()
                        kotlinx.coroutines.delay(500)
                    }
                }

                // Check inactivity periodically
                LaunchedEffect(Unit) {
                    while (true) {
                        sessionTimeoutManager.checkInactivity()
                        kotlinx.coroutines.delay(60_000)
                    }
                }

                when {
                    isRinging -> {
                        FindMeOverlay(onStop = {
                            findMyScannerService.stop()
                            isRinging = false
                        })
                    }
                    lockState == LockState.FULL_RELOGIN -> {
                        // Full re-login required — show login screen
                        LaunchedEffect(Unit) {
                            authRepository.logout()
                        }
                        AppNavigation(startDestination = Routes.LOGIN)
                    }
                    else -> {
                        AppNavigation(startDestination = startDest)
                    }
                }
            }
        }
    }

    private fun isTriggerKey(keyCode: Int): Boolean =
        keyCode == 523 || keyCode == 520 || keyCode == 521 ||
        keyCode == KeyEvent.KEYCODE_F9 || keyCode == KeyEvent.KEYCODE_F10

    override fun onKeyDown(keyCode: Int, event: KeyEvent?): Boolean {
        if (!isTriggerKey(keyCode)) return super.onKeyDown(keyCode, event)
        if (isTriggerHeld) return true // Already handling

        isTriggerHeld = true
        val mode = kotlinx.coroutines.runBlocking { settingsDataStore.scanMode.first() }

        if (mode == SettingsDataStore.SCAN_MODE_RFID) {
            // RFID mode: kill any barcode decode immediately, then start RFID
            hardwareScanner.stopBarcodeScan()
            if (!isRfidScanning) {
                hardwareScanner.startRfidBulkRead()
                isRfidScanning = true
                android.util.Log.d("MainActivity", "RFID scan started (trigger down)")
            }
        } else {
            // Barcode mode: start decode (ScanManager in HOST mode)
            hardwareScanner.startBarcodeScan()
            android.util.Log.d("MainActivity", "Barcode decode started (trigger down)")
        }
        return true
    }

    override fun onKeyUp(keyCode: Int, event: KeyEvent?): Boolean {
        if (!isTriggerKey(keyCode)) return super.onKeyUp(keyCode, event)

        isTriggerHeld = false
        val mode = kotlinx.coroutines.runBlocking { settingsDataStore.scanMode.first() }

        if (mode == SettingsDataStore.SCAN_MODE_RFID) {
            // RFID mode: stop immediately when trigger released
            if (isRfidScanning) {
                hardwareScanner.stopRfid()
                isRfidScanning = false
                android.util.Log.d("MainActivity", "RFID scan stopped (trigger up)")
            }
        } else {
            // Barcode mode: stop decode
            hardwareScanner.stopBarcodeScan()
            android.util.Log.d("MainActivity", "Barcode decode stopped (trigger up)")
        }
        return true
    }

    override fun onUserInteraction() {
        super.onUserInteraction()
        sessionTimeoutManager.onUserActivity()
    }

    override fun onDestroy() {
        super.onDestroy()
        findMyScannerService.stop()
    }
}

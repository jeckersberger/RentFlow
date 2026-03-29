package com.rentflow.scanner.data.service

import com.rentflow.scanner.data.preferences.SettingsDataStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SessionTimeoutManager @Inject constructor(
    private val settingsDataStore: SettingsDataStore,
) {
    private val logoutTimeoutMs: Long
        get() = runBlocking { settingsDataStore.logoutTimeoutMinutes.first() } * 60 * 1000L

    private var lastActivityTimestamp: Long = System.currentTimeMillis()
    private var backgroundTimestamp: Long = 0L

    private val _lockState = MutableStateFlow(LockState.UNLOCKED)
    val lockState: StateFlow<LockState> = _lockState

    fun onUserActivity() {
        lastActivityTimestamp = System.currentTimeMillis()
    }

    fun onAppForeground() {
        if (backgroundTimestamp == 0L) return
        val elapsed = System.currentTimeMillis() - backgroundTimestamp
        backgroundTimestamp = 0L

        if (elapsed >= logoutTimeoutMs) {
            _lockState.value = LockState.FULL_RELOGIN
        }
    }

    fun onAppBackground() {
        backgroundTimestamp = System.currentTimeMillis()
    }

    fun checkInactivity() {
        if (_lockState.value != LockState.UNLOCKED) return
        val elapsed = System.currentTimeMillis() - lastActivityTimestamp
        if (elapsed >= logoutTimeoutMs) {
            _lockState.value = LockState.FULL_RELOGIN
        }
    }

    fun unlock() {
        _lockState.value = LockState.UNLOCKED
        lastActivityTimestamp = System.currentTimeMillis()
    }

    fun requireFullLogin() {
        _lockState.value = LockState.FULL_RELOGIN
    }

    fun reset() {
        _lockState.value = LockState.UNLOCKED
        lastActivityTimestamp = System.currentTimeMillis()
        backgroundTimestamp = 0L
    }
}

enum class LockState {
    UNLOCKED,
    FULL_RELOGIN,
}

package com.rentflow.scanner.data.preferences

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.intPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "settings")

@Singleton
class SettingsDataStore @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    val serverUrl: Flow<String> = context.dataStore.data.map { it[KEY_SERVER_URL] ?: DEFAULT_SERVER_URL }
    val language: Flow<String> = context.dataStore.data.map { it[KEY_LANGUAGE] ?: "de" }
    val scanMode: Flow<String> = context.dataStore.data.map { it[KEY_SCAN_MODE] ?: SCAN_MODE_BARCODE }
    val lockTimeoutMinutes: Flow<Int> = context.dataStore.data.map { it[KEY_LOCK_TIMEOUT] ?: DEFAULT_LOCK_TIMEOUT_MINUTES }
    val fullReloginHours: Flow<Int> = context.dataStore.data.map { it[KEY_FULL_RELOGIN] ?: DEFAULT_FULL_RELOGIN_HOURS }
    val logoutTimeoutMinutes: Flow<Int> = context.dataStore.data.map { it[KEY_LOGOUT_TIMEOUT] ?: DEFAULT_LOGOUT_TIMEOUT_MINUTES }

    val scannerHomeCards: Flow<List<String>> = context.dataStore.data.map { prefs ->
        val json = prefs[KEY_SCANNER_HOME_CARDS]
        if (json != null) {
            try {
                Gson().fromJson(json, object : TypeToken<List<String>>() {}.type)
            } catch (_: Exception) {
                DEFAULT_HOME_CARDS
            }
        } else {
            DEFAULT_HOME_CARDS
        }
    }

    val industryLabels: Flow<Map<String, String>> = context.dataStore.data.map { prefs ->
        val json = prefs[KEY_INDUSTRY_LABELS]
        if (json != null) {
            try {
                Gson().fromJson(json, object : TypeToken<Map<String, String>>() {}.type)
            } catch (_: Exception) {
                emptyMap()
            }
        } else {
            emptyMap()
        }
    }

    suspend fun setServerUrl(url: String) {
        context.dataStore.edit { it[KEY_SERVER_URL] = url }
    }

    suspend fun setLanguage(lang: String) {
        context.dataStore.edit { it[KEY_LANGUAGE] = lang }
    }

    suspend fun setScanMode(mode: String) {
        context.dataStore.edit { it[KEY_SCAN_MODE] = mode }
    }

    suspend fun setLockTimeout(minutes: Int) {
        context.dataStore.edit { it[KEY_LOCK_TIMEOUT] = minutes }
    }

    suspend fun setFullReloginTimeout(hours: Int) {
        context.dataStore.edit { it[KEY_FULL_RELOGIN] = hours }
    }

    suspend fun setLogoutTimeout(minutes: Int) {
        context.dataStore.edit { it[KEY_LOGOUT_TIMEOUT] = minutes }
    }

    suspend fun setScannerHomeCards(cards: List<String>) {
        context.dataStore.edit { it[KEY_SCANNER_HOME_CARDS] = Gson().toJson(cards) }
    }

    suspend fun setIndustryLabels(labels: Map<String, String>) {
        context.dataStore.edit { it[KEY_INDUSTRY_LABELS] = Gson().toJson(labels) }
    }

    companion object {
        const val DEFAULT_SERVER_URL = ""
        const val SCAN_MODE_BARCODE = "barcode"
        const val SCAN_MODE_RFID = "rfid"
        const val DEFAULT_LOCK_TIMEOUT_MINUTES = 30
        const val DEFAULT_FULL_RELOGIN_HOURS = 4
        const val DEFAULT_LOGOUT_TIMEOUT_MINUTES = 60
        val DEFAULT_HOME_CARDS = listOf(
            "scan_info", "checkout", "checkin", "inventory", "rfid_assign", "new_project"
        )
        private val KEY_SERVER_URL = stringPreferencesKey("server_url")
        private val KEY_LANGUAGE = stringPreferencesKey("language")
        private val KEY_SCAN_MODE = stringPreferencesKey("scan_mode")
        private val KEY_LOCK_TIMEOUT = intPreferencesKey("lock_timeout_minutes")
        private val KEY_FULL_RELOGIN = intPreferencesKey("full_relogin_hours")
        private val KEY_LOGOUT_TIMEOUT = intPreferencesKey("logout_timeout_minutes")
        private val KEY_SCANNER_HOME_CARDS = stringPreferencesKey("scanner_home_cards")
        private val KEY_INDUSTRY_LABELS = stringPreferencesKey("industry_labels")
    }
}

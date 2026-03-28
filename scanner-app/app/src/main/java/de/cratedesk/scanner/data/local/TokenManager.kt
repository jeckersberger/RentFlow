package de.cratedesk.scanner.data.local

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenManager @Inject constructor(
    private val dataStore: DataStore<Preferences>,
) {
    companion object {
        private val ACCESS_TOKEN = stringPreferencesKey("access_token")
        private val REFRESH_TOKEN = stringPreferencesKey("refresh_token")
        private val DEVICE_ID = stringPreferencesKey("device_id")
        private val SERVER_URL = stringPreferencesKey("server_url")
    }

    val accessToken: Flow<String?> = dataStore.data.map { it[ACCESS_TOKEN] }
    val deviceId: Flow<String?> = dataStore.data.map { it[DEVICE_ID] }
    val serverUrl: Flow<String?> = dataStore.data.map { it[SERVER_URL] }

    fun getTokenSync(): String? = runBlocking {
        dataStore.data.first()[ACCESS_TOKEN]
    }

    suspend fun saveTokens(accessToken: String, refreshToken: String) {
        dataStore.edit { prefs ->
            prefs[ACCESS_TOKEN] = accessToken
            prefs[REFRESH_TOKEN] = refreshToken
        }
    }

    suspend fun saveDeviceId(deviceId: String) {
        dataStore.edit { it[DEVICE_ID] = deviceId }
    }

    suspend fun saveServerUrl(url: String) {
        dataStore.edit { it[SERVER_URL] = url }
    }

    suspend fun clearAll() {
        dataStore.edit { it.clear() }
    }

    suspend fun isLoggedIn(): Boolean =
        dataStore.data.first()[ACCESS_TOKEN] != null
}

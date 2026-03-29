package com.rentflow.scanner.data.repository

import android.util.Log
import com.rentflow.scanner.data.api.ConfigApi
import com.rentflow.scanner.data.preferences.SettingsDataStore
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ConfigRepository @Inject constructor(
    private val configApi: ConfigApi,
    private val settingsDataStore: SettingsDataStore,
) {
    suspend fun fetchAndStoreIndustryConfig(): Result<Unit> {
        return try {
            val response = configApi.getIndustryConfig()
            if (response.isSuccessful && response.body()?.data != null) {
                val data = response.body()!!.data!!
                if (data.scanner_home_cards.isNotEmpty()) {
                    settingsDataStore.setScannerHomeCards(data.scanner_home_cards)
                }
                if (data.labels.isNotEmpty()) {
                    settingsDataStore.setIndustryLabels(data.labels)
                }
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to fetch industry config"))
            }
        } catch (e: Exception) {
            Log.w("ConfigRepository", "Could not fetch industry config (offline?)", e)
            Result.failure(e)
        }
    }
}

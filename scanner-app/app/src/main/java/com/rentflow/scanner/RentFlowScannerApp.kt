package com.rentflow.scanner

import android.app.Application
import android.content.res.Configuration
import androidx.appcompat.app.AppCompatDelegate
import androidx.core.os.LocaleListCompat
import androidx.hilt.work.HiltWorkerFactory
import com.rentflow.scanner.data.preferences.SettingsDataStore
import dagger.hilt.android.HiltAndroidApp
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltAndroidApp
class RentFlowScannerApp : Application(), androidx.work.Configuration.Provider {
    @Inject lateinit var workerFactory: HiltWorkerFactory
    @Inject lateinit var settingsDataStore: SettingsDataStore

    override val workManagerConfiguration: androidx.work.Configuration
        get() = androidx.work.Configuration.Builder()
            .setWorkerFactory(workerFactory)
            .build()

    override fun onCreate() {
        super.onCreate()
        // Apply saved language on app start
        CoroutineScope(Dispatchers.Main).launch {
            val lang = settingsDataStore.language.first()
            applyLanguage(lang)
        }
    }

    companion object {
        fun applyLanguage(lang: String) {
            val localeList = LocaleListCompat.forLanguageTags(lang)
            AppCompatDelegate.setApplicationLocales(localeList)
        }
    }
}

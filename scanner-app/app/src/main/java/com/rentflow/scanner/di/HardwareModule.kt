package com.rentflow.scanner.di

import android.content.Context
import android.os.Build
import com.rentflow.scanner.data.hardware.HardwareScanner
import com.rentflow.scanner.data.hardware.MockHardwareScanner
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object HardwareModule {
    @Provides
    @Singleton
    fun provideHardwareScanner(
        @ApplicationContext context: Context,
    ): HardwareScanner {
        val isEmulator = Build.FINGERPRINT.contains("generic") ||
            Build.FINGERPRINT.contains("emulator") ||
            Build.MODEL.contains("Emulator") ||
            Build.MODEL.contains("Android SDK built for x86") ||
            Build.MANUFACTURER.contains("Genymotion") ||
            Build.HARDWARE.contains("goldfish") ||
            Build.HARDWARE.contains("ranchu")

        return if (isEmulator) {
            MockHardwareScanner()
        } else {
            // Only instantiate real scanner on physical device
            // to avoid loading native libs on emulator
            try {
                com.rentflow.scanner.data.hardware.CfH906HardwareScanner(context)
            } catch (e: Exception) {
                android.util.Log.e("HardwareModule", "Failed to init CfH906, falling back to mock", e)
                MockHardwareScanner()
            }
        }
    }
}

package com.rentflow.scanner.data.service

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.*
import com.rentflow.scanner.data.api.ScannerApi
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import java.util.concurrent.TimeUnit

@HiltWorker
class FindMyScannerWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted params: WorkerParameters,
    private val scannerApi: ScannerApi,
    private val findMyScannerService: FindMyScannerService,
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        return try {
            val deviceId = android.os.Build.SERIAL ?: android.provider.Settings.Secure.getString(
                applicationContext.contentResolver, android.provider.Settings.Secure.ANDROID_ID
            )
            val response = scannerApi.checkRingCommand(deviceId)
            if (response.isSuccessful) {
                val shouldRing = response.body()?.data?.ring == true
                if (shouldRing) {
                    findMyScannerService.ring()
                }
            }
            Result.success()
        } catch (e: Exception) {
            Result.success()
        }
    }

    companion object {
        private const val WORK_NAME = "find_my_scanner_poll"

        fun startPolling(context: Context) {
            val request = PeriodicWorkRequestBuilder<FindMyScannerWorker>(
                15, TimeUnit.SECONDS,
            ).setConstraints(
                Constraints.Builder()
                    .setRequiredNetworkType(NetworkType.CONNECTED)
                    .build()
            ).build()

            WorkManager.getInstance(context).enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.KEEP,
                request,
            )
        }

        fun stopPolling(context: Context) {
            WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
        }
    }
}

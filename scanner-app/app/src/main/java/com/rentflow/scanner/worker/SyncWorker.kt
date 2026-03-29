package com.rentflow.scanner.worker

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.*
import com.rentflow.scanner.data.api.ScanRequest
import com.rentflow.scanner.data.api.ScannerApi
import com.rentflow.scanner.data.db.PendingScanDao
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import java.util.concurrent.TimeUnit

@HiltWorker
class SyncWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted params: WorkerParameters,
    private val scannerApi: ScannerApi,
    private val pendingScanDao: PendingScanDao,
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        val pending = pendingScanDao.getPending()
        if (pending.isEmpty()) return Result.success()

        val requests = pending.map { scan ->
            ScanRequest(
                barcode = scan.barcode,
                scan_type = scan.scanType,
                user_id = scan.userId,
                device_id = scan.deviceId,
                project_id = scan.projectId,
                notes = scan.notes,
            )
        }

        return try {
            val response = scannerApi.syncOffline(requests)
            if (response.isSuccessful) {
                pending.forEach { pendingScanDao.delete(it) }
                Result.success()
            } else {
                pending.forEach { pendingScanDao.incrementRetry(it.id) }
                Result.retry()
            }
        } catch (e: Exception) {
            pending.forEach { pendingScanDao.incrementRetry(it.id) }
            Result.retry()
        }
    }

    companion object {
        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()
            val request = OneTimeWorkRequestBuilder<SyncWorker>()
                .setConstraints(constraints)
                .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 30, TimeUnit.SECONDS)
                .build()
            WorkManager.getInstance(context).enqueueUniqueWork(
                "sync_scans",
                ExistingWorkPolicy.REPLACE,
                request,
            )
        }
    }
}

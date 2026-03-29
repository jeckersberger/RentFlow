package com.rentflow.scanner.data.service

import android.app.DownloadManager
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.util.Log
import android.widget.Toast
import androidx.core.content.FileProvider
import com.rentflow.scanner.BuildConfig
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.withContext
import java.io.File
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

data class UpdateInfo(
    val versionName: String,
    val versionCode: Int,
    val downloadUrl: String,
    val changelog: String,
    val size: Long = 0,
)

@Singleton
class AppUpdateService @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    companion object {
        private const val TAG = "AppUpdateService"
        private const val GITHUB_REPO = "jeckersberger/RentFlow_scanner"
        private const val RELEASES_URL = "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    }

    private val _updateAvailable = MutableStateFlow<UpdateInfo?>(null)
    val updateAvailable: StateFlow<UpdateInfo?> = _updateAvailable

    private val _downloadState = MutableStateFlow(DownloadState.IDLE)
    val downloadState: StateFlow<DownloadState> = _downloadState

    private var downloadId: Long = -1
    private var downloadReceiver: BroadcastReceiver? = null

    suspend fun checkForUpdate(): UpdateInfo? = withContext(Dispatchers.IO) {
        try {
            val connection = URL(RELEASES_URL).openConnection() as HttpURLConnection
            connection.setRequestProperty("Accept", "application/vnd.github.v3+json")
            connection.connectTimeout = 10_000
            connection.readTimeout = 10_000

            if (connection.responseCode != 200) {
                Log.w(TAG, "GitHub API returned ${connection.responseCode}")
                return@withContext null
            }

            val response = connection.inputStream.bufferedReader().readText()
            val json = org.json.JSONObject(response)

            val tagName = json.getString("tag_name")
            val changelog = json.optString("body", "")
            val assets = json.getJSONArray("assets")

            var apkUrl: String? = null
            var apkSize: Long = 0
            for (i in 0 until assets.length()) {
                val asset = assets.getJSONObject(i)
                if (asset.getString("name").endsWith(".apk")) {
                    apkUrl = asset.getString("browser_download_url")
                    apkSize = asset.getLong("size")
                    break
                }
            }

            if (apkUrl == null) {
                Log.w(TAG, "No APK found in release $tagName")
                return@withContext null
            }

            val remoteVersion = tagName.removePrefix("v")
            val currentVersion = BuildConfig.VERSION_NAME

            if (isNewerVersion(remoteVersion, currentVersion)) {
                val info = UpdateInfo(
                    versionName = remoteVersion,
                    versionCode = parseVersionCode(remoteVersion),
                    downloadUrl = apkUrl,
                    changelog = changelog,
                    size = apkSize,
                )
                _updateAvailable.value = info
                Log.d(TAG, "Update available: $currentVersion -> $remoteVersion")
                return@withContext info
            } else {
                Log.d(TAG, "App is up to date ($currentVersion)")
                _updateAvailable.value = null
                return@withContext null
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to check for updates", e)
            return@withContext null
        }
    }

    fun downloadAndInstall(updateInfo: UpdateInfo) {
        if (_downloadState.value == DownloadState.DOWNLOADING) return

        try {
            val downloadManager = context.getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager

            // Clean up old APKs
            val downloadDir = context.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS)
            downloadDir?.listFiles()?.filter { it.name.endsWith(".apk") }?.forEach { it.delete() }

            val request = DownloadManager.Request(Uri.parse(updateInfo.downloadUrl))
                .setTitle("EquipFlow Scanner v${updateInfo.versionName}")
                .setDescription("Update wird heruntergeladen...")
                .setDestinationInExternalFilesDir(
                    context,
                    Environment.DIRECTORY_DOWNLOADS,
                    "rentflow-scanner-${updateInfo.versionName}.apk"
                )
                .setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
                .setAllowedOverMetered(true)
                .setAllowedOverRoaming(true)

            downloadId = downloadManager.enqueue(request)
            _downloadState.value = DownloadState.DOWNLOADING
            Log.d(TAG, "Download started, id=$downloadId, url=${updateInfo.downloadUrl}")

            showToast("Download gestartet...")

            // Unregister any previous receiver
            downloadReceiver?.let {
                try { context.unregisterReceiver(it) } catch (_: Exception) {}
            }

            // Register receiver for download complete
            downloadReceiver = object : BroadcastReceiver() {
                override fun onReceive(ctx: Context?, intent: Intent?) {
                    val id = intent?.getLongExtra(DownloadManager.EXTRA_DOWNLOAD_ID, -1)
                    if (id != downloadId) return

                    // Check download status
                    val query = DownloadManager.Query().setFilterById(downloadId)
                    val cursor = downloadManager.query(query)
                    if (cursor != null && cursor.moveToFirst()) {
                        val statusIndex = cursor.getColumnIndex(DownloadManager.COLUMN_STATUS)
                        val status = cursor.getInt(statusIndex)
                        cursor.close()

                        if (status == DownloadManager.STATUS_SUCCESSFUL) {
                            Log.d(TAG, "Download complete, installing...")
                            _downloadState.value = DownloadState.INSTALLING
                            showToast("Download abgeschlossen, installiere...")
                            installApk(updateInfo.versionName)
                        } else {
                            Log.e(TAG, "Download failed with status: $status")
                            _downloadState.value = DownloadState.FAILED
                            showToast("Download fehlgeschlagen (Status: $status)")
                        }
                    } else {
                        cursor?.close()
                        Log.e(TAG, "Download query returned no results")
                        _downloadState.value = DownloadState.FAILED
                        showToast("Download fehlgeschlagen")
                    }

                    try { context.unregisterReceiver(this) } catch (_: Exception) {}
                    downloadReceiver = null
                }
            }

            val filter = IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.registerReceiver(downloadReceiver, filter, Context.RECEIVER_EXPORTED)
            } else {
                context.registerReceiver(downloadReceiver, filter)
            }

        } catch (e: Exception) {
            Log.e(TAG, "Failed to start download", e)
            _downloadState.value = DownloadState.FAILED
            showToast("Download konnte nicht gestartet werden: ${e.message}")
        }
    }

    private fun installApk(version: String) {
        val file = File(
            context.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS),
            "rentflow-scanner-$version.apk"
        )
        if (!file.exists()) {
            Log.e(TAG, "Downloaded APK not found: ${file.absolutePath}")
            _downloadState.value = DownloadState.FAILED
            showToast("APK-Datei nicht gefunden")
            return
        }

        Log.d(TAG, "Installing APK: ${file.absolutePath} (${file.length()} bytes)")

        try {
            val uri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                file,
            )

            val intent = Intent(Intent.ACTION_VIEW).apply {
                setDataAndType(uri, "application/vnd.android.package-archive")
                addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            }
            context.startActivity(intent)
            _downloadState.value = DownloadState.IDLE
        } catch (e: Exception) {
            Log.e(TAG, "Failed to launch installer", e)
            _downloadState.value = DownloadState.FAILED
            showToast("Installation konnte nicht gestartet werden: ${e.message}")
        }
    }

    private fun showToast(message: String) {
        try {
            android.os.Handler(android.os.Looper.getMainLooper()).post {
                Toast.makeText(context, message, Toast.LENGTH_LONG).show()
            }
        } catch (_: Exception) {}
    }

    private fun isNewerVersion(remote: String, current: String): Boolean {
        val remoteParts = remote.split(".").map { it.toIntOrNull() ?: 0 }
        val currentParts = current.split(".").map { it.toIntOrNull() ?: 0 }
        for (i in 0 until maxOf(remoteParts.size, currentParts.size)) {
            val r = remoteParts.getOrElse(i) { 0 }
            val c = currentParts.getOrElse(i) { 0 }
            if (r > c) return true
            if (r < c) return false
        }
        return false
    }

    private fun parseVersionCode(version: String): Int {
        val parts = version.split(".").map { it.toIntOrNull() ?: 0 }
        // major * 10000 + minor * 100 + patch → e.g. 2.0.1 = 20001
        return parts.getOrElse(0) { 0 } * 10000 +
               parts.getOrElse(1) { 0 } * 100 +
               parts.getOrElse(2) { 0 }
    }
}

enum class DownloadState {
    IDLE, DOWNLOADING, INSTALLING, FAILED
}

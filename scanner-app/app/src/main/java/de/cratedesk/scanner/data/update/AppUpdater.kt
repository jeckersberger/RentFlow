package de.cratedesk.scanner.data.update

import android.app.DownloadManager
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.util.Log
import androidx.core.content.FileProvider
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import java.io.File
import java.net.URL

/**
 * Checks GitHub Releases for new APK versions and downloads/installs updates.
 */
class AppUpdater(
    private val context: Context,
    private val githubOwner: String = "jeckersberger",
    private val githubRepo: String = "RentFlow_scanner",
) {
    companion object {
        private const val TAG = "AppUpdater"
    }

    data class ReleaseInfo(
        val tagName: String,
        val versionCode: Int,
        val downloadUrl: String,
        val releaseNotes: String,
    )

    /**
     * Check GitHub for the latest release. Returns null if no update available.
     */
    suspend fun checkForUpdate(): ReleaseInfo? = withContext(Dispatchers.IO) {
        try {
            val url = "https://api.github.com/repos/$githubOwner/$githubRepo/releases/latest"
            val json = URL(url).readText()
            val release = org.json.JSONObject(json)

            val tagName = release.getString("tag_name")
            val remoteVersion = tagName.removePrefix("v").split(".").let { parts ->
                val major = parts.getOrNull(0)?.toIntOrNull() ?: 0
                val minor = parts.getOrNull(1)?.toIntOrNull() ?: 0
                val patch = parts.getOrNull(2)?.toIntOrNull() ?: 0
                major * 10000 + minor * 100 + patch
            }

            val currentVersion = try {
                context.packageManager.getPackageInfo(context.packageName, 0).let {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
                        it.longVersionCode.toInt()
                    } else {
                        @Suppress("DEPRECATION")
                        it.versionCode
                    }
                }
            } catch (e: Exception) {
                0
            }

            Log.d(TAG, "Current: $currentVersion, Remote: $remoteVersion ($tagName)")

            if (remoteVersion <= currentVersion) {
                return@withContext null
            }

            // Find APK asset
            val assets = release.getJSONArray("assets")
            var apkUrl: String? = null
            for (i in 0 until assets.length()) {
                val asset = assets.getJSONObject(i)
                val name = asset.getString("name")
                if (name.endsWith(".apk")) {
                    apkUrl = asset.getString("browser_download_url")
                    break
                }
            }

            if (apkUrl == null) {
                Log.w(TAG, "No APK found in release $tagName")
                return@withContext null
            }

            ReleaseInfo(
                tagName = tagName,
                versionCode = remoteVersion,
                downloadUrl = apkUrl,
                releaseNotes = release.optString("body", ""),
            )
        } catch (e: Exception) {
            Log.e(TAG, "Update check failed", e)
            null
        }
    }

    /**
     * Download and install the APK update.
     */
    fun downloadAndInstall(release: ReleaseInfo) {
        val downloadManager = context.getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager

        val request = DownloadManager.Request(Uri.parse(release.downloadUrl))
            .setTitle("CrateDesk Scanner Update")
            .setDescription("Version ${release.tagName}")
            .setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
            .setDestinationInExternalPublicDir(
                Environment.DIRECTORY_DOWNLOADS,
                "cratedesk-scanner-${release.tagName}.apk"
            )

        val downloadId = downloadManager.enqueue(request)
        Log.i(TAG, "Download started: $downloadId")

        // Listen for download completion
        val receiver = object : BroadcastReceiver() {
            override fun onReceive(ctx: Context?, intent: Intent?) {
                val id = intent?.getLongExtra(DownloadManager.EXTRA_DOWNLOAD_ID, -1)
                if (id == downloadId) {
                    installApk(release.tagName)
                    context.unregisterReceiver(this)
                }
            }
        }

        context.registerReceiver(
            receiver,
            IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE),
            Context.RECEIVER_EXPORTED,
        )
    }

    private fun installApk(tagName: String) {
        val apkFile = File(
            Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS),
            "cratedesk-scanner-$tagName.apk"
        )

        if (!apkFile.exists()) {
            Log.e(TAG, "APK file not found: ${apkFile.absolutePath}")
            return
        }

        val intent = Intent(Intent.ACTION_VIEW).apply {
            val uri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                apkFile,
            )
            setDataAndType(uri, "application/vnd.android.package-archive")
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }

        context.startActivity(intent)
    }
}

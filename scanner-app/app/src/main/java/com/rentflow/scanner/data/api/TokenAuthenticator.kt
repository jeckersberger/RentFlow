package com.rentflow.scanner.data.api

import android.util.Log
import com.google.gson.Gson
import com.rentflow.scanner.data.preferences.SettingsDataStore
import com.rentflow.scanner.data.preferences.TokenManager
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody

class TokenAuthenticator(
    private val tokenManager: TokenManager,
    private val settingsDataStore: SettingsDataStore,
) : Authenticator {
    private val gson = Gson()
    private val lock = Any()

    override fun authenticate(route: Route?, response: Response): Request? {
        // Don't retry if already failed 3 times
        if (responseCount(response) >= 3) {
            Log.w(TAG, "Token refresh failed after 3 attempts")
            return null
        }

        val refreshToken = tokenManager.getRefreshToken() ?: return null

        synchronized(lock) {
            // Check if another thread already refreshed
            val currentToken = tokenManager.getAccessToken()
            val requestToken = response.request.header("Authorization")?.removePrefix("Bearer ")
            if (currentToken != null && currentToken != requestToken) {
                return response.request.newBuilder()
                    .header("Authorization", "Bearer $currentToken")
                    .build()
            }

            // Refresh the token
            return try {
                val refreshBody = gson.toJson(RefreshRequest(refreshToken))
                    .toRequestBody("application/json".toMediaType())
                val baseUrl = runBlocking {
                    val url = settingsDataStore.serverUrl.first()
                    val base = url.ifBlank { "http://localhost" }
                    if (base.endsWith("/")) base else "$base/"
                }
                val refreshRequest = Request.Builder()
                    .url("${baseUrl}api/v1/auth/refresh")
                    .post(refreshBody)
                    .build()

                val client = OkHttpClient()
                val refreshResponse = client.newCall(refreshRequest).execute()

                if (!refreshResponse.isSuccessful) {
                    Log.w(TAG, "Refresh failed: ${refreshResponse.code}")
                    tokenManager.clearTokens()
                    return null
                }

                val body = refreshResponse.body?.string() ?: return null
                val apiResponse = gson.fromJson(body, ApiResponse::class.java)
                val dataJson = gson.toJson(apiResponse.data)
                val loginResponse = gson.fromJson(dataJson, LoginResponse::class.java)

                tokenManager.saveTokens(
                    loginResponse.access_token,
                    loginResponse.refresh_token,
                    tokenManager.getTenantId() ?: "",
                )
                Log.d(TAG, "Token refreshed successfully")

                response.request.newBuilder()
                    .header("Authorization", "Bearer ${loginResponse.access_token}")
                    .build()
            } catch (e: Exception) {
                Log.e(TAG, "Token refresh error", e)
                tokenManager.clearTokens()
                null
            }
        }
    }

    private fun responseCount(response: Response): Int {
        var count = 1
        var prior = response.priorResponse
        while (prior != null) {
            count++
            prior = prior.priorResponse
        }
        return count
    }

    companion object {
        private const val TAG = "TokenAuthenticator"
    }
}

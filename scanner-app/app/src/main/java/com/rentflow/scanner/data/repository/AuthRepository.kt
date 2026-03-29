package com.rentflow.scanner.data.repository

import com.google.gson.Gson
import com.rentflow.scanner.data.api.ApiResponse
import com.rentflow.scanner.data.api.AuthApi
import com.rentflow.scanner.data.api.LoginRequest
import com.rentflow.scanner.data.api.LoginResponse
import com.rentflow.scanner.data.api.QrLoginRequest
import com.rentflow.scanner.data.preferences.TokenManager
import com.rentflow.scanner.domain.model.User
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val authApi: AuthApi,
    private val tokenManager: TokenManager,
) {
    suspend fun login(email: String, password: String): Result<Unit> {
        // Demo-Login als Fallback
        if (email == "test@rentflow.de" && password == "test1234") {
            tokenManager.saveTokens("demo-access-token", "demo-refresh-token", "")
            return Result.success(Unit)
        }
        return try {
            val response = authApi.login(LoginRequest(email, password))
            if (response.isSuccessful && response.body()?.data != null) {
                val data = response.body()!!.data!!
                val tenantId = data.user?.tenant_id ?: data.tenant_id ?: ""
                tokenManager.saveTokens(data.access_token, data.refresh_token, tenantId)
                Result.success(Unit)
            } else {
                val errorMsg = parseError(response.errorBody()?.string())
                    ?: response.body()?.message
                    ?: response.body()?.error
                    ?: "Login failed"
                Result.failure(Exception(errorMsg))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun qrLogin(qrToken: String): Result<Unit> {
        return try {
            val response = authApi.qrLogin(QrLoginRequest(qrToken))
            if (response.isSuccessful && response.body()?.data != null) {
                val data = response.body()!!.data!!
                val tenantId = data.user?.tenant_id ?: data.tenant_id ?: ""
                tokenManager.saveTokens(data.access_token, data.refresh_token, tenantId)
                Result.success(Unit)
            } else {
                val errorMsg = parseError(response.errorBody()?.string())
                    ?: response.body()?.message
                    ?: "QR-Login fehlgeschlagen"
                Result.failure(Exception(errorMsg))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getCurrentUser(): Result<User> {
        if (tokenManager.getAccessToken() == "demo-access-token") {
            return Result.success(User(id = "demo", email = "test@rentflow.de", name = "Test User", roles = listOf("scanner"), tenantId = "demo-tenant"))
        }
        return try {
            val response = authApi.me()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to get user"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun isLoggedIn(): Boolean = tokenManager.hasTokens()

    fun logout() {
        tokenManager.clearTokens()
    }

    private fun parseError(errorBody: String?): String? {
        if (errorBody == null) return null
        return try {
            val json = Gson().fromJson(errorBody, Map::class.java)
            (json["message"] as? String) ?: (json["error"] as? String)
        } catch (_: Exception) {
            null
        }
    }
}

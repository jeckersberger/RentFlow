package com.rentflow.scanner.data.api

import com.rentflow.scanner.data.preferences.TokenManager
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject

class AuthInterceptor @Inject constructor(
    private val tokenManager: TokenManager,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request().newBuilder()
        tokenManager.getAccessToken()?.let {
            request.addHeader("Authorization", "Bearer $it")
        }
        tokenManager.getTenantId()?.let {
            request.addHeader("X-Tenant-ID", it)
        }
        return chain.proceed(request.build())
    }
}

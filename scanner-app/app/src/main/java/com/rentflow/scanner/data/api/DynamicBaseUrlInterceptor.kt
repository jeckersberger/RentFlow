package com.rentflow.scanner.data.api

import com.rentflow.scanner.data.preferences.SettingsDataStore
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.Interceptor
import okhttp3.Response

class DynamicBaseUrlInterceptor(
    private val settingsDataStore: SettingsDataStore,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val serverUrl = runBlocking { settingsDataStore.serverUrl.first() }
        if (serverUrl.isBlank()) return chain.proceed(request)

        val baseHttpUrl = serverUrl.trimEnd('/').toHttpUrlOrNull() ?: return chain.proceed(request)
        val newUrl = request.url.newBuilder()
            .scheme(baseHttpUrl.scheme)
            .host(baseHttpUrl.host)
            .port(baseHttpUrl.port)
            .build()

        return chain.proceed(request.newBuilder().url(newUrl).build())
    }
}

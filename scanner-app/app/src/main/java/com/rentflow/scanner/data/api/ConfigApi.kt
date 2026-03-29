package com.rentflow.scanner.data.api

import retrofit2.Response
import retrofit2.http.GET

interface ConfigApi {
    @GET("api/v1/config/industry")
    suspend fun getIndustryConfig(): Response<ApiResponse<IndustryConfigResponse>>
}

data class IndustryConfigResponse(
    val industry: String = "",
    val scanner_home_cards: List<String> = emptyList(),
    val labels: Map<String, String> = emptyMap(),
)

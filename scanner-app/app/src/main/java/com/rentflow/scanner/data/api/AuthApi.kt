package com.rentflow.scanner.data.api

import com.rentflow.scanner.domain.model.User
import retrofit2.Response
import retrofit2.http.*

interface AuthApi {
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<ApiResponse<LoginResponse>>

    @POST("api/v1/auth/refresh")
    suspend fun refresh(@Body request: RefreshRequest): Response<ApiResponse<LoginResponse>>

    @POST("api/v1/auth/qr-login")
    suspend fun qrLogin(@Body request: QrLoginRequest): Response<ApiResponse<LoginResponse>>

    @GET("api/v1/auth/me")
    suspend fun me(): Response<ApiResponse<User>>
}

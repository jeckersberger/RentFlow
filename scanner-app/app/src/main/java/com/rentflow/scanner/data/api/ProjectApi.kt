package com.rentflow.scanner.data.api

import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.domain.model.Project
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Path
import retrofit2.http.Query

interface ProjectApi {
    @GET("api/v1/projects")
    suspend fun listProjects(@Query("search") search: String? = null): Response<ApiResponse<List<Project>>>

    @GET("api/v1/projects?status=active&sort=start_date")
    suspend fun listActiveProjects(): Response<ApiResponse<List<Project>>>

    @GET("api/v1/projects?filter=returning_today")
    suspend fun listReturningToday(): Response<ApiResponse<List<Project>>>

    @GET("api/v1/projects?filter=checked_out")
    suspend fun listCheckedOut(): Response<ApiResponse<List<Project>>>

    @GET("api/v1/projects/{id}/equipment")
    suspend fun listProjectEquipment(@Path("id") projectId: String): Response<ApiResponse<List<Equipment>>>
}

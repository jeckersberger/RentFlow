package com.rentflow.scanner.data.repository

import com.rentflow.scanner.data.api.ProjectApi
import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.domain.model.Project
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ProjectRepository @Inject constructor(
    private val projectApi: ProjectApi,
) {
    suspend fun listProjects(search: String? = null): Result<List<Project>> {
        return try {
            val response = projectApi.listProjects(search)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to load projects"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun listActiveProjects(): Result<List<Project>> {
        return try {
            val response = projectApi.listActiveProjects()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                listProjects()
            }
        } catch (e: Exception) {
            listProjects()
        }
    }

    suspend fun listReturningToday(): Result<List<Project>> {
        return try {
            val response = projectApi.listReturningToday()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to load returning projects"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun listCheckedOut(): Result<List<Project>> {
        return try {
            val response = projectApi.listCheckedOut()
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to load checked out projects"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun listProjectEquipment(projectId: String): Result<List<Equipment>> {
        return try {
            val response = projectApi.listProjectEquipment(projectId)
            if (response.isSuccessful && response.body()?.data != null) {
                Result.success(response.body()!!.data!!)
            } else {
                Result.failure(Exception("Failed to load project equipment"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

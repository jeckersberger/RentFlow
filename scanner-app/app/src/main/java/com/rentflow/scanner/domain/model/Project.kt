package com.rentflow.scanner.domain.model

data class Project(
    val id: String,
    val name: String,
    val status: String,
    val color: String?,
    val client: String? = null,
    val start_date: String? = null,
    val end_date: String? = null,
    val equipment_count: Int = 0,
    val isDryHire: Boolean = false,
)

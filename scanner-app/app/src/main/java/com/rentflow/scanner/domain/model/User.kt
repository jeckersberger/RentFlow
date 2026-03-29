package com.rentflow.scanner.domain.model

data class User(
    val id: String,
    val email: String,
    val name: String? = null,
    val first_name: String? = null,
    val last_name: String? = null,
    val roles: List<String> = emptyList(),
    val tenantId: String? = null,
    val tenant_id: String? = null,
) {
    val displayName: String
        get() = name
            ?: listOfNotNull(first_name, last_name).joinToString(" ").ifBlank { email }
}

package com.rentflow.scanner.ui.createproject

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.rentflow.scanner.data.api.AddressDto
import com.rentflow.scanner.data.api.CreateCustomerRequest
import com.rentflow.scanner.data.api.CreateProjectRequest
import com.rentflow.scanner.data.api.CustomerDto
import com.rentflow.scanner.data.repository.ScannerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class CreateProjectUiState(
    // Project
    val name: String = "",
    val startDate: String = "",
    val endDate: String = "",
    val notes: String = "",
    val isDryHire: Boolean = false,
    // Venue
    val venueStreet: String = "",
    val venueCity: String = "",
    val venuePostalCode: String = "",
    // Customer search
    val customerQuery: String = "",
    val customerResults: List<CustomerDto> = emptyList(),
    val selectedCustomer: CustomerDto? = null,
    val isSearching: Boolean = false,
    // New customer
    val showNewCustomer: Boolean = false,
    val newCustomerName: String = "",
    val newCustomerEmail: String = "",
    val newCustomerPhone: String = "",
    val newCustomerStreet: String = "",
    val newCustomerCity: String = "",
    val newCustomerPostalCode: String = "",
    // State
    val isLoading: Boolean = false,
    val error: String? = null,
    val createdProjectName: String? = null,
)

@HiltViewModel
class CreateProjectViewModel @Inject constructor(
    private val scannerRepository: ScannerRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(CreateProjectUiState())
    val uiState: StateFlow<CreateProjectUiState> = _uiState

    private var searchJob: Job? = null

    fun updateName(v: String) { _uiState.update { it.copy(name = v) } }
    fun updateStartDate(v: String) { _uiState.update { it.copy(startDate = v) } }
    fun updateEndDate(v: String) { _uiState.update { it.copy(endDate = v) } }
    fun updateNotes(v: String) { _uiState.update { it.copy(notes = v) } }
    fun updateIsDryHire(v: Boolean) { _uiState.update { it.copy(isDryHire = v) } }
    fun updateVenueStreet(v: String) { _uiState.update { it.copy(venueStreet = v) } }
    fun updateVenueCity(v: String) { _uiState.update { it.copy(venueCity = v) } }
    fun updateVenuePostalCode(v: String) { _uiState.update { it.copy(venuePostalCode = v) } }

    // Customer search with debounce
    fun updateCustomerQuery(query: String) {
        _uiState.update { it.copy(customerQuery = query, selectedCustomer = null) }
        searchJob?.cancel()
        if (query.length < 2) {
            _uiState.update { it.copy(customerResults = emptyList()) }
            return
        }
        searchJob = viewModelScope.launch {
            delay(300)
            _uiState.update { it.copy(isSearching = true) }
            scannerRepository.searchCustomers(query).onSuccess { results ->
                _uiState.update { it.copy(customerResults = results, isSearching = false) }
            }.onFailure {
                _uiState.update { it.copy(isSearching = false) }
            }
        }
    }

    fun selectCustomer(customer: CustomerDto) {
        _uiState.update {
            it.copy(
                selectedCustomer = customer,
                customerQuery = customer.name,
                customerResults = emptyList(),
            )
        }
    }

    fun clearCustomer() {
        _uiState.update { it.copy(selectedCustomer = null, customerQuery = "") }
    }

    // New customer
    fun toggleNewCustomer() {
        _uiState.update { it.copy(showNewCustomer = !it.showNewCustomer, error = null) }
    }

    fun updateNewCustomerName(v: String) { _uiState.update { it.copy(newCustomerName = v) } }
    fun updateNewCustomerEmail(v: String) { _uiState.update { it.copy(newCustomerEmail = v) } }
    fun updateNewCustomerPhone(v: String) { _uiState.update { it.copy(newCustomerPhone = v) } }
    fun updateNewCustomerStreet(v: String) { _uiState.update { it.copy(newCustomerStreet = v) } }
    fun updateNewCustomerCity(v: String) { _uiState.update { it.copy(newCustomerCity = v) } }
    fun updateNewCustomerPostalCode(v: String) { _uiState.update { it.copy(newCustomerPostalCode = v) } }

    fun saveNewCustomer() {
        val state = _uiState.value
        if (state.newCustomerName.isBlank()) {
            _uiState.update { it.copy(error = "Kundenname ist erforderlich") }
            return
        }
        if (state.newCustomerStreet.isBlank() || state.newCustomerCity.isBlank()) {
            _uiState.update { it.copy(error = "Adresse (Straße + Stadt) ist erforderlich") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            scannerRepository.createCustomer(
                CreateCustomerRequest(
                    name = state.newCustomerName.trim(),
                    email = state.newCustomerEmail.trim().ifBlank { null },
                    phone = state.newCustomerPhone.trim().ifBlank { null },
                    address_street = state.newCustomerStreet.trim(),
                    address_city = state.newCustomerCity.trim(),
                    address_postcode = state.newCustomerPostalCode.trim().ifBlank { null },
                    address_country = "DE",
                )
            ).fold(
                onSuccess = { customer ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            selectedCustomer = customer,
                            customerQuery = customer.name,
                            showNewCustomer = false,
                            newCustomerName = "",
                            newCustomerEmail = "",
                            newCustomerPhone = "",
                            newCustomerStreet = "",
                            newCustomerCity = "",
                            newCustomerPostalCode = "",
                        )
                    }
                },
                onFailure = { e ->
                    _uiState.update { it.copy(isLoading = false, error = e.message) }
                },
            )
        }
    }

    fun submit() {
        val state = _uiState.value
        if (state.name.isBlank()) {
            _uiState.update { it.copy(error = "Projektname ist erforderlich") }
            return
        }

        val customer = state.selectedCustomer
        val venueAddr = if (state.venueStreet.isNotBlank() || state.venueCity.isNotBlank()) {
            AddressDto(
                street = state.venueStreet.trim().ifBlank { null },
                city = state.venueCity.trim().ifBlank { null },
                postal_code = state.venuePostalCode.trim().ifBlank { null },
                country = "DE",
            )
        } else null

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            scannerRepository.createProject(
                CreateProjectRequest(
                    name = state.name.trim(),
                    client_name = customer?.name,
                    client_email = customer?.email,
                    client_phone = customer?.phone,
                    client_address = customer?.let {
                        AddressDto(
                            street = it.address_street,
                            city = it.address_city,
                            postal_code = it.address_postcode,
                            country = it.address_country,
                        )
                    },
                    venue_address = venueAddr,
                    start_date = state.startDate.trim().ifBlank { null },
                    end_date = state.endDate.trim().ifBlank { null },
                    notes = state.notes.trim().ifBlank { null },
                    is_dry_hire = state.isDryHire,
                )
            ).fold(
                onSuccess = {
                    _uiState.update { it.copy(isLoading = false, createdProjectName = state.name.trim()) }
                },
                onFailure = { e ->
                    _uiState.update { it.copy(isLoading = false, error = e.message) }
                },
            )
        }
    }

    fun reset() {
        _uiState.value = CreateProjectUiState()
    }
}

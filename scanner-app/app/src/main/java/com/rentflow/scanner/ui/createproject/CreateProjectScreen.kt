package com.rentflow.scanner.ui.createproject

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Success

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CreateProjectScreen(
    onBack: () -> Unit,
    viewModel: CreateProjectViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.create_project_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
            )
        },
    ) { padding ->
        if (state.createdProjectName != null) {
            SuccessView(state.createdProjectName!!, onBack, viewModel::reset, Modifier.padding(padding))
        } else {
            ProjectForm(state, viewModel, Modifier.padding(padding))
        }
    }
}

@Composable
private fun SuccessView(name: String, onBack: () -> Unit, onReset: () -> Unit, modifier: Modifier) {
    Column(
        modifier = modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(Icons.Default.CheckCircle, null, Modifier.size(96.dp), tint = Success)
        Spacer(Modifier.height(16.dp))
        Text(stringResource(R.string.create_project_success), style = MaterialTheme.typography.headlineSmall, color = Success, textAlign = TextAlign.Center)
        Spacer(Modifier.height(8.dp))
        Text(name, style = MaterialTheme.typography.titleMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(Modifier.height(8.dp))
        Text(stringResource(R.string.create_project_success_hint), style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant, textAlign = TextAlign.Center)
        Spacer(Modifier.height(32.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            OutlinedButton(onClick = onBack) { Text(stringResource(R.string.create_project_back)) }
            Button(onClick = onReset, colors = ButtonDefaults.buttonColors(containerColor = Cyan)) {
                Icon(Icons.Default.Add, null, Modifier.padding(end = 8.dp))
                Text(stringResource(R.string.create_project_another))
            }
        }
    }
}

@Composable
private fun ProjectForm(state: CreateProjectUiState, vm: CreateProjectViewModel, modifier: Modifier) {
    Column(
        modifier = modifier.fillMaxSize().padding(horizontal = 16.dp).verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Spacer(Modifier.height(4.dp))

        // --- Project name ---
        SectionHeader(stringResource(R.string.create_project_section_project))
        OutlinedTextField(
            value = state.name, onValueChange = vm::updateName,
            label = { Text(stringResource(R.string.create_project_name) + " *") },
            singleLine = true, modifier = Modifier.fillMaxWidth(),
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
            isError = state.error != null && state.name.isBlank(),
        )

        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            OutlinedTextField(
                value = state.startDate, onValueChange = vm::updateStartDate,
                label = { Text(stringResource(R.string.create_project_start)) },
                placeholder = { Text("TT.MM.JJJJ") }, singleLine = true,
                modifier = Modifier.weight(1f),
            )
            OutlinedTextField(
                value = state.endDate, onValueChange = vm::updateEndDate,
                label = { Text(stringResource(R.string.create_project_end)) },
                placeholder = { Text("TT.MM.JJJJ") }, singleLine = true,
                modifier = Modifier.weight(1f),
            )
        }

        Row(verticalAlignment = Alignment.CenterVertically) {
            Checkbox(checked = state.isDryHire, onCheckedChange = vm::updateIsDryHire)
            Text(stringResource(R.string.create_project_dry_hire), style = MaterialTheme.typography.bodyLarge)
        }

        // --- Customer ---
        HorizontalDivider(Modifier.padding(vertical = 4.dp))
        SectionHeader(stringResource(R.string.create_project_section_customer))

        if (state.selectedCustomer != null) {
            Card(Modifier.fillMaxWidth()) {
                Row(Modifier.fillMaxWidth().padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f)) {
                        Text(state.selectedCustomer.name, style = MaterialTheme.typography.titleMedium)
                        state.selectedCustomer.email?.let {
                            Text(it, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                        state.selectedCustomer.address_city?.let {
                            Text(
                                listOfNotNull(state.selectedCustomer.address_street, "$it ${state.selectedCustomer.address_postcode ?: ""}".trim()).joinToString(", "),
                                style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    }
                    IconButton(onClick = vm::clearCustomer) {
                        Icon(Icons.Default.Close, null)
                    }
                }
            }
        } else {
            OutlinedTextField(
                value = state.customerQuery, onValueChange = vm::updateCustomerQuery,
                label = { Text(stringResource(R.string.create_project_customer_search)) },
                singleLine = true, modifier = Modifier.fillMaxWidth(),
                trailingIcon = {
                    if (state.isSearching) CircularProgressIndicator(Modifier.size(20.dp))
                    else Icon(Icons.Default.Search, null)
                },
            )

            // Search results dropdown
            if (state.customerResults.isNotEmpty()) {
                Card(Modifier.fillMaxWidth()) {
                    Column {
                        state.customerResults.forEach { customer ->
                            ListItem(
                                headlineContent = { Text(customer.name) },
                                supportingContent = {
                                    Text(listOfNotNull(customer.email, customer.address_city).joinToString(" · "),
                                        style = MaterialTheme.typography.bodySmall)
                                },
                                modifier = Modifier.clickable { vm.selectCustomer(customer) },
                            )
                            if (customer != state.customerResults.last()) HorizontalDivider()
                        }
                    }
                }
            }

            TextButton(onClick = vm::toggleNewCustomer) {
                Icon(Icons.Default.PersonAdd, null, Modifier.padding(end = 8.dp))
                Text(stringResource(R.string.create_project_new_customer))
            }
        }

        // New customer form
        if (state.showNewCustomer) {
            Card(Modifier.fillMaxWidth()) {
                Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(stringResource(R.string.create_project_new_customer), style = MaterialTheme.typography.titleSmall)
                    OutlinedTextField(
                        value = state.newCustomerName, onValueChange = vm::updateNewCustomerName,
                        label = { Text(stringResource(R.string.create_project_customer_name) + " *") },
                        singleLine = true, modifier = Modifier.fillMaxWidth(),
                    )
                    OutlinedTextField(
                        value = state.newCustomerStreet, onValueChange = vm::updateNewCustomerStreet,
                        label = { Text(stringResource(R.string.create_project_street) + " *") },
                        singleLine = true, modifier = Modifier.fillMaxWidth(),
                    )
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        OutlinedTextField(
                            value = state.newCustomerPostalCode, onValueChange = vm::updateNewCustomerPostalCode,
                            label = { Text(stringResource(R.string.create_project_postal_code)) },
                            singleLine = true, modifier = Modifier.width(100.dp),
                            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                        )
                        OutlinedTextField(
                            value = state.newCustomerCity, onValueChange = vm::updateNewCustomerCity,
                            label = { Text(stringResource(R.string.create_project_city) + " *") },
                            singleLine = true, modifier = Modifier.weight(1f),
                        )
                    }
                    OutlinedTextField(
                        value = state.newCustomerEmail, onValueChange = vm::updateNewCustomerEmail,
                        label = { Text(stringResource(R.string.create_project_email)) },
                        singleLine = true, modifier = Modifier.fillMaxWidth(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
                    )
                    OutlinedTextField(
                        value = state.newCustomerPhone, onValueChange = vm::updateNewCustomerPhone,
                        label = { Text(stringResource(R.string.create_project_phone)) },
                        singleLine = true, modifier = Modifier.fillMaxWidth(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                    )
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                        TextButton(onClick = vm::toggleNewCustomer) {
                            Text(stringResource(R.string.rfid_assign_cancel))
                        }
                        Button(onClick = vm::saveNewCustomer, enabled = !state.isLoading) {
                            Text(stringResource(R.string.create_project_save_customer))
                        }
                    }
                }
            }
        }

        // --- Venue ---
        HorizontalDivider(Modifier.padding(vertical = 4.dp))
        SectionHeader(stringResource(R.string.create_project_section_venue))
        OutlinedTextField(
            value = state.venueStreet, onValueChange = vm::updateVenueStreet,
            label = { Text(stringResource(R.string.create_project_street)) },
            singleLine = true, modifier = Modifier.fillMaxWidth(),
        )
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            OutlinedTextField(
                value = state.venuePostalCode, onValueChange = vm::updateVenuePostalCode,
                label = { Text(stringResource(R.string.create_project_postal_code)) },
                singleLine = true, modifier = Modifier.width(100.dp),
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
            )
            OutlinedTextField(
                value = state.venueCity, onValueChange = vm::updateVenueCity,
                label = { Text(stringResource(R.string.create_project_city)) },
                singleLine = true, modifier = Modifier.weight(1f),
            )
        }

        // --- Notes ---
        HorizontalDivider(Modifier.padding(vertical = 4.dp))
        OutlinedTextField(
            value = state.notes, onValueChange = vm::updateNotes,
            label = { Text(stringResource(R.string.create_project_notes)) },
            modifier = Modifier.fillMaxWidth().height(100.dp), maxLines = 4,
        )

        state.error?.let {
            Text(it, color = MaterialTheme.colorScheme.error)
        }

        Spacer(Modifier.height(4.dp))

        Button(
            onClick = vm::submit, modifier = Modifier.fillMaxWidth().height(56.dp),
            enabled = !state.isLoading,
            colors = ButtonDefaults.buttonColors(containerColor = Cyan),
        ) {
            if (state.isLoading) {
                CircularProgressIndicator(Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary)
            } else {
                Icon(Icons.Default.Save, null, Modifier.padding(end = 8.dp))
                Text(stringResource(R.string.create_project_save))
            }
        }

        Spacer(Modifier.height(16.dp))
    }
}

@Composable
private fun SectionHeader(title: String) {
    Text(title, style = MaterialTheme.typography.titleSmall, color = Cyan, modifier = Modifier.padding(top = 4.dp))
}

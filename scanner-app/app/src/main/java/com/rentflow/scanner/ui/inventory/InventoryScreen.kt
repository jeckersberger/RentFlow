package com.rentflow.scanner.ui.inventory

import androidx.compose.animation.animateContentSize
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.rentflow.scanner.R
import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.ui.theme.Cyan
import com.rentflow.scanner.ui.theme.Error
import com.rentflow.scanner.ui.theme.Success
import com.rentflow.scanner.ui.theme.Warning

private enum class InventoryFilter { ALL, FOUND, MISSING, UNEXPECTED }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun InventoryScreen(
    onBack: () -> Unit,
    onCompleted: () -> Unit,
    viewModel: InventoryViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsState()
    var selectedFilter by remember { mutableStateOf(InventoryFilter.ALL) }

    LaunchedEffect(state.completed) {
        if (state.completed) onCompleted()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.inventory_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = null)
                    }
                },
            )
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
            if (state.selectedJob == null && state.selectedZone == null) {
                // --- Job / Zone selection ---
                if (state.jobs.isNotEmpty()) {
                    Text(
                        stringResource(R.string.inventory_select_job),
                        style = MaterialTheme.typography.titleLarge,
                    )
                    Spacer(Modifier.height(16.dp))
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        items(state.jobs) { job ->
                            Card(
                                onClick = { viewModel.selectJob(job) },
                                modifier = Modifier.fillMaxWidth(),
                            ) {
                                Column(modifier = Modifier.padding(16.dp)) {
                                    Text(job.name, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
                                    job.zone?.let {
                                        Text(it, style = MaterialTheme.typography.bodyMedium, color = Cyan)
                                    }
                                    if (job.expectedCount > 0) {
                                        Text(
                                            stringResource(R.string.inventory_soll_items, job.expectedCount),
                                            style = MaterialTheme.typography.bodyMedium,
                                        )
                                    }
                                }
                            }
                        }
                    }
                } else {
                    // Fallback: zone selection
                    Text(
                        stringResource(R.string.inventory_select_zone),
                        style = MaterialTheme.typography.titleLarge,
                    )
                    Spacer(Modifier.height(16.dp))
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        items(state.zones) { zone ->
                            Card(
                                onClick = { viewModel.selectZone(zone) },
                                modifier = Modifier.fillMaxWidth(),
                            ) {
                                Column(modifier = Modifier.padding(16.dp)) {
                                    Text(zone.name, style = MaterialTheme.typography.bodyLarge)
                                    Text(
                                        stringResource(R.string.inventory_soll_items, zone.expectedItemCount),
                                        style = MaterialTheme.typography.bodyMedium,
                                    )
                                }
                            }
                        }
                    }
                }
            } else {
                // --- Inventory scanning view ---
                Text(
                    "${stringResource(R.string.inventory_title)}: ${state.selectedZone!!.name}",
                    style = MaterialTheme.typography.titleLarge,
                )
                Spacer(Modifier.height(12.dp))

                // Summary: Gefunden X / Soll Y
                val foundCount = state.foundItems.size
                val expectedCount = state.expectedItems.size
                val progress = if (expectedCount > 0) foundCount.toFloat() / expectedCount else 0f

                Column {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            stringResource(R.string.inventory_found_of, foundCount, expectedCount),
                            style = MaterialTheme.typography.titleMedium,
                            color = if (foundCount == expectedCount && expectedCount > 0) Success else Cyan,
                        )
                        if (state.unexpectedItems.isNotEmpty()) {
                            Text(
                                stringResource(R.string.inventory_unexpected_count, state.unexpectedItems.size),
                                style = MaterialTheme.typography.bodySmall,
                                color = Warning,
                            )
                        }
                    }
                    Spacer(Modifier.height(6.dp))
                    LinearProgressIndicator(
                        progress = { progress },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(8.dp)
                            .clip(RoundedCornerShape(4.dp)),
                        color = if (foundCount == expectedCount && expectedCount > 0) Success else Cyan,
                        trackColor = MaterialTheme.colorScheme.surfaceVariant,
                    )
                }

                Spacer(Modifier.height(12.dp))

                // Filter chips
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    FilterChip(
                        selected = selectedFilter == InventoryFilter.ALL,
                        onClick = { selectedFilter = InventoryFilter.ALL },
                        label = {
                            Text(stringResource(R.string.inventory_all, state.scannedItems.size + state.missingItems.size))
                        },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Cyan.copy(alpha = 0.2f),
                            selectedLabelColor = Cyan,
                        ),
                    )
                    FilterChip(
                        selected = selectedFilter == InventoryFilter.FOUND,
                        onClick = { selectedFilter = InventoryFilter.FOUND },
                        label = { Text(stringResource(R.string.inventory_found_count, state.foundItems.size)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Success.copy(alpha = 0.2f),
                            selectedLabelColor = Success,
                        ),
                    )
                    FilterChip(
                        selected = selectedFilter == InventoryFilter.MISSING,
                        onClick = { selectedFilter = InventoryFilter.MISSING },
                        label = { Text(stringResource(R.string.inventory_missing_count, state.missingItems.size)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Error.copy(alpha = 0.2f),
                            selectedLabelColor = Error,
                        ),
                    )
                    FilterChip(
                        selected = selectedFilter == InventoryFilter.UNEXPECTED,
                        onClick = { selectedFilter = InventoryFilter.UNEXPECTED },
                        label = { Text(stringResource(R.string.inventory_unexpected_count_filter, state.unexpectedItems.size)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Warning.copy(alpha = 0.2f),
                            selectedLabelColor = Warning,
                        ),
                    )
                }

                Spacer(Modifier.height(8.dp))

                state.error?.let {
                    Text(it, color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(8.dp))
                }

                // Item list
                val displayItems: List<Pair<Equipment, Color>> = when (selectedFilter) {
                    InventoryFilter.ALL -> {
                        val found = state.foundItems.map { it to Success }
                        val missing = state.missingItems.map { it to Error }
                        val unexpected = state.unexpectedItems.map { it to Warning }
                        found + missing + unexpected
                    }
                    InventoryFilter.FOUND -> state.foundItems.map { it to Success }
                    InventoryFilter.MISSING -> state.missingItems.map { it to Error }
                    InventoryFilter.UNEXPECTED -> state.unexpectedItems.map { it to Warning }
                }

                LazyColumn(
                    modifier = Modifier.weight(1f).animateContentSize(),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    items(displayItems, key = { it.first.id }) { (equipment, statusColor) ->
                        InventoryItemCard(equipment = equipment, statusColor = statusColor)
                    }
                    if (displayItems.isEmpty()) {
                        item {
                            Box(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 32.dp),
                                contentAlignment = Alignment.Center,
                            ) {
                                Text(
                                    when (selectedFilter) {
                                        InventoryFilter.ALL -> stringResource(R.string.inventory_empty_all)
                                        InventoryFilter.FOUND -> stringResource(R.string.inventory_empty_found)
                                        InventoryFilter.MISSING -> stringResource(R.string.inventory_empty_missing)
                                        InventoryFilter.UNEXPECTED -> stringResource(R.string.inventory_empty_unexpected)
                                    },
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                )
                            }
                        }
                    }
                }

                Spacer(Modifier.height(16.dp))
                Button(
                    onClick = viewModel::completeInventory,
                    enabled = !state.isLoading,
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                ) {
                    Text(stringResource(R.string.inventory_confirm))
                }
            }
        }
    }
}

@Composable
private fun InventoryItemCard(
    equipment: Equipment,
    statusColor: Color,
    modifier: Modifier = Modifier,
) {
    Card(modifier = modifier.fillMaxWidth()) {
        Row(modifier = Modifier.fillMaxWidth()) {
            // Color-coded left border
            Box(
                modifier = Modifier
                    .width(6.dp)
                    .fillMaxHeight()
                    .defaultMinSize(minHeight = 72.dp)
                    .background(statusColor),
            )
            Column(modifier = Modifier.padding(12.dp).weight(1f)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        equipment.name,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.weight(1f),
                    )
                    Text(
                        when (statusColor) {
                            Success -> stringResource(R.string.inventory_status_found)
                            Error -> stringResource(R.string.inventory_status_missing)
                            Warning -> stringResource(R.string.inventory_status_unexpected)
                            else -> ""
                        },
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.Bold,
                        color = statusColor,
                        modifier = Modifier
                            .clip(RoundedCornerShape(4.dp))
                            .background(statusColor.copy(alpha = 0.15f))
                            .padding(horizontal = 6.dp, vertical = 2.dp),
                    )
                }
                Spacer(Modifier.height(4.dp))
                Text(
                    stringResource(R.string.equipment_barcode, equipment.barcode),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                equipment.category.takeIf { it.isNotBlank() }?.let {
                    Text(
                        stringResource(R.string.equipment_category, it),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            }
        }
    }
}

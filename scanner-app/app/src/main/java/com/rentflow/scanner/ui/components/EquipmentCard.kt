package com.rentflow.scanner.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.rentflow.scanner.R
import com.rentflow.scanner.domain.model.Equipment
import com.rentflow.scanner.ui.theme.Warning

@Composable
fun EquipmentCard(
    equipment: Equipment,
    onClick: (() -> Unit)? = null,
    adHoc: Boolean = false,
    onRemove: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
) {
    Card(
        onClick = onClick ?: {},
        enabled = onClick != null,
        modifier = modifier.fillMaxWidth(),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(equipment.name, style = MaterialTheme.typography.titleLarge, modifier = Modifier.weight(1f))
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    if (adHoc) {
                        Text(
                            stringResource(R.string.spontan),
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = FontWeight.Bold,
                            color = Warning,
                            modifier = Modifier
                                .clip(RoundedCornerShape(4.dp))
                                .background(Warning.copy(alpha = 0.15f))
                                .padding(horizontal = 6.dp, vertical = 2.dp),
                        )
                    }
                    StatusBadge(equipment.status)
                    if (onRemove != null) {
                        IconButton(onClick = onRemove, modifier = Modifier.size(32.dp)) {
                            Icon(
                                Icons.Default.Close,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.error,
                                modifier = Modifier.size(18.dp),
                            )
                        }
                    }
                }
            }
            Spacer(Modifier.height(8.dp))
            Text(stringResource(R.string.equipment_barcode, equipment.barcode), style = MaterialTheme.typography.bodyMedium)
            Text(stringResource(R.string.equipment_category, equipment.category), style = MaterialTheme.typography.bodyMedium)
            equipment.location?.let {
                Text(stringResource(R.string.equipment_location, it), style = MaterialTheme.typography.bodyMedium)
            }
            equipment.projectName?.let {
                Text(stringResource(R.string.equipment_project, it), style = MaterialTheme.typography.bodyMedium)
            }
        }
    }
}

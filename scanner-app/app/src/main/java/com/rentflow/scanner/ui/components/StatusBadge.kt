package com.rentflow.scanner.ui.components

import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import com.rentflow.scanner.R
import com.rentflow.scanner.domain.model.EquipmentStatus
import com.rentflow.scanner.ui.theme.*

@Composable
fun StatusBadge(status: EquipmentStatus) {
    val (color, label) = when (status) {
        EquipmentStatus.AVAILABLE -> Success to stringResource(R.string.status_available)
        EquipmentStatus.CHECKED_OUT -> Warning to stringResource(R.string.status_checked_out)
        EquipmentStatus.IN_MAINTENANCE -> Cyan to stringResource(R.string.status_maintenance)
        EquipmentStatus.DAMAGED -> Error to stringResource(R.string.status_damaged)
        EquipmentStatus.RETIRED -> TextSecondary to stringResource(R.string.status_retired)
    }
    Badge(containerColor = color) { Text(label, color = Color.White) }
}

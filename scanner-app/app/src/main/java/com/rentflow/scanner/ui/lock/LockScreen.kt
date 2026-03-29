package com.rentflow.scanner.ui.lock

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.QrCodeScanner
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import com.rentflow.scanner.R
import com.rentflow.scanner.ui.theme.Cyan

@Composable
fun LockScreen(
    onUnlock: () -> Unit,
    onFullLogin: () -> Unit,
    onQrLogin: () -> Unit,
    onPasswordSubmit: (String) -> Unit,
    error: String? = null,
    isLoading: Boolean = false,
) {
    var password by remember { mutableStateOf("") }
    var passwordVisible by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Icon(
            Icons.Default.Lock,
            contentDescription = null,
            modifier = Modifier.size(64.dp),
            tint = Cyan,
        )
        Spacer(Modifier.height(16.dp))
        Text(
            stringResource(R.string.lock_title),
            style = MaterialTheme.typography.headlineMedium,
        )
        Text(
            stringResource(R.string.lock_subtitle),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(32.dp))

        OutlinedTextField(
            value = password,
            onValueChange = { password = it },
            label = { Text(stringResource(R.string.login_password)) },
            visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
            trailingIcon = {
                IconButton(onClick = { passwordVisible = !passwordVisible }) {
                    Icon(
                        if (passwordVisible) Icons.Default.VisibilityOff else Icons.Default.Visibility,
                        contentDescription = null,
                    )
                }
            },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
        )

        error?.let {
            Spacer(Modifier.height(8.dp))
            Text(it, color = MaterialTheme.colorScheme.error)
        }

        Spacer(Modifier.height(16.dp))

        Button(
            onClick = { onPasswordSubmit(password) },
            enabled = password.isNotBlank() && !isLoading,
            modifier = Modifier.fillMaxWidth().height(56.dp),
        ) {
            if (isLoading) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary)
            } else {
                Text(stringResource(R.string.lock_unlock))
            }
        }

        Spacer(Modifier.height(16.dp))

        OutlinedButton(
            onClick = onQrLogin,
            modifier = Modifier.fillMaxWidth().height(56.dp),
        ) {
            Icon(Icons.Default.QrCodeScanner, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
            Text(stringResource(R.string.login_qr))
        }

        Spacer(Modifier.height(12.dp))

        TextButton(onClick = onFullLogin) {
            Text(stringResource(R.string.lock_switch_user))
        }
    }
}

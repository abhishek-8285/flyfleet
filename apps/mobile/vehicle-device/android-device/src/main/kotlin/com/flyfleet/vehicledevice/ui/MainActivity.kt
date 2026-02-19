package com.flyfleet.vehicledevice.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.flyfleet.vehicledevice.ui.boot.SecureBootScreen
import com.flyfleet.vehicledevice.ui.theme.FlyFleetTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            FlyFleetTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    SecureBootScreen()
                }
            }
        }
    }
}

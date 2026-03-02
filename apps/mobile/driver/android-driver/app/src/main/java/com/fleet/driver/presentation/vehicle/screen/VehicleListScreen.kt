package com.fleet.driver.presentation.vehicle.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.filled.DirectionsCar
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.compose.*
import com.fleet.driver.data.remote.model.Vehicle
import com.fleet.driver.presentation.vehicle.viewmodel.VehicleViewModel

/**
 * Vehicle List Screen
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VehicleListScreen(
    viewModel: VehicleViewModel = hiltViewModel(),
    onVehicleSelected: (String) -> Unit
) {
    val vehicles by viewModel.vehicles.collectAsState()
    val myVehicles by viewModel.myVehicles.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val errorMessage by viewModel.errorMessage.collectAsState()
    val searchQuery by viewModel.searchQuery.collectAsState()
    
    var showFilters by remember { mutableStateOf(false) }
    var selectedStatusFilter by remember { mutableStateOf<String?>(null) }
    var selectedTypeFilter by remember { mutableStateOf<String?>(null) }
    var showAssignedOnly by remember { mutableStateOf(false) }
    
    // Auto-refresh periodically
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(30000) // Refresh every 30 seconds
        viewModel.refresh()
    }
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Vehicles") },
                actions = {
                    IconButton(onClick = { showFilters = !showFilters }) {
                        Icon(
                            imageVector = Icons.Default.FilterList,
                            contentDescription = "Filter"
                        )
                    }
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(
                            imageVector = Icons.Default.Refresh,
                            contentDescription = "Refresh"
                        )
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = {
                    val suggestedVehicle = vehicles.firstOrNull { !viewModel.isVehicleAssignedToMe(it.id) }
                    if (suggestedVehicle != null) {
                        onVehicleSelected(suggestedVehicle.id)
                    } else {
                        viewModel.refresh()
                    }
                },
                containerColor = MaterialTheme.colorScheme.primary
            ) {
                Icon(
                    imageVector = Icons.Default.Add,
                    contentDescription = "Add Vehicle"
                )
            }
        }
    ) { paddingValues ->
        Box(modifier = Modifier.fillMaxSize().padding(paddingValues)) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
            ) {
                // Search Bar
                OutlinedTextField(
                    value = searchQuery,
                    onValueChange = { viewModel.searchVehicles(it) },
                    label = { Text("Search vehicles...") },
                    leadingIcon = {
                        Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = "Search"
                        )
                    },
                    trailingIcon = {
                        if (searchQuery.isNotEmpty()) {
                            IconButton(onClick = { viewModel.clearSearch() }) {
                                Icon(
                                    imageVector = Icons.Default.Clear,
                                    contentDescription = "Clear"
                                )
                            }
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    singleLine = true
                )
                
                // My Vehicles Section
                if (myVehicles.isNotEmpty()) {
                    Card(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 8.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.primaryContainer
                        )
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "My Assigned Vehicles",
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.Bold
                                )
                                Text(
                                    text = "${myVehicles.size}",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onPrimaryContainer
                                )
                            }
                            
                            Spacer(modifier = Modifier.height(8.dp))
                            
                            LazyColumn(
                                modifier = Modifier.height(120.dp)
                            ) {
                                items(myVehicles.take(3)) { vehicle ->
                                    MyVehicleItem(
                                        vehicle = vehicle,
                                        onClick = { onVehicleSelected(vehicle.id) }
                                    )
                                    if (vehicle != myVehicles.take(3).last()) {
                                        Divider()
                                    }
                                }
                            }
                            
                            if (myVehicles.size > 3) {
                                TextButton(
                                    onClick = { showAssignedOnly = true },
                                    modifier = Modifier.align(Alignment.End)
                                ) {
                                    Text("View All (${myVehicles.size})")
                                }
                            }
                        }
                    }
                }
                
                // Filters Sheet
                if (showFilters) {
                    FilterSheet(
                        currentStatus = selectedStatusFilter,
                        currentType = selectedTypeFilter,
                        onStatusSelected = { status ->
                            selectedStatusFilter = status
                            viewModel.filterByStatus(status)
                        },
                        onTypeSelected = { type ->
                            selectedTypeFilter = type
                            viewModel.filterByType(type)
                        },
                        onDismiss = { showFilters = false }
                    )
                }
                
                // Vehicle List
                val displayedVehicles = if (showAssignedOnly) {
                    vehicles.filter { v -> viewModel.isVehicleAssignedToMe(v.id) }
                } else {
                    vehicles
                }

                if (showAssignedOnly) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "Showing assigned vehicles only",
                            style = MaterialTheme.typography.bodySmall
                        )
                        TextButton(onClick = { showAssignedOnly = false }) {
                            Text("Show all fleet")
                        }
                    }
                }

                if (isLoading) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        CircularProgressIndicator()
                    }
                } else if (displayedVehicles.isEmpty()) {
                    EmptyState(
                        onRetry = { viewModel.refresh() }
                    )
                } else {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        items(displayedVehicles) { vehicle ->
                            VehicleListItem(
                                vehicle = vehicle,
                                isAssignedToMe = viewModel.isVehicleAssignedToMe(vehicle.id),
                                onClick = { onVehicleSelected(vehicle.id) }
                            )
                        }
                    }
                }
            }
            
            // Error Snackbar
            errorMessage?.let { message ->
                Snackbar(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(16.dp)
                        .clip(RoundedCornerShape(8.dp)),
                    containerColor = MaterialTheme.colorScheme.errorContainer,
                    contentColor = MaterialTheme.colorScheme.onErrorContainer
                ) {
                    Text(text = message)
                }
                
                LaunchedEffect(errorMessage) {
                    kotlinx.coroutines.delay(5000)
                    viewModel.clearError()
                }
            }
        }
    }
}

@Composable
private fun VehicleListItem(
    vehicle: Vehicle,
    isAssignedToMe: Boolean,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = vehicle.registrationNumber,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold
                )
                
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = when (vehicle.status) {
                            "available" -> Icons.Default.CheckCircle
                            "in_use" -> Icons.Default.DirectionsCar
                            "maintenance" -> Icons.Default.Build
                            "retired" -> Icons.Default.Block
                            else -> Icons.Default.Help
                        },
                        contentDescription = vehicle.status,
                        tint = when (vehicle.status) {
                            "available" -> Color.Green
                            "in_use" -> Color.Blue
                            "maintenance" -> Color.Yellow
                            "retired" -> Color.Red
                            else -> Color.Gray
                        },
                        modifier = Modifier.size(16.dp)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = vehicle.status.replace("_", " ").uppercase(),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "${vehicle.make} ${vehicle.model} (${vehicle.year})",
                    style = MaterialTheme.typography.bodyMedium
                )
                Text(
                    text = vehicle.type.replaceFirstChar { it.uppercase() },
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.primary
                )
            }
            
            if (vehicle.currentLocation != null) {
                Spacer(modifier = Modifier.height(8.dp))
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = Icons.Default.LocationOn,
                        contentDescription = "Location",
                        modifier = Modifier.size(16.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = vehicle.currentLocation.address ?: "Location available",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            
            if (isAssignedToMe) {
                Spacer(modifier = Modifier.height(8.dp))
                Card(
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.secondaryContainer
                    )
                ) {
                    Text(
                        text = "ASSIGNED TO YOU",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSecondaryContainer,
                        modifier = Modifier.padding(4.dp)
                    )
                }
            }
        }
    }
}

@Composable
private fun MyVehicleItem(
    vehicle: Vehicle,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = Icons.Default.DirectionsCar,
            contentDescription = "Vehicle",
            tint = MaterialTheme.colorScheme.primary
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(
            modifier = Modifier.weight(1f)
        ) {
            Text(
                text = vehicle.registrationNumber,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )
            Text(
                text = "${vehicle.make} ${vehicle.model}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun FilterSheet(
    currentStatus: String?,
    currentType: String?,
    onStatusSelected: (String?) -> Unit,
    onTypeSelected: (String?) -> Unit,
    onDismiss: () -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Filter Vehicles",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.padding(bottom = 16.dp)
            )
            
            // Status Filter
            Text(
                text = "Status",
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            val statuses = listOf("available", "in_use", "maintenance", "retired")
            statuses.forEach { status ->
                FilterChip(
                    selected = currentStatus == status,
                    onClick = { onStatusSelected(if (currentStatus == status) null else status) },
                    label = { Text(status.replace("_", " ").uppercase()) },
                    modifier = Modifier.padding(end = 8.dp)
                )
            }
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Type Filter
            Text(
                text = "Type",
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            val types = listOf("truck", "van", "car", "motorcycle")
            types.forEach { type ->
                FilterChip(
                    selected = currentType == type,
                    onClick = { onTypeSelected(if (currentType == type) null else type) },
                    label = { Text(type.replaceFirstChar { it.uppercase() }) },
                    modifier = Modifier.padding(end = 8.dp)
                )
            }
        }
    }
}

@Composable
private fun EmptyState(
    onRetry: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = Icons.Default.DirectionsCar,
            contentDescription = "No vehicles",
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "No vehicles found",
            style = MaterialTheme.typography.headlineSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "Try adjusting your search or filters",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Button(onClick = onRetry) {
            Text("Retry")
        }
    }
}

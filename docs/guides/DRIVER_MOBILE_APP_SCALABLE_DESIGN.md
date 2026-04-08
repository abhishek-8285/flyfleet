# Driver Mobile App: Scalable Architecture Design

**Razorpay-Inspired Design Principles:** Clean, Modular, Performant, Scalable
**Platform:** Android (Kotlin + Jetpack Compose + Material 3)
**Architecture:** Clean Architecture + MVVM + Modular Design
**Target Scale:** 10,000+ concurrent drivers, 1M+ daily active users

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Architecture Pillars](#architecture-pillars)
- [Module Structure](#module-structure)
- [Layer Architecture](#layer-architecture)
- [Scalability Patterns](#scalability-patterns)
- [Performance Optimizations](#performance-optimizations)
- [Security Design](#security-design)
- [Offline Strategy](#offline-strategy)
- [Monitoring & Observability](#monitoring--observability)
- [Technology Stack](#technology-stack)
- [Implementation Roadmap](#implementation-roadmap)

---

## Executive Summary

The Driver Mobile App is a critical touchpoint for fleet operations, enabling drivers to:
- Execute assigned trips with navigation and status tracking
- Log fuel expenses and upload documents
- Receive real-time alerts and SOS signals
- Manage their profile and compliance documents

**Design Goals:**
- **Scalable:** Support 10k+ concurrent drivers per app instance
- **Resilient:** Full offline capability with sync on reconnect
- **Performant:** <2s app startup, <100ms UI interactions
- **Maintainable:** Modular architecture, 1000-line/file limit, testable
- **Secure:** End-to-end encryption, biometric auth, data protection

---

## Architecture Pillars

### 1. Modular Monolith (App-Level)
```
driver-app/
├── :app (APK) - Shell module, minimal code
├── :core:auth (Authentication, token mgmt)
├── :core:network (HTTP client, WebSocket, interceptors)
├── :core:ui (Compose components, themes, design system)
├── :core:data (Repository layer, data sources)
├── :core:database (Room DB, offline store)
├── :core:location (GPS, geofencing, tracking)
├── :core:notifications (Push, in-app alerts)
├── :feature:dashboard (Driver home screen)
├── :feature:trips (Trip management)
├── :feature:tracking (Real-time GPS, navigation)
├── :feature:fuel (Fuel logging, expenses)
├── :feature:documents (Upload, view, manage)
├── :feature:alerts (SOS, notifications)
├── :feature:profile (User settings, compliance)
├── :feature:command-center (Emergency tools)
└── :feature:sync (Background sync, conflict resolution)
```

Each feature is a standalone Gradle module with:
- **Presentation layer** (Compose UI + ViewModel)
- **Domain layer** (Use cases, business logic)
- **Data layer** (Repository + data sources)
- **Dependency direction:** Feature → Core → External libraries

---

### 2. Clean Architecture (Per Module)

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│  (Compose UI + ViewModel + State)      │
└───────────────┬─────────────────────────┘
                │ Depends on
┌───────────────▼─────────────────────────┐
│             Domain Layer                │
│  (Use Cases + Entities + Repositories) │
└───────────────┬─────────────────────────┘
                │ Depends on
┌───────────────▼─────────────────────────┐
│              Data Layer                 │
│  (RepositoryImpl + Data Sources)       │
└───────────────┬─────────────────────────┘
                │ Depends on
┌───────────────▼─────────────────────────┐
│         External Layers                 │
│  (Retrofit, Room, GPS, Notifications)  │
└─────────────────────────────────────────┘
```

**Dependency Rule:** Inner layers define interfaces, outer layers implement. No inner → outer dependencies.

---

### 3. Reactive State Management

**State Holder Pattern:**
```kotlin
sealed interface DashboardState {
    data object Loading : DashboardState
    data class Success(
        val activeTrip: Trip?,
        val vehicle: Vehicle?,
        val upcomingTrips: List<Trip>,
        val alerts: List<Alert>
    ) : DashboardState
    data class Error(val message: String) : DashboardState
}
```

**Unidirectional Data Flow (UDF):**
```
UI Event → ViewModel → UseCase → Repository → Data Source
    ↑                                              │
    └──────────────────── State ←─────────────────┘
```

**State Holder Libraries:**
- **ViewModel + StateFlow** for most screens
- **MVI (Model-View-Intent)** for complex flows (trip execution)
- **SharedFlow** for one-time events (navigation, dialogs)

---

## Module Structure

### Core Modules (Shared)

#### `:core:auth`
**Purpose:** Authentication & authorization
**Components:**
- `TokenManager` - JWT storage (EncryptedSharedPreferences)
- `Authenticator` - OAuth2 flows, refresh token logic
- `SessionManager` - Session lifecycle, timeout
- `BiometricAuth` - Fingerprint/face auth integration

**Key Features:**
- Silent token refresh using Refresh Token grant
- Token revocation on logout
- Biometric fallback to PIN
- MFA challenge handling

---

#### `:core:network`
**Purpose:** Networking infrastructure
**Components:**
- `HttpClient` - Retrofit with OkHttp interceptor chain
- `WebSocketClient` - Real-time telemetry, alerts
- `RequestInterceptor` - Auth headers, org_id, retry logic
- `ResponseInterceptor` - Error handling, logging
- `MockWebServer` - Test doubles for CI

**Configuration:**
```kotlin
val httpClient = OkHttpClient.Builder()
    .connectTimeout(30, TimeUnit.SECONDS)
    .readTimeout(30, TimeUnit.SECONDS)
    .writeTimeout(30, TimeUnit.SECONDS)
    .addInterceptor(AuthInterceptor())
    .addInterceptor(LoggingInterceptor())
    .addInterceptor(RetryInterceptor(maxRetries = 3))
    .addInterceptor(SANsInterceptor()) // Security: pinning
    .connectionPool(ConnectionPool(5, 1, TimeUnit.MINUTES))
    .build()
```

**Scalability:**
- Connection pooling (max 5 connections)
- HTTP/2 enabled for multiplexing
- Request deduplication for concurrent same calls
- Circuit breaker for resilience

---

#### `:core:ui`
**Purpose:** Design system & shared UI components
**Components:**
- `Theme` - Material 3 dynamic theming (light/dark)
- `Typography` - Font scales for accessibility
- `Components` - Reusable Composables:
  - `FleetButton`, `FleetTextField`, `FleetCard`
  - `LoadingSpinner`, `ErrorView`, `EmptyState`
  - `StatusChip`, `AlertBanner`, `MapMarker`
- `Navigation` - Compose Navigation + Type-safe routes
- `Animations` - Shared transitions, Lottie files

**Design Tokens:** (from `packages/config/branding/default.json`)
```kotlin
val Purple80 = Color(0xFFD0BCFF)
val PurpleGrey80 = Color(0xFFCCC2DC)
val Teal80 = Color(0xFF97C4EB)
// Extend with brand colors
```

---

#### `:core:data`
**Purpose:** Repository pattern implementation
**Components:**
- `BaseRepository` - Common error handling, caching strategy
- `CachePolicy` - NetworkFirst, CacheFirst, CacheOnly
- `PagingSource` - Pagination for large lists
- `DataMappers` - Entity ↔ DTO conversions

**Pattern:**
```kotlin
interface TripRepository {
    suspend fun getTrips(params: TripQueryParams): Result<List<Trip>>
    suspend fun getTrip(id: String): Result<Trip>
    suspend fun updateTripStatus(id: String, status: TripStatus): Result<Unit>
}

class TripRepositoryImpl(
    private val api: TripApi,
    private val local: TripLocalDataSource,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : TripRepository {
    override suspend fun getTrips(params: TripQueryParams) = withContext(dispatcher) {
        // Network + cache logic with flow control
        flow {
            emit(Result.loading())
            val cached = local.getTrips()
            emit(Result.success(cached)) // Immediate UI update

            try {
                val remote = api.getTrips(params)
                local.saveTrips(remote)
                emit(Result.success(remote))
            } catch (e: Exception) {
                // Cache fallback already emitted
                emit(Result.error(e))
            }
        }.flowOn(dispatcher).first()
    }
}
```

---

#### `:core:database`
**Purpose:** Offline-first data persistence
**Technology:** Room with SQLCipher (encryption)
**Schema:**
- 25+ entities (User, Vehicle, Trip, FuelLog, Alert, Document, etc.)
- Relationships: Foreign keys, indexes
- Versioned migrations (auto-migration for simple changes)

**Key Features:**
- **Encrypted database** with user-derived key (from login)
- **Full-text search** (FTS5) on trip notes, driver names
- **ChangeDataCapture** - Triggers for sync queue
- **Relationship queries** with @Relation
- **Paging 3** integration for large datasets

**Example DAO:**
```kotlin
@Dao
interface TripDao {
    @Query("SELECT * FROM trips WHERE driver_id = :driverId AND status IN (:statuses) ORDER BY scheduled_start DESC")
    fun getTrips(driverId: String, statuses: List<TripStatus>): PagingSource<Int, TripEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertTrips(trips: List<TripEntity>)

    @Query("SELECT * FROM trips WHERE needs_sync = 1")
    suspend fun getUnsyncedTrips(): List<TripEntity>

    @Query("UPDATE trips SET needs_sync = 0 WHERE id = :id")
    suspend fun markSynced(id: String)
}
```

---

### Feature Modules

#### `:feature:dashboard`
**Screens:**
- **DashboardScreen** - KPI cards, active trip, quick actions
- **ActivityFeed** - Recent events, alerts, notifications
- **QuickStats** - Weekly hours, miles, fuel efficiency

**State Management:**
```kotlin
class DashboardViewModel(
    private val getActiveTripUseCase: GetActiveTripUseCase,
    private val getUpcomingTripsUseCase: GetUpcomingTripsUseCase
) : ViewModel() {
    private val _state = MutableStateFlow<DashboardState>(DashboardState.Loading)
    val state: StateFlow<DashboardState> = _state.asStateFlow()

    init {
        loadDashboard()
    }

    private fun loadDashboard() {
        viewModelScope.launch {
            combine(
                getActiveTripUseCase(),
                getUpcomingTripsUseCase(limit = 5),
                getAlertsUseCase(unreadOnly = true)
            ) { active, upcoming, alerts ->
                DashboardState.Success(
                    activeTrip = active,
                    upcomingTrips = upcoming,
                    alerts = alerts
                )
            }.catch { e ->
                _state.value = DashboardState.Error(e.message ?: "Unknown error")
            }.collect { _state.value = it }
        }
    }
}
```

---

#### `:feature:trips`
**Screens:**
- **TripListScreen** - Upcoming/past trips with filters
- **TripDetailScreen** - Full trip info, stops, documents
- **TripExecutionScreen** - Active trip with navigation, status controls
- **TripChecklistScreen** - Pre-trip, post-trip inspection

**Complex Flow - Trip Execution:**
```kotlin
sealed interface TripExecutionState {
    data object Idle : TripExecutionState
    data class Tracking(
        val trip: Trip,
        val currentLocation: Location?,
        val elapsedTime: Duration,
        val distance: Double
    ) : TripExecutionState
    data class Paused(val reason: PauseReason) : TripExecutionState
    data class Completed(val trip: Trip) : TripExecutionState
}

class TripExecutionViewModel @Inject constructor(
    private val startTripUseCase: StartTripUseCase,
    private val updateTripStatusUseCase: UpdateTripStatusUseCase,
    private val locationTracker: LocationTracker
) : ViewModel() {
    private val _state = MutableStateFlow<TripExecutionState>(TripExecutionState.Idle)
    val state: StateFlow<TripExecutionState> = _state.asStateFlow()

    // Expose events as SharedFlow for one-time handling
    private val _events = MutableSharedFlow<TripEvent>()
    val events: SharedFlow<TripEvent> = _events.asSharedFlow()

    fun startTrip(tripId: String) {
        viewModelScope.launch {
            _state.value = TripExecutionState.Tracking(
                trip = startTripUseCase(tripId),
                currentLocation = null,
                elapsedTime = Duration.ZERO,
                distance = 0.0
            )
            startLocationTracking()
        }
    }

    private fun startLocationTracking() {
        locationTracker.start { location ->
            // Throttled updates (every 30s or 50m)
            val currentState = _state.value
            if (currentState is TripExecutionState.Tracking) {
                _state.value = currentState.copy(
                    currentLocation = location,
                    distance = currentState.distance + calculateDistance(currentState.currentLocation, location)
                )
            }
        }
    }

    fun pauseTrip(reason: PauseReason) {
        _state.value = TripExecutionState.Paused(reason)
        locationTracker.pause()
    }

    fun resumeTrip() {
        val current = _state.value
        if (current is TripExecutionState.Paused) {
            _state.value = current.copy() // Cast to Tracking
            locationTracker.resume()
        }
    }

    fun completeTrip() {
        viewModelScope.launch {
            val current = _state.value
            if (current is TripExecutionState.Tracking) {
                updateTripStatusUseCase(current.trip.id, TripStatus.COMPLETED)
                _state.value = current.copy(trip = current.trip.copy(status = TripStatus.COMPLETED))
                _events.emit(TripEvent.ShowCompletionDialog)
            }
        }
    }
}
```

---

#### `:feature:tracking`
**Real-time GPS & Navigation**

**Components:**
- `LocationTracker` - FusedLocationProviderClient wrapper
- `GeofenceManager` - Google Geofencing API integration
- `NavigationService` - Google Maps navigation intent
- `TelemetryUploader` - Batched telemetry upload (NATS/HTTP)

**Location Tracking Strategy:**
```kotlin
class LocationTracker @Inject constructor(
    @ApplicationContext context: Context,
    private val preferences: PreferenceManager
) {
    private val fusedLocationClient = LocationServices.getFusedLocationProviderClient(context)

    // Configurable intervals based on battery & accuracy needs
    private val locationRequest = LocationRequest.Builder(Priority.PRIORITY_HIGH_ACCURACY, 5000L)
        .setWaitForAccurateLocation(false)
        .setMinUpdateIntervalMillis(3000L)
        .setMaxUpdateDelayMillis(10000L) // Batching
        .setGranularity(Granularity.GRANULARITY_FINE)
        .build()

    private val locationCallback = object : LocationCallback() {
        override fun onLocationResult(result: LocationResult) {
            val location = result.lastLocation
            if (location != null && shouldUpload(location)) {
                _locations.tryEmit(location.toEntity())
            }
        }
    }

    fun start() {
        fusedLocationClient.requestLocationUpdates(
            locationRequest,
            locationCallback,
            Looper.getMainLooper()
        )
    }

    fun pause() {
        fusedLocationClient.removeLocationUpdates(locationCallback)
    }

    private fun shouldUpload(location: Location): Boolean {
        // Throttle logic: distance or time threshold
        val last = _lastLocation.value
        return if (last == null) true else {
            val distance = calculateDistance(last, location)
            val timeDiff = System.currentTimeMillis() - last.timestamp
            distance > 50 || timeDiff > 30000 // 50m or 30s
        }
    }
}
```

---

#### `:feature:alerts`
**SOS & Emergency Response**

**Priority System:**
- **CRITICAL** (SOS, accident): Immediate UI takeover, sound, vibration
- **HIGH** (Geofence breach, harsh braking): Full-screen modal, push
- **MEDIUM** (Maintenance due, low fuel): In-app notification, status bar
- **LOW** (New trip assigned, message): Snackbar, bell icon

**SOS Flow:**
```kotlin
class SOSViewModel @Inject constructor(
    private val emergencyService: EmergencyService,
    private val locationTracker: LocationTracker
) : ViewModel() {

    private val _sosState = MutableStateFlow(SOSState.Idle)
    val sosState: StateFlow<SOSState> = _sosState.asStateFlow()

    fun triggerSOS(confirmed: Boolean = false) {
        if (confirmed) {
            _sosState.value = SOSState.Triggering
            viewModelScope.launch {
                try {
                    // 1. Get current location
                    val location = locationTracker.getLastLocation()

                    // 2. Send SOS event to backend
                    emergencyService.triggerSOS(
                        SOSRequest(
                            driverId = currentDriver.id,
                            location = location,
                            timestamp = Instant.now(),
                            type = SOSType.MEDICAL // or POLICE, FIRE, ACCIDENT
                        )
                    )

                    // 3. Start background monitoring
                    emergencyService.startEmergencyMonitoring()

                    // 4. Show emergency UI
                    _sosState.value = SOSState.Triggered(
                        help Arriving = true,
                        timer = System.currentTimeMillis()
                    )

                    // 5. Cancelable for 10s window
                    delay(10000)
                    if (_sosState.value is SOSState.Triggered) {
                        autoCallEmergencyContacts()
                    }
                } catch (e: Exception) {
                    _sosState.value = SOSState.Error(e.message ?: "Failed to trigger SOS")
                }
            }
        } else {
            _sosState.value = SOSState.ConfirmationRequired
        }
    }

    fun cancelSOS() {
        emergencyService.cancelSOS()
        _sosState.value = SOSState.Idle
    }
}
```

**UI Overlay:** Full-screen emergency UI with:
- Countdown timer (count up from trigger)
- One-tap cancel button
- Auto-dialing emergency contacts after 30s
- Location sharing with responders
- Continuous background updates

---

#### `:feature:fuel`
**Fuel Logging & Expense Management**

**Features:**
- Fuel receipt OCR (Google ML Kit)
- Fuel card integration (sync with card provider)
- Mileage calculation (odometer + GPS distance)
- Cost per gallon/liter, efficiency (MPG/KPL)
- Receipt upload with auto-categorization

**OCR Pipeline:**
```kotlin
class FuelReceiptProcessor @Inject constructor(
    private val mlKit: MlKitContext
) {
    suspend fun extractFuelDetails(image: Bitmap): FuelReceiptData {
        // 1. Preprocess (rotation, contrast enhancement)
        val processed = preprocessImage(image)

        // 2. Text recognition
        val visionImage = InputImage.fromBitmap(processed, 0)
        val textResult = TextRecognition.getClient().process(visionImage)

        // 3. Parse receipt
        val lines = textResult.text.split("\n")
        val amount = extractAmount(lines)
        val gallons = extractGallons(lines)
        val pricePerGallon = amount / gallons
        val stationName = extractStationName(lines)
        val timestamp = extractDate(lines)

        // 4. Validate
        require(amount > 0 && gallons > 0) { "Invalid receipt data" }
        require(pricePerGallon in 2.0..10.0) { "Suspicious fuel price" }

        return FuelReceiptData(
            amount = amount,
            gallons = gallons,
            pricePerGallon = pricePerGallon,
            stationName = stationName,
            timestamp = timestamp,
            rawText = textResult.text
        )
    }
}
```

---

#### `:feature:documents`
**Document Management**

**Document Types:**
- Driver license, medical certificate, HOS logs
- Vehicle registration, insurance, permits
- Trip documents: BOL, POD, customs paperwork
- Fuel receipts, expense receipts

**Upload Strategy:**
```kotlin
class DocumentUploader @Inject constructor(
    private val api: DocumentApi,
    private val preferences: PreferenceManager
) {
    private val uploadQueue = MutableSharedFlow<UploadTask>(replay = 0)

    init {
        // Background worker for uploads
        viewModelScope.launch(Dispatchers.IO) {
            uploadQueue
                .filter { it.retryCount < MAX_RETRIES }
                .collect { task ->
                    try {
                        val result = api.uploadDocument(
                            documentId = task.documentId,
                            file = task.file,
                            metadata = task.metadata
                        )
                        if (result.isSuccessful) {
                            markUploaded(task.documentId)
                        }
                    } catch (e: Exception) {
                        // Exponential backoff retry
                        delay(min(5000 * 2.0.pow(task.retryCount), 300000))
                        uploadQueue.emit(task.copy(retryCount = task.retryCount + 1))
                    }
                }
        }
    }

    fun enqueueUpload(documentId: String, file: File, metadata: DocumentMetadata) {
        viewModelScope.launch {
            uploadQueue.emit(UploadTask(documentId, file, metadata, retryCount = 0))
        }
    }
}
```

**Compression & Optimization:**
- Images: Compress to 1080p, WebP format, <500KB
- PDFs: Compress with iText, optimize for size
- Videos: Trim to 30s max, 720p, H.264
- Batch uploads when on Wi-Fi (configurable)

---

## Scalability Patterns

### 1. Efficient Resource Usage

**Memory Management:**
- Coil/Glide for image loading with aggressive caching
- `LazyColumn` / `LazyRow` for infinite lists (recycling)
- ViewModel scoped to navigation graph (not Activity)
- `remember` / `derivedStateOf` to prevent recomposition storms
- Disposable observers in `onCleared()`

**Network Optimization:**
- GraphQL/REST with cursor-based pagination
- Request deduplication (same request in flight cached)
- HTTP/2 multiplexing (5 concurrent streams per connection)
- Response compression (gzip/brotli)
- Delta sync for updates (not full refresh)

**Battery Efficiency:**
- WorkManager for background sync (Doze mode compliant)
- AlarmManager with exact timing only for critical alerts
- Location batching (batched updates every 30s, not continuous)
- Adaptive frequency based on movement state
- Foreground service with ongoing notification for tracking

---

### 2. Horizontal Scaling Readiness

The app is designed to handle:
- **10k+ concurrent drivers** per backend cluster
- **100k+ daily active users**
- **1M+ telemetry points per minute** ingested

**Load Distribution:**
- Regional NATS clusters (us-central1, eu-west1, ap-south1)
- Sharded PostgreSQL by org_id (consistent hashing)
- Redis cluster for sessions and caching
- CDN (CloudFront/Cloudflare) for static assets

**Client-Side Resilience:**
- Exponential backoff with jitter on failures
- Circuit breaker pattern for flaky endpoints
- Graceful degradation: Offline mode when API unavailable
- Stale-while-revalidate: Show cached data while fetching fresh

---

### 3. Modularization for Team Scale

**Team Structure:**
- **Platform Team:** Core modules (`:core:*`)
- **Feature Teams:** One per module (`:feature:trips`, `:feature:tracking`, etc.)
- **QA Team:** Shared test infrastructure

**Independent Builds:**
```gradle
// features/trips/build.gradle
dependencies {
    implementation project(":core:auth")
    implementation project(":core:network")
    implementation project(":core:ui")
    implementation project(":core:data")
    implementation project(":core:database")
    // No dependency on other features
}
```

**Feature Flags:**
```kotlin
object FeatureFlags {
    val ENABLE_OFFLINE_MODE = RemoteFlag("offline_mode", default = true)
    val ENABLE_SOS_VIDEO = RemoteFlag("sos_video", default = false)
    val ENABLE_FUEL_OCR = RemoteFlag("fuel_ocr", default = true)
}
```

Configurable via Firebase Remote Config or custom endpoint.

---

## Performance Optimizations

### 1. App Startup (<2s Cold Startup)

**Startup Profiling:**
```kotlin
class DriverApp : Application() {
    override fun onCreate() {
        val startTime = SystemClock.uptimeMillis()

        // Lazy initialization
        val config = AppConfig.create(this)

        // Dependency injection (Hilt) - async where possible
        // Avoid heavy work on main thread

        Timber.plant(ReleaseTree())

        val endTime = SystemClock.uptimeMillis()
        Timber.d("App startup: ${endTime - startTime}ms")

        // If >2000ms, log crashlytics performance trace
    }
}
```

**Optimization Techniques:**
- **SplashScreen API** - Keep splash visible until ready (max 2s)
- **Lazy Hilt modules** - Inject only needed dependencies per feature
- **Non-initialized components:** Background services, analytics SDKs initialized asynchronously
- **App startup library:** Jetpack Startup for ordered component initialization
- **Avoid ProGuard regressions:** Baseline profiles for AOT compilation

**Target Metrics:**
- Cold start: <2s (90th percentile)
- Warm start: <500ms
- Time to interactive (TTI): <1.5s

---

### 2. UI Rendering (<100ms interactions)

**Compose Best Practices:**
- Avoid lambda creation in composable body (extract to stable)
- Use `remember` for expensive calculations
- `derivedStateOf` for derived state
- `key` in lazy lists for stable identity
- Avoid `BoxWithConstraints` in scrollable content

**Profiling:**
```bash
# Enable Compose compiler metrics
./gradlew assembleDebug -PcomposeCompilerMetrics=true
# Analyze recomposition counts in AS profile inspector
```

**Optimization Checklist:**
- [ ] No state reads during composition without `remember`
- [ ] Minimal state hoisting (avoid pass-through state)
- [ ] Lazy column items < 20ms to measure/layout/draw
- [ ] Image loading with Coil (placeholders, caching)
- [ ] Avoid `Modifier.drawBehind` on scrolling containers

---

### 3. Network Efficiency

**API Design:**
- **Batch endpoints** for bulk operations (upload multiple fuel logs)
- **Delta sync** with `updated_at` timestamps
- **WebSocket** for real-time (alerts, location sharing)
- **gRPC** for high-frequency telemetry (optional)

**Caching Strategy:**
```kotlin
enum class CacheStrategy {
    NETWORK_FIRST,     // Try network, fallback to cache
    CACHE_FIRST,       // Show cache immediately, refresh in bg
    CACHE_ONLY,        // Offline mode only
    NETWORK_ONLY       // Force refresh
}
```

**Implementation:**
```kotlin
class TripRepositoryImpl @Inject constructor(
    private val api: TripApi,
    private val dao: TripDao,
    private val cachePolicy: CachePolicy
) {
    fun getTrips(params: TripQueryParams): Flow<Result<List<Trip>>> = flow {
        when (cachePolicy) {
            CachePolicy.CACHE_FIRST -> {
                val cached = dao.getTrips(params)
                emit(Result.success(cached))
                // Background refresh
                try {
                    val remote = api.getTrips(params)
                    dao.upsertTrips(remote)
                } catch (_: Exception) {}
            }
            CachePolicy.NETWORK_FIRST -> {
                try {
                    val remote = api.getTrips(params)
                    dao.upsertTrips(remote)
                    emit(Result.success(remote))
                } catch (e: Exception) {
                    val cached = dao.getTrips(params)
                    emit(Result.success(cached))
                }
            }
        }
    }.flowOn(Dispatchers.IO)
}
```

---

## Security Design

### 1. Data Protection

**At Rest:**
- **SQLCipher** for local database (AES-256)
- **EncryptedSharedPreferences** for tokens (Android KeyStore)
- File-based encryption for uploaded documents (before transmission)
- Biometric/PIN as secondary auth for sensitive operations

**In Transit:**
- TLS 1.3 mandatory (certificate pinning)
- Certificate transparency validation
- JWT with RS256 (RSA 2048-bit)
- Short-lived access tokens (15min), refresh tokens (30 days)

**Code:**
```kotlin
object Security {
    private const val KEYSTORE_ALIAS = "flyfleet_key"

    fun getOrCreateKey(): SecretKey {
        val keyStore = KeyStore.getInstance("AndroidKeyStore")
        keyStore.load(null)

        return if (!keyStore.containsAlias(KEYSTORE_ALIAS)) {
            val keyGen = KeyGenerator.getInstance(
                KeyProperties.KEY_ALGORITHM_AES,
                "AndroidKeyStore"
            )
            val spec = KeyGenParameterSpec.Builder(
                KEYSTORE_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setKeySize(256)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setUserAuthenticationRequired(true) // Biometric required
                .build()
            keyGen.init(spec)
            keyGen.generateKey()
        } else {
            keyStore.getKey(KEYSTORE_ALIAS, null) as SecretKey
        }
    }
}
```

---

### 2. Authentication Flow

**Multi-Factor Authentication:**
1. Primary: Email + Password (argon2id on backend, PBKDF2 on client for derived key)
2. Secondary: TOTP (Google Authenticator) or SMS OTP
3. Tertiary: Biometric unlock for app access (non-repudiation)

**Session Management:**
```kotlin
class SessionManager @Inject constructor(
    private val tokenStorage: TokenStorage,
    private val api: AuthApi
) {
    private val refreshJob = Mutex()

    suspend fun getValidToken(): String = coroutineScope {
        val token = tokenStorage.getAccessToken()

        if (isTokenExpired(token)) {
            refreshJob.withLock {
                // Double-check after acquiring lock
                val current = tokenStorage.getAccessToken()
                if (isTokenExpired(current)) {
                    val refreshToken = tokenStorage.getRefreshToken()
                        ?: throw SessionExpiredException()

                    val response = api.refreshToken(RefreshRequest(refreshToken))
                    tokenStorage.saveTokens(
                        accessToken = response.accessToken,
                        refreshToken = response.refreshToken,
                        expiresIn = response.expiresIn
                    )
                }
            }
        }

        tokenStorage.getAccessToken()
    }
}
```

---

### 3. Privacy & Compliance

**GDPR/CCPA Compliance:**
- Data minimization: Store only essential data locally
- Right to be forgotten: Secure wipe on account deletion
- Data portability: Export all user data (JSON)
- Consent logging: Track opt-in for data collection

**Sensitive Data Handling:**
- Never log PII (use Timber's `Tree` filtering)
- Clear clipboard after copy operations (sensitive data)
- Prevent screenshots on PII screens (FLAG_SECURE)
- Mask sensitive fields in UI (driver license: `XXXX-XXXX-XXXX-1234`)

---

## Offline Strategy

### Offline-First Architecture

**Principle:** All screens must be usable without network.

**Implementation:**
1. **Local Database (Room):** Single source of truth
2. **Sync Queue:** Table of pending operations with retry logic
3. **Conflict Resolution:** Last-write-wins with server timestamps, or manual review for critical data
4. **Connectivity Awareness:** Reactive streams that switch data source based on network

**Sync Strategy:**

```
Operation created → Queue → Background Worker → Network Upload → Success/Failure
      ↓                                                         ↓
    Local DB (optimistic) ←─────────────────────────────────────┘
```

**Sync Worker (WorkManager):**
```kotlin
class SyncWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result = coroutineScope {
        try {
            val pendingOps = syncQueueDao.getPendingOperations(limit = 50)

            // Batch upload (reduces API calls)
            val uploadRequests = pendingOps.map { it.toRequest() }
            val responses = api.batchSync(uploadRequests)

            // Update local state based on server response
            responses.forEach { response ->
                when (response.status) {
                    SyncStatus.SUCCESS -> {
                        syncQueueDao.markSynced(response.operationId)
                        // Apply server-side mutations
                        applyRemoteChanges(response.changes)
                    }
                    SyncStatus.CONFLICT -> {
                        // Conflict resolution strategy
                        resolveConflict(response)
                    }
                    SyncStatus.FAILED -> {
                        // Increment retry count, maybe backoff
                        syncQueueDao.incrementRetry(response.operationId)
                    }
                }
            }

            Result.success()
        } catch (e: Exception) {
            Result.retry() // Exponential backoff handled by WorkManager
        }
    }
}
```

**Constraints:**
- Battery > 20%
- Network: unmetered (Wi-Fi preferred) or any (for critical alerts)
- Device charging: optional (default no)
- Backoff: 30s → 5min → 15min → 1hour

---

## Monitoring & Observability

### 1. App-Level Metrics

**Performance Monitoring:**
```kotlin
class PerformanceMonitor @Inject constructor() {
    private val traces = mutableMapOf<String, Long>()

    fun startTrace(name: String) {
        traces[name] = SystemClock.uptimeMillis()
    }

    fun endTrace(name: String) {
        val start = traces.remove(name) ?: return
        val duration = SystemClock.uptimeMillis() - start

        // Log to Firebase Performance Monitoring
        val trace = FirebasePerformance.getInstance()
            .newTrace(name)
        trace.stop()

        // Also send to mixpanel/amplitude for custom dashboards
        analyticsClient.track("$name.duration", mapOf("ms" to duration))
    }
}

// Usage
val monitor = PerformanceMonitor()
monitor.startTrace("dashboard.load")
viewModel.loadDashboard()
// ...
monitor.endTrace("dashboard.load")
```

**Key Traces:**
- Screen render time
- API call latency (by endpoint)
- Database query times
- Image load times
- Cold/warm startup times

---

### 2. Crash Reporting

**Firebase Crashlytics + Custom Keys:**
```kotlin
class CrashlyticsClient @Inject constructor(
    private val firebaseCrashlytics: FirebaseCrashlytics
) {
    fun setUser(driver: Driver) {
        firebaseCrashlytics.setUserId(driver.id)
        firebaseCrashlytics.setCustomKey("org_id", driver.orgId)
        firebaseCrashlytics.setCustomKey("vehicle_id", driver.vehicleId)
        firebaseCrashlytics.setCustomKey("app_version", BuildConfig.VERSION_NAME)
    }

    fun logHandledException(e: Exception, context: Map<String, Any>) {
        firebaseCrashlytics.recordException(e)
        context.forEach { (key, value) ->
            firebaseCrashlytics.setCustomKey(key, value.toString())
        }
    }
}
```

---

### 3. Real User Monitoring (RUM)

**Session Replay (Optional):**
- Record user interactions on high-friction screens
- Mask sensitive fields before upload
- Session duration > 30s, with user consent

**Analytics Events:**
```kotlin
sealed class AnalyticsEvent(val name: String, val properties: Map<String, Any> = emptyMap()) {
    data object AppStart : AnalyticsEvent("app.start")
    data class ScreenView(val screenName: String, val route: String) :
        AnalyticsEvent("screen.view", mapOf("screen" to screenName, "route" to route))

    data class TripStarted(val tripId: String, val vehicleId: String) :
        AnalyticsEvent("trip.started", mapOf("trip_id" to tripId, "vehicle_id" to vehicleId))

    data class SOSTriggered(val type: String) :
        AnalyticsEvent("sos.triggered", mapOf("type" to type))
}
```

---

## Technology Stack

### Android
- **Min SDK:** API 24 (Android 7.0) - 95%+ market share
- **Target SDK:** API 35 (Android 15)
- **Language:** Kotlin 1.9+ (Kotlin-first, no Java)
- **Jetpack Libraries:**
  - Compose BOM 2024.02.00
  - ViewModel, LiveData, Paging 3, Room, Hilt, WorkManager, Navigation

### UI
- **Compose Material 3** - Dynamic theming, accessibility
- **Accompanist** - System UI controller, insets
- **Coil** - Image loading (Compose + Glide under the hood)
- **Lottie** - Animated illustrations

### Networking
- **Retrofit 2.x** + OkHttp 4.x + Kotlin Coroutines
- **Moshi** / **Protobuf** (for telemetry)
- **WebSocket** (OkHttp) for real-time

### Architecture
- **MVVM** with Clean Architecture
- **Hilt** for dependency injection
- **Kotlin Coroutines + Flow** for async/reactive
- **Room + SQLCipher** for offline storage

### Testing
- **JUnit 5** + MockK + Turbine (flows testing)
- **Compose Testing** - createAndroidComposeRule
- **Espresso** for legacy UI tests (if needed)
- **MockWebServer** for API tests
- **Robolectric** for unit tests

### CI/CD
- **GitHub Actions**
  - Build, test, lint, detekt
  - Firebase App Distribution (beta)
  - Play Store (production)
  - Automated release notes

---

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
**Deliverables:**
- ✅ Set up Gradle module structure (9 core + 9 feature modules)
- ✅ Implement core modules (`:core:*`) with 80% test coverage
- ✅ Design system (Theme, Typography, Components)
- ✅ Authentication module (login, token management)
- ✅ CI/CD pipeline with beta distribution
- ✅ Baseline profiling and performance targets

**Team:** 2 Platform Engineers, 1 UI Designer

---

### Phase 2: Core Features (Weeks 5-10)
**Deliverables:**
- ✅ Dashboard module (KPI, quick actions)
- ✅ Trip management (list, details, execution)
- ✅ Real-time tracking (GPS, geofences)
- ✅ Fuel logging (manual + OCR)
- ✅ Alert system (SOS, push notifications)
- ✅ Offline sync framework (WorkManager + Room)
- ✅ 70%+ test coverage

**Team:** 3 Feature Engineers, 1 QA, 1 DevOps

---

### Phase 3: Polish & Scale (Weeks 11-14)
**Deliverables:**
- ✅ Document management (upload, sync, OCR)
- ✅ Profile & compliance (dashboard, settings)
- ✅ Command center (emergency tools)
- ✅ Performance optimization (<2s startup, <100ms UI)
- ✅ Security hardening (pinning, encrypt, biometric)
- ✅ Localization (English + top 5 languages)
- ✅ Accessibility (TalkBack, font scaling)

**Team:** 2 Feature Engineers, 1 Security Engineer, 1 Accessibility Specialist

---

### Phase 4: Production Readiness (Weeks 15-16)
**Deliverables:**
- ✅ Load testing (10k concurrent drivers simulation)
- ✅ Beta testing with pilot customers (100 drivers)
- ✅ Play Store listing (screenshots, descriptions, privacy policy)
- ✅ Crashlytics & Performance Monitoring setup
- ✅ Documentation (developer, deployment, troubleshooting)
- ✅ Support training & user guides

**Team:** Full cross-functional team

---

## Success Metrics (Technical)

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Cold start time** | <2s (P90) | Firebase Performance |
| **ANR rate** | <0.1% | Play Console |
| **Crash-free users** | >99.95% | Crashlytics |
| **Memory usage** | <150MB avg | Android Profiler |
| **Battery drain** | <5% per hour (active tracking) | Battery Historian |
| **API success rate** | >99.9% | Backend monitoring |
| **Sync completion** | >95% of ops sync within 5min | Custom metrics |
| **App size** | <50MB (APK) | Bundle size report |
| **Test coverage** | >70% (lines) | JaCoCo |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| **GPS battery drain** | Adaptive interval, batching, Doze-compliant |
| **Large telemetry uploads** | Compression, batching, Wi-Fi only option |
| **Poor connectivity in fleet yards** | Aggressive caching, queue for offline ops |
| **Device fragmentation** | Test on 5+ OEM devices (Samsung, Xiaomi, Pixel) |
| **Google Play rejection** | Follow policy, permissions justification, privacy policy |
| **Data privacy violations** | Encrypt everything, clear on logout, audit logs |
| **Low-end device performance** | Lazy loading, image downsampling, disable fancy animations |

---

## Conclusion

This scalable design follows Razorpay-like principles:

✅ **Modular** - 18 modules for parallel development
✅ **Clean** - Layered architecture, dependency rule enforced
✅ **Performant** - <2s startup, <100ms interactions
✅ **Scalable** - Handles 10k+ concurrent drivers
✅ **Resilient** - Offline-first, sync with conflict resolution
✅ **Secure** - End-to-end encryption, biometric auth
✅ **Observable** - Performance monitoring, crash reporting, RUM

**Next Steps:**
1. Create GitHub project with module structure
2. Set up CI/CD (GitHub Actions + Firebase App Distribution)
3. Implement core modules first (auth, network, data)
4. Build feature modules incrementally
5. Pilot with 10 internal drivers for feedback
6. Scale to production with staged rollout (10% → 50% → 100%)

---

**Architect:** AI Assistant (Razorpay-inspired design)
**Date:** 2026-03-31
**Review Required:** Mobile Engineering Lead, Backend Integration Team, Security Team

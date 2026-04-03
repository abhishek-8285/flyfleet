# Driver Mobile App Core - Product Requirements Document

**Document Version:** 1.0  
**Date:** March 31, 2026  
**Author:** Product Team  
**Status:** Draft for Review  
**Reviewers:** Engineering Lead, Design Lead, Security Team  

---

## Executive Summary

### Problem Statement
Professional drivers need a reliable, intuitive mobile application to manage their daily operations including trip execution, vehicle documentation, compliance reporting, and emergency response. Current solutions are fragmented, offline-unreliable, and lack the polish expected by modern users.

### Solution Overview
A production-grade Android driver application with **Razorpay-like UX** — clean, fast, trustworthy — that enables drivers to:
- Execute trips with turn-by-turn navigation
- Log fuel, hours-of-service (HOS), and incidents
- Access vehicle documents and compliance info
- Trigger SOS emergencies instantly
- Work seamlessly offline with automatic sync

### Target Users
| Persona | Description | Primary Goals |
|---------|-------------|---------------|
| **Rajesh the Long-Haul Driver** | 35-50 years, drives 10+ hours/day, low tech comfort | Simple UI, quick actions, minimal typing |
| **Priya the City Delivery Driver** | 25-40 years, urban routes, high trip volume | Fast check-in/out, quick fuel logging |
| **Amit the Fleet Owner-Operator** | 30-55 years, owns 2-5 vehicles, drives occasionally | Compliance tracking, expense documentation |

---

## Goals & Success Metrics

### Product Goals
1. **Reliability:** 99.5% uptime for core trip workflows
2. **Offline-First:** Full functionality without network, seamless sync on reconnect
3. **Speed:** <2s cold start, <500ms screen transitions
4. **Safety:** SOS trigger in <3 taps, <1s response time
5. **Compliance:** 100% FMCSA/India HOS compliance

### Success Metrics (KPIs)
| Metric | Target | Measurement |
|--------|--------|-------------|
| **Driver Adoption Rate** | >90% within 30 days | % of assigned drivers active weekly |
| **Trip Completion Rate** | >95% | Trips completed / Trips assigned |
| **Offline Sync Success** | >99% | Successful syncs / Total sync attempts |
| **SOS Response Time** | <60 seconds p95 | Time from trigger to fleet acknowledgment |
| **App Store Rating** | >4.5 stars | Google Play Store rating |
| **Crash-Free Sessions** | >99.5% | Sessions without crashes |
| **Fuel Log Accuracy** | >98% | Fuel logs matching receipts |

---

## Feature Requirements

### Epic 1: Authentication & Onboarding

#### Feature 1.1: Driver Login
**User Story:** As a driver, I want to log in quickly so I can start my trip.

**Requirements:**
- Email/phone + password login
- OTP-based login (SMS/WhatsApp)
- Biometric authentication (fingerprint/face)
- Remember device option (30-day session)
- MFA support for high-security orgs

**Acceptance Criteria:**
- [ ] Login completes in <3 seconds
- [ ] Biometric login works in <1 second
- [ ] Session persists across app restarts
- [ ] Failed login shows clear error message

**API Endpoints:**
```
POST /api/v1/auth/login
POST /api/v1/auth/otp/request
POST /api/v1/auth/otp/verify
POST /api/v1/auth/biometric/register
POST /api/v1/auth/biometric/verify
```

---

#### Feature 1.2: First-Time Onboarding
**User Story:** As a new driver, I want a guided setup so I understand how to use the app.

**Requirements:**
- 4-screen onboarding carousel (value prop)
- Permission requests (location, camera, notifications)
- Profile completion flow
- Vehicle assignment confirmation
- Quick tutorial on core actions

**Acceptance Criteria:**
- [ ] Onboarding completes in <2 minutes
- [ ] All required permissions explained before request
- [ ] Driver can skip and complete later
- [ ] Tutorial shows actual app UI (not illustrations)

---

### Epic 2: Trip Management

#### Feature 2.1: Trip Dashboard (Home Screen)
**User Story:** As a driver, I want to see my active and upcoming trips at a glance.

**Requirements:**
- Active trip card with progress indicator
- Upcoming trips list (next 7 days)
- Quick actions: Start Trip, Check Vehicle, Contact Dispatcher
- Trip status badges (On Time, Delayed, Completed)
- Earnings summary for completed trips (if applicable)

**Screen Layout:**
```
┌─────────────────────────────────────┐
│ [Menu]                    [Profile] │
│                                     │
│  Good Morning, Rajesh! 👋           │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ 🚛 ACTIVE TRIP                  │ │
│ │ Delhi → Mumbai (45% complete)   │ │
│ │ ETA: 6h 30m | 580 km remaining  │ │
│ │ [View Details] [Navigate]       │ │
│ └─────────────────────────────────┘ │
│                                     │
│  UPCOMING TRIPS                     │
│  ┌────────────────────────────────┐ │
│  │ Tomorrow, 6:00 AM              │ │
│  │ Mumbai → Pune | TATA-456      │ │
│  │ [Accept] [Decline]             │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌──────┐ ┌──────┐ ┌──────┐        │
│  │ 📋   │ │ ⛽   │ │ 🆘   │        │
│  │ Trips│ │ Fuel │ │ SOS  │        │
│  └──────┘ └──────┘ └──────┘        │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
GET /api/v1/trips/active
GET /api/v1/trips/upcoming
GET /api/v1/trips/history
```

---

#### Feature 2.2: Trip Details & Navigation
**User Story:** As a driver, I want to see trip details and navigate to destinations.

**Requirements:**
- Route visualization with stops
- Turn-by-turn navigation (Google Maps/Mapbox integration)
- Stop management (check-in/check-out at each stop)
- ETA updates based on traffic
- Document upload per stop (POD, BOL)
- Delay reporting with reason codes

**Screen Layout:**
```
┌─────────────────────────────────────┐
│ ← Back              [Complete Stop] │
│                                     │
│  STOP 2 OF 4                        │
│  ┌────────────────────────────────┐ │
│  │ 📍 Mumbai Warehouse            │ │
│  │ Check-in: 10:45 AM             │ │
│  │ Status: ✅ Checked In          │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 🗺️ NAVIGATION                 │ │
│  │ 120 km | 2h 15m               │ │
│  │ [Start Navigation]             │ │
│  └────────────────────────────────┘ │
│                                     │
│  ACTIONS                            │
│  ┌──────────┐ ┌──────────┐         │
│  │ 📸 Upload│ │ 📝 Notes  │         │
│  │  POD     │ │          │         │
│  └──────────┘ └──────────┘         │
│  ┌──────────┐ ┌──────────┐         │
│  │ ⏱️ Delay │ │ 📞 Contact│         │
│  │ Report   │ │ Dispatcher│         │
│  └──────────┘ └──────────┘         │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
GET /api/v1/trips/{tripId}
POST /api/v1/trips/{tripId}/stops/{stopId}/checkin
POST /api/v1/trips/{tripId}/stops/{stopId}/checkout
POST /api/v1/trips/{tripId}/stops/{stopId}/documents
POST /api/v1/trips/{tripId}/delay
PUT /api/v1/trips/{tripId}/status
```

---

#### Feature 2.3: Trip Start/End Workflow
**User Story:** As a driver, I want to easily start and complete trips.

**Requirements:**
- Pre-trip inspection checklist (vehicle condition)
- Odometer reading capture (photo + manual)
- Fuel level confirmation
- Digital signature for trip acceptance
- Post-trip inspection on completion
- Trip summary with earnings/durations

**Acceptance Criteria:**
- [ ] Pre-trip inspection completes in <2 minutes
- [ ] Photo capture works offline
- [ ] Signature pad responsive on low-end devices
- [ ] Trip summary generates instantly on completion

---

### Epic 3: Fuel Management

#### Feature 3.1: Fuel Logging
**User Story:** As a driver, I want to log fuel purchases quickly for reimbursement.

**Requirements:**
- Quick fuel log (<30 seconds)
- Photo receipt capture with OCR (future)
- Auto-fill from fuel card transactions
- Manual entry: liters, price, station, odometer
- GPS location auto-capture
- Expense category selection

**Screen Layout:**
```
┌─────────────────────────────────────┐
│ ← Back                   [Save]     │
│                                     │
│  LOG FUEL                           │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ ⛽ Station Name                │ │
│  │ Indian Oil, NH-48             │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌──────────┐ ┌──────────┐         │
│  │ Liters   │ │ Price    │         │
│  │ 45.5 L   │ │ ₹4,250   │         │
│  └──────────┘ └──────────┘         │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 📍 Location                    │ │
│  │ Mumbai-Pune Highway            │ │
│  │ ✅ Captured automatically      │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 📸 Receipt Photo               │ │
│  │ [Take Photo] [Upload]          │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 🔢 Odometer                    │ │
│  │ 45,230 km                      │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
POST /api/v1/fuel/logs
GET /api/v1/fuel/logs
POST /api/v1/fuel/logs/{id}/receipt
GET /api/v1/fuel/cards/transactions
```

---

### Epic 4: Hours of Service (HOS)

#### Feature 4.1: HOS Status Dashboard
**User Story:** As a driver, I want to see my remaining drive time to stay compliant.

**Requirements:**
- Real-time HOS countdown (India: 9h drive, 14h on-duty)
- Visual progress bars for each limit
- Break reminders (30 min after 5h drive)
- Violation warnings (30 min before limit)
- Daily/weekly summary view
- ELD integration (auto-duty status)

**Screen Layout:**
```
┌─────────────────────────────────────┐
│ ← Back                             │
│                                     │
│  HOURS OF SERVICE                   │
│  Today, March 31                    │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ ⏱️ DRIVING TIME                │ │
│  │ ████████░░░░░░░░ 6h 15m / 9h  │ │
│  │ 2h 45m remaining               │ │
│  └────────────────────────────────┘ │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 🕐 ON-DUTY TIME                │ │
│  │ ██████░░░░░░░░░ 8h 30m / 14h  │ │
│  │ 5h 30m remaining               │ │
│  └────────────────────────────────┘ │
│                                     │
│  STATUS: ✅ DRIVING                 │
│  Last break: 2h 15m ago             │
│  Next break due: 2h 45m             │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 📊 WEEKLY SUMMARY              │ │
│  │ Total drive: 42h / 48h limit  │ │
│  │ Total on-duty: 58h / 70h limit│ │
│  │ [View Details]                 │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
GET /api/v1/driver/hos/status
GET /api/v1/driver/hos/daily
GET /api/v1/driver/hos/weekly
POST /api/v1/driver/hos/status
```

---

#### Feature 4.2: Duty Status Changes
**User Story:** As a driver, I want to quickly change my duty status.

**Requirements:**
- One-tap status change (Driving, On-Duty, Sleeper, Off-Duty)
- Annotation for status changes (optional)
- Automatic detection from vehicle motion (ELD)
- Manual override with reason
- Audit trail of all changes

---

### Epic 5: Vehicle & Documents

#### Feature 5.1: Vehicle Assignment
**User Story:** As a driver, I want to see my assigned vehicle details.

**Requirements:**
- Vehicle info (make, model, year, registration)
- Current odometer reading
- Fuel level
- Last maintenance date
- Next service due
- Vehicle health status

**API Endpoints:**
```
GET /api/v1/vehicles/assigned
GET /api/v1/vehicles/{vehicleId}
GET /api/v1/vehicles/{vehicleId}/health
```

---

#### Feature 5.2: Document Vault
**User Story:** As a driver, I want to access my vehicle and personal documents.

**Requirements:**
- Vehicle documents (RC, insurance, permits, PUC)
- Driver documents (license, medical certificate, badges)
- Trip documents (e-way bills, challans)
- Expiration alerts (30 days before expiry)
- Offline access to all documents
- Quick share via WhatsApp/email

**Screen Layout:**
```
┌─────────────────────────────────────┐
│ ← Back                    [+ Upload]│
│                                     │
│  MY DOCUMENTS                       │
│                                     │
│  VEHICLE DOCUMENTS                  │
│  ┌────────────────────────────────┐ │
│  │ 📄 Registration Certificate    │ │
│  │ Valid until: Dec 2027 ✅       │ │
│  │ [View] [Download] [Share]      │ │
│  └────────────────────────────────┘ │
│  ┌────────────────────────────────┐ │
│  │ 📄 Insurance Policy            │ │
│  │ Valid until: Jun 2026 ⚠️       │ │
│  │ [View] [Download] [Share]      │ │
│  └────────────────────────────────┘ │
│                                     │
│  DRIVER DOCUMENTS                   │
│  ┌────────────────────────────────┐ │
│  │ 📄 Driving License             │ │
│  │ Valid until: Aug 2028 ✅       │ │
│  │ [View] [Download] [Share]      │ │
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
GET /api/v1/documents
GET /api/v1/documents/{id}
POST /api/v1/documents/upload
DELETE /api/v1/documents/{id}
```

---

### Epic 6: Alerts & Emergency

#### Feature 6.1: SOS Emergency Button
**User Story:** As a driver in distress, I want to trigger an emergency alert instantly.

**Requirements:**
- **3-tap maximum** to trigger SOS
- Hardware button support (volume buttons)
- Accidental trigger prevention (3-second hold)
- Immediate location + vehicle info sent
- Two-way communication with fleet manager
- Emergency contacts notified automatically
- Works offline (sends when reconnects)

**Screen Layout (Emergency Overlay):**
```
┌─────────────────────────────────────┐
│                                     │
│         🆘 SOS ACTIVATED 🆘         │
│                                     │
│  Help is on the way!                │
│                                     │
│  📍 Location sent to:               │
│  • Fleet Control Room               │
│  • Emergency Contact: Wife (+91...) │
│  • Nearest Hospital: 12 km          │
│                                     │
│  ┌────────────────────────────────┐ │
│  │ Your Location:                 │ │
│  │ NH-48, Near Pune Bypass        │ │
│  │ Accuracy: 15 meters            │ │
│  └────────────────────────────────┘ │
│                                     │
│  [Cancel False Alarm] [Call 112]    │
│                                     │
│  Fleet manager is calling you...    │
│  [Answer] [Decline]                 │
│                                     │
└─────────────────────────────────────┘
```

**API Endpoints:**
```
POST /api/v1/emergency/sos/trigger
POST /api/v1/emergency/sos/cancel
GET /api/v1/emergency/sos/status
POST /api/v1/emergency/sos/acknowledge
```

---

#### Feature 6.2: Alert Inbox
**User Story:** As a driver, I want to see and manage my alerts.

**Requirements:**
- Alert categories: Safety, Maintenance, Compliance, Trip
- Priority badges (Critical, Warning, Info)
- Acknowledge/resolve actions
- Alert history (30 days)
- Push notifications for critical alerts

**Alert Types:**
| Alert | Trigger | Action Required |
|-------|---------|-----------------|
| **Geofence Breach** | Exit designated zone | Acknowledge |
| **Speed Violation** | Exceed limit by 20% | Acknowledge |
| **Maintenance Due** | Service interval reached | Schedule |
| **Document Expiring** | <30 days to expiry | Upload new |
| **HOS Warning** | <1h to limit | Take break |
| **Harsh Event** | Harsh brake/accel | Review |

**API Endpoints:**
```
GET /api/v1/alerts
POST /api/v1/alerts/{id}/acknowledge
POST /api/v1/alerts/{id}/resolve
```

---

### Epic 7: Profile & Settings

#### Feature 7.1: Driver Profile
**User Story:** As a driver, I want to manage my personal information.

**Requirements:**
- Personal info (name, phone, email, address)
- Emergency contacts (name, relationship, phone)
- License details (number, expiry, classes)
- Profile photo
- Preferred language (Hindi, English, Tamil, etc.)
- Notification preferences

**API Endpoints:**
```
GET /api/v1/driver/profile
PUT /api/v1/driver/profile
PUT /api/v1/driver/emergency-contacts
```

---

#### Feature 7.2: App Settings
**User Story:** As a driver, I want to customize app behavior.

**Requirements:**
- **Notifications:** Toggle by category (alerts, trip updates, promotions)
- **Privacy:** Location sharing controls
- **Offline Mode:** Data usage settings
- **Language:** Multi-language support
- **Accessibility:** Font size, high contrast
- **About:** Version, licenses, support contact

---

### Epic 8: Offline & Sync

#### Feature 8.1: Offline-First Architecture
**User Story:** As a driver in low-connectivity areas, I want the app to work without internet.

**Requirements:**
- All core features work offline (trip, fuel, HOS, documents)
- Local database (Room) for all entities
- Mutation queue with retry logic
- Conflict resolution (server-wins with merge)
- Sync status indicator
- Manual sync trigger

**Sync States:**
- ✅ **Up to date** - All data synced
- 🔄 **Syncing...** - Upload/download in progress
- ⚠️ **Action required** - Conflicts or errors need attention
- ❌ **Sync failed** - Network error, will retry

**Technical Spec:**
```kotlin
// Sync Worker Configuration
class SyncWorkerConfig {
    val syncInterval = 30.seconds  // Periodic sync
    val retryBaseDelay = 2.seconds
    val retryMaxDelay = 5.minutes
    val maxRetries = 5
    val jitterPercent = 0.2
}

// Conflict Resolution
enum class ConflictStrategy {
    SERVER_WINS,      // Default for trips, HOS
    CLIENT_WINS,      // For notes, drafts
    MERGE,            // For comments, attachments
    MANUAL_RESOLVE    // For critical conflicts
}
```

---

## Screen Architecture

### Information Architecture

```
Driver App
├── Authentication Flow
│   ├── Login Screen
│   ├── OTP Verification
│   ├── Biometric Setup
│   └── Onboarding (4 screens)
│
├── Main App (Bottom Navigation)
│   ├── Home (Trip Dashboard)
│   ├── Trips (List + Details)
│   ├── Fuel (Log + History)
│   ├── HOS (Status + Logs)
│   └── Profile (Settings + Documents)
│
├── Trip Flow
│   ├── Trip Details
│   ├── Navigation View
│   ├── Stop Management
│   ├── Document Upload
│   └── Trip Summary
│
├── Emergency Flow
│   ├── SOS Trigger (Overlay)
│   ├── Alert Inbox
│   └── Emergency Contacts
│
└── Utility Screens
    ├── Vehicle Inspection
    ├── Delay Reporting
    ├── Photo Capture
    └── Signature Pad
```

### Navigation Structure

**Bottom Navigation (5 tabs):**
1. 🏠 **Home** - Active trip, quick actions
2. 📋 **Trips** - All trips, history
3. ⛽ **Fuel** - Log fuel, view history
4. 🕐 **HOS** - Hours of service
5. 👤 **Profile** - Settings, documents

**Floating Action Button (Contextual):**
- Home tab: Start Trip / SOS
- Trips tab: New Trip (if allowed)
- Fuel tab: Log Fuel
- HOS tab: Change Status

---

## Design System (Razorpay-Inspired)

### Visual Design Principles

1. **Clarity Over Cleverness** - Every element has a purpose
2. **Speed is a Feature** - Optimistic UI, instant feedback
3. **Trust Through Polish** - Pixel-perfect, consistent spacing
4. **Accessibility First** - WCAG 2.1 AA compliance

### Color Palette

```kotlin
// Primary Brand
val FlyFleetBlue = Color(0xFF2563EB)        // Primary actions
val FlyFleetBlueDark = Color(0xFF1D4ED8)    // Hover/pressed
val FlyFleetBlueLight = Color(0xFF60A5FA)   // Disabled/background

// Semantic Colors
val SuccessGreen = Color(0xFF10B981)        // Completed, safe
val WarningAmber = Color(0xFFF59E0B)        // Warnings, expiring
val ErrorRed = Color(0xFFEF4444)            // Errors, urgent
val InfoBlue = Color(0xFF3B82F6)            // Informational

// Neutral Palette
val Gray50 = Color(0xFFF9FAFB)              // Background
val Gray100 = Color(0xFFF3F4F6)             // Borders
val Gray500 = Color(0xFF6B7280)             // Secondary text
val Gray900 = Color(0xFF111827)             // Primary text

// Dark Mode
val DarkBg = Color(0xFF1F2937)
val DarkSurface = Color(0xFF374151)
```

### Typography

```kotlin
// Font Family
val FontFamily = "Inter"  // Clean, modern, highly legible

// Type Scale
val TextStyles = mapOf(
    "Display" to TextStyle(fontSize = 32.sp, fontWeight = Bold),
    "Headline" to TextStyle(fontSize = 24.sp, fontWeight = SemiBold),
    "Title" to TextStyle(fontSize = 20.sp, fontWeight = Medium),
    "Body" to TextStyle(fontSize = 16.sp, fontWeight = Regular),
    "Caption" to TextStyle(fontSize = 14.sp, fontWeight = Regular),
    "Button" to TextStyle(fontSize = 16.sp, fontWeight = SemiBold),
)
```

### Spacing System

```kotlin
// 8-point grid system
val Spacing = mapOf(
    "xs" to 4.dp,
    "sm" to 8.dp,
    "md" to 16.dp,
    "lg" to 24.dp,
    "xl" to 32.dp,
    "2xl" to 48.dp,
)
```

### Component Library

**Buttons:**
```kotlin
// Primary Button (Filled)
PrimaryButton(
    text = "Start Trip",
    onClick = { ... },
    enabled = true,
    loading = false,
    size = ButtonSize.Large,  // or Medium, Small
    icon = Icons.Rounded.PlayArrow,
)

// Secondary Button (Outlined)
SecondaryButton(
    text = "View Details",
    onClick = { ... },
)

// Tertiary Button (Text-only)
TextButton(
    text = "Skip",
    onClick = { ... },
)

// Destructive Button
DestructiveButton(
    text = "Cancel Trip",
    onClick = { ... },
)
```

**Cards:**
```kotlin
// Trip Card
TripCard(
    trip = trip,
    onClick = { navigateToDetails() },
    showActions = true,
)

// Alert Card
AlertCard(
    alert = alert,
    onAcknowledge = { ... },
    onResolve = { ... },
)
```

**Input Fields:**
```kotlin
// Text Field
TextField(
    value = text,
    onValueChange = { ... },
    label = "Station Name",
    placeholder = "Enter station name",
    error = errorMessage,
    leadingIcon = Icons.Rounded.Search,
)

// Number Field
NumberField(
    value = liters,
    onValueChange = { ... },
    label = "Liters",
    suffix = "L",
)

// Photo Upload
PhotoUpload(
    onCapture = { ... },
    onUpload = { ... },
    maxFiles = 3,
)
```

**Progress Indicators:**
```kotlin
// Linear Progress (HOS, Trip progress)
LinearProgress(
    progress = 0.65f,  // 0.0 - 1.0
    label = "6h 30m / 9h",
    color = ProgressColor.Warning,  // Green, Amber, Red
)

// Circular Progress (Loading)
CircularProgress(
    size = 48.dp,
    strokeWidth = 4.dp,
)

// Skeleton Loader
SkeletonLoader(
    shape = SkeletonShape.Card,
    shimmer = true,
)
```

**Dialogs & Bottom Sheets:**
```kotlin
// Confirmation Dialog
ConfirmationDialog(
    title = "Cancel Trip?",
    message = "This action cannot be undone.",
    confirmText = "Cancel Trip",
    dismissText = "Keep Driving",
    onConfirm = { ... },
    onDismiss = { ... },
    type = DialogType.Destructive,
)

// Bottom Sheet (Actions)
ActionBottomSheet(
    title = "Trip Actions",
    actions = listOf(
        Action("View Route", Icons.Rounded.Map),
        Action("Upload Document", Icons.Rounded.AttachFile),
        Action("Contact Dispatcher", Icons.Rounded.Phone),
    ),
    onActionSelected = { ... },
)
```

---

## Technical Architecture

### App Architecture (MVVM + Clean)

```
┌─────────────────────────────────────────┐
│              UI Layer                   │
│  (Composables / Views / ViewModels)     │
├─────────────────────────────────────────┤
│            Domain Layer                 │
│  (Use Cases / Business Logic)           │
├─────────────────────────────────────────┤
│            Data Layer                   │
│  (Repositories / API / Database)        │
└─────────────────────────────────────────┘
```

### Module Structure

```
apps/mobile/driver/android-driver/
├── app/
│   ├── src/main/
│   │   ├── AndroidManifest.xml
│   │   ├── kotlin/com/flyfleet/driver/
│   │   │   ├── DriverApp.kt                    # Application class
│   │   │   ├── di/                             # Dependency Injection
│   │   │   │   ├── AppModule.kt
│   │   │   │   ├── NetworkModule.kt
│   │   │   │   └── DatabaseModule.kt
│   │   │   ├── ui/                             # UI Layer
│   │   │   │   ├── theme/                      # Design System
│   │   │   │   │   ├── Color.kt
│   │   │   │   │   ├── Type.kt
│   │   │   │   │   └── Shape.kt
│   │   │   │   ├── components/                 # Reusable Components
│   │   │   │   │   ├── Buttons.kt
│   │   │   │   │   ├── Cards.kt
│   │   │   │   │   ├── Dialogs.kt
│   │   │   │   │   └── Inputs.kt
│   │   │   │   ├── screens/                    # Screen Implementations
│   │   │   │   │   ├── login/
│   │   │   │   │   ├── home/
│   │   │   │   │   ├── trips/
│   │   │   │   │   ├── fuel/
│   │   │   │   │   ├── hos/
│   │   │   │   │   ├── profile/
│   │   │   │   │   └── emergency/
│   │   │   │   └── navigation/                 # Navigation Graph
│   │   │   │       └── NavGraph.kt
│   │   │   ├── domain/                         # Business Logic
│   │   │   │   ├── models/                     # Domain Models
│   │   │   │   │   ├── Trip.kt
│   │   │   │   │   ├── Driver.kt
│   │   │   │   │   ├── Vehicle.kt
│   │   │   │   │   └── Alert.kt
│   │   │   │   ├── repositories/               # Repository Interfaces
│   │   │   │   │   ├── TripRepository.kt
│   │   │   │   │   ├── FuelRepository.kt
│   │   │   │   │   └── AuthRepository.kt
│   │   │   │   └── usecases/                   # Use Cases
│   │   │   │       ├── StartTripUseCase.kt
│   │   │   │       ├── LogFuelUseCase.kt
│   │   │   │       └── TriggerSOSUseCase.kt
│   │   │   ├── data/                           # Data Layer
│   │   │   │   ├── api/                        # API Client
│   │   │   │   │   ├── ApiService.kt
│   │   │   │   │   ├── models/                 # API Models
│   │   │   │   │   └── interceptors/           # Auth, Logging
│   │   │   │   ├── local/                      # Local Database
│   │   │   │   │   ├── DriverDatabase.kt
│   │   │   │   │   ├── dao/                    # Data Access Objects
│   │   │   │   │   └── entities/               # Room Entities
│   │   │   │   ├── repositories/               # Repository Implementations
│   │   │   │   └── sync/                       # Sync Engine
│   │   │   │       ├── SyncWorker.kt
│   │   │   │       ├── MutationQueue.kt
│   │   │   │       └── ConflictResolver.kt
│   │   │   ├── service/                        # Background Services
│   │   │   │   ├── LocationTrackingService.kt
│   │   │   │   ├── SyncForegroundService.kt
│   │   │   │   └── SOSBroadcastReceiver.kt
│   │   │   └── util/                           # Utilities
│   │   │       ├── Constants.kt
│   │   │       ├── Extensions.kt
│   │   │       └── PermissionHelper.kt
│   │   └── res/                                # Resources
│   │       ├── values/                         # Strings, Colors, Themes
│   │       ├── drawable/                       # Icons, Images
│   │       └── raw/                            # Animations, Lottie
│   └── build.gradle.kts
├── build.gradle.kts
└── settings.gradle.kts
```

### Key Dependencies

```kotlin
// build.gradle.kts
dependencies {
    // Core Android
    implementation("androidx.core:core-ktx:1.12.0")
    implementation("androidx.appcompat:appcompat:1.6.1")
    implementation("com.google.android.material:material:1.11.0")
    
    // Jetpack Compose (Modern UI)
    implementation(platform("androidx.compose:compose-bom:2024.02.00"))
    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.material3:material3")
    implementation("androidx.compose.ui:ui-tooling-preview")
    
    // Navigation
    implementation("androidx.navigation:navigation-compose:2.7.7")
    
    // Architecture Components
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.7.0")
    implementation("androidx.lifecycle:lifecycle-runtime-compose:2.7.0")
    
    // Dependency Injection (Hilt)
    implementation("com.google.dagger:hilt-android:2.50")
    kapt("com.google.dagger:hilt-compiler:2.50")
    implementation("androidx.hilt:hilt-navigation-compose:1.1.0")
    
    // Networking
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.12.0")
    
    // Local Database (Room)
    implementation("androidx.room:room-runtime:2.6.1")
    implementation("androidx.room:room-ktx:2.6.1")
    kapt("androidx.room:room-compiler:2.6.1")
    
    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-play-services:1.7.3")
    
    // Location
    implementation("com.google.android.gms:play-services-location:21.1.0")
    
    // Maps
    implementation("com.google.maps.android:maps-compose:4.3.0")
    
    // Image Loading
    implementation("io.coil-kt:coil-compose:2.5.0")
    
    // Security
    implementation("androidx.security:security-crypto:1.1.0-alpha06")
    implementation("androidx.biometric:biometric:1.1.0")
    
    // Work Manager (Background Sync)
    implementation("androidx.work:work-runtime-ktx:2.9.0")
    
    // Logging
    implementation("com.jakewharton.timber:timber:5.0.1")
    
    // Testing
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.mockito:mockito-core:5.10.0")
    androidTestImplementation("androidx.test.ext:junit:1.1.5")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.5.1")
}
```

---

## API Contract Summary

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Email/password login |
| POST | `/api/v1/auth/otp/request` | Request OTP |
| POST | `/api/v1/auth/otp/verify` | Verify OTP |
| POST | `/api/v1/auth/biometric/register` | Register biometric |
| POST | `/api/v1/auth/biometric/verify` | Biometric login |
| POST | `/api/v1/auth/logout` | Logout |
| POST | `/api/v1/auth/refresh` | Refresh token |

### Trips
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/trips/active` | Get active trip |
| GET | `/api/v1/trips/upcoming` | Get upcoming trips |
| GET | `/api/v1/trips/history` | Get trip history |
| GET | `/api/v1/trips/{id}` | Get trip details |
| POST | `/api/v1/trips/{id}/start` | Start trip |
| POST | `/api/v1/trips/{id}/complete` | Complete trip |
| POST | `/api/v1/trips/{id}/stops/{stopId}/checkin` | Check in at stop |
| POST | `/api/v1/trips/{id}/stops/{stopId}/checkout` | Check out at stop |
| POST | `/api/v1/trips/{id}/delay` | Report delay |

### Fuel
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/fuel/logs` | Get fuel logs |
| POST | `/api/v1/fuel/logs` | Create fuel log |
| GET | `/api/v1/fuel/cards/transactions` | Get fuel card transactions |

### HOS
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/driver/hos/status` | Get HOS status |
| GET | `/api/v1/driver/hos/daily` | Get daily HOS log |
| GET | `/api/v1/driver/hos/weekly` | Get weekly HOS summary |
| POST | `/api/v1/driver/hos/status` | Update duty status |

### Documents
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/documents` | List documents |
| GET | `/api/v1/documents/{id}` | Get document |
| POST | `/api/v1/documents/upload` | Upload document |
| DELETE | `/api/v1/documents/{id}` | Delete document |

### Alerts & Emergency
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/alerts` | Get alerts |
| POST | `/api/v1/alerts/{id}/acknowledge` | Acknowledge alert |
| POST | `/api/v1/emergency/sos/trigger` | Trigger SOS |
| POST | `/api/v1/emergency/sos/cancel` | Cancel false alarm |
| GET | `/api/v1/emergency/sos/status` | Get SOS status |

### Driver Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/driver/profile` | Get profile |
| PUT | `/api/v1/driver/profile` | Update profile |
| GET | `/api/v1/vehicles/assigned` | Get assigned vehicle |

---

## Security Requirements

### Authentication & Authorization
1. **JWT Tokens** - Access token (15 min) + Refresh token (7 days)
2. **Biometric** - Android BiometricPrompt with fallback to PIN
3. **Session Management** - Max 3 concurrent sessions per driver
4. **Token Storage** - EncryptedSharedPreferences
5. **Certificate Pinning** - Prevent MITM attacks

### Data Protection
1. **Encryption at Rest** - SQLCipher for Room database
2. **Encryption in Transit** - TLS 1.3 with certificate pinning
3. **Sensitive Fields** - Encrypt PII (phone, license number)
4. **Secure Wipe** - Remote wipe capability for lost devices

### Compliance
1. **GDPR/DPDP** - Data export, right to deletion
2. **FMCSA ELD** - Tamper-proof audit trail
3. **India HOS** - Compliance with Motor Vehicles Act

---

## Performance Requirements

### App Performance
| Metric | Target | Measurement |
|--------|--------|-------------|
| Cold Start | <2 seconds | Time to first frame |
| Warm Start | <1 second | Time to interactive |
| Screen Transition | <300ms | Navigation latency |
| API Response (p95) | <3 seconds | Network latency |
| Memory Usage | <150MB | RSS memory |
| CPU Usage | <20% idle | Background operation |
| Battery Drain | <3%/hour | Active usage |

### Offline Performance
| Metric | Target |
|--------|--------|
| Local Query | <100ms |
| Mutation Queue | Unlimited (storage-bound) |
| Sync Success Rate | >99% |
| Conflict Rate | <1% |

---

## Accessibility Requirements

### WCAG 2.1 AA Compliance
1. **Text Size** - Support system font scaling up to 200%
2. **Color Contrast** - Minimum 4.5:1 for normal text
3. **Touch Targets** - Minimum 48x48 dp
4. **Screen Reader** - Full TalkBack support
5. **Keyboard Navigation** - External keyboard support
6. **Reduced Motion** - Respect system animation settings

---

## Localization

### Supported Languages (Phase 1)
1. **English** (Default)
2. **Hindi** (हिन्दी)
3. **Tamil** (தமிழ்)
4. **Telugu** (తెలుగు)
5. **Kannada** (ಕನ್ನಡ)

### Localization Strategy
- All strings in `strings.xml` with context
- RTL support (for future Arabic markets)
- Number/date formatting per locale
- Localized error messages

---

## Testing Strategy

### Test Pyramid
```
          /\
         /  \        E2E Tests (10%)
        /----\       UI flows, critical paths
       /      \
      /--------\     Integration Tests (20%)
     /          \    API, Database, Sync
    /------------\
   /              \   Unit Tests (70%)
  /----------------\  Use cases, repositories
```

### Test Coverage Targets
| Layer | Coverage Target |
|-------|-----------------|
| Domain (Use Cases) | 90% |
| Data (Repositories) | 80% |
| UI (ViewModels) | 70% |
| UI (Composables) | 50% |

### Key Test Scenarios
1. **Offline Sync** - Create trip offline, reconnect, verify sync
2. **Conflict Resolution** - Edit same trip on two devices
3. **SOS Trigger** - Verify emergency flow end-to-end
4. **HOS Compliance** - Verify limit calculations
5. **Battery Drain** - 8-hour continuous operation test

---

## Rollout Plan

### Phase 1: MVP (6 weeks)
**Features:**
- Authentication (login, OTP, biometric)
- Trip dashboard + details
- Trip start/complete workflow
- Fuel logging
- HOS status + duty changes
- Basic profile
- Offline sync (core entities)

**Target:** 100 drivers (pilot fleet)

---

### Phase 2: Enhanced Operations (4 weeks)
**Features:**
- Navigation integration (Google Maps)
- Stop management (check-in/out)
- Document upload (POD, BOL)
- Alert inbox
- Vehicle inspection
- Delay reporting

**Target:** 500 drivers (expanded pilot)

---

### Phase 3: Safety & Compliance (4 weeks)
**Features:**
- SOS emergency button
- HOS violation warnings
- Document expiration alerts
- ELD integration (auto-duty status)
- Multi-language support

**Target:** 2,000 drivers (regional rollout)

---

### Phase 4: Scale & Polish (4 weeks)
**Features:**
- Performance optimization
- Accessibility improvements
- Advanced analytics
- Receipt OCR (fuel)
- Voice commands (beta)

**Target:** 10,000+ drivers (national rollout)

---

## Open Questions

| Question | Owner | Due Date |
|----------|-------|----------|
| Fuel card integration partner? | Partnerships | Apr 15 |
| ELD hardware vendor selection? | Hardware Team | Apr 10 |
| SMS provider for OTP (India)? | Backend Team | Apr 5 |
| Map provider (Google vs Mapbox)? | Engineering | Apr 5 |
| Biometric fallback flow for low-end devices? | Design | Apr 3 |

---

## Appendix

### A. Competitive Analysis
| Feature | FlyFleet | BlackBuck | Rivigo | Porter |
|---------|----------|-----------|--------|--------|
| Offline Mode | ✅ | ⚠️ Limited | ❌ | ⚠️ Limited |
| SOS Button | ✅ | ❌ | ✅ | ❌ |
| Multi-language | ✅ (5) | ✅ (2) | ❌ | ✅ (3) |
| Fuel OCR | 🔜 | ❌ | ❌ | ❌ |
| ELD Integration | ✅ | ✅ | ✅ | ❌ |
| Razorpay-like UX | ✅ | ❌ | ❌ | ⚠️ |

### B. User Research Insights
- **Top Driver Pain Point:** "App doesn't work on highway (no network)"
- **Most Requested Feature:** "One-tap fuel log with photo"
- **Safety Concern:** "How do I get help if stranded at night?"
- **Compliance Worry:** "Will I get fined if HOS calculation is wrong?"

### C. Technical Debt Register
1. **Legacy Code:** vehicle-device module needs refactor
2. **Testing Gap:** Zero E2E tests currently
3. **Performance:** Cold start >5s on low-end devices
4. **Security:** Certificate pinning not implemented

---

## Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Product Manager | | | |
| Engineering Lead | | | |
| Design Lead | | | |
| Security Lead | | | |
| QA Lead | | | |

---

**Document History:**
- v1.0 (Mar 31, 2026) - Initial draft

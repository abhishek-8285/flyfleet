# Driver Mobile App - Design Specification

**Version:** 1.0  
**Date:** March 31, 2026  
**Design System:** FlyFleet Mobile Design System (Razorpay-inspired)  
**Platform:** Android (Jetpack Compose)  

---

## Design Principles

### 1. Clarity Over Cleverness
Every element serves a purpose. No decorative UI. Drivers need information fast, especially while operating vehicles.

### 2. Speed is a Feature
- Optimistic UI updates (show success before API confirms)
- Skeleton loaders instead of spinners
- Prefetch likely next screens
- Cache aggressively

### 3. Trust Through Polish
- Consistent 8px grid
- Pixel-perfect alignment
- Smooth animations (but respect reduced motion)
- Error states are helpful, not scary

### 4. Accessibility First
- WCAG 2.1 AA minimum
- 48dp touch targets
- 4.5:1 contrast ratio
- Full TalkBack support

---

## Color System

### Primary Palette

```kotlin
// src/main/java/com/flyfleet/driver/ui/theme/Color.kt

// Brand Primary
val FlyFleetBlue = Color(0xFF2563EB)        // Primary actions, links
val FlyFleetBlueDark = Color(0xFF1D4ED8)    // Hover, pressed states
val FlyFleetBlueLight = Color(0xFF60A5FA)   // Disabled, backgrounds
val FlyFleetBlueSubtle = Color(0xFFDBEAFE)  // Subtle backgrounds

// Semantic Colors
val SuccessGreen = Color(0xFF10B981)        // Completed, safe, on-time
val SuccessGreenLight = Color(0xFFD1FAE5)
val WarningAmber = Color(0xFFF59E0B)        // Warnings, expiring soon
val WarningAmberLight = Color(0xFFFEF3C7)
val ErrorRed = Color(0xFFEF4444)            // Errors, urgent, violations
val ErrorRedLight = Color(0xFFFEE2E2)
val InfoBlue = Color(0xFF3B82F6)            // Informational messages
val InfoBlueLight = Color(0xFFDBEAFE)

// Neutral Grays
val Gray50 = Color(0xFFF9FAFB)              // App background
val Gray100 = Color(0xFFF3F4F6)             // Card backgrounds, borders
val Gray200 = Color(0xFFE5E7EB)             // Dividers
val Gray300 = Color(0xFFD1D5DB)             // Disabled borders
val Gray400 = Color(0xFF9CA3AF)             // Disabled text
val Gray500 = Color(0xFF6B7280)             // Secondary text
val Gray600 = Color(0xFF4B5563)             // Body text
val Gray700 = Color(0xFF374151)             // Primary text
val Gray800 = Color(0xFF1F2937)
val Gray900 = Color(0xFF111827)             // Headings

// Dark Mode
val DarkBg = Color(0xFF0F172A)
val DarkSurface = Color(0xFF1E293B)
val DarkSurfaceElevated = Color(0xFF334155)
```

### Usage Guidelines

| Use Case | Light Mode | Dark Mode |
|----------|------------|-----------|
| **Primary Button** | FlyFleetBlue | FlyFleetBlue |
| **Primary Button Pressed** | FlyFleetBlueDark | FlyFleetBlueLight |
| **Success State** | SuccessGreen | SuccessGreen |
| **Warning Badge** | WarningAmber | WarningAmber |
| **Error State** | ErrorRed | ErrorRed |
| **Background** | Gray50 | DarkBg |
| **Card Surface** | White | DarkSurface |
| **Primary Text** | Gray900 | Gray50 |
| **Secondary Text** | Gray500 | Gray400 |

---

## Typography

### Font Family

```kotlin
// Using Inter font for clean, modern legibility
// Fallback: Roboto (system default on Android)

val InterFontFamily = FontFamily(
    regular = Font(R.font.inter_regular, FontWeight.Normal),
    medium = Font(R.font.inter_medium, FontWeight.Medium),
    semiBold = Font(R.font.inter_semibold, FontWeight.SemiBold),
    bold = Font(R.font.inter_bold, FontWeight.Bold),
)
```

### Type Scale

```kotlin
// src/main/java/com/flyfleet/driver/ui/theme/Type.kt

val Typography = Typography(
    // Display - Large headers
    displayLarge = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 32.sp,
        lineHeight = 40.sp,
        fontWeight = FontWeight.Bold,
        letterSpacing = (-0.5).sp,
    ),
    displayMedium = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 28.sp,
        lineHeight = 36.sp,
        fontWeight = FontWeight.Bold,
        letterSpacing = (-0.25).sp,
    ),
    
    // Headline - Section headers
    headlineLarge = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 24.sp,
        lineHeight = 32.sp,
        fontWeight = FontWeight.SemiBold,
        letterSpacing = 0.sp,
    ),
    headlineMedium = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 20.sp,
        lineHeight = 28.sp,
        fontWeight = FontWeight.SemiBold,
        letterSpacing = 0.sp,
    ),
    
    // Title - Card titles, screen titles
    titleLarge = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 18.sp,
        lineHeight = 24.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.15.sp,
    ),
    titleMedium = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 16.sp,
        lineHeight = 22.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.15.sp,
    ),
    titleSmall = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.1.sp,
    ),
    
    // Body - Main content text
    bodyLarge = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 16.sp,
        lineHeight = 24.sp,
        fontWeight = FontWeight.Normal,
        letterSpacing = 0.5.sp,
    ),
    bodyMedium = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        fontWeight = FontWeight.Normal,
        letterSpacing = 0.25.sp,
    ),
    bodySmall = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 12.sp,
        lineHeight = 16.sp,
        fontWeight = FontWeight.Normal,
        letterSpacing = 0.4.sp,
    ),
    
    // Label - Buttons, tabs, captions
    labelLarge = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.1.sp,
    ),
    labelMedium = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 12.sp,
        lineHeight = 16.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.5.sp,
    ),
    labelSmall = TextStyle(
        fontFamily = InterFontFamily,
        fontSize = 11.sp,
        lineHeight = 16.sp,
        fontWeight = FontWeight.Medium,
        letterSpacing = 0.5.sp,
    ),
)
```

### Usage Examples

| Style | Use Case | Example |
|-------|----------|---------|
| `displayLarge` | App bar titles | "Trip Details" |
| `headlineLarge` | Section headers | "Upcoming Trips" |
| `titleLarge` | Card titles | "Delhi → Mumbai" |
| `bodyLarge` | Body text | Trip descriptions |
| `labelLarge` | Buttons | "Start Trip" |
| `labelSmall` | Captions | "Updated 5m ago" |

---

## Spacing System

### 8-Point Grid

```kotlin
// All spacing follows 8px grid (4px for tight spaces)

val Spacing = object {
    val xs = 4.dp    // Tight spacing (icon to text)
    val sm = 8.dp    // Default spacing (list items)
    val md = 16.dp   // Card padding
    val lg = 24.dp   // Section spacing
    val xl = 32.dp   // Large section
    val xxl = 48.dp  // Major sections
}
```

### Layout Spacing

```
Screen Padding: 16dp (md)
Card Padding: 16dp (md)
List Item Padding: 16dp horizontal, 12dp vertical
Between Cards: 16dp (md)
Between Sections: 24dp (lg)
```

---

## Component Library

### Buttons

#### Primary Button (Filled)

```kotlin
@Composable
fun PrimaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    loading: Boolean = false,
    icon: ImageVector? = null,
    size: ButtonSize = ButtonSize.Large,
)
```

**Specs:**
- Height: Large (56dp), Medium (48dp), Small (40dp)
- Corner Radius: 12dp
- Background: FlyFleetBlue (enabled), Gray300 (disabled)
- Text: White (enabled), Gray500 (disabled)
- Elevation: 0dp (flat design)

**Usage:**
```kotlin
PrimaryButton(
    text = "Start Trip",
    onClick = { viewModel.startTrip() },
    icon = Icons.Rounded.PlayArrow,
    size = ButtonSize.Large,
)
```

---

#### Secondary Button (Outlined)

```kotlin
@Composable
fun SecondaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    icon: ImageVector? = null,
)
```

**Specs:**
- Height: Same as Primary
- Border: 2dp FlyFleetBlue
- Background: Transparent
- Text: FlyFleetBlue

**Usage:**
```kotlin
SecondaryButton(
    text = "View Details",
    onClick = { navigateToDetails() },
)
```

---

#### Destructive Button

```kotlin
@Composable
fun DestructiveButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
)
```

**Specs:**
- Background: ErrorRed
- Text: White

**Usage:**
```kotlin
DestructiveButton(
    text = "Cancel Trip",
    onClick = { showCancelDialog() },
)
```

---

#### Icon Button

```kotlin
@Composable
fun IconButton(
    icon: ImageVector,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    contentDescription: String,
    size: IconButtonSize = IconButtonSize.Medium,
)
```

**Specs:**
- Size: Medium (48dp), Large (56dp)
- Icon: 24dp
- Touch Target: Minimum 48dp

---

### Cards

#### Trip Card

```kotlin
@Composable
fun TripCard(
    trip: Trip,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    showActions: Boolean = true,
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│  🚛 Delhi → Mumbai                  │  ← Title (titleLarge)
│  1,450 km | ETA: 18h 30m            │  ← Subtitle (bodyMedium, Gray500)
│                                     │
│  ████████░░░░░░░░░░ 45%             │  ← Progress bar
│  650 km completed                   │  ← Progress text (bodySmall)
│                                     │
│  ┌──────────┐ ┌──────────┐         │
│  │ [Navigate]│ │ [Details]│         │  ← Action buttons
│  └──────────┘ └──────────┘         │
└─────────────────────────────────────┘
```

**Specs:**
- Background: White (light), DarkSurface (dark)
- Corner Radius: 16dp
- Padding: 16dp
- Elevation: 0dp (flat with border)
- Border: 1dp Gray200

---

#### Alert Card

```kotlin
@Composable
fun AlertCard(
    alert: Alert,
    onAcknowledge: () -> Unit,
    onResolve: () -> Unit,
    modifier: Modifier = Modifier,
)
```

**Layout by Priority:**
```
┌─────────────────────────────────────┐
│  🚨 CRITICAL                        │  ← Priority badge (ErrorRed background)
│  Harsh Brake Detected               │  ← Title
│  Today, 10:45 AM | NH-48, Km 245    │  ← Subtitle
│                                     │
│  Speed dropped from 80 to 30 km/h   │  ← Description (bodyMedium)
│  in 3 seconds.                      │
│                                     │
│  [Acknowledge] [Dismiss]            │  ← Actions
└─────────────────────────────────────┘
```

**Priority Colors:**
| Priority | Badge BG | Icon |
|----------|----------|------|
| Critical | ErrorRedLight | ErrorRed |
| Warning | WarningAmberLight | WarningAmber |
| Info | InfoBlueLight | InfoBlue |

---

### Input Fields

#### Text Field

```kotlin
@Composable
fun FlyFleetTextField(
    value: String,
    onValueChange: (String) -> Unit,
    modifier: Modifier = Modifier,
    label: String,
    placeholder: String? = null,
    error: String? = null,
    leadingIcon: ImageVector? = null,
    trailingIcon: ImageVector? = null,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default,
    maxLines: Int = 1,
)
```

**States:**
```
Default:
┌─────────────────────────────────────┐
│  Station Name                       │  ← Label (labelSmall, Gray500)
│  Enter station name                 │  ← Placeholder (Gray400)
└─────────────────────────────────────┘
Border: 1dp Gray300

Focused:
┌─────────────────────────────────────┐
│  Station Name                       │
│  Indian Oil │                       │  ← Cursor blinking
└─────────────────────────────────────┘
Border: 2dp FlyFleetBlue

Error:
┌─────────────────────────────────────┐
│  Station Name                       │
│  This field is required             │  ← Error text (ErrorRed)
└─────────────────────────────────────┘
Border: 2dp ErrorRed
```

**Specs:**
- Min Height: 56dp
- Corner Radius: 12dp
- Padding: 16dp
- Label: Always visible (no floating label)

---

#### Number Field

```kotlin
@Composable
fun NumberField(
    value: String,
    onValueChange: (String) -> Unit,
    modifier: Modifier = Modifier,
    label: String,
    suffix: String? = null,
    prefix: String? = null,
    error: String? = null,
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│  Liters                    [L]      │  ← Suffix on right
│  45.5                               │
└─────────────────────────────────────┘
```

---

#### Photo Upload

```kotlin
@Composable
fun PhotoUpload(
    photos: List<Uri>,
    onCapture: () -> Unit,
    onUpload: () -> Unit,
    onRemove: (Uri) -> Unit,
    modifier: Modifier = Modifier,
    maxPhotos: Int = 3,
    label: String = "Add Photos",
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│  Receipt Photo                      │
│                                     │
│  ┌───────┐ ┌───────┐ ┌───────┐     │
│  │  📷   │ │       │ │       │     │
│  │ Add   │ │       │ │       │     │
│  └───────┘ └───────┘ └───────┘     │
│                                     │
│  [Take Photo] [Upload from Gallery] │
└─────────────────────────────────────┘
```

**With Photos:**
```
┌─────────────────────────────────────┐
│  Receipt Photo                      │
│                                     │
│  ┌───────┐ ┌───────┐ ┌───────┐     │
│  │ [IMG] │ │ [IMG] │ │  📷   │     │
│  │   ×   │ │   ×   │ │  Add  │     │
│  └───────┘ └───────┘ └───────┘     │
└─────────────────────────────────────┘
```

---

### Progress Indicators

#### Linear Progress

```kotlin
@Composable
fun FlyFleetLinearProgress(
    progress: Float,  // 0.0f to 1.0f
    modifier: Modifier = Modifier,
    label: String? = null,
    showValue: Boolean = false,
    color: ProgressColor = ProgressColor.Primary,
)
```

**Layout:**
```
HOS Progress:
┌─────────────────────────────────────┐
│  ⏱️ Driving Time                    │
│  ████████░░░░░░░░░░ 6h 15m / 9h    │  ← Progress bar + label
│  2h 45m remaining                   │
└─────────────────────────────────────┘
```

**Colors:**
| Color | Use Case |
|-------|----------|
| Primary (Blue) | Trip progress |
| Success (Green) | On track, under limit |
| Warning (Amber) | Approaching limit (<2h) |
| Error (Red) | Over limit, violation |

**Specs:**
- Height: 8dp
- Track: Gray200
- Corner Radius: 4dp (rounded)

---

#### Circular Progress (Loading)

```kotlin
@Composable
fun FlyFleetCircularProgress(
    modifier: Modifier = Modifier,
    size: Dp = 48.dp,
    strokeWidth: Dp = 4.dp,
    color: Color = FlyFleetBlue,
)
```

**Usage:** Loading states, pull-to-refresh

---

#### Skeleton Loader

```kotlin
@Composable
fun SkeletonLoader(
    modifier: Modifier = Modifier,
    shape: SkeletonShape,
    shimmer: Boolean = true,
)
```

**Shapes:**
- `Text(width = 200.dp, height = 16.dp)`
- `Title(width = 300.dp, height = 24.dp)`
- `Card(width = 350.dp, height = 120.dp)`
- `Avatar(size = 48.dp)`
- `Thumbnail(width = 100.dp, height = 100.dp)`

**Animation:**
- Shimmer duration: 1000ms
- Direction: Left to right

---

### Dialogs

#### Confirmation Dialog

```kotlin
@Composable
fun ConfirmationDialog(
    title: String,
    message: String,
    confirmText: String,
    dismissText: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
    modifier: Modifier = Modifier,
    type: DialogType = DialogType.Normal,
    icon: ImageVector? = null,
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│                                     │
│            🚨                       │  ← Icon (optional)
│                                     │
│       Cancel Trip?                  │  ← Title (headlineMedium)
│                                     │
│  This action cannot be undone. The  │  ← Message (bodyMedium)
│  trip will be reassigned to another │
│  driver.                            │
│                                     │
│  ┌──────────────┐ ┌──────────────┐ │
│  │  Keep Driving│ │  Cancel Trip │ │  ← Buttons (full width)
│  └──────────────┘ └──────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Types:**
| Type | Icon | Confirm Button |
|------|------|----------------|
| Normal | None | Primary |
| Destructive | Warning | Destructive |
| Info | Info | Primary |
| Success | Check | Primary |

---

#### Bottom Sheet (Action Sheet)

```kotlin
@Composable
fun ActionBottomSheet(
    title: String,
    actions: List<BottomSheetAction>,
    onActionSelected: (BottomSheetAction) -> Unit,
    onDismiss: () -> Unit,
    modifier: Modifier = Modifier,
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│  ─────                              │  ← Handle (drag indicator)
│                                     │
│  Trip Actions                       │  ← Title
│                                     │
│  ┌────────────────────────────────┐ │
│  │ 🗺️  View Route                 │ │  ← Action item
│  └────────────────────────────────┘ │
│  ┌────────────────────────────────┐ │
│  │ 📎  Upload Document            │ │
│  └────────────────────────────────┘ │
│  ┌────────────────────────────────┐ │
│  │ 📞  Contact Dispatcher         │ │
│  └────────────────────────────────┘ │
│  ┌────────────────────────────────┐ │
│  │ 🚨  Report Issue               │ │
│  └────────────────────────────────┘ │
│                                     │
│  [Cancel]                           │  ← Cancel button
└─────────────────────────────────────┘
```

**Specs:**
- Corner Radius: 24dp (top corners only)
- Max Height: 90% of screen
- Padding: 16dp

---

### Navigation

#### Bottom Navigation Bar

```kotlin
@Composable
fun FlyFleetBottomNavigation(
    items: List<BottomNavItem>,
    selectedItemId: String,
    onItemSelected: (String) -> Unit,
    modifier: Modifier = Modifier,
)
```

**Items:**
```kotlin
enum class BottomNavItem(
    val icon: ImageVector,
    val selectedIcon: ImageVector,
    val label: String,
    val route: String,
) {
    Home(
        icon = Icons.Outlined.Home,
        selectedIcon = Icons.Filled.Home,
        label = "Home",
        route = "home",
    ),
    Trips(
        icon = Icons.Outlined.List,
        selectedIcon = Icons.Filled.List,
        label = "Trips",
        route = "trips",
    ),
    Fuel(
        icon = Icons.Outlined.LocalGasStation,
        selectedIcon = Icons.Filled.LocalGasStation,
        label = "Fuel",
        route = "fuel",
    ),
    HOS(
        icon = Icons.Outlined.AccessTime,
        selectedIcon = Icons.Filled.AccessTime,
        label = "HOS",
        route = "hos",
    ),
    Profile(
        icon = Icons.Outlined.Person,
        selectedIcon = Icons.Filled.Person,
        label = "Profile",
        route = "profile",
    ),
}
```

**Layout:**
```
┌─────────────────────────────────────┐
│  [App Content]                      │
│                                     │
├─────────────────────────────────────┤
│  🏠      📋      ⛽      🕐      👤   │  ← Icons
│ Home    Trips    Fuel     HOS  Profile│  ← Labels
└─────────────────────────────────────┘
```

**Specs:**
- Height: 80dp (with label)
- Icon Size: 24dp
- Active: FlyFleetBlue
- Inactive: Gray500
- Badge: ErrorRed (for notifications)

---

#### Top App Bar

```kotlin
@Composable
fun FlyFleetTopAppBar(
    title: String,
    modifier: Modifier = Modifier,
    showBackButton: Boolean = true,
    onBackClick: () -> Unit = { },
    actions: @Composable RowScope.() -> Unit = { },
)
```

**Layout:**
```
┌─────────────────────────────────────┐
│  ←  Trip Details           [⋮]     │  ← Back + Title + Actions
├─────────────────────────────────────┤
│  [Content]                          │
└─────────────────────────────────────┘
```

**Specs:**
- Height: 64dp
- Elevation: 0dp (flat)
- Background: Surface color

---

### Chips

#### Filter Chip

```kotlin
@Composable
fun FilterChip(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    icon: ImageVector? = null,
)
```

**States:**
```
Unselected:
┌──────────────┐
│ ○ All Trips  │  ← Gray border
└──────────────┘

Selected:
┌──────────────┐
│ ● All Trips  │  ← Blue background
└──────────────┘
```

---

#### Status Chip

```kotlin
@Composable
fun StatusChip(
    status: TripStatus,
    modifier: Modifier = Modifier,
)
```

**Status Colors:**
| Status | Background | Text |
|--------|------------|------|
| Active | SuccessGreenLight | SuccessGreen |
| Completed | Gray200 | Gray700 |
| Cancelled | ErrorRedLight | ErrorRed |
| Delayed | WarningAmberLight | WarningAmber |

---

## Screen Templates

### List Screen Template

```kotlin
@Composable
fun ListScreenTemplate(
    title: String,
    items: List<T>,
    onItemSelected: (T) -> Unit,
    emptyState: @Composable () -> Unit,
    loading: Boolean = false,
    floatingActionButton: @Composable () -> Unit = { },
    itemContent: @Composable ColumnScope.(T) -> Unit,
) {
    Scaffold(
        topBar = {
            FlyFleetTopAppBar(title = title)
        },
        floatingActionButton = floatingActionButton,
    ) { padding ->
        when {
            loading -> {
                // Show skeleton loaders
                SkeletonList(count = 5)
            }
            items.isEmpty() -> {
                // Show empty state
                emptyState()
            }
            else -> {
                LazyColumn(
                    contentPadding = padding,
                    verticalArrangement = Arrangement.spacedBy(Spacing.sm),
                ) {
                    items(items) { item ->
                        itemContent(item)
                    }
                }
            }
        }
    }
}
```

---

### Detail Screen Template

```kotlin
@Composable
fun DetailScreenTemplate(
    title: String,
    onBackClick: () -> Unit,
    actions: @Composable RowScope.() -> Unit = { },
    content: @Composable ColumnScope.() -> Unit,
) {
    Scaffold(
        topBar = {
            FlyFleetTopAppBar(
                title = title,
                showBackButton = true,
                onBackClick = onBackClick,
                actions = actions,
            )
        },
    ) { padding ->
        LazyColumn(
            contentPadding = padding,
            verticalArrangement = Arrangement.spacedBy(Spacing.lg),
        ) {
            content()
        }
    }
}
```

---

### Form Screen Template

```kotlin
@Composable
fun FormScreenTemplate(
    title: String,
    onSubmit: () -> Unit,
    submitText: String = "Save",
    submitEnabled: Boolean = true,
    loading: Boolean = false,
    content: @Composable ColumnScope.() -> Unit,
) {
    Scaffold(
        topBar = {
            FlyFleetTopAppBar(title = title, showBackButton = true)
        },
        bottomBar = {
            PrimaryButton(
                text = submitText,
                onClick = onSubmit,
                enabled = submitEnabled && !loading,
                loading = loading,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(Spacing.md),
            )
        },
    ) { padding ->
        LazyColumn(
            contentPadding = padding,
            verticalArrangement = Arrangement.spacedBy(Spacing.lg),
        ) {
            content()
        }
    }
}
```

---

## Empty States

### No Trips

```kotlin
@Composable
fun NoTripsEmptyState(
    onRefreshClick: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
        modifier = Modifier
            .fillMaxSize()
            .padding(Spacing.xxl),
    ) {
        Icon(
            imageVector = Icons.Outlined.EventBusy,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = Gray400,
        )
        
        Spacer(modifier = Modifier.height(Spacing.lg))
        
        Text(
            text = "No trips scheduled",
            style = MaterialTheme.typography.titleLarge,
            color = Gray700,
        )
        
        Spacer(modifier = Modifier.height(Spacing.sm))
        
        Text(
            text = "New trips will appear here once\nassigned by your fleet manager",
            style = MaterialTheme.typography.bodyMedium,
            color = Gray500,
            textAlign = TextAlign.Center,
        )
        
        Spacer(modifier = Modifier.height(Spacing.lg))
        
        SecondaryButton(
            text = "Refresh",
            onClick = onRefreshClick,
        )
    }
}
```

---

### No Network

```kotlin
@Composable
fun NoNetworkState(
    onRetryClick: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
        modifier = Modifier
            .fillMaxSize()
            .padding(Spacing.xxl),
    ) {
        Icon(
            imageVector = Icons.Outlined.SignalWifiOff,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = Gray400,
        )
        
        Spacer(modifier = Modifier.height(Spacing.lg))
        
        Text(
            text = "You're offline",
            style = MaterialTheme.typography.titleLarge,
            color = Gray700,
        )
        
        Spacer(modifier = Modifier.height(Spacing.sm))
        
        Text(
            text = "Most features work offline.\nChanges will sync when connected.",
            style = MaterialTheme.typography.bodyMedium,
            color = Gray500,
            textAlign = TextAlign.Center,
        )
        
        Spacer(modifier = Modifier.height(Spacing.lg))
        
        PrimaryButton(
            text = "Try Again",
            onClick = onRetryClick,
        )
    }
}
```

---

## Error States

### API Error

```kotlin
@Composable
fun ApiErrorBanner(
    message: String,
    onDismiss: () -> Unit,
    onRetry: (() -> Unit)? = null,
) {
    Card(
        colors = CardDefaults.cardColors(
            containerColor = ErrorRedLight,
        ),
        modifier = Modifier
            .fillMaxWidth()
            .padding(Spacing.sm),
    ) {
        Row(
            modifier = Modifier.padding(Spacing.md),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(Spacing.sm),
        ) {
            Icon(
                imageVector = Icons.Rounded.Error,
                contentDescription = null,
                tint = ErrorRed,
            )
            
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = ErrorRed,
                modifier = Modifier.weight(1f),
            )
            
            if (onRetry != null) {
                TextButton(
                    onClick = onRetry,
                    content = { Text("Retry") },
                )
            }
            
            IconButton(
                onClick = onDismiss,
                content = {
                    Icon(
                        imageVector = Icons.Rounded.Close,
                        contentDescription = "Dismiss",
                    )
                },
            )
        }
    }
}
```

---

## Animations

### Shared Element Transitions

```kotlin
// Trip card to details transition
val sharedTransitionSpec = fadeIn(animationSpec = tween(300)) + 
    slideInHorizontally(animationSpec = tween(300))

AnimatedContent(
    targetState = selectedTrip,
    transitionSpec = { sharedTransitionSpec },
) { trip ->
    TripDetails(trip)
}
```

### Progress Bar Animation

```kotlin
val animatedProgress by animateFloatAsState(
    targetValue = progress,
    animationSpec = tween(
        durationMillis = 500,
        easing = FastOutSlowInEasing,
    ),
)

LinearProgressIndicator(
    progress = animatedProgress,
    modifier = Modifier.fillMaxWidth(),
)
```

### Shimmer Animation

```kotlin
@Composable
fun shimmerBrush(): Brush {
    return Brush.linearGradient(
        colors = listOf(
            Gray200,
            Gray100,
            Gray200,
        ),
    )
}

Box(
    modifier = Modifier
        .fillMaxWidth()
        .height(100.dp)
        .background(shimmerBrush())
        .shimmerAnimation(enabled = true),
)
```

---

## Dark Mode

### Implementation

```kotlin
// Dark theme definition
private val DarkColorScheme = darkColorScheme(
    primary = FlyFleetBlue,
    background = DarkBg,
    surface = DarkSurface,
    onPrimary = Color.White,
    onBackground = Color.White,
    onSurface = Color.White,
    // ... other colors
)

@Composable
fun FlyFleetTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    
    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography,
        content = content,
    )
}
```

### Dark Mode Adjustments

| Element | Light | Dark |
|---------|-------|------|
| Background | Gray50 | DarkBg (#0F172A) |
| Card Surface | White | DarkSurface (#1E293B) |
| Elevated Surface | White | DarkSurfaceElevated (#334155) |
| Primary Text | Gray900 | Gray50 |
| Secondary Text | Gray500 | Gray400 |
| Borders | Gray200 | Gray700 |
| Skeleton Base | Gray200 | Gray700 |
| Skeleton Highlight | Gray100 | Gray600 |

---

## Accessibility

### Content Descriptions

```kotlin
// Always provide content descriptions for icons
Icon(
    imageVector = Icons.Rounded.Home,
    contentDescription = "Home",  // Required for TalkBack
)

// For images with meaning
AsyncImage(
    model = vehicleImageUrl,
    contentDescription = "Vehicle: TATA Ace",  // Describe what the image shows
)
```

### Touch Target Size

```kotlin
// Minimum 48dp touch target
IconButton(
    onClick = { },
    modifier = Modifier.size(48.dp),  // Explicit minimum
) {
    Icon(
        imageVector = Icons.Rounded.Menu,
        contentDescription = "Menu",
        modifier = Modifier.size(24.dp),  // Icon can be smaller
    )
}
```

### Font Scaling

```kotlin
// Respect system font scale
val fontScale = LocalDensity.current.fontScale
Text(
    text = "Hello",
    fontSize = (16.sp * fontScale),  // Scale with system
)
```

---

## File Structure

```
apps/mobile/driver/android-driver/app/src/main/
├── kotlin/com/flyfleet/driver/
│   └── ui/
│       ├── theme/
│       │   ├── Color.kt          # Color definitions
│       │   ├── Type.kt           # Typography
│       │   ├── Shape.kt          # Corner radius, shapes
│       │   ├── Spacing.kt        # Spacing system
│       │   └── Theme.kt          # Theme composition
│       ├── components/
│       │   ├── buttons/
│       │   │   ├── PrimaryButton.kt
│       │   │   ├── SecondaryButton.kt
│       │   │   └── IconButton.kt
│       │   ├── cards/
│       │   │   ├── TripCard.kt
│       │   │   ├── AlertCard.kt
│       │   │   └── DocumentCard.kt
│       │   ├── inputs/
│       │   │   ├── TextField.kt
│       │   │   ├── NumberField.kt
│       │   │   └── PhotoUpload.kt
│       │   ├── progress/
│       │   │   ├── LinearProgress.kt
│       │   │   ├── CircularProgress.kt
│       │   │   └── SkeletonLoader.kt
│       │   ├── dialogs/
│       │   │   ├── ConfirmationDialog.kt
│       │   │   └── ActionBottomSheet.kt
│       │   └── navigation/
│       │       ├── BottomNavigation.kt
│       │       └── TopAppBar.kt
│       └── screens/
│           ├── login/
│           ├── home/
│           ├── trips/
│           ├── fuel/
│           ├── hos/
│           ├── profile/
│           └── emergency/
└── res/
    ├── values/
    │   ├── colors.xml
    │   ├── strings.xml
    │   └── themes.xml
    ├── font/
    │   ├── inter_regular.ttf
    │   ├── inter_medium.ttf
    │   ├── inter_semibold.ttf
    │   └── inter_bold.ttf
    └── drawable/
        ├── ic_logo.xml
        └── ...
```

---

## Design Handoff Checklist

- [ ] All colors defined in `Color.kt`
- [ ] Typography scale in `Type.kt`
- [ ] Spacing system in `Spacing.kt`
- [ ] All components implemented with variants
- [ ] Light and dark theme support
- [ ] Empty states for all list screens
- [ ] Error states for all async operations
- [ ] Loading states (skeleton preferred)
- [ ] Accessibility (TalkBack, font scaling)
- [ ] Animations documented
- [ ] Responsive layouts (phone, tablet)
- [ ] Export to Figma for design team

---

## Resources

- **Figma Design File:** [Link to Figma]
- **Icon Library:** Material Icons Rounded
- **Font:** Inter (Google Fonts)
- **Lottie Animations:** [Link to Lottie files]

---

**Last Updated:** March 31, 2026  
**Maintained By:** Design Team

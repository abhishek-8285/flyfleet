# FlyFleet Driver App - Figma Mockup Specifications

**Version:** 1.0  
**Date:** March 31, 2026  
**Design Tool:** Figma  
**Device Frame:** Android Phone (360x800 - standard MDPI)  
**Design System:** FlyFleet Mobile (Razorpay-inspired)

---

## 📐 Figma Setup

### Frame Configuration
```
Device: Android Large (360x800)
Grid: 4px baseline grid (8px primary)
Layout Grid: 
  - Columns: 4
  - Gutter: 16px
  - Margin: 16px
  - Type: Stretch
```

### Color Styles (Create as Figma Color Styles)

#### Primary Colors
| Style Name | Value | Usage |
|------------|-------|-------|
| `Primary/FlyFleetBlue` | #2563EB | Primary buttons, links, active states |
| `Primary/FlyFleetBlueDark` | #1D4ED8 | Hover, pressed states |
| `Primary/FlyFleetBlueLight` | #60A5FA | Disabled, backgrounds |
| `Primary/FlyFleetBlueSubtle` | #DBEAFE | Subtle backgrounds |

#### Semantic Colors
| Style Name | Value | Usage |
|------------|-------|-------|
| `Semantic/Success` | #10B981 | Completed states, safe |
| `Semantic/SuccessLight` | #D1FAE5 | Success backgrounds |
| `Semantic/Warning` | #F59E0B | Warnings, expiring |
| `Semantic/WarningLight` | #FEF3C7 | Warning backgrounds |
| `Semantic/Error` | #EF4444 | Errors, urgent, SOS |
| `Semantic/ErrorLight` | #FEE2E2 | Error backgrounds |
| `Semantic/Info` | #3B82F6 | Informational |
| `Semantic/InfoLight` | #DBEAFE | Info backgrounds |

#### Neutral Colors
| Style Name | Value | Usage |
|------------|-------|-------|
| `Neutral/Gray50` | #F9FAFB | App background |
| `Neutral/Gray100` | #F3F4F6 | Card backgrounds |
| `Neutral/Gray200` | #E5E7EB | Borders, dividers |
| `Neutral/Gray300` | #D1D5DB | Disabled borders |
| `Neutral/Gray400` | #9CA3AF | Disabled text |
| `Neutral/Gray500` | #6B7280 | Secondary text |
| `Neutral/Gray600` | #4B5563 | Body text |
| `Neutral/Gray700` | #374151 | Primary text |
| `Neutral/Gray800` | #1F2937 | Headings |
| `Neutral/Gray900` | #111827 | Headings (dark) |

#### Dark Mode Colors
| Style Name | Value | Usage |
|------------|-------|-------|
| `Dark/Bg` | #0F172A | Dark mode background |
| `Dark/Surface` | #1E293B | Dark mode cards |
| `Dark/SurfaceElevated` | #334155 | Dark mode elevated |

---

## 🔤 Text Styles (Create as Figma Text Styles)

### Display
| Style Name | Font | Size | Weight | Line Height | Letter Spacing |
|------------|------|------|--------|-------------|----------------|
| `Display/Large` | Inter | 32px | Bold | 40px | -0.5px |
| `Display/Medium` | Inter | 28px | Bold | 36px | -0.25px |

### Headline
| Style Name | Font | Size | Weight | Line Height | Letter Spacing |
|------------|------|------|--------|-------------|----------------|
| `Headline/Large` | Inter | 24px | SemiBold | 32px | 0 |
| `Headline/Medium` | Inter | 20px | SemiBold | 28px | 0 |

### Title
| Style Name | Font | Size | Weight | Line Height | Letter Spacing |
|------------|------|------|--------|-------------|----------------|
| `Title/Large` | Inter | 18px | Medium | 24px | 0.15px |
| `Title/Medium` | Inter | 16px | Medium | 22px | 0.15px |
| `Title/Small` | Inter | 14px | Medium | 20px | 0.1px |

### Body
| Style Name | Font | Size | Weight | Line Height | Letter Spacing |
|------------|------|------|--------|-------------|----------------|
| `Body/Large` | Inter | 16px | Regular | 24px | 0.5px |
| `Body/Medium` | Inter | 14px | Regular | 20px | 0.25px |
| `Body/Small` | Inter | 12px | Regular | 16px | 0.4px |

### Label
| Style Name | Font | Size | Weight | Line Height | Letter Spacing |
|------------|------|------|--------|-------------|----------------|
| `Label/Large` | Inter | 14px | Medium | 20px | 0.1px |
| `Label/Medium` | Inter | 12px | Medium | 16px | 0.5px |
| `Label/Small` | Inter | 11px | Medium | 16px | 0.5px |

---

## 🎨 Effect Styles

### Shadows
| Style Name | Value | Usage |
|------------|-------|-------|
| `Shadow/Sm` | 0 1px 2px rgba(0,0,0,0.05) | Small cards, chips |
| `Shadow/Md` | 0 4px 6px rgba(0,0,0,0.1) | Cards, buttons |
| `Shadow/Lg` | 0 10px 15px rgba(0,0,0,0.1) | Modals, dialogs |
| `Shadow/Xl` | 0 20px 25px rgba(0,0,0,0.15) | Bottom sheets |
| `Shadow/Inner` | inset 0 2px 4px rgba(0,0,0,0.06) | Input fields |

### Blur
| Style Name | Value | Usage |
|------------|-------|-------|
| `Blur/Sm` | 4px | Subtle backgrounds |
| `Blur/Md` | 8px | Backdrop blur |
| `Blur/Lg` | 16px | Heavy blur |

---

## 📱 Screen Specifications

### Screen 1: Login Screen

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│                                     │
│           [Logo]                    │  ← 120x120px, center
│           FlyFleet                  │  ← Headline/Large, Gray900
│                                     │
│  ┌─────────────────────────────┐   │  ← Card, White, 16px radius
│  │                             │   │
│  │  Welcome back! 👋            │   │  ← Headline/Medium, Gray900
│  │  Sign in to continue         │   │  ← Body/Medium, Gray500
│  │                             │   │
│  │  ┌─────────────────────┐   │   │  ← Input field
│  │  │ 📧 Email            │   │   │
│  │  │ driver@example.com  │   │   │
│  │  └─────────────────────┘   │   │
│  │                             │   │
│  │  ┌─────────────────────┐   │   │  ← Input field
│  │  │ 🔒 Password         │   │   │
│  │  │ ••••••••            │   │   │
│  │  └─────────────────────┘   │   │
│  │                             │   │
│  │  ☐ Remember this device     │   │  ← Checkbox + Label/Medium
│  │                             │   │
│  │  ┌─────────────────────┐   │   │  ← Primary Button
│  │  │     Sign In         │   │   │
│  │  └─────────────────────┘   │   │
│  │                             │   │
│  │  ──────── Or ────────      │   │  ← Divider + Label/Medium
│  │                             │   │
│  │  ┌─────────────────────┐   │   │  ← Secondary Button
│  │  │   Send OTP          │   │   │
│  │  └─────────────────────┘   │   │
│  │                             │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │  ← Bottom card
│  │ Need help? Contact support  │   │  ← Body/Small, Primary
│  └─────────────────────────────┘   │
│                                     │
└─────────────────────────────────────┘
```

#### Component Specifications

**Logo:**
- Size: 120x120px
- Position: Center, 80px from top
- Asset: `logo-full.svg` (provided by brand team)

**Input Card:**
- Position: 240px from top
- Padding: 24px
- Background: White
- Corner Radius: 16px
- Shadow: `Shadow/Md`

**Input Fields:**
- Height: 56px
- Background: `Neutral/Gray50`
- Border: 1px `Neutral/Gray200`
- Corner Radius: 12px
- Padding: 16px
- Label: `Label/Medium`, `Neutral/Gray500` (above field)
- Text: `Body/Large`, `Neutral/Gray900`

**Primary Button:**
- Height: 56px
- Background: `Primary/FlyFleetBlue`
- Corner Radius: 12px
- Text: `Label/Large`, White
- Full width

**Secondary Button:**
- Height: 56px
- Background: Transparent
- Border: 2px `Primary/FlyFleetBlue`
- Corner Radius: 12px
- Text: `Label/Large`, `Primary/FlyFleetBlue`
- Full width

**Divider:**
- Line: 1px `Neutral/Gray200`
- Text: `Label/Medium`, `Neutral/Gray400`
- Padding: 8px

---

### Screen 2: Home (Trip Dashboard)

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ [☰]                    [🔔] [👤]   │  ← Top bar, 64px
│                                     │
│  Good Morning, Rajesh! 👋           │  ← Headline/Medium, 24px top
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🚛 ACTIVE TRIP                │ │  ← Card, SuccessGreen Light bg
│  │                               │ │
│  │  Delhi → Mumbai               │ │  ← Title/Large, Gray900
│  │  45% complete                 │ │  ← Body/Medium, SuccessGreen
│  │                               │ │
│  │  ████████░░░░░░░░ 650/1450km  │ │  ← Progress bar
│  │                               │ │
│  │  ⏱️ 6h 30m remaining          │ │  ← Body/Small, Gray600
│  │  📍 Near Jaipur               │ │
│  │                               │ │
│  │  [View Details] [Navigate]    │ │  ← Two buttons
│  └───────────────────────────────┘ │
│                                     │
│  QUICK ACTIONS                      │  ← Title/Small, Gray700
│  ┌─────────┐ ┌─────────┐ ┌───────┐ │
│  │   📋    │ │   ⛽    │ │  🆘   │ │  ← Icon buttons, 80x80
│  │  Trips  │ │  Fuel   │ │  SOS  │ │
│  └─────────┘ └─────────┘ └───────┘ │
│                                     │
│  UPCOMING TRIPS                     │  ← Title/Small, Gray700
│  ┌───────────────────────────────┐ │
│  │ Tomorrow, 6:00 AM             │ │
│  │ Mumbai → Pune | TATA-456      │ │
│  │ [Accept] [Decline]            │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ Mar 2, 2:00 PM                │ │
│  │ Pune → Bangalore | TATA-456   │ │
│  │ [View Details]                │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

#### Active Trip Card Specifications

**Card:**
- Position: 24px from greeting
- Padding: 20px
- Background: `Semantic/SuccessLight`
- Corner Radius: 16px
- Border: 1px `Semantic/Success`

**Progress Bar:**
- Height: 8px
- Track: White (50% opacity)
- Fill: `Semantic/Success`
- Corner Radius: 4px
- Margin: 12px vertical

**Action Buttons (Inline):**
- Height: 40px
- Padding: 16px horizontal
- Background: `Semantic/Success`
- Corner Radius: 8px
- Text: `Label/Large`, White
- Spacing: 12px between

---

### Screen 3: Trip Details

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ ← Back              [Complete Stop] │  ← Top bar
│                                     │
│  ┌───────────────────────────────┐ │
│  │        [MAP VIEW]             │ │  ← 200px height
│  │     🚛 → → → 📍              │ │
│  └───────────────────────────────┘ │
│                                     │
│  STOP 2 OF 4                        │  ← Title/Medium, center
│  ┌───────────────────────────────┐ │
│  │ 📍 Mumbai Warehouse           │ │  ← Card
│  │                               │ │
│  │  Address:                     │ │
│  │  Andheri East, Mumbai         │ │
│  │                               │ │
│  │  Check-in: 10:45 AM ✅        │ │
│  │  Status: Checked In           │ │
│  │                               │ │
│  │  Contact: +91 98765 43210     │ │
│  └───────────────────────────────┘ │
│                                     │
│  NEXT STOP                          │  ← Label/Medium, Gray500
│  ┌───────────────────────────────┐ │
│  │ 📍 Pune Distribution Center   │ │
│  │ 120 km | 2h 15m               │ │
│  │ [Start Navigation]            │ │
│  └───────────────────────────────┘ │
│                                     │
│  ACTIONS                            │  ← Title/Small
│  ┌──────────┐ ┌──────────┐         │
│  │ 📸 Upload│ │ 📝 Notes  │         │  ← 160x80 cards
│  │  POD     │ │          │         │
│  └──────────┘ └──────────┘         │
│  ┌──────────┐ ┌──────────┐         │
│  │ ⏱️ Delay │ │ 📞 Contact│         │
│  │ Report   │ │ Dispatcher│         │
│  └──────────┘ └──────────┘         │
│                                     │
└─────────────────────────────────────┘
```

#### Map Card Specifications

**Map Container:**
- Height: 200px
- Corner Radius: 16px
- Overflow: Hidden
- Shadow: `Shadow/Md`

**Map Placeholder:**
- Background: `Neutral/Gray200`
- Route line: `Primary/FlyFleetBlue`, 4px stroke
- Start marker: Green pin icon
- End marker: Red pin icon
- Vehicle: Truck icon on path

---

### Screen 4: Fuel Log

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ ← Back                    [Save]    │  ← Top bar
│                                     │
│  LOG FUEL                           │  ← Headline/Medium
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ⛽ Station Name               │ │
│  │ Indian Oil, NH-48             │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌──────────┐ ┌──────────┐         │
│  │ Liters   │ │ Price    │         │
│  │ 45.5 L   │ │ ₹4,250   │         │
│  └──────────┘ └──────────┘         │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📍 Location                   │ │
│  │ Mumbai-Pune Highway           │ │
│  │ ✅ Captured automatically     │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📸 Receipt Photo              │ │
│  │                               │ │
│  │    ┌─────┐  ┌─────┐          │ │
│  │    │ 📷  │  │     │          │ │
│  │    │ Add │  │     │          │ │
│  │    └─────┘  └─────┘          │ │
│  │                               │ │
│  │  [Take Photo] [Upload]        │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🔢 Odometer                   │ │
│  │ 45,230 km                     │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 💳 Payment Method             │ │
│  │ ○ Fuel Card  ○ Cash          │ │
│  └───────────────────────────────┘ │
│                                     │
│         [Save Fuel Log]             │  ← Primary Button, fixed bottom
└─────────────────────────────────────┘
```

#### Input Grid Specifications

**Two-Column Grid:**
- Gap: 16px
- Each column: calc(50% - 8px)
- Input height: 64px (taller for numbers)

**Photo Upload Card:**
- Padding: 24px
- Background: White
- Corner Radius: 16px
- Border: 2px dashed `Neutral/Gray300`

**Photo Thumbnails:**
- Size: 80x80px
- Corner Radius: 8px
- Border: 1px `Neutral/Gray200`
- Remove button: 24x24px circle, top-right

---

### Screen 5: HOS Status

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ ← Back                             │
│                                     │
│  HOURS OF SERVICE                   │
│  Today, March 31                    │
│                                     │
│  STATUS: ✅ DRIVING                 │  ← Badge, center
│  Last break: 2h 15m ago             │
│  Next break due: 2h 45m             │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ⏱️ DRIVING TIME               │ │
│  │                               │ │
│  │  ████████░░░░░░░░             │ │  ← 6h 15m / 9h
│  │  6h 15m / 9h                  │ │
│  │  2h 45m remaining             │ │
│  │                               │ │
│  │  ⚠️ Break needed in 2h 45m    │ │  ← Warning row
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🕐 ON-DUTY TIME               │ │
│  │                               │ │
│  │  ██████░░░░░░░░░░             │ │  ← 8h 30m / 14h
│  │  8h 30m / 14h                 │ │
│  │  5h 30m remaining             │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🛌 OFF-DUTY (Last 7 Days)     │ │
│  │                               │ │
│  │  ████████████████░░           │ │  ← 58h / 70h
│  │  58h / 70h                    │ │
│  │  12h remaining                │ │
│  └───────────────────────────────┘ │
│                                     │
│  DUTY STATUS                        │  ← Title/Small
│  ┌────────┐ ┌────────┐ ┌────────┐ │
│  │Driving │ │On-Duty │ │ Sleeper│ │  ← Toggle buttons
│  │   ●    │ │        │ │        │ │
│  └────────┘ └────────┘ └────────┘ │
│  ┌────────┐                        │
│  │Off-Duty│                        │
│  │        │                        │
│  └────────┘                        │
│                                     │
└─────────────────────────────────────┘
```

#### Progress Card Specifications

**Card:**
- Padding: 20px
- Background: White
- Corner Radius: 16px
- Shadow: `Shadow/Sm`
- Margin: 12px between cards

**Progress Bar Colors:**
| Status | Color |
|--------|-------|
| <50% | `Semantic/Success` |
| 50-75% | `Semantic/Warning` |
| >75% | `Semantic/Error` |

**Warning Row:**
- Background: `Semantic/WarningLight`
- Padding: 12px
- Corner Radius: 8px
- Icon: ⚠️ `Semantic/Warning`
- Text: `Body/Small`, `Semantic/Warning`

---

### Screen 6: SOS Emergency (Overlay)

**Frame:** 360x800  
**Background:** Overlay with blur

#### Layout Structure
```
┌─────────────────────────────────────┐
│                                     │
│         🚨 SOS ACTIVATED 🚨         │  ← Headline/Large, Error
│                                     │
│  Help is on the way!                │  ← Body/Large
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📍 Location sent to:          │ │
│  │ • Fleet Control Room          │ │
│  │ • Emergency Contact: Wife     │ │
│  │ • Nearest Hospital: 12 km     │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Your Location:                │ │
│  │ NH-48, Near Pune Bypass       │ │
│  │ Accuracy: 15 meters           │ │
│  │                               │ │
│  │     [MAP PREVIEW]             │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Fleet manager is calling...   │ │
│  │                               │ │
│  │      [Answer]  [Decline]      │ │
│  └───────────────────────────────┘ │
│                                     │
│  [Cancel False Alarm]               │  ← Secondary Button
│                                     │
└─────────────────────────────────────┘
```

#### Overlay Specifications

**Background:**
- Fill: Black at 60% opacity
- Backdrop blur: 8px
- Animation: Fade in 300ms

**Content Card:**
- Position: Center vertically
- Padding: 32px
- Background: White
- Corner Radius: 24px
- Shadow: `Shadow/Xl`
- Max width: 320px

**SOS Header:**
- Icon: 64x64px
- Text: `Headline/Large`, `Semantic/Error`
- Animation: Pulse (scale 1.0 → 1.05, 1s infinite)

**Call Buttons:**
- Size: 64px circles
- Answer: `Semantic/Success` background
- Decline: `Semantic/Error` background
- Icons: 32x32px, White

---

### Screen 7: Alert Inbox

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ ← Back        Alerts        [⋮]    │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📊 3 Unread | 12 Total        │ │  ← Summary card
│  └───────────────────────────────┘ │
│                                     │
│  TODAY                              │  ← Label/Medium, Gray500
│  ┌───────────────────────────────┐ │
│  │ 🚨 CRITICAL                   │ │  ← Red badge
│  │                               │ │
│  │  Harsh Brake Detected         │ │  ← Title/Medium
│  │  Today, 10:45 AM              │ │
│  │  NH-48, Km 245                │ │
│  │                               │ │
│  │  Speed dropped from 80 to     │ │
│  │  30 km/h in 3 seconds.        │ │
│  │                               │ │
│  │  [Acknowledge] [Dismiss]      │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ ⚠️ WARNING                    │ │  ← Amber badge
│  │                               │ │
│  │  Maintenance Due              │ │
│  │  Today, 8:00 AM               │ │
│  │  TATA-456                     │ │
│  │                               │ │
│  │  Service interval reached.    │ │
│  │  Schedule within 7 days.      │ │
│  │                               │ │
│  │  [Schedule] [Dismiss]         │ │
│  └───────────────────────────────┘ │
│                                     │
│  YESTERDAY                          │
│  ┌───────────────────────────────┐ │
│  │ ℹ️ INFO                       │ │  ← Blue badge
│  │  Trip Completed               │ │
│  │  Yesterday, 6:30 PM           │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

#### Alert Card Specifications

**Card by Priority:**

| Priority | Background | Border | Icon |
|----------|------------|--------|------|
| Critical | `Semantic/ErrorLight` | 2px `Semantic/Error` | 🚨 |
| Warning | `Semantic/WarningLight` | 2px `Semantic/Warning` | ⚠️ |
| Info | `Semantic/InfoLight` | 2px `Semantic/Info` | ℹ️ |

**Card Layout:**
- Padding: 16px
- Corner Radius: 12px
- Margin: 12px vertical

**Badge:**
- Height: 24px
- Padding: 4px 12px
- Corner Radius: 4px
- Text: `Label/Small`, uppercase

---

### Screen 8: Profile & Settings

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

#### Layout Structure
```
┌─────────────────────────────────────┐
│ ← Back        Profile        [Edit] │
│                                     │
│  ┌───────────────────────────────┐ │
│  │       [Photo: 80px]           │ │
│  │                               │ │
│  │      Rajesh Kumar             │ │  ← Headline/Medium, center
│  │      Driver ID: DL-4521       │ │  ← Body/Small, Gray500
│  │                               │ │
│  │      [📞] [✉️] [📄]          │ │  ← Action icons
│  └───────────────────────────────┘ │
│                                     │
│  ACCOUNT                            │  ← Label/Medium, Gray500
│  ┌───────────────────────────────┐ │
│  │ 👤 Personal Information   →   │ │  ← List item
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🚗 Assigned Vehicle       →   │ │
│  │    TATA Ace | MH-12-AB-1234   │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📄 My Documents           →   │ │
│  │    3 expiring soon ⚠️         │ │
│  └───────────────────────────────┘ │
│                                     │
│  PREFERENCES                        │
│  ┌───────────────────────────────┐ │
│  │ 🔔 Notifications          →   │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🌐 Language               →   │ │
│  │    English                    │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🌙 Dark Mode              ⊂⊃  │ │  ← Toggle switch
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔒 Privacy                →   │ │
│  └───────────────────────────────┘ │
│                                     │
│  SUPPORT                            │
│  ┌───────────────────────────────┐ │
│  │ ❓ Help Center            →   │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📞 Contact Support        →   │ │
│  └───────────────────────────────┘ │
│                                     │
│         [Sign Out]                  │  ← Destructive Button
│                                     │
└─────────────────────────────────────┘
```

#### Profile Card Specifications

**Card:**
- Padding: 24px
- Background: White
- Corner Radius: 16px
- Shadow: `Shadow/Md`
- Text alignment: Center

**Photo:**
- Size: 80x80px
- Corner Radius: 40px (circle)
- Border: 3px `Primary/FlyFleetBlue`
- Placeholder: `Neutral/Gray200` with person icon

**List Items:**
- Height: 64px (or auto with subtitle)
- Padding: 16px
- Icon: 24x24px, left
- Chevron: 20x20px, `Neutral/Gray400`, right
- Border: 1px `Neutral/Gray100` (between items)

---

## 🧩 Component Library Specifications

### Buttons

#### Primary Button
```
Frame: Auto width, 56px height (Large)
       Auto width, 48px height (Medium)
       Auto width, 40px height (Small)

Fill: Primary/FlyFleetBlue
Corner Radius: 12px
Padding: 16px horizontal, 16px vertical

Text: Label/Large, White

States:
- Default: Fill #2563EB
- Hover: Fill #1D4ED8
- Pressed: Fill #1D4ED8, Scale 98%
- Disabled: Fill #E5E7EB, Text #9CA3AF
```

#### Secondary Button
```
Frame: Auto width, 56px height

Fill: Transparent
Stroke: 2px Primary/FlyFleetBlue
Corner Radius: 12px
Padding: 16px horizontal, 16px vertical

Text: Label/Large, Primary/FlyFleetBlue

States:
- Default: Stroke #2563EB
- Hover: Fill #DBEAFE (10%)
- Pressed: Fill #DBEAFE (20%), Scale 98%
- Disabled: Stroke #E5E7EB, Text #9CA3AF
```

#### Destructive Button
```
Frame: Auto width, 56px height

Fill: Semantic/Error
Corner Radius: 12px
Padding: 16px horizontal, 16px vertical

Text: Label/Large, White
```

#### Icon Button
```
Frame: 48x48px (Medium)
       56x56px (Large)

Fill: Transparent (default)
      Primary/FlyFleetBlue (filled variant)

Icon: 24x24px, centered
Corner Radius: 12px

States:
- Default: Transparent
- Hover: Fill #F3F4F6 (10%)
- Pressed: Fill #E5E7EB (20%), Scale 95%
```

---

### Input Fields

#### Text Field
```
Frame: Full width, 56px height

Fill: Neutral/Gray50
Stroke: 1px Neutral/Gray200 (default)
        2px Primary/FlyFleetBlue (focused)
        2px Semantic/Error (error)
Corner Radius: 12px
Padding: 16px

Label: Label/Medium, Neutral/Gray500 (above field)
Text: Body/Large, Neutral/Gray900
Placeholder: Body/Large, Neutral/Gray400

States:
- Default: Stroke #E5E7EB
- Focused: Stroke #2563EB
- Error: Stroke #EF4444, Error text below
- Disabled: Fill #F9FAFB, Text #9CA3AF
```

#### Number Field
```
Same as Text Field with:
- Keyboard type: Numeric
- Suffix/Prefix: Label/Medium, Neutral/Gray500 (right/left)
```

---

### Cards

#### Trip Card
```
Frame: Full width, Auto height (min 120px)

Fill: White
Stroke: 1px Neutral/Gray200
Corner Radius: 16px
Padding: 20px
Shadow: Shadow/Sm

Content:
- Title: Title/Large, Neutral/Gray900
- Subtitle: Body/Medium, Neutral/Gray500
- Progress: 8px height, 4px radius
- Actions: Button row, 12px gap
```

#### Alert Card
```
Frame: Full width, Auto height

Fill: Varies by priority (see Alert Inbox)
Stroke: 2px Semantic color
Corner Radius: 12px
Padding: 16px

Content:
- Badge: 24px height, Label/Small
- Title: Title/Medium
- Subtitle: Body/Small, Neutral/Gray500
- Description: Body/Medium
- Actions: Button row
```

---

### Progress Indicators

#### Linear Progress
```
Frame: Full width, 8px height

Track:
- Fill: Neutral/Gray200
- Corner Radius: 4px

Fill:
- Height: 8px
- Corner Radius: 4px
- Color: Varies by progress
  - 0-50%: Semantic/Success
  - 50-75%: Semantic/Warning
  - 75-100%: Semantic/Error

Label: Body/Small, Neutral/Gray600 (above or below)
```

#### Circular Progress (Loading)
```
Frame: 48x48px (default)

Stroke: 4px
Track: Neutral/Gray200
Fill: Primary/FlyFleetBlue

Animation: Rotate 360°, 1s infinite
```

---

### Chips

#### Filter Chip
```
Frame: Auto width, 32px height

Fill: Transparent (unselected)
      Primary/FlyFleetBlue (selected)
Stroke: 1px Neutral/Gray300 (unselected)
Corner Radius: 16px
Padding: 8px 16px

Text: Label/Medium
      Neutral/Gray600 (unselected)
      White (selected)
```

#### Status Chip
```
Frame: Auto width, 24px height

Fill: Varies by status (see Status Chip section)
Corner Radius: 4px
Padding: 4px 8px

Text: Label/Small
```

---

## 📎 Asset Requirements

### Icons (24x24px, SVG format)

| Icon Name | Usage | Style |
|-----------|-------|-------|
| `ic-home` | Bottom nav | Outline + Filled |
| `ic-trips` | Bottom nav | Outline + Filled |
| `ic-fuel` | Bottom nav | Outline + Filled |
| `ic-hos` | Bottom nav | Outline + Filled |
| `ic-profile` | Bottom nav | Outline + Filled |
| `ic-back` | Top bar | Outline |
| `ic-menu` | Top bar | Outline |
| `ic-notifications` | Top bar | Outline + Badge |
| `ic-edit` | Edit actions | Outline |
| `ic-camera` | Photo upload | Outline |
| `ic-upload` | Photo upload | Outline |
| `ic-location` | Location fields | Outline |
| `ic-phone` | Contact actions | Outline |
| `ic-email` | Contact actions | Outline |
| `ic-document` | Documents | Outline |
| `ic-check` | Success states | Outline |
| `ic-close` | Dismiss actions | Outline |
| `ic-error` | Error states | Outline |
| `ic-warning` | Warning states | Filled |
| `ic-info` | Info states | Outline |
| `ic-chevron-right` | Navigation | Outline |
| `ic-chevron-down` | Expand | Outline |
| `ic-more-vertical` | More actions | Outline |
| `ic-sos` | Emergency | Filled |
| `ic-vehicle` | Vehicle info | Outline |
| `ic-driver` | Driver info | Outline |
| `ic-navigation` | Navigate action | Outline |
| `ic-clock` | Time-related | Outline |
| `ic-calendar` | Date-related | Outline |

### Illustrations

| Illustration | Size | Usage |
|--------------|------|-------|
| `illustration-empty-trips` | 200x200px | Empty state |
| `illustration-no-network` | 200x200px | Offline state |
| `illustration-success` | 120x120px | Success confirmation |
| `illustration-error` | 120x120px | Error state |
| `illustration-onboarding-1` | 280x280px | Onboarding screen 1 |
| `illustration-onboarding-2` | 280x280px | Onboarding screen 2 |
| `illustration-onboarding-3` | 280x280px | Onboarding screen 3 |
| `illustration-onboarding-4` | 280x280px | Onboarding screen 4 |

### Lottie Animations

| Animation | Duration | Usage |
|-----------|----------|-------|
| `anim-sos-pulse` | 1s loop | SOS overlay |
| `anim-sync-spinner` | 1.5s loop | Sync indicator |
| `anim-success-check` | 800ms | Success confirmation |
| `anim-loading-dots` | 1s loop | Loading states |
| `anim-confetti` | 2s one-shot | Trip completion |

---

## 🎬 Micro-interaction Specifications

### Button Press
```
Duration: 150ms
Easing: ease-out
Effect: Scale 98%
Shadow: Reduce elevation on press
```

### Card Tap
```
Duration: 200ms
Easing: ease-out
Effect: Ripple from tap point
Ripple color: Primary/FlyFleetBlue at 20%
```

### Input Focus
```
Duration: 200ms
Easing: ease-out
Effect: Border color change + slight scale (101%)
Label: Color change to Primary
```

### List Item Swipe
```
Direction: Left or Right
Reveal: Action buttons underneath
Threshold: 80px for commit
Snap: Animate to open/closed
```

### Bottom Sheet Slide
```
Duration: 300ms
Easing: cubic-bezier(0.4, 0.0, 0.2, 1)
Effect: Slide up + Fade in
Backdrop: Fade in 200ms
```

---

## 📱 Responsive Breakpoints

| Breakpoint | Width | Layout Changes |
|------------|-------|----------------|
| Compact | 320-359px | Reduce padding to 12px, stack 2-col grids |
| Default | 360-399px | Standard layout |
| Expanded | 400px+ | Increase card max-width, wider grids |

---

## ♿ Accessibility Checklist

For each screen, verify:
- [ ] Color contrast ratio ≥ 4.5:1 for text
- [ ] Touch targets ≥ 48x48dp
- [ ] All icons have content descriptions
- [ ] Focus order is logical
- [ ] Error messages are descriptive
- [ ] Success states are announced

---

## 📤 Handoff Checklist

Before handing off to development:
- [ ] All color styles created
- [ ] All text styles created
- [ ] All component variants created
- [ ] All screens designed
- [ ] Assets exported (SVG for icons, PNG for illustrations)
- [ ] Prototype links created
- [ ] Red notes for animations
- [ ] Design specs documented (this file)

---

**Created:** March 31, 2026  
**Designer:** [Your Name]  
**Status:** Ready for Development

# FlyFleet Driver App - Additional Screen Specifications

**Continuation of:** [`FIGMA_MOCKUP_SPECS.md`](./FIGMA_MOCKUP_SPECS.md)  
**Date:** March 31, 2026

---

## 📱 Additional Screens (9-18)

### Screen 9: Onboarding Carousel (4 Screens)

**Frame:** 360x800  
**Background:** `Primary/FlyFleetBlue` (gradient to `Primary/FlyFleetBlueDark`)

#### Screen 1: Welcome
```
┌─────────────────────────────────────┐
│                                     │
│        [Illustration: 280px]        │  ← Truck + Driver
│                                     │
│  Welcome to FlyFleet                │  ← Headline/Large, White
│                                     │
│  Your complete fleet management     │  ← Body/Large, White (80%)
│  companion for seamless trips       │
│                                     │
│         ○ ● ● ●                     │  ← Page indicators
│                                     │
│         [Get Started]               │  ← Primary Button (White)
│                                     │
└─────────────────────────────────────┘
```

#### Screen 2: Trips
```
┌─────────────────────────────────────┐
│                                     │
│        [Illustration: 280px]        │  ← Map with route
│                                     │
│  Manage Your Trips                  │
│                                     │
│  View, plan, and complete trips     │
│  with real-time tracking            │
│                                     │
│         ● ○ ● ●                     │
│                                     │
│         [Continue]                  │
│                                     │
└─────────────────────────────────────┘
```

#### Screen 3: Compliance
```
┌─────────────────────────────────────┐
│                                     │
│        [Illustration: 280px]        │  ← Documents + Check
│                                     │
│  Stay Compliant                     │
│                                     │
│  Automatic HOS tracking and         │
│  document management                │
│                                     │
│         ● ● ○ ●                     │
│                                     │
│         [Continue]                  │
│                                     │
└─────────────────────────────────────┘
```

#### Screen 4: Safety
```
┌─────────────────────────────────────┐
│                                     │
│        [Illustration: 280px]        │  ← SOS + Shield
│                                     │
│  Emergency Support                  │
│                                     │
│  One-tap SOS and 24/7               │
│  fleet support                      │
│                                     │
│         ● ● ● ○                     │
│                                     │
│         [Complete Setup]            │
│                                     │
└─────────────────────────────────────┘
```

**Onboarding Card Specs:**
- Illustration: 280x280px, centered
- Title: Headline/Medium, White, 24px below illustration
- Description: Body/Large, White 80% opacity, 16px below title
- Page Indicators: 8px circles, 8px gap
  - Active: White
  - Inactive: White 40%
- Button: White background, Blue text, fixed 80px from bottom

---

### Screen 10: Vehicle Inspection

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Inspection            │
│                                     │
│  PRE-TRIP INSPECTION                │  ← Headline/Small
│  Trip: Delhi → Mumbai               │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📸 Exterior Photos            │ │
│  │ ┌────┐ ┌────┐ ┌────┐ ┌────┐  │ │
│  │ │Front│ │Rear│ │Left│ │Right│ │ │
│  │ └────┘ └────┘ └────┘ └────┘  │ │
│  └───────────────────────────────┘ │
│                                     │
│  VEHICLE CONDITION                  │
│  ┌───────────────────────────────┐ │
│  │ ⭕ Tires                       │ │  ← Radio list
│  │ ○ Good  ○ Damaged  ○ Missing  │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ ⭕ Lights                      │ │
│  │ ○ Good  ○ Damaged  ○ Missing  │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ ⭕ Mirrors                     │ │
│  │ ○ Good  ○ Damaged  ○ Missing  │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ ⭕ Fluids                      │ │
│  │ ○ Good  ○ Damaged  ○ Missing  │ │
│  └───────────────────────────────┘ │
│                                     │
│  ODOMETER                           │
│  ┌───────────────────────────────┐ │
│  │ 45,230 km                     │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📝 Notes (Optional)           │ │
│  │                               │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ✅ Confirm & Start Trip       │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Photo Grid Specs:**
- 4 columns, equal width
- Each: 72x72px square
- Corner Radius: 8px
- Background: White
- Border: 2px dashed `Neutral/Gray300`
- Icon: 32x32px, centered
- Label: `Label/Small`, below each

**Radio Group:**
- Option height: 48px
- Radio button: 20px circle
- Spacing: 16px between options
- Selected: Fill `Primary/FlyFleetBlue`, White check

---

### Screen 11: Trip Summary (Post-Trip)

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│            Trip Complete! 🎉        │  ← Headline/Medium, center
│                                     │
│        [Animation: 120px]           │  ← Confetti Lottie
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Delhi → Mumbai                │ │  ← Summary card
│  │                               │ │
│  │  ✅ Completed                 │ │
│  │  Mar 31, 2026 | 2:45 PM       │ │
│  │                               │ │
│  │  📏 1,450 km                  │ │
│  │  ⏱️ 18h 30m                   │ │
│  │  ⛽ 145 L fuel                │ │
│  └───────────────────────────────┘ │
│                                     │
│  EARNINGS                           │  ← Title/Small
│  ┌───────────────────────────────┐ │
│  │ Base Fare         ₹12,500     │ │
│  │ Fuel Surcharge     ₹2,800     │ │
│  │ Toll Charges       ₹1,200     │ │
│  │ ─────────────────────────      │ │
│  │ Total           ₹16,500       │ │  ← Title/Medium
│  └───────────────────────────────┘ │
│                                     │
│  RATING                             │
│  ┌───────────────────────────────┐ │
│  │ Rate your experience          │ │
│  │     ☆ ☆ ☆ ☆ ☆                 │ │  ← 5-star rating
│  │                               │ │
│  │ 📝 Add feedback (Optional)    │ │
│  └───────────────────────────────┘ │
│                                     │
│         [Submit & Continue]         │  ← Primary Button
│                                     │
└─────────────────────────────────────┘
```

**Earnings Row Specs:**
- Height: 48px
- Label: `Body/Medium`, `Neutral/Gray600`
- Value: `Body/Medium`, `Neutral/Gray900`
- Total row: `Title/Medium`, `Primary/FlyFleetBlue`
- Divider: 1px `Neutral/Gray200`

**Star Rating:**
- Star size: 40x40px
- Gap: 8px
- Empty: `Neutral/Gray300`
- Filled: `Semantic/Warning`
- Animation: Scale up on tap

---

### Screen 12: Document Viewer

**Frame:** 360x800  
**Background:** `Neutral/Gray900` (dark for document viewing)

```
┌─────────────────────────────────────┐
│ ← Back  [Download] [Share] [⋮]     │  ← Dark top bar
│                                     │
│  ┌───────────────────────────────┐ │
│  │                               │ │
│  │     [DOCUMENT PREVIEW]        │ │  ← Full width PDF/Image
│  │                               │ │
│  │     Registration Certificate  │ │
│  │     Vehicle No: MH-12-AB-1234 │ │
│  │     Valid Until: Dec 2027     │ │
│  │                               │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Registration Certificate      │ │  ← Bottom sheet
│  │                               │ │
│  │  Type: RC                     │ │
│  │  Valid Until: Dec 15, 2027    │ │
│  │  Status: ✅ Valid             │ │
│  │  Uploaded: Jan 10, 2024       │ │
│  │                               │ │
│  │  [Replace] [Delete]           │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Document Preview:**
- Full width
- Height: calc(100% - 200px)
- Background: White (for contrast)
- Zoom controls: Floating bottom-right

**Bottom Sheet:**
- Height: 200px
- Background: White
- Corner Radius: 24px (top only)
- Padding: 24px
- Shadow: `Shadow/Xl`

---

### Screen 13: Navigation View

**Frame:** 360x800  
**Background:** Map (full screen)

```
┌─────────────────────────────────────┐
│ ← Back        [Layers] [⋮]         │  ← Floating top bar
│                                     │
│                                     │
│        [FULL SCREEN MAP]            │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Next turn in 500m             │ │  ← Navigation card
│  │ → Turn right on NH-48         │ │
│  │                               │ │
│  │  ┌───┐                        │ │
│  │  │ → │  500m                 │ │
│  │  └───┘                        │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 2h 15m | 120 km | 3:45 PM     │ │  ← Bottom bar
│  │ [Mute] [Overview] [Stops]     │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Navigation Card:**
- Position: Floating, 16px from top
- Background: White
- Corner Radius: 16px
- Shadow: `Shadow/Lg`
- Padding: 16px
- Arrow icon: 64x64px, `Primary/FlyFleetBlue`

**Bottom Bar:**
- Height: 80px
- Background: White
- Padding: 16px
- Border radius: 16px (top corners)

---

### Screen 14: Stop Management

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Stops          [+ Add]│
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ✅ STOP 1                     │ │  ← Completed
│  │ Delhi Warehouse               │ │
│  │ 8:00 AM - 9:30 AM             │ │
│  │ Checked in: 8:05 AM           │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🟢 STOP 2 (CURRENT)           │ │  ← Active, green
│  │ Mumbai Warehouse              │ │
│  │ ETA: 2:00 PM                  │ │
│  │                               │ │
│  │  ┌──────────┐ ┌──────────┐   │ │
│  │  │ Check In │ │ Navigate │   │ │
│  │  └──────────┘ └──────────┘   │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ⚪ STOP 3                     │ │  ← Upcoming
│  │ Pune Distribution Center      │ │
│  │ ETA: Tomorrow, 10:00 AM       │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ⚪ STOP 4                     │ │
│  │ Bangalore Hub                 │ │
│  │ ETA: Tomorrow, 8:00 PM        │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Stop Card States:**

| State | Background | Border | Icon |
|-------|------------|--------|------|
| Completed | `Neutral/Gray100` | None | ✅ Green |
| Active | `Semantic/SuccessLight` | 2px `Semantic/Success` | 🟢 Green |
| Upcoming | White | 1px `Neutral/Gray200` | ⚪ Gray |

**Check-In Button:**
- Full width within card
- Height: 48px
- Background: `Semantic/Success`
- Text: `Label/Large`, White

---

### Screen 15: Delay Reporting

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Report Delay          │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ ⚠️ REPORT DELAY               │ │  ← Warning header
│  │ Trip: Delhi → Mumbai          │ │
│  │ Current Location: NH-48       │ │
│  └───────────────────────────────┘ │
│                                     │
│  SELECT REASON                      │
│  ┌───────────────────────────────┐ │
│  │ 🚗 Traffic                    │ │  ← Selectable cards
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔧 Vehicle Breakdown          │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🌧️ Weather Conditions         │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📝 Documentation Delay        │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🏥 Medical Emergency          │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ Other                         │ │
│  └───────────────────────────────┘ │
│                                     │
│  ESTIMATED DELAY                    │
│  ┌──────────┐ ┌──────────┐         │
│  │ Hours    │ │ Minutes  │         │
│  │   2      │ │   30     │         │
│  └──────────┘ └──────────┘         │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📝 Additional Notes           │ │
│  │ (Optional)                    │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📸 Upload Proof (Optional)    │ │
│  └───────────────────────────────┘ │
│                                     │
│         [Submit Report]             │  ← Primary Button
│                                     │
└─────────────────────────────────────┘
```

**Reason Card:**
- Height: 56px
- Padding: 16px
- Background: White
- Corner Radius: 12px
- Border: 1px `Neutral/Gray200`
- Selected: Border 2px `Primary/FlyFleetBlue`, Background `Primary/FlyFleetBlueSubtle`

---

### Screen 16: Fuel History

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Fuel History   [🔍]  │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 📊 This Month                 │ │  ← Summary
│  │ 450 L | ₹42,500               │ │
│  │ Avg: ₹94.44/L                 │ │
│  └───────────────────────────────┘ │
│                                     │
│  FILTER                             │
│  [This Month ▼] [All Stations ▼]   │
│                                     │
│  TODAY                              │
│  ┌───────────────────────────────┐ │
│  │ ⛽ Indian Oil                 │ │
│  │ NH-48, Mumbai-Pune Highway    │ │
│  │ Today, 10:45 AM               │ │
│  │                               │ │
│  │ 45.5 L | ₹4,250 | ₹93.41/L   │ │
│  │                               │ │
│  │ [📄 Receipt] [📍 Map]         │ │
│  └───────────────────────────────┘ │
│                                     │
│  YESTERDAY                          │
│  ┌───────────────────────────────┐ │
│  │ ⛽ HP Petrol Pump             │ │
│  │ Delhi-Jaipur Highway          │ │
│  │ Yesterday, 6:30 PM            │ │
│  │                               │ │
│  │ 38 L | ₹3,572 | ₹94.00/L     │ │
│  │                               │ │
│  │ [📄 Receipt] [📍 Map]         │ │
│  └───────────────────────────────┘ │
│                                     │
│  [Load More]                        │  ← Text button, center
│                                     │
└─────────────────────────────────────┘
```

**Fuel Entry Card:**
- Padding: 16px
- Background: White
- Corner Radius: 12px
- Margin: 12px vertical

**Action Buttons (Inline):**
- Height: 32px
- Padding: 12px
- Background: `Neutral/Gray100`
- Corner Radius: 8px
- Text: `Label/Small`, `Neutral/Gray700`
- Icon: 16x16px

---

### Screen 17: Emergency Contacts

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Emergency Contacts    │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🚨 In case of emergency:      │ │  ← Info card
│  │ Tap SOS button on home screen │ │
│  │ or call 112 for immediate help│ │
│  └───────────────────────────────┘ │
│                                     │
│  PERSONAL CONTACTS                  │
│  ┌───────────────────────────────┐ │
│  │ 👤 Wife                       │ │
│  │ Priya Kumar                   │ │
│  │ +91 98765 43210               │ │
│  │             [📞] [✉️]         │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 👤 Father                     │ │
│  │ Ramesh Kumar                  │ │
│  │ +91 98765 12345               │ │
│  │             [📞] [✉️]         │ │
│  └───────────────────────────────┘ │
│                                     │
│  [Add Personal Contact]             │  ← Secondary Button
│                                     │
│  FLEET EMERGENCY                    │
│  ┌───────────────────────────────┐ │
│  │ 🏢 Fleet Control Room         │ │
│  │ 24/7 Support                  │ │
│  │ 1800-123-4567                 │ │
│  │             [📞] [✉️]         │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🚑 Medical Emergency          │ │
│  │ Nearest Hospital: 12 km       │ │
│  │ City Hospital, Pune           │ │
│  │             [📞] [🗺️]         │ │
│  └───────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Contact Card:**
- Padding: 16px
- Background: White
- Corner Radius: 12px
- Margin: 12px vertical

**Action Icons:**
- Size: 40x40px circles
- Background: `Primary/FlyFleetBlueSubtle`
- Icon: 20x20px, `Primary/FlyFleetBlue`
- Spacing: 8px between

---

### Screen 18: Settings

**Frame:** 360x800  
**Background:** `Neutral/Gray50`

```
┌─────────────────────────────────────┐
│ ← Back        Settings              │
│                                     │
│  NOTIFICATIONS                      │
│  ┌───────────────────────────────┐ │
│  │ 🔔 Trip Updates           ⊂⊃  │ │  ← Toggle
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔔 Alerts                 ⊂⊃  │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔔 Fuel Reminders         ⊂⊃  │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔔 Promotions             ⊂⊃  │ │
│  └───────────────────────────────┘ │
│                                     │
│  PRIVACY                            │
│  ┌───────────────────────────────┐ │
│  │ 📍 Location Sharing       →   │ │
│  │    Always                    │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📊 Usage Data             ⊂⊃  │ │
│  └───────────────────────────────┘ │
│                                     │
│  APP SETTINGS                       │
│  ┌───────────────────────────────┐ │
│  │ 🌐 Language               →   │ │
│  │    English                   │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📱 App Version            →   │ │
│  │    1.0.0 (Build 45)          │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 📄 Terms of Service       →   │ │
│  └───────────────────────────────┘ │
│  ┌───────────────────────────────┐ │
│  │ 🔒 Privacy Policy         →   │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ 🗑️ Clear Cache                │ │
│  │    45 MB                     │ │
│  └───────────────────────────────┘ │
│                                     │
│         [Sign Out]                  │  ← Destructive Button
│                                     │
└─────────────────────────────────────┘
```

**Toggle Switch:**
```
Frame: 52x32px

Track:
- On: Fill Semantic/Success
- Off: Fill Neutral/Gray300
- Corner Radius: 16px

Thumb:
- Size: 28x28px
- Fill: White
- Shadow: Shadow/Md
- Position: Right (on), Left (off)

Animation: 200ms ease-out
```

---

## 🎨 Component Variations

### Toggle Switch States

| State | Track Color | Thumb Position |
|-------|-------------|----------------|
| On | `Semantic/Success` | Right (52px - 28px = 24px) |
| Off | `Neutral/Gray300` | Left (0px) |
| Disabled On | `Semantic/Success` 50% | Right, opacity 50% |
| Disabled Off | `Neutral/Gray300` 50% | Left, opacity 50% |

---

### Empty State Variations

| Screen | Illustration | Title | Subtitle |
|--------|-------------|-------|----------|
| No Trips | `illustration-empty-trips` | No trips scheduled | New trips will appear here once assigned |
| No Network | `illustration-no-network` | You're offline | Most features work offline. Changes will sync when connected. |
| No Alerts | `illustration-success` | All caught up! | No new alerts at this time |
| No Documents | `illustration-empty-documents` | No documents | Upload your vehicle and personal documents here |

---

### Loading State Variations

| Type | Component | Usage |
|------|-----------|-------|
| Skeleton Card | Gray rectangle with `Shadow/Sm` | List loading |
| Circular Spinner | 48px, `Primary/FlyFleetBlue` | Button loading |
| Shimmer | Animated gradient | Image loading |
| Progress Bar | Linear, indeterminate | Sync progress |

---

## 📐 Spacing Quick Reference

```
Screen Padding: 16px
Card Padding: 16-20px
List Item Padding: 16px
Between Cards: 12-16px
Between Sections: 24px
Between Elements: 8-16px

Button Height:
- Large: 56px
- Medium: 48px
- Small: 40px

Input Height: 56px
Card Corner Radius: 12-16px
Button Corner Radius: 12px
Icon Size: 24x24px
Touch Target Minimum: 48x48dp
```

---

## 🎯 Design Review Checklist

Before finalizing each screen:
- [ ] Uses approved color styles
- [ ] Uses approved text styles
- [ ] All touch targets ≥ 48x48dp
- [ ] Color contrast ≥ 4.5:1 for text
- [ ] Consistent spacing (8px grid)
- [ ] All icons have labels or context
- [ ] Empty states designed
- [ ] Loading states designed
- [ ] Error states designed
- [ ] Success states designed

---

**Created:** March 31, 2026  
**Continued from:** `FIGMA_MOCKUP_SPECS.md`  
**Status:** Complete - All 18 screens specified

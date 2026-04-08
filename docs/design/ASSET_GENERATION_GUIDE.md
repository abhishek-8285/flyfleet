# FlyFleet Driver App - Asset Generation Guide

**Version:** 1.0  
**Date:** March 31, 2026  
**Purpose:** Guide for creating/collecting all visual assets needed for the Driver App

---

## 📋 Asset Overview

| Category | Count | Formats | Priority |
|----------|-------|---------|----------|
| Icons | 28 | SVG | 🔴 P0 |
| Illustrations | 8 | PNG (transparent) | 🟡 P1 |
| Lottie Animations | 5 | JSON | 🟡 P1 |
| Product Screenshots | 18 | PNG | 🟢 P2 |
| Marketing Assets | 4 | PNG/SVG | 🟢 P2 |

**Total:** 63 assets

---

## 🎨 Icon Specifications

### Technical Requirements

```
Format: SVG (vector, scalable)
Size: 24x24px (canonical), scale for @2x, @3x
Stroke: 2px (consistent across all icons)
Color: #111827 (Gray900) - override via CSS
Style: Outline (mostly), some filled
Corner Radius: 2px (rounded joins)
```

### Icon Grid Template

```
┌────────────────────────────┐
│                            │
│    ┌──────────────────┐   │
│    │                  │   │
│    │   24x24px Area   │   │
│    │                  │   │
│    │      [Icon]      │   │
│    │                  │   │
│    └──────────────────┘   │
│                            │
│    48x48px Touch Target   │
│                            │
└────────────────────────────┘
```

### Icon List (28 total)

#### Navigation Icons (Bottom Bar)

| Icon Name | File Name | States | Description |
|-----------|-----------|--------|-------------|
| Home | `ic-home.svg` | Outline + Filled | House with door |
| Trips | `ic-trips.svg` | Outline + Filled | Clipboard with checklist |
| Fuel | `ic-fuel.svg` | Outline + Filled | Fuel pump nozzle |
| HOS | `ic-hos.svg` | Outline + Filled | Clock with hands |
| Profile | `ic-profile.svg` | Outline + Filled | Person silhouette |

**Reference:** Material Icons: `home`, `assignment`, `local_gas_station`, `access_time`, `person`

---

#### Action Icons

| Icon Name | File Name | Description |
|-----------|-----------|-------------|
| Back | `ic-back.svg` | Chevron left |
| Menu | `ic-menu.svg` | Hamburger (3 lines) |
| Notifications | `ic-notifications.svg` | Bell with dot |
| Edit | `ic-edit.svg` | Pencil |
| Camera | `ic-camera.svg` | Camera body |
| Upload | `ic-upload.svg` | Arrow up from box |
| Location | `ic-location.svg` | Map pin |
| Phone | `ic-phone.svg` | Phone handset |
| Email | `ic-email.svg` | Envelope |
| Document | `ic-document.svg` | Paper with fold |

**Reference:** Material Icons: `arrow_back`, `menu`, `notifications`, `edit`, `camera_alt`, `upload`, `location_on`, `phone`, `email`, `description`

---

#### Status Icons

| Icon Name | File Name | Description | Fill |
|-----------|-----------|-------------|------|
| Check | `ic-check.svg` | Checkmark | Outline |
| Close | `ic-close.svg` | X mark | Outline |
| Error | `ic-error.svg` | Exclamation in circle | Outline |
| Warning | `ic-warning.svg` | Exclamation in triangle | **Filled** |
| Info | `ic-info.svg` | i in circle | Outline |

**Reference:** Material Icons: `check_circle`, `close`, `error`, `warning`, `info`

---

#### Navigation & Control

| Icon Name | File Name | Description |
|-----------|-----------|-------------|
| Chevron Right | `ic-chevron-right.svg` | Right arrow |
| Chevron Down | `ic-chevron-down.svg` | Down arrow |
| More Vertical | `ic-more-vertical.svg` | 3 dots vertical |
| SOS | `ic-sos.svg` | SOS text with beacon |
| Vehicle | `ic-vehicle.svg` | Truck side view |
| Driver | `ic-driver.svg` | Driver license card |
| Navigation | `ic-navigation.svg` | Arrow in diamond |
| Clock | `ic-clock.svg` | Analog clock |
| Calendar | `ic-calendar.svg` | Calendar grid |

**Reference:** Material Icons: `chevron_right`, `expand_more`, `more_vert`, `sos`, `local_shipping`, `badge`, `navigation`, `schedule`, `calendar_today`

---

### Icon Export Checklist

- [ ] All icons created in SVG format
- [ ] 24x24px viewBox
- [ ] 2px stroke width (consistent)
- [ ] Outlines converted to strokes where possible
- [ ] Named with `ic-` prefix
- [ ] Organized in `/icons/` folder
- [ ] Both outline and filled variants (where needed)
- [ ] Tested on light and dark backgrounds

---

## 🖼️ Illustration Specifications

### Technical Requirements

```
Format: PNG with transparent background
Sizes: 
  - Large: 280x280px (onboarding)
  - Medium: 200x200px (empty states)
  - Small: 120x120px (success/error)
Style: Flat, modern, minimal
Colors: Match FlyFleet palette
Stroke: 2-3px outlines
```

### Illustration List

#### Onboarding (4 illustrations)

| Name | File Name | Size | Description |
|------|-----------|------|-------------|
| Welcome | `illustration-onboarding-welcome.png` | 280x280px | Happy driver with truck, smartphone in hand, road in background |
| Trips | `illustration-onboarding-trips.png` | 280x280px | Map with route pins, truck on path, checkmarks |
| Compliance | `illustration-onboarding-compliance.png` | 280x280px | Documents with checkmarks, shield, clock |
| Safety | `illustration-onboarding-safety.png` | 280x280px | SOS beacon, shield, phone with support agent |

**Style Reference:** Blush Design, ManyStories style  
**Colors:** Primary blue (#2563EB), Success green (#10B981), Warning amber (#F59E0B)

---

#### Empty States (4 illustrations)

| Name | File Name | Size | Description |
|------|-----------|------|-------------|
| No Trips | `illustration-empty-trips.png` | 200x200px | Sad truck parked, calendar with no entries |
| No Network | `illustration-no-network.png` | 200x200px | Phone with WiFi slash, cloud with X |
| No Documents | `illustration-empty-documents.png` | 200x200px | Empty folder, document with question mark |
| Success | `illustration-success.png` | 120x120px | Checkmark in circle, confetti |

---

#### Feedback States (2 illustrations)

| Name | File Name | Size | Description |
|------|-----------|------|-------------|
| Error | `illustration-error.png` | 120x120px | Exclamation mark, red background circle |
| Loading | `illustration-loading.png` | 120x120px | Spinning loader, progress indicator |

---

### Illustration Style Guide

**DO:**
- ✅ Use flat design with minimal shading
- ✅ Consistent 2-3px stroke outlines
- ✅ Limited color palette (brand colors)
- ✅ Simple, recognizable shapes
- ✅ Transparent backgrounds
- ✅ Expressive characters (when included)

**DON'T:**
- ❌ Gradients or complex shading
- ❌ Photorealistic elements
- ❌ Too many details (keep it simple)
- ❌ Inconsistent stroke widths
- ❌ White backgrounds (must be transparent)
- ❌ Text within illustrations

---

### Illustration Export Checklist

- [ ] PNG format with transparency
- [ ] Correct size (280px, 200px, or 120px)
- [ ] @2x and @3x variants exported
- [ ] Colors match brand palette
- [ ] Named with `illustration-` prefix
- [ ] Organized in `/illustrations/` folder
- [ ] Tested on light and dark backgrounds

---

## 🎬 Lottie Animation Specifications

### Technical Requirements

```
Format: JSON (Bodymovin/Lottie)
Size: Canvas 200x200px (scalable)
Frame Rate: 30fps
Colors: Match brand palette
Style: Consistent with illustrations
```

### Animation List

| Name | File Name | Duration | Loop | Description |
|------|-----------|----------|------|-------------|
| SOS Pulse | `anim-sos-pulse.json` | 1s | Yes | Red circle pulsing outward |
| Sync Spinner | `anim-sync-spinner.json` | 1.5s | Yes | Circular progress with dots |
| Success Check | `anim-success-check.json` | 800ms | No | Checkmark drawing + bounce |
| Loading Dots | `anim-loading-dots.json` | 1s | Yes | Three dots bouncing |
| Confetti | `anim-confetti.json` | 2s | No | Confetti falling from top |

---

#### SOS Pulse Animation

**Specs:**
```
Canvas: 200x200px
Elements: 3 concentric circles
Animation: Scale from 1.0 → 1.5, opacity 100% → 0%
Timing: Staggered (0ms, 300ms, 600ms)
Color: #EF4444 (Error Red)
Easing: ease-out
```

---

#### Success Check Animation

**Specs:**
```
Canvas: 200x200px
Elements: Circle background + Checkmark
Animation:
  0-400ms: Circle scales 0 → 1 (bounce)
  400-600ms: Checkmark draws (stroke-dashoffset)
  600-800ms: Checkmark bounces (scale 1.1 → 1)
Color: Circle #10B981, Checkmark White
Easing: ease-out
```

---

#### Confetti Animation

**Specs:**
```
Canvas: 200x200px
Elements: 20-30 confetti pieces
Animation: Fall from top with rotation
Colors: #2563EB, #10B981, #F59E0B, #EF4444
Timing: Random stagger 0-500ms
Duration: 2s (one-shot, no loop)
```

---

### Lottie Export Checklist

- [ ] JSON format (Bodymovin v5+)
- [ ] All shapes converted to vectors
- [ ] No images embedded
- [ ] Colors match brand palette
- [ ] Loop settings correct
- [ ] File size < 50KB each
- [ ] Named with `anim-` prefix
- [ ] Organized in `/animations/` folder
- [ ] Tested in Lottie preview

---

## 📸 Product Screenshot Specifications

### Technical Requirements

```
Format: PNG
Size: 1080x2400px (Full HD phone)
Device Frame: Pixel 7 or similar
Orientation: Portrait
Quality: Maximum (lossless)
```

### Screenshot List (18 screens)

| # | Screen | File Name |
|---|--------|-----------|
| 1 | Login | `screenshot-01-login.png` |
| 2 | Home (Trip Dashboard) | `screenshot-02-home.png` |
| 3 | Trip Details | `screenshot-03-trip-details.png` |
| 4 | Fuel Log | `screenshot-04-fuel-log.png` |
| 5 | HOS Status | `screenshot-05-hos.png` |
| 6 | SOS Emergency | `screenshot-06-sos.png` |
| 7 | Alert Inbox | `screenshot-07-alerts.png` |
| 8 | Profile | `screenshot-08-profile.png` |
| 9 | Onboarding 1 | `screenshot-09-onboarding-1.png` |
| 10 | Onboarding 2 | `screenshot-10-onboarding-2.png` |
| 11 | Onboarding 3 | `screenshot-11-onboarding-3.png` |
| 12 | Onboarding 4 | `screenshot-12-onboarding-4.png` |
| 13 | Vehicle Inspection | `screenshot-13-inspection.png` |
| 14 | Trip Summary | `screenshot-14-summary.png` |
| 15 | Document Viewer | `screenshot-15-documents.png` |
| 16 | Navigation | `screenshot-16-navigation.png` |
| 17 | Stop Management | `screenshot-17-stops.png` |
| 18 | Settings | `screenshot-18-settings.png` |

---

### Screenshot Guidelines

**DO:**
- ✅ Use realistic data (not "Lorem Ipsum")
- ✅ Show typical content (not empty states)
- ✅ Include status bar (time, battery, signal)
- ✅ Use Indian context (names, locations, ₹)
- ✅ Consistent device frame for all
- ✅ High resolution (for marketing use)

**DON'T:**
- ❌ Show sensitive data (real phone numbers, etc.)
- ❌ Use placeholder text in final screenshots
- ❌ Crop out status bar
- ❌ Mix different device frames
- ❌ Low resolution or compressed

---

### Screenshot Data Guidelines

**Use realistic Indian context:**
- Names: Rajesh Kumar, Priya Sharma, Amit Patel
- Locations: Delhi, Mumbai, Pune, Bangalore, NH-48
- Phone: +91 98765 43210
- Vehicle: TATA Ace, MH-12-AB-1234
- Currency: ₹ (Rupees)
- Date Format: 31 Mar 2026

---

## 🎯 Marketing Assets

### App Store Icons

| Asset | File Name | Size | Format |
|-------|-----------|------|--------|
| App Icon | `app-icon-1024.png` | 1024x1024px | PNG |
| App Icon | `app-icon-512.png` | 512x512px | PNG |
| App Icon | `app-icon-192.png` | 192x192px | PNG |
| App Icon | `app-icon-48.png` | 48x48px | PNG |

**Design:** FlyFleet logo on brand blue background, no text

---

### Feature Graphics

| Asset | File Name | Size | Usage |
|-------|-----------|------|-------|
| Feature Graphic | `feature-graphic-1024x500.png` | 1024x500px | Google Play |
| Promo Video | `promo-video-30s.mp4` | 1920x1080 | YouTube |
| Screenshots Grid | `screenshots-grid.png` | 2400x1600 | Website |

---

## 📁 File Organization

```
assets/
├── icons/
│   ├── navigation/
│   │   ├── ic-home.svg
│   │   ├── ic-trips.svg
│   │   ├── ic-fuel.svg
│   │   ├── ic-hos.svg
│   │   └── ic-profile.svg
│   ├── actions/
│   │   ├── ic-back.svg
│   │   ├── ic-menu.svg
│   │   ├── ic-camera.svg
│   │   └── ...
│   ├── status/
│   │   ├── ic-check.svg
│   │   ├── ic-close.svg
│   │   ├── ic-error.svg
│   │   └── ...
│   └── control/
│       ├── ic-chevron-right.svg
│       ├── ic-more-vertical.svg
│       └── ...
│
├── illustrations/
│   ├── onboarding/
│   │   ├── illustration-onboarding-welcome.png
│   │   ├── illustration-onboarding-trips.png
│   │   ├── illustration-onboarding-compliance.png
│   │   └── illustration-onboarding-safety.png
│   ├── empty-states/
│   │   ├── illustration-empty-trips.png
│   │   ├── illustration-no-network.png
│   │   └── ...
│   └── feedback/
│       ├── illustration-success.png
│       └── illustration-error.png
│
├── animations/
│   ├── anim-sos-pulse.json
│   ├── anim-sync-spinner.json
│   ├── anim-success-check.json
│   ├── anim-loading-dots.json
│   └── anim-confetti.json
│
├── screenshots/
│   ├── screenshot-01-login.png
│   ├── screenshot-02-home.png
│   └── ...
│
└── marketing/
    ├── app-icon-1024.png
    ├── feature-graphic-1024x500.png
    └── ...
```

---

## 🛠️ Tools & Resources

### Recommended Tools

| Task | Tool | Cost |
|------|------|------|
| Icon Design | Figma, Illustrator | Free / Paid |
| Illustration | Figma, Procreate | Free / Paid |
| Lottie Animation | After Effects + Bodymovin | Paid |
| Lottie (Free Alternative) | Haiku Animator | Free |
| Screenshot | Android Emulator | Free |
| File Optimization | TinyPNG, SVGO | Free |

### Resources

- **Material Icons:** https://fonts.google.com/icons
- **LottieFiles:** https://lottiefiles.com/ (free animations)
- **unDraw:** https://undraw.co/ (free illustrations)
- **Figma Community:** https://figma.com/community (free assets)

---

## ✅ Asset Quality Checklist

Before delivering assets:

### Icons
- [ ] All SVGs are valid XML
- [ ] Consistent 2px stroke
- [ ] No rasterized elements
- [ ] Named correctly
- [ ] Both variants (outline/filled) where needed

### Illustrations
- [ ] PNG with transparency
- [ ] Correct dimensions
- [ ] @2x and @3x exported
- [ ] Colors match brand
- [ ] No white backgrounds

### Animations
- [ ] JSON valid (test in Lottie preview)
- [ ] File size < 50KB
- [ ] Loop settings correct
- [ ] Colors match brand
- [ ] Smooth at 30fps

### Screenshots
- [ ] High resolution (1080x2400+)
- [ ] Realistic data
- [ ] Indian context
- [ ] Status bar visible
- [ ] Consistent device frame

---

## 📦 Delivery Format

### For Development Handoff

1. **Organized folder structure** (as above)
2. **Asset manifest** (CSV with all files)
3. **Export settings document** (this file)
4. **Figma file link** (if applicable)
5. **Lottie preview links** (for animations)

### Asset Manifest Template

```csv
File Name,Category,Size,Format,Status,Notes
ic-home.svg,Icon,24x24,SVG,✅ Complete,Outline + Filled variants
illustration-welcome.png,Illustration,280x280,PNG,⏳ In Progress,Onboarding 1
anim-sos-pulse.json,Animation,200x200,JSON,⏳ In Progress,1s loop
```

---

**Created:** March 31, 2026  
**Maintained By:** Design Team  
**Next Review:** When adding new assets

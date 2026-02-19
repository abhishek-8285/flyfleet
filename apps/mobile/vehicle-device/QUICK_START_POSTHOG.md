# PostHog Analytics - Quick Start Guide

**Status:** ✅ Ready to Deploy
**Current:** PostHog Cloud (Free Tier)
**Future:** Self-Hosted (when you scale)

---

## 5-Minute Setup

### Step 1: Sign Up for PostHog Cloud (2 minutes)

1. Go to https://posthog.com
2. Click "Get Started Free"
3. Sign up with email or GitHub
4. **No credit card required**

### Step 2: Create Project (1 minute)

1. Click "Create Project"
2. Name: **"FlyFleet Vehicle Device"**
3. Region: Choose closest (US or EU)
4. Click "Create"

### Step 3: Get API Key (1 minute)

1. Click gear icon ⚙️ (Project Settings)
2. Click **"API Keys"**
3. Copy **"Project API Key"** (starts with `phc_`)
4. Copy **"Instance URL"** (https://app.posthog.com)

### Step 4: Update build.gradle (1 minute)

Open: `apps/mobile/vehicle-device/android-device/build.gradle`

Find this section:
```kotlin
defaultConfig {
    // PostHog Configuration (defaults to cloud)
    buildConfigField "String", "POSTHOG_API_KEY", "\"phc_YOUR_POSTHOG_API_KEY\""
    buildConfigField "String", "POSTHOG_HOST", "\"https://app.posthog.com\""
}
```

Replace with your API key:
```kotlin
defaultConfig {
    buildConfigField "String", "POSTHOG_API_KEY", "\"phc_YOUR_ACTUAL_API_KEY_HERE\""
    buildConfigField "String", "POSTHOG_HOST", "\"https://app.posthog.com\""
}
```

### Step 5: Build & Test (Optional, 5 minutes)

```bash
cd apps/mobile/vehicle-device/android-device
./gradlew assembleDebug

# Install on device
adb install -r build/outputs/apk/debug/app-debug.apk

# Check logs
adb logcat | grep PostHog
```

You should see:
```
PostHog Analytics initialized
PostHog event: device_started - {...}
```

### Step 6: View Events in PostHog (1 minute)

1. Go to PostHog dashboard
2. Click **"Activity"** tab
3. See events streaming in real-time! 🎉

---

## What's Tracked Automatically

### Device Lifecycle
- ✅ Device started
- ✅ Device stopped
- ✅ Device registered

### Camera Operations
- ✅ Camera started/stopped
- ✅ Clip recorded
- ✅ Clip uploaded (success/failure)

### Telemetry
- ✅ Telemetry batch uploaded

### Driver Operations
- ✅ Driver login (success/failure)
- ✅ Driver logout
- ✅ HOS status updated

### Events
- ✅ Safety events detected
- ✅ Errors tracked

---

## Free Tier Limits

| Metric | Limit | Your Usage (Estimate) |
|--------|-------|----------------------|
| **Events/Month** | 1,000,000 | ~100,000 (100 devices) |
| **Data Retention** | 12 months | ✅ Included |
| **Dashboards** | Unlimited | ✅ Unlimited |
| **Feature Flags** | Unlimited | ✅ Unlimited |
| **Users** | Unlimited | ✅ Unlimited |

**You're covered for ~1,000 devices on free tier!**

---

## Your First Dashboard

### Create "Device Health" Dashboard

1. Go to **Dashboards** → **New Dashboard**
2. Name: **"Device Health"**
3. Click **"Add insight"**

#### Insight 1: Active Devices (Last 24h)

```
Event: device_started
Filter: Last 24 hours
Aggregation: Count unique (device_id)
```

#### Insight 2: Error Rate

```
Event: error
Filter: Last 24 hours
Aggregation: Count
```

#### Insight 3: Upload Success Rate

```
Event: clip_uploaded
Filter: success = true
Aggregation: Count
```

4. **Save dashboard**

---

## Feature Flags (Optional)

### Example: Gradual AI Rollout

1. Go to **Feature Flags** → **New Flag**
2. Key: `ai_detection`
3. Name: "AI Detection"
4. Rollout: **10%** (start small)
5. Save

### Use in Android App

```kotlin
// In your code
if (analytics.isFeatureEnabled("ai_detection")) {
    // Enable AI detection
    startAIDetection()
}
```

### Update Rollout

1. Monitor error rates
2. If stable, increase to **25%** → **50%** → **100%**
3. No app update needed! 🎉

---

## When to Migrate to Self-Hosted

### Indicators

- ✅ Events > 1M/month (~1,000+ devices)
- ✅ Need custom data retention (>12 months)
- ✅ Need data sovereignty (GDPR requirements)
- ✅ Want to reduce costs at scale
- ✅ Need custom integrations

### Cost Comparison

| Scale | Cloud Cost | Self-Hosted Cost |
|-------|-----------|------------------|
| 100 devices | **Free** | $70-100/month |
| 1,000 devices | **Free** | $70-100/month |
| 5,000 devices | $365/month | $100-200/month |
| 10,000 devices | $730/month | $200-300/month |

**Break-even:** ~3,000-5,000 devices

### Migration Guide

See: `CLOUD_TO_SELF_HOSTED_MIGRATION.md`

**Migration effort:** 2-4 hours
**Risk:** Low (can rollback anytime)

---

## Troubleshooting

### Events Not Appearing

1. **Check API key** is correct
2. **Check internet** connectivity
3. **Enable debug mode:**
   ```kotlin
   debug = true // In VehicleDeviceApplication.kt
   ```
4. **Check logs:**
   ```bash
   adb logcat | grep PostHog
   ```

### High Battery Usage

1. **Increase flush interval:**
   ```kotlin
   flushInterval = 60 // seconds (default 30)
   ```
2. **Reduce max queue size:**
   ```kotlin
   maxQueueSize = 500 // (default 1000)
   ```

### Feature Flags Not Working

1. **Check flag is released** in PostHog
2. **Check flag key** matches exactly
3. **Wait 5 minutes** for flags to sync

---

## Resources

### Documentation
- **PostHog Docs:** https://posthog.com/docs
- **Android SDK:** https://posthog.com/docs/libraries/android
- **Feature Flags:** https://posthog.com/docs/feature-flags

### Support
- **Slack Community:** https://posthog.com/slack
- **GitHub Issues:** https://github.com/PostHog/posthog/issues
- **Email Support:** support@posthog.com (cloud only)

### Tutorials
- **Getting Started:** https://posthog.com/tutorials
- **Product Analytics:** https://posthog.com/docs/product-analytics
- **Session Recording:** https://posthog.com/docs/session-replay

---

## Next Steps

### Today (15 minutes)
- [x] Sign up for PostHog Cloud
- [x] Create project
- [x] Get API key
- [x] Update build.gradle
- [ ] Build and deploy app
- [ ] Verify events appearing

### This Week
- [ ] Create Device Health dashboard
- [ ] Set up error rate alerts
- [ ] Create feature flag for AI detection
- [ ] Test feature flag rollout

### This Month
- [ ] Monitor event volume
- [ ] Optimize event tracking
- [ ] Create additional dashboards
- [ ] Train team on PostHog

### Next Quarter
- [ ] Evaluate migration to self-hosted
- [ ] If >1M events/month, plan migration
- [ ] Set up self-hosted test environment

---

## Summary

✅ **PostHog Cloud** - Free for up to 1M events/month
✅ **Easy Setup** - 5 minutes to get started
✅ **No Credit Card** - Free tier is truly free
✅ **Feature Rich** - Analytics, flags, error tracking
✅ **Easy Migration** - Self-host when you scale

**You're all set! Start tracking events now!** 🚀

---

## Questions?

**Common Questions:**

**Q: Can I use the same API key for debug and release?**
A: Yes, but recommended to use separate projects for better isolation.

**Q: What happens if I exceed 1M events?**
A: PostHog will notify you. You can upgrade to paid tier or migrate to self-hosted.

**Q: Can I export my data?**
A: Yes, PostHog allows data export in CSV/JSON format.

**Q: Is my data secure?**
A: Yes, PostHog is SOC 2 Type II certified and GDPR compliant.

**Q: How do I delete data for GDPR?**
A: Use `analytics.reset()` in app, or delete via PostHog dashboard.

---

**Happy Tracking! 📊**

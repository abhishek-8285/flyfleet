# PostHog Cloud to Self-Hosted Migration Guide

**Current:** PostHog Cloud (Free Tier - 1M events/month)
**Future:** Self-Hosted (Full Control)
**Migration Effort:** Low (1-2 hours)

---

## Phase 1: Start with PostHog Cloud (Now)

### Setup Steps (15 Minutes)

#### 1. Sign Up for PostHog Cloud

1. Go to https://posthog.com
2. Click "Get Started Free"
3. Sign up with email or GitHub
4. No credit card required

#### 2. Create Project

1. Click "Create Project"
2. Name: "FlyFleet Vehicle Device"
3. Region: Choose closest to your users (US/EU)
4. Click "Create"

#### 3. Get API Key

1. Go to **Project Settings** (gear icon)
2. Click **API Keys**
3. Copy **Project API Key** (starts with `phc_`)
4. Copy **Instance URL** (https://app.posthog.com)

#### 4. Update Android App Config

**Option A: Hardcode (Development)**

```kotlin
// In DeviceConfig.kt companion object
private const val DEFAULT_POSTHOG_API_KEY = "phc_YOUR_ACTUAL_API_KEY_HERE"
private const val DEFAULT_POSTHOG_HOST = "https://app.posthog.com"
```

**Option B: Remote Config (Production)**

```kotlin
// Fetch from backend config endpoint
// Or use Firebase Remote Config until migration
```

**Option C: Build Config (Recommended)**

```kotlin
// In build.gradle
android {
    buildTypes {
        debug {
            buildConfigField "String", "POSTHOG_API_KEY", "\"phc_DEBUG_KEY\""
            buildConfigField "String", "POSTHOG_HOST", "\"https://app.posthog.com\""
        }
        release {
            buildConfigField "String", "POSTHOG_API_KEY", "\"phc_RELEASE_KEY\""
            buildConfigField "String", "POSTHOG_HOST", "\"https://app.posthog.com\""
        }
    }
}

// In DeviceConfig.kt
fun getPostHogApiKey(context: Context): String {
    return BuildConfig.POSTHOG_API_KEY
}

fun getPostHogHost(context: Context): String {
    return BuildConfig.POSTHOG_HOST
}
```

#### 5. Test Event Tracking

```kotlin
// In Application or MainActivity
analytics.trackEvent("test_event", mapOf("source" to "android_app"))
```

Check PostHog dashboard:
1. Go to **Activity** tab
2. Should see event within 30 seconds
3. If not, check logs:
   ```kotlin
   debug = true // Enable debug logging
   ```

#### 6. Create First Dashboard

1. Go to **Dashboards** → **New Dashboard**
2. Name: "Device Health"
3. Add insights:
   - Active devices (last 24h)
   - Error rate
   - Upload success rate
4. Save dashboard

---

## Phase 2: Monitor & Scale (1-6 Months)

### Free Tier Limits

| Metric | Limit | Your Usage (Estimate) |
|--------|-------|----------------------|
| **Events/Month** | 1M | ~100k (100 devices) |
| **Retention** | 12 months | ✅ Included |
| **Dashboards** | Unlimited | ✅ Unlimited |
| **Feature Flags** | Unlimited | ✅ Unlimited |
| **Users** | Unlimited | ✅ Unlimited |

### When to Migrate

Migrate to self-hosted when:
- ✅ Events > 1M/month (~1000 devices)
- ✅ Need custom data retention
- ✅ Need data sovereignty (GDPR)
- ✅ Want to reduce costs at scale
- ✅ Need custom integrations

### Cost Comparison

| Scale | Cloud Cost | Self-Hosted Cost |
|-------|-----------|------------------|
| 100 devices | Free | $70-100/month |
| 1,000 devices | Free | $70-100/month |
| 5,000 devices | $365/month | $100-200/month |
| 10,000 devices | $730/month | $200-300/month |

**Break-even point:** ~3,000-5,000 devices

---

## Phase 3: Migrate to Self-Hosted (Future)

### Prerequisites

- Docker & Docker Compose installed
- Server with:
  - 4 CPU cores
  - 8GB RAM
  - 100GB storage
  - Ubuntu 20.04+ or similar
- Domain name (e.g., analytics.flyfleet.com)
- SSL certificate (Let's Encrypt)

### Migration Steps (2-4 Hours)

#### Step 1: Deploy Self-Hosted PostHog

```bash
# SSH into server
ssh user@analytics.flyfleet.com

# Create directory
mkdir -p /opt/posthog
cd /opt/posthog

# Create docker-compose.yml
nano docker-compose.yml
# (Paste the docker-compose from COMPLETE_OPEN_SOURCE_STACK.md)

# Start services
docker-compose up -d

# Check status
docker-compose ps
```

#### Step 2: Configure Domain & SSL

```bash
# Install nginx
sudo apt-get install nginx

# Create nginx config
sudo nano /etc/nginx/sites-available/posthog

# Add config:
server {
    listen 80;
    server_name analytics.flyfleet.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Enable site
sudo ln -s /etc/nginx/sites-available/posthog /etc/nginx/sites-enabled/

# Get SSL certificate
sudo certbot --nginx -d analytics.flyfleet.com

# Restart nginx
sudo systemctl restart nginx
```

#### Step 3: Export Data from PostHog Cloud

**Note:** PostHog Cloud allows data export, but historical data migration is complex. For most cases, it's easier to:

1. **Run both in parallel** for 30 days
2. **Switch to self-hosted** for new data
3. **Keep Cloud access** for historical data (read-only)

If you need to migrate historical data:

```bash
# Export from Cloud (PostHog team can help)
# They provide CSV/JSON export

# Import to self-hosted
# Use PostHog import tools
```

#### Step 4: Update Android App

**Option A: Remote Config Switch**

```kotlin
// In DeviceConfig.kt
fun getPostHogHost(context: Context): String {
    return prefs.getString(KEY_POSTHOG_HOST, "https://app.posthog.com") 
        ?: "https://app.posthog.com"
}

// Update via remote config or backend
fun migrateToSelfHosted(host: String) {
    prefs.edit().putString(KEY_POSTHOG_HOST, host).apply()
}
```

**Option B: App Update**

```kotlin
// In build.gradle (release build)
buildConfigField "String", "POSTHOG_HOST", "\"https://analytics.flyfleet.com\""

// Rebuild and deploy app
```

#### Step 5: Test & Verify

1. **Deploy updated app** to test device
2. **Check events** appearing in self-hosted PostHog
3. **Verify dashboards** showing data
4. **Test feature flags** working
5. **Monitor for 24-48 hours**

#### Step 6: Switch All Devices

1. **Gradual rollout:**
   - 1% → 10% → 50% → 100%
2. **Monitor error rates**
3. **Check event volume**
4. **Verify no data loss**

#### Step 7: Decommission Cloud (Optional)

After 30 days of stable self-hosted operation:

1. **Export final data** from Cloud
2. **Cancel Cloud subscription**
3. **Keep Cloud account** for historical access (free)
4. **Update documentation**

---

## Migration Checklist

### Pre-Migration

- [ ] Server provisioned
- [ ] Domain configured
- [ ] SSL certificate installed
- [ ] Docker Compose deployed
- [ ] Self-hosted PostHog accessible
- [ ] API key generated
- [ ] Android app updated
- [ ] Test device ready

### Migration Day

- [ ] Deploy app to 1% of devices
- [ ] Monitor event flow
- [ ] Check dashboards
- [ ] Verify feature flags
- [ ] No errors in logs
- [ ] Increase to 10%
- [ ] Monitor for 24 hours
- [ ] Increase to 50%
- [ ] Monitor for 24 hours
- [ ] Increase to 100%

### Post-Migration

- [ ] All devices migrated
- [ ] Data flowing correctly
- [ ] Dashboards working
- [ ] Alerts configured
- [ ] Documentation updated
- [ ] Team trained
- [ ] Cloud subscription cancelled (optional)

---

## Rollback Plan

If migration fails:

### Immediate Rollback (< 1 hour)

```kotlin
// Revert DeviceConfig to cloud
fun getPostHogHost(context: Context): String {
    return "https://app.posthog.com" // Back to cloud
}
```

### Gradual Rollback

1. **Stop rollout** at current percentage
2. **Revert app** to cloud config
3. **Deploy reverted app**
4. **Monitor for 24 hours**
5. **Investigate issues**
6. **Retry migration** when ready

---

## Common Issues & Solutions

### Issue: Events Not Appearing

**Solution:**
1. Check API key is correct
2. Check host is reachable:
   ```bash
   curl https://analytics.flyfleet.com/api/health
   ```
3. Check PostHog logs:
   ```bash
   docker-compose logs -f posthog
   ```
4. Enable debug mode in app

### Issue: High Memory Usage

**Solution:**
```yaml
# In docker-compose.yml
services:
  posthog:
    environment:
      - WORKER_MEMORY_LIMIT=4g
      - WEB_MEMORY_LIMIT=2g
```

### Issue: Slow Queries

**Solution:**
1. Add database indexes
2. Increase ClickHouse resources
3. Reduce data retention period
4. Archive old data

---

## Support Resources

### PostHog Cloud
- **Docs:** https://posthog.com/docs
- **Slack:** https://posthog.com/slack
- **Email:** support@posthog.com

### Self-Hosted
- **Docs:** https://posthog.com/docs/self-host
- **GitHub:** https://github.com/PostHog/posthog
- **Community:** https://posthog.com/community

---

## Cost Tracking

### Cloud Costs (Current)

| Month | Devices | Events | Cost |
|-------|---------|--------|------|
| Month 1 | 100 | 100k | Free |
| Month 2 | 200 | 200k | Free |
| Month 3 | 500 | 500k | Free |
| Month 6 | 1,000 | 1M | Free |
| Month 12 | 2,000 | 2M | $182/month |

### Self-Hosted Costs (Future)

| Component | Monthly Cost |
|-----------|-------------|
| Server (4 CPU, 8GB) | $50-100 |
| Database Server | $30-50 |
| Storage (100GB) | $10-20 |
| Bandwidth | $10-30 |
| **Total** | **$100-200/month** |

**Savings at 5,000 devices:** $165-530/month

---

## Conclusion

**Current Strategy:**
✅ Start with PostHog Cloud (free)
✅ Migrate to self-hosted at 3,000-5,000 devices
✅ Keep Cloud for historical data access
✅ Gradual migration with rollback plan

**Timeline:**
- **Now:** PostHog Cloud setup (15 minutes)
- **Month 6:** Evaluate migration (if >1M events/month)
- **Month 12:** Migrate to self-hosted (2-4 hours)

**Risk:** Low (easy to migrate, can rollback anytime)

**Ready to start with PostHog Cloud!** 🚀

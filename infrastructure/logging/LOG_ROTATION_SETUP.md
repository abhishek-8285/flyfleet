# Log Rotation Setup for FlyFleet

## Problem Addressed

The static analysis identified large log files in `tmp/logs/`:
- `backend_persist.log: 45MB`
- `backend_final.log: 4.5MB`

Without rotation, these logs can grow indefinitely and cause:
- Disk space exhaustion
- I/O performance degradation
- Difficulty managing log history

## Solution

This document describes how to set up logrotate to automatically rotate and compress logs.

## Files Created

- `infrastructure/logging/flyfleet-logrotate.conf` - Main configuration

## Installation

### Option 1: System-wide (recommended for production)

```bash
# Copy configuration to logrotate.d
sudo cp infrastructure/logging/flyfleet-logrotate.conf /etc/logrotate.d/flyfleet

# Test configuration syntax
sudo logrotate -d /etc/logrotate.d/flyfleet

# Force immediate rotation (dry-run to see what would happen)
sudo logrotate -vf /etc/logrotate.d/flyfleet
```

### Option 2: Manual rotation (development)

If logrotate is not available (e.g., Termux, minimal Docker containers):

```bash
# Create manual rotation script
./scripts/rotate-logs.sh

# Or add to crontab (daily at 2am)
0 2 * * * /data/data/com.termux/files/home/flyfleet/scripts/rotate-logs.sh
```

## Configuration Details

### Application Logs (`tmp/logs/*.log`)
- **Frequency:** Daily
- **Retention:** 7 days
- **Compression:** gzip
- **Permissions:** 0640 (owner: u0_a412, group: u0_a412)

### Container Logs (postgres.log, nats.log, redis.log)
- **Frequency:** Weekly
- **Retention:** 4 weeks
- **Compression:** gzip
- **Post-rotate:** Send HUP signal to container for log reopening

## Verification

Check current log sizes:
```bash
ls -lh tmp/logs/*.log
```

After rotation (next day), you should see:
```bash
ls -lh tmp/logs/
# -rw-r--r--  1 u0_a412 u0_a412   0 Mar 28 00:00 latest.log
# -rw-r--r--  1 u0_a412 u0_a412  4.5M Mar 27 00:00 backend_final.log.gz
# -rw-r--r--  1 u0_a412 u0_a412 45.1M Mar 26 00:00 backend_persist.log.gz
```

## Monitoring

Add alert for high disk usage:
```bash
# Alert if /data partition > 80%
df -h /data | tail -1 | awk '{print $5}' | tr -d '%' | xargs -I {} bash -c 'if [ {} -gt 80 ]; then echo "Disk usage critical: {}%"; fi'
```

## Manual Rotation on Demand

```bash
# Force logrotate now
sudo logrotate -vf /etc/logrotate.d/flyfleet

# Or via script
./scripts/rotate-logs.sh
```

## Troubleshooting

**Logs not rotating?**
```bash
# Check logrotate status
sudo logrotate -d /etc/logrotate.d/flyfleet

# Verify files match pattern
ls -la tmp/logs/*.log
```

**Permission denied?**
```bash
# Ensure correct ownership
sudo chown -R u0_a412:u0_a412 tmp/logs/
```

**Logrotate command not found?**
On Termux, logrotate may not be installed. Use manual script instead:
```bash
./scripts/rotate-logs.sh
```

---

## Related

- Performance profiling: `PERFORMANCE_PROFILING_STRATEGY.md`
- Log management best practices: Keep logs <100MB each, rotate daily for app logs

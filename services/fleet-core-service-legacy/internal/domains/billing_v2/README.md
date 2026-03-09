# Zero-Cost Invoice Generation System

Professional invoice generation with **₹0 cost** using Chrome Headless, local storage, in-memory cache, and SendGrid free tier.

## Features

✅ **PDF Generation** - Chrome Headless (perfect formatting)
✅ **Local Storage** - Organized by year/month
✅ **In-Memory Cache** - 30-day TTL, auto-cleanup
✅ **Email Delivery** - SendGrid free tier (100/day)
✅ **Google Drive Backup** - 15GB free storage
✅ **GST Compliance** - Indian tax requirements
✅ **Health Checks** - Monitor all components
✅ **Zero Cost** - All free services

---

## Quick Start

### 1. Install Dependencies

```bash
# Install Chrome/Chromium
sudo apt-get update
sudo apt-get install -y chromium-browser

# Install rclone (for Google Drive backup)
curl https://rclone.org/install.sh | sudo bash
```

### 2. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your settings
vim .env
```

### 3. Setup Google Drive Backup (Optional)

```bash
# Configure rclone
rclone config

# Follow prompts:
# - Name: gdrive
# - Storage: Google Drive
# - Authenticate via browser
```

### 4. Create Directories

```bash
mkdir -p /opt/flyfleet/data/{invoices,backups}
mkdir -p /var/log/flyfleet
chmod -R 755 /opt/flyfleet/data
```

### 5. Run Application

```bash
cd services/fleet-core-service-legacy
go run cmd/main.go
```

---

## Environment Variables

```bash
# ===== Invoice Storage =====
INVOICE_STORAGE_PATH=/opt/flyfleet/data/invoices
BACKUP_DIR=/opt/flyfleet/data/backups

# ===== Cache =====
CACHE_MAX_SIZE=1000          # Max invoices in cache
CACHE_TTL=720h              # Cache TTL (30 days)

# ===== PDF Generation =====
CHROME_PATH=/usr/bin/chromium
CHROME_TIMEOUT=30s

# ===== Email (SendGrid) =====
SENDGRID_API_KEY=SG.xxxxx    # Get from sendgrid.com
FROM_EMAIL=billing@flyfleet.com
FROM_NAME=FlyFleet Billing
INVOICE_CC_EMAIL=accounting@flyfleet.com  # Optional

# ===== Backup =====
RCLONE_REMOTE=gdrive:flyfleet-backup
RETENTION_DAYS=90
```

---

## API Endpoints

### Download Invoice PDF

```http
GET /api/v1/billing/invoices/{id}/pdf
```

**Response:**
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename=INV-2024-000001.pdf`

**Example:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8084/api/v1/billing/invoices/550e8400-e29b-41d4-a716-446655440000/pdf \
  -o invoice.pdf
```

### Email Invoice

```http
POST /api/v1/billing/invoices/{id}/email
Content-Type: application/json

{
  "email": "customer@example.com",
  "customer_name": "John Doe"
}
```

**Response:**
```json
{
  "status": "sent"
}
```

### Health Check

```http
GET /api/v1/billing/health
```

**Response:**
```json
{
  "status": "healthy",
  "components": {
    "storage": "healthy (1.23 MB, 45 files)",
    "cache": "healthy (12 entries, 0.15 MB)",
    "email": "configured"
  }
}
```

### Billing Statistics

```http
GET /api/v1/billing/stats
```

**Response:**
```json
{
  "storage": {
    "total_size_bytes": 1290000,
    "total_size_mb": 1.23,
    "file_count": 45,
    "base_path": "/opt/flyfleet/data/invoices"
  },
  "cache": {
    "entry_count": 12,
    "total_size_bytes": 157000,
    "total_size_mb": 0.15,
    "total_hits": 234,
    "max_size": 1000
  }
}
```

---

## Backup System

### Manual Backup

```bash
# Run backup
./scripts/backup-invoices.sh

# Check status
./scripts/backup-invoices.sh status

# Verify backup
./scripts/backup-invoices.sh verify
```

### Automatic Backup (Cron)

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/flyfleet/scripts/backup-invoices.sh >> /var/log/flyfleet/backup.log 2>&1
```

### Backup Commands

| Command | Description |
|---------|-------------|
| `backup` | Run full backup (default) |
| `sync` | Sync existing backups to cloud |
| `cleanup` | Clean up old local backups |
| `verify` | Verify latest backup integrity |
| `status` | Show backup status |
| `help` | Show help message |

---

## Directory Structure

```
/opt/flyfleet/
├── data/
│   ├── invoices/
│   │   ├── 2024/
│   │   │   ├── 03/
│   │   │   │   ├── INV-2024-000001.pdf
│   │   │   │   └── INV-2024-000002.pdf
│   │   │   └── 04/
│   │   └── 2025/
│   └── backups/
│       ├── invoices_20240315_020000.tar.gz
│       └── invoices_20240316_020000.tar.gz
├── scripts/
│   └── backup-invoices.sh
└── logs/
    └── backup.log
```

---

## Cost Breakdown

| Service | Free Tier | Your Cost |
|---------|-----------|-----------|
| **PDF Generation** | Chrome (open source) | ₹0 |
| **Storage** | Local disk (~1GB/year) | ₹0 |
| **Backup** | Google Drive (15GB) | ₹0 |
| **Email** | SendGrid (100/day) | ₹0 |
| **Cache** | In-memory (~100MB) | ₹0 |
| **Total** | | **₹0/month** |

### Capacity

- **Storage:** 15GB = ~75,000 invoices (200KB each)
- **Email:** 3,000 invoices/month (100/day)
- **Cache:** 1,000 invoices (30-day TTL)

---

## Troubleshooting

### Chrome Not Found

```
Error: chrome/chromium not found
```

**Solution:**
```bash
sudo apt-get install -y chromium-browser
export CHROME_PATH=/usr/bin/chromium
```

### SendGrid API Key Missing

```
Warning: SendGrid API key not configured, skipping email
```

**Solution:**
1. Sign up at https://sendgrid.com/free
2. Get API key from dashboard
3. Set `SENDGRID_API_KEY` environment variable

### Permission Denied

```
Error: failed to create directory: permission denied
```

**Solution:**
```bash
sudo chown -R $USER:$USER /opt/flyfleet/data
chmod -R 755 /opt/flyfleet/data
```

### rclone Not Configured

```
Error: rclone remote 'gdrive' is not configured
```

**Solution:**
```bash
rclone config
# Follow prompts to configure Google Drive
```

---

## Monitoring

### Check Health

```bash
curl http://localhost:8084/api/v1/billing/health | jq
```

### Check Storage Usage

```bash
du -sh /opt/flyfleet/data/invoices
find /opt/flyfleet/data/invoices -name "*.pdf" | wc -l
```

### Check Cache Stats

```bash
curl http://localhost:8084/api/v1/billing/stats | jq '.cache'
```

### Check Backup Status

```bash
./scripts/backup-invoices.sh status
```

---

## Security

### File Permissions

```bash
# Set secure permissions
chmod 755 /opt/flyfleet/data
chmod 644 /opt/flyfleet/data/invoices/**/*.pdf
```

### API Access

All invoice endpoints require authentication via JWT token.

### Backup Encryption

Google Drive encrypts data at rest automatically.

---

## Scaling Up (When Needed)

| Trigger | Upgrade To | Cost |
|---------|------------|------|
| >3,000 invoices/month | SendGrid Paid | $15/month |
| >15GB storage | Google One (100GB) | ₹130/month |
| Need CDN | Cloudflare + S3 | $5/month |
| >100K invoices | Async queue (SQS) | $50/month |

---

## GST Compliance (India)

### Required Fields

All invoices include:
- ✅ Sequential invoice number
- ✅ GSTIN (your and customer's)
- ✅ HSN/SAC code per item
- ✅ Tax breakup (CGST+SGST or IGST)
- ✅ Place of supply
- ✅ Invoice date

### Retention

Invoices retained for **7 years** (legal requirement).

---

## Support

For issues or questions:
- Check logs: `/var/log/flyfleet/backup.log`
- Health check: `/api/v1/billing/health`
- Documentation: See this README

---

## License

Part of FlyFleet - Enterprise Fleet Management Platform

# Local-First Storage Implementation

**Date:** 2026-02-15
**Approach:** Device-local storage with optional cloud sync

---

## Overview

The FlyFleet Vehicle Device app uses a **local-first storage architecture** where video clips and evidence are stored on the device by default, with optional cloud synchronization when connectivity allows.

### Why Local-First?

1. **Offline Operation**: Vehicles operate in areas with poor/no connectivity
2. **Cost Efficiency**: No cloud storage costs for continuous recording (24/7)
3. **Performance**: Zero upload latency during critical events
4. **Privacy**: Sensitive data stays on device until explicitly uploaded
5. **Bandwidth Optimization**: Only upload event clips, not continuous footage

---

## Storage Architecture

### Local Storage (Default)

```
Device Storage: /sdcard/flyfleet/
├── clips/              # Event-triggered clips
│   ├── clip-uuid-1.mp4
│   ├── clip-uuid-2.mp4
│   └── ...
├── ring_buffer/        # Continuous recording buffer
│   ├── segment-001.mp4
│   ├── segment-002.mp4
│   └── ...
└── metadata/           # Clip metadata and manifests
    ├── manifest.json
    └── attestations.db
```

### Cloud Storage (Optional)

```
Cloud Storage: s3://flyfleet-evidence/{org_id}/{device_id}/
├── clips/
│   ├── clip-uuid-1.mp4
│   └── ...
└── metadata/
    └── manifests/
```

---

## API Changes

### `/media/upload-url` (POST)

**Request:**
```json
{
  "device_id": "DEVICE-001",
  "clip_id": "clip-uuid-123",
  "size_bytes": 52428800,
  "content_type": "video/mp4",
  "duration_seconds": 60,
  "event_type": "harsh_brake_detected",
  "storage_mode": "local"  // "local" (default) or "cloud"
}
```

**Response (Local Storage):**
```json
{
  "local_path": "/sdcard/flyfleet/clips/clip-uuid-123.mp4",
  "object_key": "clips/clip-uuid-123.mp4",
  "storage_mode": "local",
  "max_local_size_mb": 5120
}
```

**Response (Cloud Storage):**
```json
{
  "upload_url": "https://storage.flyfleet.local/upload/clip-uuid-123",
  "expires_at": 1708016400,
  "object_key": "clips/clip-uuid-123.mp4",
  "storage_mode": "cloud",
  "max_local_size_mb": 0
}
```

### `/media/confirm` (POST)

**Request (Local Storage):**
```json
{
  "clip_id": "clip-uuid-123",
  "checksum_sha256": "abc123...",
  "duration_seconds": 60,
  "uploaded_bytes": 52428800,
  "storage_mode": "local",
  "local_path": "/sdcard/flyfleet/clips/clip-uuid-123.mp4"
}
```

**Response:**
```json
{
  "status": "confirmed",
  "storage_location": "/sdcard/flyfleet/clips/clip-uuid-123.mp4",
  "storage_mode": "local",
  "sync_status": "pending_sync"
}
```

---

## Database Schema

### `media_uploads` Table

```sql
CREATE TABLE media_uploads (
    id UUID PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    clip_id VARCHAR(255) UNIQUE NOT NULL,
    object_key TEXT NOT NULL,
    checksum_sha256 VARCHAR(64) NOT NULL,
    duration_seconds INTEGER,
    uploaded_bytes BIGINT,
    status VARCHAR(20) DEFAULT 'pending',
    storage_mode VARCHAR(20) DEFAULT 'local',  -- 'local' or 'cloud'
    sync_status VARCHAR(20) DEFAULT 'pending_sync',  -- 'pending_sync', 'syncing', 'synced', 'failed'
    local_path TEXT,  -- Path on device for local storage
    cloud_url TEXT,   -- Cloud URL if synced
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## Sync Strategy

### When to Sync

1. **Event-Triggered Sync**: Critical events (collision, panic) trigger immediate sync
2. **Wi-Fi Sync**: When device connects to depot Wi-Fi
3. **Scheduled Sync**: Daily/weekly sync during low-activity periods
4. **Storage Pressure**: When local storage > 80% full

### Sync Priority Tiers

| Tier | Events | Sync Timing |
|------|--------|-------------|
| **Tier 1** | Collision, panic, harsh brake | Immediate (cellular) |
| **Tier 2** | Speeding, distraction | Next Wi-Fi connection |
| **Tier 3** | Routine clips | Scheduled sync |
| **Tier 4** | Ring buffer (non-event) | Never (overwritten) |

### Sync Workflow

```
1. Device stores clip locally
   ↓
2. Backend marks as "pending_sync"
   ↓
3. Sync service detects pending clips
   ↓
4. Upload to cloud storage (when connected)
   ↓
5. Backend updates sync_status to "synced"
   ↓
6. Local clip marked for deletion (optional)
```

---

## Storage Management

### Local Storage Quotas

| Storage Type | Default Quota | Configurable |
|--------------|---------------|--------------|
| Event Clips | 5 GB | Yes (2-10 GB) |
| Ring Buffer | 10 GB | Yes (5-20 GB) |
| Total | 15 GB | Yes (10-30 GB) |

### Retention Policies

**Local Storage:**
- Event clips: 14 days (configurable 7-30 days)
- Ring buffer: 15 minutes rolling (configurable 5-30 minutes)
- Critical evidence: Until manually deleted (legal hold)

**Cloud Storage:**
- Event clips: 30 days (configurable 7-90 days)
- Critical evidence: 7 years (regulatory requirement)

### Automatic Cleanup

```kotlin
// Android app implements storage manager
class StorageManager {
    fun cleanup() {
        // Delete clips older than retention period
        deleteExpiredClips()
        
        // Maintain minimum free space
        if (freeSpace < minReserveMB) {
            deleteOldestNonCriticalClips()
        }
        
        // Remove synced clips (if configured)
        if (autoDeleteSynced) {
            deleteSyncedClips()
        }
    }
}
```

---

## Security

### Local Storage Security

1. **Encryption at Rest**: AES-256-GCM for all video files
2. **Key Management**: Android Keystore with hardware backing
3. **Access Control**: App-only access (no user access)
4. **Tamper Detection**: Hash chain for all clips

### File Naming

```
/sdcard/flyfleet/clips/{clip_id}.mp4.enc
```

- `.enc` extension indicates encrypted file
- Clip ID is UUID v4 (unguessable)
- No PII in filenames

### Checksum Verification

```json
{
  "clip_id": "clip-uuid-123",
  "checksum_sha256": "abc123...",  // Verified on upload
  "chain_hash": "def456..."  // Hash chain for integrity
}
```

---

## Android Implementation

### Storage Locations

```kotlin
// External storage (recommended for large files)
val externalClipsDir = File(
    Environment.getExternalStorageDirectory(),
    "flyfleet/clips"
)

// App-specific storage (more secure, less space)
val appClipsDir = File(
    context.getExternalFilesDir(null),
    "clips"
)

// Internal storage (most secure, least space)
val internalClipsDir = File(
    context.filesDir,
    "clips"
)
```

### Permissions Required

```xml
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE"
    android:maxSdkVersion="29" />
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE"
    android:maxSdkVersion="32" />
```

**Note:** Android 10+ (API 29+) uses scoped storage - no permissions needed for app-specific directories.

---

## Backend Sync Service (Future)

A separate sync service will handle cloud synchronization:

```go
type SyncService struct {
    // Fetches clips pending sync
    GetPendingSyncClips() []Clip
    
    // Uploads to cloud storage
    UploadToCloud(clip Clip) error
    
    // Updates sync status
    MarkAsSynced(clipID string) error
    
    // Deletes local clip after sync (optional)
    DeleteLocalClip(clipID string) error
}
```

---

## Benefits Summary

| Benefit | Description |
|---------|-------------|
| **Offline-First** | Works perfectly in zero-connectivity areas |
| **Cost Savings** | ~90% reduction in cloud storage costs |
| **Performance** | Zero upload latency during events |
| **Privacy** | Data stays on device unless explicitly synced |
| **Bandwidth** | Only sync event clips, not continuous footage |
| **Scalability** | Backend doesn't need to handle continuous uploads |

---

## Migration Path

### Phase 1 (Current): Local-Only
- All clips stored locally
- No cloud sync
- Manual export via USB

### Phase 2 (Next): Local + Optional Sync
- Critical events auto-sync
- Wi-Fi sync for non-critical
- Configurable sync policies

### Phase 3 (Future): Hybrid
- Intelligent sync based on connectivity
- Edge processing on device
- Cloud archival for compliance

---

## Conclusion

The local-first storage approach provides the best balance of:
- ✅ Reliability (works offline)
- ✅ Cost efficiency (minimal cloud storage)
- ✅ Performance (no upload latency)
- ✅ Privacy (data stays on device)
- ✅ Scalability (backend handles metadata, not video streams)

This is the industry-standard approach for fleet telematics devices (used by Samsara, Motive, KeepTruckin).

# Admin Support DR API

Disaster Recovery admin APIs served through Web BFF under:

- Base path: `/api/v1/admin/dr`
- Upstream service path: `/api/v1/dr`

## Backups

### List backups
- `GET /api/v1/admin/dr/backups`
- Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "FULL|INCREMENTAL|SNAPSHOT",
      "status": "PENDING|IN_PROGRESS|COMPLETED|FAILED",
      "s3_key": "string",
      "size_bytes": 1234,
      "created_by": "uuid",
      "created_at": "2026-02-19T12:00:00Z",
      "completed_at": "2026-02-19T12:10:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get backup detail
- `GET /api/v1/admin/dr/backups/{id}`

### Verify backup
- `POST /api/v1/admin/dr/backups/{id}/verify`
- Response:
```json
{
  "id": "uuid",
  "backup_id": "uuid",
  "status": "VERIFIED",
  "checksum": "sha256",
  "verified_at": "2026-02-19T12:15:00Z",
  "verified_by": "uuid"
}
```

## Restore Jobs

### Create restore job
- `POST /api/v1/admin/dr/restore-jobs`
- Request:
```json
{
  "backup_id": "uuid",
  "target_environment": "STAGING|UAT|PROD"
}
```
- Response:
```json
{
  "data": {
    "id": "uuid",
    "backup_id": "uuid",
    "status": "PENDING|IN_PROGRESS|COMPLETED|FAILED",
    "target_environment": "STAGING",
    "started_at": "2026-02-19T12:20:00Z"
  }
}
```

### List restore jobs
- `GET /api/v1/admin/dr/restore-jobs`

### Get restore job
- `GET /api/v1/admin/dr/restore-jobs/{id}`

### Get restore job logs
- `GET /api/v1/admin/dr/restore-jobs/{id}/logs`
- Response:
```json
{
  "data": [
    {
      "level": "INFO|WARN|ERROR",
      "message": "string",
      "timestamp": "2026-02-19T12:25:00Z"
    }
  ]
}
```

## DR Tests

### Create DR test
- `POST /api/v1/admin/dr/tests`
- Request:
```json
{
  "name": "Monthly failover drill",
  "description": "Optional text",
  "test_type": "FAILOVER|BACKUP|RESTORE",
  "target_env": "STAGING|UAT|PROD",
  "success_criteria": "RTO: 30 min, RPO: 5 min",
  "scheduled_at": "2026-02-20T02:00:00Z"
}
```
- Response:
```json
{
  "data": {
    "id": "uuid",
    "test_id": "DRT-1739960000",
    "status": "SCHEDULED|RUNNING",
    "progress_percent": 0
  }
}
```

### List DR tests
- `GET /api/v1/admin/dr/tests`
- Response:
```json
{
  "data": [],
  "total": 0,
  "limit": 10,
  "offset": 0
}
```

### Get DR test detail
- `GET /api/v1/admin/dr/tests/{id}`
- Response:
```json
{
  "data": {
    "id": "uuid",
    "test_id": "DRT-1739960000",
    "name": "Monthly failover drill",
    "status": "RUNNING",
    "progress_percent": 50
  },
  "steps": [
    {
      "id": "uuid",
      "step_order": 1,
      "step_name": "Initialize test context",
      "status": "PASSED|RUNNING|PENDING|SKIPPED"
    }
  ],
  "logs": [
    {
      "id": "uuid",
      "timestamp": "2026-02-19T12:30:00Z",
      "level": "RUN|INFO|WARN|ERROR",
      "message": "Step 1 completed"
    }
  ]
}
```

### Abort DR test
- `POST /api/v1/admin/dr/tests/{id}/abort`
- Response:
```json
{
  "status": "aborted"
}
```

### DR test metrics
- `GET /api/v1/admin/dr/tests/metrics`
- Response:
```json
{
  "data": {
    "total_tests": 12,
    "success_rate": 91.6,
    "avg_recovery_time": "95s",
    "next_scheduled": "2026-02-22T02:00:00Z"
  }
}
```

### DR test settings
- `GET /api/v1/admin/dr/tests/settings`
- `PUT /api/v1/admin/dr/tests/settings`
- Body for update:
```json
{
  "schedule_enabled": true,
  "frequency": "WEEKLY|MONTHLY",
  "schedule_day": "SATURDAY|SUNDAY",
  "schedule_time": "02:00",
  "intensity": "NORMAL|HIGH",
  "notification_emails": ["ops@example.com"]
}
```

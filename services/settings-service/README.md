# settings-service

## Purpose
settings-service runtime component in FlyFleet platform.

## Run
Use the root docker compose stack or run the service binary from cmd/main.go.

## Health
- Endpoint: /health.

## API
- `GET /settings/export`
  - Returns a full snapshot of current platform settings, feature flags, localization, templates, notification channels, and limits.
- `GET /settings/audit`
  - Lists settings audit logs with optional filters:
  - `entity_type`, `action`, `limit`, `offset`.
  - Defaults: `limit=50`, `offset=0`.
  - Sensitive fields in audit `changes` payload are redacted before persistence.
- `POST /settings/import`
  - Restores a settings snapshot with upsert behavior.
  - Applies changes transactionally (all-or-nothing).
  - Supports optional `"replace_existing": true` to delete feature flags that are not present in import payload.
  - Returns aggregate import counters (`feature_flags_created`, `templates_updated`, etc.).

### Import Request Example
```json
{
  "platform_settings": {
    "system_name": "Enterprise Ops",
    "support_email": "ops@example.com",
    "default_timezone": "UTC"
  },
  "feature_flags": [
    {
      "key": "new-dispatch-flow",
      "scope": "GLOBAL",
      "enabled": true
    }
  ],
  "replace_existing": false
}
```

### Import Validation Rules
- At least one section must be present (`platform_settings`, `feature_flags`, `localization_settings`, `templates`, `notification_channels`, or `system_limits`).
- `platform_settings.support_email` must be a valid email when `platform_settings` is provided.
- `platform_settings.default_timezone` must be a valid IANA timezone.
- `feature_flags[*].key` must be non-empty.
- `feature_flags[*].key` must be unique (case-insensitive) within one import request.
- `templates[*].key` must be non-empty.
- `templates[*].key` must be unique (case-insensitive) within one import request.
- `localization_settings[*].language_code` must be non-empty.
- `localization_settings[*].language_code` must be unique (case-insensitive) within one import request.
- `notification_channels[*].channel_type` must be non-empty.
- `notification_channels[*].channel_type` must be unique (case-insensitive) within one import request.
- `system_limits.api_rate_limit` cannot be negative.

## Ownership
See docs/SERVICE_OWNERSHIP.md for domain responsibility.

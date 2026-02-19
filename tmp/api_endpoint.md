# FlyFleet API Reference

This document provides a comprehensive reference for the `fleet-core-service` and all auxiliary microservices, including the `web-bff` aggregation layer and "archived" but active system modules.
All endpoints are prefixed with `/api` and require a valid JWT token in the `Authorization` header.

## 1. Vehicle Management
**Base Path**: `/v1/vehicles`

### Create Vehicle
- **Endpoint**: `POST /api/v1/vehicles`
- **Request Body**:
```json
{
  "registration_number": "KA-01-HH-1234",
  "model": "Tata Ace",
  "make": "Tata",
  "year": 2023,
  "color": "White",
  "capacity": 1000,
  "fuel_type": "diesel",
  "vin": "ABC1234567890",
  "registration": "REG12345",
  "insurance": "INS98765"
}
```
- **Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "org_id": "550e8400-e29b-41d4-a716-446655440001",
  "registration_number": "KA-01-HH-1234",
  "model": "Tata Ace",
  "make": "Tata",
  "year": 2023,
  "color": "White",
  "capacity": 1000,
  "fuel_type": "diesel",
  "vin": "ABC1234567890",
  "registration": "REG12345",
  "insurance": "INS98765",
  "status": "active",
  "created_at": "2023-10-25T10:00:00Z",
  "updated_at": "2023-10-25T10:00:00Z"
}
```

### List Vehicles
- **Endpoint**: `GET /api/v1/vehicles`
- **Query Params**: `org_id` (uuid), `page` (int), `limit` (int)
- **Response** (200 OK):
```json
{
  "vehicles": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "org_id": "550e8400-e29b-41d4-a716-446655440001",
      "registration_number": "KA-01-HH-1234",
      "model": "Tata Ace",
      "make": "Tata",
      "year": 2023,
      "color": "White",
      "capacity": 1000,
      "fuel_type": "diesel",
      "vin": "ABC1234567890",
      "registration": "REG12345",
      "insurance": "INS98765",
      "status": "active",
      "created_at": "2023-10-25T10:00:00Z",
      "updated_at": "2023-10-25T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

---

## 2. Driver Management
**Base Path**: `/v1/drivers`

### Create Driver
- **Endpoint**: `POST /api/v1/drivers`
- **Request Body**: `{"full_name": "John Doe", "license_number": "DL12345"}`

### Driver Analytics
- **Endpoint**: `GET /api/v1/drivers/:id/analytics`
- **Response**: `{"safety_score": 95, "trips_completed": 120}`

---

## 3. Trip Management
**Base Path**: `/v1/trips`
*(Standard endpoints: Create, List, Get, Update Status)*

---

## 4. Maintenance & Work Orders
**Base Path**: `/v1/maintenance`, `/v1/work-orders`

### Create Schedule
- **Endpoint**: `POST /api/v1/maintenance/schedules`
- **Request Body**: `{"vehicle_id": "uuid", "type": "preventive", "due_date": "ISO8601"}`

### List Work Orders
- **Endpoint**: `GET /api/v1/work-orders`
- **Response**: List of work orders with `status`, `priority`, `estimated_cost`.

### Inventory Management
- **Endpoint**: `GET /api/v1/inventory`
- **Response**: List of parts with `quantity_on_hand` and `reorder_point`.

---

## 5. IoT & Device Diagnostics
**Base Path**: `/v1/admin/devices`, `/v1/diagnostics`

### List Devices
- **Endpoint**: `GET /api/v1/admin/devices`
- **Response**: List of devices with `status` (online/offline), `battery_level`.

### Device Telemetry
- **Endpoint**: `GET /api/v1/admin/devices/:id/telemetry`
- **Response**: Historical sensor data (speed, fuel, temp).

### Device Sensors & Logs
- **Endpoint**: `GET /api/v1/admin/devices/:id/sensors`
- **Response**: Real-time sensor readings (IO status, analog inputs).
- **Endpoint**: `GET /api/v1/admin/devices/:id/logs`
- **Response**: Device debug logs and connection history.

### Device Commands
- **Endpoint**: `POST /api/v1/admin/devices/:id/commands`
- **Request Body**: `{"command": "ENGINE_CUTOFF", "params": {...}}`

### Firmware Management
- **Endpoint**: `GET /api/firmware`
- **Response**: List of available firmware versions.
- **Endpoint**: `POST /api/firmware/rollouts`
- **Request Body**: `{"version": "v2.1.0", "device_ids": ["uuid1", "uuid2"]}`
- **Response**: Rollout job status.

---

## 6. Advanced Tracking (Telematics)
**Base Path**: `/v1/tracking`

### Ingest GPS Telemetry
- **Endpoint**: `POST /api/v1/tracking/telemetry/gps`
- **Request Body**: `{"latitude": 12.97, "longitude": 77.59, "speed": 65, "timestamp": "ISO8601"}`

### Ingest OBD/Fuel Telemetry
- **Endpoint**: `POST /api/v1/tracking/telemetry/obd`
- **Request Body**: `{"rpm": 2500, "engine_load": 45, "coolant_temp": 90}`
- **Endpoint**: `POST /api/v1/tracking/telemetry/fuel`
- **Request Body**: `{"fuel_level_percent": 85, "range_km": 450}`

### Vehicle History (Playback)
- **Endpoint**: `GET /api/v1/tracking/vehicles/:id/history`
- **Query Params**: `start_time`, `end_time`
- **Response**: Array of breadcrumb locations for map playback.

---

## 7. Integrations & Webhooks
**Base Path**: `/api/integrations`, `/api/webhooks`

### List Integrations
- **Endpoint**: `GET /api/integrations`
- **Response**: List of active integrations (Slack, SAP, etc.) with health status.

### Manage Webhooks
- **Endpoint**: `POST /api/webhooks`
- **Request Body**: `{"url": "https://callback.com", "events": ["TRIP_START"]}`

---

## 8. Reports Module
**Base Path**: `/api/v1/admin/reports`

### Scheduled Reports
- **Endpoint**: `GET /api/v1/admin/reports`
- **Response**: List of scheduled reports.
- **Endpoint**: `POST /api/v1/admin/reports`
- **Request Body**:
```json
{
  "name": "Weekly Fuel Usage",
  "report_type": "FUEL_USAGE",
  "format": "PDF",
  "frequency": "WEEKLY",
  "delivery_time": "08:00:00",
  "recipients": ["manager@flyfleet.io"]
}
```

### Report Executions
- **Endpoint**: `GET /api/v1/admin/reports/:id/executions`
- **Response**: History of report generation runs with S3 URLs.

---

## 9. Data Governance
**Base Path**: `/api/v1/admin/governance`

### Data Requests (GDPR/Compliance)
- **Endpoint**: `POST /api/v1/admin/governance/requests`
- **Request Body**: `{"type": "EXPORT", "subject_id": "user-uuid", "purpose": "User Request"}`
- **Endpoint**: `GET /api/v1/admin/governance/requests`
- **Response**: List of pending/completed data requests.

### Governance Settings
- **Endpoint**: `GET /api/v1/admin/governance/settings`
- **Response**: Data retention policies and classification rules.

---

## 10. Disaster Recovery (DR)
**Base Path**: `/api/v1/admin/dr`

### Backups
- **Endpoint**: `GET /api/v1/admin/dr/backups`
- **Response**: List of available system backups.
- **Endpoint**: `POST /api/v1/admin/dr/backups`
- **Request Body**: `{"type": "INCREMENTAL"}`

### Restore Operations
- **Endpoint**: `POST /api/v1/admin/dr/restore`
- **Request Body**: `{"backup_id": "uuid", "target": "staging"}`

---

## 11. System Monitoring & Observability
**Base Path**: `/v1/admin/monitoring`

### System Health
- **Endpoint**: `GET /api/v1/admin/monitoring/health`
- **Response**: Operational status and critical errors.

### Metrics Hub
- **Endpoint**: `GET /api/v1/admin/monitoring/metrics`
- **Response**: CPU/Memory usage per service, DB latency.

---

## 12. Global Search & Timeline
**Base Path**: `/api/search`, `/api/timeline`

### Global Search
- **Endpoint**: `GET /api/search?q=query`
- **Response**: Unified results for Vehicles, Drivers, Trips, Alerts.

### Activity Timeline
- **Endpoint**: `GET /api/timeline`
- **Response**: Chronological feed of system events.

---

## 13. Security & Authentication
**Base Path**: `/api/security`, `/api/v1/auth`

### Security Overview
- **Endpoint**: `GET /api/security/overview`
- **Response**: MFA status, Active Sessions, Compliance score.

### Login Activity & Audits
- **Endpoint**: `GET /api/security/logins`
- **Response**: List of recent login attempts.

---

## 14. Customer Support Command Center (BFF)
**Base Path**: `/bff/support`

### Organization Overview
- **Endpoint**: `GET /bff/support/orgs/{id}/summary`
- **Response**: High-level health of a specific organization.

### Incident Management
- **Endpoint**: `GET /bff/support/incidents/feed`
- **Response**: Live feed of support tickets and system alerts.

---

## 15. Billing & Finance
**Base Path**: `/v1/admin/billing`

### Invoices
- **Endpoint**: `GET /api/v1/admin/billing/invoices`
- **Response**: List of invoices with payment status.

### Settings
- **Endpoint**: `GET /api/v1/admin/billing-settings`
- **Response**: Tax rates, Currency config, Auto-invoicing toggles.

---

## 16. Video Telematics
**Base Path**: `/api/v1/video`

### Live Streaming
- **Endpoint**: `POST /api/v1/video/streams`
- **Request Body**: `{"camera_id": "CAM01", "vehicle_id": "uuid"}`
- **Endpoint**: `GET /api/v1/video/streams`
- **Response**: List of active/inactive streams.

### Recordings (DVR)
- **Endpoint**: `GET /api/v1/video/recordings`
- **Query Params**: `vehicle_id`, `start_time`, `end_time`
- **Response**: List of available clips with S3 URLs.

### AI Analytics
- **Endpoint**: `GET /api/v1/video/analytics/recording/:id`
- **Response**: `{"objects_detected": ["person", "car"], "confidence": 0.95}`

### Retention Policies
- **Endpoint**: `POST /api/v1/video/retention/policies`
- **Request Body**: `{"retention_days": 30, "auto_delete": true}`
- **Endpoint**: `GET /api/v1/video/retention/policies`

---

## 17. Route Optimization
**Base Path**: `/api/v1/routes`

### Optimization Requests
- **Endpoint**: `POST /api/v1/routes/requests`
- **Request Body**:
```json
{
  "name": "Morning Delivery Run",
  "stops": [{"lat": 12.97, "lng": 77.59}, {"lat": 12.90, "lng": 77.65}],
  "vehicles": ["uuid1", "uuid2"],
  "constraints": {"time_window_start": "08:00", "time_window_end": "12:00"}
}
```

### Optimization Results
- **Endpoint**: `GET /api/v1/routes/results/:id`
- **Response**:
```json
{
  "total_distance_km": 45.2,
  "routes": [
    {"vehicle_id": "uuid1", "path": [...]}
  ]
}
```

### Background Jobs
- **Endpoint**: `GET /api/v1/routes/jobs`
- **Response**: Status of long-running optimization tasks.

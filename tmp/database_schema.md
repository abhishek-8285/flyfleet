# FlyFleet Database Schema Reference

This document provides a comprehensive reference for the FlyFleet PostgreSQL database schema.
All tables use `UUID` for primary keys and include `created_at` and `updated_at` timestamps.

## 1. Core Identity & Auth
### `organizations`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK, Default: `gen_random_uuid()` | Unique Organization ID |
| `name` | VARCHAR(255) | NOT NULL | Organization Name |
| `domain` | VARCHAR(255) | UNIQUE | Corporate domain for auto-join |
| `status` | VARCHAR(50) | Default: `active` | Organization status |

### `users`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Unique User ID |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | User Email |
| `password_hash` | TEXT | NOT NULL | Bcrypt hash |
| `name` | VARCHAR(255) | | Full Name |
| `mfa_enabled` | BOOLEAN | Default: `FALSE` | MFA Status |
| `status` | VARCHAR(50) | Default: `active` | User status |

### `user_organizations`
Junction table for Multi-tenancy.
| Column | Type | Constraints | Description |
|---|---|---|---|
| `user_id` | UUID | FK -> `users(id)` | User |
| `org_id` | UUID | FK -> `organizations(id)` | Organization |
| `role` | VARCHAR(50) | Default: `member` | Role (admin, member, driver) |

---

## 2. Fleet Assets
### `vehicles`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Vehicle ID |
| `org_id` | UUID | FK -> `organizations(id)` | Owner Organization |
| `license_plate` | VARCHAR(20) | UNIQUE, NOT NULL | Registration Number |
| `vin` | VARCHAR(50) | UNIQUE | Vehicle Identification Number |
| `make` | VARCHAR(50) | NOT NULL | e.g. Toyota |
| `model` | VARCHAR(50) | NOT NULL | e.g. Corolla |
| `year` | INTEGER | NOT NULL | Manufacturing Year |
| `status` | VARCHAR(20) | Default: `active` | active, maintenance, retired |
| `fuel_type` | VARCHAR(20) | | petrol, diesel, electric |
| `odometer` | BIGINT | Default: 0 | Current Odometer reading |

### `drivers`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Driver ID |
| `org_id` | UUID | FK -> `organizations(id)` | Helper Organization |
| `full_name` | VARCHAR(100) | Generated | First + Last Name |
| `license_number` | VARCHAR(50) | NOT NULL | Driving License |
| `status` | VARCHAR(20) | Default: `active` | active, suspended, inactive |
| `current_vehicle_id` | UUID | FK -> `vehicles(id)` | Currently assigned vehicle |
| `rating` | DOUBLE | Default: 5.0 | Driver Score |

---

## 3. Operations
### `trips`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Trip ID |
| `org_id` | UUID | FK -> `organizations(id)` | |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `driver_id` | UUID | FK -> `drivers(id)` | |
| `status` | VARCHAR(50) | Default: `planned` | planned, in_progress, completed |
| `start_location` | JSONB | | {lat, lng, address} |
| `end_location` | JSONB | | {lat, lng, address} |
| `waypoints` | JSONB | | Array of stops |
| `fuel_consumed_liters` | DECIMAL | | |

### `fuel_logs`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Log ID |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `driver_id` | UUID | FK -> `drivers(id)` | |
| `quantity` | FLOAT | NOT NULL | Liters/Gallons |
| `unit_price` | FLOAT | NOT NULL | Cost per unit |
| `total_cost` | FLOAT | NOT NULL | Total transaction cost |
| `odometer` | FLOAT | NOT NULL | Odometer at time of fill-up |
| `status` | VARCHAR | | PENDING, APPROVED, REJECTED |
| `receipt_url` | TEXT | | URL to image |

---

## 4. Maintenance
### `maintenance_schedules`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Schedule ID |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `type` | VARCHAR | | preventive, routine |
| `due_date` | TIMESTAMP | | Date deadline |
| `due_odometer` | BIGINT | | Mileage deadline |
| `status` | VARCHAR | | scheduled, overdue |
| `service_items` | JSONB | | Array of tasks |
| `parts_required` | JSONB | | Array of parts |

### `work_orders`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Work Order ID |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `maintenance_type` | VARCHAR | | repair, inspection |
| `priority` | VARCHAR | | low, medium, high |
| `status` | VARCHAR | | open, in_progress, completed |
| `estimated_cost` | FLOAT | | |
| `actual_cost` | FLOAT | | |

---

## 5. IoT & Diagnostics
### `devices`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Device ID |
| `serial_number` | VARCHAR(100) | UNIQUE, NOT NULL | IMEI or Serial |
| `status` | VARCHAR | Default: `active` | online, offline |
| `battery_level` | INTEGER | | 0-100 |
| `last_heartbeat` | TIMESTAMP | | Last seen time |

### `device_telemetry`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `device_id` | UUID | FK -> `devices(id)` | |
| `payload` | JSONB | NOT NULL | Raw sensor data |
| `timestamp` | TIMESTAMP | | Event time |

### `geofences`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `org_id` | UUID | FK -> `organizations(id)` | |
| `type` | VARCHAR(20) | | circle, polygon |
| `center` | JSONB | | {lat, lng} for circle |
| `coordinates` | JSONB | | Points for polygon |
| `vehicles` | JSONB | | Restricted vehicle IDs |

---

## 6. Advanced Tracking & Analytics
### `telemetry_analytics`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `analysis_type` | VARCHAR | | speed_analysis, fuel_efficiency |
| `data` | JSONB | | Analytical results |
| `confidence` | DECIMAL | | 0.0 - 1.0 |
| `timestamp` | TIMESTAMP | | |

### `telemetry_anomalies`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `anomaly_type` | VARCHAR | | unusual_speed, fuel_theft |
| `data` | JSONB | | Anomaly details |
| `anomaly_score` | DECIMAL | | Severity score |
| `recommendation` | TEXT | | Actionable advice |

---

## 7. Integrations & Webhooks
### `integrations`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `type` | VARCHAR(50) | | SLACK, SAP, ETC |
| `provider` | VARCHAR | | Twilio, AWS, SAP |
| `config` | JSONB | | Encrypted credentials |
| `status` | VARCHAR(20) | | active, error |

### `api_clients`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `client_id` | VARCHAR | UNIQUE | |
| `client_secret_hash` | TEXT | | |
| `scopes` | TEXT[] | | Assigned permissions |
| `status` | VARCHAR | | active, revoked |

### `webhooks`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `url` | TEXT | NOT NULL | Target URL |
| `events` | TEXT[] | | Subscribed events |
| `secret` | VARCHAR | | Signing secret |
| `status` | VARCHAR | | active, failing |

---

## 8. Security & Authentication
### `mfa_policies`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Policy ID |
| `org_id` | UUID | FK -> `organizations(id)` | |
| `require_mfa` | BOOLEAN | Default: FALSE | Enforce MFA |
| `allowed_methods` | JSONB | | ["totp", "sms"] |
| `max_attempts` | INTEGER | Default: 5 | |

### `mfa_events`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Event ID |
| `user_id` | UUID | FK -> `users(id)` | |
| `event_type` | VARCHAR | | CHALLENGE, VERIFY |
| `status` | VARCHAR | | success, failure |
| `method` | VARCHAR | | totp, sms |

### `audit_settings`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `org_id` | UUID | FK -> `organizations(id)` | |
| `retention_days_online` | INTEGER | | |
| `tracked_event_types` | JSONB | | List of events to log |

---

## 9. Billing & Finance
### `invoices`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Invoice ID |
| `org_id` | UUID | | |
| `customer_name` | TEXT | | |
| `invoice_date` | DATE | | |
| `due_date` | DATE | | |
| `status` | TEXT | Default: PENDING | |
| `total_amount` | DECIMAL | | |
| `balance_due` | DECIMAL | | |

### `invoice_items`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Item ID |
| `invoice_id` | UUID | FK -> `invoices(id)` | |
| `description` | TEXT | | |
| `quantity` | INTEGER | | |
| `unit_price` | DECIMAL | | |
| `total` | DECIMAL | | |

### `billing_settings`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `org_id` | UUID | PK | |
| `default_tax_rate` | DECIMAL | | |
| `base_currency` | TEXT | Default: USD | |
| `auto_invoicing_enabled` | BOOLEAN | | |

---

## 10. Customer Support
### `support_tickets`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `ticket_id` | VARCHAR | UNIQUE | Human ID (INC-123) |
| `org_id` | UUID | FK -> `organizations(id)` | |
| `title` | VARCHAR | | |
| `status` | VARCHAR | | OPEN, RESOLVED |
| `priority` | VARCHAR | | LOW, CRITICAL |
| `assigned_to` | UUID | FK -> `users(id)` | Agent |
| `category` | VARCHAR | | Technical, Billing |

### `incident_comments`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `ticket_id` | UUID | FK -> `support_tickets(id)` | |
| `content` | TEXT | | |
| `comment_type` | VARCHAR | | INTERNAL, EXTERNAL |
| `is_private` | BOOLEAN | | |

### `user_sessions`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `user_id` | UUID | FK -> `users(id)` | |
| `session_id` | VARCHAR | UNIQUE | |
| `ip_address` | VARCHAR | | |
| `is_active` | BOOLEAN | | |
| `device_info` | JSONB | | UA, OS, Browser |

### `org_support_metrics`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `org_id` | UUID | FK -> `organizations(id)` | |
| `open_tickets` | INTEGER | | |
| `sla_breaches` | INTEGER | | |
| `avg_resolution_time` | INTERVAL | | |

---

## 11. Reports Module
### `scheduled_reports`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `organization_id` | UUID | | |
| `name` | VARCHAR | | Report Name |
| `report_type` | VARCHAR | | VEHICLE_USAGE, FUEL |
| `frequency` | VARCHAR | | DAILY, WEEKLY |
| `cron_expression` | VARCHAR | | |
| `delivery_time` | TIME | | |
| `status` | VARCHAR | | ACTIVE, PAUSED |

### `report_executions`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `report_id` | UUID | FK -> `scheduled_reports(id)` | |
| `started_at` | TIMESTAMP | | |
| `finished_at` | TIMESTAMP | | |
| `status` | VARCHAR | | SUCCESS, FAILED |
| `file_url` | TEXT | | S3 URL |
| `error_message` | TEXT | | |

---

## 12. Data Governance
### `governance_requests`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `organization_id` | UUID | | |
| `requester_id` | UUID | | |
| `request_type` | VARCHAR | | EXPORT, DELETION |
| `data_scope` | TEXT | | |
| `purpose` | TEXT | | |
| `status` | VARCHAR | | PENDING, APPROVED |
| `expiry_date` | TIMESTAMP | | |

---

## 13. Disaster Recovery (DR)
### `backups`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `backup_type` | VARCHAR | | FULL, INCREMENTAL |
| `status` | VARCHAR | | SCHEDULED, COMPLETED |
| `start_time` | TIMESTAMP | | |
| `size_bytes` | BIGINT | | |
| `location` | VARCHAR | | |

---

## 14. Advanced Analytics
### `predictive_models`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | VARCHAR | PK | Model ID |
| `name` | VARCHAR | | e.g. "Fuel Usage Predictor" |
| `model_type` | VARCHAR | | regression, classification |
| `algorithm` | VARCHAR | | random_forest, xgboost |
| `accuracy` | DECIMAL | | 0.0 - 1.0 |
| `status` | VARCHAR | | active, training |

### `fleet_analytics`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | VARCHAR | PK | |
| `organization_id` | UUID | | |
| `average_fuel_efficiency` | DECIMAL | | km/l |
| `idle_time_percentage` | DECIMAL | | % |
| `safety_score` | DECIMAL | | 0-100 |
| `maintenance_cost` | DECIMAL | | |

---

## 15. Route Optimization
### `route_optimization_requests`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Request ID |
| `organization_id` | UUID | FK -> `organizations(id)` | |
| `status` | VARCHAR | | pending, completed |
| `constraints` | JSONB | | Time windows, vehicle types |
| `stops` | JSONB | | Array of locations |

### `route_optimization_results`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | Result ID |
| `request_id` | UUID | FK -> `requests(id)` | |
| `routes` | JSONB | | Optimized paths |
| `total_distance` | DECIMAL | | Total km |
| `total_cost` | DECIMAL | | Estimated cost |

---

## 16. Video Telematics
### `recordings`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PK | |
| `vehicle_id` | UUID | FK -> `vehicles(id)` | |
| `trip_id` | UUID | FK -> `trips(id)` | |
| `s3_url` | TEXT | | |
| `event_type` | VARCHAR | | CRASH, HARSH_BRAKING |
| `duration_seconds` | INTEGER | | |
| `timestamp` | TIMESTAMP | | |

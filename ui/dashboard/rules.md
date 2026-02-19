Blunt answer. Based on the screens shown, these are the required APIs and exact data columns.

---

## 1) MAIN DASHBOARD OVERVIEW SCREEN

### API: GET /dashboard/summary

**Required response columns**

* live_vehicle_count
* active_trip_count
* active_alert_count
* critical_alert_count
* warning_alert_count
* last_updated
* organization_id

---

### API: GET /dashboard/live-map

**Columns**

* vehicle_id
* latitude
* longitude
* speed
* heading
* status
* driver_name
* last_update_time

---

### API: GET /dashboard/critical-vehicle-status

**Columns**

* vehicle_id
* driver_name
* status
* location
* last_update
* sla_aging
* alert_type

---

### API: GET /dashboard/recent-activity

**Columns**

* activity_id
* type
* message
* timestamp
* severity
* vehicle_id

---

### API: GET /notifications

**Columns**

* notification_id
* title
* message
* type
* is_read
* created_at

---

## 2) ADD NEW VEHICLE SCREEN

### API: POST /vehicles

**Request columns**

* vin_number
* vehicle_number
* license_plate
* vehicle_type
* model
* year
* telematics_device_id
* organization_id
* assigned_driver_id

**Table: vehicles**

Columns required:

* vehicle_id
* vin_number
* vehicle_number
* license_plate
* vehicle_type
* model
* year
* status
* telematics_device_id
* organization_id
* created_at

---

## 3) VEHICLE DETAIL SCREEN

### API: GET /vehicles/{id}

**Columns**

* vehicle_id
* vehicle_number
* model
* year
* status
* last_latitude
* last_longitude
* speed
* fuel_level
* engine_status
* last_update_time
* assigned_driver

---

### API: GET /vehicles/{id}/alerts

**Columns**

* alert_id
* vehicle_id
* alert_type
* severity
* message
* created_at
* resolved_at

---

### API: GET /vehicles/{id}/telemetry

**Columns**

* latitude
* longitude
* speed
* fuel_level
* engine_temp
* timestamp

---

## 4) DRIVER MANAGEMENT SCREEN

### API: GET /drivers

**Columns**

* driver_id
* name
* license_number
* phone
* status
* assigned_vehicle
* last_trip
* organization_id

---

### API: POST /drivers

**Columns**

* name
* license_number
* phone
* email
* organization_id

---

## 5) MAINTENANCE SCREEN

### API: GET /maintenance/overview

**Columns**

* total_maintenance_due
* critical_issues
* fleet_health_percent

---

### API: GET /maintenance/logs

**Columns**

* maintenance_id
* vehicle_id
* type
* status
* scheduled_date
* completed_date
* cost
* description

---

## CORE DATABASE TABLES REQUIRED

### vehicles table

* vehicle_id
* vin_number
* vehicle_number
* license_plate
* vehicle_type
* model
* year
* status
* last_latitude
* last_longitude
* speed
* fuel_level
* engine_status
* telematics_device_id
* organization_id

---

### drivers table

* driver_id
* name
* license_number
* phone
* email
* status
* assigned_vehicle_id
* organization_id

---

### trips table

* trip_id
* vehicle_id
* driver_id
* start_location
* end_location
* start_time
* end_time
* status

---

### alerts table

* alert_id
* vehicle_id
* type
* severity
* message
* created_at
* resolved_at

---

### telemetry table

* telemetry_id
* vehicle_id
* latitude
* longitude
* speed
* fuel_level
* engine_temp
* timestamp

---

### maintenance table

* maintenance_id
* vehicle_id
* type
* status
* scheduled_date
* completed_date
* cost

---

### notifications table

* notification_id
* user_id
* title
* message
* is_read
* created_at

---

### organizations table

* organization_id
* name
* timezone

---

### users table

* user_id
* name
* email
* role
* organization_id

---

These APIs and columns fully cover every element visible in the provided screens.

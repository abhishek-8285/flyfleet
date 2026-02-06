

---

# PROPER MICRO-SERVICE BOUNDARIES

First rule:

**One Service = One Domain = One Database Owner**

---

## NEW SERVICE MODEL (CLEAN VERSION)

You should have exactly these core services:

1. API Gateway
2. Auth & Identity
3. Vehicle
4. Driver
5. Tracking
6. Trip
7. Geofence
8. Alert & Notification
9. Maintenance
10. Route Optimization
11. Analytics
12. Integration (India)
13. Video
14. Platform Config

---

# CENTRAL RULE

### Every data entity has ONE OWNER

| Data      | Owner Service       |
| --------- | ------------------- |
| Vehicles  | Vehicle Service     |
| Drivers   | Driver Service      |
| Telemetry | Tracking Service    |
| Trips     | Trip Service        |
| Geofences | Geofence Service    |
| Alerts    | Alert Service       |
| Reports   | Analytics Service   |
| FASTag    | Integration Service |
| Video     | Video Service       |

---

## ARCHITECTURE FLOW

```
Client → API Gateway → Services
```

No direct access to services.

---

# FIXED SERVICE RESPONSIBILITIES

---

## 1. API GATEWAY (NEW REQUIRED)

### ONLY ENTRY POINT

Responsibilities:

* Authentication
* Rate limiting
* Routing
* Aggregation
* Logging
* Versioning

---

### Gateway Endpoints:

```
/auth/*
/vehicles/*
/drivers/*
/trips/*
/tracking/*
/analytics/*
```

Gateway talks to internal services.

---

# 2. AUTH SERVICE – PURE IDENTITY

ONLY THESE APIs:

```
POST /auth/register
POST /auth/login
POST /auth/refresh
GET  /auth/profile
PUT  /auth/profile
POST /auth/roles
GET  /auth/permissions
```

---

### Remove from Auth:

* Vehicle logic
* Driver logic
* Business rules

Auth = ONLY identity.

---

# 3. VEHICLE SERVICE – MASTER DATA ONLY

Vehicle service should ONLY handle:

* Vehicle CRUD
* Assign driver
* Vehicle metadata

---

### VALID ENDPOINTS

```
POST   /vehicles
GET    /vehicles
GET    /vehicles/:id
PUT    /vehicles/:id
DELETE /vehicles/:id

POST   /vehicles/:id/assign-driver
POST   /vehicles/:id/unassign-driver
GET    /vehicles/search
```

---

### REMOVE FROM VEHICLE SERVICE

These must be deleted:

```
/vehicles/:id/analytics
/vehicles/:id/location
/vehicles/:id/fuel-level
```

Reason:

Not vehicle domain.

---

# 4. DRIVER SERVICE – PEOPLE DOMAIN

Valid APIs:

```
POST   /drivers
GET    /drivers
GET    /drivers/:id
PUT    /drivers/:id
DELETE /drivers/:id

POST   /drivers/:id/assign-vehicle
POST   /drivers/:id/unassign-vehicle
POST   /drivers/:id/documents
GET    /drivers/:id/documents
```

---

### REMOVE

```
/drivers/:id/analytics
/drivers/:id/behavior-score
```

Those belong to Analytics.

---

# 5. TRACKING SERVICE – ONLY TELEMETRY

This service must be PURELY:

* Location ingestion
* Device data
* Real-time streaming

---

### VALID ENDPOINTS

```
POST /telemetry/gps
POST /telemetry/obd
POST /telemetry/batch
GET  /tracking/history
GET  /tracking/live
```

---

### REMOVE

Any analytics from tracking service.

---

### OWNERSHIP

Tracking Service OWNS:

* GPS history
* Real-time location

---

# 6. TRIP SERVICE – BUSINESS WORKFLOW

Trip service must only handle:

Trip lifecycle.

---

### Correct APIs

```
POST /trips
GET  /trips
GET  /trips/:id
PATCH /trips/:id
DELETE /trips/:id
```

---

### REMOVE RPC STYLE:

DELETE THESE:

```
/trips/:id/start
/trips/:id/pause
/trips/:id/resume
```

---

### REPLACE WITH:

```
PATCH /trips/:id
{
  "status": "STARTED"
}
```

---

# 7. GEOFENCE SERVICE – SINGLE OWNER

Only this service handles:

* Geofence CRUD
* Boundary checks
* Events

---

### ONLY HERE:

```
POST /geofences
GET  /geofences
PUT  /geofences/:id
DELETE /geofences/:id
```

---

### REMOVE FROM ALERT SERVICE:

All geofence APIs.

---

# 8. ALERT SERVICE – PURE NOTIFICATIONS

Alert service must NOT own geofences.

---

### Responsibilities:

* Receive events
* Create alerts
* Notify users

---

### Valid APIs

```
GET  /alerts
POST /alerts/acknowledge
POST /alerts/resolve
POST /alerts/rules
```

---

### NO CRUD FOR GEOFENCE HERE

---

# 9. MAINTENANCE SERVICE

Valid domain:

* Schedules
* Logs
* Reminders

---

```
POST /maintenance/schedules
GET  /maintenance/schedules
POST /maintenance/logs
GET  /maintenance/upcoming
```

---

# 10. ROUTE OPTIMIZATION

Pure algorithmic service.

---

```
POST /route-optimization/optimize
GET  /route-optimization/results
```

---

### It should NOT store trips.

Just compute routes.

---

# 11. ANALYTICS SERVICE – READ ONLY

Analytics must be:

A consumer of events.

---

### ONLY HERE:

```
GET /analytics/vehicles
GET /analytics/drivers
GET /analytics/trips
GET /analytics/forecast
```

---

### REMOVE ALL ANALYTICS FROM:

* Vehicle service
* Driver service
* Tracking service

---

# 12. INDIA INTEGRATION SERVICE – WORKFLOW ONLY

Change completely.

---

### REMOVE CRUD LIKE:

```
PUT /fastag/transactions/:id
DELETE /fastag/transactions/:id
```

---

### Replace with:

```
POST /fastag/sync
POST /vahan/verify
POST /sarathi/verify
POST /ewaybill/validate
```

---

### Integration service = BUSINESS ACTIONS ONLY

Not database tables.

---

# 13. VIDEO SERVICE

Split responsibilities:

* Streaming
* Storage
* Analytics

---

### Core APIs

```
POST /video/streams
GET  /video/recordings
POST /video/analyze
```

---

# EVENT DRIVEN MODEL (REAL)

---

### Required Events:

```
VehicleCreated
DriverAssigned
TripStarted
LocationUpdated
GeofenceEntered
AlertGenerated
```

---

### Data Flow:

Tracking → Event → Geofence
Geofence → Event → Alert
Trip → Event → Analytics

---

# DATABASE RULES

Each service must have:

* Own database
* No cross joins
* Event sync only

---

# SECURITY MODEL

Add:

* API Gateway JWT validation
* RBAC service
* Tenant isolation

---

# RECOMMENDED FINAL ARCHITECTURE

```
Clients
   ↓
API Gateway
   ↓
Auth Service
Vehicle Service
Driver Service
Tracking Service
Trip Service
Geofence Service
Alert Service
Analytics Service
Integration Service
Video Service
Maintenance Service
```

---

# WHAT TO DELETE FROM CURRENT DESIGN

Immediately remove:

* Geofence CRUD from Alert
* Analytics from Vehicle
* Analytics from Driver
* Duplicate telemetry endpoints
* CRUD APIs in Integration

---



# Comprehensive Public API Reference

This document lists all public API endpoints extracted from the enterprise schema, grouped by their functional category, along with a brief description of their workings.

## Addresses (5 endpoints)
- **`GET`** `/addresses`: List all addresses
- **`POST`** `/addresses`: Create an address
- **`DELETE`** `/addresses/{id}`: Delete an address
- **`GET`** `/addresses/{id}`: Retrieve an address
- **`PATCH`** `/addresses/{id}`: Update an address

## Alerts (5 endpoints)
- **`DELETE`** `/alerts/configurations`: Delete alert configurations.
- **`GET`** `/alerts/configurations`: Get Alert Configurations.
- **`PATCH`** `/alerts/configurations`: Update alert configurations.
- **`POST`** `/alerts/configurations`: Create alert configurations.
- **`GET`** `/alerts/incidents/stream`: Get Alert Incidents.

## Assets (8 endpoints)
- **`DELETE`** `/assets`: Delete an existing asset.
- **`GET`** `/assets`: List all assets.
- **`PATCH`** `/assets`: Update an existing asset.
- **`POST`** `/assets`: Create a new asset.
- **`GET`** `/v1/fleet/assets/locations`: List current location for all assets
- **`GET`** `/v1/fleet/assets/reefers`: List stats for all reefers
- **`GET`** `/v1/fleet/assets/{asset_id}/locations`: List historical locations for a given asset
- **`GET`** `/v1/fleet/assets/{asset_id}/reefer`: List stats for a given reefer

## Attributes (5 endpoints)
- **`GET`** `/attributes`: List all attributes by entity type
- **`POST`** `/attributes`: Create an attribute
- **`DELETE`** `/attributes/{id}`: Deleting an attribute
- **`GET`** `/attributes/{id}`: Retrieve an attribute
- **`PATCH`** `/attributes/{id}`: Update an attribute

## Auth Token for Driver (1 endpoints)
- **`POST`** `/fleet/drivers/auth-token`: Create auth token for a driver

## Beta APIs (49 endpoints)
- **`GET`** `/assets/depreciation`: [beta] Get depreciation transactions.
- **`GET`** `/assets/inputs/stream`: [beta] List asset inputs data in an organization.
- **`GET`** `/beta/aemp/Fleet/{pageNumber}`: [beta] Get a list of AEMP equipment
- **`GET`** `/beta/fleet/drivers/efficiency`: [beta] List driver efficiency
- **`PATCH`** `/beta/fleet/equipment/{id}`: [beta] Update an equipment
- **`GET`** `/beta/fleet/hos/drivers/eld-events`: [beta] Get driver HOS ELD events
- **`GET`** `/beta/fleet/trailers/stats`: [beta] Get trailer stats
- **`GET`** `/beta/fleet/trailers/stats/feed`: [beta] Get trailer stats feed
- **`GET`** `/beta/fleet/trailers/stats/history`: [beta] Get trailer stats history
- **`PATCH`** `/beta/fleet/vehicles/{id}/immobilizer`: [beta] Update engine immobilizer state of a vehicle.
- **`DELETE`** `/beta/industrial/jobs`: [beta] Deletes an existing job
- **`GET`** `/beta/industrial/jobs`: [beta] Fetches all jobs
- **`PATCH`** `/beta/industrial/jobs`: [beta] Patches a job
- **`POST`** `/beta/industrial/jobs`: [beta] Create a job
- **`GET`** `/detections/stream`: [beta] Returns a list of all detections in an organization.
- **`GET`** `/devices`: [beta] List devices.
- **`GET`** `/fleet/vehicles/immobilizer/stream`: [beta] Get engine immobilizer states
- **`POST`** `/functions/{name}/runs`: [beta] Start a Function run.
- **`PATCH`** `/hos/daily-logs/log-meta-data`: [beta] Update the shippingDocs field of an existing assignment.
- **`GET`** `/hub/customProperties`: [beta] List custom properties for a specific hub
- **`POST`** `/hub/plan/orders`: [beta] Create orders in bulk
- **`DELETE`** `/qualification-records`: [beta] Delete a qualification record.
- **`GET`** `/qualification-records`: [beta] Get a list of specified qualification records.
- **`PATCH`** `/qualification-records`: [beta] Update a qualification record.
- **`POST`** `/qualification-records`: [beta] Create a qualification record.
- **`POST`** `/qualification-records/archive`: [beta] Archive a qualification record.
- **`GET`** `/qualification-records/stream`: [beta] Get a stream of filtered qualification records.
- **`POST`** `/qualification-records/unarchive`: [beta] Unarchive a qualification record.
- **`GET`** `/qualification-types`: [beta] Get a list of qualification types.
- **`POST`** `/readings`: [beta] Ingest Readings
- **`GET`** `/readings/definitions`: [beta] Get Readings Definitions
- **`GET`** `/readings/history`: [beta] Get Readings History and Feed
- **`GET`** `/readings/latest`: [beta] Get Readings Snapshot
- **`GET`** `/reports/configs`: [beta] Get custom report configs
- **`GET`** `/reports/datasets`: [beta] Get custom report datasets
- **`GET`** `/reports/runs`: [beta] Get custom report runs
- **`POST`** `/reports/runs`: [beta] Create a custom report run
- **`GET`** `/reports/runs/data`: [beta] Get custom report run data
- **`DELETE`** `/ridership/accounts`: [beta] Delete a ridership account.
- **`GET`** `/ridership/accounts`: [beta] List ridership accounts.
- **`POST`** `/ridership/accounts`: [beta] Create a ridership account.
- **`PUT`** `/ridership/accounts`: [beta] Update a ridership account.
- **`GET`** `/ridership/accounts/{id}`: [beta] Get a ridership account.
- **`GET`** `/safety-scores/drivers`: [beta] Get driver scores
- **`GET`** `/safety-scores/drivers/trips`: [beta] Get driver scores trip breakdown
- **`GET`** `/safety-scores/tag-group`: [beta] Get tags combined score
- **`GET`** `/safety-scores/tags`: [beta] Get tag scores
- **`GET`** `/safety-scores/vehicles`: [beta] Get vehicle scores
- **`GET`** `/safety-scores/vehicles/trips`: [beta] Get vehicle scores trip breakdown

## Carrier Proposed Assignments (3 endpoints)
- **`GET`** `/fleet/carrier-proposed-assignments`: Retrieve assignments
- **`POST`** `/fleet/carrier-proposed-assignments`: Create an assignment
- **`DELETE`** `/fleet/carrier-proposed-assignments/{id}`: Delete an assignment

## Coaching (3 endpoints)
- **`GET`** `/coaching/driver-coach-assignments`: Get driver coach assignments.
- **`PUT`** `/coaching/driver-coach-assignments`: Put driver coach assignments.
- **`GET`** `/coaching/sessions/stream`: Get coaching sessions.

## Contacts (5 endpoints)
- **`GET`** `/contacts`: List all contacts
- **`POST`** `/contacts`: Create a contact
- **`DELETE`** `/contacts/{id}`: Delete a contact
- **`GET`** `/contacts/{id}`: Retrieve a contact
- **`PATCH`** `/contacts/{id}`: Update a contact

## Documents (7 endpoints)
- **`GET`** `/fleet/document-types`: Fetch document types
- **`GET`** `/fleet/documents`: Fetch all documents
- **`POST`** `/fleet/documents`: Create document
- **`POST`** `/fleet/documents/pdfs`: Create a document PDF
- **`GET`** `/fleet/documents/pdfs/{id}`: Query a document PDF
- **`DELETE`** `/fleet/documents/{id}`: Delete document
- **`GET`** `/fleet/documents/{id}`: Fetch document

## Driver QR Codes (3 endpoints)
- **`DELETE`** `/drivers/qr-codes`: Revoke driver's QR code
- **`GET`** `/drivers/qr-codes`: Get driver QR codes
- **`POST`** `/drivers/qr-codes`: Create new QR code for driver

## Driver-Trailer Assignments (3 endpoints)
- **`GET`** `/driver-trailer-assignments`: Get currently active driver-trailer assignments for driver.
- **`PATCH`** `/driver-trailer-assignments`: Update an existing driver-trailer assignment.
- **`POST`** `/driver-trailer-assignments`: Create a new driver-trailer assignment

## Driver-Vehicle Assignments (4 endpoints)
- **`DELETE`** `/fleet/driver-vehicle-assignments`: Delete API generated driver-vehicle assignments
- **`GET`** `/fleet/driver-vehicle-assignments`: Get all driver-vehicle assignments
- **`PATCH`** `/fleet/driver-vehicle-assignments`: Update API generated driver-vehicle assignments
- **`POST`** `/fleet/driver-vehicle-assignments`: Create a new driver-vehicle assignment

## Drivers (5 endpoints)
- **`GET`** `/fleet/drivers`: List all drivers
- **`POST`** `/fleet/drivers`: Create a driver
- **`POST`** `/fleet/drivers/remote-sign-out`: Sign out a driver
- **`GET`** `/fleet/drivers/{id}`: Retrieve a driver
- **`PATCH`** `/fleet/drivers/{id}`: Update a driver

## Equipment (8 endpoints)
- **`GET`** `/fleet/equipment`: List all equipment
- **`GET`** `/fleet/equipment/locations`: Get most recent locations for all equipment
- **`GET`** `/fleet/equipment/locations/feed`: Follow feed of equipment locations
- **`GET`** `/fleet/equipment/locations/history`: Get historical equipment locations
- **`GET`** `/fleet/equipment/stats`: Get most recent stats for all equipment
- **`GET`** `/fleet/equipment/stats/feed`: Follow a feed of equipment stats
- **`GET`** `/fleet/equipment/stats/history`: Get historical equipment stats
- **`GET`** `/fleet/equipment/{id}`: Retrieve a unit of equipment

## Forms (7 endpoints)
- **`GET`** `/form-submissions`: Get a list of specified form submissions.
- **`PATCH`** `/form-submissions`: Update a single form submission.
- **`POST`** `/form-submissions`: Create a form submission.
- **`GET`** `/form-submissions/pdf-exports`: Return a PDF export that has already been generated for a form submission.
- **`POST`** `/form-submissions/pdf-exports`: Create a PDF export for an existing form submission.
- **`GET`** `/form-submissions/stream`: Get a stream of filtered form submissions.
- **`GET`** `/form-templates`: Get a list of form templates.

## Fuel and Energy (5 endpoints)
- **`GET`** `/driver-efficiency/drivers`: Get Driver efficiency data grouped by drivers.
- **`GET`** `/driver-efficiency/vehicles`: Get Driver efficiency data grouped by vehicles.
- **`GET`** `/fleet/reports/drivers/fuel-energy`: Get fuel and energy efficiency driver reports.
- **`GET`** `/fleet/reports/vehicles/fuel-energy`: Get fuel and energy efficiency vehicle reports.
- **`POST`** `/fuel-purchase`: Create a fuel purchase transaction.

## Gateways (3 endpoints)
- **`GET`** `/gateways`: List all gateways
- **`POST`** `/gateways`: Activate a new gateway
- **`DELETE`** `/gateways/{id}`: Deactivate a gateway

## Hours of Service (6 endpoints)
- **`GET`** `/fleet/hos/clocks`: Get HOS clocks
- **`GET`** `/fleet/hos/daily-logs`: Get all driver HOS daily logs
- **`GET`** `/fleet/hos/logs`: Get HOS logs
- **`GET`** `/fleet/hos/violations`: Get all driver HOS violations
- **`POST`** `/v1/fleet/drivers/{driver_id}/hos/duty_status`: Set a duty status for a specific driver
- **`GET`** `/v1/fleet/hos_authentication_logs`: Get HOS signin and signout

## Hubs (6 endpoints)
- **`GET`** `/hub/capacities`: List capacities for a specific hub
- **`PATCH`** `/hub/location/{id}`: Update a location
- **`GET`** `/hub/locations`: List locations for a specific hub
- **`POST`** `/hub/locations`: Create locations in bulk
- **`GET`** `/hub/skills`: List skills for a specific hub
- **`GET`** `/hubs`: List all hubs for the organization

## IFTA (4 endpoints)
- **`GET`** `/fleet/reports/ifta/jurisdiction`: Get IFTA jurisdiction reports.
- **`GET`** `/fleet/reports/ifta/vehicle`: Get IFTA vehicle reports.
- **`POST`** `/ifta-detail/csv`: Create a job to generate csv files of IFTA mileage segments.
- **`GET`** `/ifta-detail/csv/{id}`: Get information about an existing IFTA detail job.

## Idling (1 endpoints)
- **`GET`** `/idling/events`: Get idling events.

## Industrial (17 endpoints)
- **`GET`** `/industrial/assets`: List all assets
- **`POST`** `/industrial/assets`: Create an asset
- **`DELETE`** `/industrial/assets/{id}`: Delete an existing asset
- **`PATCH`** `/industrial/assets/{id}`: Update an asset
- **`PATCH`** `/industrial/assets/{id}/data-outputs`: Writes to data outputs on an asset
- **`GET`** `/industrial/data-inputs`: List all data inputs
- **`GET`** `/industrial/data-inputs/data-points`: List most recent data points for data inputs
- **`GET`** `/industrial/data-inputs/data-points/feed`: Follow a real-time feed of data points for data inputs
- **`GET`** `/industrial/data-inputs/data-points/history`: List historical data points for data inputs
- **`GET`** `/v1/industrial/vision/cameras`: Fetch industrial cameras
- **`GET`** `/v1/industrial/vision/cameras/{camera_id}/programs`: Fetch industrial camera programs
- **`GET`** `/v1/industrial/vision/run/camera/{camera_id}`: Fetch the latest run for a camera or program
- **`GET`** `/v1/industrial/vision/runs`: Fetch runs
- **`GET`** `/v1/industrial/vision/runs/{camera_id}`: Fetch runs by camera
- **`GET`** `/v1/industrial/vision/runs/{camera_id}/{program_id}/{started_at_ms}`: Fetch runs by camera and program
- **`POST`** `/v1/machines/history`: Get machine history
- **`POST`** `/v1/machines/list`: Get machines

## Issues (3 endpoints)
- **`GET`** `/issues`: Get a list of specified issues.
- **`PATCH`** `/issues`: Update a single issue.
- **`GET`** `/issues/stream`: Get a stream of filtered issues.

## Legacy (1 endpoints)
- **`GET`** `/v1/fleet/assets`: [Legacy] List all assets

## Legacy APIs (8 endpoints)
- **`GET`** `/fleet/defects/history`: [legacy] Get all defects
- **`GET`** `/fleet/drivers/vehicle-assignments`: [legacy] Get all vehicles assigned to a set of drivers
- **`GET`** `/fleet/dvirs/history`: [legacy] Get all DVIRs
- **`GET`** `/fleet/reports/vehicle/idling`: [legacy] Get vehicle idling reports.
- **`GET`** `/fleet/safety-events`: [legacy] List all safety events.
- **`GET`** `/fleet/safety-events/audit-logs/feed`: [legacy] Fetches safety activity event feed
- **`GET`** `/fleet/vehicles/driver-assignments`: [legacy] Get all drivers assigned to a set of vehicles
- **`GET`** `/v1/fleet/vehicles/{vehicleId}/safety/harsh_event`: [legacy] Fetch harsh events

## Live Sharing Links (4 endpoints)
- **`DELETE`** `/live-shares`: Delete non-expired Live Sharing Link
- **`GET`** `/live-shares`: Get Live Sharing Links
- **`PATCH`** `/live-shares`: Update non-expired Live Sharing Link
- **`POST`** `/live-shares`: Create Live Sharing Link

## Location and Speed (1 endpoints)
- **`GET`** `/assets/location-and-speed/stream`: List asset location and speed data in an organization.

## Maintenance (9 endpoints)
- **`GET`** `/defect-types`: Get DVIR defect types.
- **`GET`** `/defects/stream`: Stream DVIR defects.
- **`GET`** `/defects/{id}`: Get a single DVIR defect by ID.
- **`GET`** `/dvirs/stream`: Stream DVIRs
- **`GET`** `/dvirs/{id}`: Get a single DVIR by ID.
- **`PATCH`** `/fleet/defects/{id}`: Update a defect
- **`POST`** `/fleet/dvirs`: Create a mechanic DVIR
- **`PATCH`** `/fleet/dvirs/{id}`: Resolve a DVIR
- **`GET`** `/v1/fleet/maintenance/list`: Get vehicles with engine faults or check lights

## Media (3 endpoints)
- **`GET`** `/cameras/media`: List uploaded media by time range.
- **`GET`** `/cameras/media/retrieval`: Get details for a media retrieval request
- **`POST`** `/cameras/media/retrieval`: Create a media retrieval request

## Messages (2 endpoints)
- **`GET`** `/v1/fleet/messages`: Get all messages.
- **`POST`** `/v1/fleet/messages`: Send a message to a list of driver ids.

## Organization Info (1 endpoints)
- **`GET`** `/me`: Get information about your organization

## Plans (2 endpoints)
- **`POST`** `/hub/plan`: Create a new plan
- **`GET`** `/hub/plans`: List plans for a specific hub

## Preview APIs (3 endpoints)
- **`POST`** `/preview/fleet/drivers/create-auth-token`: [preview] Create auth token for a driver
- **`DELETE`** `/preview/fleet/vehicles/{id}/lock`: [preview] Unlock a vehicle.
- **`PUT`** `/preview/fleet/vehicles/{id}/lock`: [preview] Lock a vehicle.

## Route Events (1 endpoints)
- **`GET`** `/route-events/stream`: Get route events stream

## Routes (8 endpoints)
- **`GET`** `/fleet/routes`: Fetch all routes
- **`POST`** `/fleet/routes`: Create a route
- **`GET`** `/fleet/routes/audit-logs/feed`: Get route updates
- **`DELETE`** `/fleet/routes/{id}`: Delete a route.
- **`GET`** `/fleet/routes/{id}`: Fetch a route
- **`PATCH`** `/fleet/routes/{id}`: Update a route
- **`GET`** `/hub/plan/routes`: List routes for a specific plan
- **`DELETE`** `/v1/fleet/dispatch/routes/{route_id_or_external_id}`: Delete a route

## Safety (4 endpoints)
- **`GET`** `/safety-events`: Get Safety Events
- **`GET`** `/safety-events/stream`: Get Safety Events Stream
- **`GET`** `/v1/fleet/drivers/{driverId}/safety/score`: Fetch driver safety score
- **`GET`** `/v1/fleet/vehicles/{vehicleId}/safety/score`: Fetch vehicle safety scores

## Sensors (6 endpoints)
- **`POST`** `/v1/sensors/cargo`: Get cargo status
- **`POST`** `/v1/sensors/door`: Get door status
- **`POST`** `/v1/sensors/history`: Get sensor history
- **`POST`** `/v1/sensors/humidity`: Get humidity
- **`POST`** `/v1/sensors/list`: Get all sensors
- **`POST`** `/v1/sensors/temperature`: Get temperature

## Settings (5 endpoints)
- **`GET`** `/fleet/settings/compliance`: Get compliance settings
- **`PATCH`** `/fleet/settings/compliance`: Update compliance settings
- **`GET`** `/fleet/settings/driver-app`: Get driver app settings
- **`PATCH`** `/fleet/settings/driver-app`: Update driver app settings
- **`GET`** `/fleet/settings/safety`: Get safety settings

## Speeding Intervals (1 endpoints)
- **`GET`** `/speeding-intervals/stream`: Get Speeding Intervals

## Tachograph (EU Only) (3 endpoints)
- **`GET`** `/fleet/drivers/tachograph-activity/history`: Get driver tachograph activity
- **`GET`** `/fleet/drivers/tachograph-files/history`: Get tachograph driver files
- **`GET`** `/fleet/vehicles/tachograph-files/history`: Get tachograph vehicle files

## Tags (6 endpoints)
- **`GET`** `/tags`: List all tags
- **`POST`** `/tags`: Create a tag
- **`DELETE`** `/tags/{id}`: Delete a tag
- **`GET`** `/tags/{id}`: Retrieve a tag
- **`PATCH`** `/tags/{id}`: Update a tag
- **`PUT`** `/tags/{id}`: Update a tag

## Trailer Assignments (2 endpoints)
- **`GET`** `/v1/fleet/trailers/assignments`: List trailer assignments for all trailers
- **`GET`** `/v1/fleet/trailers/{trailerId}/assignments`: List trailer assignments for a given trailer

## Trailers (5 endpoints)
- **`GET`** `/fleet/trailers`: List all trailers
- **`POST`** `/fleet/trailers`: Creates a new trailer asset
- **`DELETE`** `/fleet/trailers/{id}`: Delete a trailer
- **`GET`** `/fleet/trailers/{id}`: Retrieve a trailer
- **`PATCH`** `/fleet/trailers/{id}`: Update a trailer

## TrainingAssignments (4 endpoints)
- **`DELETE`** `/training-assignments`: Delete training assignments.
- **`PATCH`** `/training-assignments`: Update training assignments.
- **`POST`** `/training-assignments`: Create training assignments.
- **`GET`** `/training-assignments/stream`: Get a stream of filtered training assignments.

## TrainingCourses (1 endpoints)
- **`GET`** `/training-courses`: Get a list of filtered training courses.

## Trips (2 endpoints)
- **`GET`** `/trips/stream`: Get Trips Stream
- **`GET`** `/v1/fleet/trips`: Get vehicle trips

## Users (6 endpoints)
- **`GET`** `/user-roles`: List all user roles
- **`GET`** `/users`: List all users
- **`POST`** `/users`: Create a user
- **`DELETE`** `/users/{id}`: Delete a user
- **`GET`** `/users/{id}`: Retrieve a user
- **`PATCH`** `/users/{id}`: Update a user

## Vehicle Locations (3 endpoints)
- **`GET`** `/fleet/vehicles/locations`: Locations snapshot
- **`GET`** `/fleet/vehicles/locations/feed`: Locations feed
- **`GET`** `/fleet/vehicles/locations/history`: Historical locations

## Vehicle Stats (3 endpoints)
- **`GET`** `/fleet/vehicles/stats`: Stats snapshot
- **`GET`** `/fleet/vehicles/stats/feed`: Stats feed
- **`GET`** `/fleet/vehicles/stats/history`: Historical stats

## Vehicles (3 endpoints)
- **`GET`** `/fleet/vehicles`: List all vehicles.
- **`GET`** `/fleet/vehicles/{id}`: Retrieve a vehicle
- **`PATCH`** `/fleet/vehicles/{id}`: Update a vehicle

## Webhooks (5 endpoints)
- **`GET`** `/webhooks`: List all webhooks belonging to a specific org.
- **`POST`** `/webhooks`: Create a webhook
- **`DELETE`** `/webhooks/{id}`: Delete a webhook with the given ID
- **`GET`** `/webhooks/{id}`: Retrieve a webhook with given ID
- **`PATCH`** `/webhooks/{id}`: Update a specific webhook's information.

## Work Orders (7 endpoints)
- **`POST`** `/maintenance/invoice-scans`: Process an invoice scan.
- **`GET`** `/maintenance/service-tasks`: Gets service tasks.
- **`DELETE`** `/maintenance/work-orders`: Deletes a work order.
- **`GET`** `/maintenance/work-orders`: Gets work orders.
- **`PATCH`** `/maintenance/work-orders`: Updates a work order.
- **`POST`** `/maintenance/work-orders`: Creates a work order.
- **`GET`** `/maintenance/work-orders/stream`: Stream work orders.


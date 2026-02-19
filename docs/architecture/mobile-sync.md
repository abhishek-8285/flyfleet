# Mobile Offline Sync Specification

## Scope
Applies to all mobile clients (driver, customer, operator) for mutation reliability under unstable networks.

## Goals
1. Users can complete core workflows while offline.
2. Operations sync deterministically after reconnect.
3. Conflicts are explicit and resolvable.
4. Sync state is observable to user and backend.

## Data Model
Each local mutation is stored as an operation record:
- `op_id` (UUID)
- `entity_type` (trip, work_order, incident, command)
- `entity_id`
- `action` (create, update, cancel, acknowledge, complete)
- `payload`
- `dedupe_key`
- `created_at`
- `next_retry_at`
- `retry_count`
- `status` (`pending`, `in_flight`, `applied`, `conflict`, `dead_letter`)
- `last_error`

## Sync Loop
1. App writes mutation to local DB and updates UI optimistically.
2. Sync worker picks oldest `pending` op whose `next_retry_at <= now`.
3. Worker sends operation with idempotency key.
4. On success:
   - mark `applied`
   - reconcile local entity from server response
5. On transient failure (network/5xx/429):
   - increment retry count
   - set exponential backoff with jitter
6. On terminal failure (4xx validation/auth):
   - mark `dead_letter`
   - surface user remediation action

## Retry Policy
1. Base delay: 2s
2. Backoff: `min(2^retry_count * base, 10m)`
3. Jitter: +/- 20%
4. Max retries: 8 before `dead_letter`
5. Retry budget reset when app restarts only if operation is still pending

## Conflict Policy
1. Conflict detection keys:
   - entity version/etag
   - server `updated_at`
2. Default strategy: server-wins with local rebase attempt.
3. Domain overrides:
   - trip status transitions: server authority only
   - notes/comments: merge append
   - forms: field-level merge where possible
4. If unresolved:
   - mark `conflict`
   - show conflict resolution UI with both versions and action choices

## Auth and Security
1. Tokens stored in secure storage only.
2. Sync worker pauses on auth failure and requests token refresh.
3. If refresh fails, queue remains durable until re-login.
4. Sensitive payload fields encrypted at rest in local DB where required.

## Observability
1. Emit events:
   - `sync.op.enqueued`
   - `sync.op.retry`
   - `sync.op.applied`
   - `sync.op.conflict`
   - `sync.op.dead_letter`
2. Metrics:
   - queue depth
   - apply latency
   - retry rate
   - conflict rate
3. Include trace/session/org identifiers in sync logs.

## User Experience Rules
1. Show sync indicator states: `Up to date`, `Syncing`, `Action required`.
2. Dead-letter items must be discoverable and retryable.
3. Optimistic updates must clearly show unsynced status when applicable.

## Acceptance Tests
1. Offline create -> reconnect -> applied in order.
2. Multi-op sequence preserves causal order.
3. Duplicate send does not create duplicate server records.
4. Conflict path surfaces resolution UI.
5. Auth expiry pauses and resumes queue after refresh.

## Current Android Driver Runtime Behavior
Implemented in `apps/mobile/customer/android/driver-app/app/src/main/java/com/flyfleet/driver/service/LocationTrackingService.kt`:

1. Background sync flush runs every 30 seconds even without fresh GPS events.
2. Sync upload is single-flight via mutex to avoid overlapping batch submits.
3. Failed uploads use exponential backoff:
   - base delay 5s
   - cap 5m
   - max retries 5
4. HTTP `409/422` responses are treated as conflict-class terminal failures and moved to dead-letter.
5. Records that exceed retry budget are moved to dead-letter and excluded from future upload attempts.
6. Dead-letter IDs and recent dead-letter event log are persisted in app preferences for restart durability.

## Operational Tooling
1. Dead-letter inspector script:
   - `scripts/mobile_dead_letter_inspector.py`
2. Usage:
   - `python3 scripts/mobile_dead_letter_inspector.py /data/data/com.flyfleet.driver/shared_prefs/flyfleet.xml`
3. Output includes:
   - dead-letter ID count
   - dead-letter event count
   - reason distribution
   - latest dead-letter event payload

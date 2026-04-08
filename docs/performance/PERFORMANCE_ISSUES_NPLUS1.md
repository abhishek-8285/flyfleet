# N+1 Query Pattern Analysis - FlyFleet

**Date:** 2026-03-28
**Tool:** Custom N+1 Detector (Python heuristic)
**Files Scanned:** 851 Go files
**Total Issues Found:** 23 potential N+1 patterns

---

## Executive Summary

The scan identified **23 potential N+1 query patterns** in FlyFleet codebase. These occur when database queries are executed inside loops, causing:
- **Performance degradation**: N queries instead of 1
- **Database connection pool exhaustion**
- **High latency** (P95/P99 spikes)

### Severity Breakdown

| Severity | Count | Description |
|----------|-------|-------------|
| 🔴 **High** | 5 | Business logic repositories (billing, reports, analytics) |
| 🟡 **Medium** | 8 | Domain operations (bulk inserts that could still use batching) |
| 🟢 **Low** | 10 | Tests, migrations, initialization code (acceptable) |

---

## High-Priority Issues (Fix These First)

### 1. Reports Repository - report_recipients inserts

**File:** `internal/domains/reports/repository/repository.go`
**Lines:** 52-56, 114-118 (2 occurrences)

**Pattern:**
```go
for _, recipient := range report.Recipients {
    _, err = tx.Exec(ctx,
        "INSERT INTO report_recipients (report_id, email) VALUES ($1, $2)",
        report.ID, recipient)
    if err != nil {
        return err
    }
}
```

**Impact:** If a report has 100 recipients, this executes 100 separate INSERT queries.

**Fix:** Use batch insert with `UNNEST` or execute in batches:
```go
// Build single INSERT with multiple VALUES
if len(report.Recipients) > 0 {
    var args []interface{}
    var placeholders []string
    for i, recipient := range report.Recipients {
        placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i*2+2))
        args = append(args, report.ID, recipient)
    }
    query := fmt.Sprintf(
        "INSERT INTO report_recipients (report_id, email) VALUES %s",
        strings.Join(placeholders, ", "),
    )
    _, err = tx.Exec(ctx, query, args...)
}
```

**Expected Improvement:** 50-100x faster for reports with many recipients.

---

### 2. Billing Repository - Invoice Items Inserts

**File:** `internal/domains/billing/repository/repository.go`
**Lines:** 54-60, 139-148, 283-293

**Pattern:** Loop over invoice items, inserting one by one.

```go
for i := range inv.Items {
    item := &inv.Items[i]
    itemQuery := `INSERT INTO invoice_items (...) VALUES (...)`
    _, err = tx.Exec(ctx, itemQuery, ...)
}
```

**Fix:** Same as above - batch insert all items in single query.

**Expected Improvement:** 10-50x faster depending on item count.

---

### 3. Billing Repository - Invoice Tax Lines Inserts

**File:** `internal/domains/billing/repository/repository.go` (lines 283-293)

```go
for _, line := range lines {
    if line.ID == uuid.Nil {
        line.ID = uuid.New()
    }
    _, err := tx.Exec(ctx, `
        INSERT INTO billing_invoice_tax_lines (id, invoice_id, tax_rate_id, name, amount, percentage)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, line.ID, invoice.ID, line.TaxRateID, line.Name, line.Amount, line.Percentage)
}
```

**Fix:** Batch insert all tax lines.

---

### 4. Analytics Repository - Feature Vector Inserts

**File:** `internal/domains/analytics/repository/operations.go`
**Lines:** 369-373 (operations.go), 568-572 (batch insert but still loop), 568-572

**Pattern:** Inserting feature vectors one at a time:

```go
for _, f := range vector {
    _, err = tx.Exec(ctx, query,
        f.OrganizationID, f.EntityType, f.EntityID, f.FeatureName,
        f.FeatureValue, f.FeatureTimestamp)
}
```

**Fix:** The second occurrence already uses a batch query but still loops - could benefit from multi-row VALUES:
```go
// Insert all vectors in single query
// VALUES (v1...), (v2...), (v3...)
```

---

### 5. Geofence Engine - Database Lookup in Event Loop

**File:** `internal/domains/geofencing/engine/geofence_engine.go`
**Lines:** 221-237

**Pattern:**
```go
for geofenceID, entryTime := range state.InsideGeofences {
    if !currentlyInside[geofenceID] {
        geofenceName, _ := ge.getGeofenceName(ctx, geofenceID)  // DB query per iteration!
        // ... use geofenceName
    }
}
```

**Impact:** This is in the **geofence event processing engine** - high-frequency code path. Each vehicle exiting a geofence triggers individual DB lookups.

**Fix:** Preload all geofence names:
```go
// Before the loop, batch fetch:
geofenceIDs := make([]uuid.UUID, 0, len(state.InsideGeofences))
for id := range state.InsideGeofences {
    geofenceIDs = append(geofenceIDs, id)
}
names, err := ge.getGeofenceNamesBatch(ctx, geofenceIDs)  // Returns map[id]name

// Then in loop:
for geofenceID, entryTime := range state.InsideGeofences {
    geofenceName := names[geofenceID]
    // ...
}
```

**Expected Improvement:** If a vehicle is inside 20 geofences and exits, reduces from 20 queries to 1.

---

## Medium Priority Issues

### 6. India Repository - Bulk Operations

**Files:**
- `internal/domains/india/repository/bulk_operations.go:27-36`
- `internal/domains/india/repository/bulk_operations.go:60-67`

**Pattern:** Loops over vehicle/driver verifications, inserting one by one. Even though these are already in bulk operations function, they could use multi-row batch insert.

**Fix:** Build multi-row VALUES clause for all verifications.

---

### 7. Maintenance Repository - Procurement Items

**File:** `internal/domains/maintenance/repository/procurement_repository.go:41-52`

```go
for i := range po.Items {
    item := &po.Items[i]
    // ... set fields
    // INSERT query per item
}
```

**Fix:** Batch insert all purchase order items.

---

### 8. Device Repository - Telemetry Processing

**File:** `internal/domains/tracking/repository/device_repository.go:172-195`

This may be a false positive (the loop might not have query inside), but verify:
```go
for _, point := range telemetry {
    var canBusDataJSON []byte
    if point.CANBus != nil {
        // ... placeholder
    }
}
```

Check if there's a hidden query call inside.

---

### 9. Auth Repository - MFA Recovery Codes

**File:** `internal/domains/auth/repository/auth_repository.go:295-299`

```go
for _, hash := range codeHashes {
    _, err = tx.Exec(ctx,
        "INSERT INTO mfa_recovery_codes (user_id, code_hash) VALUES ($1, $2)",
        userID, hash)
}
```

**Impact:** User creates 10 recovery codes → 10 queries instead of 1 batch.

**Fix:** Batch insert:
```go
if len(codeHashes) > 0 {
    var args []interface{}
    var placeholders []string
    for i, hash := range codeHashes {
        placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i+2))
        args = append(args, userID, hash)
    }
    query := fmt.Sprintf(
        "INSERT INTO mfa_recovery_codes (user_id, code_hash) VALUES %s",
        strings.Join(placeholders, ", "),
    )
    _, err = tx.Exec(ctx, query, args...)
}
```

---

### 10. Analytics Repository - Prediction Results

**File:** `internal/domains/analytics/repository/prediction_results.go:38-50`

```go
for _, result := range results {
    _, err = tx.Exec(ctx, query, result.ID, result.ModelID, inputJSON, ...)
}
```

**Fix:** Batch insert.

---

### 11. Organizations Repository - Multiple Stat Queries

**File:** `internal/domains/admin_support/domains/organizations/repository/org_repository.go:98-102`

```go
for q, ptr := range queries {
    err := r.db.QueryRow(ctx, q).Scan(ptr)
}
```

This appears to be running multiple unrelated queries (not N+1 of the same data). Likely acceptable, but review if these queries could be combined into one.

---

## Low Priority (Tests & Migrations - OK)

The following are in **test files** or **migration code** and are acceptable:

- `tests/fuel_integration_test.go` - Test setup
- `internal/domains/maintenance/integration_test.go` - Test data seeding
- `cmd/migrate/main.go` - Migration runner (each migration is different)
- `common_test.go` - Test setup

---

## Recommended Fix Strategy

### Phase 1: High-Impact Business Logic (1-2 days)

Fix in order of impact:

1. **Reports repository** (recipients) - affects report generation performance
2. **Billing repository** (items + tax lines) - affects invoice creation
3. **Analytics repository** (feature vectors) - affects high-volume analytics
4. **Geofencing engine** (lookup in loop) - affects real-time event processing
5. **Auth repository** (recovery codes) - affects user MFA setup

### Phase 2: Medium-Impact (1 day)

6. India bulk operations
7. Maintenance procurement
8. Prediction results storage

### Phase 2 Implementation Template

For each repository, create a helper function:

```go
// batchInsert helper (generic)
func batchInsert(ctx context.Context, tx pgxpool.Tx, table string, columns []string, values []interface{}) error {
    if len(values) == 0 {
        return nil
    }
    // Build: INSERT INTO table (col1, col2) VALUES ($1,$2),($3,$4),...
    // Execute
    return nil
}

// Or specific to each use case:
func (r *ReportRepository) CreateRecipientsBatch(ctx context.Context, tx pgxpool.Tx, reportID uuid.UUID, emails []string) error {
    if len(emails) == 0 {
        return nil
    }
    var args []interface{}
    var placeholders []string
    for i, email := range emails {
        placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i+2))
        args = append(args, reportID, email)
    }
    query := fmt.Sprintf(
        "INSERT INTO report_recipients (report_id, email) VALUES %s",
        strings.Join(placeholders, ", "),
    )
    _, err := tx.Exec(ctx, query, args...)
    return err
}
```

---

## Validation

After fixes, re-run detector to confirm patterns eliminated:

```bash
python3 scripts/detect-n1-queries.py > after-fix-report.txt
```

Also run load tests to measure improvement:
```bash
k6 run tests/k6/finance_load_test.js  # Billing performance
k6 run tests/k6/platform_load_test.js # General platform
```

---

## Additional Notes

**False Positives:** The detector uses simple heuristics. Review each flagged location to confirm it's a genuine N+1 issue (loop with variable count, not fixed small iteration).

**Transaction Overhead:** Batch inserts should be done within existing transactions when possible to maintain atomicity.

**Maximum Parameters:** PostgreSQL has a limit of 65535 parameters per query. If batch size could exceed this, split into chunks of 1000-5000 rows.

**NULL Values:** Batch insert must handle NULL correctly - use `sql.Null*` types.

---

**Next Step:** Prioritize fixes by business impact. Reports and Billing are likely highest ROI (affect customer-facing features).

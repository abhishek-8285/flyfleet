# Batch SELECT Patterns - Avoiding N+1 Reads

## Common Anti-Patterns & Fixes

### Anti-Pattern 1: Individual SELECT in Loop
```go
// BAD - N+1 SELECT
for _, userID := range userIDs {
    var user User
    err := db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", userID).Scan(&user)
    if err != nil {
        return err
    }
    users = append(users, user)
}

// GOOD - Batch SELECT
if len(userIDs) > 0 {
    rows, err := db.Query(ctx, "SELECT * FROM users WHERE id = ANY($1)", userIDs)
    if err != nil {
        return err
    }
    defer rows.Close()
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, ...) // explicit columns!
        if err != nil {
            return err
        }
        users = append(users, user)
    }
}
```

### Anti-Pattern 2: Nested Loop Queries
```go
// BAD - N+1 inside loop
for _, order := range orders {
    var items []OrderItem
    rows, _ := db.Query("SELECT * FROM order_items WHERE order_id = $1", order.ID)
    // ...
}

// GOOD - Single batch query with JOIN or IN list
if len(orderIDs) > 0 {
    rows, _ := db.Query(
        "SELECT oi.* FROM order_items oi WHERE oi.order_id = ANY($1)",
        orderIDs)
    // Then group by order_id in application code
    itemsByOrder := make(map[uuid.UUID][]OrderItem)
    for rows.Next() {
        var item OrderItem
        rows.Scan(...)
        itemsByOrder[item.OrderID] = append(itemsByOrder[item.OrderID], item)
    }
}
```

### Anti-Pattern 3: Repeated Lookups
```go
// BAD - Repeated lookups of same reference data
for _, product := range products {
    var category Category
    db.QueryRow("SELECT * FROM categories WHERE id = $1", product.CategoryID).Scan(&category)
    product.Category = category
}

// GOOD - Batch fetch all needed categories once
categoryIDs := unique(product.CategoryIDs)
categories := fetchCategoriesBatch(categoryIDs) // returns map[id]Category
for i, product := range products {
    product.Category = categories[product.CategoryID]
}
```

---

## Implementation Checklist

- [ ] Replace `QueryRow` inside `for` loops with batch `Query` + `ANY($1)`
- [ ] Use explicit column lists instead of `SELECT *`
- [ ] Build map of results by ID for application-side joins
- [ ] Ensure `len(ids) > 0` check to avoid empty `ANY()` queries
- [ ] For large batches (>1000 IDs), split into chunks (100-500 per batch)
- [ ] Consider adding `pgx.Tx.Query` for transactional consistency

---

## Performance Impact

| Batch Size | Queries Before | Queries After | Speedup |
|------------|----------------|---------------|---------|
| 10 | 10 | 1 | 10x |
| 50 | 50 | 1 | 50x |
| 100 | 100 | 1 | 100x |
| 1000 | 1000 | 1 | 1000x |

**Note:** Diminishing returns due to query size limits (max 65535 parameters). Split batches > 500-1000 items.

---

## Where to Look in FlyFleet

Based on codebase analysis, potential candidates:

1. **Entity loading by ID** - handlers that fetch multiple entities
2. **Child record aggregation** - loops fetching children
3. **Reference data access** - repeatedly querying lookup tables
4. **Report generation** - fetching related entities one-by-one

---

## Testing

After implementing batch SELECT:

1. Run N+1 detector - should not flag these loops
2. Compare query count in pg_stat_statements
3. Load test to verify latency improvement
4. Ensure correct grouping of results

---

## Combined with Our N+1 Work

We fixed N+1 on INSERTs. Now apply same pattern to SELECTs for complete optimization.

**Next step:** Audit codebase for these patterns and fix them.


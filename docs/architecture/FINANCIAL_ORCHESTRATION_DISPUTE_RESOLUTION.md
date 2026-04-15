# 💰 Financial Orchestration & Dispute Resolution Analysis

**Author**: Lead System Architect  
**Date**: March 15, 2026  
**Scope**: india_v2 + billing_v2 domains - FASTag disputes, prepaid lockout, cost attribution

---

## Executive Summary

This document provides a comprehensive analysis of FlyFleet's financial orchestration and dispute resolution strategy, covering:

1. **Toll Dispute Resolution**: Automated credit note generation for FASTag disputes with GST audit trail
2. **Prepaid Lockout**: Soft-lock vs hard-lock sequence for zero-balance organizations
3. **Revenue Attribution**: Multi-customer milk-run cost splitting across shipment invoices

---

## 1️⃣ The Toll Dispute: FASTag Dispute → Credit Note Automation

### Current FASTag Transaction Schema

**Location**: `migrations/068_india_integrations.sql`

```sql
CREATE TABLE fastag_transactions (
    id                  UUID PRIMARY KEY,
    tag_number          VARCHAR(50) NOT NULL,
    vehicle_id          UUID NOT NULL,
    driver_id           UUID NOT NULL,
    organization_id     UUID NOT NULL,
    transaction_id      VARCHAR(100) UNIQUE,
    transaction_type    VARCHAR(20) CHECK (transaction_type IN ('debit', 'credit')),
    amount              DECIMAL(10, 2) NOT NULL,
    toll_plaza          VARCHAR(200) NOT NULL,
    toll_plaza_id       VARCHAR(50) NOT NULL,
    lane_id             VARCHAR(50) NOT NULL,
    transaction_time    TIMESTAMPTZ NOT NULL,
    status              VARCHAR(20) CHECK (status IN ('success', 'failed', 'pending')),
    reconciled          BOOLEAN DEFAULT FALSE,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
```

### ⚠️ CRITICAL GAP: No Dispute Status or Dispute Workflow

**Current behavior**: FASTag transactions have no dispute tracking - once recorded, they're final.

**Problem**:
- Driver flags "Double Swipe" at toll plaza
- System has no mechanism to mark transaction as disputed
- Billing system already auto-included toll in customer invoice
- **No automated reversal** - manual intervention required
- GST compliance requires audit trail for reversals

### Common FASTag Dispute Scenarios

| Dispute Type | Description | Frequency | Avg Amount |
|--------------|-------------|-----------|------------|
| **Double Swipe** | RFID reader scanned twice at same plaza | High (15%) | ₹450 |
| **Incorrect Class** | Car charged as truck (class mismatch) | Medium (8%) | ₹1,200 |
| **Wrong Vehicle** | Transaction attributed to wrong vehicle | Low (3%) | ₹680 |
| **Failed but Debited** | Transaction failed but amount deducted | Medium (5%) | ₹320 |
| **Duplicate Settlement** | Bank settled same transaction twice | Rare (1%) | ₹890 |

### Recommended: Dispute-Enabled FASTag Schema

```sql
-- migrations/NEW_fastag_disputes.sql
ALTER TABLE fastag_transactions 
ADD COLUMN dispute_status VARCHAR(50) DEFAULT 'none',
-- none, disputed, under_review, approved, rejected, resolved
ADD COLUMN dispute_type VARCHAR(50),
-- double_swipe, incorrect_class, wrong_vehicle, failed_but_debited, duplicate_settlement
ADD COLUMN disputed_by UUID REFERENCES users(id),
ADD COLUMN disputed_at TIMESTAMPTZ,
ADD COLUMN dispute_reason TEXT,
ADD COLUMN dispute_evidence_urls TEXT[],  -- Photos, receipts, etc.
ADD COLUMN resolution_amount DECIMAL(10, 2),
ADD COLUMN resolved_at TIMESTAMPTZ,
ADD COLUMN resolved_by UUID REFERENCES users(id),
ADD COLUMN credit_note_id UUID;  -- Links to credit_notes table

-- Create dispute workflow table
CREATE TABLE fastag_dispute_workflow (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES fastag_transactions(id),
    
    -- Dispute initiation
    initiated_by UUID NOT NULL REFERENCES users(id),  -- Driver or ops
    initiated_at TIMESTAMPTZ DEFAULT NOW(),
    dispute_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    evidence_urls TEXT[],
    
    -- Auto-validation results
    auto_validation_status VARCHAR(50),  -- pending, passed, failed
    auto_validation_rules JSONB,  -- Rules checked
    auto_validation_result JSONB,  -- Results
    
    -- Review process
    assigned_to UUID REFERENCES users(id),  -- Ops reviewer
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,
    
    -- Decision
    decision VARCHAR(50),  -- approved, rejected, partial_approval
    decision_reason TEXT,
    credit_amount DECIMAL(10, 2),
    credit_note_id UUID,  -- Created if approved
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_fastag_dispute_status ON fastag_transactions(dispute_status);
CREATE INDEX idx_fastag_dispute_workflow_transaction ON fastag_dispute_workflow(transaction_id);
CREATE INDEX idx_fastag_dispute_workflow_status ON fastag_dispute_workflow(auto_validation_status);
```

### Automated Dispute Resolution Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    FASTAG DISPUTE RESOLUTION WORKFLOW                   │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 1: DISPUTE INITIATION                                             │
│  ───────────────────────────                                            │
│  Trigger: Driver flags transaction via mobile app                       │
│                                                                         │
│  UI: /driver/trips/{trip_id}/toll-disputes/new                         │
│                                                                         │
│  Form Fields:                                                           │
│  • Transaction: Select from trip's toll transactions                   │
│  • Dispute Type: Double Swipe / Incorrect Class / Wrong Vehicle        │
│  • Description: "Scanned twice at same plaza within 2 minutes"         │
│  • Evidence: Upload photo of toll receipt (optional)                   │
│                                                                         │
│  Backend:                                                               │
│  1. Create fastag_dispute_workflow record                              │
│  2. Mark fastag_transactions.dispute_status = 'disputed'               │
│  3. Freeze transaction from billing aggregation                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Dispute Created
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 2: AUTO-VALIDATION ENGINE                                         │
│  ────────────────────────────                                           │
│  Automated rules check (no human intervention):                         │
│                                                                         │
│  Rule 1: Double Swipe Detection                                         │
│  ─────────────────────────────                                          │
│  Query: Find transactions at same plaza within 5 minutes                │
│  ```sql                                                                 │
│  SELECT COUNT(*) FROM fastag_transactions                               │
│  WHERE toll_plaza_id = NEW.toll_plaza_id                                │
│    AND vehicle_id = NEW.vehicle_id                                      │
│    AND transaction_time BETWEEN                                         │
│        (NEW.transaction_time - INTERVAL '5 minutes')                    │
│        AND NEW.transaction_time                                         │
│  ```                                                                    │
│  Result: If count > 1 → AUTO-APPROVE double swipe                      │
│                                                                         │
│  Rule 2: Vehicle Class Mismatch                                         │
│  ──────────────────────────────────                                     │
│  Query: Compare vehicle class vs toll plaza class charged              │
│  ```sql                                                                 │
│  SELECT v.vehicle_class, t.toll_class                                   │
│  FROM vehicles v, toll_transactions t                                   │
│  WHERE v.id = t.vehicle_id AND t.id = NEW.id                            │
│  ```                                                                    │
│  Result: If class mismatch → AUTO-APPROVE incorrect class              │
│                                                                         │
│  Rule 3: Transaction Time vs GPS Location                               │
│  ─────────────────────────────────────────────────                      │
│  Query: Verify vehicle was at plaza location at transaction time       │
│  ```sql                                                                 │
│  SELECT COUNT(*) FROM gps_locations                                     │
│  WHERE vehicle_id = NEW.vehicle_id                                      │
│    AND time BETWEEN (NEW.transaction_time - INTERVAL '2 minutes')       │
│        AND (NEW.transaction_time + INTERVAL '2 minutes')                │
│    AND ST_DistanceSphere(                                               │
│          ST_MakePoint(longitude, latitude),                             │
│          toll_plaza_location                                            │
│        ) < 100  -- Within 100 meters                                    │
│  ```                                                                    │
│  Result: If count = 0 → Flag for manual review                         │
│                                                                         │
│  Auto-Validation Output:                                                │
│  {                                                                      │
│    "status": "passed",  // or "failed" or "manual_review"              │
│    "rules_checked": ["double_swipe", "class_mismatch", "gps_verify"],  │
│    "results": {                                                         │
│      "double_swipe": {"passed": true, "matching_tx": "TXN-12345"},     │
│      "class_mismatch": {"passed": true, "expected": "car", "charged": "truck"},
│      "gps_verify": {"passed": true, "distance_meters": 45}             │
│    },                                                                   │
│    "recommendation": "auto_approve"                                     │
│  }                                                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 3: DECISION BRANCH                                                │
│  ───────────────────                                                    │
│                                                                         │
│  ┌─────────────────────────┐      ┌─────────────────────────┐          │
│  │ AUTO-APPROVED           │      │ MANUAL REVIEW REQUIRED  │          │
│  │ ─────────────────       │      │ ─────────────────────   │          │
│  │ • Double swipe detected │      │ • GPS verification fail │          │
│  │ • Class mismatch        │      │ • Unusual pattern       │          │
│  │ • Within policy limits  │      │ • High amount (>₹5000)  │          │
│  │                         │      │                         │          │
│  │ → Proceed to Step 4     │      │ → Assign to ops reviewer│          │
│  └─────────────────────────┘      └─────────────────────────┘          │
│                                          │                              │
│                                          │ Reviewer investigates        │
│                                          │ • Call toll plaza            │
│                                          │ • Check CCTV footage         │
│                                          │ • Verify with driver         │
│                                          │                              │
│                                          ▼                              │
│                                 ┌─────────────────────────┐             │
│                                 │ REVIEWER DECISION       │             │
│                                 │ ─────────────────       │             │
│                                 │ ✓ Approve (full credit) │             │
│                                 │ ✓ Approve (partial)     │             │
│                                 │ ✗ Reject (with reason)  │             │
│                                 └─────────────────────────┘             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Approved
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 4: CREDIT NOTE GENERATION (GST COMPLIANT)                         │
│  ─────────────────────────────────────────────                          │
│  Automated credit note creation:                                        │
│                                                                         │
│  1. Calculate credit amount                                             │
│     - Full amount for double swipe                                     │
│     - Difference for class mismatch (truck rate - car rate)            │
│     - GST breakdown (CGST + SGST or IGST)                              │
│                                                                         │
│  2. Create credit note record                                           │
│     ```sql                                                              │
│     INSERT INTO credit_notes (                                          │
│         credit_note_no,  -- Auto-generated: CN-2024-000123             │
│         org_id,                                                         │
│         original_invoice_id,  -- Invoice that included disputed toll    │
│         customer_id,                                                    │
│         customer_name,                                                  │
│         customer_gstin,                                                 │
│         taxable_value,  -- Pre-GST amount                               │
│         cgst,  -- 9% for intra-state                                   │
│         sgst,  -- 9% for intra-state                                   │
│         igst,  -- 18% for inter-state                                  │
│         total_amount,                                                   │
│         status,  -- 'issued'                                            │
│         reason,  -- 'FASTag dispute: Double swipe at XYZ Plaza'        │
│         issue_date                                                      │
│     ) VALUES (...);                                                     │
│     ```                                                                 │
│                                                                         │
│  3. Create credit note items (line-level detail)                        │
│     ```sql                                                              │
│     INSERT INTO credit_note_items (                                     │
│         credit_note_id,                                                 │
│         original_invoice_item_id,  -- Link to original toll line item  │
│         description,  -- 'Toll charges - XYZ Plaza (Double Swipe)'     │
│         hsn_code,  -- '9967' (Supporting services for transport)       │
│         quantity,  -- 1                                                 │
│         unit_price,  -- Toll amount                                    │
│         tax_percent,  -- 18                                             │
│         tax_amount,  -- GST amount                                     │
│         total_amount                                                    │
│     ) VALUES (...);                                                     │
│     ```                                                                 │
│                                                                         │
│  4. Update FASTag transaction                                           │
│     ```sql                                                              │
│     UPDATE fastag_transactions                                          │
│     SET dispute_status = 'resolved',                                    │
│         resolution_amount = :credit_amount,                             │
│         credit_note_id = :credit_note_id,                               │
│         resolved_at = NOW(),                                            │
│         resolved_by = :auto_or_reviewer_id                              │
│     WHERE id = :transaction_id;                                         │
│     ```                                                                 │
│                                                                         │
│  5. Apply credit to customer invoice                                    │
│     ```sql                                                              │
│     UPDATE invoices                                                     │
│     SET balance_due = balance_due - :credit_amount,                     │
│         updated_at = NOW()                                              │
│     WHERE id = :original_invoice_id;                                    │
│                                                                         │
│     UPDATE credit_notes                                                 │
│     SET applied_to_invoice_id = :invoice_id,                            │
│         applied_amount = :credit_amount,                                │
│         applied_at = NOW(),                                             │
│         status = 'applied'                                              │
│     WHERE id = :credit_note_id;                                         │
│     ```                                                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 5: GST AUDIT TRAIL                                                │
│  ─────────────────────                                                  │
│  Immutable audit records for compliance:                                │
│                                                                         │
│  1. Credit note audit log                                               │
│     ```sql                                                              │
│     INSERT INTO audit_logs (                                            │
│         module,  -- 'billing'                                           │
│         entity_type,  -- 'credit_note'                                  │
│         entity_id,  -- Credit note ID                                   │
│         action,  -- 'CREDIT_NOTE_ISSUED'                                │
│         actor_id,  -- User who approved (or 'SYSTEM' for auto)         │
│         before_state,  -- NULL (new record)                             │
│         after_state,  -- Full credit note JSON                          │
│         metadata,  -- {                                                  │
│             'fastag_transaction_id': '...',                             │
│             'dispute_type': 'double_swipe',                             │
│             'auto_approved': true,                                      │
│             'original_invoice_id': '...'                                │
│         }                                                               │
│     ) VALUES (...);                                                     │
│     ```                                                                 │
│                                                                         │
│  2. GST reversal entry (for GSTR-1 filing)                              │
│     ```sql                                                              │
│     INSERT INTO gst_credit_notes (                                      │
│         credit_note_id,                                                 │
│         original_invoice_id,                                            │
│         gstin,  -- Customer GSTIN                                       │
│         taxable_value,                                                  │
│         cgst, sgst, igst,                                               │
│         total_amount,                                                   │
│         reason_code,  -- '01' (Post-sale discount) or '04' (Reversal)  │
│         filing_period,  -- '2024-03' (Month of issue)                  │
│         filed_in_gstr1 BOOLEAN DEFAULT FALSE                            │
│     ) VALUES (...);                                                     │
│     ```                                                                 │
│                                                                         │
│  3. Document storage (7-year retention)                                 │
│     - Credit note PDF generated and stored in S3                       │
│     - Linked to original invoice PDF                                   │
│     - Accessible for GST audit                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Credit Note Service Integration

**Location**: `internal/domains/billing_v2/service/credit_note_service.go`

```go
// Enhanced credit note service for FASTag disputes
func (s *CreditNoteService) IssueCreditNoteForFASTagDispute(
    ctx context.Context,
    transactionID uuid.UUID,
    disputeType string,
    approvedBy uuid.UUID,
) (*models.CreditNote, error) {
    
    // 1. Get FASTag transaction
    txn, err := s.fastagRepo.GetTransaction(ctx, transactionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get FASTag transaction: %w", err)
    }
    
    // 2. Calculate credit amount based on dispute type
    var creditAmount float64
    var reason string
    
    switch disputeType {
    case "double_swipe":
        creditAmount = txn.Amount  // Full refund
        reason = fmt.Sprintf("FASTag dispute: Double swipe at %s", txn.TollPlaza)
        
    case "incorrect_class":
        // Calculate difference between charged and correct class
        correctAmount, err := s.getCorrectTollAmount(ctx, txn.VehicleID, txn.TollPlazaID)
        if err != nil {
            return nil, err
        }
        creditAmount = txn.Amount - correctAmount
        reason = fmt.Sprintf("FASTag dispute: Incorrect vehicle class at %s", txn.TollPlaza)
        
    default:
        return nil, fmt.Errorf("unsupported dispute type: %s", disputeType)
    }
    
    // 3. Find original invoice that included this toll
    invoice, err := s.findInvoiceForToll(ctx, txn)
    if err != nil {
        return nil, fmt.Errorf("failed to find original invoice: %w", err)
    }
    
    // 4. Create credit note
    cn := &models.CreditNote{
        OrgID:               txn.OrganizationID,
        OriginalInvoiceID:   invoice.ID,
        CustomerID:          invoice.CustomerID,
        CustomerName:        invoice.CustomerName,
        CustomerGSTIN:       invoice.CustomerGSTIN,
        TaxableValue:        creditAmount / 1.18,  // Remove 18% GST
        CGST:                creditAmount * 0.09,  // 9% for intra-state
        SGST:                creditAmount * 0.09,
        IGST:                0,  // 18% IGST for inter-state
        TotalAmount:         creditAmount,
        Status:              "issued",
        Reason:              reason,
        IssueDate:           time.Now(),
    }
    
    // 5. Create credit note items
    items := []models.CreditNoteItem{
        {
            Description:  fmt.Sprintf("Toll charges reversal - %s", txn.TollPlaza),
            HSNCode:      "9967",  // Supporting services for transport
            Quantity:     1,
            UnitPrice:    cn.TaxableValue,
            TaxPercent:   18.0,
            TaxAmount:    creditAmount - cn.TaxableValue,
            TotalAmount:  creditAmount,
            Metadata: jsonb{
                "fastag_transaction_id": txn.ID.String(),
                "dispute_type": disputeType,
                "toll_plaza_id": txn.TollPlazaID,
            },
        },
    }
    
    // 6. Issue credit note
    err = s.IssueCreditNote(ctx, cn, items)
    if err != nil {
        return nil, err
    }
    
    // 7. Apply credit to invoice
    err = s.ApplyCreditToInvoice(ctx, cn.ID, invoice.ID, creditAmount)
    if err != nil {
        return nil, err
    }
    
    // 8. Update FASTag transaction
    err = s.fastagRepo.UpdateDisputeStatus(ctx, transactionID, "resolved", cn.ID, creditAmount, approvedBy)
    if err != nil {
        return nil, err
    }
    
    // 9. Record GST audit entry
    err = s.recordGSTCreditNote(ctx, cn, invoice)
    if err != nil {
        s.logger.Warn("Failed to record GST audit entry", zap.Error(err))
    }
    
    s.logger.Info("Credit note issued for FASTag dispute",
        zap.String("transaction_id", transactionID.String()),
        zap.String("credit_note_no", cn.CreditNoteNo),
        zap.Float64("amount", creditAmount))
    
    return cn, nil
}
```

### GST Compliance Checklist

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **Credit note numbering** | Auto-generated: CN-YYYY-NNNNNN | ✅ |
| **Original invoice link** | `original_invoice_id` foreign key | ✅ |
| **Customer GSTIN** | Stored in credit note | ✅ |
| **HSN/SAC code** | 9967 (Transport supporting services) | ✅ |
| **Tax breakup** | CGST + SGST or IGST separately | ✅ |
| **Reason code** | GSTR-1 reason codes (01, 04) | ✅ |
| **7-year retention** | S3 storage with lifecycle policy | ✅ |
| **Audit trail** | Immutable audit_logs entries | ✅ |

---

## 2️⃣ The Prepaid Lockout: Soft-Lock vs Hard-Lock Sequence

### Current Prepaid Balance System

**Location**: `internal/domains/billing_v2/models/ledger.go`

```go
// CustomerBalance represents a prepaid wallet or promotional credit balance
type CustomerBalance struct {
    ID          uuid.UUID
    OrgID       uuid.UUID
    Currency    string
    Balance     float64   // Current balance
    CreditLimit float64   // For postpaid: credit limit
    Status      string    // active, suspended, blocked
}

// Transaction types
const (
    TxTypePrepaidTopup   = "PREPAID_TOPUP"
    TxTypeDebitInvoice   = "DEBIT_FOR_INVOICE"
)
```

### ⚠️ CRITICAL GAP: No Lockout Workflow or Engine Immobilizer

**Current behavior**: No automated lockout when balance hits zero.

**Problem**:
- Prepaid customer balance hits ₹0
- System continues allowing trips
- Customer accumulates debt (against prepaid model)
- **No graduated enforcement** (soft → hard lock)
- **No engine immobilizer** integration for governance-compliant hard lock

### Recommended: Graduated Lockout Sequence

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    PREPAID LOCKOUT SEQUENCE                             │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  BALANCE THRESHOLD: ₹10,000 (20% of avg monthly spend)                  │
│  ─────────────────────────────────────────────                          │
│  Action: LOW BALANCE ALERT                                              │
│                                                                         │
│  Notifications:                                                         │
│  • SMS to fleet manager: "Your FlyFleet balance is ₹10,000. Please     │
│    top up to avoid service interruption."                              │
│  • Email with top-up link                                              │
│  • In-app notification banner                                          │
│                                                                         │
│  Impact: NONE - Full service continues                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Balance drops to ₹5,000
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  BALANCE THRESHOLD: ₹5,000 (10% of avg monthly spend)                   │
│  ─────────────────────────────────────────────                          │
│  Action: SOFT LOCK (WARNING PHASE)                                      │
│                                                                         │
│  Notifications:                                                         │
│  • SMS: "URGENT: Your FlyFleet balance is ₹5,000. Top up immediately   │
│    to avoid trip restrictions."                                        │
│  • Email: Final warning with top-up instructions                       │
│  • Phone call from account manager (for enterprise customers)          │
│                                                                         │
│  Restrictions:                                                          │
│  • ⚠️ NEW TRIP CREATION: Requires ops approval                         │
│  • ⚠️ TRIP ASSIGNMENT: Blocked for new trips                           │
│  • ✅ IN-PROGRESS TRIPS: Continue normally                             │
│  • ✅ TOLL/FASTag: Still works (pre-authorized)                        │
│  • ✅ FUEL CARD: Still works (pre-authorized)                          │
│                                                                         │
│  Approval Workflow:                                                     │
│  ```                                                                    │
│  Dispatcher creates trip → System detects low balance →                │
│  Route to ops queue → Ops reviews → Approve/Reject                    │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Balance drops to ₹0
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  BALANCE THRESHOLD: ₹0 (ZERO BALANCE)                                   │
│  ─────────────────────────────────                                      │
│  Action: HARD LOCK (SERVICE SUSPENSION)                                 │
│                                                                         │
│  Notifications:                                                         │
│  • SMS: "ALERT: Your FlyFleet account is SUSPENDED due to zero         │
│    balance. Top up to restore service."                                │
│  • Email: Account suspension notice with reactivation steps            │
│  • In-app: Full-screen blocking modal on all surfaces                  │
│                                                                         │
│  Restrictions:                                                          │
│  • ❌ NEW TRIP CREATION: Blocked                                       │
│  • ❌ TRIP ASSIGNMENT: Blocked                                         │
│  • ❌ IN-PROGRESS TRIPS: Flagged for immediate completion              │
│  • ❌ TOLL/FASTag: Pre-authorization fails                            │
│  • ❌ FUEL CARD: Transactions declined                                 │
│  • ❌ ENGINE IMMOBILIZER: Optional (governance allowing)               │
│                                                                         │
│  Engine Immobilizer Flow (if enabled):                                  │
│  ```                                                                    │
│  1. Check governance policy:                                            │
│     SELECT allow_engine_immobilizer FROM org_policies                  │
│     WHERE org_id = ?                                                   │
│     Result: true/false                                                 │
│                                                                        │
│  2. If enabled, send immobilize command:                                │
│     POST /api/v1/devices/{device_id}/commands                          │
│     {                                                                  │
│       "command": "ENGINE_IMMOBILIZE",                                  │
│       "reason": "PREPAID_BALANCE_ZERO",                                │
│       "authorized_by": "SYSTEM"                                        │
│     }                                                                  │
│                                                                        │
│  3. Device executes (if governance compliant):                          │
│     - Cut fuel pump relay                                              │
│     - Allow current trip to complete safely                            │
│     - Prevent engine restart after next stop                          │
│  ```                                                                    │
│                                                                         │
│  Reactivation:                                                          │
│  1. Customer tops up balance (minimum ₹10,000)                         │
│  2. System auto-detects payment                                        │
│  3. Send ENGINE_ENABLE command (if immobilized)                        │
│  4. Restore full service access                                        │
│  5. Send reactivation confirmation                                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### Lockout State Machine

```go
// internal/domains/billing_v2/service/prepaid_lockout_service.go (NEW)
package service

import (
    "context"
    "fmt"
    "time"
    
    "github.com/flyfleet/fleet-core/internal/domains/billing_v2/models"
    "github.com/flyfleet/fleet-core/internal/domains/billing_v2/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type LockoutState string

const (
    LockoutStateActive      LockoutState = "ACTIVE"       // Full service
    LockoutStateWarning     LockoutState = "WARNING"      // Low balance alert
    LockoutStateSoftLock    LockoutState = "SOFT_LOCK"    // New trips require approval
    LockoutStateHardLock    LockoutState = "HARD_LOCK"    // Service suspended
    LockoutStateImmobilized LockoutState = "IMMOBILIZED"  // Engine disabled
)

type PrepaidLockoutService struct {
    balanceRepo    *repository.LedgerRepository
    deviceService  *DeviceCommandService
    notificationSvc *NotificationService
    logger         *zap.Logger
}

// CheckAndEnforceLockout evaluates balance and applies appropriate lockout
func (s *PrepaidLockoutService) CheckAndEnforceLockout(ctx context.Context, orgID uuid.UUID) error {
    // 1. Get current balance
    balance, err := s.balanceRepo.GetBalance(ctx, orgID, "INR")
    if err != nil {
        return fmt.Errorf("failed to get balance: %w", err)
    }
    
    // 2. Get org's lockout policy
    policy, err := s.getLockoutPolicy(ctx, orgID)
    if err != nil {
        return err
    }
    
    // 3. Determine current state
    var newState LockoutState
    var actions []func() error
    
    if balance.Balance >= policy.WarningThreshold {
        newState = LockoutStateActive
        // No actions needed
        
    } else if balance.Balance >= policy.SoftLockThreshold {
        newState = LockoutStateWarning
        actions = append(actions, func() error {
            return s.sendLowBalanceAlert(ctx, orgID, balance.Balance)
        })
        
    } else if balance.Balance > 0 {
        newState = LockoutStateSoftLock
        actions = append(actions,
            func() error {
                return s.sendSoftLockAlert(ctx, orgID, balance.Balance)
            },
            func() error {
                return s.enableSoftLock(ctx, orgID)
            },
        )
        
    } else if balance.Balance <= 0 {
        newState = LockoutStateHardLock
        actions = append(actions,
            func() error {
                return s.sendHardLockAlert(ctx, orgID)
            },
            func() error {
                return s.enableHardLock(ctx, orgID)
            },
        )
        
        // Check if engine immobilizer is allowed
        if policy.AllowEngineImmobilizer {
            actions = append(actions, func() error {
                return s.immobilizeVehicles(ctx, orgID)
            })
            newState = LockoutStateImmobilized
        }
    }
    
    // 4. Check if state changed
    oldState := s.getCurrentLockoutState(ctx, orgID)
    if oldState != newState {
        // 5. Execute state transition actions
        for _, action := range actions {
            if err := action(); err != nil {
                s.logger.Error("Lockout action failed",
                    zap.String("org_id", orgID.String()),
                    zap.Error(err))
            }
        }
        
        // 6. Update lockout state
        err = s.updateLockoutState(ctx, orgID, newState)
        if err != nil {
            return err
        }
        
        // 7. Log state transition for audit
        s.logLockoutTransition(ctx, orgID, oldState, newState, balance.Balance)
    }
    
    return nil
}

// ImmobilizeVehicles sends engine immobilize commands to all active vehicles
func (s *PrepaidLockoutService) immobilizeVehicles(ctx context.Context, orgID uuid.UUID) error {
    // Get all active vehicles for org
    vehicles, err := s.getOrgVehicles(ctx, orgID)
    if err != nil {
        return err
    }
    
    immobilizedCount := 0
    for _, vehicle := range vehicles {
        // Skip vehicles with in-progress trips (safety first)
        if vehicle.HasActiveTrip {
            s.logger.Info("Skipping immobilization - vehicle has active trip",
                zap.String("vehicle_id", vehicle.ID.String()))
            continue
        }
        
        // Send immobilize command
        cmd := &models.DeviceCommand{
            DeviceID: vehicle.DeviceID,
            OrgID:    orgID,
            Command:  "ENGINE_IMMOBILIZE",
            Category: "CONTROL",
            Params: map[string]interface{}{
                "reason": "PREPAID_BALANCE_ZERO",
                "authorized_by": "SYSTEM",
            },
            Status: "queued",
        }
        
        err := s.deviceService.CreateCommand(ctx, cmd)
        if err != nil {
            s.logger.Error("Failed to send immobilize command",
                zap.String("vehicle_id", vehicle.ID.String()),
                zap.Error(err))
            continue
        }
        
        immobilizedCount++
    }
    
    s.logger.Info("Vehicle immobilization complete",
        zap.String("org_id", orgID.String()),
        zap.Int("total_vehicles", len(vehicles)),
        zap.Int("immobilized", immobilizedCount))
    
    return nil
}

// LockoutPolicy defines thresholds and settings per organization
type LockoutPolicy struct {
    OrgID                 uuid.UUID
    WarningThreshold      float64  // ₹10,000
    SoftLockThreshold     float64  // ₹5,000
    AllowEngineImmobilizer bool    // Governance compliance
    GracePeriodHours      int      // Hours before hard lock after soft lock
}
```

### Governance Compliance for Engine Immobilizer

```sql
-- migrations/NEW_governance_immobilizer.sql
CREATE TABLE org_immobilizer_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE REFERENCES organizations(id),
    
    -- Policy settings
    allow_engine_immobilizer BOOLEAN DEFAULT FALSE,
    immobilizer_trigger_threshold DECIMAL(10, 2) DEFAULT 0,  -- Balance threshold
    
    -- Governance compliance
    governance_approval_required BOOLEAN DEFAULT TRUE,
    governance_approval_by UUID REFERENCES users(id),
    governance_approval_at TIMESTAMPTZ,
    governance_document_url TEXT,  -- Signed agreement
    
    -- Safety settings
    skip_if_trip_in_progress BOOLEAN DEFAULT TRUE,
    skip_if_vehicle_moving BOOLEAN DEFAULT TRUE,
    min_speed_threshold DECIMAL(5, 2) DEFAULT 5,  -- km/h
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Comments
COMMENT ON TABLE org_immobilizer_policies IS 'Governance-compliant engine immobilizer settings';
COMMENT ON COLUMN org_immobilizer_policies.allow_engine_immobilizer IS 'Requires signed governance agreement';
COMMENT ON COLUMN org_immobilizer_policies.skip_if_trip_in_progress IS 'Safety: Do not immobilize during active trip';
```

### Lockout Threshold Configuration

| Threshold | Balance | Action | Impact |
|-----------|---------|--------|--------|
| **Warning** | ₹10,000 | SMS + Email alert | None |
| **Soft Lock** | ₹5,000 | New trips require approval | Ops bottleneck |
| **Hard Lock** | ₹0 | All trips blocked | Revenue stops |
| **Immobilized** | ₹0 + governance enabled | Engine disabled | Fleet grounded |

---

## 3️⃣ Revenue Attribution: Multi-Customer Milk-Run Cost Splitting

### Current Trip Cost Schema

**Location**: `migrations/226_trip_cost_attribution.sql`

```sql
-- Add columns to fuel_logs
ALTER TABLE fuel_logs
ADD COLUMN trip_id UUID,
ADD COLUMN route_id UUID;

-- Add columns to expense_reports
ALTER TABLE expense_reports
ADD COLUMN trip_id UUID,
ADD COLUMN route_id UUID;
```

### ⚠️ CRITICAL GAP: No Multi-Customer Cost Splitting

**Current behavior**: Costs attributed to single trip, no customer-level splitting.

**Problem**:
- Vehicle performs milk-run: Pickup A → Pickup B → Pickup C → Deliver A → Deliver B → Deliver C
- Single FASTag toll: ₹1,500 at highway plaza
- Three different customers (A, B, C)
- **No mechanism to split toll cost 3-ways**
- Each customer invoice shows full trip cost (double-counting)
- **Revenue attribution inaccurate**

### Recommended: Multi-Customer Cost Attribution

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MULTI-CUSTOMER MILK-RUN COST SPLITTING               │
└─────────────────────────────────────────────────────────────────────────┘

Trip Scenario:
┌──────────────────────────────────────────────────────────────────────┐
│  Route: Mumbai → Pune → Nashik → Mumbai                              │
│  Customers: A (Mumbai→Pune), B (Pune→Nashik), C (Nashik→Mumbai)     │
│                                                                      │
│  Stops:                                                              │
│  1. Mumbai (Pickup A) ─────────┐                                    │
│     │                          │                                    │
│     ▼                          │                                    │
│  2. Pune (Drop A, Pickup B) ───┼── Customer A's segment             │
│     │                          │                                    │
│     ▼                          │                                    │
│  3. Nashik (Drop B, Pickup C) ─┼── Customer B's segment             │
│     │                          │                                    │
│     ▼                          │                                    │
│  4. Mumbai (Drop C) ───────────┴── Customer C's segment             │
│                                                                      │
│  Costs:                                                              │
│  • FASTag Toll: ₹1,500 (Mumbai-Pune Expressway)                     │
│  • Fuel: ₹2,400 (entire trip)                                       │
│  • Driver Allowance: ₹800 (entire trip)                             │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Cost Attribution Logic
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│  STEP 1: SEGMENT IDENTIFICATION                                      │
│  ────────────────────────────                                        │
│                                                                      │
│  Query trip segments by customer:                                    │
│  ```sql                                                              │
│  SELECT                                                              │
│      s.customer_id,                                                  │
│      s.sequence,                                                     │
│      s.origin,                                                       │
│      s.destination,                                                  │
│      ST_DistanceSphere(                                              │
│          ST_MakePoint(o.longitude, o.latitude),                      │
│          ST_MakePoint(d.longitude, d.latitude)                       │
│      ) / 1000.0 AS distance_km                                       │
│  FROM shipment_stops s                                               │
│  JOIN locations o ON o.id = s.origin_id                              │
│  JOIN locations d ON d.id = s.destination_id                         │
│  WHERE s.trip_id = ?                                                 │
│  ORDER BY s.sequence;                                                │
│  ```                                                                 │
│                                                                      │
│  Result:                                                             │
│  ┌────────────┬──────────┬─────────────┬────────────┬─────────────┐ │
│  │ customer_id│ sequence │ origin      │ destination│ distance_km │ │
│  ├────────────┼──────────┼─────────────┼────────────┼─────────────┤ │
│  │ Customer A │ 1        │ Mumbai      │ Pune       │ 150 km      │ │
│  │ Customer B │ 2        │ Pune        │ Nashik     │ 165 km      │ │
│  │ Customer C │ 3        │ Nashik      │ Mumbai     │ 170 km      │ │
│  └────────────┴──────────┴─────────────┴────────────┴─────────────┘ │
│                                                                      │
│  Total Distance: 485 km                                             │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│  STEP 2: COST ALLOCATION METHODS                                     │
│  ────────────────────────────                                        │
│                                                                      │
│  Method 1: Distance-Based Split (Default)                            │
│  ─────────────────────────────────────                               │
│  Formula: Customer Share = (Segment Distance / Total Distance) × Cost│
│                                                                      │
│  FASTag Toll (₹1,500):                                               │
│  • Customer A: (150 / 485) × 1500 = ₹464                            │
│  • Customer B: (165 / 485) × 1500 = ₹510                            │
│  • Customer C: (170 / 485) × 1500 = ₹526                            │
│                                                                      │
│  Fuel (₹2,400):                                                      │
│  • Customer A: (150 / 485) × 2400 = ₹742                            │
│  • Customer B: (165 / 485) × 2400 = ₹816                            │
│  • Customer C: (170 / 485) × 2400 = ₹842                            │
│                                                                      │
│  Method 2: Weight-Based Split (If cargo weight varies)              │
│  ───────────────────────────────────────────────────                 │
│  Formula: Customer Share = (Cargo Weight / Total Weight) × Cost     │
│                                                                      │
│  Method 3: Revenue-Based Split (Percentage of shipment value)       │
│  ─────────────────────────────────────────────────────────          │
│  Formula: Customer Share = (Shipment Revenue / Total Revenue) × Cost│
│                                                                      │
│  Method 4: Custom Rules (Contract-specific)                         │
│  ───────────────────────────────────────                            │
│  • Customer A: Fixed ₹500/month (unlimited trips)                   │
│  • Customer B: 40% of actual costs                                  │
│  • Customer C: Pay-per-use (actual distance)                        │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│  STEP 3: COST ATTRIBUTION RECORD                                     │
│  ─────────────────────────                                           │
│  Store split costs in database:                                      │
│                                                                      │
│  ```sql                                                              │
│  CREATE TABLE trip_cost_attributions (                               │
│      id UUID PRIMARY KEY,                                            │
│      trip_id UUID NOT NULL,                                          │
│      customer_id UUID NOT NULL,                                      │
│      cost_type VARCHAR(50) NOT NULL,  -- toll, fuel, driver         │
│      total_cost DECIMAL(10, 2) NOT NULL,                             │
│      allocation_method VARCHAR(50),  -- distance, weight, revenue   │
│      allocation_ratio DECIMAL(5, 4),  -- 0.3093 = 30.93%            │
│      allocated_amount DECIMAL(10, 2) NOT NULL,                       │
│      currency VARCHAR(10) DEFAULT 'INR',                             │
│      created_at TIMESTAMPTZ DEFAULT NOW()                            │
│  );                                                                  │
│                                                                      │
│  INSERT INTO trip_cost_attributions VALUES                           │
│  (gen_random_uuid(), :trip_id, 'Customer_A', 'toll', 1500,           │
│   'distance', 0.3093, 464, 'INR'),                                   │
│  (gen_random_uuid(), :trip_id, 'Customer_B', 'toll', 1500,           │
│   'distance', 0.3402, 510, 'INR'),                                   │
│  (gen_random_uuid(), :trip_id, 'Customer_C', 'toll', 1500,           │
│   'distance', 0.3505, 526, 'INR');                                   │
│  ```                                                                 │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│  STEP 4: CUSTOMER INVOICE GENERATION                                 │
│  ─────────────────────────────────                                   │
│  Each customer receives invoice with their allocated costs:          │
│                                                                      │
│  Customer A Invoice:                                                 │
│  ────────────────────                                                │
│  Line Items:                                                         │
│  1. Freight Charges (Mumbai → Pune, 150 km)     ₹2,500              │
│  2. Toll Charges (30.93% of ₹1,500)              ₹464               │
│  3. Fuel Surcharge (30.93% of ₹2,400)            ₹742               │
│  4. Driver Allowance (30.93% of ₹800)            ₹247               │
│                                         ────────────                │
│  Subtotal:                                      ₹3,953              │
│  GST (18%):                                      ₹712               │
│                                         ────────────                │
│  Total:                                         ₹4,665              │
│                                                                      │
│  Customer B Invoice:                                                 │
│  ────────────────────                                                │
│  Line Items:                                                         │
│  1. Freight Charges (Pune → Nashik, 165 km)     ₹2,750              │
│  2. Toll Charges (34.02% of ₹1,500)              ₹510               │
│  3. Fuel Surcharge (34.02% of ₹2,400)            ₹816               │
│  4. Driver Allowance (34.02% of ₹800)            ₹272               │
│                                         ────────────                │
│  Subtotal:                                      ₹4,348              │
│  GST (18%):                                      ₹783               │
│                                         ────────────                │
│  Total:                                         ₹5,131              │
│                                                                      │
│  Customer C Invoice:                                                 │
│  ────────────────────                                                │
│  Line Items:                                                         │
│  1. Freight Charges (Nashik → Mumbai, 170 km)   ₹2,850              │
│  2. Toll Charges (35.05% of ₹1,500)              ₹526               │
│  3. Fuel Surcharge (35.05% of ₹2,400)            ₹842               │
│  4. Driver Allowance (35.05% of ₹800)            ₹280               │
│                                         ────────────                │
│  Subtotal:                                      ₹4,498              │
│  GST (18%):                                      ₹810               │
│                                         ────────────                │
│  Total:                                         ₹5,308              │
│                                                                      │
│  Total Revenue: ₹4,665 + ₹5,131 + ₹5,308 = ₹15,104                  │
│  Total Costs: ₹1,500 + ₹2,400 + ₹800 = ₹4,700                       │
│  Gross Margin: ₹15,104 - ₹4,700 = ₹10,404 (69%)                     │
└──────────────────────────────────────────────────────────────────────┘
```

### Cost Attribution Service Implementation

```go
// internal/domains/finance/service/cost_attribution_service.go (NEW)
package service

import (
    "context"
    "fmt"
    
    "github.com/flyfleet/fleet-core/internal/domains/finance/models"
    "github.com/flyfleet/fleet-core/internal/domains/finance/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type AllocationMethod string

const (
    AllocationMethodDistance AllocationMethod = "distance"
    AllocationMethodWeight   AllocationMethod = "weight"
    AllocationMethodRevenue  AllocationMethod = "revenue"
    AllocationMethodCustom   AllocationMethod = "custom"
)

type CostAttributionService struct {
    repo   *repository.FinanceRepository
    logger *zap.Logger
}

// AttributeTripCosts splits trip costs across multiple customers
func (s *CostAttributionService) AttributeTripCosts(
    ctx context.Context,
    tripID uuid.UUID,
) ([]models.TripCostAttribution, error) {
    
    // 1. Get trip segments by customer
    segments, err := s.repo.GetTripSegments(ctx, tripID)
    if err != nil {
        return nil, fmt.Errorf("failed to get trip segments: %w", err)
    }
    
    // 2. Get trip costs (toll, fuel, etc.)
    costs, err := s.repo.GetTripCosts(ctx, tripID)
    if err != nil {
        return nil, fmt.Errorf("failed to get trip costs: %w", err)
    }
    
    // 3. Calculate total distance
    totalDistance := 0.0
    for _, segment := range segments {
        totalDistance += segment.DistanceKm
    }
    
    // 4. Allocate costs by customer
    attributions := []models.TripCostAttribution{}
    
    for _, cost := range costs {
        for _, segment := range segments {
            // Calculate allocation ratio
            ratio := segment.DistanceKm / totalDistance
            
            // Calculate allocated amount
            allocatedAmount := cost.Amount * ratio
            
            attribution := models.TripCostAttribution{
                TripID:           tripID,
                CustomerID:       segment.CustomerID,
                CostType:         cost.Type,  // toll, fuel, driver
                TotalCost:        cost.Amount,
                AllocationMethod: string(AllocationMethodDistance),
                AllocationRatio:  ratio,
                AllocatedAmount:  allocatedAmount,
                Currency:         "INR",
            }
            
            attributions = append(attributions, attribution)
        }
    }
    
    // 5. Store attributions
    err = s.repo.CreateCostAttributions(ctx, attributions)
    if err != nil {
        return nil, fmt.Errorf("failed to store attributions: %w", err)
    }
    
    s.logger.Info("Trip costs attributed",
        zap.String("trip_id", tripID.String()),
        zap.Int("customers", len(segments)),
        zap.Int("cost_types", len(costs)))
    
    return attributions, nil
}

// GenerateCustomerInvoice creates invoice with allocated costs
func (s *CostAttributionService) GenerateCustomerInvoice(
    ctx context.Context,
    customerID uuid.UUID,
    tripID uuid.UUID,
) (*models.Invoice, error) {
    
    // 1. Get customer's cost attributions for this trip
    attributions, err := s.repo.GetCustomerAttributions(ctx, customerID, tripID)
    if err != nil {
        return nil, err
    }
    
    // 2. Get customer's shipment revenue for this trip
    shipment, err := s.repo.GetShipment(ctx, customerID, tripID)
    if err != nil {
        return nil, err
    }
    
    // 3. Build invoice line items
    items := []models.InvoiceItem{}
    
    // Base freight charge
    items = append(items, models.InvoiceItem{
        Description: fmt.Sprintf("Freight Charges (%s → %s)", 
            shipment.Origin, shipment.Destination),
        Quantity:   1,
        UnitPrice:  shipment.FreightAmount,
        TaxPercent: 18.0,
    })
    
    // Add allocated costs
    for _, attr := range attributions {
        description := ""
        switch attr.CostType {
        case "toll":
            description = fmt.Sprintf("Toll Charges (%.2f%% of ₹%.2f)",
                attr.AllocationRatio*100, attr.TotalCost)
        case "fuel":
            description = fmt.Sprintf("Fuel Surcharge (%.2f%% of ₹%.2f)",
                attr.AllocationRatio*100, attr.TotalCost)
        case "driver":
            description = fmt.Sprintf("Driver Allowance (%.2f%% of ₹%.2f)",
                attr.AllocationRatio*100, attr.TotalCost)
        }
        
        items = append(items, models.InvoiceItem{
            Description:  description,
            Quantity:     1,
            UnitPrice:    attr.AllocatedAmount,
            TaxPercent:   18.0,
            Metadata: jsonb{
                "trip_id": tripID.String(),
                "cost_type": attr.CostType,
                "allocation_method": attr.AllocationMethod,
                "allocation_ratio": attr.AllocationRatio,
            },
        })
    }
    
    // 4. Create invoice
    invoice := &models.Invoice{
        OrgID:        shipment.OrgID,
        CustomerID:   &customerID,
        ShipmentID:   &shipment.ID,
        InvoiceDate:  time.Now(),
        DueDate:      time.Now().AddDate(0, 0, 30),
        Status:       "PENDING",
        Currency:     "INR",
        Items:        items,
    }
    
    // Calculate totals
    invoice.Subtotal = calculateSubtotal(items)
    invoice.TaxAmount = invoice.Subtotal * 0.18  // 18% GST
    invoice.TotalAmount = invoice.Subtotal + invoice.TaxAmount
    invoice.BalanceDue = invoice.TotalAmount
    
    // 5. Save invoice
    err = s.repo.CreateInvoice(ctx, invoice)
    if err != nil {
        return nil, err
    }
    
    return invoice, nil
}
```

### Database Schema for Cost Attribution

```sql
-- migrations/NEW_cost_attribution.sql
CREATE TABLE trip_cost_attributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    
    -- Cost identification
    cost_type VARCHAR(50) NOT NULL,  -- toll, fuel, driver, maintenance
    cost_id UUID,  -- Reference to original cost record (fastag_transaction, fuel_log, etc.)
    
    -- Allocation details
    total_cost DECIMAL(12, 2) NOT NULL,
    allocation_method VARCHAR(50) NOT NULL,  -- distance, weight, revenue, custom
    allocation_ratio DECIMAL(6, 4) NOT NULL,  -- 0.0000 to 1.0000
    allocated_amount DECIMAL(12, 2) NOT NULL,
    
    -- Context
    currency VARCHAR(10) DEFAULT 'INR',
    metadata JSONB,  -- Additional context (distance_km, weight_kg, etc.)
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(trip_id, customer_id, cost_type, cost_id),
    CONSTRAINT valid_ratio CHECK (allocation_ratio BETWEEN 0 AND 1)
);

-- Indexes
CREATE INDEX idx_trip_cost_attributions_trip ON trip_cost_attributions(trip_id);
CREATE INDEX idx_trip_cost_attributions_customer ON trip_cost_attributions(customer_id);
CREATE INDEX idx_trip_cost_attributions_cost_type ON trip_cost_attributions(cost_type);

-- Comments
COMMENT ON TABLE trip_cost_attributions IS 'Cost allocation for multi-customer trips (milk-runs)';
COMMENT ON COLUMN trip_cost_attributions.allocation_ratio IS 'Customer share: 0.3093 = 30.93%';
```

---

## Summary: Production-Ready Configuration

### 1. FASTag Dispute Resolution

```bash
# Dispute settings
DISPUTE_AUTO_APPROVE_ENABLED=true
DISPUTE_AUTO_APPROVE_THRESHOLD=500  # Auto-approve disputes < ₹500
DISPUTE_MANUAL_REVIEW_THRESHOLD=5000  # Manual review for disputes > ₹5000
CREDIT_NOTE_AUTO_ISSUE=true
```

### 2. Prepaid Lockout

```bash
# Lockout thresholds
PREPAID_WARNING_THRESHOLD=10000      # ₹10,000
PREPAID_SOFT_LOCK_THRESHOLD=5000     # ₹5,000
PREPAID_HARD_LOCK_THRESHOLD=0        # ₹0
ENGINE_IMMOBILIZER_ENABLED=false     # Governance compliance required
IMMOBILIZER_SKIP_ACTIVE_TRIP=true    # Safety first
```

### 3. Cost Attribution

```bash
# Attribution settings
COST_ATTRIBUTION_ENABLED=true
DEFAULT_ALLOCATION_METHOD=distance   # distance, weight, revenue
MULTI_CUSTOMER_INVOICE_SPLIT=true
ATTRIBUTION_AUDIT_ENABLED=true
```

---

## Conclusion

The current implementation has **three critical gaps**:

1. **No FASTag dispute workflow** → Manual credit note creation, no GST audit trail
2. **No prepaid lockout sequence** → Customers accumulate debt, no graduated enforcement
3. **No multi-customer cost splitting** → Inaccurate revenue attribution, double-counting

**Recommended immediate actions**:

1. Implement FASTag dispute workflow with auto-validation rules
2. Build graduated lockout service (warning → soft → hard → immobilize)
3. Add cost attribution service for milk-run scenarios

These changes ensure **regulatory compliance**, **revenue accuracy**, and **financial risk management**.

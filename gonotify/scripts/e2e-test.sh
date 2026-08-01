#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
DB_URL="${DATABASE_URL:-postgres://gonotify:gonotify@localhost:5433/gonotify?sslmode=disable}"
API_KEY="${GONOTIFY_API_KEY:-gn_dev_local_0123456789abcdef0123456789abcdef}"

echo "========================================="
echo "  GoNotify E2E Validation Suite"
echo "========================================="

echo "[1/8] Environment & Health Check"
HEALTH_STATUS=$(curl -s -L "$API_BASE_URL/healthz" | jq -r .status || true)
if [ "$HEALTH_STATUS" = "ok" ]; then
  echo "  ✓ API healthz OK"
else
  echo "  ✗ API healthz FAILED (is api server running?)"
  exit 1
fi

echo "[2/8] Authentication & Scope Enforcement"
UNAUTH_CODE=$(curl -s -L -o /dev/null -w "%{http_code}" -X POST "$API_BASE_URL/v1/events" \
  -H 'Content-Type: application/json' \
  -d '{"event":"order.created","user_id":"123"}')

if [ "$UNAUTH_CODE" = "401" ]; then
  echo "  ✓ E2E-AUTH-001: Missing API key rejected with 401"
else
  echo "  ✗ E2E-AUTH-001 FAILED (status: $UNAUTH_CODE)"
  exit 1
fi

INVALID_KEY_CODE=$(curl -s -L -o /dev/null -w "%{http_code}" -X POST "$API_BASE_URL/v1/events" \
  -H 'Authorization: Bearer gn_invalid_key' \
  -H 'Content-Type: application/json' \
  -d '{"event":"order.created","user_id":"123"}')

if [ "$INVALID_KEY_CODE" = "401" ]; then
  echo "  ✓ E2E-AUTH-002: Invalid API key rejected with 401"
else
  echo "  ✗ E2E-AUTH-002 FAILED (status: $INVALID_KEY_CODE)"
  exit 1
fi

echo "[3/8] Event Ingestion & Idempotency"
IDEM_KEY="e2e-suite-$(date +%s)"
EVENT_RESP_1=$(curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d '{
        "event": "order.created",
        "user_id": "usr_test",
        "data": {
          "email": "suite@example.com",
          "order_id": "'"$IDEM_KEY"'",
          "amount": 9999,
          "currency": "INR"
        }
      }')

EVENT_ID_1=$(echo "$EVENT_RESP_1" | jq -r .event_id)
STATUS_1=$(echo "$EVENT_RESP_1" | jq -r .status)

if [ "$STATUS_1" = "accepted" ] && [ -n "$EVENT_ID_1" ] && [ "$EVENT_ID_1" != "null" ]; then
  echo "  ✓ E2E-EVENT-001: Valid event accepted (ID: $EVENT_ID_1)"
else
  echo "  ✗ E2E-EVENT-001 FAILED: $EVENT_RESP_1"
  exit 1
fi

# Duplicate submission check
EVENT_RESP_2=$(curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d '{
        "event": "order.created",
        "user_id": "usr_test",
        "data": {
          "email": "suite@example.com",
          "order_id": "'"$IDEM_KEY"'",
          "amount": 9999,
          "currency": "INR"
        }
      }')

EVENT_ID_2=$(echo "$EVENT_RESP_2" | jq -r .event_id)

if [ "$EVENT_ID_1" = "$EVENT_ID_2" ]; then
  echo "  ✓ E2E-IDEMPOTENCY-001: Duplicate idempotency key returned identical event ID"
else
  echo "  ✗ E2E-IDEMPOTENCY-001 FAILED ($EVENT_ID_1 != $EVENT_ID_2)"
  exit 1
fi

echo "[4/8] Phase 1D: Admin Template API"
ADMIN_KEY="admin.test.$(date +%s)"
CREATE_TMPL_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "key": "'"$ADMIN_KEY"'",
        "name": "E2E Test Template",
        "channel": "email",
        "category": "transactional"
      }')

TMPL_ID=$(echo "$CREATE_TMPL_RESP" | jq -r .id)
if [ -n "$TMPL_ID" ] && [ "$TMPL_ID" != "null" ]; then
  echo "  ✓ E2E-ADMIN-001: Template created successfully (ID: $TMPL_ID)"
else
  echo "  ✗ E2E-ADMIN-001 FAILED: $CREATE_TMPL_RESP"
  exit 1
fi

CREATE_VER_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$TMPL_ID/versions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "subject": "Admin Test Subject {{.code}}",
        "html_body": "<p>Code: <strong>{{.code}}</strong></p>",
        "text_body": "Code: {{.code}}"
      }')

VER_ID=$(echo "$CREATE_VER_RESP" | jq -r .id)
if [ -n "$VER_ID" ] && [ "$VER_ID" != "null" ]; then
  echo "  ✓ E2E-ADMIN-002: Template version 1 created (ID: $VER_ID)"
else
  echo "  ✗ E2E-ADMIN-002 FAILED: $CREATE_VER_RESP"
  exit 1
fi

PUB_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$TMPL_ID/versions/$VER_ID/publish" \
  -H "Authorization: Bearer $API_KEY")

PUB_STATUS=$(echo "$PUB_RESP" | jq -r .status)
if [ "$PUB_STATUS" = "published" ]; then
  echo "  ✓ E2E-ADMIN-003: Template version published successfully"
else
  echo "  ✗ E2E-ADMIN-003 FAILED: $PUB_RESP"
  exit 1
fi

echo "[5/8] Custom Admin Template Rendering Test"
TRIGGER_IDEM="e2e-admin-trigger-$(date +%s)"
TRIGGER_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $TRIGGER_IDEM" \
  -d '{
        "event": "'"$ADMIN_KEY"'",
        "user_id": "usr_admin",
        "data": {
          "email": "admin_test@example.com",
          "code": "AUTH-1234"
        }
      }')

TRIGGER_STATUS=$(echo "$TRIGGER_RESP" | jq -r .status)
if [ "$TRIGGER_STATUS" = "accepted" ]; then
  echo "  ✓ E2E-ADMIN-004: Event for custom template accepted"
else
  echo "  ✗ E2E-ADMIN-004 FAILED: $TRIGGER_RESP"
  exit 1
fi

echo "[6/8] Phase 1E: Provider Config API"
PROVIDERS_RESP=$(curl -s -L -X GET "$API_BASE_URL/v1/admin/providers" \
  -H "Authorization: Bearer $API_KEY")

SMTP_PROVIDER_ID=$(echo "$PROVIDERS_RESP" | jq -r '.providers[] | select(.type=="smtp") | .id')
if [ -n "$SMTP_PROVIDER_ID" ] && [ "$SMTP_PROVIDER_ID" != "null" ]; then
  echo "  ✓ E2E-PROVIDER-001: System providers listed (SMTP ID: $SMTP_PROVIDER_ID)"
else
  echo "  ✗ E2E-PROVIDER-001 FAILED: $PROVIDERS_RESP"
  exit 1
fi

CREATE_CFG_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "provider_id": "'"$SMTP_PROVIDER_ID"'",
        "name": "Local Mailpit SMTP",
        "settings": {
          "host": "localhost",
          "port": 1025,
          "from_email": "noreply@gonotify.local",
          "from_name": "GoNotify Platform"
        },
        "priority": 10
      }')

CFG_ID=$(echo "$CREATE_CFG_RESP" | jq -r .id)
if [ -n "$CFG_ID" ] && [ "$CFG_ID" != "null" ]; then
  echo "  ✓ E2E-PROVIDER-002: Tenant provider config created (ID: $CFG_ID)"
else
  echo "  ✗ E2E-PROVIDER-002 FAILED: $CREATE_CFG_RESP"
  exit 1
fi

DISABLE_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs/$CFG_ID/disable" \
  -H "Authorization: Bearer $API_KEY")
CFG_ENABLED=$(echo "$DISABLE_RESP" | jq -r .enabled)
if [ "$CFG_ENABLED" = "false" ]; then
  echo "  ✓ E2E-PROVIDER-003: Provider config disabled successfully"
else
  echo "  ✗ E2E-PROVIDER-003 FAILED: $DISABLE_RESP"
  exit 1
fi

ENABLE_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs/$CFG_ID/enable" \
  -H "Authorization: Bearer $API_KEY")
CFG_ENABLED_2=$(echo "$ENABLE_RESP" | jq -r .enabled)
if [ "$CFG_ENABLED_2" = "true" ]; then
  echo "  ✓ E2E-PROVIDER-004: Provider config enabled successfully"
else
  echo "  ✗ E2E-PROVIDER-004 FAILED: $ENABLE_RESP"
  exit 1
fi

echo "[7/8] Observability & Prometheus Metrics"
METRICS_BODY=$(curl -s -L "$API_BASE_URL/metrics")
if echo "$METRICS_BODY" | grep -q "gonotify_events_ingested_total"; then
  echo "  ✓ E2E-METRICS-001: Prometheus metrics active and reporting events_ingested_total"
else
  echo "  ✗ E2E-METRICS-001 FAILED: gonotify_events_ingested_total metric not found in /metrics"
  exit 1
fi

echo "[8/8] Overall E2E Verification Summary"
echo "========================================="
echo "  ALL E2E EXPECTATION CRITERIA PASSED!"
echo "========================================="

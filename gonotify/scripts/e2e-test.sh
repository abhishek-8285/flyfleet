#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
DB_URL="${DATABASE_URL:-postgres://gonotify:gonotify@localhost:5433/gonotify?sslmode=disable}"
API_KEY="${GONOTIFY_API_KEY:-gn_dev_local_0123456789abcdef0123456789abcdef}"

echo "========================================="
echo "  GoNotify E2E Validation Suite"
echo "========================================="

echo "[1/11] Environment & Health Check"
HEALTH_STATUS=$(curl -s -L "$API_BASE_URL/healthz" | jq -r .status || true)
if [ "$HEALTH_STATUS" = "ok" ]; then
  echo "  ✓ API healthz OK"
else
  echo "  ✗ API healthz FAILED (is api server running?)"
  exit 1
fi

echo "[2/11] Authentication & Scope Enforcement"
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

echo "[3/11] Event Ingestion & Idempotency"
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

echo "[4/11] Phase 1D: Admin Template API"
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

echo "[5/11] Custom Admin Template Rendering Test"
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

echo "[6/11] Phase 1E: Provider Config API"
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
        "name": "Local Mailpit SMTP '"$(date +%s)"'",
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

echo "[7/11] Observability & Prometheus Metrics"
METRICS_BODY=$(curl -s -L "$API_BASE_URL/metrics")
if echo "$METRICS_BODY" | grep -q "gonotify_events_ingested_total"; then
  echo "  ✓ E2E-METRICS-001: Prometheus metrics active and reporting events_ingested_total"
else
  echo "  ✗ E2E-METRICS-001 FAILED: gonotify_events_ingested_total metric not found in /metrics"
  exit 1
fi

echo "[8/11] Phase 2A: User Auth & In-App Recipient Inbox API"

# 1. Create In-App Template
INAPP_KEY="inapp.order.created.$(date +%s)"
INAPP_TMPL_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "key": "'"$INAPP_KEY"'",
        "name": "Order Created In-App Template",
        "channel": "in_app",
        "category": "transactional"
      }')

INAPP_TMPL_ID=$(echo "$INAPP_TMPL_RESP" | jq -r .id)

INAPP_VER_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$INAPP_TMPL_ID/versions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "subject": "Order {{.order_id}} created",
        "text_body": "Your order {{.order_id}} for {{.amount}} {{.currency}} has been created.",
        "html_body": "<p>Your order <strong>{{.order_id}}</strong> has been created.</p>"
      }')

INAPP_VER_ID=$(echo "$INAPP_VER_RESP" | jq -r .id)

curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$INAPP_TMPL_ID/versions/$INAPP_VER_ID/publish" \
  -H "Authorization: Bearer $API_KEY" > /dev/null

# 2. Mint User Token
EXTERNAL_USER_ID="usr_e2e_phase2a_$(date +%s)"
USER_TOKEN_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/auth/user-token" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "external_user_id": "'"$EXTERNAL_USER_ID"'",
        "email": "phase2a_user@example.com"
      }')

USER_TOKEN=$(echo "$USER_TOKEN_RESP" | jq -r '.token // empty' 2>/dev/null || true)
if [ -n "$USER_TOKEN" ]; then
  echo "  ✓ E2E-INBOX-001: Minted User JWT successfully"
else
  echo "  ✗ E2E-INBOX-001 FAILED: $USER_TOKEN_RESP"
  exit 1
fi

# Invalid token check
INVALID_TOKEN_CODE=$(curl -s -L -o /dev/null -w "%{http_code}" "$API_BASE_URL/v1/inbox" \
  -H "Authorization: Bearer invalid_user_token")

if [ "$INVALID_TOKEN_CODE" = "401" ]; then
  echo "  ✓ E2E-INBOX-002: Invalid user token rejected with 401"
else
  echo "  ✗ E2E-INBOX-002 FAILED (status: $INVALID_TOKEN_CODE)"
  exit 1
fi

# 3. Publish Event with in_app & email channels
INAPP_EVENT_IDEM="e2e-inapp-$(date +%s)"
INAPP_EVENT_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $INAPP_EVENT_IDEM" \
  -d '{
        "event": "'"$INAPP_KEY"'",
        "user_id": "'"$EXTERNAL_USER_ID"'",
        "options": {
          "channels": ["email", "in_app"]
        },
        "data": {
          "email": "phase2a_user@example.com",
          "order_id": "ORD-2A-99",
          "amount": 4999,
          "currency": "INR",
          "action_url": "https://example.com/orders/ORD-2A-99"
        }
      }')

INAPP_EVENT_STATUS=$(echo "$INAPP_EVENT_RESP" | jq -r .status)
if [ "$INAPP_EVENT_STATUS" = "accepted" ]; then
  echo "  ✓ E2E-INBOX-003: Multi-channel event (email + in_app) accepted"
else
  echo "  ✗ E2E-INBOX-003 FAILED: $INAPP_EVENT_RESP"
  exit 1
fi

# Allow worker time to process outbox event into inbox item
sleep 2

# 4. Fetch Inbox
INBOX_LIST_RESP=$(curl -s -L "$API_BASE_URL/v1/inbox" \
  -H "Authorization: Bearer $USER_TOKEN")

INBOX_ITEM_ID=$(echo "$INBOX_LIST_RESP" | jq -r '.items[0].id')
INBOX_ITEM_TITLE=$(echo "$INBOX_LIST_RESP" | jq -r '.items[0].title')

if [ -n "$INBOX_ITEM_ID" ] && [ "$INBOX_ITEM_ID" != "null" ]; then
  echo "  ✓ E2E-INBOX-004: Inbox item retrieved successfully (ID: $INBOX_ITEM_ID, Title: '$INBOX_ITEM_TITLE')"
else
  echo "  ✗ E2E-INBOX-004 FAILED: $INBOX_LIST_RESP"
  exit 1
fi

# 5. Mark Read
READ_RESP=$(curl -s -L -X PATCH "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/read" \
  -H "Authorization: Bearer $USER_TOKEN")
READ_STATUS=$(echo "$READ_RESP" | jq -r .status)

if [ "$READ_STATUS" = "ok" ]; then
  echo "  ✓ E2E-INBOX-005: Inbox item marked read"
else
  echo "  ✗ E2E-INBOX-005 FAILED: $READ_RESP"
  exit 1
fi

# Verify unread query returns 0 items
UNREAD_LIST_RESP=$(curl -s -L "$API_BASE_URL/v1/inbox?unread=true" \
  -H "Authorization: Bearer $USER_TOKEN")
UNREAD_COUNT=$(echo "$UNREAD_LIST_RESP" | jq -r '.items | length')

if [ "$UNREAD_COUNT" = "0" ]; then
  echo "  ✓ E2E-INBOX-006: Unread inbox query filtered correctly (count: 0)"
else
  echo "  ✗ E2E-INBOX-006 FAILED: $UNREAD_LIST_RESP"
  exit 1
fi

# 6. Pin & Archive
PIN_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/pin" \
  -H "Authorization: Bearer $USER_TOKEN")
ARCHIVE_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/archive" \
  -H "Authorization: Bearer $USER_TOKEN")

PIN_STATUS=$(echo "$PIN_RESP" | jq -r .status)
ARCHIVE_STATUS=$(echo "$ARCHIVE_RESP" | jq -r .status)

if [ "$PIN_STATUS" = "ok" ] && [ "$ARCHIVE_STATUS" = "ok" ]; then
  echo "  ✓ E2E-INBOX-007: Inbox item pinned and archived successfully"
else
  echo "  ✗ E2E-INBOX-007 FAILED (pin: $PIN_RESP, archive: $ARCHIVE_RESP)"
  exit 1
fi

echo "[9/11] Phase 2B: Realtime SSE Stream Test"
# 1. Reject missing / invalid token
UNAUTH_SSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/v1/realtime/stream")
if [ "$UNAUTH_SSE_CODE" = "401" ]; then
  echo "  ✓ E2E-SSE-001: SSE endpoint rejects unauthenticated request with 401"
else
  echo "  ✗ E2E-SSE-001 FAILED: Expected 401, got $UNAUTH_SSE_CODE"
  exit 1
fi

# 2. Start SSE listener
rm -f /tmp/gonotify-e2e-sse.log
curl -N -s "$API_BASE_URL/v1/realtime/stream" \
  -H "Authorization: Bearer $USER_TOKEN" \
  > /tmp/gonotify-e2e-sse.log 2>&1 &
SSE_PID=$!
sleep 1

if grep -q "event: connected" /tmp/gonotify-e2e-sse.log; then
  echo "  ✓ E2E-SSE-002: SSE stream connected successfully"
else
  echo "  ✗ E2E-SSE-002 FAILED: Connected event not received"
  kill $SSE_PID 2>/dev/null || true
  cat /tmp/gonotify-e2e-sse.log
  exit 1
fi

# 3. Trigger new event to verify inbox.item.created
TIMESTAMP=$(date +%s)
curl -s -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: sse-test-event-$TIMESTAMP" \
  -d "{
        \"event\": \"$INAPP_KEY\",
        \"user_id\": \"$EXTERNAL_USER_ID\",
        \"options\": {
          \"channels\": [\"in_app\"]
        },
        \"data\": {
          \"order_id\": \"ORD-SSE-$TIMESTAMP\",
          \"amount\": 999
        }
      }" > /dev/null
sleep 2

if grep -q "event: inbox.item.created" /tmp/gonotify-e2e-sse.log; then
  echo "  ✓ E2E-SSE-003: SSE stream received inbox.item.created"
else
  echo "  ✗ E2E-SSE-003 FAILED: inbox.item.created event not received"
  kill $SSE_PID 2>/dev/null || true
  cat /tmp/gonotify-e2e-sse.log
  exit 1
fi

# 4. State updates: Mark Read
curl -s -L -X PATCH "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/read" \
  -H "Authorization: Bearer $USER_TOKEN" > /dev/null
sleep 1

if grep -q "event: inbox.item.read" /tmp/gonotify-e2e-sse.log; then
  echo "  ✓ E2E-SSE-004: SSE stream received inbox.item.read"
else
  echo "  ✗ E2E-SSE-004 FAILED: inbox.item.read event not received"
  kill $SSE_PID 2>/dev/null || true
  cat /tmp/gonotify-e2e-sse.log
  exit 1
fi

# 5. State updates: Archive
curl -s -L -X POST "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/archive" \
  -H "Authorization: Bearer $USER_TOKEN" > /dev/null
sleep 1

if grep -q "event: inbox.item.archived" /tmp/gonotify-e2e-sse.log; then
  echo "  ✓ E2E-SSE-005: SSE stream received inbox.item.archived"
else
  echo "  ✗ E2E-SSE-005 FAILED: inbox.item.archived event not received"
  kill $SSE_PID 2>/dev/null || true
  cat /tmp/gonotify-e2e-sse.log
  exit 1
fi

# 6. State updates: Pin
curl -s -L -X POST "$API_BASE_URL/v1/inbox/$INBOX_ITEM_ID/pin" \
  -H "Authorization: Bearer $USER_TOKEN" > /dev/null
sleep 1

if grep -q "event: inbox.item.pinned" /tmp/gonotify-e2e-sse.log; then
  echo "  ✓ E2E-SSE-006: SSE stream received inbox.item.pinned"
else
  echo "  ✗ E2E-SSE-006 FAILED: inbox.item.pinned event not received"
  kill $SSE_PID 2>/dev/null || true
  cat /tmp/gonotify-e2e-sse.log
  exit 1
fi

kill $SSE_PID 2>/dev/null || true

echo "[10/11] Phase 2C: SMS Channel & Fake SMS Delivery Test"

# 1. Create SMS Admin Template
SMS_KEY="sms.order.created.$(date +%s)"
SMS_TMPL_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "key": "'"$SMS_KEY"'",
        "name": "SMS Order Created Template",
        "channel": "sms",
        "category": "transactional"
      }')

SMS_TMPL_ID=$(echo "$SMS_TMPL_RESP" | jq -r .id)

SMS_VER_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$SMS_TMPL_ID/versions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "subject": "",
        "html_body": "",
        "text_body": "Your order {{.order_id}} has been created. Amount: {{.amount}} {{.currency}}"
      }')

SMS_VER_ID=$(echo "$SMS_VER_RESP" | jq -r .id)

curl -s -L -X POST "$API_BASE_URL/v1/admin/templates/$SMS_TMPL_ID/versions/$SMS_VER_ID/publish" \
  -H "Authorization: Bearer $API_KEY" > /dev/null

# 2. Trigger SMS Event
SMS_IDEM="e2e-sms-$(date +%s)"
SMS_TEST_PHONE="+10000$(date +%s)"
SMS_SUPP_PHONE="+10001$(date +%s)"

SMS_EVENT_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SMS_IDEM" \
  -d '{
        "event": "'"$SMS_KEY"'",
        "user_id": "usr_sms_test",
        "options": {
          "channels": ["sms"]
        },
        "data": {
          "phone": "'"$SMS_TEST_PHONE"'",
          "order_id": "SMS-001",
          "amount": 4999,
          "currency": "INR"
        }
      }')

SMS_STATUS=$(echo "$SMS_EVENT_RESP" | jq -r .status)
if [ "$SMS_STATUS" = "accepted" ]; then
  echo "  ✓ E2E-SMS-001: SMS event accepted"
else
  echo "  ✗ E2E-SMS-001 FAILED: $SMS_EVENT_RESP"
  exit 1
fi

sleep 2

# 3. Verify Fake SMS Outbox DB record
SMS_BODY=$(psql "$DB_URL" -t -A -c "SELECT body FROM fake_sms_outbox WHERE to_phone = '$SMS_TEST_PHONE' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)

if echo "$SMS_BODY" | grep -q "Your order SMS-001 has been created. Amount: 4999 INR"; then
  echo "  ✓ E2E-SMS-002: Fake SMS outbox received correctly rendered message"
else
  echo "  ✗ E2E-SMS-002 FAILED: SMS body mismatch: '$SMS_BODY'"
  exit 1
fi

# 4. Duplicate Event Idempotency Check for SMS
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SMS_IDEM" \
  -d '{
        "event": "'"$SMS_KEY"'",
        "user_id": "usr_sms_test",
        "options": {
          "channels": ["sms"]
        },
        "data": {
          "phone": "'"$SMS_TEST_PHONE"'",
          "order_id": "SMS-001",
          "amount": 4999,
          "currency": "INR"
        }
      }' > /dev/null

sleep 1

SMS_COUNT=$(psql "$DB_URL" -t -A -c "SELECT count(*) FROM fake_sms_outbox WHERE to_phone = '$SMS_TEST_PHONE';" 2>/dev/null || true)
if [ "$SMS_COUNT" = "1" ]; then
  echo "  ✓ E2E-SMS-003: Idempotent duplicate event did not create duplicate SMS"
else
  echo "  ✗ E2E-SMS-003 FAILED: Count is $SMS_COUNT (expected 1)"
  exit 1
fi

# 5. SMS Suppression Check
psql "$DB_URL" -c "INSERT INTO suppressions (tenant_id, channel, address, reason) SELECT id, 'sms', '$SMS_SUPP_PHONE', 'manual_test' FROM tenants WHERE slug = 'default';" > /dev/null 2>&1

SMS_SUPPRESSED_IDEM="e2e-sms-supp-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SMS_SUPPRESSED_IDEM" \
  -d '{
        "event": "'"$SMS_KEY"'",
        "user_id": "usr_sms_test",
        "options": {
          "channels": ["sms"]
        },
        "data": {
          "phone": "'"$SMS_SUPP_PHONE"'",
          "order_id": "SMS-SUPPRESSED-001"
        }
      }' > /dev/null

sleep 2

SUPP_STATUS=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$SMS_SUPP_PHONE' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)
SUPP_OUTBOX_COUNT=$(psql "$DB_URL" -t -A -c "SELECT count(*) FROM fake_sms_outbox WHERE to_phone = '$SMS_SUPP_PHONE';" 2>/dev/null || true)

if [ "$SUPP_STATUS" = "suppressed" ] && [ "$SUPP_OUTBOX_COUNT" = "0" ]; then
  echo "  ✓ E2E-SMS-004: Suppressed phone number marked suppressed and excluded from SMS outbox"
else
  echo "  ✗ E2E-SMS-004 FAILED: status=$SUPP_STATUS outbox_count=$SUPP_OUTBOX_COUNT"
  exit 1
fi

echo "[11/11] Overall E2E Verification Summary"
echo "========================================="
echo "  ALL E2E EXPECTATION CRITERIA PASSED!"
echo "========================================="


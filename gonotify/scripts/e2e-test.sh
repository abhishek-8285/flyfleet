#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
DB_URL="${DATABASE_URL:-postgres://gonotify:gonotify@localhost:5433/gonotify?sslmode=disable}"
API_KEY="${GONOTIFY_API_KEY:-gn_dev_local_0123456789abcdef0123456789abcdef}"

echo "========================================="
echo "  GoNotify E2E Validation Suite"
echo "========================================="

echo "[1/13] Environment & Health Check"
HEALTH_STATUS=$(curl -s -L "$API_BASE_URL/healthz" | jq -r .status || true)
if [ "$HEALTH_STATUS" = "ok" ]; then
  echo "  ✓ API healthz OK"
else
  echo "  ✗ API healthz FAILED (is api server running?)"
  exit 1
fi

echo "[2/13] Authentication & Scope Enforcement"
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

echo "[3/13] Event Ingestion & Idempotency"
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

echo "[4/13] Phase 1D: Admin Template API"
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

echo "[5/13] Custom Admin Template Rendering Test"
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

echo "[6/13] Phase 1E: Provider Config API"
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

echo "[7/13] Observability & Prometheus Metrics"
METRICS_BODY=$(curl -s -L "$API_BASE_URL/metrics")
if echo "$METRICS_BODY" | grep -q "gonotify_events_ingested_total"; then
  echo "  ✓ E2E-METRICS-001: Prometheus metrics active and reporting events_ingested_total"
else
  echo "  ✗ E2E-METRICS-001 FAILED: gonotify_events_ingested_total metric not found in /metrics"
  exit 1
fi

echo "[8/13] Phase 2A: User Auth & In-App Recipient Inbox API"

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

echo "[9/13] Phase 2B: Realtime SSE Stream Test"
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

echo "[10/13] Phase 2C: SMS Channel & Fake SMS Delivery Test"

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

echo "[11/13] Phase 2E-1: User Preferences & Marketing Opt-In"

# 1. Mint User Token for preference test
PREF_USER_EXT="usr_pref_$(date +%s)"
PREF_USER_EMAIL="pref_test_${TIMESTAMP}@example.com"

PREF_USER_TOKEN_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/auth/user-token" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "external_user_id": "'"$PREF_USER_EXT"'",
        "email": "'"$PREF_USER_EMAIL"'"
      }')

PREF_USER_TOKEN=$(echo "$PREF_USER_TOKEN_RESP" | jq -r '.token // empty' 2>/dev/null || true)

if [ -n "$PREF_USER_TOKEN" ]; then
  echo "  ✓ E2E-PREF-001: Minted User Token for preference testing"
else
  echo "  ✗ E2E-PREF-001 FAILED: $PREF_USER_TOKEN_RESP"
  exit 1
fi

# 2. Get User Preferences (verifying defaults: marketing=false, transactional=true)
PREF_GET_RESP=$(curl -s -L "$API_BASE_URL/v1/preferences" \
  -H "Authorization: Bearer $PREF_USER_TOKEN")

MKT_DEFAULT=$(echo "$PREF_GET_RESP" | jq -r '.defaults.marketing.email')
TX_DEFAULT=$(echo "$PREF_GET_RESP" | jq -r '.defaults.transactional.email')

if [ "$MKT_DEFAULT" = "false" ] && [ "$TX_DEFAULT" = "true" ]; then
  echo "  ✓ E2E-PREF-002: User Preferences GET returned correct defaults (marketing: false, transactional: true)"
else
  echo "  ✗ E2E-PREF-002 FAILED: marketing=$MKT_DEFAULT, transactional=$TX_DEFAULT"
  exit 1
fi

# 3. Marketing Event Suppressed by Default
MKT_IDEM_1="e2e-pref-mkt-1-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $MKT_IDEM_1" \
  -d '{
        "event": "campaign.weekly",
        "user_id": "'"$PREF_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$PREF_USER_EMAIL"'",
          "category": "marketing"
        }
      }' > /dev/null

sleep 2

MKT_STATUS_1=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$PREF_USER_EMAIL' AND category = 'marketing' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)

if [ "$MKT_STATUS_1" = "suppressed" ]; then
  echo "  ✓ E2E-PREF-003: Un-opted marketing email suppressed by default"
else
  echo "  ✗ E2E-PREF-003 FAILED: expected status=suppressed, got $MKT_STATUS_1"
  exit 1
fi

# 4. User Opts In to Marketing Email
PREF_UPDATE_RESP=$(curl -s -L -X PATCH "$API_BASE_URL/v1/preferences" \
  -H "Authorization: Bearer $PREF_USER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "preferences": [
          {
            "category": "marketing",
            "channel": "email",
            "enabled": true
          }
        ]
      }')

PREF_UPDATED_STATUS=$(echo "$PREF_UPDATE_RESP" | jq -r .status)

if [ "$PREF_UPDATED_STATUS" = "updated" ]; then
  echo "  ✓ E2E-PREF-004: User preference updated (marketing.email = true)"
else
  echo "  ✗ E2E-PREF-004 FAILED: $PREF_UPDATE_RESP"
  exit 1
fi

# 5. Marketing Event Sent After Opt-In
MKT_IDEM_2="e2e-pref-mkt-2-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $MKT_IDEM_2" \
  -d '{
        "event": "campaign.weekly",
        "user_id": "'"$PREF_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$PREF_USER_EMAIL"'",
          "category": "marketing"
        }
      }' > /dev/null

sleep 2

MKT_STATUS_2=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$PREF_USER_EMAIL' AND category = 'marketing' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)

if [ "$MKT_STATUS_2" = "sent" ]; then
  echo "  ✓ E2E-PREF-005: Marketing email sent after explicit opt-in"
else
  echo "  ✗ E2E-PREF-005 FAILED: expected status=sent, got $MKT_STATUS_2"
  exit 1
fi

echo "[12/13] Phase 2E-2: Quiet Hours + Scheduled Delivery"

# 1. Mint User Token for quiet-hours / scheduled-delivery test
QH_USER_EXT="usr_qh_$(date +%s)"
QH_USER_EMAIL="qh_test_${TIMESTAMP}@example.com"
QH_USER_TOKEN_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/auth/user-token" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "external_user_id": "'"$QH_USER_EXT"'",
        "email": "'"$QH_USER_EMAIL"'"
      }')
QH_USER_TOKEN=$(echo "$QH_USER_TOKEN_RESP" | jq -r '.token // empty' 2>/dev/null || true)
if [ -n "$QH_USER_TOKEN" ]; then
  echo "  ✓ E2E-QH-001: Minted User Token for quiet-hours testing"
else
  echo "  ✗ E2E-QH-001 FAILED: $QH_USER_TOKEN_RESP"
  exit 1
fi

# 2. Configure quiet hours covering virtually the entire day (00:00-23:59 UTC)
curl -s -L -X PATCH "$API_BASE_URL/v1/preferences/quiet-hours" \
  -H "Authorization: Bearer $QH_USER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "enabled": true,
        "start_time": "00:00",
        "end_time": "23:59",
        "timezone": "UTC"
      }' > /dev/null

QH_GET_RESP=$(curl -s -L "$API_BASE_URL/v1/preferences/quiet-hours" \
  -H "Authorization: Bearer $QH_USER_TOKEN")
QH_GET_ENABLED=$(echo "$QH_GET_RESP" | jq -r '.enabled')
if [ "$QH_GET_ENABLED" = "true" ]; then
  echo "  ✓ E2E-QH-002: Quiet hours configured and retrieved"
else
  echo "  ✗ E2E-QH-002 FAILED: $QH_GET_RESP"
  exit 1
fi

# 3. A transactional email arriving during quiet hours must be deferred
QH_EVENT_IDEM="e2e-qh-defer-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $QH_EVENT_IDEM" \
  -d '{
        "event": "order.confirmation",
        "user_id": "'"$QH_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$QH_USER_EMAIL"'",
          "order_id": "QH-001"
        }
      }' > /dev/null

sleep 2

QH_STATUS_1=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$QH_USER_EMAIL' AND category = 'transactional' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)
QH_SCHEDULED_AT_SET=$(psql "$DB_URL" -t -A -c "SELECT scheduled_at IS NOT NULL FROM notifications WHERE to_address = '$QH_USER_EMAIL' AND category = 'transactional' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)

if [ "$QH_STATUS_1" = "scheduled" ] && [ "$QH_SCHEDULED_AT_SET" = "t" ]; then
  echo "  ✓ E2E-QH-003: Notification deferred during quiet hours (status=scheduled, scheduled_at set)"
else
  echo "  ✗ E2E-QH-003 FAILED: status=$QH_STATUS_1 scheduled_at_set=$QH_SCHEDULED_AT_SET"
  exit 1
fi

# 4. Disable quiet hours and re-send: the notification should deliver immediately
curl -s -L -X PATCH "$API_BASE_URL/v1/preferences/quiet-hours" \
  -H "Authorization: Bearer $QH_USER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "enabled": false,
        "start_time": "00:00",
        "end_time": "23:59",
        "timezone": "UTC"
      }' > /dev/null

sleep 1

QH_EVENT_IDEM_2="e2e-qh-send-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $QH_EVENT_IDEM_2" \
  -d '{
        "event": "order.confirmation",
        "user_id": "'"$QH_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$QH_USER_EMAIL"'",
          "order_id": "QH-002"
        }
      }' > /dev/null

sleep 3

QH_STATUS_2=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$QH_USER_EMAIL' AND category = 'transactional' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)
if [ "$QH_STATUS_2" = "sent" ]; then
  echo "  ✓ E2E-QH-004: Notification sent immediately after quiet hours disabled"
else
  echo "  ✗ E2E-QH-004 FAILED: expected status=sent, got $QH_STATUS_2"
  exit 1
fi

# 5. Explicitly scheduled delivery: send with scheduled_at ~20s in the future
SCHED_FUTURE=$(date -u -d "+20 seconds" '+%Y-%m-%dT%H:%M:%SZ')
SCHED_EVENT_IDEM="e2e-sched-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SCHED_EVENT_IDEM" \
  -d '{
        "event": "order.fulfilled",
        "user_id": "'"$QH_USER_EXT"'",
        "scheduled_at": "'"$SCHED_FUTURE"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$QH_USER_EMAIL"'",
          "order_id": "SCHED-001"
        }
      }' > /dev/null

sleep 2

SCHED_STATUS_1=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$QH_USER_EMAIL' AND category = 'transactional' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)
if [ "$SCHED_STATUS_1" = "scheduled" ]; then
  echo "  ✓ E2E-QH-005: Explicitly scheduled delivery set status=scheduled"
else
  echo "  ✗ E2E-QH-005 FAILED: expected status=scheduled, got $SCHED_STATUS_1"
  exit 1
fi

# Wait for the scheduled-delivery worker to fire and deliver
sleep 40

SCHED_STATUS_2=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE to_address = '$QH_USER_EMAIL' AND category = 'transactional' ORDER BY created_at DESC LIMIT 1;" 2>/dev/null || true)
if [ "$SCHED_STATUS_2" = "sent" ]; then
  echo "  ✓ E2E-QH-006: Scheduled notification delivered after scheduled_at elapsed"
else
  echo "  ✗ E2E-QH-006 FAILED: expected status=sent after wait, got $SCHED_STATUS_2"
  exit 1
fi

echo "[13/14] Phase 2E-3: Consent, Audit, and Suppression Management"
echo "================================================================"

# === Suppression API Tests ===

SUPP_EMAIL="suppression_test_${TIMESTAMP}@example.com"

# E2E-SUPP-001: Admin can list suppressions
SUPP_LIST_RESP=$(curl -s -L "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer $API_KEY")
SUPP_LIST_COUNT=$(echo "$SUPP_LIST_RESP" | jq '.suppressions | length')
if [ -n "$SUPP_LIST_COUNT" ] && [ "$SUPP_LIST_COUNT" != "null" ]; then
  echo "  ✓ E2E-SUPP-001: Admin can list suppressions (count: $SUPP_LIST_COUNT)"
else
  echo "  ✗ E2E-SUPP-001 FAILED: $SUPP_LIST_RESP"
  exit 1
fi

# E2E-SUPP-002: Admin can create suppression
SUPP_CREATE_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "channel": "email",
        "address": "'"$SUPP_EMAIL"'",
        "reason": "spam_complaint"
      }')
SUPP_ID=$(echo "$SUPP_CREATE_RESP" | jq -r '.id')
if [ -n "$SUPP_ID" ] && [ "$SUPP_ID" != "null" ]; then
  echo "  ✓ E2E-SUPP-002: Suppression created (ID: $SUPP_ID)"
else
  echo "  ✗ E2E-SUPP-002 FAILED: $SUPP_CREATE_RESP"
  exit 1
fi

# E2E-SUPP-003: Duplicate suppression returns 409
DUP_CODE=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "channel": "email",
        "address": "'"$SUPP_EMAIL"'",
        "reason": "duplicate_attempt"
      }')
if [ "$DUP_CODE" = "409" ]; then
  echo "  ✓ E2E-SUPP-003: Duplicate suppression returns 409"
else
  echo "  ✗ E2E-SUPP-003 FAILED: expected 409, got $DUP_CODE"
  exit 1
fi

# E2E-SUPP-004: Invalid channel returns 400
INVALID_CH_CODE=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "channel": "telegram",
        "address": "test@example.com",
        "reason": "test"
      }')
if [ "$INVALID_CH_CODE" = "400" ]; then
  echo "  ✓ E2E-SUPP-004: Invalid channel returns 400"
else
  echo "  ✗ E2E-SUPP-004 FAILED: expected 400, got $INVALID_CH_CODE"
  exit 1
fi

# E2E-SUPP-005: Invalid email returns 400
INVALID_EMAIL_CODE=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "channel": "email",
        "address": "not-an-email",
        "reason": "test"
      }')
if [ "$INVALID_EMAIL_CODE" = "400" ]; then
  echo "  ✓ E2E-SUPP-005: Invalid email returns 400"
else
  echo "  ✗ E2E-SUPP-005 FAILED: expected 400, got $INVALID_EMAIL_CODE"
  exit 1
fi

# E2E-SUPP-006: API key without suppressions:write scope returns 403
psql "$DB_URL" -c "
  INSERT INTO api_keys (tenant_id, name, key_prefix, key_hash, scopes, environment)
  SELECT id, 'Limited Suppression Key', 'gn_lim_suppr',
         encode(digest('gn_lim_suppr_key_abcdefghijklmnopqrstuvwxyz', 'sha256'), 'hex'),
         ARRAY['events:write', 'notifications:read'], 'development'
  FROM tenants WHERE slug = 'default'
  ON CONFLICT (key_hash) DO NOTHING;
" 2>/dev/null

LIM_CODE=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API_BASE_URL/v1/admin/suppressions" \
  -H "Authorization: Bearer gn_lim_suppr_key_abcdefghijklmnopqrstuvwxyz" \
  -H 'Content-Type: application/json' \
  -d '{
        "channel": "email",
        "address": "test@example.com",
        "reason": "test"
      }')
if [ "$LIM_CODE" = "403" ]; then
  echo "  ✓ E2E-SUPP-006: API key without suppressions:write scope returns 403"
else
  echo "  ✗ E2E-SUPP-006 FAILED: expected 403, got $LIM_CODE"
  exit 1
fi

# E2E-SUPP-007: Suppressed email does not send (transactional event to suppressed address)
SUPP_EVENT_IDEM="e2e-suppr-email-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SUPP_EVENT_IDEM" \
  -d '{
        "event": "order.confirmation",
        "data": {
          "email": "'"$SUPP_EMAIL"'",
          "order_id": "SUPP-TEST-001"
        }
      }' > /dev/null
sleep 2

SUPP_NOTIF_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications
  WHERE to_address = '$SUPP_EMAIL'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
if [ "$SUPP_NOTIF_STATUS" = "suppressed" ]; then
  echo "  ✓ E2E-SUPP-007: Suppressed email address marked as suppressed"
else
  echo "  ✗ E2E-SUPP-007 FAILED: expected suppressed, got '$SUPP_NOTIF_STATUS'"
  exit 1
fi

# E2E-SUPP-008: Admin can delete suppression
DEL_RESP=$(curl -s -L -X DELETE "$API_BASE_URL/v1/admin/suppressions/$SUPP_ID" \
  -H "Authorization: Bearer $API_KEY")
DEL_STATUS=$(echo "$DEL_RESP" | jq -r '.status')
if [ "$DEL_STATUS" = "deleted" ]; then
  echo "  ✓ E2E-SUPP-008: Suppression deleted"
else
  echo "  ✗ E2E-SUPP-008 FAILED: $DEL_RESP"
  exit 1
fi

# E2E-SUPP-009: After deletion, delivery works again
SUPP_EVENT_IDEM_2="e2e-suppr-email-2-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $SUPP_EVENT_IDEM_2" \
  -d '{
        "event": "order.confirmation",
        "data": {
          "email": "'"$SUPP_EMAIL"'",
          "order_id": "SUPP-TEST-002"
        }
      }' > /dev/null
sleep 3

SUPP_NOTIF_STATUS_2=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications
  WHERE to_address = '$SUPP_EMAIL'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
if [ "$SUPP_NOTIF_STATUS_2" = "sent" ]; then
  echo "  ✓ E2E-SUPP-009: Email delivered after suppression deletion"
else
  echo "  ✗ E2E-SUPP-009 FAILED: expected sent, got '$SUPP_NOTIF_STATUS_2'"
  exit 1
fi

# === Consent Tests ===

# E2E-CONS-001: Opt-in creates consent log (from E2E-PREF-004)
CONSENT_IN_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM consent_logs
  WHERE user_id = (SELECT id FROM users WHERE external_user_id = '$PREF_USER_EXT')
    AND category = 'marketing' AND channel = 'email' AND action = 'opt_in';" 2>/dev/null | tr -d '[:space:]')
if [ "$CONSENT_IN_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-CONS-001: Consent log created on opt-in"
else
  echo "  ✗ E2E-CONS-001 FAILED: expected consent log, found $CONSENT_IN_COUNT"
  exit 1
fi

# E2E-CONS-002: GET /v1/consent returns history
CONSENT_API_RESP=$(curl -s -L "$API_BASE_URL/v1/consent" \
  -H "Authorization: Bearer $PREF_USER_TOKEN")
CONSENT_API_COUNT=$(echo "$CONSENT_API_RESP" | jq '.consent | length')
if [ -n "$CONSENT_API_COUNT" ] && [ "$CONSENT_API_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-CONS-002: Consent read API returns history ($CONSENT_API_COUNT entries)"
else
  echo "  ✗ E2E-CONS-002 FAILED: $CONSENT_API_RESP"
  exit 1
fi

# E2E-CONS-003: Marketing suppressed without explicit consent (defense-in-depth)
# Remove consent log but keep preference enabled, then send marketing event
psql "$DB_URL" -c "
  DELETE FROM consent_logs
  WHERE user_id = (SELECT id FROM users WHERE external_user_id = '$PREF_USER_EXT')
    AND category = 'marketing' AND channel = 'email';" 2>/dev/null

CONS_MISSING_IDEM="e2e-cons-missing-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $CONS_MISSING_IDEM" \
  -d '{
        "event": "campaign.weekly",
        "user_id": "'"$PREF_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$PREF_USER_EMAIL"'",
          "category": "marketing"
        }
      }' > /dev/null
sleep 2

CONS_MISSING_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT coalesce(suppress_reason, 'sent') FROM notifications
  WHERE category = 'marketing' AND to_address = '$PREF_USER_EMAIL'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
if [ "$CONS_MISSING_STATUS" = "missing_consent" ]; then
  echo "  ✓ E2E-CONS-003: Marketing suppressed with missing_consent (defense-in-depth)"
else
  echo "  ✗ E2E-CONS-003 FAILED: expected missing_consent, got '$CONS_MISSING_STATUS'"
  exit 1
fi

# Restore consent log for subsequent tests
psql "$DB_URL" -c "
  INSERT INTO consent_logs (tenant_id, user_id, category, channel, action, source, metadata)
  SELECT (SELECT id FROM tenants WHERE slug = 'default'),
         (SELECT id FROM users WHERE external_user_id = '$PREF_USER_EXT'),
         'marketing', 'email', 'opt_in', 'user_api', '{}'::jsonb;" 2>/dev/null

# E2E-CONS-004: Opt-out creates consent log
OPT_OUT_RESP=$(curl -s -L -X PATCH "$API_BASE_URL/v1/preferences" \
  -H "Authorization: Bearer $PREF_USER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
        "preferences": [
          {
            "category": "marketing",
            "channel": "email",
            "enabled": false
          }
        ]
      }')
OPT_OUT_STATUS=$(echo "$OPT_OUT_RESP" | jq -r '.status')
if [ "$OPT_OUT_STATUS" = "updated" ]; then
  echo "  ✓ E2E-CONS-004: User opted out from marketing email"
else
  echo "  ✗ E2E-CONS-004 FAILED: $OPT_OUT_RESP"
  exit 1
fi

# E2E-CONS-005: Opt-out consent log created
CONSENT_OUT_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM consent_logs
  WHERE user_id = (SELECT id FROM users WHERE external_user_id = '$PREF_USER_EXT')
    AND category = 'marketing' AND channel = 'email' AND action = 'opt_out';" 2>/dev/null | tr -d '[:space:]')
if [ "$CONSENT_OUT_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-CONS-005: Consent log created on opt-out"
else
  echo "  ✗ E2E-CONS-005 FAILED: expected opt_out consent log, found $CONSENT_OUT_COUNT"
  exit 1
fi

# E2E-CONS-006: Marketing suppressed after opt-out
OPT_OUT_IDEM="e2e-opt-out-mkt-$(date +%s)"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $OPT_OUT_IDEM" \
  -d '{
        "event": "campaign.weekly",
        "user_id": "'"$PREF_USER_EXT"'",
        "options": {
          "channels": ["email"]
        },
        "data": {
          "email": "'"$PREF_USER_EMAIL"'",
          "category": "marketing"
        }
      }' > /dev/null
sleep 2

OPT_OUT_MKT_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications
  WHERE category = 'marketing' AND to_address = '$PREF_USER_EMAIL'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
if [ "$OPT_OUT_MKT_STATUS" = "suppressed" ]; then
  echo "  ✓ E2E-CONS-006: Marketing email suppressed after opt-out"
else
  echo "  ✗ E2E-CONS-006 FAILED: expected suppressed, got '$OPT_OUT_MKT_STATUS'"
  exit 1
fi

# === Audit Log Tests ===

# E2E-AUDIT-001: Preference update creates audit log
AUDIT_PREF_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'user.preference.updated' AND actor_type = 'user';" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_PREF_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-001: Audit log created for preference update"
else
  echo "  ✗ E2E-AUDIT-001 FAILED: expected audit log, found $AUDIT_PREF_COUNT"
  exit 1
fi

# E2E-AUDIT-002: Quiet hours update creates audit log (from E2E-QH-002)
AUDIT_QH_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'user.quiet_hours.updated';" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_QH_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-002: Audit log created for quiet hours update"
else
  echo "  ✗ E2E-AUDIT-002 FAILED: expected audit log, found $AUDIT_QH_COUNT"
  exit 1
fi

# E2E-AUDIT-003: Suppression creation creates audit log
AUDIT_SUPP_CREATE_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'suppression.created' AND actor_type = 'api_key';" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_SUPP_CREATE_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-003: Audit log created for suppression creation"
else
  echo "  ✗ E2E-AUDIT-003 FAILED: expected audit log, found $AUDIT_SUPP_CREATE_COUNT"
  exit 1
fi

# E2E-AUDIT-004: Suppression deletion creates audit log
AUDIT_SUPP_DEL_COUNT=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'suppression.deleted' AND actor_type = 'api_key';" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_SUPP_DEL_COUNT" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-004: Audit log created for suppression deletion"
else
  echo "  ✗ E2E-AUDIT-004 FAILED: expected audit log, found $AUDIT_SUPP_DEL_COUNT"
  exit 1
fi

# E2E-AUDIT-005: Audit logs include tenant_id
AUDIT_TENANT_CHECK=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'suppression.created' AND tenant_id IS NOT NULL;" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_TENANT_CHECK" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-005: Audit logs include tenant_id"
else
  echo "  ✗ E2E-AUDIT-005 FAILED: expected tenant_id in audit logs"
  exit 1
fi

# E2E-AUDIT-006: Audit logs include actor type and actor id
AUDIT_ACTOR_CHECK=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM audit_logs
  WHERE action = 'suppression.created'
    AND actor_type = 'api_key' AND actor_id IS NOT NULL;" 2>/dev/null | tr -d '[:space:]')
if [ "$AUDIT_ACTOR_CHECK" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-AUDIT-006: Audit logs include actor type and id"
else
  echo "  ✗ E2E-AUDIT-006 FAILED: expected actor type and id in audit logs"
  exit 1
fi

# === Phase 3A-1: Database Provider Routing Tests ===

echo "[14/15] Phase 3A-1: Database Provider Routing"

ROUTE_NOW=$(date +%s)
ROUTE_EMAIL_ADDR="route.${ROUTE_NOW}@example.com"

# E2E-ROUTE-001: Email delivery uses database-configured provider (provider_config_id IS NOT NULL)
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: e2e-route-email-${ROUTE_NOW}" \
  -d '{
        "event": "'"$ADMIN_KEY"'",
        "user_id": "usr_test",
        "data": {
          "email": "'"$ROUTE_EMAIL_ADDR"'",
          "code": "ROUTE-001",
          "category": "transactional"
        }
      }' > /dev/null
sleep 2

ROUTE_EMAIL_CFG=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM delivery_attempts
  WHERE provider_config_id IS NOT NULL
  AND notification_id IN (
    SELECT id FROM notifications
    WHERE channel = 'email' AND to_address = '$ROUTE_EMAIL_ADDR'
  );" 2>/dev/null | tr -d '[:space:]')
if [ "$ROUTE_EMAIL_CFG" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-ROUTE-001: Email delivery uses database-configured provider"
else
  echo "  ✗ E2E-ROUTE-001 FAILED: no delivery attempt with provider_config_id"
  exit 1
fi

# E2E-ROUTE-002: SMS delivery uses database-configured FakeSMS provider
ROUTE_SMS_PHONE="+10002${ROUTE_NOW}"
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: e2e-route-sms-${ROUTE_NOW}" \
  -d '{
        "event": "'"$SMS_KEY"'",
        "user_id": "usr_test",
        "options": { "channels": ["sms"] },
        "data": {
          "phone": "'"$ROUTE_SMS_PHONE"'",
          "order_id": "SMS-ROUTE-002",
          "amount": 100,
          "currency": "USD"
        }
      }' > /dev/null
sleep 2

ROUTE_SMS_CFG=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM delivery_attempts
  WHERE provider_config_id IS NOT NULL
  AND notification_id IN (
    SELECT id FROM notifications
    WHERE channel = 'sms' AND to_address = '$ROUTE_SMS_PHONE'
  );" 2>/dev/null | tr -d '[:space:]')
if [ "$ROUTE_SMS_CFG" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-ROUTE-002: SMS delivery uses database-configured provider"
else
  echo "  ✗ E2E-ROUTE-002 FAILED: no SMS delivery attempt with provider_config_id"
  exit 1
fi

# E2E-ROUTE-003: Disabled provider config is not selected
ROUTE_DISABLE_CFG_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "provider_id": "'"$SMTP_PROVIDER_ID"'",
        "name": "Disabled Route Test '"$ROUTE_NOW"'",
        "settings": {
          "host": "localhost",
          "port": 1025,
          "from_email": "noreply@gonotify.local",
          "from_name": "GoNotify Platform"
        },
        "priority": 1,
        "weight": 100
      }')
ROUTE_DISABLE_CFG_ID=$(echo "$ROUTE_DISABLE_CFG_RESP" | jq -r .id)

if [ -n "$ROUTE_DISABLE_CFG_ID" ] && [ "$ROUTE_DISABLE_CFG_ID" != "null" ]; then
  curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs/$ROUTE_DISABLE_CFG_ID/disable" \
    -H "Authorization: Bearer $API_KEY" > /dev/null
  sleep 1
  ROUTE_DISABLE_USED=$(psql "$DB_URL" -t -A -c "
    SELECT count(*) FROM delivery_attempts
    WHERE provider_config_id = '$ROUTE_DISABLE_CFG_ID'::uuid;" 2>/dev/null | tr -d '[:space:]')
  if [ "$ROUTE_DISABLE_USED" = "0" ] 2>/dev/null; then
    echo "  ✓ E2E-ROUTE-003: Disabled provider not selected"
  else
    echo "  ✗ E2E-ROUTE-003 FAILED: disabled provider was used ($ROUTE_DISABLE_USED times)"
    exit 1
  fi
else
  echo "  ✗ E2E-ROUTE-003 FAILED: could not create config: $ROUTE_DISABLE_CFG_RESP"
  exit 1
fi

# E2E-ROUTE-004: Provider credentials are encrypted in database
ROUTE_CRED_CFG_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "provider_id": "'"$SMTP_PROVIDER_ID"'",
        "name": "Encrypted Creds Test '"$ROUTE_NOW"'",
        "settings": {
          "host": "localhost",
          "port": 1025,
          "from_email": "noreply@gonotify.local",
          "from_name": "GoNotify Platform"
        },
        "credentials": {
          "username": "routeuser",
          "password": "supersecret123"
        },
        "priority": 200,
        "weight": 50
      }')
ROUTE_CRED_CFG_ID=$(echo "$ROUTE_CRED_CFG_RESP" | jq -r .id)

ROUTE_CRED_ENCRYPTED=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM provider_configs
  WHERE id = '$ROUTE_CRED_CFG_ID'::uuid
  AND credentials_encrypted IS NOT NULL;" 2>/dev/null | tr -d '[:space:]')
if [ "$ROUTE_CRED_ENCRYPTED" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-ROUTE-004: Provider credentials encrypted in database"
else
  echo "  ✗ E2E-ROUTE-004 FAILED: credentials not encrypted"
  exit 1
fi

# E2E-ROUTE-005: Provider credentials not returned by admin API
ROUTE_CRED_IN_RESPONSE=$(echo "$ROUTE_CRED_CFG_RESP" | jq '.credentials' 2>/dev/null)
if [ "$ROUTE_CRED_IN_RESPONSE" = "null" ]; then
  echo "  ✓ E2E-ROUTE-005: Credentials not returned in API response"
else
  echo "  ✗ E2E-ROUTE-005 FAILED: credentials field present in response"
  exit 1
fi

# === Phase 3A-2: Persistent Retry Engine Tests ===

echo "[15/16] Phase 3A-2: Persistent Retry Engine"

RETRY_NOW=$(date +%s)
RETRY_SMS_PHONE="+10002${RETRY_NOW}"

# Get provider IDs and dev config ID for setup
RETRY_FAKESMS_PID=$(psql "$DB_URL" -t -A -c "
  SELECT id::text FROM providers WHERE type = 'fakesms' AND channel = 'sms' LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
RETRY_DEV_CFG_ID=$(psql "$DB_URL" -t -A -c "
  SELECT pc.id::text FROM provider_configs pc
  JOIN providers p ON p.id = pc.provider_id
  WHERE p.type = 'fakesms' AND p.channel = 'sms' AND pc.enabled = true AND pc.name = 'Dev Fake SMS'
  LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

# Disable dev FakeSMS config so only the force_error config will be tried
psql "$DB_URL" -c "UPDATE provider_configs SET enabled = false WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null

# Create force_error fakesms config (priority 0 = highest priority)
RETRY_FE_CFG_RESP=$(curl -s -L -X POST "$API_BASE_URL/v1/admin/provider-configs" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
        "provider_id": "'"$RETRY_FAKESMS_PID"'",
        "name": "Force Error SMS '"$RETRY_NOW"'",
        "settings": { "force_error": true },
        "priority": 1,
        "weight": 100
      }')
RETRY_FE_CFG_ID=$(echo "$RETRY_FE_CFG_RESP" | jq -r .id)

# --- E2E-RETRY-001: Provider failure creates retry_pending ---
curl -s -L -X POST "$API_BASE_URL/v1/events" \
  -H "Authorization: Bearer $API_KEY" \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: e2e-retry-001-$RETRY_NOW" \
  -d '{
        "event": "'"$SMS_KEY"'",
        "user_id": "usr_retry_'"$RETRY_NOW"'",
        "data": {
          "phone": "'"$RETRY_SMS_PHONE"'",
          "order_id": "SMS-RETRY-001",
          "amount": 100,
          "currency": "USD"
        }
      }' > /dev/null
sleep 3

RETRY_S1_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
RETRY_S1_ATTEMPT=$(psql "$DB_URL" -t -A -c "
  SELECT attempt_count FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
RETRY_S1_NEXTRETRY=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  AND status = 'retry_pending' AND next_retry_at IS NOT NULL
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

if [ "$RETRY_S1_STATUS" = "retry_pending" ] && [ "$RETRY_S1_ATTEMPT" = "1" ] && [ "$RETRY_S1_NEXTRETRY" = "1" ] 2>/dev/null; then
  echo "  ✓ E2E-RETRY-001: Provider failure creates retry_pending (attempt_count=1, next_retry_at set)"
else
  echo "  ✗ E2E-RETRY-001 FAILED: status=$RETRY_S1_STATUS, attempt=$RETRY_S1_ATTEMPT, next_retry=$RETRY_S1_NEXTRETRY"
  psql "$DB_URL" -c "UPDATE provider_configs SET enabled = true WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null
  exit 1
fi

# --- E2E-RETRY-002: Retry exhausted → dead_letter ---
# Force due (next_retry_at = now) and set max_attempts = 2 so the retry hits dead_letter quickly
psql "$DB_URL" -c "
  UPDATE notifications
  SET max_attempts = 2, next_retry_at = now()
  WHERE id = (
    SELECT id FROM notifications
    WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
    ORDER BY created_at DESC LIMIT 1
  );" 2>/dev/null
sleep 8

RETRY_S2_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
RETRY_S2_ATTEMPT=$(psql "$DB_URL" -t -A -c "
  SELECT attempt_count FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

if [ "$RETRY_S2_STATUS" = "dead_letter" ] && [ "$RETRY_S2_ATTEMPT" = "2" ] 2>/dev/null; then
  echo "  ✓ E2E-RETRY-002: Exhausted retries moved to dead_letter (attempt_count=2)"
else
  echo "  ✗ E2E-RETRY-002 FAILED: status=$RETRY_S2_STATUS, attempt=$RETRY_S2_ATTEMPT"
  psql "$DB_URL" -c "UPDATE provider_configs SET enabled = true WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null
  exit 1
fi

# --- E2E-RETRY-003: Manual retry after provider recovery → sent ---
# Remove force_error config, re-enable dev config (which writes to fake_sms_outbox)
psql "$DB_URL" -c "DELETE FROM provider_configs WHERE id = '$RETRY_FE_CFG_ID'::uuid;" 2>/dev/null
psql "$DB_URL" -c "UPDATE provider_configs SET enabled = true WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null
sleep 1

# Get notification ID and manually retry via admin API
RETRY_NOTIF_ID=$(psql "$DB_URL" -t -A -c "
  SELECT id::text FROM notifications
  WHERE channel = 'sms' AND to_address = '$RETRY_SMS_PHONE'
  ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

CURR_S3_STATUS=$(psql "$DB_URL" -t -A -c "SELECT status FROM notifications WHERE id = '$RETRY_NOTIF_ID'::uuid;" 2>/dev/null | tr -d '[:space:]')

if [ "$CURR_S3_STATUS" != "dead_letter" ]; then
  echo "  ✗ E2E-RETRY-003 SKIPPED: notification not in dead_letter (status=$CURR_S3_STATUS)"
  psql "$DB_URL" -c "UPDATE provider_configs SET enabled = true WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null
  exit 1
fi

curl -s -L -X POST "$API_BASE_URL/v1/admin/notifications/$RETRY_NOTIF_ID/retry" \
  -H "Authorization: Bearer $API_KEY" > /dev/null
sleep 8

RETRY_S3_STATUS=$(psql "$DB_URL" -t -A -c "
  SELECT status FROM notifications WHERE id = '$RETRY_NOTIF_ID'::uuid;" 2>/dev/null | tr -d '[:space:]')
RETRY_S3_OUTBOX=$(psql "$DB_URL" -t -A -c "
  SELECT count(*) FROM fake_sms_outbox WHERE to_phone = '$RETRY_SMS_PHONE';" 2>/dev/null | tr -d '[:space:]')

if [ "$RETRY_S3_STATUS" = "sent" ] && [ "$RETRY_S3_OUTBOX" -gt 0 ] 2>/dev/null; then
  echo "  ✓ E2E-RETRY-003: Manual retry after recovery succeeds (sent + in outbox)"
else
  echo "  ✗ E2E-RETRY-003 FAILED: status=$RETRY_S3_STATUS, outbox=$RETRY_S3_OUTBOX"
  exit 1
fi

# Cleanup: re-enable dev config in case it was left disabled
psql "$DB_URL" -c "UPDATE provider_configs SET enabled = true WHERE id = '$RETRY_DEV_CFG_ID'::uuid;" 2>/dev/null

echo ""
echo "[16/16] Overall E2E Verification Summary"
echo "========================================="
echo "  ALL E2E EXPECTATION CRITERIA PASSED!"
echo "========================================="



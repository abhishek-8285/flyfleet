#!/usr/bin/env bash
set -euo pipefail

RUNTIME_GATE_DISABLED="${RUNTIME_GATE_DISABLED:-0}"
if [[ "$RUNTIME_GATE_DISABLED" == "1" ]]; then
  echo "runtime_integration_gate=SKIPPED reason=RUNTIME_GATE_DISABLED"
  exit 0
fi

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
BASE_URL="${BASE_URL:-http://localhost:8081}"
KEEP_STACK="${KEEP_STACK:-0}"
PRE_CLEAN="${PRE_CLEAN:-1}"
BUILD_STACK="${BUILD_STACK:-1}"
RUN_MIGRATIONS="${RUN_MIGRATIONS:-1}"
STRICT_MIGRATIONS="${STRICT_MIGRATIONS:-0}"
AUTH_EMAIL="${AUTH_EMAIL:-admin@flyfleet.com}"
: "${AUTH_PASSWORD:?ERROR: AUTH_PASSWORD environment variable is required. Set it with your admin password.}"
: "${JWT_SECRET:?ERROR: JWT_SECRET environment variable is required. Set it with a strong secret key.}"
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:8080}"

SERVICES=(
  postgres
  redis
  nats
  auth-service
  fleet-core-service
  alert-service
  admin-support-service
  billing-service
  tracking-service
  reports-service
  settings-service
  integrations-service
  web-bff
)

CONTAINER_NAMES=(
  fleet-postgres
  fleet-redis
  fleet-nats
  fleet-auth-service
  fleet-core-service
  fleet-alert-service
  fleet-admin-support-service
  fleet-billing-service
  fleet-tracking-service
  fleet-reports-service
  fleet-settings-service
  fleet-integrations-service
  fleet-web-bff
)

compose_cmd() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    docker compose -f "$COMPOSE_FILE" "$@"
    return
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose -f "$COMPOSE_FILE" "$@"
    return
  fi
  echo "error=docker_compose_not_available"
  exit 1
}

remove_conflicting_containers() {
  if ! command -v docker >/dev/null 2>&1; then
    return
  fi
  for c in "${CONTAINER_NAMES[@]}"; do
    if docker ps -a --format '{{.Names}}' | grep -qx "$c"; then
      docker rm -f "$c" >/dev/null 2>&1 || true
    fi
  done
}

run_migrations() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "migrations=skipped_no_docker"
    return
  fi

  if ! docker ps --format '{{.Names}}' | grep -qx "fleet-postgres"; then
    echo "migrations=skipped_no_postgres_container"
    return
  fi

  echo "migrations=running"
  shopt -s nullglob
  # Run baseline fix
  echo "migration_file=fix_all_missing_schema.sql"
  cat scripts/fix_all_missing_schema.sql | docker exec -i fleet-postgres psql -v ON_ERROR_STOP=1 -U fleet_user -d fleet_db >/dev/null
  # Run standard migrations
  for file in infrastructure/migrations/*.sql; do
    echo "migration_file=$(basename "$file")"
    if ! cat "$file" | docker exec -i fleet-postgres psql -v ON_ERROR_STOP=1 -U fleet_user -d fleet_db >/dev/null; then
      echo "migration_status=FAILED file=$(basename "$file")"
      if [[ "$STRICT_MIGRATIONS" == "1" ]]; then
        return 1
      fi
    fi
  done
  # Run runtime gate seeds
  echo "migration_file=seed_runtime_gate.sql"
  cat scripts/seed_runtime_gate.sql | docker exec -i fleet-postgres psql -v ON_ERROR_STOP=1 -U fleet_user -d fleet_db >/dev/null
}

cleanup() {
  if [[ "$KEEP_STACK" == "1" ]]; then
    echo "stack_cleanup=skipped"
    return
  fi
  echo "stack_cleanup=running"
  compose_cmd down >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "runtime_gate=starting"
echo "compose_file=$COMPOSE_FILE"
export JWT_SECRET

if [[ "$PRE_CLEAN" == "1" ]]; then
  echo "stack_preclean=running"
  compose_cmd down --remove-orphans >/dev/null 2>&1 || true
  remove_conflicting_containers
fi

if [[ "$BUILD_STACK" == "1" ]]; then
  echo "stack_build=running"
  compose_cmd build web-bff auth-service fleet-core-service alert-service admin-support-service billing-service tracking-service reports-service settings-service integrations-service >/dev/null
fi

compose_cmd up -d --force-recreate "${SERVICES[@]}"

if [[ "$RUN_MIGRATIONS" == "1" ]]; then
  run_migrations
fi

# Wait for web-bff health
for i in {1..90}; do
  code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || true)
  if [[ "$code" == "200" ]]; then
    echo "web_bff_health=PASS"
    break
  fi
  sleep 2
  if [[ "$i" == "90" ]]; then
    echo "web_bff_health=FAIL"
    compose_cmd ps
    exit 1
  fi
done

ALLOW_404=1 bash scripts/smoke_bff.sh "$BASE_URL"
bash scripts/check_api_latency_smoke.sh "$BASE_URL" 1200

# Authenticated smoke against protected routes
echo "# authenticated smoke"

login_resp=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$AUTH_EMAIL\",\"password\":\"$AUTH_PASSWORD\"}" || true)

token=$(python3 - <<'PY' "$login_resp"
import json
import sys

try:
    data = json.loads(sys.argv[1] or "{}")
except Exception:
    print("")
    raise SystemExit(0)

print(data.get("token") or data.get("access_token") or "")
PY
)

if [[ -z "$token" ]]; then
  echo "auth_login=FAIL"
  echo "login_response=$login_resp"
  exit 1
fi

echo "auth_login=PASS"

echo "# service health integration matrix"
SERVICE_HEALTH_ENDPOINTS=(
  "auth-service|http://localhost:8080/health"
  "fleet-core-service|http://localhost:8084/health"
  "alert-service|http://localhost:8085/health"
  "admin-support-service|http://localhost:8086/health"
  "billing-service|http://localhost:8087/health"
  "tracking-service|http://localhost:8082/health"
  "reports-service|http://localhost:8083/health"
  "integrations-service|http://localhost:8090/health"
  "settings-service|http://localhost:8091/health"
  "web-bff|http://localhost:8081/health"
)

service_health_fail=0
for entry in "${SERVICE_HEALTH_ENDPOINTS[@]}"; do
  name="${entry%%|*}"
  url="${entry##*|}"
  code=$(curl -sS -o /tmp/runtime_service_health_body.txt -w "%{http_code}" "$url" || true)
  echo "service_health $name code=$code"
  if [[ "$code" == "000" ]]; then
    service_health_fail=1
    continue
  fi
  if [[ "$code" =~ ^[0-9]{3}$ ]] && [[ "$code" -ge 500 ]]; then
    service_health_fail=1
  fi
done
if [[ "$service_health_fail" -ne 0 ]]; then
  echo "service_health_matrix=FAIL"
  exit 1
fi
echo "service_health_matrix=PASS"

# Ensure authenticated org exists in fleet-core DB for FK-safe rollout tests.
org_id=$(python3 - <<'PY' "$token"
import base64
import json
import sys

token = (sys.argv[1] or "").strip()
parts = token.split(".")
if len(parts) < 2:
    print("")
    raise SystemExit(0)

payload = parts[1]
padding = "=" * (-len(payload) % 4)
try:
    data = json.loads(base64.urlsafe_b64decode(payload + padding).decode("utf-8"))
except Exception:
    print("")
    raise SystemExit(0)

print(data.get("org_id", ""))
PY
)

if [[ -n "$org_id" ]] && command -v docker >/dev/null 2>&1; then
  docker exec fleet-postgres psql -U fleet_user -d fleet_db -c \
    "INSERT INTO organizations (id, name, domain, status) VALUES ('$org_id', 'Runtime Gate Org', 'runtime-gate.local', 'active') ON CONFLICT (id) DO NOTHING;" >/dev/null
  echo "runtime_org_seed=PASS org_id=$org_id"
else
  echo "runtime_org_seed=SKIPPED"
fi

echo "# firmware rollout lifecycle smoke"

firmware_payload=$(python3 - <<'PY'
import json, time
ts = int(time.time())
print(json.dumps({
    "version": f"ci-{ts}",
    "hardware_target": "tracker-v2",
    "file_url": "https://example.invalid/firmware/ci.bin",
    "file_size": 1048576,
    "checksum": "ci-checksum",
    "release_notes": "runtime integration rollout test",
    "is_critical": False
}))
PY
)

firmware_code=$(curl -sS -o /tmp/runtime_firmware_body.txt -w "%{http_code}" \
  -X POST "$BASE_URL/api/v1/firmware" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d "$firmware_payload" || true)
echo "firmware_create_code=$firmware_code"
if [[ "$firmware_code" != "201" ]]; then
  echo "firmware_lifecycle=FAIL stage=create_firmware"
  cat /tmp/runtime_firmware_body.txt
  exit 1
fi

firmware_id=$(python3 - <<'PY'
import json, pathlib
data = json.loads(pathlib.Path("/tmp/runtime_firmware_body.txt").read_text(encoding="utf-8"))
print(data.get("id", ""))
PY
)
if [[ -z "$firmware_id" ]]; then
  echo "firmware_lifecycle=FAIL stage=parse_firmware_id"
  cat /tmp/runtime_firmware_body.txt
  exit 1
fi

rollout_payload=$(python3 - <<'PY' "$firmware_id"
import json, sys, time
firmware_id = sys.argv[1]
ts = int(time.time())
print(json.dumps({
    "firmware_id": firmware_id,
    "name": f"CI Rollout {ts}",
    "device_group": "all",
    "total_devices": 3
}))
PY
)

rollout_code=$(curl -sS -o /tmp/runtime_rollout_body.txt -w "%{http_code}" \
  -X POST "$BASE_URL/api/v1/firmware/rollouts" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d "$rollout_payload" || true)
echo "rollout_create_code=$rollout_code"
if [[ "$rollout_code" != "201" ]]; then
  echo "firmware_lifecycle=FAIL stage=create_rollout"
  cat /tmp/runtime_rollout_body.txt
  exit 1
fi

rollout_id=$(python3 - <<'PY'
import json, pathlib
data = json.loads(pathlib.Path("/tmp/runtime_rollout_body.txt").read_text(encoding="utf-8"))
print(data.get("id", ""))
PY
)
if [[ -z "$rollout_id" ]]; then
  echo "firmware_lifecycle=FAIL stage=parse_rollout_id"
  cat /tmp/runtime_rollout_body.txt
  exit 1
fi

rollout_get_code=$(curl -sS -o /tmp/runtime_rollout_get_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id" || true)
echo "rollout_get_code=$rollout_get_code"
if [[ "$rollout_get_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=get_rollout"
  cat /tmp/runtime_rollout_get_body.txt
  exit 1
fi

transition_statuses=(in_progress paused in_progress completed)
for status in "${transition_statuses[@]}"; do
  payload="{\"status\":\"$status\"}"
  if [[ "$status" == "completed" ]]; then
    payload="{\"status\":\"$status\",\"reason\":\"runtime gate completion\"}"
  fi
  code=$(curl -sS -o /tmp/runtime_rollout_transition_body.txt -w "%{http_code}" \
    -X PATCH "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/status" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$payload" || true)
  echo "rollout_transition status=$status code=$code"
  if [[ "$code" != "200" ]]; then
    echo "firmware_lifecycle=FAIL stage=transition status=$status"
    cat /tmp/runtime_rollout_transition_body.txt
    exit 1
  fi
done

events_code=$(curl -sS -o /tmp/runtime_rollout_events_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20" || true)
echo "rollout_events_code=$events_code"
if [[ "$events_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_events"
  cat /tmp/runtime_rollout_events_body.txt
  exit 1
fi

events_count=$(python3 - <<'PY'
import json, pathlib
try:
    data = json.loads(pathlib.Path("/tmp/runtime_rollout_events_body.txt").read_text(encoding="utf-8"))
except Exception:
    print("0")
    raise SystemExit(0)
print(len(data) if isinstance(data, list) else 0)
PY
)
echo "rollout_events_count=$events_count"
if [[ "${events_count:-0}" -lt 2 ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_events_count"
  cat /tmp/runtime_rollout_events_body.txt
  exit 1
fi

limit_max_boundary_code=$(curl -sS -o /tmp/runtime_rollout_limit_max_boundary_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=500&offset=1" || true)
echo "rollout_limit_max_boundary_code=$limit_max_boundary_code"
if [[ "$limit_max_boundary_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_limit_max_boundary"
  cat /tmp/runtime_rollout_limit_max_boundary_body.txt
  exit 1
fi

has_reason=$(python3 - <<'PY'
import json, pathlib
try:
    data = json.loads(pathlib.Path("/tmp/runtime_rollout_events_body.txt").read_text(encoding="utf-8"))
except Exception:
    print("0")
    raise SystemExit(0)
flag = 0
if isinstance(data, list):
    for event in data:
        meta = event.get("metadata") if isinstance(event, dict) else None
        if isinstance(meta, dict) and str(meta.get("reason", "")).strip():
            flag = 1
            break
print(flag)
PY
)
echo "rollout_events_reason_present=$has_reason"
if [[ "$has_reason" != "1" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_event_reason_missing"
  cat /tmp/runtime_rollout_events_body.txt
  exit 1
fi

note_code=$(curl -sS -o /tmp/runtime_rollout_note_body.txt -w "%{http_code}" \
  -X POST "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"message":"runtime gate operator note","metadata":{"source":"runtime-gate"}}' || true)
echo "rollout_note_code=$note_code"
if [[ "$note_code" != "201" ]]; then
  echo "firmware_lifecycle=FAIL stage=add_rollout_note"
  cat /tmp/runtime_rollout_note_body.txt
  exit 1
fi

note_events_code=$(curl -sS -o /tmp/runtime_rollout_note_events_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=note" || true)
echo "rollout_note_events_code=$note_events_code"
if [[ "$note_events_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_note_events"
  cat /tmp/runtime_rollout_note_events_body.txt
  exit 1
fi

note_filter_check=$(python3 - <<'PY'
import json, pathlib
try:
    data = json.loads(pathlib.Path("/tmp/runtime_rollout_note_events_body.txt").read_text(encoding="utf-8"))
except Exception:
    print("0 0")
    raise SystemExit(0)
if not isinstance(data, list):
    print("0 0")
    raise SystemExit(0)
count = len(data)
all_note = int(count > 0 and all(isinstance(event, dict) and event.get("event_type") == "rollout.note" for event in data))
print(f"{count} {all_note}")
PY
)
note_events_count="$(echo "$note_filter_check" | awk '{print $1}')"
note_events_all_note="$(echo "$note_filter_check" | awk '{print $2}')"
echo "rollout_note_events_count=$note_events_count"
echo "rollout_note_events_all_note=$note_events_all_note"
if [[ "${note_events_count:-0}" -lt 1 || "$note_events_all_note" != "1" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_note_filter"
  cat /tmp/runtime_rollout_note_events_body.txt
  exit 1
fi

note_events_canonical_code=$(curl -sS -o /tmp/runtime_rollout_note_events_canonical_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=rollout.note" || true)
echo "rollout_note_events_canonical_code=$note_events_canonical_code"
if [[ "$note_events_canonical_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_note_events_canonical"
  cat /tmp/runtime_rollout_note_events_canonical_body.txt
  exit 1
fi

transition_events_code=$(curl -sS -o /tmp/runtime_rollout_transition_events_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=transition" || true)
echo "rollout_transition_events_code=$transition_events_code"
if [[ "$transition_events_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_transition_events"
  cat /tmp/runtime_rollout_transition_events_body.txt
  exit 1
fi

transition_filter_check=$(python3 - <<'PY'
import json, pathlib
try:
    data = json.loads(pathlib.Path("/tmp/runtime_rollout_transition_events_body.txt").read_text(encoding="utf-8"))
except Exception:
    print("0 0")
    raise SystemExit(0)
if not isinstance(data, list):
    print("0 0")
    raise SystemExit(0)
count = len(data)
all_transition = int(count > 0 and all(isinstance(event, dict) and event.get("event_type") == "rollout.transition" for event in data))
print(f"{count} {all_transition}")
PY
)
transition_events_count="$(echo "$transition_filter_check" | awk '{print $1}')"
transition_events_all_transition="$(echo "$transition_filter_check" | awk '{print $2}')"
echo "rollout_transition_events_count=$transition_events_count"
echo "rollout_transition_events_all_transition=$transition_events_all_transition"
if [[ "${transition_events_count:-0}" -lt 1 || "$transition_events_all_transition" != "1" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_transition_filter"
  cat /tmp/runtime_rollout_transition_events_body.txt
  exit 1
fi

transition_events_canonical_code=$(curl -sS -o /tmp/runtime_rollout_transition_events_canonical_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=rollout.transition" || true)
echo "rollout_transition_events_canonical_code=$transition_events_canonical_code"
if [[ "$transition_events_canonical_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_transition_events_canonical"
  cat /tmp/runtime_rollout_transition_events_canonical_body.txt
  exit 1
fi

created_events_code=$(curl -sS -o /tmp/runtime_rollout_created_events_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=created" || true)
echo "rollout_created_events_code=$created_events_code"
if [[ "$created_events_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_created_events"
  cat /tmp/runtime_rollout_created_events_body.txt
  exit 1
fi

created_filter_check=$(python3 - <<'PY'
import json, pathlib
try:
    data = json.loads(pathlib.Path("/tmp/runtime_rollout_created_events_body.txt").read_text(encoding="utf-8"))
except Exception:
    print("0 0")
    raise SystemExit(0)
if not isinstance(data, list):
    print("0 0")
    raise SystemExit(0)
count = len(data)
all_created = int(count > 0 and all(isinstance(event, dict) and event.get("event_type") == "rollout.created" for event in data))
print(f"{count} {all_created}")
PY
)
created_events_count="$(echo "$created_filter_check" | awk '{print $1}')"
created_events_all_created="$(echo "$created_filter_check" | awk '{print $2}')"
echo "rollout_created_events_count=$created_events_count"
echo "rollout_created_events_all_created=$created_events_all_created"
if [[ "${created_events_count:-0}" -lt 1 || "$created_events_all_created" != "1" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_created_filter"
  cat /tmp/runtime_rollout_created_events_body.txt
  exit 1
fi

created_events_canonical_code=$(curl -sS -o /tmp/runtime_rollout_created_events_canonical_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=20&event_type=rollout.created" || true)
echo "rollout_created_events_canonical_code=$created_events_canonical_code"
if [[ "$created_events_canonical_code" != "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=list_rollout_created_events_canonical"
  cat /tmp/runtime_rollout_created_events_canonical_body.txt
  exit 1
fi

invalid_event_filter_code=$(curl -sS -o /tmp/runtime_rollout_bad_event_filter_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?event_type=invalid_value" || true)
echo "rollout_invalid_event_filter_code=$invalid_event_filter_code"
if [[ "$invalid_event_filter_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_invalid_event_filter_guard"
  cat /tmp/runtime_rollout_bad_event_filter_body.txt
  exit 1
fi

invalid_limit_code=$(curl -sS -o /tmp/runtime_rollout_bad_limit_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=0" || true)
echo "rollout_invalid_limit_code=$invalid_limit_code"
if [[ "$invalid_limit_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_invalid_limit_guard"
  cat /tmp/runtime_rollout_bad_limit_body.txt
  exit 1
fi

limit_too_large_code=$(curl -sS -o /tmp/runtime_rollout_limit_too_large_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=501" || true)
echo "rollout_limit_too_large_code=$limit_too_large_code"
if [[ "$limit_too_large_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_limit_too_large_guard"
  cat /tmp/runtime_rollout_limit_too_large_body.txt
  exit 1
fi

invalid_offset_code=$(curl -sS -o /tmp/runtime_rollout_bad_offset_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?offset=-1" || true)
echo "rollout_invalid_offset_code=$invalid_offset_code"
if [[ "$invalid_offset_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_invalid_offset_guard"
  cat /tmp/runtime_rollout_bad_offset_body.txt
  exit 1
fi

invalid_limit_format_code=$(curl -sS -o /tmp/runtime_rollout_bad_limit_format_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?limit=abc" || true)
echo "rollout_invalid_limit_format_code=$invalid_limit_format_code"
if [[ "$invalid_limit_format_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_invalid_limit_format_guard"
  cat /tmp/runtime_rollout_bad_limit_format_body.txt
  exit 1
fi

invalid_offset_format_code=$(curl -sS -o /tmp/runtime_rollout_bad_offset_format_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/events?offset=abc" || true)
echo "rollout_invalid_offset_format_code=$invalid_offset_format_code"
if [[ "$invalid_offset_format_code" != "400" ]]; then
  echo "firmware_lifecycle=FAIL stage=rollout_invalid_offset_format_guard"
  cat /tmp/runtime_rollout_bad_offset_format_body.txt
  exit 1
fi

# completed -> scheduled must fail
invalid_code=$(curl -sS -o /tmp/runtime_rollout_invalid_transition_body.txt -w "%{http_code}" \
  -X PATCH "$BASE_URL/api/v1/firmware/rollouts/$rollout_id/status" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"status":"scheduled"}' || true)
echo "rollout_invalid_transition code=$invalid_code"
if [[ "$invalid_code" == "200" ]]; then
  echo "firmware_lifecycle=FAIL stage=invalid_transition_guard"
  cat /tmp/runtime_rollout_invalid_transition_body.txt
  exit 1
fi
echo "firmware_lifecycle=PASS"

ota_orchestration_output="$(python3 scripts/ota_rollout_orchestrator.py || true)"
echo "$ota_orchestration_output"
if ! grep -q "ota_rollout_orchestration_simulation=PASS" <<<"$ota_orchestration_output"; then
  echo "ota_orchestration_runtime=FAIL"
  exit 1
fi
echo "ota_orchestration_runtime=PASS"

settings_export_code=$(curl -sS -o /tmp/runtime_settings_export_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/admin/settings/export" || true)
echo "settings_export_code=$settings_export_code"
if [[ "$settings_export_code" != "200" ]]; then
  echo "settings_export_runtime=FAIL"
  cat /tmp/runtime_settings_export_body.txt
  exit 1
fi
if ! python3 - <<'PY'
import json, sys
with open("/tmp/runtime_settings_export_body.txt", "r", encoding="utf-8") as f:
    payload = json.load(f)
if "platform_settings" not in payload or "exported_at" not in payload:
    sys.exit(1)
print("settings_export_semantics=PASS")
PY
then
  echo "settings_export_semantics=FAIL"
  cat /tmp/runtime_settings_export_body.txt
  exit 1
fi

settings_import_code=$(curl -sS -o /tmp/runtime_settings_import_body.txt -w "%{http_code}" \
  -X POST "$BASE_URL/api/v1/admin/settings/import" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"feature_flags":[{"key":"runtime-gate-flag","scope":"GLOBAL","enabled":true}],"replace_existing":false}' || true)
echo "settings_import_code=$settings_import_code"
if [[ "$settings_import_code" != "200" ]]; then
  echo "settings_import_runtime=FAIL"
  cat /tmp/runtime_settings_import_body.txt
  exit 1
fi
if ! python3 - <<'PY'
import json, sys
with open("/tmp/runtime_settings_import_body.txt", "r", encoding="utf-8") as f:
    payload = json.load(f)
required = {"platform_updated", "feature_flags_created", "feature_flags_updated", "feature_flags_deleted"}
if not required.issubset(payload.keys()):
    sys.exit(1)
print("settings_import_semantics=PASS")
PY
then
  echo "settings_import_semantics=FAIL"
  cat /tmp/runtime_settings_import_body.txt
  exit 1
fi

settings_import_invalid_code=$(curl -sS -o /tmp/runtime_settings_import_invalid_body.txt -w "%{http_code}" \
  -X POST "$BASE_URL/api/v1/admin/settings/import" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{}' || true)
echo "settings_import_invalid_code=$settings_import_invalid_code"
if [[ "$settings_import_invalid_code" != "400" ]]; then
  echo "settings_import_validation_runtime=FAIL"
  cat /tmp/runtime_settings_import_invalid_body.txt
  exit 1
fi

settings_audit_invalid_limit_code=$(curl -sS -o /tmp/runtime_settings_audit_invalid_limit_body.txt -w "%{http_code}" \
  -H "Authorization: Bearer $token" \
  "$BASE_URL/api/v1/admin/settings/audit?limit=abc" || true)
echo "settings_audit_invalid_limit_code=$settings_audit_invalid_limit_code"
if [[ "$settings_audit_invalid_limit_code" != "400" ]]; then
  echo "settings_audit_validation_runtime=FAIL"
  cat /tmp/runtime_settings_audit_invalid_limit_body.txt
  exit 1
fi

AUTH_ENDPOINTS=(
  "/api/v1/dashboard/stats|200"
  "/api/v1/admin/governance/requests|200"
  "/api/v1/admin/work-orders|200"
  "/api/v1/admin/devices?page=1&size=5|200"
  "/api/v1/admin/notifications/logs|200"
  "/api/v1/admin/billing/invoices|200"
  "/api/v1/admin/dr/backups|200"
  "/api/v1/admin/dr/tests|200"
  "/api/v1/admin/dr/tests/metrics|200"
  "/api/v1/admin/dr/tests/settings|200"
  "/api/v1/admin/reports/schedules|200"
  "/api/v1/admin/reports/executions|200"
  "/api/v1/admin/settings/platform|200"
  "/api/v1/admin/settings/audit?limit=10|200"
  "/api/v1/admin/finance/expenses/reports?limit=5|200"
  "/api/v1/admin/revenue/forecast?route_id=BLR-MYS|200"
  "/api/v1/search?q=test|200"
  "/api/v1/timeline|200"
  "/api/devices/summary|200"
  "/api/devices/twins|200"
)

auth_fail=0
for entry in "${AUTH_ENDPOINTS[@]}"; do
  path="${entry%%|*}"
  expected="${entry##*|}"
  code=$(curl -sS -o /tmp/runtime_auth_body.txt -w "%{http_code}" \
    -H "Authorization: Bearer $token" \
    "$BASE_URL$path" || true)
  printf 'auth %-42s -> %s (expected %s)\n' "$path" "$code" "$expected"
  if [[ "$code" != "$expected" ]]; then
    auth_fail=1
    body="$(tr '\n' ' ' < /tmp/runtime_auth_body.txt | cut -c1-240)"
    printf 'auth_response_preview %s -> %s\n' "$path" "$body"
  fi
done

if [[ "$auth_fail" -ne 0 ]]; then
  echo "authenticated_smoke=FAIL"
  exit 1
fi

echo "authenticated_smoke=PASS"

echo "runtime_integration_gate=PASS"

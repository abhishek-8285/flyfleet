import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";
import { check, sleep } from "k6";
import http from "k6/http";

export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp up to 10 users
    { duration: "1m", target: 10 }, // Stay at 10 users (10 simultaneous SOS)
    { duration: "30s", target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<200"], // 95% of requests must complete below 200ms
    http_req_failed: ["rate<0.01"], // Error rate should be less than 1%
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080/api/v1/emergency";
const ORG_ID = "00000000-0000-0000-0000-000000000001";

export default function () {
  const vehicleId = uuidv4();
  const deviceId = uuidv4();

  // 1. Trigger SOS (Standard)
  const standardPayload = JSON.stringify({
    vehicle_id: vehicleId,
    incident_type: "SOS_PANIC",
    panic_source: "DRIVER_MOBILE",
    latitude: 28.6139,
    longitude: 77.209,
    metadata: { battery: "85%" },
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      "X-Org-ID": ORG_ID,
    },
  };

  let res = http.post(`${BASE_URL}/sos-trigger`, standardPayload, params);
  check(res, {
    "standard status 200 or 409": (r) => r.status === 200 || r.status === 409,
  });

  // 2. Trigger SOS (AIS-140 Strict Validation)
  const aisPayload = JSON.stringify({
    device_id: deviceId,
    vehicle_id: vehicleId,
    lat: 19.076,
    lng: 72.8777,
    timestamp: Date.now(),
    panic_flag: true,
  });

  res = http.post(`${BASE_URL}/sos-trigger/ais140`, aisPayload, params);
  check(res, {
    "ais140 status 202 or 409": (r) => r.status === 202 || r.status === 409,
    "ais140 latency < 100ms": (r) => r.timings.duration < 100,
  });

  // 3. Test Invalid AIS-140 (Missing Flag)
  const invalidAisPayload = JSON.stringify({
    device_id: deviceId,
    vehicle_id: vehicleId,
    lat: 19.076,
    lng: 72.8777,
    timestamp: Date.now(),
    panic_flag: false,
  });

  res = http.post(`${BASE_URL}/sos-trigger/ais140`, invalidAisPayload, params);
  check(res, {
    "invalid ais140 status 400": (r) => r.status === 400,
  });

  // 4. List Active Incidents & Timeline
  const listRes = http.get(`${BASE_URL}/incidents`, params);
  check(listRes, {
    "list status 200": (r) => r.status === 200,
  });

  sleep(Math.random() * 2 + 1);
}

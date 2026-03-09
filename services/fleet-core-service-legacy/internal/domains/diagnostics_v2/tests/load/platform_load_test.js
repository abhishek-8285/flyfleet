import { check, sleep } from "k6";
import http from "k6/http";

export let options = {
  stages: [
    { duration: "1m", target: 100 }, // ramp up to 100 users
    { duration: "3m", target: 500 }, // stay at 500 users
    { duration: "1m", target: 0 }, // ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests must be under 500ms
    http_req_failed: ["rate<0.01"], // error rate < 1%
  },
};

export default function () {
  const BASE_URL = "http://localhost:8080";

  // Simulate concurrent diagnostics telemetry
  let telemetryRes = http.post(
    `${BASE_URL}/diagnostics/devices/sim-batch/telemetry`,
    JSON.stringify({
      gps_lat: 12.0,
      gps_lon: 77.0,
      speed: 50,
      timestamp: new Date().toISOString(),
    }),
    {
      headers: {
        "Content-Type": "application/json",
        Authorization: "DeviceToken prod-scale-token",
      },
    },
  );

  check(telemetryRes, { "telemetry success": (r) => r.status === 201 });

  // Simulate concurrent portal dashboard access
  let dashboardRes = http.get(
    `${BASE_URL}/diagnostics/dashboard/fleet-health`,
    {
      headers: { Authorization: "Bearer portal-jwt-token" },
    },
  );

  check(dashboardRes, { "dashboard success": (r) => r.status === 200 });

  sleep(1);
}

import { check, sleep } from "k6";
import http from "k6/http";

export let options = {
  vus: 50, // 50 virtual users simulating devices
  duration: "30s",
  thresholds: {
    http_req_duration: ["p(95)<200"], // 95% of requests must complete below 200ms
    http_req_failed: ["rate<0.01"], // error rate must be < 1%
  },
};

export default function () {
  const url =
    "http://localhost:8080/diagnostics/devices/123e4567-e89b-12d3-a456-426614174000/telemetry";
  const payload = JSON.stringify({
    gps_lat: 12.9716,
    gps_lon: 77.5946,
    speed: 60,
    battery_voltage: 12.5,
    timestamp: new Date().toISOString(),
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      "Device-Token": "test-token",
    },
  };

  let res = http.post(url, payload, params);

  check(res, {
    "status is 201": (r) => r.status === 201,
  });

  sleep(1);
}

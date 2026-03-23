import { check } from "k6";
import http from "k6/http";

export const options = {
  scenarios: {
    driver_load: {
      executor: "constant-arrival-rate",
      rate: 100, // 100 behavior events per second
      timeUnit: "1s",
      duration: "5m",
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<200"], // 95% of requests must complete below 200ms
    http_req_failed: ["rate<0.01"], // Error rate must be < 1%
  },
};

const BASE_URL = "http://localhost:8080/api/v1";
const DRIVER_ID = "41b5650b-c777-4ac0-ba21-136cb5a45fd7"; // Test driver
const TOKEN =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzI5MDIzMjIsIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMSIsInVzZXJfaWQiOiJ0ZXN0LXVzZXItMTIzIn0.d6wYUqD34qJsaNMFn8amhMh86QqYXpabQBXLqtL80xk";

export default function () {
  const payload = JSON.stringify({
    event_type: "SPEEDING",
    speed: 90 + Math.floor(Math.random() * 20),
    timestamp: new Date().toISOString(),
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${TOKEN}`,
    },
  };

  const res = http.post(
    `${BASE_URL}/drivers/${DRIVER_ID}/update-behavior`,
    payload,
    params,
  );

  check(res, {
    "is status 200": (r) => r.status === 200,
  });
}

import { check, sleep } from "k6";
import http from "k6/http";

export const options = {
  stages: [
    { duration: "30s", target: 50 }, // Ramp up
    { duration: "1m", target: 500 }, // Spike to 500 concurrent users
    { duration: "30s", target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<200"], // 95% of requests must complete below 200ms
    http_req_failed: ["rate<0.01"], // Error rate must be < 1%
  },
};

export default function () {
  const payload = JSON.stringify({
    email: "loadtest@example.com",
    password: "StrongPassword123!",
  });
  const headers = { "Content-Type": "application/json" };
  const res = http.post("http://localhost:8080/api/v1/auth/login", payload, {
    headers,
  });

  check(res, {
    "is status 200": (r) => r.status === 200,
    "has access_token": (r) => r.json("access_token") !== undefined,
  });
  sleep(1);
}

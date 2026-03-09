import { check, sleep } from "k6";
import http from "k6/http";

export const options = {
  stages: [
    { duration: "2m", target: 500 }, // Ramp-up to 500 users
    { duration: "5m", target: 500 }, // Stay at 500 users
    { duration: "2m", target: 1000 }, // Push to 1000 users
    { duration: "5m", target: 1000 }, // Stay at 1000 users
    { duration: "2m", target: 0 }, // Ramp-down
  ],
  thresholds: {
    http_req_duration: ["p(95)<300"], // 95% of requests must be below 300ms
    http_req_failed: ["rate<0.01"], // Less than 1% failure rate
  },
};

const BASE_URL = __ENV.API_URL || "http://localhost:8080";

export default function () {
  const customerId = "00000000-0000-0000-0000-000000000000"; // Global Mock

  // 1. Dashboard View
  let res = http.get(`${BASE_URL}/portal/dashboard`, {
    headers: { "X-Customer-ID": customerId },
  });
  check(res, { "dashboard status is 200": (r) => r.status === 200 });

  sleep(1);

  // 2. Search Shipments
  res = http.get(`${BASE_URL}/portal/shipments/search?status=IN_TRANSIT`, {
    headers: { "X-Customer-ID": customerId },
  });
  check(res, { "search status is 200": (r) => r.status === 200 });

  sleep(2);

  // 3. Track Specific Shipment
  const shipmentId = "00000000-0000-0000-0000-000000000001";
  res = http.get(`${BASE_URL}/portal/shipments/${shipmentId}/tracking`, {
    headers: { "X-Customer-ID": customerId },
  });
  check(res, { "tracking status is 200": (r) => r.status === 200 });

  sleep(3);
}

import { check, sleep } from "k6";
import http from "k6/http";

export let options = {
  vus: 10, // Virtual Users
  duration: "30s",
};

const BASE_URL = "http://localhost:8080"; // Update to actual test server

export default function () {
  const payload = JSON.stringify({
    html_content: "<html><body><h1>Load Test Invoice</h1></body></html>",
    sample_data: { invoice_number: "LOAD-001" },
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  let res = http.post(
    `${BASE_URL}/api/v1/billing/invoices/preview`,
    payload,
    params,
  );

  check(res, {
    "status is 200": (r) => r.status === 200,
    "content type is pdf": (r) =>
      r.headers["Content-Type"] === "application/pdf",
  });

  sleep(1);
}

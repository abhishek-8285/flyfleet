// k6 load tests for Dispatch Flow endpoints
// Run: k6 run load-tests/dispatch_flow.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// ── Configuration ───────────────────────────────────────────────────────────

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 VUs
    { duration: '1m', target: 50 },    // Peak load: 50 concurrent dispatchers
    { duration: '30s', target: 100 },  // Stress test: 100 VUs
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<1000'],  // 95% of requests < 1s
    'http_req_failed': ['rate<0.01'],      // <1% error rate
    'errors': ['rate<0.05'],                // <5% custom error rate
  },
};

// ── Setup ───────────────────────────────────────────────────────────────────

export function setup() {
  return {
    baseUrl: __ENV.BASE_URL || 'http://localhost:8084',
    token: __ENV.JWT_TOKEN || 'test-token',
  };
}

// ── Test: Dispatch Suggestion ───────────────────────────────────────────────

export function suggestTest(data) {
  const res = http.post(`${data.baseUrl}/api/v1/dispatch/suggest`, JSON.stringify({
    shipment_id: generateUUID(),
    pickup_latitude: 28.6139,
    pickup_longitude: 77.2090,
    vehicle_type: 'LCV',
    cargo_weight_kg: 250.0,
    top_n: 3,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
  });

  check(res, {
    'suggest status 200': (r) => r.status === 200,
    'suggest < 500ms': (r) => r.timings.duration < 500,
    'suggest has suggestions': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.suggestions && body.suggestions.length > 0;
      } catch (e) {
        return false;
      }
    },
  });

  if (res.status !== 200) {
    errorRate.add(1);
  }

  sleep(1);
}

// ── Test: Dispatch Assign ───────────────────────────────────────────────────

export function assignTest(data) {
  const res = http.post(`${data.baseUrl}/api/v1/dispatch/assign`, JSON.stringify({
    shipment_id: generateUUID(),
    vehicle_id: generateUUID(),
    driver_id: generateUUID(),
    scheduled_start: new Date(Date.now() + 15 * 60000).toISOString(),
    auto_start: true,
    auto_complete: true,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
  });

  check(res, {
    'assign status 201': (r) => r.status === 201,
    'assign < 1s': (r) => r.timings.duration < 1000,
    'assign returns trip_id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.trip_id && body.trip_id !== '';
      } catch (e) {
        return false;
      }
    },
  });

  if (res.status !== 201) {
    errorRate.add(1);
  }

  sleep(2);
}

// ── Test: Dispatch Dashboard ────────────────────────────────────────────────

export function dashboardTest(data) {
  const res = http.get(`${data.baseUrl}/api/v1/dispatch/dashboard`, {
    headers: {
      'Authorization': `Bearer ${data.token}`,
    },
    tags: { name: 'dashboard' },
  });

  check(res, {
    'dashboard status 200': (r) => r.status === 200,
    'dashboard < 1s': (r) => r.timings.duration < 1000,
    'dashboard has summary': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.summary && body.summary.total_today !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (res.status !== 200) {
    errorRate.add(1);
  }

  sleep(5);
}

// ── Test: Bulk Assign (50 shipments) ───────────────────────────────────────

export function bulkAssignTest(data) {
  const assignments = [];
  for (let i = 0; i < 50; i++) {
    assignments.push({
      shipment_id: generateUUID(),
      vehicle_id: generateUUID(),
      driver_id: generateUUID(),
    });
  }

  const start = Date.now();
  const res = http.post(`${data.baseUrl}/api/v1/dispatch/bulk-assign`, JSON.stringify({
    assignments: assignments,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
    tags: { name: 'bulk_assign' },
  });
  const duration = Date.now() - start;

  check(res, {
    'bulk assign status 200': (r) => r.status === 200,
    'bulk assign < 10s': (r) => duration < 10000,
    'bulk assign all succeeded': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.succeeded === 50 && body.failed === 0;
      } catch (e) {
        return false;
      }
    },
  });

  if (res.status !== 200) {
    errorRate.add(1);
  }

  sleep(10);
}

// ── Test: Auth Failure ──────────────────────────────────────────────────────

export function authFailureTest(data) {
  const res = http.post(`${data.baseUrl}/api/v1/dispatch/assign`, JSON.stringify({
    shipment_id: generateUUID(),
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer invalid-token',
    },
  });

  check(res, {
    'auth failure returns 401': (r) => r.status === 401,
  });

  sleep(1);
}

// ── Helpers ─────────────────────────────────────────────────────────────────

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

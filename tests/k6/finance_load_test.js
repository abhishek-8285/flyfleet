import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';
const EXISTING_DRIVER_ID = __ENV.DRIVER_ID || 'e1a510a4-d48c-4b54-bb9e-191de2067d3d';

const commonHeaders = {
  'Content-Type': 'application/json',
};
if (AUTH_TOKEN) {
  commonHeaders.Authorization = `Bearer ${AUTH_TOKEN}`;
}

http.setResponseCallback(http.expectedStatuses({ min: 200, max: 299 }, 429));
const throttledRequests = new Counter('throttled_requests');

export const options = {
  discardResponseBodies: true,
  scenarios: {
    fuel_ingest: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 120,
      maxVUs: 300,
      exec: 'fuelIngest',
    },
    expense_submit: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 80,
      maxVUs: 200,
      exec: 'expenseSubmit',
    },
  },
  thresholds: {
    http_req_failed: ['rate==0'],
    http_req_duration: ['p(95)<200'],
  },
};

function uuidSeed(i) {
  return `00000000-0000-0000-0000-${String(i).padStart(12, '0')}`;
}

function randomUUIDv4() {
  const hex = () => Math.floor(Math.random() * 16).toString(16);
  let out = '';
  for (let i = 0; i < 32; i++) out += hex();
  return `${out.slice(0, 8)}-${out.slice(8, 12)}-4${out.slice(13, 16)}-a${out.slice(17, 20)}-${out.slice(20, 32)}`;
}

export function fuelIngest() {
  const payload = JSON.stringify({
    org_id: '00000000-0000-0000-0000-000000000001',
    vehicle_id: randomUUIDv4(),
    driver_id: EXISTING_DRIVER_ID,
    liters: 40,
    amount: 4000,
    currency: 'INR',
    transaction_date: new Date().toISOString(),
  });

  const res = http.post(`${BASE_URL}/api/v1/finance/fuel/ingest`, payload, { headers: commonHeaders });
  if (res.status === 429) throttledRequests.add(1);
  check(res, {
    'fuel ingest status is 2xx or 429': (r) => (r.status >= 200 && r.status < 300) || r.status === 429,
  });
}

let seq = 1;
export function expenseSubmit() {
  const payload = JSON.stringify({
    org_id: '00000000-0000-0000-0000-000000000001',
    submitter_id: uuidSeed(seq++),
    period_start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    period_end: new Date().toISOString(),
    total_amount: 1000,
    currency: 'INR',
    items: [
      {
        category: 'FUEL',
        amount: 1000,
        gst_amount: 180,
      },
    ],
  });

  const res = http.post(`${BASE_URL}/api/v1/finance/expenses/reports`, payload, { headers: commonHeaders });
  if (res.status === 429) throttledRequests.add(1);
  check(res, {
    'expense status is 2xx or 429': (r) => (r.status >= 200 && r.status < 300) || r.status === 429,
  });
}

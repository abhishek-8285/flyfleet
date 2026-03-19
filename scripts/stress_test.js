import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    scenarios: {
        telemetry: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 50000 },
                { duration: '2m', target: 100000 },
                { duration: '10m', target: 100000 },
                { duration: '2m', target: 0 },
            ],
            gracefulRampDown: '30s',
        },
        chaos_sos: {
            executor: 'constant-arrival-rate',
            rate: 1, // ~1 SOS per second across all VUs
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 10,
            startTime: '4m', // Start during peak
            maxVUs: 50,
        },
        chaos_revocation: {
            executor: 'constant-arrival-rate',
            rate: 2, // ~2 revocations per second
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 10,
            startTime: '4m',
            maxVUs: 50,
        }
    },
    thresholds: {
        'http_req_duration{scenario:telemetry}': ['p(95)<100'],
        'http_req_duration{scenario:chaos_sos}': ['p(95)<200'],
        'http_req_failed': ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    // Telemetry load
    const deviceId = `vehicle-${__VU}`;
    const payload = JSON.stringify({
        device_id: deviceId,
        timestamp: Date.now(),
        latitude: 12.9716 + (Math.random() - 0.5) * 0.1,
        longitude: 77.5946 + (Math.random() - 0.5) * 0.1,
        speed: Math.random() * 100,
        ignition: true,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Organization-ID': '00000000-0000-0000-0000-000000000001', // Simulation Org
        },
    };

    const res = http.post(`${BASE_URL}/api/v1/telemetry/gps`, payload, params);
    
    check(res, {
        'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    });

    // 1Hz GPS pings
    sleep(1);
}

export function chaos_sos() {
    const payload = JSON.stringify({
        device_id: `vehicle-${randomIntBetween(1, 100000)}`,
        vehicle_id: `VINFLEET-${randomString(8)}`,
        incident_type: 'panic',
        latitude: 12.9716,
        longitude: 77.5946,
    });

    const res = http.post(`${BASE_URL}/api/v1/emergency/sos-trigger`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(res, { 'sos accepted': (r) => r.status === 201 || r.status === 202 });
}

export function chaos_revocation() {
    const userId = '00000000-0000-0000-0000-000000000002';
    const sessionId = `revoked-session-${randomIntBetween(1, 1000)}`;
    
    const res = http.post(`${BASE_URL}/api/v1/auth/users/${userId}/sessions/${sessionId}/revoke`, null, {
        headers: { 'X-Secret-Chaos-Key': 'performance-test' },
    });

    check(res, { 'revocation accepted': (r) => r.status === 200 || r.status === 204 });
}

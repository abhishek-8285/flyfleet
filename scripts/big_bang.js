import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

export const options = {
    scenarios: {
        // Stage 1 & 2: The Network Firehose
        telemetry_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 50000 }, // Stage 1: Ramp 0 -> 50K
                { duration: '10m', target: 100000 }, // Stage 2: Maintain 100K -> Peak
            ],
            exec: 'gpsPing',
        },
        
        // Chaos Event: The Zombie Killer
        chaos_zombie_event: {
            executor: 'shared-iterations',
            vus: 500,
            iterations: 500,
            startTime: '5m', // Trigger exactly at Minute 5
            exec: 'revokeSession',
        },
        
        // Financial Shield: The Idempotency Double-Click Guard
        financial_shield: {
            executor: 'shared-iterations',
            vus: 100,
            iterations: 1000,
            startTime: '2m', // Start testing this after ramp-up is done
            exec: 'doubleClickTransaction',
        },
        
        // Automated SRE Defense: The Pool Shedding Check
        sre_audit_guard: {
            executor: 'constant-vus',
            vus: 1, // Single admin VU continuously polling
            duration: '12m',
            exec: 'pollAnalytics',
        }
    }
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8084';
const JWT_TOKEN = __ENV.JWT_TOKEN || 'test-jwt-token'; 
const ORG_ID = 'test-tenant-100k';

export function gpsPing() {
    const vuId = exec.vu.idInTest;
    
    const payload = JSON.stringify({
        vehicle_id: `veh-${vuId}`,
        timestamp: new Date().toISOString(),
        data: {
            latitude: 28.7041 + (Math.random() * 0.1),
            longitude: 77.1025 + (Math.random() * 0.1),
            speed: Math.floor(Math.random() * 80),
            heading: Math.floor(Math.random() * 360)
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${JWT_TOKEN}`,
            'X-Org-ID': ORG_ID,
            'X-Session-ID': `session-${vuId}` 
        }
    };

    const res = http.post(`${BASE_URL}/api/v1/device/telemetry/gps`, payload, params);
    
    // We expect 201 Created usually, but if the VU hit the 5m Zombie Chaos event, it will bounce back as 401.
    check(res, {
        'Telemetry ping accepted (2xx) or Revoked (401)': (r) => r.status === 201 || r.status === 200 || r.status === 401,
    });
    
    // 1Hz Ping Rate mapping to prompt requirement
    sleep(1);
}

export function revokeSession() {
    // Exactly 500 iterative sessions revoked here
    const targetVu = exec.scenario.iterationInTest + 1;

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${JWT_TOKEN}`,
            'X-Org-ID': ORG_ID,
        }
    };

    const payload = JSON.stringify({
        device_id: `session-${targetVu}`
    });

    const res = http.post(`${BASE_URL}/api/v1/auth/mfa/revoke`, payload, params);
    check(res, { 'Chaos Event - Session Revoked': (r) => r.status === 200 || r.status === 204 });
}

export function doubleClickTransaction() {
    const reqId = `fastag-request-${exec.vu.idInTest}-${exec.scenario.iterationInTest}`;
    
    const payload = JSON.stringify({
        amount: 500,
        toll_plaza_id: "T123",
        vehicle_id: "veh-test"
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${JWT_TOKEN}`,
            'X-Org-ID': ORG_ID,
            'X-Request-ID': reqId
        }
    };

    // Simulate clicking the "Pay FASTag" button 2 times instantly
    const requests = Array(2).fill({
        method: 'POST',
        url: `${BASE_URL}/api/v1/india/fastag/transaction`,
        body: payload,
        params: params
    });

    const responses = http.batch(requests);
    
    let okCount = 0;
    let conflictCount = 0;

    responses.forEach(r => {
        if (r.status === 200 || r.status === 201) okCount++;
        if (r.status === 409) conflictCount++;
    });

    // We verify Phase 4: Financial Shield strictly enforces 1 execution, 1 reject.
    check(okCount, { 'Financial Shield: Exactly 1 Success': (v) => v === 1 });
    check(conflictCount, { 'Financial Shield: Exactly 1 Duplicate Bridled (409)': (v) => v === 1 });
    
    sleep(1);
}

export function pollAnalytics() {
    const params = {
        headers: {
            'Authorization': `Bearer ${JWT_TOKEN}`,
            'X-Org-ID': ORG_ID,
        }
    };
    
    const res = http.get(`${BASE_URL}/api/v1/analytics/reports`, params);
    
    // As the critical pool hits >90% during 100K load, Phase 5 SRE guard should shed this request
    check(res, {
        'SRE Guard: Analytics served (200) or gracefully shed (429)': (r) => r.status === 200 || r.status === 429
    });

    // Check performance constraints to prevent cascade failures
    check(res, {
        'Critical Latency Bound Check': (r) => r.timings.duration < 1000 // Ensure shedding is fast
    });

    sleep(2);
}

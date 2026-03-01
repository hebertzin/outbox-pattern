/**
 * Users Service — k6 Load Test
 *
 * Scenarios (sequential):
 *   1. Smoke      : 1 VU × 30 s — sanity check before load
 *   2. Steady     : ramp 0→50 VUs in 30 s, hold 4 min, ramp down 30 s
 *   3. Spike      : ramp 0→200 VUs in 10 s, hold 30 s, ramp down 10 s
 *
 * Thresholds enforced globally:
 *   • p90 response time < 30 ms
 *   • error rate        < 1 %
 *
 * Run locally (service must be up via docker-compose):
 *   k6 run tests/load/load-test.js
 *   k6 run tests/load/load-test.js -e BASE_URL=http://localhost:8080
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ── Config ──────────────────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://app:8080';

const HEADERS = { 'Content-Type': 'application/json' };

// ── Custom metrics ───────────────────────────────────────────────────────────

const errorRate   = new Rate('error_rate');
const createTrend = new Trend('duration_create_user', true);

// ── Scenario / threshold config ──────────────────────────────────────────────

export const options = {
  scenarios: {
    // 1. Smoke — 1 VU for 30 s
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
      tags: { scenario: 'smoke' },
    },

    // 2. Steady load — ramp up to 50 VUs, hold, ramp down
    steady: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '4m',  target: 50 },
        { duration: '30s', target: 0  },
      ],
      startTime: '31s',
      tags: { scenario: 'steady' },
    },

    // 3. Spike — ramp from 0 to 200 VUs in 10 s
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 200 },
        { duration: '30s', target: 200 },
        { duration: '10s', target: 0   },
      ],
      startTime: '6m32s',
      tags: { scenario: 'spike' },
    },
  },

  thresholds: {
    // Global p90 and error-rate targets
    'http_req_duration': ['p(90) < 30'],
    'error_rate':        ['rate < 0.01'],

    // Per-endpoint latency budget
    'duration_create_user': ['p(90) < 30'],
  },
};

// ── Default function ─────────────────────────────────────────────────────────

export default function main() {
  createUser();

  // ~10 ms think-time between requests
  sleep(0.01);
}

// ── Request helpers ──────────────────────────────────────────────────────────

function createUser() {
  const payload = JSON.stringify({
    email:    `loadtest-${uuidv4()}@example.com`,
    password: 'loadtestpassword123',
  });

  const res = http.post(`${BASE_URL}/api/v1/users`, payload, { headers: HEADERS });

  createTrend.add(res.timings.duration);

  const ok = check(res, {
    'create: status is 201': (r) => r.status === 201,
    'create: has id': (r) => {
      try { return JSON.parse(r.body).data.id !== ''; } catch { return false; }
    },
  });

  errorRate.add(!ok);
}

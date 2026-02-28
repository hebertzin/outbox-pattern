/**
 * Transaction Service — k6 Load Test
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
 *
 * Run via Docker (from transaction-service/):
 *   docker compose -f docker-compose.yml -f docker-compose.load-test.yml run --rm k6
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { randomItem, uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ── Config ──────────────────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://app:8080';

const HEADERS = { 'Content-Type': 'application/json' };

// Pre-defined user pool — keeps foreign-key constraints simple
const USER_IDS = [
  'load-user-a', 'load-user-b', 'load-user-c',
  'load-user-d', 'load-user-e', 'load-user-f',
];

// ── Custom metrics ───────────────────────────────────────────────────────────

const errorRate   = new Rate('error_rate');
const createTrend = new Trend('duration_create_transaction', true);
const statusTrend = new Trend('duration_get_status',        true);
const balanceTrend = new Trend('duration_get_balance',      true);

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

    // Per-endpoint latency budgets
    'duration_create_transaction': ['p(90) < 30'],
    'duration_get_status':         ['p(90) < 30'],
    'duration_get_balance':        ['p(90) < 30'],
  },
};

// ── Setup — seed a transaction ID that GET /status tests can use ──────────────

export function setup() {
  const fromUser = USER_IDS[0];
  const toUser   = USER_IDS[1];

  const res = http.post(
    `${BASE_URL}/api/v1/transactions`,
    JSON.stringify({ from_user_id: fromUser, to_user_id: toUser, amount: 1000 }),
    { headers: HEADERS },
  );

  const ok = res.status === 201 || res.status === 200;
  if (!ok) {
    console.warn(`setup: seed transaction failed — status ${res.status}`);
    return { txID: '' };
  }

  const body = JSON.parse(res.body);
  const txID = body.data ? body.data.id : '';
  console.info(`setup: seeded transaction id=${txID}`);
  return { txID };
}

// ── Default function — traffic mix ───────────────────────────────────────────
//
//   60 % GET  /balance/{userId}
//   30 % POST /transactions
//   10 % GET  /transactions/{id}

export default function main(data) {
  const roll = Math.random();

  if (roll < 0.60) {
    getBalance();
  } else if (roll < 0.90) {
    createTransaction();
  } else {
    getTransactionStatus(data.txID);
  }

  // ~10 ms think-time between requests
  sleep(0.01);
}

// ── Request helpers ──────────────────────────────────────────────────────────

function createTransaction() {
  const from = randomItem(USER_IDS);
  let to = randomItem(USER_IDS);
  while (to === from) to = randomItem(USER_IDS);

  const payload = JSON.stringify({
    from_user_id: from,
    to_user_id:   to,
    amount:       Math.floor(Math.random() * 900) + 100,
    description:  'load test',
  });

  const params = {
    headers: {
      ...HEADERS,
      'Idempotency-Key': uuidv4(),  // unique per request — no duplicates
    },
  };

  group('POST /api/v1/transactions', () => {
    const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, params);

    createTrend.add(res.timings.duration);

    const ok = check(res, {
      'create: status is 2xx': (r) => r.status === 201 || r.status === 200,
      'create: has id':        (r) => {
        try { return JSON.parse(r.body).data.id !== ''; } catch { return false; }
      },
    });

    errorRate.add(!ok);
  });
}

function getTransactionStatus(txID) {
  if (!txID) return;

  group('GET /api/v1/transactions/:id', () => {
    const res = http.get(`${BASE_URL}/api/v1/transactions/${txID}`, { headers: HEADERS });

    statusTrend.add(res.timings.duration);

    const ok = check(res, {
      'status: 200': (r) => r.status === 200,
      'status: has id': (r) => {
        try { return JSON.parse(r.body).data.id !== ''; } catch { return false; }
      },
    });

    errorRate.add(!ok);
  });
}

function getBalance() {
  const userID = randomItem(USER_IDS);

  group('GET /api/v1/balance/:userId', () => {
    const res = http.get(`${BASE_URL}/api/v1/balance/${userID}`, { headers: HEADERS });

    balanceTrend.add(res.timings.duration);

    const ok = check(res, {
      'balance: 200': (r) => r.status === 200,
      'balance: has balance field': (r) => {
        try {
          const body = JSON.parse(r.body);
          return typeof body.data.balance === 'number';
        } catch { return false; }
      },
    });

    errorRate.add(!ok);
  });
}

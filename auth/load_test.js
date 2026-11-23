import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate } from 'k6/metrics';

// BASE_URL is injected by the CD pipeline via -e BASE_URL=<cloud_run_url>
const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const ID_TOKEN = __ENV.ID_TOKEN || '';
const P95_MS = __ENV.P95_MS || (ID_TOKEN ? '3000' : '500');

const errores5xx = new Rate('errores_5xx_o_timeout');
const rechazos429 = new Counter('rechazos_429');

export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Ramp-up to 20 users
    { duration: '30s', target: 20 }, // Stay at 20 users
    { duration: '10s', target: 0 },  // Ramp-down to 0
  ],
  thresholds: {
    http_req_duration: [`p(95)<${P95_MS}`],
    errores_5xx_o_timeout: ['rate<0.02'],
    checks: ['rate>0.95'],
  },
};

export default function runLoadTest() {
  // Use X-Serverless-Authorization so Cloud Run IAM does not collide with app JWT headers.
  const authHeaders = ID_TOKEN ? { 'X-Serverless-Authorization': `Bearer ${ID_TOKEN}` } : {};

  // 1. Health Endpoint Spike
  const healthRes = http.get(`${BASE_URL}/health`, { headers: authHeaders, timeout: '15s' });
  if (healthRes.status === 429) {
    rechazos429.add(1);
  } else {
    errores5xx.add(healthRes.status === 0 || healthRes.status >= 500);
    check(healthRes, {
      'health check returns 200': (r) => r.status === 200,
    });
  }

  // 2. Mock Login Request (exercises validation/auth pipeline without mutating state)
  const loginPayload = JSON.stringify({
    email: 'loadtest@example.com',
    password: 'wrongpassword123',
  });

  const params = {
    headers: { 'Content-Type': 'application/json', ...authHeaders },
    timeout: '15s',
  };

  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, params);
  if (loginRes.status === 429) {
    rechazos429.add(1);
  } else {
    errores5xx.add(loginRes.status === 0 || loginRes.status >= 500);
    // We expect a 401 Unauthorized or 400 Bad Request because the credentials/payload are invalid.
    check(loginRes, {
      'login gracefully rejects': (r) => r.status === 401 || r.status === 400,
    });
  }

  // Wait 1 second between iterations per virtual user
  sleep(1);
}

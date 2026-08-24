import http from 'k6/http';
import { check, sleep } from 'k6';

// BASE_URL is injected by the CD pipeline via -e BASE_URL=<cloud_run_url>
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Ramp-up to 20 users
    { duration: '30s', target: 20 }, // Stay at 20 users
    { duration: '10s', target: 0 },  // Ramp-down to 0
  ],
  thresholds: {
    // 95% of requests must complete below 500ms
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  // 1. Health Endpoint Spike
  let healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check returns 200': (r) => r.status === 200,
  });

  // 2. Mock Login Request (to test latency of auth logic/DB/bcrypt)
  const loginPayload = JSON.stringify({
    email: 'loadtest@example.com',
    password: 'wrongpassword123',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, params);

  // We expect a 401 Unauthorized or 400 Bad Request because the credentials are wrong.
  check(loginRes, {
    'login gracefully rejects': (r) => r.status === 401 || r.status === 400,
  });

  // Wait 1 second between iterations per virtual user
  sleep(1);
}

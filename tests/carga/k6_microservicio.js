// Parameterized read-only k6 load test for direct Cloud Run microservice verification in CD.
// Never mutates database state (safe for shared Neon/Firestore testing environments).
//
// Usage:
//   docker run --rm -i \
//     -e SERVICE=notifications \
//     -e BASE_URL=https://notifications-testing-xxx.a.run.app \
//     -e ID_TOKEN="<gcp-oidc-id-token>" \
//     grafana/k6:0.54.0 run - < tests/carga/k6_microservicio.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const SERVICE = (__ENV.SERVICE || '').trim();
const BASE_URL = (__ENV.BASE_URL || '').replace(/\/$/, '');
const ID_TOKEN = __ENV.ID_TOKEN || '';
const P95_MS = __ENV.P95_MS || '3000';

const SERVICE_PROFILES = {
  auth: {
    healthPath: '/health',
    protectedPath: '/api/v1/auth/me',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  notifications: {
    healthPath: '/health',
    protectedPath: '/api/v1/notifications',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  education: {
    healthPath: '/api/salud',
    protectedPath: '/api/cursos',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  'human-promotion': {
    healthPath: '/health',
    protectedPath: '/api/v1/human-promotion/fichas?participante_id=loadtest',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  'labor-integration': {
    healthPath: '/health',
    protectedPath: '/api/v1/labor-integration/applications/me',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  campaigns: {
    healthPath: '/health',
    protectedPath: '/api/v1/campaigns',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  dashboard: {
    healthPath: '/health',
    protectedPath: '/api/v1/dashboard/resumen?desde=2026-01-01&hasta=2026-01-31',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  rxresume: {
    healthPath: '/health',
    protectedPath: '/api/v1/rxresume/me',
    protectedMethod: 'GET',
    expectedStatuses: [401],
    testInvalidToken: true,
  },
  'cv-renderer': {
    healthPath: '/health',
    protectedPath: '/render',
    protectedMethod: 'POST',
    protectedBody: '{}',
    expectedStatuses: [400],
    testInvalidToken: false,
  },
};

const profile = SERVICE_PROFILES[SERVICE];
if (!profile) {
  throw new Error(`Unsupported SERVICE="${SERVICE}". Expected one of: ${Object.keys(SERVICE_PROFILES).join(', ')}`);
}
if (!BASE_URL) {
  throw new Error('BASE_URL environment variable is required');
}

const errores = new Rate('errores_5xx_o_timeout');
const rechazos429 = new Counter('rechazos_429');
const latencia = new Trend('latencia_servicio', true);

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '25s', target: 10 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_duration: [`p(95)<${P95_MS}`],
    errores_5xx_o_timeout: ['rate<0.02'],
    checks: ['rate>0.95'],
  },
  summaryTrendStats: ['avg', 'med', 'p(90)', 'p(95)', 'max'],
};

function recordResult(res, label, expectedStatuses) {
  if (res.status === 429) {
    rechazos429.add(1, { servicio: SERVICE });
    return;
  }
  errores.add(res.status === 0 || res.status >= 500, { servicio: SERVICE });
  latencia.add(res.timings.duration, { servicio: SERVICE });
  check(res, {
    [label]: (r) => expectedStatuses.includes(r.status),
  });
}

function sendProtected(headers) {
  const url = `${BASE_URL}${profile.protectedPath}`;
  const params = { headers, tags: { servicio: SERVICE, endpoint: 'protected' }, timeout: '15s' };
  if (profile.protectedMethod === 'POST') {
    const postHeaders = Object.assign({ 'Content-Type': 'application/json' }, headers);
    return http.post(url, profile.protectedBody || '{}', Object.assign({}, params, { headers: postHeaders }));
  }
  return http.get(url, params);
}

export default function () {
  // X-Serverless-Authorization authenticates with Cloud Run IAM without overriding app Authorization headers.
  const cloudRunHeaders = ID_TOKEN ? { 'X-Serverless-Authorization': `Bearer ${ID_TOKEN}` } : {};

  const healthRes = http.get(`${BASE_URL}${profile.healthPath}`, {
    headers: cloudRunHeaders,
    tags: { servicio: SERVICE, endpoint: 'health' },
    timeout: '15s',
  });
  recordResult(healthRes, `${SERVICE} ${profile.healthPath} -> 200`, [200]);

  const protectedRes = sendProtected(cloudRunHeaders);
  recordResult(
    protectedRes,
    `${SERVICE} ${profile.protectedPath} -> ${profile.expectedStatuses.join('/')}`,
    profile.expectedStatuses
  );

  if (profile.testInvalidToken) {
    const invalidHeaders = Object.assign({ Authorization: 'Bearer token-invalido-de-carga' }, cloudRunHeaders);
    const invalidRes = sendProtected(invalidHeaders);
    recordResult(
      invalidRes,
      `${SERVICE} ${profile.protectedPath} invalid JWT -> ${profile.expectedStatuses.join('/')}`,
      profile.expectedStatuses
    );
  }

  sleep(1);
}

import http from 'k6/http';
import { check, sleep } from 'k6';

// BASE_URL is injected by the CD pipeline via -e BASE_URL=<cloud_run_url>
const BASE_URL = (__ENV.BASE_URL || 'http://localhost:8210').replace(/\/$/, '');
const ID_TOKEN = __ENV.ID_TOKEN || '';
const P95_MS = __ENV.P95_MS || (ID_TOKEN ? '3000' : '500');

export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Ramp-up to 20 users
    { duration: '30s', target: 20 }, // Stay at 20 users
    { duration: '10s', target: 0 },  // Ramp-down to 0
  ],
  thresholds: {
    http_req_duration: [`p(95)<${P95_MS}`],
    checks: ['rate>0.99'],
  },
};

export default function () {
  // Use X-Serverless-Authorization so Cloud Run IAM does not collide with app JWT headers.
  const cloudRunHeaders = ID_TOKEN ? { 'X-Serverless-Authorization': `Bearer ${ID_TOKEN}` } : {};

  // 1. Health Endpoint Spike (público, sin JWT)
  let healthRes = http.get(`${BASE_URL}/api/salud`, { headers: cloudRunHeaders, timeout: '15s' });
  check(healthRes, {
    'health check returns 200 or 429': (r) => r.status === 200 || r.status === 429,
  });

  // 2. Lectura de cursos sin token (para ejercitar AuthMiddleware bajo
  // carga sin depender de un JWT válido real ni de Contentful).
  let cursosRes = http.get(`${BASE_URL}/api/cursos`, { headers: cloudRunHeaders, timeout: '15s' });
  check(cursosRes, {
    'cursos sin token rechaza con 401 o 429': (r) => r.status === 401 || r.status === 429,
  });

  // 3. Lectura de cursos con un token con formato inválido (ejercita la
  // rama de validación de JWTService.ValidateToken bajo carga).
  const params = {
    headers: Object.assign(
      { Authorization: 'Bearer token-invalido-de-carga' },
      cloudRunHeaders
    ),
    timeout: '15s',
  };
  let invalidTokenRes = http.get(`${BASE_URL}/api/cursos`, params);
  check(invalidTokenRes, {
    'cursos con token inválido rechaza con 401 o 429': (r) => r.status === 401 || r.status === 429,
  });

  // Wait 1 second between iterations per virtual user
  sleep(1);
}

import http from 'k6/http';
import { check, sleep } from 'k6';

// BASE_URL is injected by the CD pipeline via -e BASE_URL=<cloud_run_url>
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8210';

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
  // 1. Health Endpoint Spike (público, sin JWT)
  let healthRes = http.get(`${BASE_URL}/api/salud`);
  check(healthRes, {
    'health check returns 200': (r) => r.status === 200,
  });

  // 2. Lectura de cursos sin token (para ejercitar AuthMiddleware bajo
  // carga sin depender de un JWT válido real ni de Contentful).
  let cursosRes = http.get(`${BASE_URL}/api/cursos`);
  check(cursosRes, {
    'cursos sin token rechaza con 401': (r) => r.status === 401,
  });

  // 3. Lectura de cursos con un token con formato inválido (ejercita la
  // rama de validación de JWTService.ValidateToken bajo carga).
  const params = {
    headers: {
      Authorization: 'Bearer token-invalido-de-carga',
    },
  };
  let invalidTokenRes = http.get(`${BASE_URL}/api/cursos`, params);
  check(invalidTokenRes, {
    'cursos con token inválido rechaza con 401': (r) => r.status === 401,
  });

  // Wait 1 second between iterations per virtual user
  sleep(1);
}

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 20 },
    { duration: '30s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.05'],   // error rate must be under 5%
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export default function () {
  // 1. Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health status is 200': (r) => r.status === 200,
  });

  // 2. Validate token internal endpoint
  const validateRes = http.get(`${BASE_URL}/internal/validate-token`, {
    headers: { Authorization: 'Bearer mock_token_for_load_test' },
  });
  check(validateRes, {
    'validate token handled': (r) => r.status === 200 || r.status === 401,
  });

  sleep(1);
}

export function handleSummary(data) {
  const jsonReport = JSON.stringify(data, null, 2);
  const htmlReport = `<!DOCTYPE html>
<html>
<head>
  <title>k6 Performance Test Report - Auth Service</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; background: #f8fafc; }
    h1 { color: #485839; }
    .card { background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 1rem; }
    .metric { font-size: 1.2rem; margin: 0.5rem 0; }
  </style>
</head>
<body>
  <h1>Reporte de Rendimiento k6 - Auth Service</h1>
  <div class="card">
    <h2>Resumen de Ejecución</h2>
    <div class="metric"><strong>Duración Total:</strong> ${data.state.testRunDurationMs} ms</div>
    <div class="metric"><strong>Peticiones Totales:</strong> ${data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0}</div>
    <div class="metric"><strong>Tiempo Promedio de Respuesta:</strong> ${data.metrics.http_req_duration ? data.metrics.http_req_duration.values.avg.toFixed(2) : 0} ms</div>
    <div class="metric"><strong>Tasa de Fallos:</strong> ${data.metrics.http_req_failed ? (data.metrics.http_req_failed.values.rate * 100).toFixed(2) : 0}%</div>
  </div>
</body>
</html>`;

  return {
    'reports/performance/k6-summary.json': jsonReport,
    'reports/performance/k6-report.html': htmlReport,
  };
}

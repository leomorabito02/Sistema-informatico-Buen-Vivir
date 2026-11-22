// Prueba de carga de TODOS los microservicios a través del API Gateway
// (D6, 22/11/2026). La corre el CD solo contra TESTING; nunca contra
// producción.
//
// Sin escrituras: testing comparte Firestore y Neon con el stack local.
// Cada servicio recibe pedidos GET públicos (lectura liviana) o pedidos
// sin token que deben responder 401 (miden gateway + servicio + validación
// del JWT sin tocar la base).
//
// Los servicios limitan a 100 pedidos/min por IP y el runner sale de una
// sola IP: un 429 se cuenta aparte (rechazos_429) y no como error. Lo que
// falla la prueba es un 5xx, un timeout o una latencia alta.
//
// Uso local contra testing (solo lectura):
//   docker run --rm -i -e BASE_URL=https://<gateway-testing> \
//     grafana/k6:0.54.0 run - < tests/carga/k6_servicios_gateway.js
import http from "k6/http";
import { check } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";

const BASE = (__ENV.BASE_URL || "http://localhost:8080").replace(/\/$/, "");
const ORIGIN = __ENV.ORIGIN || "https://testing.ceibuenvivir.pages.dev";

const errores = new Rate("errores_5xx_o_timeout");
const rechazos429 = new Counter("rechazos_429");
const latenciaPorServicio = new Trend("latencia_servicio", true);

// [servicio, ruta, estados esperados]
const RUTAS = [
  ["auth", "/api/v1/auth/me", [401]],
  ["education", "/api/cursos", [401]],
  ["human-promotion", "/api/v1/human-promotion/fichas", [401]],
  ["campaigns", "/api/v1/campaigns/publicas", [200]],
  ["labor-integration", "/api/v1/labor-integration/offers", [200]],
  ["dashboard", "/api/v1/dashboard/resumen", [401]],
  ["notifications", "/api/v1/notifications", [401]],
  ["rxresume", "/api/v1/rxresume/me", [401]],
];

export const options = {
  scenarios: {
    // Carga sostenida y baja: 4 pedidos/s durante 60 s, repartidos.
    sostenida: {
      executor: "constant-arrival-rate",
      rate: 4,
      timeUnit: "1s",
      duration: "60s",
      preAllocatedVUs: 4,
      maxVUs: 16,
    },
    // Pico corto de concurrencia: rampa a 15 usuarios simultáneos.
    pico: {
      executor: "ramping-vus",
      startTime: "65s",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 15 },
        { duration: "20s", target: 15 },
        { duration: "5s", target: 0 },
      ],
    },
  },
  thresholds: {
    // Cloud Run escala a cero: el p95 admite arranques en frío.
    http_req_duration: ["p(95)<3000"],
    "http_req_duration{scenario:sostenida}": ["p(95)<2500"],
    errores_5xx_o_timeout: ["rate<0.02"],
    checks: ["rate>0.9"],
  },
  summaryTrendStats: ["avg", "med", "p(90)", "p(95)", "max"],
};

export default function () {
  const [servicio, ruta, esperados] = RUTAS[Math.floor(Math.random() * RUTAS.length)];
  const res = http.get(`${BASE}${ruta}`, {
    headers: { Origin: ORIGIN },
    tags: { servicio },
    timeout: "15s",
  });
  if (res.status === 429) {
    rechazos429.add(1, { servicio });
    return;
  }
  const fallo = res.status === 0 || res.status >= 500;
  errores.add(fallo, { servicio });
  latenciaPorServicio.add(res.timings.duration, { servicio });
  check(res, { [`${servicio} ${ruta} -> ${esperados.join("/")}`]: (r) => esperados.includes(r.status) });
}

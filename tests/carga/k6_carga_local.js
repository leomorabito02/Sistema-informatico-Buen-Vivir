// Prueba de carga controlada contra el stack Docker LOCAL. Auditoría
// 2026-11-11, Problema 2.
//
// OJO: el stack local usa Firestore y Neon reales (compartidos con
// testing). Por eso la carga fuerte va solo a rutas que NO tocan la base
// (salud y rechazos 401), y la lectura pública con base va a un ritmo bajo
// y fijo. No hay escrituras.
//
// El gateway nginx limita a 10 pedidos/s por IP (ráfaga 20). Como k6 sale
// de una sola IP, para medir los SERVICIOS se les pega directo a su puerto
// (education :8210, human-promotion :8085) y al gateway se le pega por
// debajo del límite (escenario "gateway").
//
// Uso (desde la raíz del repo, con el stack levantado):
//   docker run --rm -v "$PWD/tests/carga:/scripts" grafana/k6:0.54.0 run \
//     --summary-export=/scripts/resultado-carga.json /scripts/k6_carga_local.js
import http from "k6/http";
import { check } from "k6";
import { Rate } from "k6/metrics";

const HOST = __ENV.HOST || "host.docker.internal";
const GW = `http://${HOST}:8080`;
const EDU = `http://${HOST}:8210`;
const PH = `http://${HOST}:8085`;
const errores5xx = new Rate("errores_5xx");

export const options = {
  scenarios: {
    // 1) Gateway por debajo de su límite (8/s): debe responder todo bien.
    gateway: {
      executor: "constant-arrival-rate",
      exec: "gateway",
      rate: 8,
      timeUnit: "1s",
      duration: "40s",
      preAllocatedVUs: 5,
      maxVUs: 20,
    },
    // 2) Servicios directo, sin base: rampa a 50 usuarios simultáneos.
    servicios: {
      executor: "ramping-vus",
      exec: "servicios",
      startVUs: 0,
      stages: [
        { duration: "15s", target: 50 },
        { duration: "30s", target: 50 },
        { duration: "10s", target: 0 },
      ],
    },
    // 3) Lectura pública con base real (biblioteca, Firestore): 3/s fijo.
    publico: {
      executor: "constant-arrival-rate",
      exec: "publico",
      rate: 3,
      timeUnit: "1s",
      duration: "40s",
      preAllocatedVUs: 5,
      maxVUs: 10,
    },
  },
  thresholds: {
    "http_req_duration{scenario:gateway}": ["p(95)<1500"],
    "http_req_duration{scenario:servicios}": ["p(95)<500"],
    "http_req_duration{scenario:publico}": ["p(95)<2000"],
    "errores_5xx{scenario:gateway}": ["rate<0.01"],
    "errores_5xx{scenario:servicios}": ["rate<0.01"],
    "errores_5xx{scenario:publico}": ["rate<0.01"],
  },
};

function registrar(res) {
  errores5xx.add(res.status >= 500);
}

const GW_RUTAS = [
  ["/health", [200]],
  ["/api/salud", [200]],
  ["/api/v1/campaigns/publicas", [200]],
  ["/api/cursos", [401]],
  ["/api/v1/users", [401]],
];

export function gateway() {
  const [ruta, ok] = GW_RUTAS[Math.floor(Math.random() * GW_RUTAS.length)];
  const r = http.get(`${GW}${ruta}`, { tags: { ruta: `gw ${ruta}` } });
  registrar(r);
  check(r, { [`gateway ${ruta} -> ${ok}`]: (x) => ok.includes(x.status) });
}

const SVC_RUTAS = [
  [`${EDU}/api/salud`, 200, "edu salud"],
  [`${EDU}/api/cursos`, 401, "edu cursos sin token"],
  [`${EDU}/api/participante/cursos`, 401, "edu participante sin token"],
  [`${PH}/health`, 200, "ph health"],
  [`${PH}/api/v1/human-promotion/fichas?participante_id=x`, 401, "ph fichas sin token"],
];

export function servicios() {
  const [url, ok, nombre] = SVC_RUTAS[Math.floor(Math.random() * SVC_RUTAS.length)];
  const r = http.get(url, { tags: { ruta: nombre } });
  registrar(r);
  check(r, { [`${nombre} -> ${ok} (o 429)`]: (x) => x.status === ok || x.status === 429 });
}

export function publico() {
  const r = http.get(`${EDU}/api/biblioteca`, { tags: { ruta: "edu biblioteca" } });
  registrar(r);
  check(r, { "biblioteca 200": (x) => x.status === 200 });
}

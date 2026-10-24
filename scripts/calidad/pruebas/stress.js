// ST-1 Stress: ¿con cuántos usuarios se rompe Educación?
//
// Escalera de 10 → 20 → 30 → 40 usuarios, 1 minuto en cada escalón (con
// 10 s de subida entre escalones). La escala parte del pico de la prueba
// de carga (scripts/calidad/pruebas/carga.js, 10 usuarios) y llega a 4 veces ese
// valor. Solo lee: no escribe en Contentful.
//
// Cada pedido lleva la etiqueta del escalón en que salió, así el resumen
// muestra el p95 y los errores de cada uno. El primer escalón con p95 > 2 s
// o errores > 5 % es el punto de quiebre, y hace fallar la prueba. Se corta
// antes si los errores pasan del 20 %.
//
// ESCALON (por defecto 1m) acorta los escalones para probar el script.

import { sleep } from 'k6';
import { preparar, recorrerCurso } from './comun.js';

const ESCALON = __ENV.ESCALON || '1m';
const SUBIDA = '10s';
const NIVELES = [10, 20, 30, 40];

function segundos(d) {
  const n = parseFloat(d);
  return d.endsWith('m') ? n * 60 : n;
}

const stages = [];
for (const n of NIVELES) {
  stages.push({ duration: SUBIDA, target: n });
  stages.push({ duration: ESCALON, target: n });
}
stages.push({ duration: SUBIDA, target: 0 });

const thresholds = {
  http_req_duration: ['p(95)<2000'],
  http_req_failed: [
    'rate<0.05',
    { threshold: 'rate<0.20', abortOnFail: true, delayAbortEval: '20s' },
  ],
};
for (const n of NIVELES) {
  thresholds[`http_req_duration{escalon:${n}}`] = ['p(95)<2000'];
  thresholds[`http_req_failed{escalon:${n}}`] = ['rate<0.05'];
}

export const options = {
  stages: stages,
  thresholds: thresholds,
  summaryTrendStats: ['avg', 'med', 'p(90)', 'p(95)', 'max'],
};

export function setup() {
  return preparar();
}

// El escalón según el tiempo transcurrido: la subida hacia un nivel cuenta
// como parte de ese nivel.
function escalonActual(data) {
  const transcurrido = (Date.now() - data.inicio) / 1000;
  const porNivel = segundos(SUBIDA) + segundos(ESCALON);
  const i = Math.min(Math.floor(transcurrido / porNivel), NIVELES.length - 1);
  return NIVELES[i];
}

export default function stress(data) {
  recorrerCurso(data, { escalon: String(escalonActual(data)) });
  sleep(1);
}

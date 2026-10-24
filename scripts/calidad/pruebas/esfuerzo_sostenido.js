// ST-2 Esfuerzo sostenido: ¿Educación anda igual de bien después de media
// hora con carga normal?
//
// 10 usuarios constantes (el pico de scripts/calidad/pruebas/carga.js) durante 30
// minutos, con las mismas lecturas que el stress. Solo lee: no escribe en
// Contentful.
//
// Compara el p95 de los primeros 5 minutos con el de los últimos 5. Pasa si
// el final no es más de 20 % peor que el inicio y los errores quedan por
// debajo del 5 %. Como k6 no puede comparar dos métricas en un umbral, el
// resultado de esa comparación queda en sostenido-degradacion.json y el
// pipeline falla si dice "aprobado": false. La memoria del contenedor la
// mide el pipeline con docker stats al empezar y al terminar.
//
// DURACION y VENTANA (por defecto 30m y 5m) acortan la prueba para probar
// el script.

import { sleep } from 'k6';
import { Trend } from 'k6/metrics';
import { preparar, recorrerCurso } from './comun.js';

const DURACION = __ENV.DURACION || '30m';
const VENTANA = __ENV.VENTANA || '5m';
const REPORTES = __ENV.REPORTES || 'reportes-servicio';
const USUARIOS = 10;

function segundos(d) {
  const n = parseFloat(d);
  return d.endsWith('m') ? n * 60 : n;
}

const latenciaInicio = new Trend('latencia_inicio', true);
const latenciaFinal = new Trend('latencia_final', true);

export const options = {
  scenarios: {
    sostenido: { executor: 'constant-vus', vus: USUARIOS, duration: DURACION },
  },
  thresholds: {
    http_req_failed: [
      'rate<0.05',
      { threshold: 'rate<0.05', abortOnFail: true, delayAbortEval: '2m' },
    ],
    latencia_inicio: ['p(95)>=0'],
    latencia_final: ['p(95)>=0'],
  },
  summaryTrendStats: ['avg', 'med', 'p(90)', 'p(95)', 'max'],
};

export function setup() {
  return preparar();
}

export default function sostenido(data) {
  const transcurrido = (Date.now() - data.inicio) / 1000;
  const respuestas = recorrerCurso(data, {});
  for (const r of respuestas) {
    if (transcurrido < segundos(VENTANA)) {
      latenciaInicio.add(r.timings.duration);
    } else if (transcurrido >= segundos(DURACION) - segundos(VENTANA)) {
      latenciaFinal.add(r.timings.duration);
    }
  }
  sleep(1);
}

export function handleSummary(data) {
  const p95 = (m) => (data.metrics[m] ? data.metrics[m].values['p(95)'] : null);
  const inicio = p95('latencia_inicio');
  const final = p95('latencia_final');
  const errores = data.metrics.http_req_failed ? data.metrics.http_req_failed.values.rate : null;
  const degradacion = inicio && final ? (final - inicio) / inicio : null;
  const resultado = {
    usuarios: USUARIOS,
    duracion: DURACION,
    ventana: VENTANA,
    p95_inicio_ms: inicio,
    p95_final_ms: final,
    degradacion: degradacion,
    errores: errores,
    aprobado: degradacion !== null && degradacion <= 0.2 && errores !== null && errores < 0.05,
  };
  return {
    stdout: JSON.stringify(resultado, null, 2) + '\n',
    [`${REPORTES}/sostenido-degradacion.json`]: JSON.stringify(resultado, null, 2),
  };
}

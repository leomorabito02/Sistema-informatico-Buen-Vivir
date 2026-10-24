// Partes compartidas por stress.js y esfuerzo_sostenido.js.
//
// Los dos escenarios solo leen: catálogo de cursos (Contentful, con la
// caché de 5 min de Educación), actividades e inscripciones (Firestore).
// No escriben en Contentful, así no gastan pedidos de su plan gratuito.
// Los tokens se firman como en scripts/calidad/pruebas/carga.js, con el JWT_SECRET
// del entorno de calidad.

import http from 'k6/http';
import { check } from 'k6';
import crypto from 'k6/crypto';
import encoding from 'k6/encoding';

export const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:18210';

export function token(sub, role) {
  const secret = __ENV.JWT_SECRET;
  if (!secret) {
    throw new Error('Falta JWT_SECRET (el del entorno de calidad)');
  }
  const now = Math.floor(Date.now() / 1000);
  const header = encoding.b64encode(JSON.stringify({ alg: 'HS256', typ: 'JWT' }), 'rawurl');
  const payload = encoding.b64encode(
    JSON.stringify({ sub: sub, role: role, type: 'access', iat: now, exp: now + 4 * 3600 }),
    'rawurl'
  );
  const signature = crypto.hmac('sha256', secret, `${header}.${payload}`, 'base64rawurl');
  return `${header}.${payload}.${signature}`;
}

function lista(res) {
  const body = res.json();
  return Array.isArray(body) ? body : body.data || [];
}

// Lee el catálogo una vez (calienta la caché) y elige los cursos que cargó
// cargar_por_api.py, que son los primeros del catálogo.
export function preparar() {
  const adminToken = token('admin-prueba', 'role_admin');
  const res = http.get(`${BASE_URL}/api/cursos`, {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  if (res.status !== 200) {
    throw new Error(`No se pudo leer el catálogo: ${res.status}`);
  }
  const cursos = lista(res).slice(0, 3).map((c) => c.id);
  if (cursos.length === 0) {
    throw new Error('El catálogo está vacío: ¿se corrió cargar_por_api.py?');
  }
  return { adminToken: adminToken, cursos: cursos, inicio: Date.now() };
}

// Lo que pide cada usuario virtual en cada vuelta: lo mismo que la pantalla
// de un curso (catálogo, actividades e inscriptos).
export function recorrerCurso(data, tags) {
  const params = { headers: { Authorization: `Bearer ${data.adminToken}` }, tags: tags };
  const curso = data.cursos[Math.floor(Math.random() * data.cursos.length)];
  const respuestas = [
    http.get(`${BASE_URL}/api/cursos`, params),
    http.get(`${BASE_URL}/api/cursos/${curso}/actividades`, params),
    http.get(`${BASE_URL}/api/cursos/${curso}/inscripciones`, params),
  ];
  for (const r of respuestas) {
    check(r, { 'responde 200': (x) => x.status === 200 });
  }
  return respuestas;
}

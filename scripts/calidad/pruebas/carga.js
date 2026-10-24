import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';
import encoding from 'k6/encoding';

// BASE_URL is injected by the runner or defaults to local server
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8210';
const JWT_SECRET = __ENV.JWT_SECRET || '';
const CONTENTFUL_CMA_TOKEN = __ENV.CONTENTFUL_CMA_TOKEN || '';
const CONTENTFUL_SPACE_ID = __ENV.CONTENTFUL_SPACE_ID || '';
const CONTENTFUL_ENVIRONMENT = __ENV.CONTENTFUL_ENVIRONMENT || 'testing';
const CONTENTFUL_CMA_BASE_URL = __ENV.CONTENTFUL_CMA_BASE_URL || 'https://api.contentful.com';
const ESCRIBIR_CADA = 5;

export const options = {
  stages: [
    { duration: '5s', target: 5 },   // Ramp-up inicial
    { duration: '15s', target: 10 }, // Carga controlada para respetar límites de CMA
    { duration: '5s', target: 0 },   // Ramp-down
  ],
  thresholds: {
    // 95% de las solicitudes deben completarse en menos de 800ms considerando llamadas a Contentful
    http_req_duration: ['p(95)<800'],
    checks: ['rate>0.95'],
  },
};

function generateJWT(sub, role, secret) {
  if (!secret) {
    return '';
  }
  const header = encoding.b64encode(JSON.stringify({ alg: 'HS256', typ: 'JWT' }), 'rawurl');
  const now = Math.floor(Date.now() / 1000);
  const payload = encoding.b64encode(
    JSON.stringify({
      sub: sub,
      role: role,
      type: 'access',
      iat: now,
      exp: now + 3600,
    }),
    'rawurl'
  );
  const signature = crypto.hmac('sha256', secret, `${header}.${payload}`, 'base64rawurl');
  return `${header}.${payload}.${signature}`;
}

export function setup() {
  const adminToken = generateJWT('usr-loadtest-admin', 'role_admin', JWT_SECRET);
  if (!adminToken) {
    throw new Error('JWT_SECRET is required to run real authenticated load tests');
  }

  const authHeaders = {
    headers: {
      Authorization: `Bearer ${adminToken}`,
      'Content-Type': 'application/json',
    },
  };

  // Verificar salud de la API antes de iniciar
  let health = http.get(`${BASE_URL}/api/salud`);
  check(health, { 'api is ready': (r) => r.status === 200 });

  // 1. Obtener o crear trainingArea
  let trainingAreaId = '';
  let bootstrapAreaId = '';
  let areasRes = http.get(`${BASE_URL}/api/areas-formativas`, authHeaders);
  if (areasRes.status === 200) {
    const areas = areasRes.json();
    if (Array.isArray(areas) && areas.length > 0) {
      trainingAreaId = areas[0].id;
    }
  }

  if (!trainingAreaId) {
    const createAreaRes = http.post(
      `${BASE_URL}/api/areas-formativas`,
      JSON.stringify({
        name: '[LOAD-TEST] Area Formativa',
        description: 'Area temporal para pruebas de carga',
      }),
      authHeaders
    );
    if (createAreaRes.status === 201) {
      trainingAreaId = createAreaRes.json('id');
      bootstrapAreaId = trainingAreaId;
    }
  }

  // 2. Obtener o crear material
  let materialId = '';
  let bootstrapMaterialId = '';
  let matRes = http.get(`${BASE_URL}/api/materiales`, authHeaders);
  if (matRes.status === 200) {
    const mats = matRes.json();
    if (Array.isArray(mats) && mats.length > 0) {
      materialId = mats[0].id;
    }
  }

  if (!materialId) {
    const createMatRes = http.post(
      `${BASE_URL}/api/materiales`,
      JSON.stringify({
        title: '[LOAD-TEST] Material',
        type: 'Link',
        url: 'https://example.com/load-test',
      }),
      authHeaders
    );
    if (createMatRes.status === 201) {
      materialId = createMatRes.json('id');
      bootstrapMaterialId = materialId;
    }
  }

  return {
    adminToken: adminToken,
    trainingAreaId: trainingAreaId,
    materialId: materialId,
    bootstrapAreaId: bootstrapAreaId,
    bootstrapMaterialId: bootstrapMaterialId,
  };
}

export default function loadTest(data) {
  const authHeaders = {
    headers: {
      Authorization: `Bearer ${data.adminToken}`,
      'Content-Type': 'application/json',
    },
  };

  // 1. Lectura real CDA de cursos (catálogo en Contentful)
  let cursosRes = http.get(`${BASE_URL}/api/cursos`, authHeaders);
  check(cursosRes, {
    'GET /api/cursos responde 200': (r) => r.status === 200,
  });

  // 2. Creación real CMA de curso (escritura en Contentful con prefijo [LOAD-TEST]).
  // Solo escribe el primer usuario virtual, una de cada ESCRIBIR_CADA vueltas:
  // unos 4 cursos por corrida en vez de ~150, para no gastar los pedidos del
  // plan gratuito de Contentful (riesgo RP-04). La escritura igual se mide.
  if (data.trainingAreaId && data.materialId && __VU === 1 && __ITER % ESCRIBIR_CADA === 0) {
    const coursePayload = JSON.stringify({
      title: `[LOAD-TEST] Curso ${Date.now()}-${__VU}-${__ITER}`,
      description: {
        nodeType: 'document',
        data: {},
        content: [
          {
            nodeType: 'paragraph',
            data: {},
            content: [
              {
                nodeType: 'text',
                value: 'Descripción de prueba de carga con Contentful',
                marks: [],
                data: {},
              },
            ],
          },
        ],
      },
      responsibleTeacher: 'usr-loadtest-teacher',
      duration: '10 horas',
      mode: 'Virtual',
      trainingAreaId: data.trainingAreaId,
      materialIds: [data.materialId],
    });

    let createRes = http.post(`${BASE_URL}/api/cursos`, coursePayload, authHeaders);
    check(createRes, {
      'POST /api/cursos responde 201': (r) => r.status === 201,
    });
  }

  sleep(1);
}

function deleteContentfulEntry(cmaBase, spaceId, env, cmaHeaders, entry) {
  const entryId = entry.sys.id;
  if (entry.sys.publishedVersion) {
    http.del(
      `${cmaBase}/spaces/${spaceId}/environments/${env}/entries/${entryId}/published`,
      null,
      cmaHeaders
    );
  }
  http.del(
    `${cmaBase}/spaces/${spaceId}/environments/${env}/entries/${entryId}`,
    null,
    cmaHeaders
  );
}

function purgeLoadTestEntries(cmaBase, spaceId, env, token) {
  if (!token || !spaceId) {
    return;
  }
  const cmaHeaders = {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  };
  const cmaUrl = `${cmaBase}/spaces/${spaceId}/environments/${env}/entries?content_type=course&query=%5BLOAD-TEST%5D`;
  const res = http.get(cmaUrl, cmaHeaders);
  if (res.status !== 200) {
    return;
  }
  const items = res.json('items');
  if (!Array.isArray(items)) {
    return;
  }
  for (const item of items) {
    deleteContentfulEntry(cmaBase, spaceId, env, cmaHeaders, item);
  }
}

export function teardown(data) {
  const authHeaders = {
    headers: {
      Authorization: `Bearer ${data.adminToken}`,
      'Content-Type': 'application/json',
    },
  };

  // 1. Eliminar áreas/materiales de bootstrap si fueron creados
  if (data.bootstrapMaterialId) {
    http.del(`${BASE_URL}/api/materiales/${data.bootstrapMaterialId}`, null, authHeaders);
  }
  if (data.bootstrapAreaId) {
    http.del(`${BASE_URL}/api/areas-formativas/${data.bootstrapAreaId}`, null, authHeaders);
  }

  // 2. Purgar cursos de prueba directamente de Contentful CMA
  purgeLoadTestEntries(
    CONTENTFUL_CMA_BASE_URL,
    CONTENTFUL_SPACE_ID,
    CONTENTFUL_ENVIRONMENT,
    CONTENTFUL_CMA_TOKEN
  );
}



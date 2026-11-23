/**
 * Contenido de la página de Ayuda (heurística 10 de Nielsen: ayuda
 * centrada en la tarea, con pasos y corta). Los nombres de botones y
 * secciones son los que se ven en pantalla: si cambia un botón, cambiar
 * también el paso que lo nombra.
 */
export interface AyudaTema {
  pregunta: string;
  pasos: string[];
}

const PARTICIPANTE: AyudaTema[] = [
  {
    pregunta: "¿Cómo entro a un curso?",
    pasos: [
      "En el menú, abrí Mis cursos.",
      "En la tarjeta del curso, tocá Ir al curso.",
    ],
  },
  {
    pregunta: "¿Cómo entrego una actividad?",
    pasos: [
      "Dentro del curso, buscá la actividad en Actividades y tocá Realizar.",
      "Escribí tu respuesta o tocá Adjuntar archivo.",
      "Tocá Entregar. Al entregar se registra tu asistencia y sumás las habilidades de esa actividad.",
    ],
  },
  {
    pregunta: "¿Dónde veo mi recorrido en el centro?",
    pasos: ["En el menú, abrí Mi proceso."],
  },
  {
    pregunta: "¿Cómo me postulo a una oferta de trabajo?",
    pasos: [
      "Antes, revisá que tu CV esté al día en Mi CV.",
      "Abrí Ofertas laborales, pestaña Ofertas disponibles.",
      "En la oferta, tocá Postularme y confirmá.",
      "Tus postulaciones quedan en la pestaña Mis postulaciones.",
    ],
  },
  {
    pregunta: "¿Cómo armo mi CV?",
    pasos: [
      "Abrí Mi CV y completá tus datos, habilidades e idiomas.",
      "Tocá Guardar y generar currículum.",
      "Con Descargar PDF lo bajás; con Compartir CV obtenés un link.",
    ],
  },
];

const DOCENTE: AyudaTema[] = [
  {
    pregunta: "¿Dónde veo los cursos que doy?",
    pasos: [
      "En el menú, abrí Mis cursos.",
      "Tocá Ver detalle para entrar al curso o Materiales para ver los materiales.",
    ],
  },
  {
    pregunta: "¿Cómo cargo una actividad?",
    pasos: [
      "Entrá al curso y tocá Nueva actividad.",
      "Completá la consigna y la fecha, y elegí las habilidades que desarrolla.",
      "Guardá. Los participantes la ven en Actividades.",
    ],
  },
  {
    pregunta: "¿Cómo veo lo que entregaron?",
    pasos: ["Dentro del curso, en cada actividad tocá Entregas."],
  },
  {
    pregunta: "¿Cómo se toma la asistencia?",
    pasos: [
      "Se registra sola cuando el participante entrega la actividad.",
      "En el menú, Asistencia muestra la de tus cursos.",
    ],
  },
  {
    pregunta: "¿Dónde veo cómo avanza el grupo?",
    pasos: [
      "Seguimiento educativo: asistencia y avance de cada participante en tus cursos.",
    ],
  },
];

const GESTION_COMUN: AyudaTema[] = [
  {
    pregunta: "¿Cómo inscribo a una persona en un curso?",
    pasos: [
      "En el menú, Cursos → Inscripciones.",
      "Elegí el curso y la persona, y tocá Inscribir.",
      "También podés hacerlo desde el Catálogo de cursos con Inscribir participante.",
    ],
  },
  {
    pregunta: "¿Cómo cargo la asistencia general del centro?",
    pasos: [
      "En el menú, Asistencia → Asistencia general.",
      "Marcá Presente o Ausente para cada persona.",
      "Tocá Guardar asistencia.",
    ],
  },
  {
    pregunta: "¿Dónde veo la asistencia de cada curso?",
    pasos: ["En el menú, Asistencia → Asistencia por curso."],
  },
  {
    pregunta: "¿Cómo veo la ficha de un participante?",
    pasos: [
      "En el menú, Promoción humana → Fichas personales.",
      "Buscá a la persona por nombre o apellido para ver o crear su ficha.",
    ],
  },
  {
    pregunta: "¿Cómo manejo las ofertas laborales?",
    pasos: [
      "Abrí Inserción laboral.",
      "Las ofertas que publica una empresa quedan Pendientes: revisalas y tocá Aprobar.",
      "Indicá la cantidad de vacantes al publicar o editar, y usá Cerrar oferta si la búsqueda finaliza antes.",
      "Con Postular participante postulás a alguien; con Ver postulados ves quiénes se postularon.",
      "En Finalizadas e historial ves las ofertas vencidas o cerradas junto al contacto de la empresa y los postulantes para seguimiento.",
    ],
  },
];

const RESPONSABLE: AyudaTema[] = [
  {
    pregunta: "¿Cómo creo un curso?",
    pasos: [
      "En el menú, Cursos → Gestión de cursos.",
      "Si hace falta, creá antes el área con Nueva área y los materiales con Nuevo material.",
      "Tocá Nuevo curso, completá los datos y elegí al docente responsable.",
    ],
  },
  ...GESTION_COMUN,
  {
    pregunta: "¿Cómo doy de alta o edito una cuenta?",
    pasos: [
      "En el menú, Usuarios → Gestión de usuarios.",
      "Tocá Nuevo usuario, o el lápiz (Editar) de una fila para cambiar sus datos.",
      "Eliminar pide confirmación y no se puede deshacer.",
    ],
  },
];

const VOLUNTARIADO: AyudaTema[] = [
  ...GESTION_COMUN,
  {
    pregunta: "¿Cómo actualizo los datos de un participante?",
    pasos: [
      "En el menú, Usuarios → Gestión de participantes.",
      "Buscalo por nombre o email y tocá el lápiz (Editar).",
    ],
  },
];

const EMPRESA: AyudaTema[] = [
  {
    pregunta: "¿Cómo publico una oferta?",
    pasos: [
      "Abrí Ofertas laborales y tocá Publicar oferta.",
      "Completá los datos (incluido el número de vacantes) y guardá.",
      "La oferta queda Pendiente hasta que el centro la aprueba.",
    ],
  },
  {
    pregunta: "¿Cómo veo quién se postuló o cierro una oferta antes de tiempo?",
    pasos: [
      "En la oferta, tocá Ver postulados para revisar perfiles y aceptar o rechazar.",
      "Si ya cubriste las vacantes, tocá Cerrar oferta para moverla a Finalizadas e historial.",
      "Desde Finalizadas e historial ves los canales del centro y qué postulantes requieren seguimiento.",
    ],
  },
  {
    pregunta: "¿Cómo me comunico con el centro?",
    pasos: ["En el menú, abrí Contacto con el centro."],
  },
];

export const AYUDA_POR_ROL: Record<string, AyudaTema[]> = {
  role_admin: RESPONSABLE,
  role_dev: RESPONSABLE,
  role_volunteer: VOLUNTARIADO,
  role_participant: PARTICIPANTE,
  role_teacher: DOCENTE,
  role_company: EMPRESA,
};

export const AYUDA_GENERAL: AyudaTema[] = [
  {
    pregunta: "¿Dónde veo mis avisos?",
    pasos: [
      "Tocá la campana, arriba a la derecha.",
      "Con Marcar todas como leídas las dejás al día.",
    ],
  },
  {
    pregunta: "¿Cómo cambio mi contraseña o mis datos?",
    pasos: [
      "Tocá tu nombre, arriba a la derecha, y elegí Mi cuenta.",
      "Datos Personales para tus datos; Seguridad para la contraseña.",
    ],
  },
  {
    pregunta: "¿Cómo descargo o pido borrar mis datos personales?",
    pasos: [
      "En Mi cuenta, abrí Privacidad y Datos (Ley 25.326).",
      "Descargar mis datos personales baja una copia; desde ahí también podés pedir la baja de la cuenta.",
    ],
  },
  {
    pregunta: "Me apareció «Tu sesión venció», ¿qué hago?",
    pasos: [
      "Tocá Seguir conectado/a para continuar donde estabas.",
      "Si tocás Cerrar sesión, lo que no hayas guardado se pierde.",
    ],
  },
  {
    pregunta: "¿Cómo cierro sesión?",
    pasos: ["Tocá tu nombre, arriba a la derecha, y elegí Cerrar sesión."],
  },
];

export const ATAJOS: { teclas: string[]; accion: string }[] = [
  { teclas: ["/"], accion: "Ir al buscador de la pantalla (si tiene)" },
  { teclas: ["?"], accion: "Abrir esta ayuda" },
  { teclas: ["Esc"], accion: "Cerrar el menú en el celular" },
];

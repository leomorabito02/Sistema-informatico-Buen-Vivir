/**
 * CONSTANTES GLOBALES DEL SISTEMA
 * Valores, roles, estados, y configuraciones
 */

// ============================================================================
// ROLES Y PERMISOS
// ============================================================================

export const ROLES = {
  ADMIN: 'admin',
  PARTICIPANTE: 'participante',
  VOLUNTARIO: 'voluntario',
  EMPRESA: 'empresa',
  DOCENTE: 'docente',
} as const;

export type RoleType = (typeof ROLES)[keyof typeof ROLES];

// ============================================================================
// ESTADOS Y STATUSES
// ============================================================================

export const COURSE_STATUSES = {
  DRAFT: 'draft',
  ACTIVO: 'activo',
  FINALIZADO: 'finalizado',
  CANCELADO: 'cancelado',
  PAUSADO: 'pausado',
} as const;

export const USER_STATUSES = {
  ACTIVO: 'activo',
  INACTIVO: 'inactivo',
  PENDIENTE_APROBACION: 'pendiente_aprobacion',
  SUSPENDIDO: 'suspendido',
} as const;

export const JOB_OFFER_STATUSES = {
  PENDIENTE: 'pendiente',
  APROBADA: 'aprobada',
  RECHAZADA: 'rechazada',
  CERRADA: 'cerrada',
  ARCHIVADA: 'archivada',
} as const;

export const ENROLLMENT_STATUSES = {
  INSCRIPTO: 'inscripto',
  EN_PROGRESO: 'en_progreso',
  COMPLETADO: 'completado',
  ABANDONADO: 'abandonado',
  SUSPENSO: 'suspenso',
} as const;

export const CAMPAIGN_STATUSES = {
  DRAFT: 'draft',
  PUBLICADA: 'publicada',
  PAUSADA: 'pausada',
  ARCHIVADA: 'archivada',
} as const;

// ============================================================================
// LÍMITES Y CUOTAS
// ============================================================================

export const LIMITS = {
  MAX_FILE_SIZE: 5242880, // 5MB
  MAX_FILE_SIZE_MB: 5,
  MAX_LOGIN_ATTEMPTS: 5,
  LOGIN_BLOCK_DURATION_MS: 15 * 60 * 1000, // 15 minutos
  MAX_PAGE_SIZE: 100,
  DEFAULT_PAGE_SIZE: 10,
  API_TIMEOUT_MS: 30000, // 30 segundos
} as const;

// ============================================================================
// FORMATOS Y VALIDACIÓN
// ============================================================================

export const ALLOWED_FILE_TYPES = [
  'pdf',
  'doc',
  'docx',
  'xls',
  'xlsx',
  'jpg',
  'jpeg',
  'png',
  'gif',
  'zip',
] as const;

export const DOCUMENT_TYPES = [
  'DNI',
  'PASAPORTE',
  'CEDULA',
] as const;

export const SKILL_LEVELS = {
  BASICO: 'basico',
  INTERMEDIO: 'intermedio',
  AVANZADO: 'avanzado',
  EXPERTO: 'experto',
} as const;

// ============================================================================
// ÁREAS FORMATIVAS
// ============================================================================

export const FORMATIVE_AREAS = [
  'Gastronomía',
  'Oficios',
  'Arte y Diseño',
  'Agricultura',
  'Tecnología',
  'Administración',
  'Salud',
  'Educación',
  'Sostenibilidad',
  'Otro',
] as const;

// ============================================================================
// COLORES Y DISEÑO
// ============================================================================

export const COLORS = {
  PRIMARY: '#16a34a', // green-600
  SECONDARY: '#0ea5e9', // sky-500
  ACCENT: '#ec4899', // pink-500
  SUCCESS: '#22c55e', // green-500
  WARNING: '#f59e0b', // amber-500
  ERROR: '#ef4444', // red-500
  NEUTRAL: '#6b7280', // gray-500
  MUTED: '#9ca3af', // gray-400
} as const;

export const BREAKPOINTS = {
  MOBILE: 320,
  TABLET: 768,
  DESKTOP: 1024,
  WIDE: 1280,
} as const;

// ============================================================================
// RUTAS DE NAVEGACIÓN
// ============================================================================

export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  COURSES: '/courses',
  COURSE_DETAIL: '/courses/:id',
  PROGRESS: '/progress',
  USERS: '/users',
  TEACHER_PORTAL: '/teacher-portal',
  JOB_OFFERS: '/ofertas-laborales',
  JOB_OFFERS_PUBLIC: '/ofertas-laborales-publico',
  CAMPAIGNS: '/gestion-campanas',
  CAMPAIGNS_PUBLIC: '/campanas-publico',
  PARTICIPANT_TRACKING: '/seguimiento-participantes',
  CV_GENERATOR: '/generar-cv',
} as const;

// ============================================================================
// MENSAJES
// ============================================================================

export const MESSAGES = {
  ERROR: {
    GENERIC: 'Ocurrió un error inesperado. Por favor, intenta más tarde.',
    NETWORK: 'Error de conexión. Verifica tu conexión a internet.',
    VALIDATION: 'Por favor, completa todos los campos requeridos.',
    UNAUTHORIZED: 'No tienes permiso para realizar esta acción.',
    NOT_FOUND: 'El recurso solicitado no fue encontrado.',
    SERVER: 'Error del servidor. Por favor, intenta más tarde.',
  },
  SUCCESS: {
    CREATED: 'Creado exitosamente.',
    UPDATED: 'Actualizado exitosamente.',
    DELETED: 'Eliminado exitosamente.',
    SAVED: 'Guardado exitosamente.',
    ENROLLED: 'Inscrito al curso exitosamente.',
  },
  LOADING: 'Cargando...',
  CONFIRM: '¿Estás seguro?',
  CANCEL: 'Cancelar',
} as const;

// ============================================================================
// TIMESTAMPS Y FECHAS
// ============================================================================

export const DATE_FORMATS = {
  FULL: 'dd/MM/yyyy HH:mm:ss',
  DATE: 'dd/MM/yyyy',
  TIME: 'HH:mm',
  ISO: "yyyy-MM-dd'T'HH:mm:ss.SSSxxx",
} as const;

export const DURATION_MS = {
  SECOND: 1000,
  MINUTE: 60 * 1000,
  HOUR: 60 * 60 * 1000,
  DAY: 24 * 60 * 60 * 1000,
  WEEK: 7 * 24 * 60 * 60 * 1000,
  MONTH: 30 * 24 * 60 * 60 * 1000,
} as const;

// ============================================================================
// NOTIFICACIONES
// ============================================================================

export const NOTIFICATION_TYPES = {
  INSCRIPCION_CURSO: 'inscripcion_curso',
  RECORDATORIO_CURSO: 'recordatorio_curso',
  OFERTA_COMPATIBLE: 'oferta_compatible',
  CAMPAÑA_NUEVA: 'campaña_nueva',
  CONTENIDO_APROBADO: 'contenido_aprobado',
  CONTENIDO_RECHAZADO: 'contenido_rechazado',
  APROBACION_VOLUNTARIO: 'aprobacion_voluntario',
  HITO_REGISTRADO: 'hito_registrado',
} as const;

export const NOTIFICATION_CHANNELS = {
  EMAIL: 'email',
  PUSH: 'push',
  IN_APP: 'in_app',
  SMS: 'sms',
} as const;

// ============================================================================
// CONFIGURACIÓN
// ============================================================================

export const CONFIG = {
  APP_NAME: 'Sistema Buen Vivir',
  ORG_NAME: 'Centro Barrial La Otra Orilla',
  VERSION: '1.0.0',
  ENVIRONMENT: 'development',
  API_BASE_URL: 'http://localhost:3000/api',
  APP_URL: 'http://localhost:5173',
} as const;

// ============================================================================
// PERMISOS POR ROL
// ============================================================================

export const ROLE_PERMISSIONS = {
  admin: [
    'create_user',
    'edit_user',
    'delete_user',
    'create_course',
    'edit_course',
    'delete_course',
    'approve_offer',
    'reject_offer',
    'view_audit_log',
    'export_data',
    'manage_volunteers',
    'create_campaign',
    'edit_campaign',
    'publish_campaign',
    'view_dashboard',
  ],
  participante: [
    'view_profile',
    'edit_profile',
    'view_courses',
    'enroll_course',
    'view_progress',
    'view_offers',
    'apply_offer',
    'view_campaigns',
    'generate_cv',
  ],
  voluntario: [
    'view_profile',
    'edit_participant_profile',
    'view_courses',
    'view_participants',
    'register_milestone',
    'create_campaign',
    'edit_campaign',
    'publish_campaign',
    'view_dashboard_limited',
  ],
  empresa: [
    'view_profile',
    'edit_profile',
    'create_offer',
    'edit_offer',
    'view_applications',
    'publish_campaign',
  ],
  docente: [
    'view_profile',
    'edit_profile',
    'view_courses',
    'upload_content',
    'view_students',
  ],
} as const;

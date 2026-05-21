/**
 * TIPOS CENTRALIZADOS DEL SISTEMA
 * Todas las interfaces, tipos y enums del proyecto
 * Importar desde: import type { User, Course } from '../types'
 */

// ============================================================================
// ENUMS GLOBALES
// ============================================================================

export enum Role {
  ADMIN = 'admin',
  PARTICIPANTE = 'participante',
  VOLUNTARIO = 'voluntario',
  EMPRESA = 'empresa',
  DOCENTE = 'docente',
}

export enum UserStatus {
  ACTIVO = 'activo',
  INACTIVO = 'inactivo',
  PENDIENTE_APROBACION = 'pendiente_aprobacion',
  SUSPENDIDO = 'suspendido',
}

export enum CourseStatus {
  DRAFT = 'draft',
  ACTIVO = 'activo',
  FINALIZADO = 'finalizado',
  CANCELADO = 'cancelado',
  PAUSADO = 'pausado',
}

export enum EnrollmentStatus {
  INSCRIPTO = 'inscripto',
  EN_PROGRESO = 'en_progreso',
  COMPLETADO = 'completado',
  ABANDONADO = 'abandonado',
  SUSPENSO = 'suspenso',
}

export enum JobOfferStatus {
  PENDIENTE = 'pendiente',
  APROBADA = 'aprobada',
  RECHAZADA = 'rechazada',
  CERRADA = 'cerrada',
  ARCHIVADA = 'archivada',
}

export enum ApplicationStatus {
  APLICADO = 'aplicado',
  EN_REVISION = 'en_revision',
  ENTREVISTA = 'entrevista',
  OFERTA = 'oferta',
  RECHAZADO = 'rechazado',
  CONTRATADO = 'contratado',
}

export enum CampaignStatus {
  DRAFT = 'draft',
  PUBLICADA = 'publicada',
  PAUSADA = 'pausada',
  ARCHIVADA = 'archivada',
}

export enum AuditAction {
  CREATE = 'create',
  READ = 'read',
  UPDATE = 'update',
  DELETE = 'delete',
  LOGIN = 'login',
  LOGOUT = 'logout',
  EXPORT = 'export',
  APPROVE = 'approve',
  REJECT = 'reject',
}

export enum NotificationType {
  INSCRIPCION_CURSO = 'inscripcion_curso',
  RECORDATORIO_CURSO = 'recordatorio_curso',
  OFERTA_COMPATIBLE = 'oferta_compatible',
  CAMPAÑA_NUEVA = 'campaña_nueva',
  CONTENIDO_APROBADO = 'contenido_aprobado',
  CONTENIDO_RECHAZADO = 'contenido_rechazado',
  APROBACION_VOLUNTARIO = 'aprobacion_voluntario',
  HITO_REGISTRADO = 'hito_registrado',
}

export enum NotificationChannel {
  EMAIL = 'email',
  PUSH = 'push',
  IN_APP = 'in_app',
  SMS = 'sms',
}

export enum FormativeArea {
  GASTRONOMIA = 'Gastronomía',
  OFICIOS = 'Oficios',
  ARTE_DISEÑO = 'Arte y Diseño',
  AGRICULTURA = 'Agricultura',
  TECNOLOGIA = 'Tecnología',
  ADMINISTRACION = 'Administración',
  SALUD = 'Salud',
  EDUCACION = 'Educación',
  SOSTENIBILIDAD = 'Sostenibilidad',
  OTRO = 'Otro',
}

export enum SkillLevel {
  BASICO = 'basico',
  INTERMEDIO = 'intermedio',
  AVANZADO = 'avanzado',
  EXPERTO = 'experto',
}

// ============================================================================
// INTERFACES DE USUARIO
// ============================================================================

export interface User {
  id: string;
  email: string;
  nombre: string;
  apellido: string;
  telefono: string;
  rol: Role;
  estado: UserStatus;
  avatar?: string;
  biografia?: string;
  fechaNacimiento?: Date;
  genero?: 'masculino' | 'femenino' | 'otro' | 'prefiero_no_decir';
  localidad?: string;
  direccion?: string;
  documento?: string;
  tipoDocumento?: 'DNI' | 'PASAPORTE' | 'CEDULA';

  // Autenticación
  passwordHash?: string;
  dos2FAEnabled: boolean;
  dos2FASecret?: string;
  lastLogin?: Date;
  loginAttempts: number;
  bloqueadoHasta?: Date;

  // OAuth
  googleId?: string;
  oauthProvider?: 'google' | 'facebook';

  // Metadata
  preferenciasNotificacion?: NotificationPreferences;
  metadatos?: Record<string, any>;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
}

export interface ParticipantProfile extends User {
  edad: number;
  estadoCivil?: 'soltero' | 'casado' | 'divorciado' | 'viudo' | 'union_estable';
  hijosACargo?: number;
  situacionHabitacional?: string;
  situacionPersonal: string;
  habilidades: Skill[];
  avances: Milestone[];
  documentacion: Document[];
  trayectoriaLaboral?: string;
  experiencia?: string;
  areasInteres?: FormativeArea[];
  disponibilidadHoraria?: string;
  voluntarioAsignado?: string;
  fechaRegistro: Date;
}

export interface Skill {
  id: string;
  nombre: string;
  descripcion: string;
  nivel: SkillLevel;
  adquiridoEn?: string;
  fechaAdquisicion?: Date;
  verificado: boolean;
}

export interface Milestone {
  id: string;
  titulo: string;
  descripcion: string;
  fecha: Date;
  responsable: string;
  tipo: 'logro' | 'dificultad' | 'cambio' | 'recomendacion';
  documentacion?: string;
  impacto?: string;
  seguimiento?: string;
}

export interface Document {
  id: string;
  nombre: string;
  tipo: 'DNI' | 'CERTIFICADO' | 'COMPROBANTE' | 'FOTO' | 'OTRO';
  url: string;
  fechaCarga: Date;
  cargadoPor: string;
  estado: 'pendiente_revision' | 'aprobado' | 'rechazado';
  observaciones?: string;
}

export interface AuditLog {
  id: string;
  usuarioId: string;
  accion: AuditAction;
  recurso: 'usuario' | 'curso' | 'oferta' | 'campaña' | 'perfil' | 'otro';
  recursoId: string;
  cambios?: {
    antes?: Record<string, any>;
    despues?: Record<string, any>;
  };
  ipOrigen: string;
  userAgent?: string;
  timestamp: Date;
  exitoso: boolean;
  mensajeError?: string;
}

export interface NotificationPreferences {
  emailCursos: boolean;
  emailOfertas: boolean;
  emailCampaigns: boolean;
  pushNotifications: boolean;
  smsNotifications: boolean;
  frecuenciaResumen?: 'diaria' | 'semanal' | 'mensual' | 'nunca';
}

// ============================================================================
// INTERFACES DE EDUCACIÓN
// ============================================================================

export interface Course {
  id: string;
  titulo: string;
  descripcion: string;
  areaFormativa: FormativeArea;
  docenteId: string;
  duracion: number;
  horasSemana?: number;
  fechaInicio: Date;
  fechaFin?: Date;
  capacidad: number;
  inscriptos: number;
  materiales: Material[];
  actividades: Activity[];
  estado: CourseStatus;
  imagenPortada?: string;
  prerequisites?: string[];
  certificado?: boolean;
  contentfulId?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Material {
  id: string;
  titulo: string;
  descripcion?: string;
  tipo: 'pdf' | 'video' | 'imagen' | 'enlace' | 'archivo';
  url: string;
  orden: number;
  requerido: boolean;
  cargadoPor: string;
  cargadoEn: Date;
  tamaño?: number;
}

export interface Activity {
  id: string;
  cursoId: string;
  titulo: string;
  descripcion: string;
  tipo: 'leccion' | 'tarea' | 'quiz' | 'proyecto' | 'asistencia';
  orden: number;
  fechaApertura: Date;
  fechaCierre?: Date;
  puntosMaximos?: number;
  porcentajePrueba?: number;
  requiereEntrega: boolean;
  formatoEntrega?: string;
  contentfulId?: string;
  createdAt: Date;
}

export interface Enrollment {
  id: string;
  participanteId: string;
  cursoId: string;
  estado: EnrollmentStatus;
  porcentajeCompletado: number;
  actividadesCompletadas: string[];
  asistencia: number;
  calificacion?: number;
  aprobado?: boolean;
  fechaInscripcion: Date;
  fechaComplecion?: Date;
  certificadoEmitido?: boolean;
  fechaCertificado?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface ActivityResponse {
  id: string;
  actividadId: string;
  participanteId: string;
  contenido: string;
  archivo?: {
    nombre: string;
    url: string;
    tipo: string;
  };
  puntuacion?: number;
  retroalimentacion?: string;
  evaluadoPor?: string;
  fechaEntrega: Date;
  fechaEvaluacion?: Date;
  createdAt: Date;
  updatedAt: Date;
}

// ============================================================================
// INTERFACES DE INSERCIÓN LABORAL
// ============================================================================

export interface JobOffer {
  id: string;
  empresaId: string;
  titulo: string;
  descripcion: string;
  requisitos: string[];
  tipoContrato: 'tiempo_completo' | 'tiempo_parcial' | 'pasantia' | 'voluntariado';
  salario?: {
    minimo?: number;
    maximo?: number;
    moneda: 'ARS' | 'USD';
  };
  ubicacion: string;
  remoto: boolean;
  habilidadesRequeridas: string[];
  areasFormativas: FormativeArea[];
  estado: JobOfferStatus;
  fechaPublicacion?: Date;
  fechaAprobacion?: Date;
  aprobadoPor?: string;
  fechaCierre: Date;
  fechaCierrePrevista?: Date;
  aplicacionesRecibidas: number;
  aplicacionesAceptadas: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface JobApplication {
  id: string;
  ofertaId: string;
  participanteId: string;
  estado: ApplicationStatus;
  cv: {
    url: string;
    generadoAutomaticamente: boolean;
  };
  cartaPresentacion?: string;
  seleccionadoParaEntrevista?: boolean;
  fechaEntrevista?: Date;
  resultadoEntrevista?: 'pendiente' | 'aprobado' | 'rechazado';
  contratado?: boolean;
  fechaContratacion?: Date;
  observaciones?: string;
  fechaAplicacion: Date;
  fechaUltimaActualizacion: Date;
}

export interface CVProfile {
  id: string;
  participanteId: string;
  nombre: string;
  apellido: string;
  telefono: string;
  email: string;
  ubicacion: string;
  enlacePortfolio?: string;
  sumario: string;
  experienciaLaboral: {
    puesto: string;
    empresa: string;
    fechaInicio: Date;
    fechaFin?: Date;
    descripcion: string;
  }[];
  educacionFormal: {
    titulo: string;
    institucion: string;
    fechaCompletacion?: Date;
  }[];
  cursosCompletados: {
    titulo: string;
    institucion: string;
    fechaCompletacion: Date;
    certificado: boolean;
  }[];
  habilidades: {
    nombre: string;
    nivel: SkillLevel;
  }[];
  idiomas: {
    idioma: string;
    nivel: 'basico' | 'intermedio' | 'avanzado';
  }[];
  ultimaGeneracion: Date;
  plantillaUsada: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface JobSuggestion {
  id: string;
  participanteId: string;
  ofertaId: string;
  matchScore: number;
  razonesMatch: string[];
  visto: boolean;
  fechaVisto?: Date;
  aplicado: boolean;
  fechaSugerencia: Date;
  fechaExpiracion: Date;
}

// ============================================================================
// INTERFACES DE CAMPAÑAS
// ============================================================================

export interface Campaign {
  id: string;
  titulo: string;
  descripcion: string;
  contenido: string;
  creadoPor: string;
  rol: Role;
  imagenes: {
    url: string;
    alt: string;
    principal: boolean;
    orden: number;
  }[];
  estado: CampaignStatus;
  fechaPublicacion?: Date;
  fechaArchivo?: Date;
  tags?: string[];
  categoria?: string;
  visibilidadPublica: boolean;
  visualizaciones: number;
  compartidas: number;
  slug: string;
  metaDescripcion?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CampaignContent {
  bloques: ContentBlock[];
}

export interface ContentBlock {
  id: string;
  tipo: 'texto' | 'imagen' | 'video' | 'galeria' | 'titulo' | 'separador';
  contenido: Record<string, any>;
  orden: number;
}

// ============================================================================
// INTERFACES DE NOTIFICACIONES
// ============================================================================

export interface Notification {
  id: string;
  usuarioId: string;
  tipo: NotificationType;
  titulo: string;
  descripcion: string;
  enlace?: string;
  canales: NotificationChannel[];
  leido: boolean;
  fechaLectura?: Date;
  datos?: Record<string, any>;
  createdAt: Date;
}

export interface EmailTemplate {
  id: string;
  tipo: NotificationType;
  asunto: string;
  contenidoHtml: string;
  variablesDisponibles: string[];
  activo: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface EmailSend {
  id: string;
  paraEmail: string;
  asunto: string;
  contenidoHtml: string;
  enviado: boolean;
  fechaEnvio?: Date;
  abierto: boolean;
  fechaApertura?: Date;
  clicEnlaces: number;
  plantilla: string;
  datos: Record<string, any>;
  createdAt: Date;
}

// ============================================================================
// INTERFACES ADICIONALES
// ============================================================================

export interface SystemConfig {
  id: string;
  jwtSecret: string;
  jwtExpiracion: number;
  refreshTokenExpiracion: number;
  dos2FAActivo: boolean;
  dos2FAMetodo: 'sms' | 'totp' | 'email';
  emailActivas: boolean;
  pushActivas: boolean;
  contentfulSpaceId?: string;
  contentfulAccessToken?: string;
  iaActiva: boolean;
  iaProvider: 'openai' | 'anthropic' | 'gemini';
  iaApiKey?: string;
  maxIntentosLogin: number;
  tiempoBloqueoMinutos: number;
  updatedAt: Date;
}

export interface DashboardStats {
  periodo: {
    inicio: Date;
    fin: Date;
  };
  participantesActivos: number;
  participantesNuevos: number;
  participantesInactivos: number;
  cursosActivos: number;
  cursosFinalizados: number;
  inscripcionesTotal: number;
  tasaCompletacion: number;
  ofertasPublicadas: number;
  ofertasAprobadas: number;
  aplicacionesRecibidas: number;
  contrataciones: number;
  habilidadesDesarrolladas: number;
  habilidadesPorArea: Record<string, number>;
  campanasPublicadas: number;
  visualizacionesTotales: number;
  generadoEn: Date;
}

export interface SearchFilters {
  query?: string;
  area?: FormativeArea;
  estado?: string;
  rol?: Role;
  ordenarPor?: 'nombre' | 'fecha' | 'relevancia';
  orden?: 'asc' | 'desc';
  pagina?: number;
  porPagina?: number;
  fechaInicio?: Date;
  fechaFin?: Date;
}

export interface PaginatedResponse<T> {
  datos: T[];
  paginacion: {
    total: number;
    pagina: number;
    porPagina: number;
    totalPaginas: number;
    tieneSiguiente: boolean;
    tieneAnterior: boolean;
  };
}

export interface ApiResponse<T> {
  exito: boolean;
  datos?: T;
  mensaje?: string;
  errores?: Array<{
    campo: string;
    mensaje: string;
  }>;
  timestamp: Date;
}

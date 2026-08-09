/**
 * COMPONENTES DEL SISTEMA BUEN VIVIR
 * Exports de todos los componentes de la aplicación
 */

// Componentes Base (Reutilizables)
export { Modal } from './base/Modal';
export { Card, CardHeader, CardBody, CardFooter } from './base/Card';
export { DataTable } from './base/DataTable';

// Componentes Principales
// The App component file is missing; disable the re-export until the component is added.
// export { App } from './App';
export { Dashboard } from './Dashboard';
export { Login } from './Login';
export { Register } from './Register';
export { LandingPage } from './LandingPage';

// Módulo de Educación
export { CourseCatalog } from './CourseCatalog';
export { CourseDetail } from './CourseDetail';
export { ProgressTracking } from './ProgressTracking';

// Módulo de Inserción Laboral
export { PublicJobOffers } from './PublicJobOffers';
// The JobApplicationForm component file is missing; disable the re-export until the component is added.
// export { JobApplicationForm } from './JobApplicationForm';
export { LaborInsertion } from './LaborInsertion';

// Módulo de Campañas
export { PublicCampaigns } from './PublicCampaigns';
export { CampaignManagement } from './CampaignManagement';

// Módulo de Promoción Humana (NUEVO)
export { PersonalPromotion } from './PersonalPromotion';

// Módulo de Notificaciones (NUEVO)
export { NotificationCenter } from './NotificationCenter';
export { NotificationSettings } from './NotificationSettings';

// Otros Módulos
export { UserManagement } from './UserManagement';
export { VolunteerUserManagement } from './VolunteerUserManagement';
export { TeacherPortal } from './TeacherPortal';
export { ParticipantTracking } from './ParticipantTracking';
export { AttendanceTracking } from './AttendanceTracking';
export { CVGenerator } from './CVGenerator';
export { JobOffers } from './JobOffers';
export { ForgotPassword } from './ForgotPassword';
export { AuditLog } from './AuditLog';

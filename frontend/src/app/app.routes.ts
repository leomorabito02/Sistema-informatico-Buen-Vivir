import { Routes } from "@angular/router";
import { authGuard } from "./core/guards/auth.guard";
import { guestGuard } from "./core/guards/guest.guard";
import { adminGuard } from "./core/guards/admin.guard";
import { adminOrVolunteerGuard } from "./core/guards/admin-volunteer.guard";
import { teacherGuard } from "./core/guards/teacher.guard";
import { participantGuard } from "./core/guards/participant.guard";

export const routes: Routes = [
  {
    path: "",
    loadComponent: () =>
      import("./pages/landing/landing.component").then(
        (m) => m.LandingComponent,
      ),
  },
  {
    path: "login",
    loadComponent: () =>
      import("./pages/login/login.component").then((m) => m.LoginComponent),
    canActivate: [guestGuard],
  },
  {
    path: "register",
    loadComponent: () =>
      import("./pages/register/register.component").then(
        (m) => m.RegisterComponent,
      ),
    canActivate: [guestGuard],
  },
  {
    path: "cv-public/:id",
    loadComponent: () =>
      import("./pages/cv-public/cv-public.component").then(
        (m) => m.CvPublicComponent,
      ),
  },
  {
    path: "pending-approval",
    loadComponent: () =>
      import("./pages/pending-approval/pending-approval.component").then(
        (m) => m.PendingApprovalComponent,
      ),
  },
  {
    path: "account-notice",
    loadComponent: () =>
      import("./pages/account-notice/account-notice.component").then(
        (m) => m.AccountNoticeComponent,
      ),
  },
  {
    path: "select-role",
    loadComponent: () =>
      import("./pages/select-role/select-role.component").then(
        (m) => m.SelectRoleComponent,
      ),
  },
  {
    path: "privacidad",
    loadComponent: () =>
      import("./pages/privacy-policy/privacy-policy.component").then(
        (m) => m.PrivacyPolicyComponent,
      ),
  },
  {
    path: "politica-privacidad",
    redirectTo: "privacidad",
    pathMatch: "full",
  },
  {
    path: "terminos",
    loadComponent: () =>
      import("./pages/terms-conditions/terms-conditions.component").then(
        (m) => m.TermsConditionsComponent,
      ),
  },
  {
    path: "terminos-condiciones",
    redirectTo: "terminos",
    pathMatch: "full",
  },
  {
    path: "cookies",
    loadComponent: () =>
      import("./pages/cookie-policy/cookie-policy.component").then(
        (m) => m.CookiePolicyComponent,
      ),
  },
  {
    path: "cookie-policy",
    redirectTo: "cookies",
    pathMatch: "full",
  },
  {
    path: "contacto",
    loadComponent: () =>
      import("./pages/contacto/contacto.component").then(
        (m) => m.ContactoComponent,
      ),
  },
  {
    // Campañas publicadas, sin iniciar sesión (HU-50, HU-51).
    path: "campanas-publico",
    loadComponent: () =>
      import("./pages/campanas-publico/campanas-publico.component").then(
        (m) => m.CampanasPublicoComponent,
      ),
  },
  {
    path: "mi-cuenta",
    redirectTo: "dashboard/mi-perfil",
    pathMatch: "prefix",
  },
  {
    path: "dashboard",
    loadComponent: () =>
      import("./pages/dashboard/dashboard.component").then(
        (m) => m.DashboardComponent,
      ),
    canActivate: [authGuard],
    children: [
      {
        path: "",
        pathMatch: "full",
        loadComponent: () =>
          import("./pages/dashboard/inicio/inicio.component").then(
            (m) => m.InicioComponent,
          ),
      },
      {
        path: "mi-perfil",
        loadComponent: () =>
          import("./pages/dashboard/mi-perfil/mi-perfil.component").then(
            (m) => m.MiPerfilComponent,
          ),
      },
      {
        path: "cursos",
        loadComponent: () =>
          import("./pages/dashboard/cursos/cursos.component").then(
            (m) => m.CursosComponent,
          ),
      },
      {
        // Detalle del curso con sus actividades y entregas. El componente
        // decide la vista por rol (gestión / docente / participante) y el
        // backend valida el acceso a cada curso.
        path: "cursos/:id",
        loadComponent: () =>
          import("./pages/dashboard/curso-detalle/curso-detalle.component").then(
            (m) => m.CursoDetalleComponent,
          ),
      },
      {
        path: "gestion-cursos",
        canActivate: [adminGuard],
        loadComponent: () =>
          import("./pages/dashboard/admin-cursos/admin-cursos.component").then(
            (m) => m.AdminCursosComponent,
          ),
      },
      {
        path: "mis-cursos",
        canActivate: [teacherGuard],
        loadComponent: () =>
          import("./pages/dashboard/mis-cursos/mis-cursos.component").then(
            (m) => m.MisCursosComponent,
          ),
      },
      {
        path: "progreso",
        canActivate: [teacherGuard],
        loadComponent: () =>
          import("./pages/dashboard/progreso/progreso.component").then(
            (m) => m.ProgresoComponent,
          ),
      },
      {
        path: "ofertas",
        loadComponent: () =>
          import("./pages/dashboard/ofertas/ofertas.component").then(
            (m) => m.OfertasComponent,
          ),
      },
      {
        path: "cv",
        loadComponent: () =>
          import("./pages/dashboard/cv/cv.component").then(
            (m) => m.CvComponent,
          ),
      },
      {
        path: "usuarios",
        loadComponent: () =>
          import("./pages/dashboard/usuarios/usuarios.component").then(
            (m) => m.UsuariosComponent,
          ),
      },
      {
        path: "seguimiento",
        loadComponent: () =>
          import("./pages/dashboard/seguimiento/seguimiento.component").then(
            (m) => m.SeguimientoComponent,
          ),
      },
      {
        path: "asistencia",
        canActivate: [teacherGuard],
        loadComponent: () =>
          import("./pages/dashboard/asistencia/asistencia.component").then(
            (m) => m.AsistenciaComponent,
          ),
      },
      {
        path: "inscripciones",
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/inscripciones/inscripciones.component").then(
            (m) => m.InscripcionesComponent,
          ),
      },
      {
        path: "asistencia-cursos",
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/asistencia-cursos/asistencia-cursos.component").then(
            (m) => m.AsistenciaCursosComponent,
          ),
      },
      {
        path: "asistencia-general",
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/asistencia-general/asistencia-general.component").then(
            (m) => m.AsistenciaGeneralComponent,
          ),
      },
      {
        path: "promocion-humana",
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/promocion-humana/promocion-humana.component").then(
            (m) => m.PromocionHumanaComponent,
          ),
      },
      {
        path: "promocion-humana/participantes/:participanteId",
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/promocion-humana/ficha/ficha.component").then(
            (m) => m.FichaComponent,
          ),
      },
      {
        path: "mi-proceso",
        canActivate: [participantGuard],
        loadComponent: () =>
          import("./pages/dashboard/mi-proceso/mi-proceso.component").then(
            (m) => m.MiProcesoComponent,
          ),
      },
      {
        path: "campanas",
        // Gestión de campañas: administración y voluntariado (HU-06).
        canActivate: [adminOrVolunteerGuard],
        loadComponent: () =>
          import("./pages/dashboard/campanas/campanas.component").then(
            (m) => m.CampanasComponent,
          ),
      },
      {
        path: "notificaciones",
        loadComponent: () =>
          import("./pages/dashboard/notificaciones/notificaciones.component").then(
            (m) => m.NotificacionesComponent,
          ),
      },
      {
        path: "indicadores",
        canActivate: [adminGuard],
        loadComponent: () =>
          import("./pages/dashboard/indicadores/dashboard-page.component").then(
            (m) => m.DashboardPageComponent,
          ),
      },
      {
        path: "ayuda",
        loadComponent: () =>
          import("./pages/dashboard/ayuda/ayuda.component").then(
            (m) => m.AyudaComponent,
          ),
      },
    ],
  },
];

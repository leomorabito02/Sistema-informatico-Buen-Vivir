import { Routes } from '@angular/router';
import { LandingComponent } from './pages/landing/landing.component';
import { LoginComponent } from './pages/login/login.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';
import { RegisterComponent } from './pages/register/register.component';
import { authGuard } from './core/guards/auth.guard';
import { guestGuard } from './core/guards/guest.guard';

export const routes: Routes = [
  {
    path: '',
    component: LandingComponent
  },
  {
    path: 'login',
    component: LoginComponent,
    canActivate: [guestGuard]
  },
  {
    path: 'register',
    component: RegisterComponent,
    canActivate: [guestGuard]
  },
  {
    path: 'pending-approval',
    loadComponent: () => import('./pages/pending-approval/pending-approval.component').then(m => m.PendingApprovalComponent)
  },
  {
    path: 'account-notice',
    loadComponent: () => import('./pages/account-notice/account-notice.component').then(m => m.AccountNoticeComponent)
  },
  {
    path: 'select-role',
    loadComponent: () => import('./pages/select-role/select-role.component').then(m => m.SelectRoleComponent)
  },
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [authGuard],
    children: [
      {
        path: 'mi-perfil',
        loadComponent: () => import('./pages/dashboard/mi-perfil/mi-perfil.component').then(m => m.MiPerfilComponent)
      },
      {
        path: 'cursos',
        loadComponent: () => import('./pages/dashboard/cursos/cursos.component').then(m => m.CursosComponent)
      },
      {
        path: 'mis-cursos',
        loadComponent: () => import('./pages/dashboard/mis-cursos/mis-cursos.component').then(m => m.MisCursosComponent)
      },
      {
        path: 'progreso',
        loadComponent: () => import('./pages/dashboard/progreso/progreso.component').then(m => m.ProgresoComponent)
      },
      {
        path: 'ofertas',
        loadComponent: () => import('./pages/dashboard/ofertas/ofertas.component').then(m => m.OfertasComponent)
      },
      {
        path: 'cv',
        loadComponent: () => import('./pages/dashboard/cv/cv.component').then(m => m.CvComponent)
      },
      {
        path: 'usuarios',
        loadComponent: () => import('./pages/dashboard/usuarios/usuarios.component').then(m => m.UsuariosComponent)
      },
      {
        path: 'seguimiento',
        loadComponent: () => import('./pages/dashboard/seguimiento/seguimiento.component').then(m => m.SeguimientoComponent)
      },
      {
        path: 'asistencia',
        loadComponent: () => import('./pages/dashboard/asistencia/asistencia.component').then(m => m.AsistenciaComponent)
      },
      {
        path: 'campanas',
        loadComponent: () => import('./pages/dashboard/campanas/campanas.component').then(m => m.CampanasComponent)
      },
      {
        path: 'insercion',
        loadComponent: () => import('./pages/dashboard/insercion/insercion.component').then(m => m.InsercionComponent)
      },
      {
        path: 'participantes',
        loadComponent: () => import('./pages/dashboard/participantes/participantes.component').then(m => m.ParticipantesComponent)
      },
      {
        path: 'notificaciones',
        loadComponent: () => import('./pages/dashboard/notificaciones/notificaciones.component').then(m => m.NotificacionesComponent)
      }
    ]
  }
];

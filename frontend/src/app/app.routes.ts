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
      }
    ]
  }
];

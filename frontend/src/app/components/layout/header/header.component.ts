import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import { MeResponse } from '../../../core/models/auth.model';
import { LucideAngularModule, Bell, Settings, LogOut } from 'lucide-angular';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: './header.component.html',
  styleUrl: './header.component.css'
})
export class HeaderComponent implements OnInit {
  private authService = inject(AuthService);

  roleName: string = '';
  
  // Icons
  Bell = Bell;
  Settings = Settings;
  LogOut = LogOut;

  ngOnInit(): void {
    this.authService.isAuthenticated().subscribe({
      next: (response: MeResponse) => {
        this.roleName = this.mapRoleIdToName(response.role_id ?? response.user?.role_id ?? '');
      },
      error: () => {
        this.roleName = '';
      }
    });
  }

  mapRoleIdToName(roleId: string): string {
    if (!roleId) return '';
    const id = roleId.toLowerCase();
    if (id === 'role_participant' || id === 'participante') return 'participante';
    if (id === 'role_admin' || id === 'responsable') return 'responsable';
    if (id === 'role_volunteer' || id === 'voluntario') return 'voluntario';
    if (id === 'role_teacher' || id === 'docente') return 'docente';
    if (id === 'role_company' || id === 'empresa') return 'empresa';
    return id.replace('role_', '');
  }

  logout(): void {
    this.authService.logout();
  }
}

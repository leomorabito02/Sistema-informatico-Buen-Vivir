import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule } from '@angular/forms';
import { RouterOutlet } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { HeaderComponent } from '../../components/layout/header/header.component';
import { SidebarComponent } from '../../components/layout/sidebar/sidebar.component';
import { FooterComponent } from '../../components/layout/footer/footer.component';

import { ToastModule } from 'primeng/toast';
import { CardModule } from 'primeng/card';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterOutlet, LucideAngularModule, HeaderComponent, SidebarComponent, FooterComponent, ToastModule, CardModule],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent {
  
}

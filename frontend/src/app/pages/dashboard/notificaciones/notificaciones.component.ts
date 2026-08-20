import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { TagModule } from 'primeng/tag';
import { BadgeModule } from 'primeng/badge';

@Component({
  selector: 'app-notificaciones',
  standalone: true,
  imports: [CommonModule, CardModule, ButtonModule, TagModule, BadgeModule],
  templateUrl: './notificaciones.component.html',
  styleUrls: ['./notificaciones.component.css']
})
export class NotificacionesComponent implements OnInit {
  notifications: any[] = [];

  ngOnInit() {}

  markAllAsRead() {
    this.notifications.forEach(n => n.read = true);
  }

  markAsRead(n: any) {
    n.read = true;
  }
}

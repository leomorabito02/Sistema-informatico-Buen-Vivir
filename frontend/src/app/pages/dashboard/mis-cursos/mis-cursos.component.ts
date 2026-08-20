import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { TagModule } from 'primeng/tag';
import { CardModule } from 'primeng/card';

@Component({
  selector: 'app-mis-cursos',
  standalone: true,
  imports: [CommonModule, TableModule, ButtonModule, TagModule, CardModule],
  templateUrl: './mis-cursos.component.html',
  styleUrls: ['./mis-cursos.component.css']
})
export class MisCursosComponent implements OnInit {
  assignedCourses: any[] = [];

  ngOnInit() {}

  getSeverity(status: string): any {
    switch (status) {
      case 'Activo': return 'success';
      case 'Planificado': return 'info';
      default: return undefined;
    }
  }
}

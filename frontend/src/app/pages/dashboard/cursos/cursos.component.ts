import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { TagModule } from 'primeng/tag';
import { InputTextModule } from 'primeng/inputtext';

interface Curso {
  title: string;
  category: string;
  instructor: string;
  schedule: string;
  seats: number;
  status: string;
  duration?: string;
  level?: string;
}


@Component({
  selector: 'app-cursos',
  standalone: true,
  imports: [CommonModule, FormsModule, TableModule, ButtonModule, CardModule, TagModule, InputTextModule],
  templateUrl: './cursos.component.html',
  styleUrls: ['./cursos.component.css']
})
export class CursosComponent implements OnInit {
  searchTerm: string = '';
  cursos: Curso[] = [];
  filteredCursos: Curso[] = [];

  ngOnInit() {}

  filterCursos() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredCursos = this.cursos.filter(c => 
      c.title.toLowerCase().includes(term) ||
      c.category.toLowerCase().includes(term) ||
      c.instructor.toLowerCase().includes(term)
    );
  }

  getSeverity(status: string): string | undefined {
    switch (status) {
      case 'Inscripciones Abiertas': return 'success';
      case 'En Curso': return 'info';
      case 'Próximamente': return 'warning';
      default: return undefined;
    }
  }
}

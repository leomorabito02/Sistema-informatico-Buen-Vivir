import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { TagModule } from 'primeng/tag';
import { InputTextModule } from 'primeng/inputtext';

@Component({
  selector: 'app-cursos',
  standalone: true,
  imports: [CommonModule, FormsModule, TableModule, ButtonModule, CardModule, TagModule, InputTextModule],
  templateUrl: './cursos.component.html',
  styleUrls: ['./cursos.component.css']
})
export class CursosComponent implements OnInit {
  searchTerm: string = '';
  cursos: any[] = [];
  filteredCursos: any[] = [];

  ngOnInit() {}

  filterCursos() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredCursos = this.cursos.filter(c => 
      c.title.toLowerCase().includes(term) ||
      c.category.toLowerCase().includes(term) ||
      c.instructor.toLowerCase().includes(term)
    );
  }

  getSeverity(status: string): any {
    switch (status) {
      case 'Inscripciones Abiertas': return 'success';
      case 'En Curso': return 'info';
      case 'Próximamente': return 'warning';
      default: return undefined;
    }
  }
}

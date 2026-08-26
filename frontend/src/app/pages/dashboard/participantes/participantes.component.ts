import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { TagModule } from 'primeng/tag';
import { InputTextModule } from 'primeng/inputtext';
import { CardModule } from 'primeng/card';

interface Participant {
  name: string;
  dni: string;
  course: string;
  status: string;
  attendance?: number;
}


@Component({
  selector: 'app-participantes',
  standalone: true,
  imports: [CommonModule, FormsModule, TableModule, ButtonModule, TagModule, InputTextModule, CardModule],
  templateUrl: './participantes.component.html',
  styleUrls: ['./participantes.component.css']
})
export class ParticipantesComponent implements OnInit {
  searchTerm: string = '';

  participants: Participant[] = [];
  filteredParticipants: Participant[] = [];

  ngOnInit() {}

  filter() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredParticipants = this.participants.filter(p =>
      p.name.toLowerCase().includes(term) ||
      p.dni.includes(term) ||
      p.course.toLowerCase().includes(term)
    );
  }
}

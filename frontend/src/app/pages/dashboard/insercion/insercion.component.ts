import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardModule } from 'primeng/card';
import { TableModule } from 'primeng/table';
import { TagModule } from 'primeng/tag';
import { ButtonModule } from 'primeng/button';

@Component({
  selector: 'app-insercion',
  standalone: true,
  imports: [CommonModule, CardModule, TableModule, TagModule, ButtonModule],
  templateUrl: './insercion.component.html',
  styleUrls: ['./insercion.component.css']
})
export class InsercionComponent implements OnInit {
  placements: any[] = [];
  linkedParticipantsCount: number = 0;
  retentionRate: string = '0%';
  activeCompaniesCount: number = 0;

  ngOnInit() {}
}

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ProgressBarModule } from 'primeng/progressbar';
import { CardModule } from 'primeng/card';
import { TagModule } from 'primeng/tag';
import { ButtonModule } from 'primeng/button';

@Component({
  selector: 'app-progreso',
  standalone: true,
  imports: [CommonModule, ProgressBarModule, CardModule, TagModule, ButtonModule],
  templateUrl: './progreso.component.html',
  styleUrls: ['./progreso.component.css']
})
export class ProgresoComponent implements OnInit {
  overallProgress = 0;
  activeCourses: any[] = [];
  completedCourses: any[] = [];

  ngOnInit() {}
}

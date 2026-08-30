import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";

interface AttendanceLog {
  date: string;
  course: string;
  status: string;
  duration?: number;
}

@Component({
  selector: "app-asistencia",
  standalone: true,
  imports: [CommonModule, TableModule, ButtonModule, TagModule, CardModule],
  templateUrl: "./asistencia.component.html",
  styleUrls: ["./asistencia.component.css"],
})
export class AsistenciaComponent implements OnInit {
  attendanceLogs: AttendanceLog[] = [];

  ngOnInit() {}
}

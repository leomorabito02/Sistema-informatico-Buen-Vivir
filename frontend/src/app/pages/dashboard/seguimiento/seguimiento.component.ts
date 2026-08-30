import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";

interface SeguimientoRecord {
  name: string;
  course: string;
  attendance: number;
  grade: number;
  status: string;
}

@Component({
  selector: "app-seguimiento",
  standalone: true,
  imports: [CommonModule, TableModule, ButtonModule, TagModule, CardModule],
  templateUrl: "./seguimiento.component.html",
  styleUrls: ["./seguimiento.component.css"],
})
export class SeguimientoComponent implements OnInit {
  records: SeguimientoRecord[] = [];

  ngOnInit() {}

  getSeverity(status: string): string | undefined {
    switch (status) {
      case "Excelente Avance":
        return "success";
      case "Asistencia Regular":
        return "info";
      case "En Riesgo de Deserción":
        return "danger";
      default:
        return undefined;
    }
  }
}

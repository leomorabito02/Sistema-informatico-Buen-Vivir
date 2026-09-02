import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TableModule } from "primeng/table";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { CardModule } from "primeng/card";
import { SelectModule } from "primeng/select";
import { InputTextModule } from "primeng/inputtext";
import { DatePickerModule } from "primeng/datepicker";
import { CheckboxModule } from "primeng/checkbox";
import { ProgressSpinnerModule } from "primeng/progressspinner";
import { MessageModule } from "primeng/message";
import { EducationService } from "../../../services/education.service";
import {
  Activity,
  Attendance,
  Course,
} from "../../../core/models/education.model";

interface Option {
  label: string;
  value: string;
}
interface AttendanceDraft {
  participanteId: string;
  presente: boolean;
}

@Component({
  selector: "app-asistencia",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    TagModule,
    CardModule,
    SelectModule,
    InputTextModule,
    DatePickerModule,
    CheckboxModule,
    ProgressSpinnerModule,
    MessageModule,
  ],
  templateUrl: "./asistencia.component.html",
  styleUrls: ["./asistencia.component.css"],
})
export class AsistenciaComponent implements OnInit {
  private education = inject(EducationService);

  loadingCourses = true;
  error = "";
  info = "";

  courseOptions: Option[] = [];
  selectedCourseId: string | null = null;

  loadingActivities = false;
  activities: Activity[] = [];

  // crear actividad
  nuevaActividadTitulo = "";
  nuevaActividadFecha: Date | null = null;
  creandoActividad = false;

  // asistencia de la actividad seleccionada
  selectedActivityId: string | null = null;
  loadingAttendance = false;
  attendance: Attendance[] = [];
  draft: AttendanceDraft[] = [];
  guardando = false;

  ngOnInit() {
    this.education.listMyCourses().subscribe({
      next: (courses: Course[]) => {
        this.courseOptions = courses.map((c) => ({
          label: c.title || c.id,
          value: c.id,
        }));
        this.loadingCourses = false;
        if (this.courseOptions.length === 1) {
          this.selectedCourseId = this.courseOptions[0].value;
          this.onCourseChange();
        }
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingCourses = false;
      },
    });
  }

  onCourseChange() {
    this.selectedActivityId = null;
    this.attendance = [];
    this.draft = [];
    if (!this.selectedCourseId) return;
    this.loadingActivities = true;
    this.error = "";
    this.education.listActivities(this.selectedCourseId).subscribe({
      next: (acts) => {
        this.activities = acts;
        this.loadingActivities = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingActivities = false;
      },
    });
  }

  crearActividad() {
    if (
      !this.selectedCourseId ||
      !this.nuevaActividadTitulo ||
      !this.nuevaActividadFecha
    )
      return;
    this.creandoActividad = true;
    this.error = "";
    this.education
      .createActivity(this.selectedCourseId, {
        titulo: this.nuevaActividadTitulo,
        fecha: this.nuevaActividadFecha.toISOString(),
      })
      .subscribe({
        next: (act) => {
          this.activities = [...this.activities, act];
          this.nuevaActividadTitulo = "";
          this.nuevaActividadFecha = null;
          this.creandoActividad = false;
          this.info = "Actividad creada.";
        },
        error: (err: Error) => {
          this.error = err.message;
          this.creandoActividad = false;
        },
      });
  }

  selectActivity(id: string) {
    this.selectedActivityId = id;
    this.info = "";
    this.loadingAttendance = true;
    this.error = "";
    this.education.listAttendance(id).subscribe({
      next: (list) => {
        this.attendance = list;
        this.draft = list.map((a) => ({
          participanteId: a.participanteId,
          presente: a.presente,
        }));
        this.loadingAttendance = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loadingAttendance = false;
      },
    });
  }

  addRow() {
    this.draft = [...this.draft, { participanteId: "", presente: true }];
  }

  removeRow(i: number) {
    this.draft = this.draft.filter((_, idx) => idx !== i);
  }

  guardar() {
    if (!this.selectedActivityId) return;
    const payload = this.draft
      .map((d) => ({
        participanteId: d.participanteId.trim(),
        presente: d.presente,
      }))
      .filter((d) => d.participanteId);
    if (payload.length === 0) return;
    this.guardando = true;
    this.error = "";
    this.education
      .registerAttendance(this.selectedActivityId, payload)
      .subscribe({
        next: () => {
          this.guardando = false;
          this.info = "Asistencia registrada.";
          this.selectActivity(this.selectedActivityId!);
        },
        error: (err: Error) => {
          this.error = err.message;
          this.guardando = false;
        },
      });
  }
}

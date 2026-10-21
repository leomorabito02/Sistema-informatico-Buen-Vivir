import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { CardModule } from "primeng/card";
import { forkJoin } from "rxjs";

import { HumanPromotionService } from "../../../services/human-promotion.service";
import { EducationService } from "../../../services/education.service";
import { HitoParticipante } from "../../../core/models/human-promotion.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { formatDMY } from "../../../shared/utils/date-ymd";
import {
  ItemLineaTiempo,
  mezclarLineaDeTiempo,
  nombresDeCursos,
} from "../../../shared/utils/linea-tiempo";

/**
 * Línea de tiempo propia del participante (HU-34). Mezcla, por fecha, los
 * hitos que el equipo marcó como visibles con su actividad en el centro
 * (inscripciones, cursos completados y asistencia por mes). El backend ya
 * filtra todo lo confidencial. Sin cantidades, porcentajes ni datos
 * numéricos (pedido del cliente).
 */
@Component({
  selector: "app-mi-proceso",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    CardModule,
  ],
  templateUrl: "./mi-proceso.component.html",
  styleUrls: ["./mi-proceso.component.css"],
})
export class MiProcesoComponent implements OnInit {
  private readonly service = inject(HumanPromotionService);
  private readonly education = inject(EducationService);

  readonly formatDMY = formatDMY;
  loading = true;
  error = "";
  items: ItemLineaTiempo<HitoParticipante>[] = [];

  ngOnInit() {
    this.cargar();
  }

  trackByKey(_: number, item: ItemLineaTiempo<HitoParticipante>) {
    return item.key;
  }

  cargar() {
    this.loading = true;
    this.error = "";
    forkJoin({
      linea: this.service.miLineaDeTiempo(),
      cursos: nombresDeCursos(this.education),
    }).subscribe({
      next: ({ linea, cursos }) => {
        this.items = mezclarLineaDeTiempo(linea.hitos, linea.actividad, cursos);
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }
}

import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { CardModule } from "primeng/card";

import { HumanPromotionService } from "../../../services/human-promotion.service";
import { HitoParticipante } from "../../../core/models/human-promotion.model";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { formatDMY } from "../../../shared/utils/date-ymd";

/**
 * Línea de tiempo propia del participante (HU-34). Solo muestra los hitos
 * que el equipo marcó como visibles; el backend ya filtra todo lo
 * confidencial. Sin cantidades, porcentajes ni datos numéricos (pedido
 * del cliente).
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

  readonly formatDMY = formatDMY;
  loading = true;
  error = "";
  hitos: HitoParticipante[] = [];

  ngOnInit() {
    this.service.miLineaDeTiempo().subscribe({
      next: (hitos) => {
        this.hitos = hitos;
        this.loading = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.loading = false;
      },
    });
  }
}

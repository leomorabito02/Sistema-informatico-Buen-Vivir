import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { CardModule } from "primeng/card";
import { InputTextModule } from "primeng/inputtext";
import { IconFieldModule } from "primeng/iconfield";
import { InputIconModule } from "primeng/inputicon";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";

import { HumanPromotionService } from "../../../services/human-promotion.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { ResultadoBusqueda } from "../../../core/models/human-promotion.model";
import { coincideBusqueda } from "../../../shared/utils/busqueda";

/**
 * Promoción Humana: listado de participantes (HU-36) y acceso a su ficha. Solo admin y
 * voluntario. El buscador filtra el listado por nombre o apellido.
 */
@Component({
  selector: "app-promocion-humana",
  standalone: true,
  imports: [
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
    CommonModule,
    FormsModule,
    RouterLink,
    CardModule,
    InputTextModule,
    IconFieldModule,
    InputIconModule,
    ButtonModule,
    TagModule,
  ],
  templateUrl: "./promocion-humana.component.html",
  styleUrls: ["./promocion-humana.component.css"],
})
export class PromocionHumanaComponent implements OnInit {
  private readonly service = inject(HumanPromotionService);

  texto = "";
  cargando = true;
  error = "";
  participantes: ResultadoBusqueda[] = [];

  ngOnInit() {
    this.cargar();
  }

  cargar() {
    this.cargando = true;
    this.error = "";
    this.service.listarParticipantes().subscribe({
      next: (lista) => {
        this.participantes = lista;
        this.cargando = false;
      },
      error: (err: Error) => {
        this.error = err.message;
        this.cargando = false;
      },
    });
  }

  /** Participantes que coinciden con el buscador (todos si está vacío). */
  get filtrados(): ResultadoBusqueda[] {
    return this.participantes.filter((p) =>
      coincideBusqueda(this.texto, p.nombre, p.apellido),
    );
  }

  trackById(_: number, r: ResultadoBusqueda) {
    return r.participante_id;
  }

  nombreCompleto(r: ResultadoBusqueda): string {
    return `${r.nombre} ${r.apellido}`.trim() || r.participante_id;
  }
}

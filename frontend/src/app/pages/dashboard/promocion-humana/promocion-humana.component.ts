import { Component, DestroyRef, OnInit, inject } from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { CardModule } from "primeng/card";
import { InputTextModule } from "primeng/inputtext";
import { IconFieldModule } from "primeng/iconfield";
import { InputIconModule } from "primeng/inputicon";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { Subject, of } from "rxjs";
import {
  catchError,
  debounceTime,
  distinctUntilChanged,
  map,
  switchMap,
  tap,
} from "rxjs/operators";

import { HumanPromotionService } from "../../../services/human-promotion.service";
import { LoadingStateComponent } from "../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../shared/components/empty-state.component";
import { ResultadoBusqueda } from "../../../core/models/human-promotion.model";

/** Tiempo de espera antes de consultar mientras se escribe. */
const DEBOUNCE_MS = 400;
const MIN_CARACTERES = 2;

/**
 * Promoción Humana: búsqueda de participantes por nombre (HU-36) y acceso
 * a su ficha de proceso personal. Solo admin y voluntario (guard + backend).
 * La búsqueda espera a que se deje de escribir (debounce) y cancela la
 * consulta anterior si llega una nueva (switchMap).
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
  private readonly destroyRef = inject(DestroyRef);
  private readonly busqueda$ = new Subject<string>();

  readonly minCaracteres = MIN_CARACTERES;
  texto = "";
  buscando = false;
  buscado = false;
  error = "";
  resultados: ResultadoBusqueda[] = [];

  ngOnInit() {
    this.busqueda$
      .pipe(
        map((t) => t.trim()),
        debounceTime(DEBOUNCE_MS),
        distinctUntilChanged(),
        tap((t) => {
          this.error = "";
          this.buscando = t.length >= MIN_CARACTERES;
          if (t.length < MIN_CARACTERES) {
            this.resultados = [];
            this.buscado = false;
          }
        }),
        switchMap((t) =>
          t.length < MIN_CARACTERES
            ? of(null)
            : this.service.buscarParticipantes(t).pipe(
                catchError((err: Error) => {
                  this.error = err.message;
                  return of([] as ResultadoBusqueda[]);
                }),
              ),
        ),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((res) => {
        this.buscando = false;
        if (res === null) return;
        this.resultados = res;
        this.buscado = true;
      });
  }

  onTextoChange(valor: string) {
    this.busqueda$.next(valor);
  }

  nombreCompleto(r: ResultadoBusqueda): string {
    return `${r.nombre} ${r.apellido}`.trim() || r.participante_id;
  }
}

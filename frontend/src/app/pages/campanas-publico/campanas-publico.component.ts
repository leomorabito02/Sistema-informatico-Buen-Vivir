import { Component, DestroyRef, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { RouterLink } from "@angular/router";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import {
  Subject,
  debounceTime,
  distinctUntilChanged,
  switchMap,
  catchError,
  of,
  tap,
} from "rxjs";
import {
  LucideAngularModule,
  ArrowLeft,
  Megaphone,
  Search,
  Calendar,
} from "lucide-angular";
import { FooterComponent } from "../../components/layout/footer/footer.component";
import { CampaignsService } from "../../services/campaigns.service";
import {
  Campana,
  DEFAULT_CAMPAIGN_PLACEHOLDER,
} from "../../core/models/campaign.model";
import { ComentariosComponent } from "../../shared/components/comentarios.component";

/**
 * Campañas publicadas, para cualquier visitante sin iniciar sesión
 * (HU-50), con búsqueda por título (HU-51) y comentarios (HU-24). La
 * portada enlaza acá.
 */
@Component({
  selector: "app-campanas-publico",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    LucideAngularModule,
    FooterComponent,
    ComentariosComponent,
  ],
  templateUrl: "./campanas-publico.component.html",
  styleUrl: "./campanas-publico.component.css",
})
export class CampanasPublicoComponent implements OnInit {
  readonly ArrowLeftIcon = ArrowLeft;
  readonly MegaphoneIcon = Megaphone;
  readonly SearchIcon = Search;
  readonly CalendarIcon = Calendar;

  private readonly service = inject(CampaignsService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly busqueda$ = new Subject<string>();

  campanas: Campana[] = [];
  /** Total de publicadas que coinciden con la búsqueda (HU-50). */
  total = 0;
  cargandoMas = false;
  private ultimaBusqueda = "";
  /** Campañas con los comentarios abiertos: se piden recién al abrirlos. */
  readonly comentariosAbiertos = new Set<string>();
  busqueda = "";
  cargando = true;
  error = "";

  ngOnInit(): void {
    this.busqueda$
      .pipe(
        debounceTime(300),
        distinctUntilChanged(),
        tap(() => {
          this.cargando = true;
          this.error = "";
        }),
        // switchMap: si se sigue escribiendo, gana la última búsqueda.
        switchMap((q) => {
          this.ultimaBusqueda = q;
          return this.service.publicasPagina(q, 0).pipe(
            catchError((e: Error) => {
              this.error = e.message;
              return of({ items: [] as Campana[], total: 0 });
            }),
          );
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((p) => {
        this.campanas = p.items;
        this.total = p.total;
        this.cargando = false;
      });
    this.busqueda$.next("");
  }

  get hayMas(): boolean {
    return this.campanas.length < this.total;
  }

  /** HU-50: trae la página siguiente y la suma a la lista. */
  verMas(): void {
    if (this.cargandoMas || !this.hayMas) return;
    this.cargandoMas = true;
    this.service
      .publicasPagina(this.ultimaBusqueda, this.campanas.length)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (p) => {
          this.campanas = [...this.campanas, ...p.items];
          this.total = p.total;
          this.cargandoMas = false;
        },
        error: (e: Error) => {
          this.error = e.message;
          this.cargandoMas = false;
        },
      });
  }

  buscar(texto: string): void {
    this.busqueda$.next(texto.trim());
  }

  /** "2026-11-05" → "05/11/2026" (sin pasar por Date, para no correr el día). */
  fechaDMY(fecha: string): string {
    const [a, m, d] = fecha.split("-");
    return d && m && a ? `${d}/${m}/${a}` : fecha;
  }

  alternarComentarios(id: string): void {
    if (this.comentariosAbiertos.has(id)) this.comentariosAbiertos.delete(id);
    else this.comentariosAbiertos.add(id);
  }

  onImageError(event: Event): void {
    const el = event.target as HTMLImageElement;
    if (el && el.src !== DEFAULT_CAMPAIGN_PLACEHOLDER) {
      el.src = DEFAULT_CAMPAIGN_PLACEHOLDER;
    }
  }
}

import { Component, EventEmitter, Input, Output, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { ButtonModule } from "primeng/button";
import { AutoCompleteModule } from "primeng/autocomplete";
import { SelectModule } from "primeng/select";
import { TagModule } from "primeng/tag";
import { MessageService } from "primeng/api";
import {
  LucideAngularModule,
  Check,
  CheckCheck,
  Plus,
  Sparkles,
  X,
} from "lucide-angular";

import { SkillsService } from "../../../../services/skills.service";
import {
  SKILL_CATEGORIES,
  Skill,
  SkillCategory,
  SkillInput,
  SkillSuggestion,
  skillKey,
} from "../../../../core/models/skills.model";
import { LoadingStateComponent } from "../../../../shared/components/loading-state.component";
import { ErrorStateComponent } from "../../../../shared/components/error-state.component";
import { EmptyStateComponent } from "../../../../shared/components/empty-state.component";

/** Igual que maxSkillsPerActivity del backend (skill_service.go). */
export const MAX_SKILLS_PER_ACTIVITY = 8;

/**
 * Sección "Habilidades que desarrolla esta actividad" de la ventana de
 * alta / edición de actividad (HU-21 / HU-58).
 *
 * No guarda nada por su cuenta: mantiene la lista `skills` del formulario
 * y la devuelve con `skillsChange`. El componente padre la guarda cuando
 * se guarda la actividad. Así las sugerencias de la IA que nadie aceptó
 * nunca llegan a la base (viven solo acá, como pide el DER).
 */
@Component({
  selector: "app-activity-skills",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    AutoCompleteModule,
    SelectModule,
    TagModule,
    LucideAngularModule,
    LoadingStateComponent,
    ErrorStateComponent,
    EmptyStateComponent,
  ],
  templateUrl: "./activity-skills.component.html",
  styleUrls: ["./activity-skills.component.css"],
})
export class ActivitySkillsComponent {
  private readonly skillsApi = inject(SkillsService);
  private readonly toast = inject(MessageService);

  @Input() cursoId = "";
  /** Vacío mientras la actividad todavía no se creó. */
  @Input() actividadId = "";
  @Input() titulo = "";
  @Input() consigna = "";
  /** Habilidades asignadas en el formulario (todavía sin guardar). */
  @Input() skills: SkillInput[] = [];
  @Output() skillsChange = new EventEmitter<SkillInput[]>();

  readonly categorias = [...SKILL_CATEGORIES];
  readonly maxSkills = MAX_SKILLS_PER_ACTIVITY;
  readonly Sparkles = Sparkles;
  readonly Check = Check;
  readonly CheckCheck = CheckCheck;
  readonly Plus = Plus;
  readonly X = X;

  // --- sugerencias de la IA ---
  sugerencias: SkillSuggestion[] = [];
  cargando = false;
  errorSugerencias = "";
  /** true después de un pedido que no trajo nada nuevo. */
  sinSugerencias = false;

  // --- carga manual ---
  manual: Skill | string | null = null;
  manualCategoria: SkillCategory | null = null;
  catalogoFiltrado: Skill[] = [];

  /** El botón de la IA se habilita recién con título y consigna. */
  get puedeSugerir(): boolean {
    return !!(this.titulo.trim() && this.consigna.trim() && this.cursoId);
  }

  get lleno(): boolean {
    return this.skills.length >= this.maxSkills;
  }

  sugerir() {
    if (!this.puedeSugerir || this.cargando) return;
    this.cargando = true;
    this.errorSugerencias = "";
    this.sinSugerencias = false;
    this.sugerencias = [];

    this.skillsApi
      .suggest({
        cursoId: this.cursoId,
        actividadId: this.actividadId || undefined,
        titulo: this.titulo.trim(),
        consigna: this.consigna.trim(),
        excluir: this.skills.map((s) => s.nombre),
      })
      .subscribe({
        next: (res) => {
          this.cargando = false;
          // El backend ya filtra repetidas; esto cubre las que se agregaron
          // a mano mientras la IA respondía.
          this.sugerencias = res.sugerencias.filter(
            (s) => !this.yaAsignada(s.nombre),
          );
          this.sinSugerencias = this.sugerencias.length === 0;
        },
        error: (err: Error) => {
          this.cargando = false;
          this.errorSugerencias = err.message;
        },
      });
  }

  aceptar(s: SkillSuggestion) {
    if (this.lleno) {
      this.avisarLleno();
      return;
    }
    this.agregar({ id: s.id, nombre: s.nombre, categoria: s.categoria });
    this.sugerencias = this.sugerencias.filter((x) => x !== s);
  }

  descartar(s: SkillSuggestion) {
    this.sugerencias = this.sugerencias.filter((x) => x !== s);
  }

  aceptarTodas() {
    const lugar = this.maxSkills - this.skills.length;
    const aceptadas = this.sugerencias.slice(0, Math.max(0, lugar));
    const nuevas = aceptadas.map((s) => ({
      id: s.id,
      nombre: s.nombre,
      categoria: s.categoria,
    }));
    this.emitir([...this.skills, ...nuevas]);
    this.sugerencias = this.sugerencias.slice(aceptadas.length);
    if (this.sugerencias.length > 0) this.avisarLleno();
  }

  quitar(skill: SkillInput) {
    this.emitir(this.skills.filter((s) => s !== skill));
  }

  buscarCatalogo(event: { query: string }) {
    this.skillsApi.searchCatalog(event.query).subscribe({
      next: (lista) => {
        this.catalogoFiltrado = lista.filter((s) => !this.yaAsignada(s.nombre));
      },
      error: () => {
        // Sin catálogo igual se puede escribir una habilidad nueva.
        this.catalogoFiltrado = [];
      },
    });
  }

  /** Nombre escrito o elegido en el autocompletado. */
  get manualNombre(): string {
    if (!this.manual) return "";
    return typeof this.manual === "string"
      ? this.manual.trim()
      : this.manual.nombre;
  }

  /**
   * Habilidad del catálogo que coincide con lo escrito (aunque no se haya
   * elegido de la lista): en ese caso se reusa su categoría y no se pide.
   */
  get manualDelCatalogo(): Skill | null {
    if (this.manual && typeof this.manual !== "string") return this.manual;
    const key = skillKey(this.manualNombre);
    return this.catalogoFiltrado.find((s) => s.id === key) ?? null;
  }

  get puedeAgregarManual(): boolean {
    if (!this.manualNombre || this.lleno) return false;
    return !!this.manualDelCatalogo || !!this.manualCategoria;
  }

  agregarManual() {
    if (!this.puedeAgregarManual) return;
    if (this.yaAsignada(this.manualNombre)) {
      this.toast.add({
        severity: "info",
        summary: "Ya está en la lista",
        detail: `"${this.manualNombre}" ya es una habilidad de esta actividad.`,
      });
      return;
    }
    const delCatalogo = this.manualDelCatalogo;
    this.agregar(
      delCatalogo
        ? {
            id: delCatalogo.id,
            nombre: delCatalogo.nombre,
            categoria: delCatalogo.categoria,
          }
        : { nombre: this.manualNombre, categoria: this.manualCategoria! },
    );
    this.manual = null;
    this.manualCategoria = null;
  }

  private agregar(skill: SkillInput) {
    if (this.yaAsignada(skill.nombre)) return;
    this.emitir([...this.skills, skill]);
  }

  private emitir(lista: SkillInput[]) {
    this.skills = lista;
    this.skillsChange.emit(lista);
  }

  private yaAsignada(nombre: string): boolean {
    const key = skillKey(nombre);
    return this.skills.some((s) => skillKey(s.nombre) === key);
  }

  private avisarLleno() {
    this.toast.add({
      severity: "warn",
      summary: "Límite de habilidades",
      detail: `Una actividad puede tener como máximo ${this.maxSkills} habilidades.`,
    });
  }
}

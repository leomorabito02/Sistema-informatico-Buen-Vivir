import { Component, Input, OnChanges, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import {
  Comentario,
  ComentariosService,
  EntidadComentable,
} from "../../services/comentarios.service";
import { AuthService } from "../../services/auth.service";
import { ProfileService } from "../../services/profile.service";

interface AutorSesion {
  nombre: string;
  apellido: string;
}

/**
 * Comentarios de un curso o una campaña (HU-24): la lista, del más nuevo
 * al más viejo, y un formulario que pide nombre y apellido (se puede
 * comentar sin cuenta). El texto se muestra con interpolación de
 * Angular, que escapa el HTML; además el backend lo guarda sanitizado.
 * Con `puedeModerar` (administrador) cada comentario tiene "Eliminar".
 */
@Component({
  selector: "app-comentarios",
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="comentarios" [attr.aria-labelledby]="tituloId">
      <h3 class="comentarios-titulo" [id]="tituloId">
        Comentarios ({{ comentarios.length }})
      </h3>

      @if (cargando) {
        <p class="comentarios-mensaje" role="status">Cargando comentarios…</p>
      } @else if (errorLista) {
        <p class="comentarios-mensaje comentarios-error" role="alert">
          {{ errorLista }}
        </p>
      } @else if (comentarios.length === 0) {
        <p class="comentarios-mensaje">
          Todavía no hay comentarios. ¡Dejá el primero!
        </p>
      }

      <ul class="comentarios-lista">
        @for (c of comentarios; track c.id) {
          <li class="comentario">
            <p class="comentario-autor">
              {{ c.nombre }} {{ c.apellido }}
              <span class="comentario-fecha">· {{ fecha(c.fecha) }}</span>
            </p>
            <p class="comentario-texto">{{ c.texto }}</p>
            @if (puedeModerar) {
              <button
                type="button"
                class="comentario-eliminar"
                (click)="eliminar(c)"
              >
                Eliminar
              </button>
            }
          </li>
        }
      </ul>

      <form class="comentarios-form" (ngSubmit)="enviar()" novalidate>
        @if (!sesionUsuario) {
          <div class="comentarios-nombre">
            <label>
              Nombre
              <input
                name="nombre"
                [(ngModel)]="nombre"
                maxlength="50"
                autocomplete="given-name"
                required
              />
            </label>
            <label>
              Apellido
              <input
                name="apellido"
                [(ngModel)]="apellido"
                maxlength="50"
                autocomplete="family-name"
                required
              />
            </label>
          </div>
        }
        <label>
          Comentario
          <textarea
            name="texto"
            [(ngModel)]="texto"
            maxlength="1000"
            rows="3"
            required
          ></textarea>
        </label>
        @if (errorForm) {
          <p class="comentarios-error" role="alert">{{ errorForm }}</p>
        }
        @if (enviado) {
          <p class="comentarios-ok" role="status">Comentario publicado.</p>
        }
        <button type="submit" class="comentarios-enviar" [disabled]="enviando">
          {{ enviando ? "Publicando…" : "Comentar" }}
        </button>
      </form>
    </section>
  `,
  styles: [
    `
      .comentarios {
        margin-top: 1rem;
        display: grid;
        gap: 0.75rem;
      }
      .comentarios-titulo {
        font-size: var(--font-size-base, 1rem);
        font-weight: 600;
        margin: 0;
      }
      .comentarios-mensaje {
        color: var(--muted-foreground, #66655b);
        font-size: var(--font-size-sm, 0.875rem);
        margin: 0;
      }
      .comentarios-error {
        color: var(--destructive, #b91c1c);
        font-size: var(--font-size-sm, 0.875rem);
        margin: 0;
      }
      .comentarios-ok {
        color: var(--primary, #485839);
        font-size: var(--font-size-sm, 0.875rem);
        margin: 0;
      }
      .comentarios-lista {
        list-style: none;
        padding: 0;
        margin: 0;
        display: grid;
        gap: 0.5rem;
      }
      .comentario {
        border: 1px solid var(--border, #e5e5e0);
        border-radius: 8px;
        padding: 0.625rem 0.75rem;
      }
      .comentario-autor {
        font-weight: 600;
        font-size: var(--font-size-sm, 0.875rem);
        margin: 0;
      }
      .comentario-fecha {
        font-weight: 400;
        color: var(--muted-foreground, #66655b);
      }
      .comentario-texto {
        margin: 0.25rem 0 0;
        white-space: pre-line;
        overflow-wrap: anywhere;
      }
      .comentario-eliminar {
        margin-top: 0.375rem;
        background: none;
        border: none;
        padding: 0;
        color: var(--destructive, #b91c1c);
        cursor: pointer;
        font-size: var(--font-size-sm, 0.875rem);
      }
      .comentarios-form {
        display: grid;
        gap: 0.5rem;
      }
      .comentarios-nombre {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 0.5rem;
      }
      .comentarios-form label {
        display: grid;
        gap: 0.25rem;
        font-size: var(--font-size-sm, 0.875rem);
      }
      .comentarios-form input,
      .comentarios-form textarea {
        font: inherit;
        padding: 0.5rem;
        border: 1px solid var(--border, #e5e5e0);
        border-radius: 6px;
        width: 100%;
        box-sizing: border-box;
      }
      .comentarios-form input:focus-visible,
      .comentarios-form textarea:focus-visible {
        outline: 2px solid var(--primary, #485839);
      }
      .comentarios-enviar {
        justify-self: start;
        padding: 0.5rem 1rem;
        border: none;
        border-radius: 6px;
        cursor: pointer;
        background: var(--primary, #485839);
        color: var(--primary-foreground, #fff);
        font: inherit;
      }
      .comentarios-enviar:disabled {
        opacity: 0.6;
        cursor: default;
      }
    `,
  ],
})
export class ComentariosComponent implements OnInit, OnChanges {
  @Input({ required: true }) entidadTipo!: EntidadComentable;
  @Input({ required: true }) entidadId!: string;
  @Input() puedeModerar = false;

  private readonly service = inject(ComentariosService);
  private readonly auth = inject(AuthService, { optional: true });
  private readonly profile = inject(ProfileService, { optional: true });
  private static siguienteId = 0;
  readonly tituloId = `comentarios-titulo-${ComentariosComponent.siguienteId++}`;

  comentarios: Comentario[] = [];
  cargando = false;
  errorLista = "";
  sesionUsuario: AutorSesion | null = null;
  nombre = "";
  apellido = "";
  texto = "";
  enviando = false;
  enviado = false;
  errorForm = "";

  ngOnInit(): void {
    this.cargarSesion();
  }

  ngOnChanges(): void {
    if (this.entidadId) this.cargar();
  }

  private cargarSesion(): void {
    if (!this.auth) return;
    if (
      typeof this.auth.isLocallyAuthenticated === "function" &&
      !this.auth.isLocallyAuthenticated()
    ) {
      return;
    }
    this.auth.isAuthenticated().subscribe({
      next: (me) => {
        const autor = this.normalizarAutor(
          me.user?.first_name,
          me.user?.last_name,
        );
        if (autor) {
          this.sesionUsuario = autor;
          return;
        }
        if (typeof this.profile?.getMyProfile !== "function") return;
        this.profile.getMyProfile().subscribe({
          next: (res) => {
            this.sesionUsuario = this.normalizarAutor(
              res.user?.first_name,
              res.user?.last_name,
            );
          },
          error: () => {
            this.sesionUsuario = null;
          },
        });
      },
      error: () => {
        this.sesionUsuario = null;
      },
    });
  }

  private normalizarAutor(first?: string, last?: string): AutorSesion | null {
    const nombre = (first ?? "").trim();
    const apellido = (last ?? "").trim();
    if (!nombre && !apellido) return null;
    return {
      nombre: nombre || apellido,
      apellido: apellido || "Buen Vivir",
    };
  }

  cargar(): void {
    this.cargando = true;
    this.errorLista = "";
    this.service.listar(this.entidadTipo, this.entidadId).subscribe({
      next: (lista) => {
        this.comentarios = lista ?? [];
        this.cargando = false;
      },
      error: (e: Error) => {
        this.errorLista = e.message;
        this.cargando = false;
      },
    });
  }

  enviar(): void {
    this.enviado = false;
    const autor = this.sesionUsuario ?? {
      nombre: this.nombre.trim(),
      apellido: this.apellido.trim(),
    };
    if (!autor.nombre || !autor.apellido) {
      this.errorForm = "Completá tu nombre y apellido para comentar";
      return;
    }
    if (!this.texto.trim()) {
      this.errorForm = "El comentario está vacío";
      return;
    }
    this.errorForm = "";
    this.enviando = true;
    this.service
      .crear({
        entidadTipo: this.entidadTipo,
        entidadId: this.entidadId,
        nombre: autor.nombre,
        apellido: autor.apellido,
        texto: this.texto,
      })
      .subscribe({
        next: (c) => {
          this.comentarios = [c, ...this.comentarios];
          this.texto = "";
          this.enviando = false;
          this.enviado = true;
        },
        error: (e: Error) => {
          this.errorForm = e.message;
          this.enviando = false;
        },
      });
  }

  eliminar(c: Comentario): void {
    if (!confirm(`¿Eliminar el comentario de ${c.nombre} ${c.apellido}?`)) {
      return;
    }
    this.service.eliminar(c.id).subscribe({
      next: () =>
        (this.comentarios = this.comentarios.filter((x) => x.id !== c.id)),
      error: (e: Error) => (this.errorLista = e.message),
    });
  }

  fecha(iso: string): string {
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleDateString("es-AR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    });
  }
}

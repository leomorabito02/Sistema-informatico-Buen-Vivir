import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { TableModule } from "primeng/table";
import { TagModule } from "primeng/tag";

export interface FilaPermisos {
  rol: string;
  roleId: string;
  puede: string;
  noPuede: string;
  pantallas: string;
}

/**
 * Matriz de roles y permisos tal como la aplica hoy el backend
 * (docs/ROLES_Y_PERMISOS.md). Decisión del 15/11/2026 sobre HU-06: es de
 * solo lectura, porque los permisos están en el código de cada servicio
 * y no en una tabla editable.
 */
export const MATRIZ_PERMISOS: FilaPermisos[] = [
  {
    rol: "Administrador",
    roleId: "role_admin",
    puede:
      "Todo: usuarios y aprobaciones, cursos, áreas y materiales, moderación de contenido docente y comentarios, fichas e hitos, auditoría, ofertas (aprobar, rechazar, borrar), campañas, dashboard y reportes.",
    noPuede: "-",
    pantallas:
      "Dashboard, Cursos, Asistencia, Participantes, Campañas, Inserción laboral, Auditoría, Roles y permisos",
  },
  {
    rol: "Voluntario",
    roleId: "role_volunteer",
    puede:
      "Alta, edición y baja de participantes (y aprobarlos); inscribir y ver asistencia; fichas, hitos y PDF; campañas; ver ofertas, postular participantes, editar y borrar ofertas.",
    noPuede: "Crear cursos, aprobar ofertas, dashboard, auditoría, roles y permisos.",
    pantallas:
      "Cursos (catálogo, inscripciones), Asistencia, Participantes, Seguimiento, Campañas, Inserción laboral",
  },
  {
    rol: "Participante",
    roleId: "role_participant",
    puede:
      "Sus cursos, actividades y entregas, su progreso, su línea de tiempo, ofertas publicadas, su CV, preferencias de notificación, comentar.",
    noPuede: "Ver datos de otros, fichas, gestionar.",
    pantallas: "Mis cursos, Mi proceso, Ofertas, Mi CV",
  },
  {
    rol: "Empresa (aprobada)",
    roleId: "role_company",
    puede:
      "Publicar ofertas (nacen pendientes), ver, editar y borrar las suyas, ver sus postulantes, renovar una oferta vencida.",
    noPuede: "Aprobar ofertas, ver pendientes de otras empresas, perfiles y fichas.",
    pantallas: "Ofertas laborales, Contacto",
  },
  {
    rol: "Docente interno",
    roleId: "role_teacher",
    puede:
      "Sus cursos publicados, actividades, entregas, asistencia, progreso, materiales de sus cursos, habilidades de sus actividades.",
    noPuede: "Fichas, dashboard, gestión de usuarios, moderación.",
    pantallas: "Mis cursos, Asistencia, Progreso, Seguimiento",
  },
  {
    rol: "Docente externo",
    roleId: "role_teacher (externo)",
    puede:
      "Subir contenido (queda pendiente de moderación), ver su estado, editarlo o borrarlo mientras está pendiente.",
    noPuede: "Todo lo demás.",
    pantallas: "Mis contenidos",
  },
  {
    rol: "Visitante (sin sesión)",
    roleId: "-",
    puede: "Campañas publicadas, biblioteca (dejando sus datos), comentarios, mapa.",
    noPuede: "Todo lo que requiere sesión.",
    pantallas: "Landing, campañas públicas, contacto",
  },
];

@Component({
  selector: "app-permisos",
  standalone: true,
  imports: [CommonModule, TableModule, TagModule],
  template: `
    <div class="page-container">
      <div class="header-section">
        <div>
          <h1 class="page-title">Roles y permisos</h1>
          <p class="page-subtitle">
            Qué puede hacer cada rol en el sistema. Es de solo lectura: los
            permisos los aplica cada servicio del backend en cada pedido. Los
            intentos de acceso no permitidos quedan en Auditoría (módulo
            Usuarios y perfiles).
          </p>
        </div>
      </div>

      <p-table [value]="filas" class="p-datatable-striped" [responsiveLayout]="'stack'">
        <ng-template pTemplate="header">
          <tr>
            <th>Rol</th>
            <th>Puede</th>
            <th>No puede</th>
            <th>Pantallas</th>
          </tr>
        </ng-template>
        <ng-template pTemplate="body" let-f>
          <tr>
            <td>
              <strong>{{ f.rol }}</strong>
              <div><code class="role-id">{{ f.roleId }}</code></div>
            </td>
            <td>{{ f.puede }}</td>
            <td>{{ f.noPuede }}</td>
            <td>{{ f.pantallas }}</td>
          </tr>
        </ng-template>
      </p-table>
    </div>
  `,
  styles: [
    `
      .page-container {
        padding: 1.5rem;
      }
      .page-title {
        margin: 0 0 0.25rem;
      }
      .page-subtitle {
        margin: 0 0 1rem;
        color: var(--text-color-secondary, #64748b);
        max-width: 70ch;
      }
      .role-id {
        font-size: 0.75rem;
        color: var(--text-color-secondary, #64748b);
      }
      @media (max-width: 640px) {
        .page-container {
          padding: 1rem;
        }
      }
    `,
  ],
})
export class PermisosComponent {
  readonly filas = MATRIZ_PERMISOS;
}

import {
  Component,
  DestroyRef,
  HostListener,
  OnInit,
  ViewEncapsulation,
  inject,
} from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { CommonModule } from "@angular/common";
import {
  IsActiveMatchOptions,
  NavigationEnd,
  Router,
  RouterLink,
  RouterLinkActive,
} from "@angular/router";
import { filter } from "rxjs";
import { MenuItem } from "primeng/api";
import { MenuModule } from "primeng/menu";
import {
  LucideAngularModule,
  LucideIconData,
  Bell,
  BookOpen,
  Briefcase,
  CalendarCheck,
  ChevronDown,
  Circle,
  CircleQuestionMark,
  CircleUserRound,
  ClipboardList,
  FileText,
  LayoutDashboard,
  Megaphone,
  Menu,
  MessageCircle,
  Sprout,
  TrendingUp,
  Users,
  X,
} from "lucide-angular";
import { AuthService } from "../../../services/auth.service";
import { ProfileService } from "../../../services/profile.service";
import { ImageService } from "../../../services/image.service";
import { MeResponse } from "../../../core/models/auth.model";

/**
 * Menú del header por rol. Solo navegación: el acceso real a cada ruta lo
 * siguen decidiendo los guards de app.routes.ts.
 *
 * - "Dashboard" solo para el administrador (role_admin / role_dev).
 * - El docente y la empresa no tienen "Inicio": entran directo a su portal
 *   (ver InicioComponent).
 */
const CURSOS_GESTION: MenuItem = {
  label: "Cursos",
  items: [
    { label: "Catálogo de cursos", routerLink: "/dashboard/cursos" },
    { label: "Gestión de cursos", routerLink: "/dashboard/gestion-cursos" },
    { label: "Inscripciones", routerLink: "/dashboard/inscripciones" },
    // HU-23 y HU-25: moderar el material de docentes externos y ver
    // quién descargó de la biblioteca.
    { label: "Contenido docente", routerLink: "/dashboard/contenido-docente" },
  ],
};

const ASISTENCIA: MenuItem = {
  label: "Asistencia",
  items: [
    {
      label: "Asistencia por curso",
      routerLink: "/dashboard/asistencia-cursos",
    },
    {
      label: "Asistencia general",
      routerLink: "/dashboard/asistencia-general",
    },
  ],
};

export const NAV_BY_ROLE: Record<string, MenuItem[]> = {
  role_admin: [
    { label: "Dashboard", routerLink: "/dashboard/indicadores" },
    CURSOS_GESTION,
    ASISTENCIA,
    {
      label: "Participantes",
      items: [
        { label: "Usuarios", routerLink: "/dashboard/usuarios" },
        {
          label: "Promoción humana",
          routerLink: "/dashboard/promocion-humana",
        },
        { label: "Seguimiento", routerLink: "/dashboard/seguimiento" },
      ],
    },
    { label: "Campañas", routerLink: "/dashboard/campanas" },
    { label: "Inserción laboral", routerLink: "/dashboard/ofertas" },
  ],
  role_volunteer: [
    {
      label: "Cursos",
      items: [
        { label: "Catálogo de cursos", routerLink: "/dashboard/cursos" },
        { label: "Inscripciones", routerLink: "/dashboard/inscripciones" },
      ],
    },
    ASISTENCIA,
    {
      label: "Participantes",
      items: [
        {
          label: "Gestión de participantes",
          routerLink: "/dashboard/usuarios",
        },
        {
          label: "Promoción humana",
          routerLink: "/dashboard/promocion-humana",
        },
        { label: "Seguimiento", routerLink: "/dashboard/seguimiento" },
      ],
    },
    // HU-06: el voluntariado crea, edita y publica campañas.
    { label: "Campañas", routerLink: "/dashboard/campanas" },
    { label: "Inserción laboral", routerLink: "/dashboard/ofertas" },
  ],
  role_participant: [
    { label: "Mis cursos", routerLink: "/dashboard/cursos" },
    { label: "Mi proceso", routerLink: "/dashboard/mi-proceso" },
    { label: "Ofertas laborales", routerLink: "/dashboard/ofertas" },
    { label: "Mi CV", routerLink: "/dashboard/cv" },
  ],
  role_teacher: [
    { label: "Mis cursos", routerLink: "/dashboard/mis-cursos" },
    { label: "Asistencia", routerLink: "/dashboard/asistencia" },
    { label: "Progreso", routerLink: "/dashboard/progreso" },
    { label: "Seguimiento", routerLink: "/dashboard/seguimiento" },
  ],
  role_company: [
    { label: "Ofertas laborales", routerLink: "/dashboard/ofertas" },
    { label: "Contacto con el centro", routerLink: "/contacto" },
  ],
};
NAV_BY_ROLE["role_dev"] = NAV_BY_ROLE["role_admin"];

/** El docente externo solo carga y consulta contenido (HU-23). */
export const NAV_DOCENTE_EXTERNO: MenuItem[] = [
  { label: "Mis contenidos", routerLink: "/dashboard/mis-contenidos" },
];

/** Ícono de cada sección de primer nivel de la barra lateral. */
const NAV_ICONS: Record<string, LucideIconData> = {
  Dashboard: LayoutDashboard,
  Cursos: BookOpen,
  "Mis cursos": BookOpen,
  Asistencia: CalendarCheck,
  Participantes: Users,
  Campañas: Megaphone,
  "Inserción laboral": Briefcase,
  "Ofertas laborales": Briefcase,
  "Mi proceso": Sprout,
  "Mi CV": FileText,
  "Mis contenidos": FileText,
  Progreso: TrendingUp,
  Seguimiento: ClipboardList,
  "Contacto con el centro": MessageCircle,
};

/** Coincidencia "la ruta actual está dentro de este link". */
const SUBSET: IsActiveMatchOptions = {
  paths: "subset",
  queryParams: "ignored",
  fragment: "ignored",
  matrixParams: "ignored",
};

const ROLE_LABELS: Record<string, string> = {
  role_admin: "Responsable",
  role_dev: "Desarrollo",
  role_volunteer: "Voluntariado",
  role_participant: "Participante",
  role_teacher: "Docente",
  role_company: "Empresa",
};

@Component({
  selector: "app-header",
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    RouterLinkActive,
    MenuModule,
    LucideAngularModule,
  ],
  templateUrl: "./header.component.html",
  styleUrl: "./header.component.css",
  // Sin encapsulación: los estilos del menú de usuario (p-menu en <body>)
  // también viven acá; el resto de los selectores va bajo .app-header.
  encapsulation: ViewEncapsulation.None,
})
export class HeaderComponent implements OnInit {
  private readonly authService = inject(AuthService);
  private readonly profileService = inject(ProfileService, { optional: true });
  private readonly imageService = inject(ImageService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  roleId = "";
  roleName = "";
  roleLabel = "";
  firstName = "";
  initials = "";
  avatarUrl: string | null = null;
  items: MenuItem[] = [];

  /** Barra lateral abierta (solo cuenta en celular/tablet). */
  sidebarOpen = false;
  /** Grupos desplegados, por etiqueta. */
  readonly expanded = new Set<string>();

  readonly userItems: MenuItem[] = [
    { label: "Mi perfil", routerLink: "/dashboard/mi-perfil" },
    { label: "Ayuda", routerLink: "/dashboard/ayuda" },
    { separator: true },
    { label: "Cerrar sesión", command: () => this.logout() },
  ];

  // Icons
  readonly Bell = Bell;
  readonly ChevronDown = ChevronDown;
  readonly CircleUserRound = CircleUserRound;
  readonly CircleQuestionMark = CircleQuestionMark;
  readonly Menu = Menu;
  readonly X = X;

  ngOnInit(): void {
    // Al cambiar de pantalla: se cierra el panel en celular y se despliega
    // el grupo de la sección actual.
    this.router.events
      .pipe(
        filter((e) => e instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.sidebarOpen = false;
        this.expandActiveGroup();
      });

    this.authService.isAuthenticated().subscribe({
      next: (response: MeResponse) => {
        this.roleId = response.role_id ?? response.user?.role_id ?? "";
        this.roleName = this.mapRoleIdToName(this.roleId);
        this.roleLabel = ROLE_LABELS[this.roleId.toLowerCase()] ?? "";
        this.firstName = response.user?.first_name?.trim() ?? "";
        this.initials = this.initialsOf(
          this.firstName,
          response.user?.last_name?.trim() ?? "",
        );
        const userPic =
          response.user?.image_url ||
          (response.user?.role_data?.['image_url'] as string | undefined) ||
          (response.user?.role_data?.['profile_picture_url'] as string | undefined);
        if (userPic) {
          this.avatarUrl = this.resolveImageUrl(userPic);
        }
        this.items =
          this.roleId === "role_teacher" && response.is_external
            ? NAV_DOCENTE_EXTERNO
            : (NAV_BY_ROLE[this.roleId.toLowerCase()] ?? []);
        this.expandActiveGroup();
        this.loadProfileAvatar();
      },
      error: () => {
        this.roleId = "";
        this.roleName = "";
        this.roleLabel = "";
        this.items = [];
        this.avatarUrl = null;
      },
    });
  }

  mapRoleIdToName(roleId: string): string {
    if (!roleId) return "";
    const id = roleId.toLowerCase();
    if (id === "role_participant" || id === "participante")
      return "participante";
    if (id === "role_admin" || id === "responsable") return "responsable";
    if (id === "role_volunteer" || id === "voluntario") return "voluntario";
    if (id === "role_teacher" || id === "docente") return "docente";
    if (id === "role_company" || id === "empresa") return "empresa";
    return id.replace("role_", "");
  }

  iconFor(item: MenuItem): LucideIconData {
    return NAV_ICONS[item.label ?? ""] ?? Circle;
  }

  /** ¿La pantalla actual pertenece a alguno de los hijos del grupo? */
  isGroupActive(item: MenuItem): boolean {
    return (item.items ?? []).some(
      (child) =>
        typeof child.routerLink === "string" &&
        this.router.isActive(child.routerLink, SUBSET),
    );
  }

  isExpanded(item: MenuItem): boolean {
    return this.expanded.has(item.label ?? "");
  }

  toggleGroup(item: MenuItem): void {
    const key = item.label ?? "";
    if (this.expanded.has(key)) this.expanded.delete(key);
    else this.expanded.add(key);
  }

  toggleSidebar(): void {
    this.sidebarOpen = !this.sidebarOpen;
  }

  @HostListener("document:keydown.escape")
  closeSidebar(): void {
    this.sidebarOpen = false;
  }

  /**
   * Atajos para quien usa teclado (heurística 7): "/" lleva al buscador de
   * la pantalla y "?" abre la Ayuda. No actúan mientras se escribe.
   */
  @HostListener("document:keydown", ["$event"])
  onShortcut(event: KeyboardEvent): void {
    if (event.ctrlKey || event.metaKey || event.altKey) return;
    const target = event.target as HTMLElement | null;
    if (
      target &&
      (target.isContentEditable ||
        ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName))
    ) {
      return;
    }
    if (event.key === "/") {
      const search = document.querySelector<HTMLInputElement>(
        'main input[type="search"], main input[placeholder^="Buscar"]',
      );
      if (search) {
        event.preventDefault();
        search.focus();
      }
    } else if (event.key === "?") {
      event.preventDefault();
      void this.router.navigateByUrl("/dashboard/ayuda");
    }
  }

  private expandActiveGroup(): void {
    for (const item of this.items) {
      if (item.items && this.isGroupActive(item)) {
        this.expanded.add(item.label ?? "");
      }
    }
  }

  private initialsOf(first: string, last: string): string {
    return ((first[0] ?? "") + (last[0] ?? "")).toUpperCase();
  }

  resolveImageUrl(url?: string | boolean | null): string {
    return this.imageService.resolveImageUrl(url);
  }

  onAvatarError(): void {
    this.avatarUrl = null;
  }

  private loadProfileAvatar(): void {
    if (!this.profileService) return;
    this.profileService
      .getMyProfile()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (profile) => {
          const pic =
            profile.user?.image_url ||
            (profile.role_data?.['image_url'] as string | undefined) ||
            (profile.role_data?.['profile_picture_url'] as string | undefined);
          if (pic) {
            this.avatarUrl = this.resolveImageUrl(pic);
          }
          if (!this.firstName && profile.user?.first_name) {
            this.firstName = profile.user.first_name.trim();
            this.initials = this.initialsOf(
              this.firstName,
              profile.user.last_name?.trim() ?? "",
            );
          }
        },
        error: () => {
          // Fallback to initials / placeholder
        },
      });
  }

  logout(): void {
    this.authService.logout();
  }
}

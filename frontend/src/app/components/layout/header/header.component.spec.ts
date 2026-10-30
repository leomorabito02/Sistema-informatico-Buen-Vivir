import { ComponentFixture, TestBed } from "@angular/core/testing";
import { HeaderComponent, NAV_BY_ROLE } from "./header.component";
import { AuthService } from "../../../services/auth.service";
import { ProfileService } from "../../../services/profile.service";
import { Router, provideRouter } from "@angular/router";
import { provideNoopAnimations } from "@angular/platform-browser/animations";
import { MenuItem } from "primeng/api";
import { of, throwError } from "rxjs";
import { Bell, ChevronDown, Circle, CircleUserRound } from "lucide-angular";

/** Todas las rutas (con hijos aplanados) de un menú. */
function routesOf(items: MenuItem[]): string[] {
  return items.flatMap((it) => [
    ...(it.routerLink ? [it.routerLink as string] : []),
    ...routesOf(it.items ?? []),
  ]);
}

describe("HeaderComponent", () => {
  let component: HeaderComponent;
  let fixture: ComponentFixture<HeaderComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockProfileService: jasmine.SpyObj<ProfileService>;

  function create(me: unknown) {
    mockAuthService.isAuthenticated.and.returnValue(of(me as any));
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "logout",
    ]);
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );

    mockProfileService = jasmine.createSpyObj("ProfileService", [
      "getMyProfile",
    ]);
    mockProfileService.getMyProfile.and.returnValue(of({} as any));

    await TestBed.configureTestingModule({
      imports: [HeaderComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: ProfileService, useValue: mockProfileService },
        provideRouter([{ path: "**", children: [] }]),
        provideNoopAnimations(),
      ],
    }).compileComponents();
  });

  it("should create and set roleName, label and menu on init", () => {
    create({ role_id: "role_participant", user: { first_name: " Ana " } });

    expect(component).toBeTruthy();
    expect(component.roleName).toBe("participante");
    expect(component.roleLabel).toBe("Participante");
    expect(component.firstName).toBe("Ana");
    expect(component.items).toBe(NAV_BY_ROLE["role_participant"]);
    expect(component.Bell).toBe(Bell);
    expect(component.ChevronDown).toBe(ChevronDown);
    expect(component.CircleUserRound).toBe(CircleUserRound);
  });

  it("should read role_id from user object if root role_id is missing", () => {
    create({ user: { role_id: "role_admin" } });

    expect(component.roleName).toBe("responsable");
    expect(component.items).toBe(NAV_BY_ROLE["role_admin"]);
  });

  it("should fallback to empty role and menu if isAuthenticated fails", () => {
    mockAuthService.isAuthenticated.and.returnValue(
      throwError(() => new Error("Unauthorized")),
    );
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleName).toBe("");
    expect(component.roleLabel).toBe("");
    expect(component.items).toEqual([]);
  });

  it("should show an empty menu for an unknown role", () => {
    create({ role_id: "role_custom" });
    expect(component.items).toEqual([]);
    expect(component.roleLabel).toBe("");
  });

  it("should map roleId to name correctly for all roles", () => {
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;

    expect(component.mapRoleIdToName("")).toBe("");
    expect(component.mapRoleIdToName("role_participant")).toBe("participante");
    expect(component.mapRoleIdToName("participante")).toBe("participante");
    expect(component.mapRoleIdToName("role_admin")).toBe("responsable");
    expect(component.mapRoleIdToName("responsable")).toBe("responsable");
    expect(component.mapRoleIdToName("role_volunteer")).toBe("voluntario");
    expect(component.mapRoleIdToName("voluntario")).toBe("voluntario");
    expect(component.mapRoleIdToName("role_teacher")).toBe("docente");
    expect(component.mapRoleIdToName("docente")).toBe("docente");
    expect(component.mapRoleIdToName("role_company")).toBe("empresa");
    expect(component.mapRoleIdToName("empresa")).toBe("empresa");
    expect(component.mapRoleIdToName("role_custom")).toBe("custom");
  });

  it("should call authService.logout on logout()", () => {
    create({ role_id: "role_participant" });
    component.logout();
    expect(mockAuthService.logout).toHaveBeenCalled();
  });

  it("should log out from the user menu item", () => {
    create({ role_id: "role_participant" });
    const salir = component.userItems.find((i) => i.label === "Cerrar sesión");
    salir?.command?.({});
    expect(mockAuthService.logout).toHaveBeenCalled();
  });

  describe("menú por rol", () => {
    it("solo el administrador tiene la opción Dashboard", () => {
      for (const [role, items] of Object.entries(NAV_BY_ROLE)) {
        const tieneDashboard = items.some((i) => i.label === "Dashboard");
        expect(tieneDashboard)
          .withContext(role)
          .toBe(role === "role_admin" || role === "role_dev");
      }
    });

    it("la empresa solo ve opciones laborales y de contacto, no cursos", () => {
      const rutas = routesOf(NAV_BY_ROLE["role_company"]);
      expect(rutas).toEqual(["/dashboard/ofertas", "/contacto"]);
    });

    it("el docente no tiene ofertas laborales ni campañas", () => {
      const rutas = routesOf(NAV_BY_ROLE["role_teacher"]);
      expect(rutas).not.toContain("/dashboard/ofertas");
      expect(rutas).not.toContain("/dashboard/campanas");
      expect(rutas[0]).toBe("/dashboard/mis-cursos");
    });

    it("el docente externo solo ve Mis contenidos (HU-23)", () => {
      create({ role_id: "role_teacher", is_external: true });
      expect(routesOf(component.items)).toEqual(["/dashboard/mis-contenidos"]);
    });

    it("el docente interno sigue viendo sus cursos", () => {
      create({ role_id: "role_teacher", is_external: false });
      expect(routesOf(component.items)[0]).toBe("/dashboard/mis-cursos");
    });

    it("el administrador modera el contenido docente", () => {
      expect(routesOf(NAV_BY_ROLE["role_admin"])).toContain(
        "/dashboard/contenido-docente",
      );
    });

    it("el participante no recibe rutas exclusivas del docente", () => {
      const rutas = routesOf(NAV_BY_ROLE["role_participant"]);
      expect(rutas).not.toContain("/dashboard/progreso");
      expect(rutas).toContain("/dashboard/mi-proceso");
    });

    it("el voluntario no tiene gestión de cursos ni usuarios de admin", () => {
      const rutas = routesOf(NAV_BY_ROLE["role_volunteer"]);
      expect(rutas).not.toContain("/dashboard/gestion-cursos");
      expect(rutas).toContain("/dashboard/inscripciones");
    });

    it("el voluntario gestiona campañas (HU-06)", () => {
      expect(routesOf(NAV_BY_ROLE["role_volunteer"])).toContain(
        "/dashboard/campanas",
      );
    });

    it("ningún ítem del menú dice Dashboard salvo el del admin", () => {
      const etiquetas = Object.entries(NAV_BY_ROLE)
        .filter(([r]) => r !== "role_admin" && r !== "role_dev")
        .flatMap(([, items]) => items.map((i) => i.label));
      expect(etiquetas).not.toContain("Dashboard");
    });
  });

  describe("barra lateral", () => {
    it("muestra las iniciales del usuario en el avatar", () => {
      create({
        role_id: "role_admin",
        user: { first_name: "sofía", last_name: "Cuozzo" },
      });
      expect(component.initials).toBe("SC");
      const el: HTMLElement = fixture.nativeElement;
      expect(el.querySelector(".avatar")?.textContent).toContain("SC");
    });

    it("cada sección del menú de todos los roles tiene su propio ícono", () => {
      create({ role_id: "role_participant" });
      for (const items of Object.values(NAV_BY_ROLE)) {
        for (const item of items) {
          expect(component.iconFor(item))
            .withContext(item.label ?? "")
            .not.toBe(Circle);
        }
      }
      expect(component.iconFor({ label: "Otra" })).toBe(Circle);
    });

    it("despliega y pliega un grupo", () => {
      create({ role_id: "role_admin" });
      const cursos = component.items.find((i) => i.label === "Cursos")!;
      expect(component.isExpanded(cursos)).toBeFalse();

      component.toggleGroup(cursos);
      fixture.detectChanges();
      expect(component.isExpanded(cursos)).toBeTrue();
      // Catálogo, gestión, inscripciones y contenido docente (HU-23).
      expect(fixture.nativeElement.querySelectorAll(".nav-subitem")).toHaveSize(
        4,
      );

      component.toggleGroup(cursos);
      expect(component.isExpanded(cursos)).toBeFalse();
    });

    it("abre el grupo de la pantalla actual al navegar", async () => {
      create({ role_id: "role_admin" });
      await TestBed.inject(Router).navigateByUrl("/dashboard/usuarios");
      const participantes = component.items.find(
        (i) => i.label === "Participantes",
      )!;
      expect(component.isGroupActive(participantes)).toBeTrue();
      expect(component.isExpanded(participantes)).toBeTrue();
    });

    it("abre y cierra el panel en celular, y lo cierra al navegar", async () => {
      create({ role_id: "role_participant" });
      component.toggleSidebar();
      fixture.detectChanges();
      expect(component.sidebarOpen).toBeTrue();
      expect(
        fixture.nativeElement.querySelector(".sidebar-backdrop"),
      ).not.toBeNull();

      component.closeSidebar();
      expect(component.sidebarOpen).toBeFalse();

      component.toggleSidebar();
      await TestBed.inject(Router).navigateByUrl("/dashboard/cv");
      expect(component.sidebarOpen).toBeFalse();
    });

    it("atajos: / enfoca el buscador y ? abre la Ayuda", () => {
      create({ role_id: "role_participant" });
      const router = TestBed.inject(Router);
      const nav = spyOn(router, "navigateByUrl").and.resolveTo(true);

      const main = document.createElement("main");
      const input = document.createElement("input");
      input.type = "search";
      main.appendChild(input);
      document.body.appendChild(main);
      try {
        document.dispatchEvent(new KeyboardEvent("keydown", { key: "/" }));
        expect(document.activeElement).toBe(input);

        // Mientras se escribe en un campo, los atajos no actúan.
        input.dispatchEvent(
          new KeyboardEvent("keydown", { key: "?", bubbles: true }),
        );
        expect(nav).not.toHaveBeenCalled();

        input.blur();
        document.dispatchEvent(new KeyboardEvent("keydown", { key: "?" }));
        expect(nav).toHaveBeenCalledWith("/dashboard/ayuda");

        nav.calls.reset();
        document.dispatchEvent(
          new KeyboardEvent("keydown", { key: "?", ctrlKey: true }),
        );
        expect(nav).not.toHaveBeenCalled();
      } finally {
        main.remove();
      }
    });

    it("cierra el panel con Escape", () => {
      create({ role_id: "role_participant" });
      component.sidebarOpen = true;
      document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
      expect(component.sidebarOpen).toBeFalse();
    });
  });

  it("renders the brand, skip link and user button", () => {
    create({ role_id: "role_admin", user: { first_name: "Sofía" } });
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector(".skip-link")?.getAttribute("href")).toBe(
      "#contenido-principal",
    );
    expect(el.querySelector(".brand-name")?.textContent).toContain(
      "Buen Vivir",
    );
    expect(el.querySelector(".user-name")?.textContent).toContain("Sofía");
    expect(el.querySelector('[aria-label="Notificaciones"]')).not.toBeNull();
  });

  describe("Avatar display and fallback", () => {
    it("renders avatar image when user has image_url", () => {
      create({
        role_id: "role_participant",
        user: {
          first_name: "Ana",
          image_url: "https://example.com/avatar.jpg",
        },
      });

      expect(component.avatarUrl).toBe("https://example.com/avatar.jpg");
      const img = fixture.nativeElement.querySelector(".avatar-img");
      expect(img).not.toBeNull();
      expect(img?.getAttribute("src")).toBe("https://example.com/avatar.jpg");
    });

    it("falls back to initials when no image is available", () => {
      create({
        role_id: "role_participant",
        user: { first_name: "Ana", last_name: "Gómez" },
      });

      expect(component.avatarUrl).toBeNull();
      const img = fixture.nativeElement.querySelector(".avatar-img");
      expect(img).toBeNull();
      const avatarText =
        fixture.nativeElement.querySelector(".avatar")?.textContent;
      expect(avatarText).toContain("AG");
    });

    it("falls back to initials when avatar image fails to load", () => {
      create({
        role_id: "role_participant",
        user: {
          first_name: "Ana",
          last_name: "Gómez",
          image_url: "https://example.com/broken.jpg",
        },
      });

      expect(component.avatarUrl).toBe("https://example.com/broken.jpg");
      component.onAvatarError();
      fixture.detectChanges();

      expect(component.avatarUrl).toBeNull();
      const img = fixture.nativeElement.querySelector(".avatar-img");
      expect(img).toBeNull();
      const avatarText =
        fixture.nativeElement.querySelector(".avatar")?.textContent;
      expect(avatarText).toContain("AG");
    });

    it("resolves relative and localhost api urls correctly", () => {
      expect(component.resolveImageUrl(null)).toBe("");
      expect(component.resolveImageUrl("")).toBe("");
      expect(
        component.resolveImageUrl("https://cdn.example.com/photo.png"),
      ).toBe("https://cdn.example.com/photo.png");
      expect(
        component.resolveImageUrl(
          "http://localhost:8080/api/v1/uploads/pic.png",
        ),
      ).toContain("/api/v1/uploads/pic.png");
      expect(component.resolveImageUrl("/uploads/pic.png")).toContain(
        "/uploads/pic.png",
      );
    });
  });
});

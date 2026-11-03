import { routes } from "./app.routes";
import { LandingComponent } from "./pages/landing/landing.component";
import { LoginComponent } from "./pages/login/login.component";
import { RegisterComponent } from "./pages/register/register.component";
import { DashboardComponent } from "./pages/dashboard/dashboard.component";
import { authGuard } from "./core/guards/auth.guard";
import { guestGuard } from "./core/guards/guest.guard";

describe("App Routes", () => {
  it("should contain the correct routes configuration", () => {
    expect(routes).toBeDefined();
    expect(routes.length).toBeGreaterThan(0);

    const landingRoute = routes.find((r) => r.path === "");
    expect(
      landingRoute?.loadComponent || landingRoute?.component,
    ).toBeDefined();

    const loginRoute = routes.find((r) => r.path === "login");
    expect(loginRoute?.loadComponent || loginRoute?.component).toBeDefined();
    expect(loginRoute?.canActivate).toEqual([guestGuard]);

    const registerRoute = routes.find((r) => r.path === "register");
    expect(
      registerRoute?.loadComponent || registerRoute?.component,
    ).toBeDefined();
    expect(registerRoute?.canActivate).toEqual([guestGuard]);

    const dashboardRoute = routes.find((r) => r.path === "dashboard");
    expect(
      dashboardRoute?.loadComponent || dashboardRoute?.component,
    ).toBeDefined();
    expect(dashboardRoute?.canActivate).toEqual([authGuard]);
    expect(dashboardRoute?.children).toHaveSize(25);
  });

  it("should lazy load pending-approval component", async () => {
    const route = routes.find((r) => r.path === "pending-approval");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load account-notice component", async () => {
    const route = routes.find((r) => r.path === "account-notice");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load verify-email component", async () => {
    const route = routes.find((r) => r.path === "verify-email");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load change-password component", async () => {
    const route = routes.find((r) => r.path === "change-password");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load select-role component", async () => {
    const route = routes.find((r) => r.path === "select-role");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load privacidad component", async () => {
    const route = routes.find((r) => r.path === "privacidad");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load terminos component", async () => {
    const route = routes.find((r) => r.path === "terminos");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load cookies component", async () => {
    const route = routes.find((r) => r.path === "cookies");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load contacto component", async () => {
    const route = routes.find((r) => r.path === "contacto");
    expect(route?.loadComponent).toBeDefined();
    if (route?.loadComponent) {
      const component = await (route.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("la página pública de campañas no pide sesión (HU-50)", async () => {
    const route = routes.find((r) => r.path === "campanas-publico");
    expect(route?.canActivate).toBeUndefined();
    const component = await (route!.loadComponent as any)();
    expect(component).toBeDefined();
  });

  it("la gestión de campañas es solo para administración y voluntariado", () => {
    const dashboard = routes.find((r) => r.path === "dashboard");
    const campanas = dashboard?.children?.find((r) => r.path === "campanas");
    expect(campanas?.canActivate?.length).toBe(1);
  });

  it("should redirect mi-cuenta to dashboard/mi-perfil", () => {
    const miCuenta = routes.find((r) => r.path === "mi-cuenta");
    expect(miCuenta?.redirectTo).toBe("dashboard/mi-perfil");
  });

  const childrenRoutes = [
    "mi-perfil",
    "cursos",
    "gestion-cursos",
    "mis-cursos",
    "progreso",
    "ofertas",
    "cv",
    "usuarios",
    "seguimiento",
    "asistencia",
    "campanas",
    "cursos/:id",
    "inscripciones",
    "asistencia-cursos",
    "asistencia-general",
    "notificaciones",
    "ayuda",
    "indicadores",
    "auditoria",
  ];

  childrenRoutes.forEach((childPath) => {
    it(`should lazy load dashboard child route: ${childPath}`, async () => {
      const dashboard = routes.find((r) => r.path === "dashboard");
      const child = dashboard?.children?.find((r) => r.path === childPath);
      expect(child?.loadComponent).toBeDefined();
      if (child?.loadComponent) {
        const component = await (child.loadComponent as any)();
        expect(component).toBeDefined();
      }
    });
  });

  it("should lazy load 404 and wildcard components", async () => {
    const route404 = routes.find((r) => r.path === "404");
    expect(route404?.loadComponent).toBeDefined();
    if (route404?.loadComponent) {
      const component = await (route404.loadComponent as any)();
      expect(component).toBeDefined();
    }

    const wildcardRoute = routes.find((r) => r.path === "**");
    expect(wildcardRoute?.loadComponent).toBeDefined();
    if (wildcardRoute?.loadComponent) {
      const component = await (wildcardRoute.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });

  it("should lazy load 500 component", async () => {
    const route500 = routes.find((r) => r.path === "500");
    expect(route500?.loadComponent).toBeDefined();
    if (route500?.loadComponent) {
      const component = await (route500.loadComponent as any)();
      expect(component).toBeDefined();
    }
  });
});

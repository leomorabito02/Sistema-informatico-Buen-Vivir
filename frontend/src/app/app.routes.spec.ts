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
    expect(landingRoute?.component).toBe(LandingComponent);

    const loginRoute = routes.find((r) => r.path === "login");
    expect(loginRoute?.component).toBe(LoginComponent);
    expect(loginRoute?.canActivate).toEqual([guestGuard]);

    const registerRoute = routes.find((r) => r.path === "register");
    expect(registerRoute?.component).toBe(RegisterComponent);
    expect(registerRoute?.canActivate).toEqual([guestGuard]);

    const dashboardRoute = routes.find((r) => r.path === "dashboard");
    expect(dashboardRoute?.component).toBe(DashboardComponent);
    expect(dashboardRoute?.canActivate).toEqual([authGuard]);
    expect(dashboardRoute?.children).toHaveSize(14);
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

  it("should lazy load mi-cuenta component and its children", async () => {
    const miCuenta = routes.find((r) => r.path === "mi-cuenta");
    expect(miCuenta?.loadComponent).toBeDefined();
    expect(miCuenta?.children?.length).toBeGreaterThan(0);
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
    "insercion",
    "participantes",
    "notificaciones",
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
});

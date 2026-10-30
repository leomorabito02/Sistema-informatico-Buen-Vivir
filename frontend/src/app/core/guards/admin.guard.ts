import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { catchError, map, of } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

/**
 * Deja pasar solo a `role_admin` / `role_dev`. Cualquier otro rol se
 * redirige al inicio con `?acceso=denegado`, y el panel muestra el aviso
 * de acceso denegado (403). Una sesión inválida va al login.
 * Se apoya en `GET /api/v1/auth/me`, igual que `authGuard`.
 */
export const adminGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      const roleId = res.role_id ?? res.user?.role_id ?? "";
      if (roleId === "role_admin" || roleId === "role_dev") return true;
      return router.createUrlTree(["/dashboard"], {
        queryParams: { acceso: "denegado" },
      });
    }),
    catchError(() => of(router.createUrlTree(["/login"]))),
  );
};

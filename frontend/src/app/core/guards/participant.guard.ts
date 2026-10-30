import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { catchError, map, of } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

/**
 * Deja pasar solo a `role_participant`. Calcado de `adminOrVolunteerGuard`,
 * para la línea de tiempo propia del participante (HU-34). El backend
 * también lo restringe: el guard solo evita mostrar una pantalla vacía.
 */
export const participantGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      const roleId = res.role_id ?? res.user?.role_id ?? "";
      if (roleId === "role_participant") return true;
      return router.createUrlTree(["/dashboard"]);
    }),
    catchError(() => of(router.createUrlTree(["/login"]))),
  );
};

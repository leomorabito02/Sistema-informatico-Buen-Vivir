import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { catchError, map, of } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

/**
 * Deja pasar solo a `role_admin` / `role_volunteer` / `role_dev`. Calcado
 * de `adminGuard`, para las pantallas nuevas de inscripción y asistencia
 * (HU-18, HU-19, HU-19bis) que el backend expone a admin y voluntario sin
 * scope de "mis cursos" (ver education/CONTEXTO_INSCRIPCIONES_ASISTENCIA.md).
 */
export const adminOrVolunteerGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      const roleId = res.role_id ?? res.user?.role_id ?? "";
      if (
        roleId === "role_admin" ||
        roleId === "role_volunteer" ||
        roleId === "role_dev"
      )
        return true;
      return router.createUrlTree(["/dashboard"]);
    }),
    catchError(() => of(router.createUrlTree(["/login"]))),
  );
};

import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { catchError, map, of } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

/**
 * Deja pasar solo a `role_teacher` (docente responsable). Evita que un
 * admin/voluntario llegue por URL directa a las pantallas de "mis
 * cursos" (scope docente) y se confunda con las pantallas equivalentes
 * de admin/voluntario (`adminOrVolunteerGuard`: asistencia-cursos,
 * asistencia-general, inscripciones).
 */
export const teacherGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      const roleId = res.role_id ?? res.user?.role_id ?? "";
      // El docente externo no tiene cursos a cargo: solo carga y consulta
      // contenido (HU-23).
      if (roleId === "role_teacher" && res.is_external) {
        return router.createUrlTree(["/dashboard/mis-contenidos"]);
      }
      if (roleId === "role_teacher" || roleId === "role_dev") return true;
      return router.createUrlTree(["/dashboard"]);
    }),
    catchError(() => of(router.createUrlTree(["/login"]))),
  );
};

/**
 * "Mis contenidos" (HU-23): cualquier docente, interno o externo. Es la
 * única pantalla del docente externo.
 */
export const anyTeacherGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      const roleId = res.role_id ?? res.user?.role_id ?? "";
      if (roleId === "role_teacher" || roleId === "role_dev") return true;
      return router.createUrlTree(["/dashboard"]);
    }),
    catchError(() => of(router.createUrlTree(["/login"]))),
  );
};

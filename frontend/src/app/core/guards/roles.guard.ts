import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { catchError, map, of } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

/**
 * Guard genérico por rol (HU-06): deja pasar solo a los roles indicados
 * (role_dev pasa siempre, como admin en el backend). Si no corresponde,
 * vuelve al inicio del dashboard. Es solo navegación: el backend valida
 * igual cada pedido.
 */
export function rolesGuard(...roles: string[]): CanActivateFn {
  return () => {
    const authService = inject(AuthService);
    const router = inject(Router);

    if (!authService.isLocallyAuthenticated()) {
      return router.createUrlTree(["/login"]);
    }

    return authService.isAuthenticated().pipe(
      map((res: MeResponse) => {
        const roleId = res.role_id ?? res.user?.role_id ?? "";
        if (roleId === "role_dev" || roles.includes(roleId)) return true;
        return router.createUrlTree(["/dashboard"]);
      }),
      catchError(() => of(router.createUrlTree(["/login"]))),
    );
  };
}

import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";
import { catchError, map, of, timeout } from "rxjs";

export const guestGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return true;
  }

  return authService.isAuthenticated().pipe(
    timeout(1500),
    map((res: MeResponse) => {
      if (res.approval_status === "pending") {
        return router.createUrlTree(["/pending-approval"]);
      }
      return router.createUrlTree(["/dashboard"]);
    }),
    catchError(() => of(true)),
  );
};

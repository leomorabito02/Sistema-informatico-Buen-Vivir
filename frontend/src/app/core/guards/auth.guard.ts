import { CanActivateFn, Router } from "@angular/router";
import { inject } from "@angular/core";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";
import { catchError, map, of } from "rxjs";

export const authGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLocallyAuthenticated()) {
    return router.createUrlTree(["/login"]);
  }

  return authService.isAuthenticated().pipe(
    map((res: MeResponse) => {
      if (res.user_status === "blocked" || res.user?.status === "blocked") {
        return router.createUrlTree(["/account-notice"]);
      }
      if (res.user?.is_email_verified === false) {
        return router.createUrlTree(["/verify-email"], {
          queryParams: res.user?.email ? { email: res.user.email } : undefined,
        });
      }
      if (res.approval_status === "pending") {
        return router.createUrlTree(["/pending-approval"]);
      }
      return true;
    }),
    catchError(() => {
      return of(router.createUrlTree(["/login"]));
    }),
  );
};

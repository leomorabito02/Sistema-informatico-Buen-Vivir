import { CanActivateFn, Router } from '@angular/router';
import { inject } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { catchError, map, of } from 'rxjs';

export const guestGuard: CanActivateFn = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.isAuthenticated().pipe(
    map((res: any) => {
      if (res.approval_status === 'pending') {
        return router.createUrlTree(['/pending-approval']);
      }
      return router.createUrlTree(['/dashboard']);
    }),
    catchError(() => of(true))
  );
};


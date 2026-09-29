import { HttpInterceptorFn, HttpErrorResponse } from "@angular/common/http";

import { inject } from "@angular/core";
import { catchError, switchMap } from "rxjs/operators";
import { throwError, BehaviorSubject, filter, take } from "rxjs";
import { AuthService } from "../../services/auth.service";

let isRefreshing = false;
const refreshTokenSubject = new BehaviorSubject<any>(null);

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const isExternal =
    req.url.startsWith("http") && !req.url.startsWith(window.location.origin);
  if (isExternal) {
    return next(req);
  }

  const clonedReq = req.clone({
    withCredentials: true,
  });

  const authService = inject(AuthService);

  return next(clonedReq).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !req.url.includes("/api/v1/auth/")) {
        if (!isRefreshing) {
          isRefreshing = true;
          refreshTokenSubject.next(null);

          const extend = confirm("Tu sesión expiró. ¿Deseas extenderla?");
          if (extend) {
            return authService.refreshToken().pipe(
              switchMap((res) => {
                isRefreshing = false;
                refreshTokenSubject.next(res);
                return next(clonedReq);
              }),
              catchError((err) => {
                isRefreshing = false;
                authService.logout();
                return throwError(() => err);
              }),
            );
          } else {
            isRefreshing = false;
            authService.logout();
            return throwError(() => error);
          }
        } else {
          return refreshTokenSubject.pipe(
            filter((token) => token != null),
            take(1),
            switchMap(() => {
              return next(clonedReq);
            }),
          );
        }
      }
      return throwError(() => error);
    }),
  );
};

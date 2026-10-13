import { HttpInterceptorFn, HttpErrorResponse } from "@angular/common/http";

import { inject } from "@angular/core";
import { catchError, switchMap } from "rxjs/operators";
import { throwError, BehaviorSubject, filter, take } from "rxjs";
import { AuthService } from "../../services/auth.service";
import { SessionPromptService } from "../../services/session-prompt.service";
import { environment } from "../../../environments/environment";

let isRefreshing = false;
const refreshTokenSubject = new BehaviorSubject<any>(null);

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const isApiUrl = !!environment.apiUrl && req.url.startsWith(environment.apiUrl);
  const isInternal =
    req.url.startsWith("/") ||
    req.url.startsWith(window.location.origin) ||
    isApiUrl;
  if (!isInternal) {
    return next(req);
  }

  const clonedReq = req.clone({
    withCredentials: true,
  });

  const authService = inject(AuthService);
  const sessionPrompt = inject(SessionPromptService);

  return next(clonedReq).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !req.url.includes("/api/v1/auth/")) {
        if (!isRefreshing) {
          isRefreshing = true;
          refreshTokenSubject.next(null);

          // Diálogo propio de la app (no el confirm() del navegador).
          return sessionPrompt.ask().pipe(
            switchMap((extend) => {
              if (!extend) {
                isRefreshing = false;
                authService.logout();
                return throwError(() => error);
              }
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
            }),
          );
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

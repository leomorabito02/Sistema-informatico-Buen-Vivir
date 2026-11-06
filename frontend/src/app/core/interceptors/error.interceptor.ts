import {
  HttpInterceptorFn,
  HttpErrorResponse,
  HttpHeaders,
} from "@angular/common/http";
import { inject } from "@angular/core";
import { Router } from "@angular/router";
import { catchError, throwError } from "rxjs";
import { AlertService } from "../../services/alert.service";

export const SKIP_GLOBAL_ERROR_HEADER = "X-Skip-Global-Error";

/**
 * Utility helper to create headers with the skip-global-error flag.
 */
export function withSkipGlobalError(
  headers?: HttpHeaders | Record<string, string | string[]>,
): { headers: HttpHeaders } {
  let httpHeaders =
    headers instanceof HttpHeaders ? headers : new HttpHeaders(headers);
  return { headers: httpHeaders.set(SKIP_GLOBAL_ERROR_HEADER, "true") };
}

function handleForbidden(
  error: HttpErrorResponse,
  router: Router,
  alertService: AlertService,
): void {
  const errorCode = error.error?.error || error.error?.code || "";
  const message = error.error?.message || "";

  if (
    errorCode === "EMAIL_NOT_VERIFIED" ||
    message.includes("verificar tu correo")
  ) {
    const email = error.error?.email || "";
    if (email) {
      void router.navigate(["/verify-email"], { queryParams: { email } });
    } else {
      void router.navigate(["/verify-email"]);
    }
    return;
  }

  if (
    errorCode === "ACCOUNT_PENDING_APPROVAL" ||
    message.includes("pendiente de aprobación")
  ) {
    void router.navigate(["/pending-approval"]);
    return;
  }

  if (errorCode === "ACCOUNT_BLOCKED" || message.includes("bloqueada")) {
    void router.navigate(["/account-notice"]);
    return;
  }

  alertService.showForbiddenDialog(
    message || "No tienes permisos para realizar esta acción.",
  );
}

function navigateToNotFound(router: Router): void {
  void router.navigate(["/404"]);
}

function notifyNotFound(
  error: HttpErrorResponse,
  alertService: AlertService,
): void {
  alertService.showError(
    "Recurso no encontrado",
    error.error?.message ||
      "El recurso solicitado no fue encontrado en el servidor.",
  );
}

/**
 * Muestra la pantalla de error sin cambiar la URL: así "Reintentar" (y F5)
 * recargan la página donde estaba el usuario, no `/500`.
 */
function navigateToServerError(router: Router): void {
  void router.navigate(["/500"], { skipLocationChange: true });
}

function notifyServerError(
  error: HttpErrorResponse,
  alertService: AlertService,
): void {
  alertService.showServerErrorDialog(
    error.error?.message ||
      "Ocurrió un error inesperado al procesar la solicitud en el servidor.",
  );
}

/**
 * Global HTTP Error Interceptor.
 * Catches HTTP error statuses and redirects to dedicated views (404, 500, verify-email, etc.)
 * or triggers global modal dialogs / toasts (429, 403, network failure).
 */
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);
  const alertService = inject(AlertService);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (req.headers.has(SKIP_GLOBAL_ERROR_HEADER)) {
        return throwError(() => error);
      }

      if (error.status === 0) {
        alertService.showOfflineToast();
      } else if (error.status === 403) {
        handleForbidden(error, router, alertService);
      } else if (error.status === 404) {
        if (req.headers.has("X-Redirect-On-404")) {
          navigateToNotFound(router);
        } else {
          notifyNotFound(error, alertService);
        }
      } else if (error.status === 429) {
        alertService.showRateLimitDialog();
      } else if (error.status >= 500 && error.status <= 504) {
        if (req.method === "GET" && !req.headers.has("X-Skip-500-Redirect")) {
          navigateToServerError(router);
        } else {
          notifyServerError(error, alertService);
        }
      }

      return throwError(() => error);
    }),
  );
};

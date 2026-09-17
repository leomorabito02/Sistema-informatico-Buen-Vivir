import { HttpInterceptorFn } from "@angular/common/http";

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  // Skip withCredentials for external (cross-origin) requests.
  // Cloudflare Workers and other public APIs respond with ACAO: *,
  // which is incompatible with withCredentials: true.
  const isExternal =
    req.url.startsWith("http") && !req.url.startsWith(window.location.origin);
  if (isExternal) {
    return next(req);
  }

  const clonedReq = req.clone({
    withCredentials: true,
  });

  return next(clonedReq);
};

import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { Observable, throwError } from "rxjs";
import { catchError, map } from "rxjs/operators";

import { environment } from "../../environments/environment";
import { withSkipGlobalError } from "../core/interceptors/error.interceptor";

export interface ConsultaContacto {
  nombre: string;
  email: string;
  asunto: string;
  mensaje: string;
  acepta_privacidad: boolean;
}

/** Error de envío con el mensaje para mostrar y, si los hay, los errores por campo. */
export class ErrorContacto extends Error {
  constructor(
    message: string,
    readonly errores: Record<string, string> = {},
  ) {
    super(message);
  }
}

/**
 * Formulario de "Contacto e información institucional": la consulta va al
 * servicio de Notificaciones, que la manda por email al centro
 * (POST /api/v1/notifications/contacto, público). No se guarda en la base.
 */
@Injectable({ providedIn: "root" })
export class ContactoService {
  private readonly http = inject(HttpClient);
  private readonly url = `${environment.apiUrl ?? ""}/api/v1/notifications/contacto`;

  enviar(consulta: ConsultaContacto): Observable<void> {
    return this.http
      .post<{ enviada: boolean }>(this.url, consulta, {
        ...withSkipGlobalError(),
        withCredentials: true,
      })
      .pipe(
        map(() => undefined),
        catchError((err: HttpErrorResponse) =>
          throwError(() => ContactoService.aError(err)),
        ),
      );
  }

  static aError(err: HttpErrorResponse): ErrorContacto {
    const body = (err.error ?? {}) as {
      error?: string;
      errores?: Record<string, string>;
    };
    if (err.status === 0) {
      return new ErrorContacto(
        "No hay conexión. Revisá tu internet y volvé a intentar.",
      );
    }
    return new ErrorContacto(
      body.error ||
        "No pudimos enviar la consulta. Escribinos a ceibuenvivir7@gmail.com.",
      body.errores ?? {},
    );
  }
}

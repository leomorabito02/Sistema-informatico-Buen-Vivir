import { Injectable, inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable, of } from "rxjs";
import { catchError } from "rxjs/operators";
import { environment } from "../../environments/environment";

export interface ImpactMetrics {
  activeParticipants: number;
  volunteers: number;
  courses?: number | string;
  updatedAt?: string;
}

const DEFAULT_METRICS: ImpactMetrics = {
  activeParticipants: 0,
  volunteers: 0,
  courses: "Pendiente",
};

@Injectable({
  providedIn: "root",
})
export class ImpactService {
  private readonly http = inject(HttpClient);

  getImpactMetrics(): Observable<ImpactMetrics> {
    const url = environment.workerImpactUrl;
    if (!url) {
      return of(DEFAULT_METRICS);
    }
    return this.http.get<ImpactMetrics>(url).pipe(
      catchError((error) => {
        console.warn(
          "Could not fetch Cloudflare Worker KV impact metrics, falling back to defaults.",
          error,
        );
        return of(DEFAULT_METRICS);
      }),
    );
  }
}

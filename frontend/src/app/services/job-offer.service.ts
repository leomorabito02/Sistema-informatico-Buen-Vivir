import { Injectable } from "@angular/core";
import { HttpClient, HttpParams } from "@angular/common/http";
import { Observable } from "rxjs";
import { environment } from "../../environments/environment";

export interface JobOffer {
  id?: string;
  empresa_id: string;
  creador_nombre?: string;
  titulo: string;
  descripcion: string;
  ubicacion: string;
  modalidad: string;
  tipo_empleo: string;
  salario: string;
  horario: string;
  requisitos: string[];
  fecha_limite?: string;
  estado: string;
  created_at?: string;
  updated_at?: string;
}

export interface JobApplication {
  id?: string;
  participante_id: string;
  /** Id de usuario: el del CV. Las postulaciones viejas no lo traen. */
  usuario_id?: string;
  /** Nombre del postulante (lo completa labor al listar los postulados). */
  nombre?: string;
  oferta_id: string;
  estado?: string;
  fecha_postulacion?: string;
}

@Injectable({
  providedIn: "root",
})
export class JobOfferService {
  private readonly apiUrl = `${environment.apiUrl}/api/v1/labor-integration/offers`;

  constructor(private readonly http: HttpClient) {}

  /** `q` lo busca labor-integration en título, empresa y ubicación (HU-45, HU-46). */
  getOffers(q = ""): Observable<JobOffer[]> {
    const texto = q.trim();
    const params = texto ? new HttpParams().set("q", texto) : undefined;
    return this.http.get<JobOffer[]>(this.apiUrl, {
      params,
      withCredentials: true,
    });
  }

  createOffer(offer: JobOffer): Observable<JobOffer> {
    return this.http.post<JobOffer>(this.apiUrl, offer, {
      withCredentials: true,
    });
  }

  updateOffer(id: string, offer: JobOffer): Observable<JobOffer> {
    return this.http.put<JobOffer>(`${this.apiUrl}/${id}`, offer, {
      withCredentials: true,
    });
  }

  deleteOffer(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`, {
      withCredentials: true,
    });
  }

  approveOffer(id: string): Observable<void> {
    return this.http.put<void>(
      `${this.apiUrl}/${id}/approve`,
      {},
      { withCredentials: true },
    );
  }

  rejectOffer(id: string): Observable<void> {
    return this.http.put<void>(
      `${this.apiUrl}/${id}/reject`,
      {},
      { withCredentials: true },
    );
  }

  applyToOffer(
    application: Partial<JobApplication>,
  ): Observable<JobApplication> {
    return this.http.post<JobApplication>(
      `${environment.apiUrl}/api/v1/labor-integration/applications`,
      application,
      { withCredentials: true },
    );
  }

  getApplicationsForOffer(offerId: string): Observable<JobApplication[]> {
    return this.http.get<JobApplication[]>(
      `${this.apiUrl}/${offerId}/applications`,
      { withCredentials: true },
    );
  }

  getMyApplications(): Observable<JobApplication[]> {
    return this.http.get<JobApplication[]>(
      `${environment.apiUrl}/api/v1/labor-integration/applications/me`,
      { withCredentials: true },
    );
  }

  updateApplicationStatus(
    applicationId: string,
    status: string,
  ): Observable<void> {
    return this.http.put<void>(
      `${environment.apiUrl}/api/v1/labor-integration/applications/${applicationId}/status`,
      { estado: status },
      { withCredentials: true },
    );
  }
}

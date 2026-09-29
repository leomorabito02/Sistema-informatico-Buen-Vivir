import { Injectable, inject } from "@angular/core";
import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { environment } from "../../environments/environment";
import { Observable, throwError } from "rxjs";
import { catchError } from "rxjs/operators";

export interface SocialProfile {
  id: string;
  network: string;
  username: string;
  url: string;
}

export interface Basics {
  name: string;
  email: string;
  phone: string;
  headline: string;
  location: string;
  website: string;
  summary: string;
  picture: string;
  profiles: SocialProfile[];
}

export interface WorkExperience {
  id: string;
  company: string;
  position: string;
  startDate: string;
  endDate: string;
  description: string;
}

export interface EducationItem {
  id: string;
  institution: string;
  degree: string;
  field: string;
  startDate: string;
  endDate: string;
  grade: string;
  status?: string;
}

export interface Skill {
  id: string;
  name: string;
  level: string;
}

export interface Certification {
  id: string;
  name: string;
  issuer: string;
  date: string;
}

export interface LanguageItem {
  id: string;
  name: string;
  level: string;
}

export interface Award {
  id: string;
  title: string;
  awarder: string;
  date: string;
  summary: string;
}

export interface VolunteerExperience {
  id: string;
  organization: string;
  position: string;
  startDate: string;
  endDate: string;
  summary: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  startDate: string;
  endDate: string;
  url: string;
  keywords: string[];
}

export interface Reference {
  id: string;
  name: string;
  description: string;
}

export interface Interest {
  id: string;
  name: string;
  keywords: string[];
}

export interface ResumeData {
  basics: Basics;
  work: WorkExperience[];
  education: EducationItem[];
  skills: Skill[];
  certifications: Certification[];
  languages: LanguageItem[];
  awards: Award[];
  volunteer: VolunteerExperience[];
  projects: Project[];
  references: Reference[];
  interests: Interest[];
  metadata?: any;
}

export interface Resume {
  id: string;
  user_id: string;
  title: string;
  pdf_url: string;
  public_url: string;
  rxresume_id: string;
  data: ResumeData;
  created_at: string;
  updated_at: string;
}

export interface ProfileInput {
  name: string;
  email: string;
  phone: string;
  location: string;
  headline: string;
}

@Injectable({
  providedIn: "root",
})
export class RxResumeService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  // getMyResume fetches the current authenticated user's CV.
  getMyResume(): Observable<Resume> {
    return this.http
      .get<Resume>(`${this.apiUrl}/api/v1/rxresume/me`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // updateMyResume saves modifications to the logged-in user's resume data.
  updateMyResume(data: ResumeData): Observable<Resume> {
    return this.http
      .put<Resume>(`${this.apiUrl}/api/v1/rxresume/me`, data, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // generateResume triggers sync with RxResume using profile data
  generateResume(profile: ProfileInput): Observable<Resume> {
    return this.http
      .post<Resume>(`${this.apiUrl}/api/v1/rxresume/me/generate`, profile, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // uploadAndParsePDF uploads a PDF file for offline extraction and GCS storage.
  uploadAndParsePDF(file: File): Observable<Resume> {
    const formData = new FormData();
    formData.append("file", file, file.name);

    return this.http
      .post<Resume>(`${this.apiUrl}/api/v1/rxresume/me/parse-pdf`, formData, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // uploadPicture uploads a profile picture to Reactive Resume
  uploadPicture(file: File): Observable<Resume> {
    const formData = new FormData();
    formData.append("file", file, file.name);

    return this.http
      .post<Resume>(`${this.apiUrl}/api/v1/rxresume/me/picture`, formData, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // getAllParticipantsResumes fetches CV list for staff roles.
  getAllParticipantsResumes(): Observable<Resume[]> {
    return this.http
      .get<Resume[]>(`${this.apiUrl}/api/v1/rxresume/participants`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // getUserResumeByAdmin fetches a specific participant's CV by userID for staff roles.
  getUserResumeByAdmin(userId: string): Observable<Resume> {
    return this.http
      .get<Resume>(`${this.apiUrl}/api/v1/rxresume/user?user_id=${userId}`, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // updateUserResumeByAdmin modifies a specific participant's CV by userID for staff roles.
  updateUserResumeByAdmin(
    userId: string,
    data: ResumeData,
  ): Observable<Resume> {
    return this.http
      .put<Resume>(
        `${this.apiUrl}/api/v1/rxresume/user?user_id=${userId}`,
        data,
        {
          withCredentials: true,
        },
      )
      .pipe(catchError(this.handleError));
  }

  // uploadUserPictureByAdmin uploads a profile picture for a specific participant's CV by userID for staff roles.
  uploadUserPictureByAdmin(userId: string, file: File): Observable<Resume> {
    const formData = new FormData();
    formData.append("file", file, file.name);

    return this.http
      .post<Resume>(`${this.apiUrl}/api/v1/rxresume/user/picture?user_id=${userId}`, formData, {
        withCredentials: true,
      })
      .pipe(catchError(this.handleError));
  }

  // getPublicResume fetches a participant's CV by userID for public viewing.
  getPublicResume(userId: string): Observable<Resume> {
    return this.http
      .get<Resume>(`${this.apiUrl}/api/v1/rxresume/public/${userId}`)
      .pipe(catchError(this.handleError));
  }

  // downloadPublicPDF triggers the PDF download for a public CV.
  downloadPublicPDF(userId: string): void {
    this.http
      .get(`${this.apiUrl}/api/v1/rxresume/public/${userId}/pdf`, {
        responseType: "blob",
      })
      .subscribe({
        next: (blob) => {
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = url;
          a.download = `cv-${userId}.pdf`;
          document.body.appendChild(a);
          a.click();
          a.remove();
          window.URL.revokeObjectURL(url);
        },
        error: (err) => {
          console.error("Error al descargar el PDF:", err);
        },
      });
  }

  private readonly handleError = (error: HttpErrorResponse) => {
    let errorMessage = "Ocurrió un error con el servicio RxResume.";
    if (
      error.error &&
      typeof error.error === "object" &&
      "message" in error.error
    ) {
      errorMessage = (error.error as { message: string }).message;
    } else if (
      typeof error.error === "string" &&
      error.error.trim().length > 0
    ) {
      errorMessage = error.error.trim();
    } else if (error.status === 401) {
      errorMessage = "No autorizado.";
    } else if (error.status === 403) {
      errorMessage = "No tiene permisos para gestionar este CV.";
    }
    return throwError(() => new Error(errorMessage));
  };
}

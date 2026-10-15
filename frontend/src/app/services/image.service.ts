import { Injectable } from "@angular/core";
import { environment } from "../../environments/environment";

@Injectable({
  providedIn: "root",
})
export class ImageService {
  /**
   * Resolves raw image URLs or paths to fully-qualified accessible URLs.
   */
  resolveImageUrl(url?: string | boolean | null): string {
    if (!url || typeof url !== "string") return "";
    if (url.startsWith("http://localhost:8080/api/v1/")) {
      const path = url.replace("http://localhost:8080", "");
      return `${environment.apiUrl || ""}${path}`;
    }
    if (url.startsWith("/")) {
      return `${environment.apiUrl || ""}${url}`;
    }
    return url;
  }
}

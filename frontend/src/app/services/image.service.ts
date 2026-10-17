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
    const trimmed = url.trim();
    if (!trimmed) return "";
    const base =
      environment.apiUrl ||
      (typeof window !== "undefined" ? window.location.origin : "");
    let fullUrl = trimmed;
    if (fullUrl.startsWith("http://localhost:8080/api/v1/")) {
      const path = fullUrl.replace("http://localhost:8080", "");
      fullUrl = `${base}${path}`;
    } else if (fullUrl.startsWith("/")) {
      fullUrl = `${base}${fullUrl}`;
    }
    try {
      return encodeURI(decodeURI(fullUrl));
    } catch {
      return encodeURI(fullUrl);
    }
  }
}

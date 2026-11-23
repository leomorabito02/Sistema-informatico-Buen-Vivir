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
    // Codifica espacios y tildes, pero respeta lo que ya viene codificado:
    // rxresume arma `?path=pictures%2Fusr-…` y encodeURI(decodeURI()) lo
    // dejaba en `pictures%252F…` (decodeURI no decodifica %2F), con lo que
    // la foto respondía 403 (22/11/2026).
    return encodeURI(fullUrl).replace(/%25([0-9A-Fa-f]{2})/g, "%$1");
  }
}

import { Injectable, inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { firstValueFrom } from "rxjs";
import { Resume, ResumeData } from "./rxresume.service";
import { ImageService } from "./image.service";

export interface GenerateCvPdfOptions {
  primaryColor?: string;
  template?: string;
  pictureUrl?: string;
}

@Injectable({
  providedIn: "root",
})
export class CvPdfService {
  private readonly http = inject(HttpClient);
  private readonly imageService = inject(ImageService);
  private enginePromise: Promise<
    typeof import("@buenvivir/cv-engine/browser")
  > | null = null;

  private loadEngine(): Promise<typeof import("@buenvivir/cv-engine/browser")> {
    this.enginePromise ??= import("@buenvivir/cv-engine/browser");
    return this.enginePromise;
  }

  /**
   * Resolves a raw image URL/path into an in-memory base64 Data URL (JPEG/PNG)
   * so @react-pdf/renderer renders it without CORS, cookie, or format issues.
   */
  async resolvePictureAsDataUrl(rawUrl?: string | null): Promise<string> {
    if (!rawUrl || typeof rawUrl !== "string" || !rawUrl.trim()) {
      return "";
    }
    const trimmed = rawUrl.trim();
    if (trimmed.startsWith("data:image/")) {
      return trimmed;
    }

    const resolvedUrl = this.imageService.resolveImageUrl(trimmed);
    if (!resolvedUrl) return "";

    // 1. Try fetching via Angular's HttpClient with cache-busting to bypass any non-CORS browser cache
    try {
      const sep = resolvedUrl.includes("?") ? "&" : "?";
      const fetchUrl = `${resolvedUrl}${sep}_t=${Date.now()}`;
      const blob = await firstValueFrom(
        this.http.get(fetchUrl, {
          responseType: "blob",
          withCredentials: true,
        }),
      );
      if (blob && blob.size > 0) {
        return await this.blobToDataUrl(blob);
      }
    } catch {
      // Fallback to Image + Canvas
    }

    // 2. Fallback to HTMLImageElement + Canvas
    const sep = resolvedUrl.includes("?") ? "&" : "?";
    const canvasUrl = `${resolvedUrl}${sep}_t=${Date.now()}`;
    return await this.loadImageViaCanvas(canvasUrl);
  }

  private async blobToDataUrl(blob: Blob): Promise<string> {
    const isStandardFormat =
      blob.type === "image/png" ||
      blob.type === "image/jpeg" ||
      blob.type === "image/jpg";

    if (isStandardFormat) {
      return new Promise((resolve) => {
        const reader = new FileReader();
        reader.onloadend = () => {
          resolve(typeof reader.result === "string" ? reader.result : "");
        };
        reader.onerror = () => resolve("");
        reader.readAsDataURL(blob);
      });
    }

    if (typeof createImageBitmap === "function") {
      try {
        const bitmap = await createImageBitmap(blob);
        const canvas = document.createElement("canvas");
        canvas.width = bitmap.width;
        canvas.height = bitmap.height;
        const ctx = canvas.getContext("2d");
        if (ctx) {
          ctx.drawImage(bitmap, 0, 0);
          return canvas.toDataURL("image/jpeg", 0.92);
        }
      } catch {
        // Fallback to Image or FileReader
      }
    }

    return new Promise((resolve) => {
      const objectUrl = URL.createObjectURL(blob);
      this.loadImageViaCanvas(objectUrl)
        .then((dataUrl) => {
          URL.revokeObjectURL(objectUrl);
          if (dataUrl) {
            resolve(dataUrl);
          } else {
            const reader = new FileReader();
            reader.onloadend = () => {
              resolve(typeof reader.result === "string" ? reader.result : "");
            };
            reader.onerror = () => resolve("");
            reader.readAsDataURL(blob);
          }
        })
        .catch(() => {
          URL.revokeObjectURL(objectUrl);
          const reader = new FileReader();
          reader.onloadend = () => {
            resolve(typeof reader.result === "string" ? reader.result : "");
          };
          reader.onerror = () => resolve("");
          reader.readAsDataURL(blob);
        });
    });
  }

  private loadImageViaCanvas(url: string): Promise<string> {
    return new Promise((resolve) => {
      if (typeof document === "undefined") {
        resolve("");
        return;
      }
      const img = new Image();
      if (!url.startsWith("blob:")) {
        img.crossOrigin = "anonymous";
      }
      img.onload = () => {
        try {
          const canvas = document.createElement("canvas");
          canvas.width = img.naturalWidth || img.width || 160;
          canvas.height = img.naturalHeight || img.height || 160;
          const ctx = canvas.getContext("2d");
          if (!ctx) {
            resolve("");
            return;
          }
          ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
          const dataUrl = canvas.toDataURL("image/jpeg", 0.92);
          resolve(dataUrl);
        } catch {
          resolve("");
        }
      };
      img.onerror = () => resolve("");
      img.src = url;
    });
  }

  async generatePdfBlob(
    resume: Resume | ResumeData,
    opts: GenerateCvPdfOptions = {},
  ): Promise<Blob> {
    const engine = await this.loadEngine();
    const resumeData =
      "data" in resume && resume.data ? resume.data : (resume as ResumeData);
    const templateFromResume =
      resumeData.metadata && typeof resumeData.metadata.template === "string"
        ? resumeData.metadata.template
        : undefined;

    const rawPic = opts.pictureUrl || resumeData.basics?.picture;
    const pictureDataUrl = await this.resolvePictureAsDataUrl(rawPic);

    return await engine.generateCvPdfBlob(resume, {
      template: opts.template || templateFromResume,
      primaryColor: opts.primaryColor,
      pictureUrl: pictureDataUrl || "",
    });
  }

  async downloadPdf(
    resume: Resume | ResumeData,
    filename = "curriculum-vitae.pdf",
    opts: GenerateCvPdfOptions = {},
  ): Promise<void> {
    const blob = await this.generatePdfBlob(resume, opts);
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1500);
  }
}

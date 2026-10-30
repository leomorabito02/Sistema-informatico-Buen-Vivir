declare module "@buenvivir/cv-engine/browser" {
  export interface GenerateCvPdfBlobOptions {
    template?: string;
    pictureUrl?: string;
    primaryColor?: string;
  }

  export function generateCvPdfBlob(
    cv: unknown,
    opts?: GenerateCvPdfBlobOptions,
  ): Promise<Blob>;
}

declare module "@buenvivir/cv-engine" {
  export interface GenerateCvPdfBufferOptions {
    template?: string;
    pictureUrl?: string;
    primaryColor?: string;
  }

  export function generateCvPdfBuffer(
    cv: unknown,
    opts?: GenerateCvPdfBufferOptions,
  ): Promise<Buffer>;
}

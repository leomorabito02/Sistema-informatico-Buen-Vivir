import { createResumePdfBlob } from "../vendor/pdf/browser";
import { mapToResumeData, type InternalResume, type InternalCvData } from "./map-to-resume-data";
import type { Template } from "../vendor/schema/templates";

// @react-pdf llama a Buffer.isBuffer() al cargar imágenes; en el navegador
// Buffer no existe y el PDF con foto fallaba con "Buffer is not defined".
const globalConBuffer = globalThis as { Buffer?: { isBuffer: (v: unknown) => boolean } };
globalConBuffer.Buffer ??= { isBuffer: () => false };

export interface GeneratePdfOptions {
  template?: Template;
  pictureUrl?: string;
  primaryColor?: string;
}

export async function generateCvPdfBlob(
  cv: InternalResume | InternalCvData,
  opts: GeneratePdfOptions = {}
): Promise<Blob> {
  const data = mapToResumeData(cv, {
    pictureUrl: opts.pictureUrl,
    overridePrimaryColor: opts.primaryColor,
  });

  return await createResumePdfBlob({
    data,
    template: opts.template ?? (data.metadata.template as Template),
  });
}

export * from "./map-to-resume-data";
export * from "./fonts";

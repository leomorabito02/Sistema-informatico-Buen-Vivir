import { createResumePdfBlob } from "../vendor/pdf/browser";
import { mapToResumeData, type InternalResume, type InternalCvData } from "./map-to-resume-data";
import type { Template } from "../vendor/schema/templates";

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

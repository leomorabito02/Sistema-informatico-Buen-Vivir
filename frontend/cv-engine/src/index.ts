import React, { createElement } from "react";
(globalThis as any).React = React;
import type { Buffer } from "node:buffer";
import { renderToBuffer } from "@react-pdf/renderer";
import { ResumeDocument } from "../vendor/pdf/document";
import { mapToResumeData, type InternalResume, type InternalCvData } from "./map-to-resume-data";
import type { Template } from "../vendor/schema/templates";

export interface GeneratePdfBufferOptions {
  template?: Template;
  pictureUrl?: string;
  primaryColor?: string;
}

export async function generateCvPdfBuffer(
  cv: InternalResume | InternalCvData,
  opts: GeneratePdfBufferOptions = {}
): Promise<Buffer> {
  const data = mapToResumeData(cv, {
    pictureUrl: opts.pictureUrl,
    overridePrimaryColor: opts.primaryColor,
  });

  const document = createElement(ResumeDocument, {
    data,
    template: opts.template ?? (data.metadata.template as Template),
  }) as Parameters<typeof renderToBuffer>[0];

  return await renderToBuffer(document);
}

export * from "./map-to-resume-data";
export * from "./fonts";

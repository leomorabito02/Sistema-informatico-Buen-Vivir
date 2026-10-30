import React from "react";
(globalThis as any).React = React;
import http, { type IncomingMessage, type ServerResponse } from "node:http";
import type { Buffer } from "node:buffer";
import { generateCvPdfBuffer } from "../index";
import type { InternalCvData, InternalResume } from "../map-to-resume-data";
import type { Template } from "../../vendor/schema/templates";

const portEnv = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env?.["PORT"];
const PORT = Number(portEnv || 5005);
const MAX_PAYLOAD_BYTES = 10 * 1024 * 1024; // 10MB limit

interface RenderRequestBody {
  cv: InternalResume | InternalCvData;
  template?: Template;
  pictureUrl?: string;
  primaryColor?: string;
}

const server = http.createServer((req: IncomingMessage, res: ServerResponse) => {
  // CORS / Security headers
  res.setHeader("X-Content-Type-Options", "nosniff");

  if (req.method === "GET" && req.url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ status: "ok", service: "cv-renderer" }));
    return;
  }

  if (req.method === "POST" && req.url === "/render") {
    let rawBody = "";
    let receivedBytes = 0;

    req.on("data", (chunk: Buffer | string) => {
      receivedBytes += chunk.length;
      if (receivedBytes > MAX_PAYLOAD_BYTES) {
        res.writeHead(413, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "Payload too large" }));
        req.destroy();
        return;
      }
      rawBody += chunk;
    });

    req.on("end", async () => {
      try {
        if (!rawBody) {
          res.writeHead(400, { "Content-Type": "application/json" });
          res.end(JSON.stringify({ error: "Missing request body" }));
          return;
        }

        const body = JSON.parse(rawBody) as RenderRequestBody;
        if (!body.cv) {
          res.writeHead(400, { "Content-Type": "application/json" });
          res.end(JSON.stringify({ error: "Field 'cv' is required" }));
          return;
        }

        const pdfBuffer = await generateCvPdfBuffer(body.cv, {
          template: body.template,
          pictureUrl: body.pictureUrl,
          primaryColor: body.primaryColor,
        });

        res.writeHead(200, {
          "Content-Type": "application/pdf",
          "Content-Length": pdfBuffer.length,
          "Cache-Control": "no-cache",
        });
        res.end(pdfBuffer);
      } catch (err: any) {
        console.error("[cv-renderer] Render error:", err);
        res.writeHead(500, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "Failed to render PDF", details: err?.message }));
      }
    });

    return;
  }

  res.writeHead(404, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ error: "Not found" }));
});

server.listen(PORT, "0.0.0.0", () => {
  console.log(`[cv-renderer] Service listening on port ${PORT}`);
});

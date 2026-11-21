import React from "react";
(globalThis as any).React = React;
import http, { type IncomingMessage, type ServerResponse } from "node:http";
import type { Buffer } from "node:buffer";
import { generateCvPdfBuffer } from "../index";
import type { InternalCvData, InternalResume } from "../map-to-resume-data";
import type { Template } from "../../vendor/schema/templates";

interface NodeProcessLike {
  env?: Record<string, string | undefined>;
  memoryUsage?: () => { rss: number; heapUsed: number; heapTotal: number };
  on?: (signal: string, listener: () => void) => void;
  exit?: (code?: number) => void;
}

const nodeProcess = (globalThis as { process?: NodeProcessLike }).process;
const PORT = Number(nodeProcess?.env?.["PORT"] || 5005);
const METRICS_ENABLED =
  (nodeProcess?.env?.["METRICS_ENABLED"] || "").toLowerCase() === "true";
const APP_ENV = nodeProcess?.env?.["APP_ENV"] || nodeProcess?.env?.["NODE_ENV"] || "development";
const MAX_PAYLOAD_BYTES = 10 * 1024 * 1024; // 10MB limit
const STARTED_AT_MS = Date.now();
const DURATION_BUCKETS = [0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30];

const httpRequests: Record<string, number> = {};
const renderBucketCounts = DURATION_BUCKETS.map(() => 0);
let renderDurationSum = 0;
let renderDurationCount = 0;
let inFlightRequests = 0;

function logJSON(
  severity: "INFO" | "WARNING" | "ERROR",
  message: string,
  extra: Record<string, unknown> = {},
): void {
  const payload = JSON.stringify({
    time: new Date().toISOString(),
    severity,
    service: "cv-renderer",
    env: APP_ENV,
    message,
    ...extra,
  });
  if (severity === "ERROR") {
    console.error(payload);
  } else {
    console.log(payload);
  }
}

function statusClass(code: number): string {
  if (code >= 500) return "5xx";
  if (code >= 400) return "4xx";
  if (code >= 300) return "3xx";
  return "2xx";
}

function severityForStatus(code: number): "INFO" | "WARNING" | "ERROR" {
  if (code >= 500) return "ERROR";
  if (code >= 400) return "WARNING";
  return "INFO";
}

function recordHttpMetric(method: string, route: string, status: number, durationSec: number): void {
  const key = `${method}|${route}|${statusClass(status)}`;
  httpRequests[key] = (httpRequests[key] || 0) + 1;
  if (route === "/render") {
    renderDurationSum += durationSec;
    renderDurationCount += 1;
    for (let i = 0; i < DURATION_BUCKETS.length; i++) {
      if (durationSec <= DURATION_BUCKETS[i]) {
        renderBucketCounts[i] += 1;
      }
    }
  }
}

function renderPrometheusMetrics(): string {
  const lines: string[] = [
    "# HELP buenvivir_http_requests_total Total HTTP requests handled by the service.",
    "# TYPE buenvivir_http_requests_total counter",
  ];
  for (const key of Object.keys(httpRequests).sort((a, b) => a.localeCompare(b))) {
    const [method, route, sc] = key.split("|");
    lines.push(
      `buenvivir_http_requests_total{service="cv-renderer",method="${method}",route="${route}",status_class="${sc}"} ${httpRequests[key]}`,
    );
  }

  lines.push(
    "# HELP buenvivir_http_request_duration_seconds HTTP request latency distribution in seconds.",
    "# TYPE buenvivir_http_request_duration_seconds histogram",
  );
  for (let i = 0; i < DURATION_BUCKETS.length; i++) {
    lines.push(
      `buenvivir_http_request_duration_seconds_bucket{service="cv-renderer",method="POST",route="/render",le="${DURATION_BUCKETS[i]}"} ${renderBucketCounts[i]}`,
    );
  }
  lines.push(
    `buenvivir_http_request_duration_seconds_bucket{service="cv-renderer",method="POST",route="/render",le="+Inf"} ${renderDurationCount}`,
    `buenvivir_http_request_duration_seconds_sum{service="cv-renderer",method="POST",route="/render"} ${renderDurationSum}`,
    `buenvivir_http_request_duration_seconds_count{service="cv-renderer",method="POST",route="/render"} ${renderDurationCount}`,
    "# HELP buenvivir_http_requests_in_flight Current number of in-flight HTTP requests.",
    "# TYPE buenvivir_http_requests_in_flight gauge",
    `buenvivir_http_requests_in_flight{service="cv-renderer"} ${inFlightRequests}`,
  );

  const mem = nodeProcess?.memoryUsage?.();
  if (mem) {
    lines.push(
      "# HELP process_resident_memory_bytes Resident memory size in bytes.",
      "# TYPE process_resident_memory_bytes gauge",
      `process_resident_memory_bytes{service="cv-renderer"} ${mem.rss}`,
      "# HELP nodejs_heap_size_used_bytes Node.js heap used in bytes.",
      "# TYPE nodejs_heap_size_used_bytes gauge",
      `nodejs_heap_size_used_bytes{service="cv-renderer"} ${mem.heapUsed}`,
    );
  }

  const uptimeSec = ((Date.now() - STARTED_AT_MS) / 1000).toFixed(3);
  lines.push(
    "# HELP process_uptime_seconds Seconds since the service process started.",
    "# TYPE process_uptime_seconds gauge",
    `process_uptime_seconds{service="cv-renderer"} ${uptimeSec}`,
  );
  return lines.join("\n") + "\n";
}

interface RenderRequestBody {
  cv: InternalResume | InternalCvData;
  template?: Template;
  pictureUrl?: string;
  primaryColor?: string;
}

const server = http.createServer((req: IncomingMessage, res: ServerResponse) => {
  res.setHeader("X-Content-Type-Options", "nosniff");

  if (METRICS_ENABLED && req.method === "GET" && req.url === "/metrics") {
    res.writeHead(200, { "Content-Type": "text/plain; version=0.0.4; charset=utf-8" });
    res.end(renderPrometheusMetrics());
    return;
  }

  const startedAt = Date.now();
  const correlationId = String(req.headers["x-correlation-id"] || "");
  if (correlationId) {
    res.setHeader("X-Correlation-ID", correlationId);
  }
  const method = req.method || "GET";
  const route = req.url === "/health" || req.url === "/render" ? req.url : "unmatched";

  inFlightRequests += 1;
  res.on("finish", () => {
    inFlightRequests = Math.max(0, inFlightRequests - 1);
    const elapsedMs = Date.now() - startedAt;
    recordHttpMetric(method, route, res.statusCode, elapsedMs / 1000);
    if (route === "/health" && res.statusCode < 400) {
      return;
    }
    logJSON(severityForStatus(res.statusCode), "http_request", {
      method,
      route,
      status: res.statusCode,
      latency_ms: elapsedMs,
      ...(correlationId ? { correlation_id: correlationId } : {}),
    });
  });

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
        logJSON("ERROR", "render_pdf_failed", {
          error: err?.message || String(err),
          ...(correlationId ? { correlation_id: correlationId } : {}),
        });
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
  logJSON("INFO", "cv-renderer listening", { port: PORT, metrics_enabled: METRICS_ENABLED });
});

function shutdown(signal: string): void {
  logJSON("INFO", "graceful_shutdown_started", { signal });
  server.close(() => {
    nodeProcess?.exit?.(0);
  });
}

nodeProcess?.on?.("SIGTERM", () => shutdown("SIGTERM"));
nodeProcess?.on?.("SIGINT", () => shutdown("SIGINT"));

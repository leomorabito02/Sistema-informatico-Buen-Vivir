/**
 * Cloudflare Worker for Landing Page Impact Counters (KV)
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    // Handle CORS preflight
    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type, Authorization",
        },
      });
    }

    // GET /api/impact -> Retrieve stats from KV Edge
    if (request.method === "GET" && url.pathname === "/api/impact") {
      const kvData = await env.IMPACT_KV.get("impact_stats");
      const defaultStats = {
        activeParticipants: 248,
        volunteers: 32,
        courses: "Pendiente",
        updatedAt: new Date().toISOString(),
      };

      const body = kvData || JSON.stringify(defaultStats);

      return new Response(body, {
        status: 200,
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
          "Cache-Control": "public, max-age=60, s-maxage=300",
        },
      });
    }

    // POST /api/impact/update -> Update stats in KV (Secured by Bearer token)
    if (request.method === "POST" && url.pathname === "/api/impact/update") {
      const authHeader = request.headers.get("Authorization") || "";
      const expectedToken = env.WORKER_SYNC_TOKEN || "default-secret-token";

      if (!authHeader.startsWith("Bearer ") || authHeader.slice(7) !== expectedToken) {
        return new Response(JSON.stringify({ error: "Unauthorized" }), {
          status: 401,
          headers: { "Content-Type": "application/json" },
        });
      }

      try {
        const payload = await request.json();

        const rawKvData = await env.IMPACT_KV.get("impact_stats");
        let existingStats = {};
        if (rawKvData) {
          try {
            existingStats = JSON.parse(rawKvData);
          } catch (e) {
            existingStats = {};
          }
        }

        const stats = {
          activeParticipants: payload.activeParticipants ?? existingStats.activeParticipants ?? 0,
          volunteers: payload.volunteers ?? existingStats.volunteers ?? 0,
          courses: payload.courses ?? existingStats.courses ?? "Pendiente",
          updatedAt: new Date().toISOString(),
        };

        await env.IMPACT_KV.put("impact_stats", JSON.stringify(stats));

        return new Response(JSON.stringify({ success: true, stats }), {
          status: 200,
          headers: {
            "Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
          },
        });
      } catch (err) {
        console.error("Error updating impact stats payload:", err);
        return new Response(JSON.stringify({ error: "Invalid JSON payload" }), {
          status: 400,
          headers: { "Content-Type": "application/json" },
        });
      }
    }

    return new Response(JSON.stringify({ error: "Not Found" }), {
      status: 404,
      headers: { "Content-Type": "application/json" },
    });
  },
};

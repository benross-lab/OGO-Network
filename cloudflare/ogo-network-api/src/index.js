const ALLOWED = new Set([
  "/v1/network",
  "/v1/validators",
  "/v1/epochs",
]);

function isAllowedPath(path) {
  return ALLOWED.has(path)
    || /^\/v1\/balances\/[^/]+$/.test(path)
    || /^\/v1\/transactions\/[^/]+(?:\/verify)?$/.test(path)
    || /^\/v1\/transactions\/(prepare|broadcast)$/.test(path)
    || /^\/v1\/addresses\/[^/]+\/transactions$/.test(path);
}

function json(status, value) {
  return new Response(JSON.stringify(value), {
    status,
    headers: {
      "content-type": "application/json; charset=utf-8",
      "cache-control": "no-store",
      "x-content-type-options": "nosniff",
    },
  });
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "access-control-allow-origin": env.CORS_ORIGIN || "*",
          "access-control-allow-methods": "GET, POST, OPTIONS",
          "access-control-allow-headers": "content-type, idempotency-key",
        },
      });
    }
    if (url.pathname === "/healthz") {
      return json(200, { ok: true, service: "ogo-network-api-edge", development_only: true });
    }
    if (!url.pathname.startsWith("/v1/") || !isAllowedPath(url.pathname)) {
      return json(404, { error: { code: "not_found", message: "endpoint not found" } });
    }
    if (!["GET", "POST"].includes(request.method)) {
      return json(405, { error: { code: "method_not_allowed", message: "method not allowed" } });
    }
    if (!env.ORIGIN_BASE_URL) {
      return json(503, { error: { code: "origin_unconfigured", message: "ORIGIN_BASE_URL is not configured" } });
    }
    const contentLength = Number(request.headers.get("content-length") || 0);
    if (contentLength > 1024 * 1024) {
      return json(413, { error: { code: "request_too_large", message: "request body exceeds 1 MiB" } });
    }
    const upstream = new URL(url.pathname + url.search, env.ORIGIN_BASE_URL);
    const headers = new Headers(request.headers);
    headers.set("x-ogo-edge", "ogo-network-api");
    // Do not add credentials or sign requests here. The origin only receives the
    // signed transaction bytes supplied by the client for broadcast.
    const response = await fetch(upstream, {
      method: request.method,
      headers,
      body: request.method === "GET" ? undefined : request.body,
    });
    const out = new Response(response.body, response);
    out.headers.set("access-control-allow-origin", env.CORS_ORIGIN || "*");
    out.headers.set("cache-control", "no-store");
    out.headers.set("x-content-type-options", "nosniff");
    return out;
  },
};

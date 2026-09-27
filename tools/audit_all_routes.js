// tools/audit_all_routes.js
// Systematic route and endpoint auditor for go-9router

const BASE_URL = process.env.BASE_URL || "http://127.0.0.1:20128";

const ENDPOINTS = [
  // 1. Core Gateway & Diagnostics API
  { method: "GET", path: "/api/health", expected: [200] },
  { method: "GET", path: "/api/version", expected: [200] },
  { method: "GET", path: "/api/init", expected: [200] },
  { method: "GET", path: "/api/locale", expected: [200] },
  { method: "GET", path: "/api/changelog", expected: [200] },
  { method: "GET", path: "/api/auth/status", expected: [200] },
  { method: "GET", path: "/api/settings/require-login", expected: [200] },
  { method: "GET", path: "/api/settings", expected: [200] },
  { method: "GET", path: "/api/settings/database", expected: [200] },
  { method: "GET", path: "/api/providers", expected: [200] },
  { method: "GET", path: "/api/providers/client", expected: [200] },
  { method: "GET", path: "/api/providers/suggested-models", expected: [200] },
  { method: "GET", path: "/api/provider-nodes", expected: [200] },
  { method: "GET", path: "/api/keys", expected: [200] },
  { method: "GET", path: "/api/combos", expected: [200] },
  { method: "GET", path: "/api/combos/presets", expected: [200] },
  { method: "GET", path: "/api/proxy-pools", expected: [200] },
  { method: "GET", path: "/api/models", expected: [200] },
  { method: "GET", path: "/api/models/alias", expected: [200] },
  { method: "GET", path: "/api/models/custom", expected: [200] },
  { method: "GET", path: "/api/models/availability", expected: [200] },
  { method: "GET", path: "/api/models/disabled", expected: [200] },
  { method: "GET", path: "/api/usage", expected: [200] },
  { method: "GET", path: "/api/usage/history", expected: [200] },
  { method: "GET", path: "/api/usage/chart", expected: [200] },
  { method: "GET", path: "/api/usage/providers", expected: [200] },
  { method: "GET", path: "/api/usage/logs", expected: [200] },
  { method: "GET", path: "/api/usage/stats", expected: [200] },
  { method: "GET", path: "/api/token-saver/summary", expected: [200] },
  {
    method: "POST",
    path: "/api/token-saver/test",
    body: { text: "Hello from audit test! Please summarize." },
    expected: [200]
  },
  { method: "GET", path: "/api/headroom", expected: [200] },
  { method: "GET", path: "/api/pxpipe", expected: [200] },
  { method: "GET", path: "/api/tunnel", expected: [200] },
  { method: "GET", path: "/api/tunnel/status", expected: [200] },
  { method: "GET", path: "/api/tunnel/tailscale-check", expected: [200] },
  { method: "GET", path: "/api/translator", expected: [200] },
  { method: "GET", path: "/api/cli-tools", expected: [200] },
  { method: "GET", path: "/api/media-providers", expected: [200] },
  { method: "GET", path: "/api/tags", expected: [200] },
  { method: "GET", path: "/api/pricing", expected: [200] },
  { method: "GET", path: "/api/payments", expected: [200] },
  { method: "GET", path: "/api/users", expected: [200] },
  { method: "GET", path: "/api/cloud", expected: [200] },
  { method: "GET", path: "/api/auth/oidc", expected: [200] },
  { method: "GET", path: "/api/auth/saml", expected: [200] },

  // 2. Standard OpenAI / Anthropic Compatible Protocol Endpoints
  { method: "GET", path: "/v1/models", expected: [200] },
  { method: "GET", path: "/api/v1/models", expected: [200] },
  { method: "GET", path: "/api/v1", expected: [200] },
  { method: "GET", path: "/api/v1beta", expected: [200] },
  {
    method: "POST",
    path: "/v1/chat/completions",
    body: {
      model: "openrouter/default",
      messages: [{ role: "user", content: "ping" }]
    },
    expected: [200, 400, 401, 502, 503] // NOT 404
  },
  {
    method: "POST",
    path: "/v1/messages",
    body: {
      model: "openrouter/default",
      messages: [{ role: "user", content: "ping" }]
    },
    expected: [200, 400, 401, 502, 503] // NOT 404
  },

  // 3. Web Dashboard Frontend Routes (Modern Pages)
  { method: "GET", path: "/", expected: [200, 307, 308] },
  { method: "GET", path: "/login", expected: [200] },
  { method: "GET", path: "/endpoint", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/endpoint", expected: [200] },
  { method: "GET", path: "/providers", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/providers", expected: [200] },
  { method: "GET", path: "/providers/new", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/providers/new", expected: [200] },
  { method: "GET", path: "/proxy-pools", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/proxy-pools", expected: [200] },
  { method: "GET", path: "/combos", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/combos", expected: [200] },
  { method: "GET", path: "/skills", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/skills", expected: [200] },
  { method: "GET", path: "/token-saver", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/token-saver", expected: [200] },
  { method: "GET", path: "/cli-tools", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/cli-tools", expected: [200] },
  { method: "GET", path: "/profile", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/profile", expected: [200] },
  { method: "GET", path: "/usage", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/usage", expected: [200] },
  { method: "GET", path: "/quota", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/quota", expected: [200] },
  { method: "GET", path: "/console-log", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/console-log", expected: [200] },
  { method: "GET", path: "/mitm", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/mitm", expected: [200] },
  { method: "GET", path: "/pxpipe", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/pxpipe", expected: [200] },
  { method: "GET", path: "/translator", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/translator", expected: [200] },
  { method: "GET", path: "/settings/pricing", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/settings/pricing", expected: [200] },
  { method: "GET", path: "/basic-chat", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/basic-chat", expected: [200] },
  { method: "GET", path: "/media-providers/web", expected: [200, 307, 308] },
  { method: "GET", path: "/dashboard/media-providers/web", expected: [200] }
];

async function runAudit() {
  console.log(`\n======================================================`);
  console.log(`  🔍 GO-9ROUTER COMPREHENSIVE ROUTE & ENDPOINT AUDIT  `);
  console.log(`  Gateway Base URL: ${BASE_URL}`);
  console.log(`======================================================\n`);

  let passCount = 0;
  let failCount = 0;
  const failures = [];

  for (const item of ENDPOINTS) {
    const url = `${BASE_URL}${item.path}`;
    const start = Date.now();
    try {
      const options = {
        method: item.method,
        headers: {
          "Accept": "application/json, text/html, */*"
        },
        redirect: "manual"
      };

      if (item.body) {
        options.headers["Content-Type"] = "application/json";
        options.body = JSON.stringify(item.body);
      }

      const res = await fetch(url, options);
      const elapsed = Date.now() - start;
      const isExpected = item.expected.includes(res.status) && res.status !== 404;

      if (isExpected) {
        passCount++;
        console.log(`[PASS] ${item.method.padEnd(5)} ${item.path.padEnd(38)} -> HTTP ${res.status} (${elapsed}ms)`);
      } else {
        failCount++;
        failures.push({ path: item.path, status: res.status, expected: item.expected });
        console.error(`[FAIL] ${item.method.padEnd(5)} ${item.path.padEnd(38)} -> HTTP ${res.status} (Expected: ${item.expected.join("/")}) (${elapsed}ms)`);
      }
    } catch (err) {
      failCount++;
      failures.push({ path: item.path, error: err.message });
      console.error(`[ERR ] ${item.method.padEnd(5)} ${item.path.padEnd(38)} -> Connection error: ${err.message}`);
    }
  }

  console.log(`\n======================================================`);
  console.log(`  📊 AUDIT SUMMARY:`);
  console.log(`  Total Routes Tested : ${ENDPOINTS.length}`);
  console.log(`  Passed (OK / Not 404): ${passCount}`);
  console.log(`  Failed / 404 Errors : ${failCount}`);
  console.log(`======================================================\n`);

  if (failures.length > 0) {
    console.error("❌ Detected issues on routes:");
    console.table(failures);
    process.exit(1);
  } else {
    console.log("✅ ALL ROUTES AUDITED SUCCESSFULLY WITH ZERO 404 ERRORS!");
    process.exit(0);
  }
}

runAudit();

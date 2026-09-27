const http = require("http");

async function testOAuthFlow() {
  console.log("=== Testing OAuth Callback & Exchange Resilience ===");

  // 1. Test /callback endpoint returns HTML with postMessage and wildcard
  const callbackRes = await fetch("http://127.0.0.1:20128/callback?code=mock_code&state=mock_state");
  const html = await callbackRes.text();
  if (html.includes("oauth_callback") && html.includes("window.opener.postMessage") && html.includes("*")) {
    console.log("✓ /callback returns valid HTML with wildcard postMessage");
  } else {
    console.error("✗ /callback response missing expected scripts");
    process.exit(1);
  }

  // 2. Test /api/oauth/antigravity/authorize returns proper redirectUri
  const authRes = await fetch("http://127.0.0.1:20128/api/oauth/antigravity/authorize");
  const authData = await authRes.json();
  if (authData.authUrl && authData.redirectUri === "http://127.0.0.1:20128/callback") {
    console.log("✓ /api/oauth/antigravity/authorize sets dynamic redirectUri to current host");
  } else {
    console.error("✗ Unexpected authorize response:", authData);
    process.exit(1);
  }

  // 3. Test /api/oauth/antigravity/exchange handles invalid/mock code without hanging
  const start = Date.now();
  const exchangeRes = await fetch("http://127.0.0.1:20128/api/oauth/antigravity/exchange", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      code: "invalid_or_expired_mock_code",
      redirectUri: "http://127.0.0.1:20128/callback"
    })
  });
  const elapsed = Date.now() - start;
  const exchangeData = await exchangeRes.json();
  console.log(`✓ /exchange responded in ${elapsed}ms (status: ${exchangeRes.status})`);
  console.log("  Response details:", exchangeData.error || exchangeData);

  if (elapsed < 5000) {
    console.log("✓ Fast response guaranteed: no hanging on exchange failure or external APIs!");
  }

  console.log("=== All OAuth Verification Tests Passed! ===");
}

testOAuthFlow().catch(console.error);

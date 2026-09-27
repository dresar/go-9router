async function addBynara() {
  console.log("Adding Bynara provider connection...");

  // 1. Login
  const loginRes = await fetch("http://127.0.0.1:20128/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password: "admin1234" })
  });
  const cookies = loginRes.headers.getSetCookie();
  const sessionCookie = cookies.find(c => c.startsWith("9r_session=")).split(";")[0];

  // 2. Check if bynara connection already exists
  const listRes = await fetch("http://127.0.0.1:20128/api/providers/client?provider=bynara", {
    headers: { cookie: sessionCookie }
  });
  const listData = await listRes.json();
  const existing = listData.connections?.find(c => c.provider === "bynara");

  const payload = {
    provider: "bynara",
    authType: "apikey",
    name: "Bynara Free Tier",
    apiKey: "sk-nry-6AA_jExFvrRZBeDN8Pt3M5KIQ99Ar-jVTGATF_mzKo0",
    providerSpecificData: {
      baseUrl: "https://router.bynara.id/v1",
      baseURL: "https://router.bynara.id/v1"
    }
  };

  let connId = "";
  if (existing) {
    console.log("Updating existing Bynara connection:", existing.id);
    const updateRes = await fetch(`http://127.0.0.1:20128/api/providers/${existing.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", cookie: sessionCookie },
      body: JSON.stringify(payload)
    });
    const updateData = await updateRes.json();
    connId = existing.id;
    console.log("Updated:", updateData);
  } else {
    console.log("Creating new Bynara connection...");
    const createRes = await fetch("http://127.0.0.1:20128/api/providers", {
      method: "POST",
      headers: { "Content-Type": "application/json", cookie: sessionCookie },
      body: JSON.stringify(payload)
    });
    const createData = await createRes.json();
    connId = createData.connection?.id;
    console.log("Created connection:", connId);
  }

  // 3. Test connection
  console.log("Testing Bynara connection...");
  const testRes = await fetch(`http://127.0.0.1:20128/api/providers/${connId}/test`, {
    method: "POST",
    headers: { cookie: sessionCookie }
  });
  const testData = await testRes.json();
  console.log("Test result:", testData);

  // 4. Test chat completion proxy through Go gateway using a bynara model (e.g. mimo-v2.5-free or qwen3.8-flash)
  console.log("Testing chat proxy through /v1/chat/completions with model mimo-v2.5-free...");
  const chatRes = await fetch("http://127.0.0.1:20128/v1/chat/completions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      model: "bynara/mimo-v2.5-free",
      messages: [{ role: "user", content: "Reply with the exact word: PONG" }],
      max_tokens: 10
    })
  });
  console.log("Chat HTTP status:", chatRes.status);
  const chatText = await chatRes.text();
  console.log("Chat response:", chatText.slice(0, 300));
}

addBynara().catch(console.error);

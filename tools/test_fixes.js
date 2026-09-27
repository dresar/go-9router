const assert = require('assert');

async function testAll() {
  console.log('====================================================');
  console.log('     Go 9Router - Verification of All Fixes         ');
  console.log('====================================================\n');

  // 1. Authenticate with admin1234
  const loginRes = await fetch('http://127.0.0.1:20128/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: 'admin1234' })
  });
  const cookies = loginRes.headers.getSetCookie();
  const sessionCookie = cookies.find(c => c.startsWith('9r_session=')).split(';')[0];
  console.log('✓ [AUTH] Logged in successfully with session cookie\n');

  // 2. Test OAuth Authorize for Antigravity
  console.log('--- Test 1: Antigravity OAuth Authorize Endpoint ---');
  const oauthRes = await fetch('http://127.0.0.1:20128/api/oauth/antigravity/authorize?redirect_uri=http://localhost:20127/callback', {
    headers: { cookie: sessionCookie }
  });
  const oauthData = await oauthRes.json();
  console.log('Status:', oauthRes.status);
  console.log('Response:', JSON.stringify(oauthData, null, 2));
  assert(oauthRes.ok, 'OAuth authorize request failed');
  assert(oauthData.authUrl && oauthData.authUrl.includes('accounts.google.com'), 'Missing or invalid Google authUrl');
  assert(oauthData.state, 'Missing state parameter');
  assert.strictEqual(oauthData.flowType, 'authorization_code', 'flowType must be authorization_code');
  console.log('✅ Antigravity OAuth Authorize URL generated correctly! (Fixes "No authorization URL returned")\n');

  // 3. Test Connection Toggle (ON / OFF)
  console.log('--- Test 2: Quota Tracker Connection Toggle (ON / OFF) ---');
  const listRes = await fetch('http://127.0.0.1:20128/api/providers/client?pageSize=5', {
    headers: { cookie: sessionCookie }
  });
  const listData = await listRes.json();
  const testConn = listData.connections[0];
  console.log(`Target Connection: [${testConn.provider}] ${testConn.name || testConn.email} (ID: ${testConn.id}), Current isActive: ${testConn.isActive}`);

  // Turn OFF
  console.log('Turning connection OFF...');
  const turnOffRes = await fetch(`http://127.0.0.1:20128/api/providers/${testConn.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', cookie: sessionCookie },
    body: JSON.stringify({ isActive: false })
  });
  const turnOffData = await turnOffRes.json();
  console.log('Turned OFF response isActive:', turnOffData.connection.isActive);
  assert.strictEqual(turnOffData.connection.isActive, false, 'Connection did not turn OFF');

  // Verify DB state
  const verifyOffRes = await fetch(`http://127.0.0.1:20128/api/providers/${testConn.id}`, {
    headers: { cookie: sessionCookie }
  });
  const verifyOffData = await verifyOffRes.json();
  assert.strictEqual(verifyOffData.connection.isActive, false, 'Database does not reflect OFF state');
  console.log('✓ Successfully confirmed connection is saved as OFF in database');

  // Turn ON again
  console.log('Turning connection ON again...');
  const turnOnRes = await fetch(`http://127.0.0.1:20128/api/providers/${testConn.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', cookie: sessionCookie },
    body: JSON.stringify({ isActive: true })
  });
  const turnOnData = await turnOnRes.json();
  console.log('Turned ON response isActive:', turnOnData.connection.isActive);
  assert.strictEqual(turnOnData.connection.isActive, true, 'Connection did not turn ON');
  console.log('✅ Quota Tracker connection ON/OFF toggle verified and working!\n');

  // 4. Test Proxy Pool Test & Toggle
  console.log('--- Test 3: Proxy Pool Test & Toggle ---');
  const proxyListRes = await fetch('http://127.0.0.1:20128/api/proxy-pools', {
    headers: { cookie: sessionCookie }
  });
  const proxyListData = await proxyListRes.json();
  const testPool = proxyListData.proxyPools[0];
  console.log(`Target Proxy Pool: [${testPool.type}] ${testPool.name} (ID: ${testPool.id})`);

  // Test proxy pool
  console.log('Testing proxy pool via POST /api/proxy-pools/:id/test...');
  const testProxyRes = await fetch(`http://127.0.0.1:20128/api/proxy-pools/${testPool.id}/test`, {
    method: 'POST',
    headers: { cookie: sessionCookie }
  });
  const testProxyData = await testProxyRes.json();
  console.log('Proxy test result:', JSON.stringify(testProxyData));
  assert(testProxyRes.ok, 'Proxy pool test failed with HTTP error');
  assert(typeof testProxyData.ok === 'boolean', 'Proxy test response missing ok field');
  console.log(`✅ Proxy test executed successfully! ok=${testProxyData.ok}, status=${testProxyData.status}, elapsed=${testProxyData.elapsedMs}ms`);

  // Toggle proxy pool isActive
  console.log('Toggling proxy pool isActive to false...');
  const poolToggleRes = await fetch(`http://127.0.0.1:20128/api/proxy-pools/${testPool.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', cookie: sessionCookie },
    body: JSON.stringify({ isActive: false })
  });
  const poolToggleData = await poolToggleRes.json();
  assert.strictEqual(poolToggleData.pool.isActive, false, 'Proxy pool failed to toggle to false');
  // Revert back
  await fetch(`http://127.0.0.1:20128/api/proxy-pools/${testPool.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', cookie: sessionCookie },
    body: JSON.stringify({ isActive: true })
  });
  console.log('✅ Proxy pool toggle verified and reverted!\n');

  // 5. Test Provider Connection Probe
  console.log('--- Test 4: Provider Connection Test & Batch Test ---');
  const connTestRes = await fetch(`http://127.0.0.1:20128/api/providers/${testConn.id}/test`, {
    method: 'POST',
    headers: { cookie: sessionCookie }
  });
  const connTestData = await connTestRes.json();
  console.log('Provider connection test response:', JSON.stringify(connTestData));
  assert(connTestRes.ok, 'Provider test request failed');
  assert(typeof connTestData.valid === 'boolean', 'Missing valid boolean in test response');
  console.log('✅ Provider connection test endpoint verified!\n');

  // 6. Test Model Ping
  console.log('--- Test 5: Model Ping (/api/models/test) ---');
  const modelTestRes = await fetch('http://127.0.0.1:20128/api/models/test', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', cookie: sessionCookie },
    body: JSON.stringify({ model: 'ag/gemini-3.8-flash' })
  });
  const modelTestData = await modelTestRes.json();
  console.log('Model test response:', JSON.stringify(modelTestData));
  assert(modelTestRes.ok, 'Model test request failed');
  assert(modelTestData.ok === true || modelTestData.valid === true, 'Model test ok missing');
  console.log('✅ Model ping test verified!\n');

  // 7. Test Quota Usage Endpoint
  console.log('--- Test 6: Quota Usage Endpoint (/api/usage/:connectionId) ---');
  const usageRes = await fetch(`http://127.0.0.1:20128/api/usage/${testConn.id}`, {
    headers: { cookie: sessionCookie }
  });
  const usageData = await usageRes.json();
  console.log('Usage response keys:', Object.keys(usageData));
  assert(usageRes.ok, 'Usage request failed');
  assert(usageData.quotas && typeof usageData.quotas === 'object', 'Usage data missing quotas object');
  console.log('✅ Quota tracker endpoint verified with proper structure!\n');

  console.log('====================================================');
  console.log('      ALL 6 VERIFICATION CHECKS PASSED 100%!        ');
  console.log('====================================================\n');
}

testAll().catch(e => {
  console.error('\n❌ Verification Failed:', e);
  process.exit(1);
});

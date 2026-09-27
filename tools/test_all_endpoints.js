const http = require('http');

async function request(path, options = {}) {
  return new Promise((resolve, reject) => {
    const req = http.request(`http://127.0.0.1:20128${path}`, options, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        let json = null;
        try { json = JSON.parse(data); } catch (e) { json = data; }
        resolve({ status: res.statusCode, headers: res.headers, body: json });
      });
    });
    req.on('error', reject);
    if (options.body) {
      req.write(typeof options.body === 'string' ? options.body : JSON.stringify(options.body));
    }
    req.end();
  });
}

async function run() {
  console.log('Testing Go 9Router Endpoints...\n');

  // 1. Login
  const login = await request('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: { password: 'admin1234' }
  });
  console.log('1. Login:', login.status, login.body);
  const cookie = (login.headers['set-cookie'] || []).map(c => c.split(';')[0]).join('; ');

  const authHeaders = { 'Cookie': cookie, 'Content-Type': 'application/json' };

  // 2. /api/providers/client?sort=provider (The user's reported error!)
  const provClient = await request('/api/providers/client?sort=provider', { headers: authHeaders });
  console.log('2. /api/providers/client?sort=provider -> status:', provClient.status, {
    connectionsCount: provClient.body?.connections?.length,
    providerOptions: provClient.body?.providerOptions,
    pagination: provClient.body?.pagination
  });

  // 3. /api/providers/suggested-models
  const suggested = await request('/api/providers/suggested-models', { headers: authHeaders });
  console.log('3. /api/providers/suggested-models -> status:', suggested.status, suggested.body);

  // 4. /api/providers/test-batch
  const testBatch = await request('/api/providers/test-batch', {
    method: 'POST',
    headers: authHeaders,
    body: { mode: 'all' }
  });
  console.log('4. /api/providers/test-batch -> status:', testBatch.status, testBatch.body);

  // 5. /api/usage/request-logs
  const reqLogs = await request('/api/usage/request-logs', { headers: authHeaders });
  console.log('5. /api/usage/request-logs -> status:', reqLogs.status, Array.isArray(reqLogs.body) ? `count: ${reqLogs.body.length}` : reqLogs.body);

  // 6. /api/usage/chart?period=7d
  const chart = await request('/api/usage/chart?period=7d', { headers: authHeaders });
  console.log('6. /api/usage/chart?period=7d -> status:', chart.status, Array.isArray(chart.body) ? `buckets: ${chart.body.length}` : chart.body);

  // 7. /api/usage/providers
  const usageProviders = await request('/api/usage/providers', { headers: authHeaders });
  console.log('7. /api/usage/providers -> status:', usageProviders.status, usageProviders.body);

  // 8. /api/settings/require-login
  const reqLogin = await request('/api/settings/require-login');
  console.log('8. /api/settings/require-login -> status:', reqLogin.status, reqLogin.body);

  // 9. /api/tunnel/status
  const tunnelStatus = await request('/api/tunnel/status', { headers: authHeaders });
  console.log('9. /api/tunnel/status -> status:', tunnelStatus.status, {
    hasTunnelObj: !!tunnelStatus.body?.tunnel,
    hasTailscaleObj: !!tunnelStatus.body?.tailscale
  });

  // 10. /api/cli-tools/all-statuses
  const cliTools = await request('/api/cli-tools/all-statuses', { headers: authHeaders });
  console.log('10. /api/cli-tools/all-statuses -> status:', cliTools.status, {
    hasClaude: !!cliTools.body?.claude,
    hasCodex: !!cliTools.body?.codex
  });

  // 11. /api/combos/presets?source=cursor
  const combosPresets = await request('/api/combos/presets?source=cursor', { headers: authHeaders });
  console.log('11. /api/combos/presets?source=cursor -> status:', combosPresets.status, combosPresets.body);

  // 12. /api/models/availability
  const availability = await request('/api/models/availability', { headers: authHeaders });
  console.log('12. /api/models/availability -> status:', availability.status, availability.body);

  // 13. /api/models/disabled
  const disabled = await request('/api/models/disabled', { headers: authHeaders });
  console.log('13. /api/models/disabled -> status:', disabled.status, disabled.body);

  // 14. /api/oauth/cursor/auto-import
  const cursorAutoImport = await request('/api/oauth/cursor/auto-import', { headers: authHeaders });
  console.log('14. /api/oauth/cursor/auto-import -> status:', cursorAutoImport.status, cursorAutoImport.body);

  console.log('\nALL TESTS EXECUTED!');
}

run().catch(console.error);

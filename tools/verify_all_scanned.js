const http = require('http');
const endpoints = require('./frontend_endpoints.json');

async function request(path, cookie) {
  return new Promise((resolve) => {
    // Replace template variables like ${id}, ${connectionId} with dummy 'default'
    const cleanPath = path
      .replace(/\$\{[^}]+\}/g, 'default')
      .replace(/\[\.\.\.[^\]]+\]/g, 'default')
      .replace(/\[[^\]]+\]/g, 'default');

    let resolved = false;
    const req = http.request(`http://127.0.0.1:20128${cleanPath}`, {
      method: 'GET',
      timeout: 1500,
      headers: {
        'Cookie': cookie,
        'Content-Type': 'application/json'
      }
    }, (res) => {
      if (res.headers['content-type']?.includes('text/event-stream') || cleanPath.includes('stream')) {
        resolved = true;
        req.destroy();
        return resolve({ path: cleanPath, status: res.statusCode });
      }
      let data = '';
      res.on('data', chunk => { data += chunk; });
      res.on('end', () => {
        if (!resolved) {
          resolved = true;
          resolve({ path: cleanPath, status: res.statusCode });
        }
      });
    });

    req.on('timeout', () => {
      if (!resolved) {
        resolved = true;
        req.destroy();
        resolve({ path: cleanPath, status: 'TIMEOUT' });
      }
    });

    req.on('error', (err) => {
      if (!resolved) {
        resolved = true;
        resolve({ path: cleanPath, status: 'ERR', error: err.message });
      }
    });

    req.end();
  });
}

async function run() {
  // 1. Login
  const loginReq = http.request('http://127.0.0.1:20128/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' }
  });
  
  const cookie = await new Promise((resolve) => {
    loginReq.on('response', (res) => {
      const c = (res.headers['set-cookie'] || []).map(x => x.split(';')[0]).join('; ');
      resolve(c);
    });
    loginReq.write(JSON.stringify({ password: 'admin1234' }));
    loginReq.end();
  });

  console.log(`Auditing all ${endpoints.length} frontend endpoints against Go backend...`);
  const results = { ok200: 0, notAllowed405: 0, notFound404: 0, other: [] };

  for (const ep of endpoints) {
    const res = await request(ep, cookie);
    if (res.status === 200 || res.status === 201) {
      results.ok200++;
    } else if (res.status === 405) {
      // 405 means endpoint exists in Go but expects POST/PUT/DELETE, which is correct for action endpoints!
      results.notAllowed405++;
    } else if (res.status === 404) {
      results.notFound404++;
      results.other.push({ endpoint: ep, status: res.status });
    } else {
      results.other.push({ endpoint: ep, status: res.status });
    }
  }

  console.log('\nAudit Summary:');
  console.log('200 OK:', results.ok200);
  console.log('405 Method Not Allowed (Route exists, requires POST/PUT/DELETE):', results.notAllowed405);
  console.log('404 Not Found:', results.notFound404);
  console.log('Other statuses / Details:', results.other);
}

run().catch(console.error);

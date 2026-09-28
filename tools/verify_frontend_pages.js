// tools/verify_frontend_pages.js
const http = require('http');

const PAGES = [
  '/',
  '/login',
  '/landing',
  '/dashboard',
  '/endpoint',
  '/dashboard/endpoint',
  '/basic-chat',
  '/dashboard/basic-chat',
  '/cli-tools',
  '/dashboard/cli-tools',
  '/combos',
  '/dashboard/combos',
  '/console-log',
  '/dashboard/console-log',
  '/media-providers/web',
  '/dashboard/media-providers/web',
  '/mitm',
  '/dashboard/mitm',
  '/playground',
  '/dashboard/playground',
  '/profile',
  '/dashboard/profile',
  '/providers',
  '/dashboard/providers',
  '/providers/new',
  '/dashboard/providers/new',
  '/proxy-pools',
  '/dashboard/proxy-pools',
  '/pxpipe',
  '/dashboard/pxpipe',
  '/quota',
  '/dashboard/quota',
  '/skills',
  '/dashboard/skills',
  '/token-saver',
  '/dashboard/token-saver',
  '/translator',
  '/dashboard/translator',
  '/usage',
  '/dashboard/usage',
  '/dashboard/settings/pricing',
  '/callback'
];

async function checkPage(path) {
  return new Promise((resolve) => {
    const req = http.request(`http://127.0.0.1:20128${path}`, {
      method: 'GET',
      headers: {
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
      }
    }, (res) => {
      let body = '';
      res.on('data', chunk => {
        if (body.length < 2000) body += chunk;
      });
      res.on('end', () => {
        const isHtml = (res.headers['content-type'] || '').includes('text/html');
        const isRedirect = [301, 302, 307, 308].includes(res.statusCode);
        const ok = (res.statusCode === 200 && isHtml) || isRedirect;
        resolve({
          path,
          status: res.statusCode,
          location: res.headers.location || null,
          contentType: res.headers['content-type'],
          isHtml,
          ok,
          snippet: body.slice(0, 120).replace(/\s+/g, ' ')
        });
      });
    });

    req.on('error', (err) => {
      resolve({ path, ok: false, error: err.message });
    });

    req.end();
  });
}

async function run() {
  console.log('='.repeat(70));
  console.log('  🧪 GO-9ROUTER FRONTEND PAGES VERIFICATION');
  console.log('  Testing all page URLs on http://127.0.0.1:20128');
  console.log('='.repeat(70) + '\n');

  let passed = 0;
  let failed = 0;
  const failures = [];

  for (const p of PAGES) {
    const result = await checkPage(p);
    if (result.ok) {
      passed++;
      const redirectInfo = result.location ? `-> Redirect to ${result.location}` : '';
      console.log(`[PASS] ${p.padEnd(32)} HTTP ${result.status} ${redirectInfo}`);
    } else {
      failed++;
      failures.push(result);
      console.error(`[FAIL] ${p.padEnd(32)} HTTP ${result.status || 'ERR'} ${result.error || ''}`);
    }
  }

  console.log('\n' + '='.repeat(70));
  console.log(`  PAGE AUDIT SUMMARY:`);
  console.log(`  Total Pages Checked : ${PAGES.length}`);
  console.log(`  Passed              : ${passed}`);
  console.log(`  Failed              : ${failed}`);
  console.log('='.repeat(70) + '\n');

  if (failed > 0) {
    console.error('Failed pages:');
    console.table(failures);
    process.exit(1);
  } else {
    console.log('🎉 ALL FRONTEND PAGES RESPOND WITH 200 OK OR VALID REDIRECT!');
  }
}

run();

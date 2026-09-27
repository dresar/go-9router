const fs = require('fs');
const path = require('path');

function walk(dir, fileList = []) {
  try {
    const files = fs.readdirSync(dir);
    for (const file of files) {
      if (file === 'node_modules' || file === '.next' || file === '.git') continue;
      const fullPath = path.join(dir, file);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory()) {
        walk(fullPath, fileList);
      } else if (/\.(jsx?|tsx?)$/.test(file)) {
        fileList.push(fullPath);
      }
    }
  } catch (e) {}
  return fileList;
}

const frontendSrc = path.resolve(__dirname, '../frontend/src');
const allFiles = walk(frontendSrc);
const apiRegex = /[\"'`](\/api\/[a-zA-Z0-9_\-\/\$\{\}]+)[\"'`\?]/g;
const endpoints = new Set();
const usages = {};

for (const f of allFiles) {
  // skip src/app/api folder itself so we only see consumer calls
  if (f.includes(path.sep + 'src' + path.sep + 'app' + path.sep + 'api')) continue;

  const content = fs.readFileSync(f, 'utf8');
  let m;
  while ((m = apiRegex.exec(content)) !== null) {
    const ep = m[1];
    endpoints.add(ep);
    if (!usages[ep]) usages[ep] = [];
    const rel = path.relative(frontendSrc, f).replace(/\\/g, '/');
    if (!usages[ep].includes(rel)) usages[ep].push(rel);
  }
}

const sorted = Array.from(endpoints).sort();
fs.writeFileSync(path.resolve(__dirname, 'frontend_endpoints.json'), JSON.stringify(sorted, null, 2));
console.log('Saved', sorted.length, 'endpoints to tools/frontend_endpoints.json');

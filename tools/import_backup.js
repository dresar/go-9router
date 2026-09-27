const fs = require('fs');
const path = require('path');
const Database = require(path.resolve(__dirname, '../frontend/node_modules/better-sqlite3'));

const backupPath = path.resolve(__dirname, '../9router-backup-2026-09-20T15-57-15-731Z.json');
const dbPath = path.resolve(__dirname, '../data/db.sqlite');

console.log('Reading backup from:', backupPath);
const payload = JSON.parse(fs.readFileSync(backupPath, 'utf8'));

console.log('Connecting to database:', dbPath);
const db = new Database(dbPath);
db.pragma('journal_mode = WAL');

const importTransaction = db.transaction(() => {
  console.log('Clearing existing data...');
  db.prepare(`DELETE FROM settings`).run();
  db.prepare(`DELETE FROM providerConnections`).run();
  db.prepare(`DELETE FROM providerNodes`).run();
  db.prepare(`DELETE FROM proxyPools`).run();
  db.prepare(`DELETE FROM apiKeys`).run();
  db.prepare(`DELETE FROM combos`).run();
  db.prepare(`DELETE FROM kv WHERE scope IN ('modelAliases', 'customModels', 'mitmAlias', 'pricing')`).run();

  // 1. Settings
  if (payload.settings) {
    db.prepare(`INSERT OR REPLACE INTO settings(id, data) VALUES(1, ?)`).run(JSON.stringify(payload.settings));
    console.log('Imported settings: OK');
  }

  // 2. Provider Connections
  const insertConn = db.prepare(`
    INSERT OR REPLACE INTO providerConnections(id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  let connCount = 0;
  for (const c of payload.providerConnections || []) {
    const { id, provider, authType, name, email, priority, isActive, createdAt, updatedAt, ...rest } = c;
    insertConn.run(
      id,
      provider,
      authType || 'oauth',
      name || null,
      email || null,
      priority !== undefined ? priority : 50,
      isActive === false ? 0 : 1,
      JSON.stringify(rest),
      createdAt || new Date().toISOString(),
      updatedAt || new Date().toISOString()
    );
    connCount++;
  }
  console.log(`Imported providerConnections: ${connCount}`);

  // 3. Provider Nodes
  const insertNode = db.prepare(`
    INSERT OR REPLACE INTO providerNodes(id, type, name, data, createdAt, updatedAt)
    VALUES(?, ?, ?, ?, ?, ?)
  `);
  let nodeCount = 0;
  for (const n of payload.providerNodes || []) {
    const { id, type, name, createdAt, updatedAt, ...rest } = n;
    insertNode.run(
      id,
      type || null,
      name || null,
      JSON.stringify(rest),
      createdAt || new Date().toISOString(),
      updatedAt || new Date().toISOString()
    );
    nodeCount++;
  }
  console.log(`Imported providerNodes: ${nodeCount}`);

  // 4. Proxy Pools
  const insertPool = db.prepare(`
    INSERT OR REPLACE INTO proxyPools(id, isActive, testStatus, data, createdAt, updatedAt)
    VALUES(?, ?, ?, ?, ?, ?)
  `);
  let poolCount = 0;
  for (const p of payload.proxyPools || []) {
    const { id, isActive, testStatus, createdAt, updatedAt, ...rest } = p;
    insertPool.run(
      id,
      isActive === false ? 0 : 1,
      testStatus || 'unknown',
      JSON.stringify(rest),
      createdAt || new Date().toISOString(),
      updatedAt || new Date().toISOString()
    );
    poolCount++;
  }
  console.log(`Imported proxyPools: ${poolCount}`);

  // 5. API Keys
  const insertKey = db.prepare(`
    INSERT OR REPLACE INTO apiKeys(id, key, name, machineId, isActive, createdAt)
    VALUES(?, ?, ?, ?, ?, ?)
  `);
  let keyCount = 0;
  for (const k of payload.apiKeys || []) {
    insertKey.run(
      k.id,
      k.key,
      k.name || null,
      k.machineId || null,
      k.isActive === false ? 0 : 1,
      k.createdAt || new Date().toISOString()
    );
    keyCount++;
  }
  console.log(`Imported apiKeys: ${keyCount}`);

  // 6. Combos
  const insertCombo = db.prepare(`
    INSERT OR REPLACE INTO combos(id, name, kind, models, createdAt, updatedAt)
    VALUES(?, ?, ?, ?, ?, ?)
  `);
  let comboCount = 0;
  for (const c of payload.combos || []) {
    insertCombo.run(
      c.id,
      c.name,
      c.kind || null,
      JSON.stringify(c.models || []),
      c.createdAt || new Date().toISOString(),
      c.updatedAt || new Date().toISOString()
    );
    comboCount++;
  }
  console.log(`Imported combos: ${comboCount}`);

  // 7. Custom Models
  const insertKV = db.prepare(`INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?, ?, ?)`);
  let customModelCount = 0;
  for (const m of payload.customModels || []) {
    const k = `${m.providerAlias || m.provider}|${m.id}|${m.type || 'llm'}`;
    insertKV.run('customModels', k, JSON.stringify(m));
    customModelCount++;
  }
  console.log(`Imported customModels: ${customModelCount}`);

  // 8. Model Aliases
  let aliasCount = 0;
  for (const [a, m] of Object.entries(payload.modelAliases || {})) {
    insertKV.run('modelAliases', a, JSON.stringify(m));
    aliasCount++;
  }
  if (aliasCount > 0) console.log(`Imported modelAliases: ${aliasCount}`);

  // 9. MITM Aliases
  let mitmCount = 0;
  for (const [tool, mappings] of Object.entries(payload.mitmAlias || {})) {
    insertKV.run('mitmAlias', tool, JSON.stringify(mappings || {}));
    mitmCount++;
  }
  if (mitmCount > 0) console.log(`Imported mitmAlias: ${mitmCount}`);

  // 10. Pricing
  let pricingCount = 0;
  for (const [provider, models] of Object.entries(payload.pricing || {})) {
    insertKV.run('pricing', provider, JSON.stringify(models));
    pricingCount++;
  }
  if (pricingCount > 0) console.log(`Imported pricing: ${pricingCount}`);
});

try {
  importTransaction();
  console.log('\n========================================');
  console.log('SUCCESS: All backup data imported cleanly!');
  console.log('========================================');
} catch (err) {
  console.error('Import failed:', err);
  process.exit(1);
} finally {
  db.close();
}

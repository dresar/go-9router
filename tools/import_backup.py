import json
import sqlite3
import os
import sys

def import_backup(backup_file, db_file):
    print(f"Loading backup from: {backup_file}")
    with open(backup_file, 'r', encoding='utf-8') as f:
        payload = json.load(f)

    os.makedirs(os.path.dirname(os.path.abspath(db_file)), exist_ok=True)
    conn = sqlite3.connect(db_file)
    conn.execute("PRAGMA journal_mode = WAL")
    cur = conn.cursor()

    # Create tables if not exists
    migrations = [
        "CREATE TABLE IF NOT EXISTS _meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)",
        "CREATE TABLE IF NOT EXISTS settings (id INTEGER PRIMARY KEY CHECK (id = 1), data TEXT NOT NULL DEFAULT '{}')",
        """CREATE TABLE IF NOT EXISTS providerConnections (
            id TEXT PRIMARY KEY, provider TEXT NOT NULL, authType TEXT NOT NULL,
            name TEXT, email TEXT, priority INTEGER, isActive INTEGER DEFAULT 1,
            data TEXT NOT NULL DEFAULT '{}', createdAt TEXT NOT NULL, updatedAt TEXT NOT NULL
        )""",
        "CREATE INDEX IF NOT EXISTS idx_pc_provider ON providerConnections(provider)",
        "CREATE INDEX IF NOT EXISTS idx_pc_provider_active ON providerConnections(provider, isActive)",
        "CREATE INDEX IF NOT EXISTS idx_pc_priority ON providerConnections(provider, priority)",
        """CREATE TABLE IF NOT EXISTS providerNodes (
            id TEXT PRIMARY KEY, type TEXT, name TEXT, data TEXT NOT NULL DEFAULT '{}',
            createdAt TEXT NOT NULL, updatedAt TEXT NOT NULL
        )""",
        "CREATE INDEX IF NOT EXISTS idx_pn_type ON providerNodes(type)",
        """CREATE TABLE IF NOT EXISTS proxyPools (
            id TEXT PRIMARY KEY, isActive INTEGER DEFAULT 1, testStatus TEXT,
            data TEXT NOT NULL DEFAULT '{}', createdAt TEXT NOT NULL, updatedAt TEXT NOT NULL
        )""",
        "CREATE INDEX IF NOT EXISTS idx_pp_active ON proxyPools(isActive)",
        """CREATE TABLE IF NOT EXISTS apiKeys (
            id TEXT PRIMARY KEY, key TEXT UNIQUE NOT NULL, name TEXT,
            machineId TEXT, isActive INTEGER DEFAULT 1, createdAt TEXT NOT NULL
        )""",
        "CREATE INDEX IF NOT EXISTS idx_ak_key ON apiKeys(key)",
        """CREATE TABLE IF NOT EXISTS combos (
            id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, kind TEXT,
            models TEXT NOT NULL DEFAULT '[]', createdAt TEXT NOT NULL, updatedAt TEXT NOT NULL
        )""",
        "CREATE INDEX IF NOT EXISTS idx_combo_name ON combos(name)",
        "CREATE TABLE IF NOT EXISTS kv (scope TEXT NOT NULL, key TEXT NOT NULL, value TEXT NOT NULL, PRIMARY KEY (scope, key))",
        "CREATE INDEX IF NOT EXISTS idx_kv_scope ON kv(scope)",
        """CREATE TABLE IF NOT EXISTS usageHistory (
            id INTEGER PRIMARY KEY AUTOINCREMENT, timestamp TEXT NOT NULL, provider TEXT,
            model TEXT, connectionId TEXT, apiKey TEXT, endpoint TEXT, promptTokens INTEGER DEFAULT 0,
            completionTokens INTEGER DEFAULT 0, cost REAL DEFAULT 0, status TEXT, tokens TEXT, meta TEXT
        )""",
        "CREATE TABLE IF NOT EXISTS usageDaily (dateKey TEXT PRIMARY KEY, data TEXT NOT NULL DEFAULT '{}')",
        """CREATE TABLE IF NOT EXISTS requestDetails (
            id TEXT PRIMARY KEY, timestamp TEXT NOT NULL, provider TEXT, model TEXT,
            connectionId TEXT, status TEXT, data TEXT NOT NULL DEFAULT '{}'
        )"""
    ]
    for m in migrations:
        cur.execute(m)

    # Clear existing
    for table in ['settings', 'providerConnections', 'providerNodes', 'proxyPools', 'apiKeys', 'combos']:
        cur.execute(f"DELETE FROM {table}")
    cur.execute("DELETE FROM kv WHERE scope IN ('modelAliases', 'customModels', 'mitmAlias', 'pricing')")

    # 1. Settings
    if 'settings' in payload:
        cur.execute("INSERT OR REPLACE INTO settings(id, data) VALUES(1, ?)", (json.dumps(payload['settings']),))
        print("Imported settings: OK")

    # 2. Provider Connections
    conn_count = 0
    for c in payload.get('providerConnections', []):
        c_copy = dict(c)
        cid = c_copy.pop('id')
        provider = c_copy.pop('provider')
        auth_type = c_copy.pop('authType', 'oauth')
        name = c_copy.pop('name', None)
        email = c_copy.pop('email', None)
        priority = c_copy.pop('priority', 50)
        is_active = 0 if c_copy.pop('isActive', True) is False else 1
        created_at = c_copy.pop('createdAt', '')
        updated_at = c_copy.pop('updatedAt', '')
        cur.execute("""
            INSERT OR REPLACE INTO providerConnections(id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt)
            VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """, (cid, provider, auth_type, name, email, priority, is_active, json.dumps(c_copy), created_at, updated_at))
        conn_count += 1
    print(f"Imported providerConnections: {conn_count}")

    # 3. Provider Nodes
    node_count = 0
    for n in payload.get('providerNodes', []):
        n_copy = dict(n)
        nid = n_copy.pop('id')
        ntype = n_copy.pop('type', None)
        name = n_copy.pop('name', None)
        created_at = n_copy.pop('createdAt', '')
        updated_at = n_copy.pop('updatedAt', '')
        cur.execute("""
            INSERT OR REPLACE INTO providerNodes(id, type, name, data, createdAt, updatedAt)
            VALUES(?, ?, ?, ?, ?, ?)
        """, (nid, ntype, name, json.dumps(n_copy), created_at, updated_at))
        node_count += 1
    print(f"Imported providerNodes: {node_count}")

    # 4. Proxy Pools
    pool_count = 0
    for p in payload.get('proxyPools', []):
        p_copy = dict(p)
        pid = p_copy.pop('id')
        is_active = 0 if p_copy.pop('isActive', True) is False else 1
        test_status = p_copy.pop('testStatus', 'unknown')
        created_at = p_copy.pop('createdAt', '')
        updated_at = p_copy.pop('updatedAt', '')
        cur.execute("""
            INSERT OR REPLACE INTO proxyPools(id, isActive, testStatus, data, createdAt, updatedAt)
            VALUES(?, ?, ?, ?, ?, ?)
        """, (pid, is_active, test_status, json.dumps(p_copy), created_at, updated_at))
        pool_count += 1
    print(f"Imported proxyPools: {pool_count}")

    # 5. API Keys
    key_count = 0
    for k in payload.get('apiKeys', []):
        cur.execute("""
            INSERT OR REPLACE INTO apiKeys(id, key, name, machineId, isActive, createdAt)
            VALUES(?, ?, ?, ?, ?, ?)
        """, (k['id'], k['key'], k.get('name'), k.get('machineId'), 0 if k.get('isActive') is False else 1, k.get('createdAt', '')))
        key_count += 1
    print(f"Imported apiKeys: {key_count}")

    # 6. Combos
    combo_count = 0
    for c in payload.get('combos', []):
        cur.execute("""
            INSERT OR REPLACE INTO combos(id, name, kind, models, createdAt, updatedAt)
            VALUES(?, ?, ?, ?, ?, ?)
        """, (c['id'], c['name'], c.get('kind'), json.dumps(c.get('models', [])), c.get('createdAt', ''), c.get('updatedAt', '')))
        combo_count += 1
    print(f"Imported combos: {combo_count}")

    # 7. Custom Models
    cm_count = 0
    for m in payload.get('customModels', []):
        k = f"{m.get('providerAlias') or m.get('provider')}|{m.get('id')}|{m.get('type', 'llm')}"
        cur.execute("INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?, ?, ?)", ('customModels', k, json.dumps(m)))
        cm_count += 1
    print(f"Imported customModels: {cm_count}")

    # 8. Model Aliases
    for a, m in payload.get('modelAliases', {}).items():
        cur.execute("INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?, ?, ?)", ('modelAliases', a, json.dumps(m)))

    # 9. MITM Aliases
    for tool, mappings in payload.get('mitmAlias', {}).items():
        cur.execute("INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?, ?, ?)", ('mitmAlias', tool, json.dumps(mappings)))

    # 10. Pricing
    for prov, models in payload.get('pricing', {}).items():
        cur.execute("INSERT OR REPLACE INTO kv(scope, key, value) VALUES(?, ?, ?)", ('pricing', prov, json.dumps(models)))

    conn.commit()
    conn.close()
    print("ALL BACKUP DATA IMPORTED CLEANLY!")

if __name__ == '__main__':
    backup_file = sys.argv[1] if len(sys.argv) > 1 else 'backup.json'
    db_file = sys.argv[2] if len(sys.argv) > 2 else 'data/db.sqlite'
    import_backup(backup_file, db_file)

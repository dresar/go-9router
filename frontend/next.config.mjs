import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const projectRoot = dirname(fileURLToPath(import.meta.url));
// CLI bundling needs workspace root so tracing includes hoisted node_modules (slim ~50MB).
// Docker / default uses projectRoot so server.js lands at /app/server.js (not nested).
const tracingRoot = process.env.NEXT_TRACING_ROOT_MODE === "workspace"
  ? join(projectRoot, "..")
  : projectRoot;
const proxyClientMaxBodySize = process.env.NINEROUTER_PROXY_CLIENT_MAX_BODY_SIZE || "128mb";

/** @type {import('next').NextConfig} */
const nextConfig = {
  distDir: process.env.NEXT_DIST_DIR || ".next",
  output: "standalone",
  allowedDevOrigins: [
    "127.0.0.1",
    "localhost",
    "127.0.0.1:20128",
    "localhost:20128",
    "127.0.0.1:20127",
    "localhost:20127"
  ],
  // `open` must stay external. It derives its own directory from `import.meta.url`, and
  // webpack replaces that with the absolute path of the BUILD machine as a string literal.
  // A release built on macOS therefore ships `file:///Users/.../open/index.js`, which
  // `fileURLToPath` rejects on Windows ("File URL path must be absolute" — no drive
  // letter). That throw happens at module scope, so every consumer of `open` dies on
  // import — including xAI/Grok token refresh, which loads the OAuth service that imports
  // it. Keeping it external preserves the real `import.meta.url` at runtime.
  serverExternalPackages: ["better-sqlite3", "sql.js", "node:sqlite", "bun:sqlite", "open"],
  turbopack: {
    root: tracingRoot
  },
  outputFileTracingRoot: tracingRoot,
  outputFileTracingExcludes: {
    "*": ["./gitbook/**/*"]
  },
  images: {
    unoptimized: true
  },
  env: {},
  experimental: {
    // #1529/#1572: LLM clients can send long context or base64 image payloads through /v1 rewrites.
    proxyClientMaxBodySize,
    // Cache fetch responses across HMR refreshes for faster dev reloads.
    serverComponentsHmrCache: true,
    // Tree-shake heavy barrel imports to cut compile + bundle size
    optimizePackageImports: ["@xyflow/react", "@dnd-kit/core", "@dnd-kit/sortable", "material-symbols", "marked"],
  },
  webpack: (config, { isServer }) => {
    // Ignore fs/path modules in browser bundle
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        path: false,
      };
    }
    // Exclude non-source dirs from watcher to reduce inotify load
    config.watchOptions = {
      ...config.watchOptions,
      aggregateTimeout: 300,
      ignored: /[\\/](node_modules|\.git|logs|\.next|\.next-cli-build|gitbook|cli|open-sse\.old|tests|docs)[\\/]/,
    };
    return config;
  },
  async redirects() {
    return [
      {
        source: "/dashboard",
        destination: "/endpoint",
        permanent: false,
      },
      {
        source: "/dashboard/:path((?!chat).*)",
        destination: "/:path",
        permanent: false,
      },
    ];
  },
  async rewrites() {
    const goBackend = process.env.GO_BACKEND_URL || "http://127.0.0.1:20128";
    return [
      {
        source: "/api/:path*",
        destination: `${goBackend}/api/:path*`
      },
      {
        source: "/v1/:path*",
        destination: `${goBackend}/v1/:path*`
      },
      {
        source: "/responses",
        destination: `${goBackend}/responses`
      },
      {
        source: "/codex/:path*",
        destination: `${goBackend}/codex/:path*`
      },
      {
        source: "/skills",
        destination: "/dashboard/skills"
      },
      {
        source: "/skills/:path*",
        destination: "/dashboard/skills/:path*"
      },
      {
        source: "/providers",
        destination: "/dashboard/providers"
      },
      {
        source: "/providers/:path*",
        destination: "/dashboard/providers/:path*"
      },
      {
        source: "/combos",
        destination: "/dashboard/combos"
      },
      {
        source: "/combos/:path*",
        destination: "/dashboard/combos/:path*"
      },
      {
        source: "/proxy-pools",
        destination: "/dashboard/proxy-pools"
      },
      {
        source: "/proxy-pools/:path*",
        destination: "/dashboard/proxy-pools/:path*"
      },
      {
        source: "/endpoint",
        destination: "/dashboard/endpoint"
      },
      {
        source: "/endpoint/:path*",
        destination: "/dashboard/endpoint/:path*"
      },
      {
        source: "/usage",
        destination: "/dashboard/usage"
      },
      {
        source: "/usage/:path*",
        destination: "/dashboard/usage/:path*"
      },
      {
        source: "/quota",
        destination: "/dashboard/quota"
      },
      {
        source: "/quota/:path*",
        destination: "/dashboard/quota/:path*"
      },
      {
        source: "/profile",
        destination: "/dashboard/profile"
      },
      {
        source: "/profile/:path*",
        destination: "/dashboard/profile/:path*"
      },
      {
        source: "/console-log",
        destination: "/dashboard/console-log"
      },
      {
        source: "/console-log/:path*",
        destination: "/dashboard/console-log/:path*"
      },
      {
        source: "/cli-tools",
        destination: "/dashboard/cli-tools"
      },
      {
        source: "/cli-tools/:path*",
        destination: "/dashboard/cli-tools/:path*"
      },
      {
        source: "/token-saver",
        destination: "/dashboard/token-saver"
      },
      {
        source: "/token-saver/:path*",
        destination: "/dashboard/token-saver/:path*"
      },
      {
        source: "/translator",
        destination: "/dashboard/translator"
      },
      {
        source: "/translator/:path*",
        destination: "/dashboard/translator/:path*"
      },
      {
        source: "/media-providers",
        destination: "/dashboard/media-providers"
      },
      {
        source: "/media-providers/:path*",
        destination: "/dashboard/media-providers/:path*"
      },
      {
        source: "/basic-chat",
        destination: "/dashboard/basic-chat"
      },
      {
        source: "/mitm",
        destination: "/dashboard/mitm"
      },
      {
        source: "/pxpipe",
        destination: "/dashboard/pxpipe"
      }
    ];
  }
};

export default nextConfig;

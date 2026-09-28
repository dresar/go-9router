import https from "https";
import http from "http";
import pkg from "../../../../package.json" with { type: "json" };

const GITHUB_REPO = "dresar/go-9router";
const VERSION_CACHE_TTL_MS = 3600000; // cache for 1h

// Survive hot reload; one cache per process
const versionCache = (global.__versionCache ??= { value: null, fetchedAt: 0 });

// Try querying Go gateway first (which runs locally on 20128)
function fetchFromGoGateway() {
  return new Promise((resolve) => {
    const req = http.get("http://127.0.0.1:20128/api/version", { timeout: 1500 }, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          const parsed = JSON.parse(data);
          if (parsed && parsed.currentCommit) {
            resolve(parsed);
            return;
          }
          resolve(null);
        } catch {
          resolve(null);
        }
      });
    });
    req.on("error", () => resolve(null));
    req.on("timeout", () => { req.destroy(); resolve(null); });
  });
}

// Fetch latest master commit from GitHub API
function fetchLatestGitHubCommit() {
  return new Promise((resolve) => {
    const headers = {
      "User-Agent": "9Router-Frontend"
    };
    if (process.env.GITHUB_TOKEN) {
      headers["Authorization"] = "token " + process.env.GITHUB_TOKEN;
    }
    const req = https.get(
      `https://api.github.com/repos/${GITHUB_REPO}/commits/master`,
      {
        timeout: 5000,
        headers
      },
      (res) => {
        let data = "";
        res.on("data", (chunk) => (data += chunk));
        res.on("end", () => {
          try {
            const parsed = JSON.parse(data);
            if (parsed && parsed.sha) {
              resolve({
                latestCommit: parsed.sha.slice(0, 7),
                latestCommitMsg: (parsed.commit?.message || "").split("\n")[0],
                latestCommitDate: parsed.commit?.committer?.date || ""
              });
              return;
            }
            resolve(null);
          } catch {
            resolve(null);
          }
        });
      }
    );
    req.on("error", () => resolve(null));
    req.on("timeout", () => { req.destroy(); resolve(null); });
  });
}

export async function GET() {
  // First attempt: Go gateway (which has live local git state)
  const goStatus = await fetchFromGoGateway();
  if (goStatus) {
    return Response.json(goStatus);
  }

  // Fallback: GitHub API with 1h cache
  let githubData = versionCache.value;
  if (!githubData || Date.now() - versionCache.fetchedAt >= VERSION_CACHE_TTL_MS) {
    githubData = await fetchLatestGitHubCommit();
    if (githubData) {
      versionCache.value = githubData;
      versionCache.fetchedAt = Date.now();
    }
  }

  const currentVersion = pkg.version;
  return Response.json({
    version: currentVersion,
    currentVersion,
    latestVersion: githubData?.latestCommit || currentVersion,
    hasUpdate: false,
    latestCommit: githubData?.latestCommit || "",
    latestCommitMsg: githubData?.latestCommitMsg || "",
    latestCommitDate: githubData?.latestCommitDate || "",
    repoUrl: `https://github.com/${GITHUB_REPO}`
  });
}


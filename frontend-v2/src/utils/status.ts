import { Connection } from "../types"

export function isConnectionInCooldown(c: Connection): boolean {
  if (c.rateLimitedUntil) {
    const until = new Date(c.rateLimitedUntil).getTime()
    if (!isNaN(until) && until > Date.now()) return true
  }

  for (const k of Object.keys(c)) {
    if (k.startsWith("modelLock_")) {
      const lockVal = (c as any)[k]
      if (typeof lockVal === "string") {
        const lockTime = new Date(lockVal).getTime()
        if (!isNaN(lockTime) && lockTime > Date.now()) return true
      }
    }
  }

  if (c.errorCode === 429) return true

  if (c.lastError) {
    const err = c.lastError.toLowerCase()
    if (err.includes("429") || err.includes("rate limit") || err.includes("quota") || err.includes("too many requests")) {
      return true
    }
  }

  return false
}

export function isConnectionError(c: Connection): boolean {
  if (isConnectionInCooldown(c)) return false
  if (c.testStatus === "error") return true
  if (c.errorCode && c.errorCode >= 400) return true
  if (c.lastError && c.lastError.trim().length > 0) return true
  return false
}

export function getProviderStatus(conns: Connection[]): {
  status: "ready" | "cooldown" | "error" | "inactive"
  label: string
  errorCount: number
} {
  const active = conns.filter(
    (c) =>
      c.isActive &&
      c.testStatus !== "unavailable" &&
      c.testStatus !== "error" &&
      (c.errorCode === 0 || c.errorCode === null || c.errorCode === undefined)
  )
  const cooldown = conns.filter((c) => isConnectionInCooldown(c))
  const errors = conns.filter((c) => isConnectionError(c))

  if (cooldown.length > 0) {
    return { status: "cooldown", label: "⏱ Cooldown", errorCount: errors.length }
  }
  if (active.length > 0) {
    return { status: "ready", label: "⊙ Ready", errorCount: errors.length }
  }
  if (errors.length > 0) {
    return { status: "error", label: `△ ${errors.length} Isu`, errorCount: errors.length }
  }
  return { status: "inactive", label: "○ Inactive", errorCount: 0 }
}

export function formatProviderName(raw: string): string {
  const map: Record<string, string> = {
    gemini: "Gemini",
    groq: "Groq",
    openrouter: "Openrouter",
    node_apinex: "Node_apinex",
    "codebuddy-intl": "Codebuddy-Intl",
    qoder: "Qoder",
    github: "Github",
    "openai-compatible-chat-tabitoken": "Openai-Compatible-Chat-Tabitoken",
    "gemini-cli": "Gemini-Cli",
    ollama: "Ollama",
    bynara: "Bynara",
    "vercel-ai-gateway": "Vercel-Ai-Gateway",
    "api-airforce": "API Airforce",
    node_b_ai_api: "B AI API",
    node_dahl_global: "Dahl Global",
    node_9inference_cloud: "9inference Cloud",
    node_openagentic_id: "Openagentic ID",
    "grok-cli": "Grok CLI",
    "opencode-go": "OpenCode Go",
    "xiaomi-tokenplan": "Xiaomi TokenPlan",
    "cloudflare-ai": "Cloudflare AI",
    clinepass: "ClinePass",
    "kilo-gateway": "Kilo Gateway"
  }
  return map[raw] || raw.charAt(0).toUpperCase() + raw.slice(1)
}

export function getProviderColor(raw: string): string {
  const map: Record<string, string> = {
    gemini: "#1a4fc4",
    groq: "#f59e0b",
    openrouter: "#7c3aed",
    github: "#333333",
    mistral: "#ff6b35",
    bynara: "#10b981",
    "cloudflare-ai": "#f48120",
    qoder: "#06b6d4"
  }
  return map[raw.toLowerCase()] || "#27272a"
}

export function maskIdentity(email?: string, name?: string): string {
  const target = email || name || ""
  if (!target) return "-"
  if (!target.includes("@")) {
    if (target.length <= 4) return target
    return `${target.slice(0, 3)}***`
  }
  const [user, domain] = target.split("@")
  if (user.length <= 3) return `${user}***@${domain}`
  return `${user.slice(0, 3)}***@${domain}`
}

export function formatTimeAgo(dateStr?: string): string {
  if (!dateStr) return "-"
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const diffSec = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diffSec < 60) return "Baru saja"
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin}m lalu`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `${diffHours}j lalu`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 30) return `${diffDays}h lalu`
  return d.toLocaleDateString()
}

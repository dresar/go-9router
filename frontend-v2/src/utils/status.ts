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

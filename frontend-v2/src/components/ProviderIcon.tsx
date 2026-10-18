import React from "react"

interface ProviderIconProps {
  provider: string
  className?: string
}

export const ProviderIcon: React.FC<ProviderIconProps> = ({ provider, className = "h-4 w-4" }) => {
  const p = (provider || "").toLowerCase()

  if (p === "gemini" || p === "gemini-cli") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
          d="M12 2C12 7.52285 7.52285 12 2 12C7.52285 12 12 16.4772 12 22C12 16.4772 16.4772 12 22 12C16.4772 12 12 7.52285 12 2Z"
          fill="url(#gemini-grad)"
        />
        <defs>
          <linearGradient id="gemini-grad" x1="2" y1="2" x2="22" y2="22" gradientUnits="userSpaceOnUse">
            <stop stopColor="#38bdf8" />
            <stop offset="0.5" stopColor="#818cf8" />
            <stop offset="1" stopColor="#c084fc" />
          </linearGradient>
        </defs>
      </svg>
    )
  }

  if (p === "groq") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="9" stroke="#f97316" strokeWidth="2.5" />
        <path d="M12 7V17M7 12H17" stroke="#f97316" strokeWidth="2.5" strokeLinecap="round" />
      </svg>
    )
  }

  if (p === "openrouter") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="3" y="3" width="8" height="8" rx="2" fill="#3b82f6" />
        <rect x="13" y="3" width="8" height="8" rx="2" fill="#60a5fa" />
        <rect x="3" y="13" width="8" height="8" rx="2" fill="#2563eb" />
        <rect x="13" y="13" width="8" height="8" rx="2" fill="#93c5fd" />
      </svg>
    )
  }

  if (p === "bynara") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M4 6L12 2L20 6V18L12 22L4 18V6Z" stroke="#10b981" strokeWidth="2" strokeLinejoin="round" />
        <path d="M12 2V22M4 6L20 18M20 6L4 18" stroke="#10b981" strokeWidth="1.5" strokeOpacity="0.6" />
      </svg>
    )
  }

  if (p === "github") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="currentColor">
        <path
          fillRule="evenodd"
          clipRule="evenodd"
          d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
        />
      </svg>
    )
  }

  if (p === "cloudflare-ai" || p.includes("cloudflare")) {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
          d="M19.35 10.04C18.67 6.59 15.64 4 12 4C9.11 4 6.6 5.64 5.35 8.04C2.34 8.36 0 10.91 0 14C0 17.31 2.69 20 6 20H19C21.76 20 24 17.76 24 15C24 12.36 21.95 10.22 19.35 10.04Z"
          fill="#f97316"
        />
      </svg>
    )
  }

  if (p === "sambanova") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M4 14C8 10 16 10 20 14M4 10C8 6 16 6 20 10M4 18C8 14 16 14 20 18" stroke="#ea580c" strokeWidth="2.5" strokeLinecap="round" />
      </svg>
    )
  }

  if (p === "cerebras") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="4" y="4" width="16" height="16" rx="4" stroke="#e11d48" strokeWidth="2" />
        <circle cx="12" cy="12" r="3" fill="#e11d48" />
        <path d="M12 2V4M12 20V22M2 12H4M20 12H22" stroke="#e11d48" strokeWidth="2" strokeLinecap="round" />
      </svg>
    )
  }

  if (p === "mistral") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M3 4H7V8H11V12H13V8H17V4H21V20H17V12H15V16H9V12H7V20H3V4Z" fill="#f59e0b" />
      </svg>
    )
  }

  if (p === "qoder") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M7 8L3 12L7 16M17 8L21 12L17 16M14 4L10 20" stroke="#06b6d4" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    )
  }

  if (p === "cline" || p === "clinepass") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="3" y="5" width="18" height="14" rx="3" stroke="#10b981" strokeWidth="2" />
        <circle cx="8" cy="12" r="1.5" fill="#10b981" />
        <circle cx="16" cy="12" r="1.5" fill="#10b981" />
        <path d="M11 15H13" stroke="#10b981" strokeWidth="2" strokeLinecap="round" />
      </svg>
    )
  }

  if (p === "xai" || p === "grok-cli") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="currentColor">
        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
      </svg>
    )
  }

  if (p === "antigravity") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="8" stroke="#8b5cf6" strokeWidth="2" strokeDasharray="3 3" />
        <circle cx="12" cy="12" r="3" fill="#8b5cf6" />
        <path d="M12 2V6M12 18V22M2 12H6M18 12H22" stroke="#8b5cf6" strokeWidth="2" strokeLinecap="round" />
      </svg>
    )
  }

  if (p === "agnes") {
    return (
      <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L15 9L22 12L15 15L12 22L9 15L2 12L9 9L12 2Z" fill="#a855f7" />
      </svg>
    )
  }

  const initials = (provider || "??").slice(0, 2).toUpperCase()
  return (
    <span className="font-mono font-bold text-[10px] text-zinc-300">
      {initials}
    </span>
  )
}

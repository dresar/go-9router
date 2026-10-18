import React from "react"
import { Connection } from "../types"
import { ProviderIcon } from "./ProviderIcon"
import { ToggleSwitch } from "./ToggleSwitch"
import { StatusBadge } from "./StatusBadge"
import {
  isConnectionInCooldown,
  isConnectionError,
  formatProviderName,
  maskIdentity,
  formatTimeAgo
} from "../utils/status"
import {
  X,
  Search,
  Power,
  ExternalLink,
  Copy,
  Check,
  ChevronDown,
  ChevronUp
} from "lucide-react"

interface ProviderGroupDrawerProps {
  provider: string | null
  connections: Connection[]
  onClose: () => void
  onToggleConnection: (id: string, nextState: boolean) => Promise<void>
  onToggleProviderGlobal: (provider: string, nextState: boolean) => Promise<void>
  onOpenInTable: (provider: string) => void
  onSelectSingleConnection: (connection: Connection) => void
}

export const ProviderGroupDrawer: React.FC<ProviderGroupDrawerProps> = ({
  provider,
  connections,
  onClose,
  onToggleConnection,
  onToggleProviderGlobal,
  onOpenInTable
}) => {
  const [search, setSearch] = React.useState("")
  const [filterStatus, setFilterStatus] = React.useState<"all" | "active" | "inactive" | "cooldown">("all")
  const [togglingId, setTogglingId] = React.useState<string | null>(null)
  const [isGlobalToggling, setIsGlobalToggling] = React.useState(false)
  const [expandedId, setExpandedId] = React.useState<string | null>(null)
  const [copiedId, setCopiedId] = React.useState<string | null>(null)

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [onClose])

  if (!provider) return null

  const providerConnections = connections.filter((c) => c.provider === provider)
  const activeCount = providerConnections.filter((c) => c.isActive).length
  const totalCount = providerConnections.length
  const isAllActive = activeCount === totalCount && totalCount > 0

  const filteredConnections = providerConnections.filter((c) => {
    if (filterStatus === "active" && !c.isActive) return false
    if (filterStatus === "inactive" && c.isActive) return false
    if (filterStatus === "cooldown" && !isConnectionInCooldown(c) && !isConnectionError(c)) return false

    if (search.trim()) {
      const q = search.toLowerCase().trim()
      const matchName = (c.name || "").toLowerCase().includes(q)
      const matchEmail = (c.email || "").toLowerCase().includes(q)
      const matchId = (c.id || "").toLowerCase().includes(q)
      const matchDisplay = (c.displayName || "").toLowerCase().includes(q)
      if (!matchName && !matchEmail && !matchId && !matchDisplay) return false
    }

    return true
  })

  const handleToggleSingle = async (c: Connection) => {
    setTogglingId(c.id)
    try {
      await onToggleConnection(c.id, !c.isActive)
    } finally {
      setTogglingId(null)
    }
  }

  const handleGlobalToggle = async () => {
    setIsGlobalToggling(true)
    try {
      await onToggleProviderGlobal(provider, !isAllActive)
    } finally {
      setIsGlobalToggling(false)
    }
  }

  const handleCopyJson = (c: Connection) => {
    navigator.clipboard.writeText(JSON.stringify(c, null, 2))
    setCopiedId(c.id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-xs">
      <div className="w-full max-w-2xl bg-[#0d0d14] border border-[#1e1e2a] rounded-xl flex flex-col max-h-[90vh] shadow-2xl">
        <div className="flex items-center justify-between border-b border-[#1e1e2a] px-5 py-4 bg-[#121218] rounded-t-xl">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#18181b] border border-[#27272a] p-2">
              <ProviderIcon provider={provider} className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-bold text-white">{formatProviderName(provider)}</h2>
                <span className="rounded bg-emerald-500/15 border border-emerald-500/30 px-2 py-0.5 text-xs font-mono text-emerald-400">
                  {activeCount} / {totalCount} Aktif
                </span>
              </div>
              <p className="text-xs text-[#71717a] font-mono mt-0.5">
                Daftar seluruh akun dan kunci terhubung
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleGlobalToggle}
              disabled={isGlobalToggling}
              className={`flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-mono font-medium transition-colors ${
                isAllActive
                  ? "border border-amber-500/30 bg-amber-500/10 text-amber-400 hover:bg-amber-500/20"
                  : "border border-emerald-500/30 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20"
              }`}
            >
              <Power className="h-3.5 w-3.5" />
              <span>{isAllActive ? "Matikan Semua" : "Aktifkan Semua"}</span>
            </button>

            <button
              type="button"
              onClick={onClose}
              className="rounded-lg p-1.5 text-[#71717a] hover:bg-[#18181b] hover:text-white transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>

        <div className="border-b border-[#1e1e2a] bg-[#121218]/60 px-5 py-3 flex flex-col sm:flex-row items-center justify-between gap-2.5">
          <div className="relative w-full sm:w-64">
            <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari akun atau email..."
              className="w-full rounded-lg border border-[#27272a] bg-[#18181b] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono"
            />
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto justify-between sm:justify-end">
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value as any)}
              className="rounded-lg border border-[#27272a] bg-[#18181b] px-2.5 py-1.5 text-xs text-white focus:border-emerald-500 focus:outline-none font-mono"
            >
              <option value="all">Semua Akun ({totalCount})</option>
              <option value="active">Aktif Saja ({activeCount})</option>
              <option value="inactive">Nonaktif ({totalCount - activeCount})</option>
              <option value="cooldown">Cooldown / Isu</option>
            </select>

            <button
              type="button"
              onClick={() => {
                onOpenInTable(provider)
                onClose()
              }}
              className="flex items-center gap-1 rounded-lg border border-[#27272a] bg-[#18181b] px-2.5 py-1.5 text-xs text-[#a1a1aa] hover:text-white transition-colors font-mono shrink-0"
              title="Buka di tabel manajemen koneksi utama"
            >
              <span>Buka di Tabel</span>
              <ExternalLink className="h-3 w-3" />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-3 text-xs">
          {filteredConnections.length === 0 ? (
            <div className="py-12 text-center text-[#71717a] font-mono">
              Tidak ada akun yang sesuai dengan filter.
            </div>
          ) : (
            filteredConnections.map((c) => {
              const inCooldown = isConnectionInCooldown(c)
              const hasErr = isConnectionError(c)
              const isToggling = togglingId === c.id
              const isExpanded = expandedId === c.id

              let badgeStatus: "ready" | "cooldown" | "error" | "inactive" = "inactive"
              let badgeLabel = "○ Nonaktif"
              if (inCooldown) {
                badgeStatus = "cooldown"
                badgeLabel = "⏱ Cooldown"
              } else if (hasErr) {
                badgeStatus = "error"
                badgeLabel = "△ Isu"
              } else if (c.isActive) {
                badgeStatus = "ready"
                badgeLabel = "⊙ Ready"
              }

              let quotaStatusText = "Normal"
              if (inCooldown) {
                quotaStatusText = "Rate Limited / Kuota Tertunda"
              } else if (hasErr) {
                quotaStatusText = "Error Autentikasi / Jaringan"
              }

              return (
                <div
                  key={c.id}
                  className="rounded-lg border border-[#1e1e2a] bg-[#121218] p-3.5 space-y-2.5 hover:border-emerald-500/30 transition-colors"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-white truncate max-w-xs text-xs">
                          {c.displayName || c.name || maskIdentity(c.email)}
                        </span>

                        <span className="rounded bg-[#18181b] border border-[#27272a] px-1.5 py-0.5 text-[10px] font-mono text-zinc-300">
                          P{c.priority ?? 1}
                        </span>

                        <span className="rounded bg-[#18181b] border border-[#27272a] px-1.5 py-0.5 text-[10px] font-mono text-zinc-400 uppercase">
                          {c.authType || "apikey"}
                        </span>

                        <StatusBadge status={badgeStatus} label={badgeLabel} />
                      </div>

                      <div className="mt-1 flex items-center gap-2 text-[11px] font-mono text-[#71717a]">
                        <span className="truncate">
                          Email: {maskIdentity(c.email, c.name)}
                        </span>
                        <span>•</span>
                        <span>ID: {c.id.slice(0, 8)}</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <ToggleSwitch
                        checked={c.isActive}
                        onChange={() => handleToggleSingle(c)}
                        disabled={isToggling}
                        label={c.isActive ? "Matikan akun ini" : "Aktifkan akun ini"}
                      />
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-[11px] font-mono pt-2 border-t border-[#1e1e2a] text-[#a1a1aa]">
                    <div>
                      <span className="text-[#71717a]">Status Kuota: </span>
                      <span className={inCooldown ? "text-amber-400" : hasErr ? "text-rose-400" : "text-emerald-400"}>
                        {quotaStatusText}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <span>Terakhir: {formatTimeAgo(c.lastUsedAt)}</span>
                      <button
                        type="button"
                        onClick={() => setExpandedId(isExpanded ? null : c.id)}
                        className="text-[#71717a] hover:text-white transition-colors"
                        title="Tampilkan raw JSON"
                      >
                        {isExpanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                      </button>
                    </div>
                  </div>

                  {inCooldown && (c.lastError || c.rateLimitedUntil) && (
                    <div className="rounded bg-amber-500/5 border border-amber-500/20 p-2 text-[10px] font-mono text-amber-300 break-words">
                      {c.lastError || `Rate limited until ${new Date(c.rateLimitedUntil || "").toLocaleTimeString()}`}
                    </div>
                  )}

                  {isExpanded && (
                    <div className="pt-2 border-t border-[#1e1e2a] space-y-2">
                      <div className="flex items-center justify-between text-[10px] font-mono text-[#71717a]">
                        <span>Metadata Diagnostik</span>
                        <button
                          type="button"
                          onClick={() => handleCopyJson(c)}
                          className="flex items-center gap-1 text-[#a1a1aa] hover:text-white"
                        >
                          {copiedId === c.id ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                          <span>{copiedId === c.id ? "Tersalin" : "Salin JSON"}</span>
                        </button>
                      </div>
                      <pre className="text-[10px] font-mono text-emerald-300 bg-[#09090b] p-2.5 rounded-lg overflow-x-auto border border-[#1e1e2a]">
                        {JSON.stringify(c, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>

        <div className="border-t border-[#1e1e2a] p-4 bg-[#121218] rounded-b-xl flex items-center justify-between text-xs font-mono">
          <span className="text-[#71717a]">
            Menampilkan {filteredConnections.length} dari {totalCount} akun
          </span>

          <button
            type="button"
            onClick={onClose}
            className="rounded-lg bg-[#27272a] hover:bg-[#3f3f46] px-4 py-1.5 font-sans font-medium text-white transition-colors"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  )
}

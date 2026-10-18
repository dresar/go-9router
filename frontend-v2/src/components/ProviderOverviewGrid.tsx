import React from "react"
import { Connection } from "../types"
import { ProviderIcon } from "./ProviderIcon"
import { isConnectionInCooldown, isConnectionError } from "../utils/status"
import {
  Layers,
  Activity,
  ChevronRight,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Search,
  Power
} from "lucide-react"

interface ProviderOverviewGridProps {
  connections: Connection[]
  onSelectProvider: (provider: string) => void
  onToggleProviderGlobal: (provider: string, nextState: boolean) => Promise<void>
}

export const ProviderOverviewGrid: React.FC<ProviderOverviewGridProps> = ({
  connections,
  onSelectProvider,
  onToggleProviderGlobal
}) => {
  const [search, setSearch] = React.useState("")
  const [togglingProvider, setTogglingProvider] = React.useState<string | null>(null)

  const providerGroups = React.useMemo(() => {
    const map = new Map<
      string,
      {
        total: number
        active: number
        errors: number
        cooldowns: number
        connections: Connection[]
      }
    >()

    for (const c of connections) {
      const p = c.provider || "unknown"
      const current = map.get(p) || {
        total: 0,
        active: 0,
        errors: 0,
        cooldowns: 0,
        connections: []
      }
      current.total += 1
      if (c.isActive) current.active += 1
      if (isConnectionInCooldown(c)) current.cooldowns += 1
      else if (isConnectionError(c)) current.errors += 1
      current.connections.push(c)
      map.set(p, current)
    }

    const list = Array.from(map.entries()).map(([provider, stats]) => ({
      provider,
      ...stats
    }))

    list.sort((a, b) => b.total - a.total || b.active - a.active)
    return list
  }, [connections])

  const filteredGroups = React.useMemo(() => {
    if (!search.trim()) return providerGroups
    const q = search.toLowerCase()
    return providerGroups.filter((g) => g.provider.toLowerCase().includes(q))
  }, [providerGroups, search])

  const totalProviders = providerGroups.length
  const activeProviders = providerGroups.filter((g) => g.active > 0).length

  const handleGlobalToggle = async (e: React.MouseEvent, provider: string, currentActive: number) => {
    e.stopPropagation()
    const nextState = currentActive === 0
    setTogglingProvider(provider)
    try {
      await onToggleProviderGlobal(provider, nextState)
    } finally {
      setTogglingProvider(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#27272a] pb-4">
        <div>
          <h1 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
            <span>Provider Gateway</span>
            <span className="text-xs font-mono text-emerald-400 font-normal">
              ({activeProviders}/{totalProviders} Aktif)
            </span>
          </h1>
          <p className="text-xs text-[#a1a1aa] mt-0.5">
            Ringkasan {connections.length} koneksi provider AI yang terdaftar di sistem 9Router
          </p>
        </div>

        <div className="relative w-full sm:w-64">
          <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari provider..."
            className="w-full rounded-md border border-[#27272a] bg-[#121214] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 font-mono"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
        {filteredGroups.map((g) => {
          const isToggling = togglingProvider === g.provider
          const isAnyActive = g.active > 0
          const isFullyActive = g.active === g.total && g.total > 0

          let statusBadge = (
            <span className="inline-flex items-center gap-1 rounded bg-[#27272a] px-1.5 py-0.5 text-[10px] font-mono text-[#a1a1aa]">
              Idle
            </span>
          )

          if (g.cooldowns > 0) {
            statusBadge = (
              <span className="inline-flex items-center gap-1 rounded bg-amber-500/10 border border-amber-500/30 px-1.5 py-0.5 text-[10px] font-mono text-amber-400">
                <Clock className="h-2.5 w-2.5" />
                <span>⏱ Cooldown</span>
              </span>
            )
          } else if (g.errors > 0 && g.active === 0) {
            statusBadge = (
              <span className="inline-flex items-center gap-1 rounded bg-rose-500/10 border border-rose-500/30 px-1.5 py-0.5 text-[10px] font-mono text-rose-400">
                <AlertTriangle className="h-2.5 w-2.5" />
                <span>{g.errors} Isu</span>
              </span>
            )
          } else if (isAnyActive) {
            statusBadge = (
              <span className="inline-flex items-center gap-1 rounded bg-emerald-500/10 border border-emerald-500/30 px-1.5 py-0.5 text-[10px] font-mono text-emerald-400">
                <CheckCircle2 className="h-2.5 w-2.5" />
                <span>Ready</span>
              </span>
            )
          }

          return (
            <div
              key={g.provider}
              onClick={() => onSelectProvider(g.provider)}
              className="group cursor-pointer rounded-lg border border-[#27272a] bg-[#121214] p-3.5 hover:border-emerald-500/40 hover:bg-[#18181b] transition-all flex flex-col justify-between"
            >
              <div>
                <div className="flex items-start justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2.5">
                    <div className="flex h-8 w-8 items-center justify-center rounded-md bg-[#18181b] border border-[#27272a] p-1.5 group-hover:border-emerald-500/40 transition-colors">
                      <ProviderIcon provider={g.provider} className="h-5 w-5" />
                    </div>
                    <div>
                      <h3 className="text-xs font-semibold text-white group-hover:text-emerald-300 transition-colors capitalize">
                        {g.provider}
                      </h3>
                      <span className="text-[10px] text-[#71717a] font-mono">
                        {g.total} akun terhubung
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {statusBadge}

                    <button
                      type="button"
                      onClick={(e) => handleGlobalToggle(e, g.provider, g.active)}
                      disabled={isToggling}
                      title={
                        isAnyActive
                          ? `Matikan seluruh akun ${g.provider}`
                          : `Aktifkan seluruh akun ${g.provider}`
                      }
                      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                        isAnyActive ? "bg-emerald-600" : "bg-[#27272a]"
                      } ${isToggling ? "opacity-50" : ""}`}
                    >
                      <span
                        className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-xs transition duration-200 ease-in-out ${
                          isAnyActive ? "translate-x-4" : "translate-x-0"
                        }`}
                      />
                    </button>
                  </div>
                </div>

                <div className="mt-3 flex items-center justify-between text-[11px] font-mono pt-2 border-t border-[#27272a]/60">
                  <span className="text-[#a1a1aa]">Status Operasi:</span>
                  <span className={isAnyActive ? "text-emerald-400 font-medium" : "text-[#71717a]"}>
                    {g.active} / {g.total} Aktif
                  </span>
                </div>
              </div>

              <div className="mt-3 flex items-center justify-between text-[11px] text-[#71717a] group-hover:text-emerald-400 transition-colors pt-2">
                <span>Kelola {g.total} Akun</span>
                <ChevronRight className="h-3.5 w-3.5 transform group-hover:translate-x-0.5 transition-transform" />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

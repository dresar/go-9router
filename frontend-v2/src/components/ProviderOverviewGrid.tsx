import React from "react"
import { Connection } from "../types"
import { Layers, Activity, ChevronRight, CheckCircle2, AlertTriangle, XCircle, Search } from "lucide-react"

interface ProviderOverviewGridProps {
  connections: Connection[]
  onSelectProvider: (provider: string) => void
}

export const ProviderOverviewGrid: React.FC<ProviderOverviewGridProps> = ({
  connections,
  onSelectProvider
}) => {
  const [search, setSearch] = React.useState("")

  const providerGroups = React.useMemo(() => {
    const map = new Map<string, { total: number; active: number; errors: number; connections: Connection[] }>()

    for (const c of connections) {
      const p = c.provider || "unknown"
      const current = map.get(p) || { total: 0, active: 0, errors: 0, connections: [] }
      current.total += 1
      if (c.isActive) current.active += 1
      if (c.testStatus === "error" || c.lastError || (c.errorCode && c.errorCode >= 400)) {
        current.errors += 1
      }
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
            Ringkasan 189 koneksi provider AI yang terdaftar di sistem 9Router
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
          const isAllActive = g.active === g.total && g.total > 0
          const hasError = g.errors > 0

          return (
            <div
              key={g.provider}
              onClick={() => onSelectProvider(g.provider)}
              className="group cursor-pointer rounded-lg border border-[#27272a] bg-[#121214] p-3.5 hover:border-emerald-500/40 hover:bg-[#18181b] transition-all flex flex-col justify-between"
            >
              <div>
                <div className="flex items-start justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2">
                    <div className="flex h-7 w-7 items-center justify-center rounded-md bg-[#27272a] text-xs font-mono font-bold text-white uppercase group-hover:bg-emerald-500/20 group-hover:text-emerald-300 transition-colors">
                      {g.provider.slice(0, 2)}
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

                  <div>
                    {g.active === 0 ? (
                      <span className="inline-flex items-center gap-1 rounded bg-[#27272a] px-1.5 py-0.5 text-[10px] font-mono text-[#a1a1aa]">
                        Mati
                      </span>
                    ) : hasError ? (
                      <span className="inline-flex items-center gap-1 rounded bg-amber-500/10 border border-amber-500/30 px-1.5 py-0.5 text-[10px] font-mono text-amber-400">
                        <AlertTriangle className="h-2.5 w-2.5" />
                        {g.errors} Isu
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 rounded bg-emerald-500/10 border border-emerald-500/30 px-1.5 py-0.5 text-[10px] font-mono text-emerald-400">
                        <CheckCircle2 className="h-2.5 w-2.5" />
                        Ready
                      </span>
                    )}
                  </div>
                </div>

                <div className="mt-3 flex items-center justify-between text-[11px] font-mono pt-2 border-t border-[#27272a]/60">
                  <span className="text-[#a1a1aa]">Status Operasi:</span>
                  <span className={g.active > 0 ? "text-emerald-400 font-medium" : "text-[#71717a]"}>
                    {g.active} / {g.total} Aktif
                  </span>
                </div>
              </div>

              <div className="mt-3 flex items-center justify-between text-[11px] text-[#71717a] group-hover:text-emerald-400 transition-colors pt-2">
                <span>Kelola Akun</span>
                <ChevronRight className="h-3.5 w-3.5 transform group-hover:translate-x-0.5 transition-transform" />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

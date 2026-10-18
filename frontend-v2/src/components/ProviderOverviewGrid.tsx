import React from "react"
import { Connection } from "../types"
import { ProviderCard } from "./ProviderCard"
import { Search } from "lucide-react"

interface ProviderOverviewGridProps {
  connections: Connection[]
  onSelectProvider: (provider: string) => void
  onToggleProviderGlobal: (provider: string, nextState: boolean) => Promise<void>
  selectedProvider?: string | null
}

export const ProviderOverviewGrid: React.FC<ProviderOverviewGridProps> = ({
  connections,
  onSelectProvider,
  onToggleProviderGlobal,
  selectedProvider
}) => {
  const [search, setSearch] = React.useState("")

  const providerGroups = React.useMemo(() => {
    const map = new Map<string, Connection[]>()

    for (const c of connections) {
      const p = c.provider || "unknown"
      const current = map.get(p) || []
      current.push(c)
      map.set(p, current)
    }

    const list = Array.from(map.entries()).map(([provider, conns]) => ({
      provider,
      conns,
      total: conns.length,
      active: conns.filter((c) => c.isActive).length
    }))

    list.sort((a, b) => b.total - a.total || b.active - a.active)
    return list
  }, [connections])

  const filteredGroups = React.useMemo(() => {
    if (!search.trim()) return providerGroups
    const q = search.toLowerCase().trim()
    return providerGroups.filter(
      (g) =>
        g.provider.toLowerCase().includes(q) ||
        g.conns.some(
          (c) =>
            (c.email || "").toLowerCase().includes(q) ||
            (c.displayName || "").toLowerCase().includes(q) ||
            (c.name || "").toLowerCase().includes(q)
        )
    )
  }, [providerGroups, search])

  const totalProviders = providerGroups.length
  const activeProviders = providerGroups.filter((g) => g.active > 0).length

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#1e1e2a] pb-4">
        <div>
          <h1 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <span>Provider Gateway</span>
            <span className="text-sm font-mono text-emerald-400 font-normal">
              ({activeProviders}/{totalProviders} Aktif)
            </span>
          </h1>
          <p className="text-xs text-[#71717a] mt-0.5">
            Ringkasan {connections.length} koneksi provider AI yang terdaftar di sistem 9Router
          </p>
        </div>

        <div className="relative w-full sm:w-64">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari provider..."
            className="w-full rounded-lg border border-[#27272a] bg-[#121218] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3.5">
        {filteredGroups.map((g) => (
          <ProviderCard
            key={g.provider}
            provider={g.provider}
            connections={g.conns}
            isSelected={selectedProvider === g.provider}
            onSelect={onSelectProvider}
            onToggleGlobal={onToggleProviderGlobal}
          />
        ))}
      </div>
    </div>
  )
}

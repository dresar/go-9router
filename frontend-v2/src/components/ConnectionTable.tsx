import React from "react"
import { Connection } from "../types"
import { ProviderIcon } from "./ProviderIcon"
import { ToggleSwitch } from "./ToggleSwitch"
import {
  isConnectionInCooldown,
  isConnectionError,
  maskIdentity,
  formatTimeAgo
} from "../utils/status"
import {
  Search,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Eye
} from "lucide-react"

interface ConnectionTableProps {
  connections: Connection[]
  initialProviderFilter?: string
  onToggleActive: (id: string, nextState: boolean) => Promise<void>
  onSelectConnection: (connection: Connection) => void
}

export const ConnectionTable: React.FC<ConnectionTableProps> = ({
  connections,
  initialProviderFilter = "all",
  onToggleActive,
  onSelectConnection
}) => {
  const [search, setSearch] = React.useState("")
  const [selectedProvider, setSelectedProvider] = React.useState(initialProviderFilter)
  const [statusFilter, setStatusFilter] = React.useState<"all" | "active" | "inactive" | "cooldown" | "issues">("all")
  const [sortBy, setSortBy] = React.useState<"priority" | "name" | "provider">("priority")
  const [togglingId, setTogglingId] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (initialProviderFilter) {
      setSelectedProvider(initialProviderFilter)
    }
  }, [initialProviderFilter])

  const providerList = React.useMemo(() => {
    const set = new Set<string>()
    for (const c of connections) {
      if (c.provider) set.add(c.provider)
    }
    return Array.from(set).sort()
  }, [connections])

  const filteredConnections = React.useMemo(() => {
    return connections
      .filter((c) => {
        if (selectedProvider !== "all" && c.provider !== selectedProvider) {
          return false
        }

        if (statusFilter === "active" && !c.isActive) return false
        if (statusFilter === "inactive" && c.isActive) return false
        if (statusFilter === "cooldown" && !isConnectionInCooldown(c)) return false
        if (statusFilter === "issues" && !isConnectionError(c)) return false

        if (search.trim()) {
          const q = search.toLowerCase()
          const matchName = (c.name || "").toLowerCase().includes(q)
          const matchEmail = (c.email || "").toLowerCase().includes(q)
          const matchProvider = (c.provider || "").toLowerCase().includes(q)
          const matchId = (c.id || "").toLowerCase().includes(q)
          const matchDisplayName = (c.displayName || "").toLowerCase().includes(q)
          if (!matchName && !matchEmail && !matchProvider && !matchId && !matchDisplayName) {
            return false
          }
        }

        return true
      })
      .sort((a, b) => {
        if (sortBy === "priority") {
          return (a.priority || 0) - (b.priority || 0)
        }
        if (sortBy === "provider") {
          return a.provider.localeCompare(b.provider)
        }
        return (a.name || a.email || "").localeCompare(b.name || b.email || "")
      })
  }, [connections, selectedProvider, statusFilter, search, sortBy])

  const handleToggle = async (c: Connection) => {
    setTogglingId(c.id)
    try {
      await onToggleActive(c.id, !c.isActive)
    } finally {
      setTogglingId(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 border-b border-[#1e1e2a] pb-4">
        <div>
          <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
            <span>Daftar Koneksi Provider</span>
            <span className="text-xs font-mono text-[#71717a] font-normal">
              ({filteredConnections.length} dari {connections.length})
            </span>
          </h2>
          <p className="text-xs text-[#71717a] mt-0.5">
            Manajemen status aktif, prioritas failover, dan observasi akun
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <div className="relative min-w-[200px] flex-1 sm:flex-initial">
            <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Filter nama, email, ID..."
              className="w-full rounded-lg border border-[#27272a] bg-[#121218] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono"
            />
          </div>

          <select
            value={selectedProvider}
            onChange={(e) => setSelectedProvider(e.target.value)}
            className="rounded-lg border border-[#27272a] bg-[#121218] px-3 py-1.5 text-xs text-white focus:border-emerald-500 focus:outline-none font-mono"
          >
            <option value="all">Semua Provider ({connections.length})</option>
            {providerList.map((p) => {
              const count = connections.filter((c) => c.provider === p).length
              return (
                <option key={p} value={p}>
                  {p} ({count})
                </option>
              )
            })}
          </select>

          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as any)}
            className="rounded-lg border border-[#27272a] bg-[#121218] px-3 py-1.5 text-xs text-white focus:border-emerald-500 focus:outline-none font-mono"
          >
            <option value="all">Semua Status</option>
            <option value="active">Aktif Saja</option>
            <option value="inactive">Nonaktif</option>
            <option value="cooldown">⏱ Cooldown</option>
            <option value="issues">Isu / Error</option>
          </select>

          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="rounded-lg border border-[#27272a] bg-[#121218] px-3 py-1.5 text-xs text-white focus:border-emerald-500 focus:outline-none font-mono"
          >
            <option value="priority">Urut Prioritas</option>
            <option value="provider">Urut Provider</option>
            <option value="name">Urut Nama</option>
          </select>
        </div>
      </div>

      <div className="overflow-x-auto rounded-lg border border-[#1e1e2a] bg-[#121218]">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="border-b border-[#1e1e2a] bg-[#0d0d14] text-[#71717a] font-mono text-[11px]">
              <th className="py-2.5 px-3 w-16 text-center">Status</th>
              <th className="py-2.5 px-3">Provider</th>
              <th className="py-2.5 px-3">Identitas / Akun</th>
              <th className="py-2.5 px-3 w-20 text-center">Prioritas</th>
              <th className="py-2.5 px-3">Kondisi Health</th>
              <th className="py-2.5 px-3 hidden lg:table-cell">Terakhir Digunakan</th>
              <th className="py-2.5 px-3 text-right w-20">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#1e1e2a]">
            {filteredConnections.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-xs text-[#71717a]">
                  Tidak ada koneksi yang sesuai dengan filter.
                </td>
              </tr>
            ) : (
              filteredConnections.map((c) => {
                const inCooldown = isConnectionInCooldown(c)
                const hasError = isConnectionError(c)
                const isToggling = togglingId === c.id

                return (
                  <tr key={c.id} className="hover:bg-[#18181b] transition-colors group">
                    <td className="py-2.5 px-3 text-center">
                      <ToggleSwitch
                        checked={c.isActive}
                        onChange={() => handleToggle(c)}
                        disabled={isToggling}
                        label={c.isActive ? "Klik untuk mematikan" : "Klik untuk mengaktifkan"}
                        size="sm"
                      />
                    </td>

                    <td className="py-2.5 px-3 font-mono font-medium text-white">
                      <span className="inline-flex items-center gap-1.5 rounded bg-[#18181b] border border-[#27272a] px-2 py-0.5 text-[11px] text-white">
                        <ProviderIcon provider={c.provider} className="h-3.5 w-3.5 shrink-0" />
                        <span>{c.provider}</span>
                      </span>
                    </td>

                    <td className="py-2.5 px-3">
                      <div className="font-medium text-white truncate max-w-xs">
                        {c.displayName || c.name || maskIdentity(c.email)}
                      </div>
                      <div className="text-[11px] font-mono text-[#71717a] truncate max-w-xs">
                        {maskIdentity(c.email, c.name)} • ID: {c.id.slice(0, 8)}
                      </div>
                    </td>

                    <td className="py-2.5 px-3 text-center font-mono">
                      <span className="rounded bg-[#27272a] px-2 py-0.5 text-[10px] text-[#f4f4f5]">
                        P{c.priority ?? 1}
                      </span>
                    </td>

                    <td className="py-2.5 px-3 font-mono text-[11px]">
                      {!c.isActive ? (
                        <span className="text-[#71717a]">Nonaktif</span>
                      ) : inCooldown ? (
                        <div className="flex items-center gap-1.5 text-amber-400">
                          <Clock className="h-3 w-3 shrink-0" />
                          <span className="truncate max-w-[200px]" title={c.lastError || "Rate limited / cooldown"}>
                            ⏱ Cooldown
                          </span>
                        </div>
                      ) : hasError ? (
                        <div className="flex items-center gap-1.5 text-rose-400">
                          <AlertTriangle className="h-3 w-3 shrink-0" />
                          <span className="truncate max-w-[200px]" title={c.lastError || "Error"}>
                            {c.errorCode ? `[${c.errorCode}] ` : ""}
                            {c.lastError || "Isu Kunci"}
                          </span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5 text-emerald-400">
                          <CheckCircle2 className="h-3 w-3 shrink-0" />
                          <span>Ready</span>
                        </div>
                      )}
                    </td>

                    <td className="py-2.5 px-3 text-[#71717a] font-mono text-[11px] hidden lg:table-cell">
                      {formatTimeAgo(c.lastUsedAt)}
                    </td>

                    <td className="py-2.5 px-3 text-right">
                      <button
                        type="button"
                        onClick={() => onSelectConnection(c)}
                        className="rounded-lg border border-[#27272a] bg-[#18181b] p-1.5 text-[#a1a1aa] hover:border-emerald-500/50 hover:bg-emerald-500/10 hover:text-emerald-400 transition-colors"
                        title="Lihat Detail JSON"
                      >
                        <Eye className="h-3.5 w-3.5" />
                      </button>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

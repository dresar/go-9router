import React from "react"
import { UsageStats } from "../types"
import { BarChart3, Activity, Zap, CheckCircle2, AlertTriangle, Layers } from "lucide-react"

interface UsageStatsViewProps {
  stats: UsageStats | null
}

export const UsageStatsView: React.FC<UsageStatsViewProps> = ({ stats }) => {
  const modelsUsage = React.useMemo(() => {
    if (!stats?.byModel) return []
    return Object.entries(stats.byModel).map(([key, val]: [string, any]) => ({
      key,
      rawModel: val.rawModel || key,
      provider: val.provider || "unknown",
      requests: val.requests || 0,
      promptTokens: val.promptTokens || 0,
      completionTokens: val.completionTokens || 0,
      lastUsed: val.lastUsed
    })).sort((a, b) => b.requests - a.requests)
  }, [stats])

  const totalRequests = modelsUsage.reduce((acc, m) => acc + m.requests, 0)
  const totalPromptTokens = modelsUsage.reduce((acc, m) => acc + m.promptTokens, 0)
  const totalCompletionTokens = modelsUsage.reduce((acc, m) => acc + m.completionTokens, 0)

  const recentRequests = stats?.recentRequests || []

  return (
    <div className="space-y-6">
      <div className="border-b border-[#27272a] pb-4">
        <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
          <span>Metrik & Ringkasan Penggunaan Gateway</span>
        </h2>
        <p className="text-xs text-[#a1a1aa] mt-0.5">
          Observabilitas beban panggilan token dan aktivitas gateway 9Router
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-[#a1a1aa] text-xs font-mono mb-1">
            <Activity className="h-4 w-4 text-emerald-400" />
            <span>Total Permintaan Terproses</span>
          </div>
          <div className="text-xl font-bold font-mono text-white mt-1">
            {totalRequests.toLocaleString()}
          </div>
        </div>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-[#a1a1aa] text-xs font-mono mb-1">
            <Zap className="h-4 w-4 text-emerald-400" />
            <span>Prompt Tokens</span>
          </div>
          <div className="text-xl font-bold font-mono text-emerald-400 mt-1">
            {totalPromptTokens.toLocaleString()}
          </div>
        </div>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-[#a1a1aa] text-xs font-mono mb-1">
            <BarChart3 className="h-4 w-4 text-emerald-400" />
            <span>Completion Tokens</span>
          </div>
          <div className="text-xl font-bold font-mono text-white mt-1">
            {totalCompletionTokens.toLocaleString()}
          </div>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-xs font-semibold text-white uppercase tracking-wider font-mono">
          Distribusi Penggunaan per Model
        </h3>

        <div className="overflow-x-auto rounded-lg border border-[#27272a] bg-[#121214]">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="border-b border-[#27272a] bg-[#18181b]/60 text-[#a1a1aa] font-mono text-[11px]">
                <th className="py-2.5 px-3">Model</th>
                <th className="py-2.5 px-3">Provider</th>
                <th className="py-2.5 px-3 text-right">Permintaan</th>
                <th className="py-2.5 px-3 text-right">Prompt Tokens</th>
                <th className="py-2.5 px-3 text-right">Completion Tokens</th>
                <th className="py-2.5 px-3 text-right">Terakhir Dipakai</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#27272a]">
              {modelsUsage.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-6 text-center text-xs text-[#71717a]">
                    Belum ada data metrik tercatat.
                  </td>
                </tr>
              ) : (
                modelsUsage.map((m, idx) => (
                  <tr key={idx} className="hover:bg-[#18181b] transition-colors font-mono">
                    <td className="py-2.5 px-3 font-medium text-white max-w-xs truncate" title={m.rawModel}>
                      {m.rawModel}
                    </td>
                    <td className="py-2.5 px-3">
                      <span className="rounded bg-[#27272a] px-2 py-0.5 text-[10px] text-zinc-300">
                        {m.provider}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 text-right font-medium text-white">
                      {m.requests.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-3 text-right text-emerald-400">
                      {m.promptTokens.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-3 text-right text-zinc-300">
                      {m.completionTokens.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-3 text-right text-[#71717a] text-[10px]">
                      {m.lastUsed ? new Date(m.lastUsed).toLocaleDateString() : "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {recentRequests.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-xs font-semibold text-white uppercase tracking-wider font-mono">
            Riwayat Permintaan Terkini
          </h3>

          <div className="overflow-x-auto rounded-lg border border-[#27272a] bg-[#121214]">
            <table className="w-full text-left text-xs border-collapse font-mono">
              <thead>
                <tr className="border-b border-[#27272a] bg-[#18181b]/60 text-[#a1a1aa] text-[11px]">
                  <th className="py-2.5 px-3">Waktu</th>
                  <th className="py-2.5 px-3">Model</th>
                  <th className="py-2.5 px-3">Provider</th>
                  <th className="py-2.5 px-3 text-right">Tokens</th>
                  <th className="py-2.5 px-3 text-center">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#27272a]">
                {recentRequests.slice(0, 15).map((req, idx) => {
                  const isOk = req.status === "ok" || req.status === "200"
                  return (
                    <tr key={idx} className="hover:bg-[#18181b] transition-colors text-[11px]">
                      <td className="py-2 px-3 text-[#a1a1aa]">
                        {new Date(req.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="py-2 px-3 font-medium text-white truncate max-w-xs">
                        {req.model}
                      </td>
                      <td className="py-2 px-3 text-zinc-400">{req.provider}</td>
                      <td className="py-2 px-3 text-right text-zinc-300">
                        {req.promptTokens + req.completionTokens}
                      </td>
                      <td className="py-2 px-3 text-center">
                        <span
                          className={`rounded px-1.5 py-0.5 text-[10px] ${
                            isOk
                              ? "bg-emerald-500/10 border border-emerald-500/30 text-emerald-400"
                              : "bg-rose-500/10 border border-rose-500/30 text-rose-400"
                          }`}
                        >
                          {req.status}
                        </span>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

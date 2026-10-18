import React from "react"
import { ModelItem, Combo, CustomModel } from "../types"
import { Cpu, Search, Layers, Play, ChevronDown, ChevronRight, Sparkles, Database } from "lucide-react"

interface ModelsAndCombosViewProps {
  models: ModelItem[]
  combos: Combo[]
  customModels: CustomModel[]
  onSelectModelForPlayground: (modelId: string) => void
}

export const ModelsAndCombosView: React.FC<ModelsAndCombosViewProps> = ({
  models,
  combos,
  customModels,
  onSelectModelForPlayground
}) => {
  const [activeSubTab, setActiveSubTab] = React.useState<"combos" | "gateway" | "custom">("combos")
  const [search, setSearch] = React.useState("")
  const [expandedComboId, setExpandedComboId] = React.useState<string | null>("558accdd-22ae-4ab2-bff7-31dc9c8809d5")

  const filteredGatewayModels = React.useMemo(() => {
    if (!search.trim()) return models
    const q = search.toLowerCase()
    return models.filter((m) => m.id.toLowerCase().includes(q) || (m.owned_by || "").toLowerCase().includes(q))
  }, [models, search])

  const filteredCustomModels = React.useMemo(() => {
    if (!search.trim()) return customModels
    const q = search.toLowerCase()
    return customModels.filter((m) => (m.id || "").toLowerCase().includes(q) || (m.providerAlias || "").toLowerCase().includes(q))
  }, [customModels, search])

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#27272a] pb-4">
        <div>
          <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
            <span>Katalog Model & Combo Router</span>
          </h2>
          <p className="text-xs text-[#a1a1aa] mt-0.5">
            Daftar {combos.length} Smart Combo, {models.length} Model Gateway Aktif, dan {customModels.length} Model Terdaftar
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex rounded-md border border-[#27272a] bg-[#121214] p-0.5 text-xs font-mono">
            <button
              onClick={() => setActiveSubTab("combos")}
              className={`px-3 py-1 rounded transition-colors ${
                activeSubTab === "combos" ? "bg-[#27272a] text-white font-medium" : "text-[#a1a1aa] hover:text-white"
              }`}
            >
              Combo ({combos.length})
            </button>
            <button
              onClick={() => setActiveSubTab("gateway")}
              className={`px-3 py-1 rounded transition-colors ${
                activeSubTab === "gateway" ? "bg-[#27272a] text-white font-medium" : "text-[#a1a1aa] hover:text-white"
              }`}
            >
              Gateway ({models.length})
            </button>
            <button
              onClick={() => setActiveSubTab("custom")}
              className={`px-3 py-1 rounded transition-colors ${
                activeSubTab === "custom" ? "bg-[#27272a] text-white font-medium" : "text-[#a1a1aa] hover:text-white"
              }`}
            >
              Custom ({customModels.length})
            </button>
          </div>
        </div>
      </div>

      {activeSubTab === "combos" && (
        <div className="space-y-3">
          <div className="grid grid-cols-1 gap-3">
            {combos.map((c) => {
              const isExpanded = expandedComboId === c.id
              return (
                <div
                  key={c.id}
                  className="rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden"
                >
                  <div
                    onClick={() => setExpandedComboId(isExpanded ? null : c.id)}
                    className="flex items-center justify-between p-4 cursor-pointer hover:bg-[#18181b] transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex h-8 w-8 items-center justify-center rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-mono font-bold text-xs">
                        {c.name.slice(0, 2)}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-semibold text-white font-mono">{c.name}</h3>
                          <span className="rounded bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.5 text-[10px] font-mono text-emerald-400 uppercase">
                            Smart Router
                          </span>
                        </div>
                        <p className="text-[11px] font-mono text-[#71717a]">
                          {c.models?.length || 0} Model Failover Terpasang • ID: {c.id.slice(0, 8)}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          onSelectModelForPlayground(c.name)
                        }}
                        className="flex items-center gap-1.5 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2.5 py-1 text-xs text-emerald-400 hover:bg-emerald-500/20 transition-colors"
                      >
                        <Play className="h-3 w-3" />
                        <span>Coba</span>
                      </button>

                      <div className="text-[#a1a1aa] p-1">
                        {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                      </div>
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="border-t border-[#27272a] bg-[#09090b] p-4">
                      <div className="text-xs font-medium text-[#a1a1aa] mb-2 flex items-center justify-between">
                        <span>Urutan Failover Pipeline (Prioritas Eksekusi)</span>
                        <span className="font-mono text-[11px] text-emerald-400">{c.models?.length || 0} Models</span>
                      </div>
                      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
                        {c.models?.map((modelName, idx) => (
                          <div
                            key={idx}
                            className="flex items-center gap-2 rounded-md border border-[#27272a] bg-[#121214] px-2.5 py-1.5 font-mono text-xs text-zinc-300"
                          >
                            <span className="flex h-4 w-4 items-center justify-center rounded-full bg-[#27272a] text-[9px] font-bold text-[#a1a1aa] shrink-0">
                              {idx + 1}
                            </span>
                            <span className="truncate">{modelName}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {activeSubTab === "gateway" && (
        <div className="space-y-3">
          <div className="relative w-full max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari model gateway..."
              className="w-full rounded-md border border-[#27272a] bg-[#121214] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
            {filteredGatewayModels.map((m) => (
              <div
                key={m.id}
                className="flex items-center justify-between rounded-lg border border-[#27272a] bg-[#121214] p-3 hover:border-emerald-500/40 transition-colors"
              >
                <div className="min-w-0 pr-2">
                  <div className="font-mono text-xs font-semibold text-white truncate" title={m.id}>
                    {m.id}
                  </div>
                  <div className="text-[10px] font-mono text-[#71717a] flex items-center gap-2 mt-0.5">
                    <span>{m.owned_by || "system"}</span>
                    {m.context_length ? (
                      <span>{Math.round(m.context_length / 1000)}k ctx</span>
                    ) : null}
                  </div>
                </div>

                <button
                  onClick={() => onSelectModelForPlayground(m.id)}
                  className="shrink-0 flex items-center gap-1 rounded border border-[#27272a] bg-[#18181b] px-2 py-1 text-[11px] text-[#a1a1aa] hover:border-emerald-500/40 hover:text-emerald-400 transition-colors"
                  title="Gunakan di Playground"
                >
                  <Play className="h-3 w-3" />
                  <span>Uji</span>
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeSubTab === "custom" && (
        <div className="space-y-3">
          <div className="relative w-full max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-[#71717a]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari model kustom..."
              className="w-full rounded-md border border-[#27272a] bg-[#121214] pl-8 pr-3 py-1.5 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
            {filteredCustomModels.map((m, idx) => (
              <div
                key={idx}
                className="flex items-center justify-between rounded-lg border border-[#27272a] bg-[#121214] p-3 hover:border-emerald-500/40 transition-colors"
              >
                <div className="min-w-0 pr-2">
                  <div className="font-mono text-xs font-semibold text-white truncate" title={m.id}>
                    {m.id}
                  </div>
                  <div className="text-[10px] font-mono text-[#71717a] flex items-center gap-2 mt-0.5">
                    <span className="text-emerald-400">{m.providerAlias || "custom"}</span>
                    <span>•</span>
                    <span className="uppercase">{m.type || "llm"}</span>
                  </div>
                </div>

                <button
                  onClick={() => onSelectModelForPlayground(m.id)}
                  className="shrink-0 flex items-center gap-1 rounded border border-[#27272a] bg-[#18181b] px-2 py-1 text-[11px] text-[#a1a1aa] hover:border-emerald-500/40 hover:text-emerald-400 transition-colors"
                >
                  <Play className="h-3 w-3" />
                  <span>Uji</span>
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

import React from "react"
import {
  Layers,
  Activity,
  Cpu,
  Terminal,
  Send,
  BarChart3,
  RefreshCw,
  ShieldCheck
} from "lucide-react"

interface HeaderProps {
  activeTab: string
  setActiveTab: (tab: string) => void
  totalConnections: number
  activeConnections: number
  totalModels: number
  onRefresh: () => void
  isLoading: boolean
}

export const Header: React.FC<HeaderProps> = ({
  activeTab,
  setActiveTab,
  totalConnections,
  activeConnections,
  totalModels,
  onRefresh,
  isLoading
}) => {
  const navItems = [
    { id: "overview", label: "Overview", icon: Layers },
    { id: "connections", label: "Connections", icon: Activity },
    { id: "models", label: "Models & Combos", icon: Cpu },
    { id: "setup", label: "Setup & Keys", icon: Terminal },
    { id: "playground", label: "Playground", icon: Send },
    { id: "metrics", label: "Metrics", icon: BarChart3 }
  ]

  return (
    <header className="sticky top-0 z-40 w-full border-b border-[#1e1e2a] bg-[#09090b]/95 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#1a1a2e] border border-[#27272a] text-white font-mono font-bold text-sm">
            9R
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold tracking-tight text-white text-base">9Router</span>
              <span className="inline-flex items-center gap-1 rounded bg-[#18181b] border border-[#27272a] px-1.5 py-0.5 text-[11px] font-mono text-emerald-400">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                v0.5.91
              </span>
            </div>
            <div className="flex items-center gap-2 text-[11px] font-mono text-[#2dd4bf]">
              <span>Port 20128</span>
              <span className="text-zinc-600">•</span>
              <span>Direct 20129</span>
            </div>
          </div>
        </div>

        <div className="hidden sm:flex items-center gap-3 rounded-lg border border-[#27272a] bg-[#121214] px-3.5 py-1.5 font-mono text-xs">
          <div>
            <span className="text-[#71717a]">Total:</span>
            <span className="font-bold text-white ml-1.5">{totalConnections}</span>
          </div>
          <span className="text-[#27272a]">|</span>
          <div>
            <span className="text-[#71717a]">Aktif:</span>
            <span className="font-bold text-emerald-400 ml-1.5">{activeConnections}</span>
          </div>
          <span className="text-[#27272a]">|</span>
          <div>
            <span className="text-[#71717a]">Models:</span>
            <span className="font-bold text-white ml-1.5">{totalModels}</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={onRefresh}
            disabled={isLoading}
            title="Perbarui Data"
            className="flex h-8 w-8 items-center justify-center rounded-lg border border-[#27272a] bg-[#18181b] text-[#a1a1aa] hover:bg-[#27272a] hover:text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      <div className="border-t border-[#1e1e2a] bg-[#0d0d14]">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-2 sm:px-6">
          <nav className="flex items-center gap-1 overflow-x-auto whitespace-nowrap">
            {navItems.map((item) => {
              const Icon = item.icon
              const active = activeTab === item.id
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                    active
                      ? "bg-[#27272a] text-white shadow-xs"
                      : "text-[#a1a1aa] hover:text-white hover:bg-[#18181b]"
                  }`}
                >
                  <Icon className={`h-3.5 w-3.5 ${active ? "text-emerald-400" : "text-[#71717a]"}`} />
                  <span>{item.label}</span>
                </button>
              )
            })}
          </nav>

          <div className="hidden sm:inline-flex items-center gap-1.5 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-1.5 text-xs font-medium text-emerald-400 shrink-0">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
            <ShieldCheck className="h-3.5 w-3.5" />
            <span>Public Access</span>
          </div>
        </div>
      </div>
    </header>
  )
}

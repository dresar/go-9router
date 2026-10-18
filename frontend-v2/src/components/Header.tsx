import React from "react"
import {
  Activity,
  Layers,
  Cpu,
  Terminal,
  Send,
  BarChart3,
  RefreshCw,
  LogIn,
  LogOut,
  ShieldCheck,
  Lock,
  Menu,
  X
} from "lucide-react"

interface HeaderProps {
  activeTab: string
  setActiveTab: (tab: string) => void
  totalConnections: number
  activeConnections: number
  totalModels: number
  authenticated: boolean
  requireLogin: boolean
  onOpenLogin: () => void
  onLogout: () => void
  onRefresh: () => void
  isLoading: boolean
}

export const Header: React.FC<HeaderProps> = ({
  activeTab,
  setActiveTab,
  totalConnections,
  activeConnections,
  totalModels,
  authenticated,
  requireLogin,
  onOpenLogin,
  onLogout,
  onRefresh,
  isLoading
}) => {
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false)

  const navItems = [
    { id: "overview", label: "Overview", icon: Layers },
    { id: "connections", label: "Connections", icon: Activity },
    { id: "models", label: "Models & Combos", icon: Cpu },
    { id: "setup", label: "Setup & Keys", icon: Terminal },
    { id: "playground", label: "Playground", icon: Send },
    { id: "usage", label: "Metrics", icon: BarChart3 }
  ]

  return (
    <header className="sticky top-0 z-40 w-full border-b border-[#27272a] bg-[#09090b]/90 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-mono font-bold text-sm">
              9R
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-semibold tracking-tight text-white text-base">9Router</span>
                <span className="inline-flex items-center gap-1 rounded bg-[#18181b] border border-[#27272a] px-1.5 py-0.5 text-[11px] font-mono text-emerald-400">
                  <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                  v0.5.91
                </span>
              </div>
              <div className="hidden sm:flex items-center gap-2 text-[11px] text-[#a1a1aa] font-mono">
                <span>Port 20128</span>
                <span>•</span>
                <span className="text-emerald-400">Direct 20129</span>
              </div>
            </div>
          </div>

          <div className="hidden lg:flex items-center gap-2 border-l border-[#27272a] pl-4">
            <div className="flex items-center gap-1.5 rounded-md bg-[#18181b] border border-[#27272a] px-2.5 py-1 text-xs font-mono">
              <span className="text-[#a1a1aa]">Total:</span>
              <span className="font-semibold text-white">{totalConnections}</span>
            </div>
            <div className="flex items-center gap-1.5 rounded-md bg-[#18181b] border border-[#27272a] px-2.5 py-1 text-xs font-mono">
              <span className="text-[#a1a1aa]">Aktif:</span>
              <span className="font-semibold text-emerald-400">{activeConnections}</span>
            </div>
            <div className="flex items-center gap-1.5 rounded-md bg-[#18181b] border border-[#27272a] px-2.5 py-1 text-xs font-mono">
              <span className="text-[#a1a1aa]">Models:</span>
              <span className="font-semibold text-white">{totalModels}</span>
            </div>
          </div>
        </div>

        <nav className="hidden md:flex items-center gap-1 bg-[#121214] p-1 rounded-lg border border-[#27272a]">
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
                <Icon className={`h-3.5 w-3.5 ${active ? "text-emerald-400" : "text-[#a1a1aa]"}`} />
                <span>{item.label}</span>
              </button>
            )
          })}
        </nav>

        <div className="flex items-center gap-2">
          <button
            onClick={onRefresh}
            disabled={isLoading}
            title="Perbarui Data"
            className="flex h-8 w-8 items-center justify-center rounded-md border border-[#27272a] bg-[#18181b] text-[#a1a1aa] hover:bg-[#27272a] hover:text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin text-emerald-400" : ""}`} />
          </button>

          {requireLogin ? (
            authenticated ? (
              <button
                onClick={onLogout}
                className="flex items-center gap-1.5 rounded-md border border-[#27272a] bg-[#18181b] px-2.5 py-1.5 text-xs text-[#a1a1aa] hover:border-rose-500/50 hover:bg-rose-500/10 hover:text-rose-400 transition-colors"
              >
                <LogOut className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Logout</span>
              </button>
            ) : (
              <button
                onClick={onOpenLogin}
                className="flex items-center gap-1.5 rounded-md border border-emerald-500/40 bg-emerald-500/10 px-2.5 py-1.5 text-xs text-emerald-400 hover:bg-emerald-500/20 transition-colors"
              >
                <LogIn className="h-3.5 w-3.5" />
                <span>Login</span>
              </button>
            )
          ) : (
            <div className="hidden sm:flex items-center gap-1.5 rounded-md border border-[#27272a] bg-[#18181b] px-2.5 py-1 text-[11px] font-mono text-emerald-400">
              <ShieldCheck className="h-3.5 w-3.5" />
              <span>Public Access</span>
            </div>
          )}

          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden flex h-8 w-8 items-center justify-center rounded-md border border-[#27272a] bg-[#18181b] text-[#a1a1aa] hover:text-white"
          >
            {mobileMenuOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </button>
        </div>
      </div>

      {mobileMenuOpen && (
        <div className="md:hidden border-t border-[#27272a] bg-[#0c0a09] px-4 py-3">
          <div className="grid grid-cols-2 gap-2 mb-3">
            <div className="rounded-md bg-[#18181b] p-2 text-xs font-mono border border-[#27272a]">
              <span className="text-[#a1a1aa] block text-[10px]">Total Koneksi</span>
              <span className="text-white font-bold">{totalConnections}</span>
            </div>
            <div className="rounded-md bg-[#18181b] p-2 text-xs font-mono border border-[#27272a]">
              <span className="text-[#a1a1aa] block text-[10px]">Koneksi Aktif</span>
              <span className="text-emerald-400 font-bold">{activeConnections}</span>
            </div>
          </div>
          <div className="flex flex-col gap-1">
            {navItems.map((item) => {
              const Icon = item.icon
              const active = activeTab === item.id
              return (
                <button
                  key={item.id}
                  onClick={() => {
                    setActiveTab(item.id)
                    setMobileMenuOpen(false)
                  }}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-xs font-medium transition-colors ${
                    active
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      : "text-[#a1a1aa] hover:bg-[#18181b] hover:text-white"
                  }`}
                >
                  <Icon className="h-4 w-4" />
                  <span>{item.label}</span>
                </button>
              )
            })}
          </div>
        </div>
      )}
    </header>
  )
}

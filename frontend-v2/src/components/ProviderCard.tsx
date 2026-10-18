import React from "react"
import { Connection } from "../types"
import { ProviderIcon } from "./ProviderIcon"
import { ToggleSwitch } from "./ToggleSwitch"
import { StatusBadge } from "./StatusBadge"
import { getProviderStatus, formatProviderName } from "../utils/status"

interface ProviderCardProps {
  provider: string
  connections: Connection[]
  isSelected?: boolean
  onSelect: (provider: string) => void
  onToggleGlobal: (provider: string, nextState: boolean) => Promise<void>
}

export const ProviderCard: React.FC<ProviderCardProps> = ({
  provider,
  connections,
  isSelected = false,
  onSelect,
  onToggleGlobal
}) => {
  const [isToggling, setIsToggling] = React.useState(false)

  const total = connections.length
  const active = connections.filter((c) => c.isActive).length
  const isAnyActive = active > 0
  const statusInfo = getProviderStatus(connections)

  const handleToggle = async () => {
    if (isToggling) return
    setIsToggling(true)
    try {
      await onToggleGlobal(provider, !isAnyActive)
    } finally {
      setIsToggling(false)
    }
  }

  return (
    <div
      onClick={() => onSelect(provider)}
      className={`group cursor-pointer rounded-xl border bg-[#121218] p-4 transition-all duration-200 hover:border-[#2a2a3a] hover:-translate-y-0.5 flex flex-col justify-between ${
        isSelected ? "border-emerald-500/80 ring-1 ring-emerald-500/30" : "border-[#1e1e2a]"
      }`}
    >
      <div>
        <div className="flex items-start justify-between gap-2 mb-3">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-[#18181b] border border-[#27272a] p-1.5">
              <ProviderIcon provider={provider} className="h-4 w-4" />
            </div>
            <div className="min-w-0 truncate">
              <h3 className="text-sm font-semibold text-white group-hover:text-emerald-300 transition-colors truncate">
                {formatProviderName(provider)}
              </h3>
              <span className="text-[11px] text-[#71717a] font-mono block">
                {total} akun terhubung
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <StatusBadge
              status={statusInfo.status}
              label={statusInfo.label}
              errorCount={statusInfo.errorCount}
            />
            <ToggleSwitch
              checked={isAnyActive}
              onChange={handleToggle}
              disabled={isToggling}
              label={isAnyActive ? `Matikan semua akun ${provider}` : `Aktifkan semua akun ${provider}`}
            />
          </div>
        </div>

        <div className="mt-3.5 flex items-center justify-between text-xs font-mono">
          <span className="text-[#71717a]">Status Operasi:</span>
          <span className={isAnyActive ? "text-emerald-400 font-semibold" : "text-zinc-500 font-semibold"}>
            {active} / {total} Aktif
          </span>
        </div>
      </div>

      <div className="mt-3.5 pt-2.5 border-t border-[#1e1e2a] flex items-center justify-between text-xs text-[#71717a] group-hover:text-emerald-400 transition-colors">
        <span className={isSelected ? "text-emerald-400 font-medium" : ""}>
          Kelola {total} Akun
        </span>
        <span className="font-mono text-sm leading-none group-hover:translate-x-0.5 transition-transform">
          ›
        </span>
      </div>
    </div>
  )
}

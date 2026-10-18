import React from "react"
import { Connection } from "../types"
import { X, Copy, Check, CheckCircle2, AlertTriangle, Power, Clock, Shield } from "lucide-react"

interface ProviderDrawerProps {
  connection: Connection | null
  onClose: () => void
  onToggleActive: (id: string, nextState: boolean) => Promise<void>
}

export const ProviderDrawer: React.FC<ProviderDrawerProps> = ({
  connection,
  onClose,
  onToggleActive
}) => {
  const [copied, setCopied] = React.useState(false)
  const [isToggling, setIsToggling] = React.useState(false)

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [onClose])

  if (!connection) return null

  const handleCopyJson = () => {
    navigator.clipboard.writeText(JSON.stringify(connection, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleToggle = async () => {
    setIsToggling(true)
    try {
      await onToggleActive(connection.id, !connection.isActive)
    } finally {
      setIsToggling(false)
    }
  }

  const hasError = connection.testStatus === "error" || connection.lastError || (connection.errorCode && connection.errorCode >= 400)

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/60 backdrop-blur-xs">
      <div className="w-full max-w-lg bg-[#0c0a09] border-l border-[#27272a] h-full flex flex-col shadow-2xl">
        <div className="flex items-center justify-between border-b border-[#27272a] px-5 py-4 bg-[#121214]">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-mono font-bold text-xs uppercase">
              {connection.provider.slice(0, 2)}
            </div>
            <div>
              <h3 className="text-sm font-semibold text-white capitalize">{connection.provider}</h3>
              <p className="text-[11px] font-mono text-[#71717a]">{connection.id}</p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="rounded-md p-1.5 text-[#a1a1aa] hover:bg-[#27272a] hover:text-white transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-5 text-xs">
          <div className="flex items-center justify-between rounded-lg border border-[#27272a] bg-[#121214] p-3.5">
            <div>
              <span className="text-white font-medium block">Status Operasi</span>
              <span className="text-[11px] text-[#a1a1aa]">
                {connection.isActive ? "Koneksi aktif dan digunakan dalam failover gateway" : "Koneksi dinonaktifkan"}
              </span>
            </div>

            <button
              onClick={handleToggle}
              disabled={isToggling}
              className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-xs font-mono font-medium transition-colors ${
                connection.isActive
                  ? "bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/20"
                  : "bg-[#27272a] text-[#a1a1aa] hover:text-white"
              }`}
            >
              <Power className="h-3.5 w-3.5" />
              <span>{connection.isActive ? "Aktif" : "Mati"}</span>
            </button>
          </div>

          <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4 space-y-3 font-mono text-[11px]">
            <h4 className="text-xs font-sans font-semibold text-white mb-2">Informasi Akun</h4>

            <div className="flex justify-between py-1 border-b border-[#27272a]/40">
              <span className="text-[#a1a1aa]">Nama:</span>
              <span className="text-white">{connection.name || "-"}</span>
            </div>

            <div className="flex justify-between py-1 border-b border-[#27272a]/40">
              <span className="text-[#a1a1aa]">Email:</span>
              <span className="text-white">{connection.email || "-"}</span>
            </div>

            <div className="flex justify-between py-1 border-b border-[#27272a]/40">
              <span className="text-[#a1a1aa]">Tipe Autentikasi:</span>
              <span className="text-emerald-400 uppercase">{connection.authType || "apikey"}</span>
            </div>

            <div className="flex justify-between py-1 border-b border-[#27272a]/40">
              <span className="text-[#a1a1aa]">Prioritas Routing:</span>
              <span className="text-white">P{connection.priority ?? 1}</span>
            </div>

            <div className="flex justify-between py-1 border-b border-[#27272a]/40">
              <span className="text-[#a1a1aa]">Dibuat Pada:</span>
              <span className="text-[#a1a1aa]">{connection.createdAt ? new Date(connection.createdAt).toLocaleString() : "-"}</span>
            </div>

            <div className="flex justify-between py-1">
              <span className="text-[#a1a1aa]">Terakhir Digunakan:</span>
              <span className="text-[#a1a1aa]">{connection.lastUsedAt ? new Date(connection.lastUsedAt).toLocaleString() : "Belum pernah"}</span>
            </div>
          </div>

          {hasError && (
            <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-4 space-y-2">
              <div className="flex items-center gap-2 text-amber-400 font-medium">
                <AlertTriangle className="h-4 w-4" />
                <span>Laporan Kesehatan & Error</span>
              </div>
              <p className="text-[11px] font-mono text-amber-200/90 break-words bg-[#121214] p-2.5 rounded-md border border-[#27272a]">
                {connection.lastError || "Terjadi kesalahan pada koneksi"}
              </p>
              {connection.lastErrorAt && (
                <div className="text-[10px] font-mono text-[#a1a1aa]">
                  Waktu error: {new Date(connection.lastErrorAt).toLocaleString()}
                </div>
              )}
            </div>
          )}

          {connection.providerSpecificData && Object.keys(connection.providerSpecificData).length > 0 && (
            <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4 space-y-2 font-mono">
              <h4 className="text-xs font-sans font-semibold text-white">Metadata Spesifik Provider</h4>
              <pre className="text-[10px] text-emerald-300 bg-[#09090b] p-3 rounded-md overflow-x-auto border border-[#27272a]">
                {JSON.stringify(connection.providerSpecificData, null, 2)}
              </pre>
            </div>
          )}
        </div>

        <div className="border-t border-[#27272a] p-4 bg-[#121214] flex items-center justify-between">
          <button
            onClick={handleCopyJson}
            className="flex items-center gap-1.5 rounded-md border border-[#27272a] bg-[#18181b] px-3 py-1.5 text-xs text-[#a1a1aa] hover:text-white transition-colors"
          >
            {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            <span>{copied ? "Tersalin" : "Salin Raw JSON"}</span>
          </button>

          <button
            onClick={onClose}
            className="rounded-md bg-[#27272a] hover:bg-[#3f3f46] px-4 py-1.5 text-xs font-medium text-white transition-colors"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  )
}

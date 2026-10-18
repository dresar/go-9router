import React from "react"
import { Clock, AlertTriangle, CheckCircle2 } from "lucide-react"

interface StatusBadgeProps {
  status: "ready" | "cooldown" | "error" | "inactive"
  label?: string
  errorCount?: number
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  label,
  errorCount
}) => {
  if (status === "cooldown") {
    return (
      <span className="inline-flex items-center gap-1 rounded bg-amber-500/15 border border-amber-500/30 px-1.5 py-0.5 text-[10px] font-mono text-amber-400">
        <Clock className="h-2.5 w-2.5" />
        <span>{label || "⏱ Cooldown"}</span>
      </span>
    )
  }

  if (status === "error") {
    return (
      <span className="inline-flex items-center gap-1 rounded bg-rose-500/15 border border-rose-500/30 px-1.5 py-0.5 text-[10px] font-mono text-rose-400">
        <AlertTriangle className="h-2.5 w-2.5" />
        <span>{label || (errorCount ? `△ ${errorCount} Isu` : "△ Error")}</span>
      </span>
    )
  }

  if (status === "ready") {
    return (
      <span className="inline-flex items-center gap-1 rounded bg-emerald-500/15 border border-emerald-500/30 px-1.5 py-0.5 text-[10px] font-mono text-emerald-400">
        <CheckCircle2 className="h-2.5 w-2.5" />
        <span>{label || "⊙ Ready"}</span>
      </span>
    )
  }

  return (
    <span className="inline-flex items-center gap-1 rounded bg-zinc-500/15 border border-zinc-500/30 px-1.5 py-0.5 text-[10px] font-mono text-zinc-400">
      <span>{label || "○ Inactive"}</span>
    </span>
  )
}

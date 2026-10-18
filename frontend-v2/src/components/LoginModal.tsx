import React from "react"
import { Lock, Eye, EyeOff, AlertCircle, X, ShieldAlert } from "lucide-react"

interface LoginModalProps {
  isOpen: boolean
  onClose: () => void
  onLogin: (password: string) => Promise<boolean>
  isDismissible?: boolean
}

export const LoginModal: React.FC<LoginModalProps> = ({
  isOpen,
  onClose,
  onLogin,
  isDismissible = true
}) => {
  const [password, setPassword] = React.useState("")
  const [showPassword, setShowPassword] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = React.useState(false)

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!password) {
      setError("Password tidak boleh kosong")
      return
    }

    setIsSubmitting(true)
    setError(null)

    try {
      const success = await onLogin(password)
      if (success) {
        setPassword("")
        onClose()
      } else {
        setError("Password salah. Silakan coba lagi.")
      }
    } catch (err: any) {
      setError(err?.message || "Gagal melakukan autentikasi")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-xs">
      <div className="w-full max-w-sm rounded-lg border border-[#27272a] bg-[#121214] p-6 shadow-2xl relative">
        {isDismissible && (
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-[#a1a1aa] hover:text-white transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        )}

        <div className="flex items-center gap-3 mb-4">
          <div className="flex h-10 w-10 items-center justify-center rounded-md bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
            <Lock className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-sm font-semibold text-white">Autentikasi Dashboard</h2>
            <p className="text-xs text-[#a1a1aa]">Masukkan password master 9Router</p>
          </div>
        </div>

        {error && (
          <div className="mb-4 flex items-center gap-2 rounded-md bg-rose-500/10 border border-rose-500/30 p-2.5 text-xs text-rose-400">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-[#a1a1aa] mb-1.5">
              Password Admin
            </label>
            <div className="relative">
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                autoFocus
                className="w-full rounded-md border border-[#27272a] bg-[#18181b] px-3 py-2 pr-10 text-xs text-white placeholder-[#52525b] focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 font-mono"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 flex items-center pr-3 text-[#71717a] hover:text-white"
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="flex-1 rounded-md bg-emerald-600 hover:bg-emerald-500 px-4 py-2 text-xs font-medium text-white transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {isSubmitting ? (
                <>
                  <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent" />
                  <span>Memverifikasi...</span>
                </>
              ) : (
                <span>Masuk Dashboard</span>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

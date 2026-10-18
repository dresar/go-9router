import React from "react"
import { ModelItem, ApiKey } from "../types"
import { api } from "../api/client"
import { Send, Play, Copy, Check, Clock, RotateCcw, AlertCircle, Sparkles } from "lucide-react"

interface PlaygroundViewProps {
  models: ModelItem[]
  selectedModel: string
  onSelectModel: (model: string) => void
  apiKeys: ApiKey[]
}

export const PlaygroundView: React.FC<PlaygroundViewProps> = ({
  models,
  selectedModel,
  onSelectModel,
  apiKeys
}) => {
  const [prompt, setPrompt] = React.useState("Tulis ringkasan 1 kalimat tentang apa itu reverse proxy gateway AI.")
  const [temperature, setTemperature] = React.useState(0.7)
  const [maxTokens, setMaxTokens] = React.useState(300)
  const [response, setResponse] = React.useState<string | null>(null)
  const [latency, setLatency] = React.useState<number | null>(null)
  const [isLoading, setIsLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [copied, setCopied] = React.useState(false)

  const activeKey = apiKeys.find((k) => k.isActive)?.key || ""

  const handleSend = async () => {
    if (!prompt.trim() || isLoading) return
    setIsLoading(true)
    setError(null)

    try {
      const res = await api.chatCompletion({
        model: selectedModel || "HERMES",
        prompt: prompt.trim(),
        apiKey: activeKey,
        temperature,
        maxTokens
      })
      setResponse(res.content)
      setLatency(res.latencyMs)
    } catch (err: any) {
      setError(err?.message || "Gagal mendapatkan respons")
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault()
      handleSend()
    }
  }

  const handleCopy = () => {
    if (!response) return
    navigator.clipboard.writeText(response)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-4">
      <div className="border-b border-[#27272a] pb-4">
        <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
          <span>AI Playground Gateway</span>
        </h2>
        <p className="text-xs text-[#a1a1aa] mt-0.5">
          Uji coba prompt langsung ke gateway 9Router dengan model pilihan
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <div className="lg:col-span-5 space-y-4">
          <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4 space-y-3">
            <div>
              <label className="block text-xs font-medium text-[#a1a1aa] mb-1">
                Model Tujuan
              </label>
              <select
                value={selectedModel}
                onChange={(e) => onSelectModel(e.target.value)}
                className="w-full rounded-md border border-[#27272a] bg-[#18181b] px-3 py-1.5 text-xs text-white focus:border-emerald-500 focus:outline-none font-mono"
              >
                <option value="HERMES">HERMES (Smart Combo)</option>
                <option value="MY-COMBO">MY-COMBO (Smart Combo)</option>
                <option value="cli-combo">cli-combo (Smart Combo)</option>
                {models.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.id}
                  </option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-2 gap-3 pt-2 border-t border-[#27272a]">
              <div>
                <label className="block text-xs font-mono text-[#a1a1aa] mb-1">
                  Temp: {temperature}
                </label>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.1"
                  value={temperature}
                  onChange={(e) => setTemperature(parseFloat(e.target.value))}
                  className="w-full accent-emerald-500"
                />
              </div>

              <div>
                <label className="block text-xs font-mono text-[#a1a1aa] mb-1">
                  Max Tokens: {maxTokens}
                </label>
                <input
                  type="number"
                  min="50"
                  max="4000"
                  step="50"
                  value={maxTokens}
                  onChange={(e) => setMaxTokens(parseInt(e.target.value) || 300)}
                  className="w-full rounded-md border border-[#27272a] bg-[#18181b] px-2.5 py-1 text-xs text-white font-mono"
                />
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <label className="text-xs font-medium text-white">Prompt Input</label>
              <span className="text-[10px] font-mono text-[#71717a]">Ctrl + Enter untuk Kirim</span>
            </div>

            <textarea
              rows={6}
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ketik instruksi atau pertanyaan di sini..."
              className="w-full rounded-md border border-[#27272a] bg-[#09090b] p-3 text-xs text-white placeholder-[#71717a] focus:border-emerald-500 focus:outline-none font-mono resize-none"
            />

            <button
              onClick={handleSend}
              disabled={isLoading || !prompt.trim()}
              className="w-full flex items-center justify-center gap-2 rounded-md bg-emerald-600 hover:bg-emerald-500 py-2 px-4 text-xs font-medium text-white transition-colors disabled:opacity-50"
            >
              {isLoading ? (
                <>
                  <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white border-t-transparent" />
                  <span>Mengirim ke Gateway...</span>
                </>
              ) : (
                <>
                  <Send className="h-3.5 w-3.5" />
                  <span>Kirim Permintaan</span>
                </>
              )}
            </button>
          </div>
        </div>

        <div className="lg:col-span-7 flex flex-col">
          <div className="flex-1 rounded-lg border border-[#27272a] bg-[#121214] flex flex-col overflow-hidden min-h-[350px]">
            <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/60">
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono font-medium text-white">Respons Model</span>
                {latency !== null && (
                  <span className="flex items-center gap-1 rounded bg-[#27272a] px-2 py-0.5 text-[10px] font-mono text-emerald-400">
                    <Clock className="h-3 w-3" />
                    {latency}ms
                  </span>
                )}
              </div>

              {response && (
                <button
                  onClick={handleCopy}
                  className="flex items-center gap-1 text-[11px] font-mono text-[#a1a1aa] hover:text-white"
                >
                  {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                  <span>{copied ? "Tersalin" : "Salin"}</span>
                </button>
              )}
            </div>

            <div className="flex-1 p-4 overflow-y-auto font-mono text-xs text-zinc-200 bg-[#09090b]/40">
              {error ? (
                <div className="flex items-start gap-2.5 rounded-md bg-rose-500/10 border border-rose-500/30 p-3 text-rose-400">
                  <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
                  <div className="break-all whitespace-pre-wrap">{error}</div>
                </div>
              ) : response ? (
                <div className="whitespace-pre-wrap leading-relaxed">{response}</div>
              ) : isLoading ? (
                <div className="flex items-center justify-center h-full text-[#71717a] gap-2">
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                  <span>Menunggu token dari backend...</span>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-full text-[#71717a] text-center p-6">
                  <Play className="h-6 w-6 mb-2 opacity-40 text-emerald-400" />
                  <p>Pilih model dan klik kirim untuk menguji throughput dan latency secara langsung.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

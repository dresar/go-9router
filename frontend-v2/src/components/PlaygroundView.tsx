import React from "react"
import { ModelItem, ApiKey } from "../types"
import { api } from "../api/client"
import {
  Send,
  Play,
  Copy,
  Check,
  Clock,
  RotateCcw,
  AlertCircle,
  Square,
  Sparkles
} from "lucide-react"

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
  const [maxTokens, setMaxTokens] = React.useState(400)
  const [response, setResponse] = React.useState<string>("")
  const [latency, setLatency] = React.useState<number | null>(null)
  const [isStreaming, setIsStreaming] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [copied, setCopied] = React.useState(false)
  const abortControllerRef = React.useRef<AbortController | null>(null)

  const activeKey = apiKeys.find((k) => k.isActive)?.key || ""

  const handleSend = async () => {
    if (!prompt.trim() || isStreaming) return
    setIsStreaming(true)
    setError(null)
    setResponse("")
    setLatency(null)

    const controller = new AbortController()
    abortControllerRef.current = controller

    try {
      const res = await api.chatCompletionStream({
        model: selectedModel || "HERMES",
        prompt: prompt.trim(),
        apiKey: activeKey,
        temperature,
        maxTokens,
        onChunk: (chunk) => {
          setResponse(chunk)
        },
        signal: controller.signal
      })
      setLatency(res.latencyMs)
    } catch (err: any) {
      if (err.name !== "AbortError") {
        setError(err?.message || "Gagal mendapatkan respons")
      }
    } finally {
      setIsStreaming(false)
      abortControllerRef.current = null
    }
  }

  const handleStop = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      setIsStreaming(false)
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

  const handleReset = () => {
    setPrompt("Tulis ringkasan 1 kalimat tentang apa itu reverse proxy gateway AI.")
    setResponse("")
    setError(null)
    setLatency(null)
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#27272a] pb-4">
        <div>
          <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
            <span>AI Playground Gateway</span>
            <span className="rounded bg-emerald-500/10 border border-emerald-500/30 px-2 py-0.5 text-xs font-mono text-emerald-400">
              Live Stream
            </span>
          </h2>
          <p className="text-xs text-[#a1a1aa] mt-0.5">
            Uji coba streaming prompt langsung ke gateway 9Router dengan latency timer monospace
          </p>
        </div>

        <button
          onClick={handleReset}
          className="flex items-center gap-1.5 rounded-md border border-[#27272a] bg-[#121214] px-3 py-1.5 text-xs text-[#a1a1aa] hover:text-white transition-colors font-mono self-start sm:self-auto"
        >
          <RotateCcw className="h-3.5 w-3.5" />
          <span>Reset Prompt</span>
        </button>
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
                <optgroup label="Smart Combo Routers">
                  <option value="HERMES">HERMES (Smart Combo)</option>
                  <option value="MY-COMBO">MY-COMBO (Smart Combo)</option>
                  <option value="cli-combo">cli-combo (Smart Combo)</option>
                </optgroup>
                <optgroup label="Direct Gateway Models">
                  {models.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.id}
                    </option>
                  ))}
                </optgroup>
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

            <div className="flex items-center gap-2">
              {isStreaming ? (
                <button
                  type="button"
                  onClick={handleStop}
                  className="w-full flex items-center justify-center gap-2 rounded-md bg-rose-600 hover:bg-rose-500 py-2 px-4 text-xs font-medium text-white transition-colors"
                >
                  <Square className="h-3.5 w-3.5 fill-current" />
                  <span>Hentikan Streaming</span>
                </button>
              ) : (
                <button
                  type="button"
                  onClick={handleSend}
                  disabled={!prompt.trim()}
                  className="w-full flex items-center justify-center gap-2 rounded-md bg-emerald-600 hover:bg-emerald-500 py-2 px-4 text-xs font-medium text-white transition-colors disabled:opacity-50"
                >
                  <Send className="h-3.5 w-3.5" />
                  <span>Kirim Streaming</span>
                </button>
              )}
            </div>
          </div>
        </div>

        <div className="lg:col-span-7 flex flex-col">
          <div className="flex-1 rounded-lg border border-[#27272a] bg-[#121214] flex flex-col overflow-hidden min-h-[380px]">
            <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/60">
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono font-medium text-white">Respons Model</span>
                {isStreaming && (
                  <span className="flex items-center gap-1 rounded bg-emerald-500/10 border border-emerald-500/30 px-2 py-0.5 text-[10px] font-mono text-emerald-400">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    <span>Streaming...</span>
                  </span>
                )}
                {latency !== null && (
                  <span className="flex items-center gap-1 rounded bg-[#27272a] px-2 py-0.5 text-[10px] font-mono text-emerald-400">
                    <Clock className="h-3 w-3" />
                    <span>{latency}ms</span>
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

            <div className="flex-1 p-4 overflow-y-auto font-mono text-xs text-zinc-200 bg-[#09090b]/60">
              {error ? (
                <div className="flex items-start gap-2.5 rounded-md bg-rose-500/10 border border-rose-500/30 p-3 text-rose-400">
                  <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
                  <div className="break-all whitespace-pre-wrap">{error}</div>
                </div>
              ) : response ? (
                <div className="whitespace-pre-wrap leading-relaxed">
                  {response}
                  {isStreaming && (
                    <span className="inline-block h-4 w-2 bg-emerald-400 ml-1 animate-pulse align-middle" />
                  )}
                </div>
              ) : isStreaming ? (
                <div className="flex items-center justify-center h-full text-[#71717a] gap-2">
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                  <span>Menghubungkan ke gateway model...</span>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-full text-[#71717a] text-center p-6">
                  <Play className="h-6 w-6 mb-2 opacity-40 text-emerald-400" />
                  <p>Pilih model dan klik kirim streaming untuk menguji throughput dan latency secara langsung.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

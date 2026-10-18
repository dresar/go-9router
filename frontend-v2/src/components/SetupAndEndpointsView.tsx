import React from "react"
import { ApiKey } from "../types"
import { Copy, Check, Terminal, Key, Globe, Shield, Code2, Eye, EyeOff } from "lucide-react"

interface SetupAndEndpointsViewProps {
  apiKeys: ApiKey[]
}

export const SetupAndEndpointsView: React.FC<SetupAndEndpointsViewProps> = ({ apiKeys }) => {
  const [copiedKey, setCopiedKey] = React.useState(false)
  const [copiedSection, setCopiedSection] = React.useState<string | null>(null)
  const [showKey, setShowKey] = React.useState(false)

  const activeKey = apiKeys.find((k) => k.isActive)?.key || "sk-7016dd129191d903-u8emvi-c0b721b9"

  const copyToClipboard = (text: string, sectionId: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSection(sectionId)
    setTimeout(() => setCopiedSection(null), 2000)
  }

  const curlSnippet = `curl -X POST https://9router.agentai.my.id/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${activeKey}" \\
  -d '{
    "model": "HERMES",
    "messages": [{"role": "user", "content": "Halo 9Router!"}]
  }'`

  const pythonSnippet = `from openai import OpenAI

client = OpenAI(
    base_url="https://9router.agentai.my.id/v1",
    api_key="${activeKey}"
)

response = client.chat.completions.create(
    model="HERMES",
    messages=[{"role": "user", "content": "Halo 9Router!"}]
)

print(response.choices[0].message.content)`

  const nodeSnippet = `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "https://9router.agentai.my.id/v1",
  apiKey: "${activeKey}"
});

const response = await client.chat.completions.create({
  model: "HERMES",
  messages: [{ role: "user", content: "Halo 9Router!" }]
});

console.log(response.choices[0].message.content);`

  const envSnippet = `OPENAI_BASE_URL="https://9router.agentai.my.id/v1"
OPENAI_API_KEY="${activeKey}"`

  return (
    <div className="space-y-6">
      <div className="border-b border-[#27272a] pb-4">
        <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
          <span>Konektivitas & Kunci Akses Gateway</span>
        </h2>
        <p className="text-xs text-[#a1a1aa] mt-0.5">
          Spesifikasi endpoint OpenAI-compatible dan konfigurasi kredensial klien
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-emerald-400 font-mono text-xs mb-1">
            <Globe className="h-4 w-4" />
            <span className="font-semibold">Endpoint Publik (HTTPS)</span>
          </div>
          <div className="mt-2 flex items-center justify-between rounded bg-[#09090b] border border-[#27272a] px-3 py-2 font-mono text-xs text-white">
            <span className="truncate">https://9router.agentai.my.id/v1</span>
            <button
              onClick={() => copyToClipboard("https://9router.agentai.my.id/v1", "ep-pub")}
              className="text-[#a1a1aa] hover:text-white ml-2 shrink-0"
            >
              {copiedSection === "ep-pub" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            </button>
          </div>
        </div>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-zinc-300 font-mono text-xs mb-1">
            <Terminal className="h-4 w-4 text-emerald-400" />
            <span className="font-semibold">Port 20128 (Auth Protected)</span>
          </div>
          <div className="mt-2 flex items-center justify-between rounded bg-[#09090b] border border-[#27272a] px-3 py-2 font-mono text-xs text-white">
            <span className="truncate">http://localhost:20128/v1</span>
            <button
              onClick={() => copyToClipboard("http://localhost:20128/v1", "ep-local")}
              className="text-[#a1a1aa] hover:text-white ml-2 shrink-0"
            >
              {copiedSection === "ep-local" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            </button>
          </div>
        </div>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
          <div className="flex items-center gap-2 text-emerald-400 font-mono text-xs mb-1">
            <Shield className="h-4 w-4" />
            <span className="font-semibold">Port 20129 (Direct LAN / No Key)</span>
          </div>
          <div className="mt-2 flex items-center justify-between rounded bg-[#09090b] border border-[#27272a] px-3 py-2 font-mono text-xs text-white">
            <span className="truncate">http://localhost:20129/v1</span>
            <button
              onClick={() => copyToClipboard("http://localhost:20129/v1", "ep-lan")}
              className="text-[#a1a1aa] hover:text-white ml-2 shrink-0"
            >
              {copiedSection === "ep-lan" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            </button>
          </div>
        </div>
      </div>

      <div className="rounded-lg border border-[#27272a] bg-[#121214] p-4">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <Key className="h-4 w-4 text-emerald-400" />
            <h3 className="text-xs font-semibold text-white">Master API Key Gateway</h3>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowKey(!showKey)}
              className="flex items-center gap-1 rounded border border-[#27272a] bg-[#18181b] px-2 py-1 text-xs text-[#a1a1aa] hover:text-white transition-colors"
            >
              {showKey ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
              <span>{showKey ? "Sembunyikan" : "Tampilkan"}</span>
            </button>
            <button
              onClick={() => {
                navigator.clipboard.writeText(activeKey)
                setCopiedKey(true)
                setTimeout(() => setCopiedKey(false), 2000)
              }}
              className="flex items-center gap-1.5 rounded bg-emerald-600 hover:bg-emerald-500 px-3 py-1 text-xs font-medium text-white transition-colors"
            >
              {copiedKey ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
              <span>{copiedKey ? "Tersalin" : "Salin Kunci"}</span>
            </button>
          </div>
        </div>

        <div className="rounded bg-[#09090b] border border-[#27272a] p-3 font-mono text-xs text-emerald-400 break-all select-all">
          {showKey ? activeKey : activeKey.replace(/^(sk-[a-f0-9]{8}).*([a-f0-9]{8})$/, "$1••••••••••••$2")}
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-semibold text-white">Panduan Integrasi Cepat</h3>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden">
          <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/50">
            <span className="text-xs font-mono font-medium text-zinc-300">cURL (Command Line)</span>
            <button
              onClick={() => copyToClipboard(curlSnippet, "curl")}
              className="flex items-center gap-1 text-[11px] text-[#a1a1aa] hover:text-white"
            >
              {copiedSection === "curl" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
              <span>Salin cURL</span>
            </button>
          </div>
          <pre className="p-4 text-xs font-mono text-zinc-300 overflow-x-auto bg-[#09090b]">
            {curlSnippet}
          </pre>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden">
            <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/50">
              <span className="text-xs font-mono font-medium text-zinc-300">Python (OpenAI SDK)</span>
              <button
                onClick={() => copyToClipboard(pythonSnippet, "py")}
                className="flex items-center gap-1 text-[11px] text-[#a1a1aa] hover:text-white"
              >
                {copiedSection === "py" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                <span>Salin</span>
              </button>
            </div>
            <pre className="p-4 text-xs font-mono text-zinc-300 overflow-x-auto bg-[#09090b]">
              {pythonSnippet}
            </pre>
          </div>

          <div className="rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden">
            <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/50">
              <span className="text-xs font-mono font-medium text-zinc-300">Node.js (TypeScript)</span>
              <button
                onClick={() => copyToClipboard(nodeSnippet, "ts")}
                className="flex items-center gap-1 text-[11px] text-[#a1a1aa] hover:text-white"
              >
                {copiedSection === "ts" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
                <span>Salin</span>
              </button>
            </div>
            <pre className="p-4 text-xs font-mono text-zinc-300 overflow-x-auto bg-[#09090b]">
              {nodeSnippet}
            </pre>
          </div>
        </div>

        <div className="rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden">
          <div className="flex items-center justify-between border-b border-[#27272a] px-4 py-2.5 bg-[#18181b]/50">
            <span className="text-xs font-mono font-medium text-zinc-300">Environment Variables (.env)</span>
            <button
              onClick={() => copyToClipboard(envSnippet, "env")}
              className="flex items-center gap-1 text-[11px] text-[#a1a1aa] hover:text-white"
            >
              {copiedSection === "env" ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
              <span>Salin .env</span>
            </button>
          </div>
          <pre className="p-4 text-xs font-mono text-zinc-300 overflow-x-auto bg-[#09090b]">
            {envSnippet}
          </pre>
        </div>
      </div>
    </div>
  )
}

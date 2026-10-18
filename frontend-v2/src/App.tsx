import React from "react"
import { Header } from "./components/Header"
import { ProviderOverviewGrid } from "./components/ProviderOverviewGrid"
import { ConnectionTable } from "./components/ConnectionTable"
import { ProviderDrawer } from "./components/ProviderDrawer"
import { ModelsAndCombosView } from "./components/ModelsAndCombosView"
import { SetupAndEndpointsView } from "./components/SetupAndEndpointsView"
import { PlaygroundView } from "./components/PlaygroundView"
import { UsageStatsView } from "./components/UsageStatsView"
import { LoginModal } from "./components/LoginModal"
import { api } from "./api/client"
import { Connection, ModelItem, Combo, CustomModel, ApiKey, UsageStats } from "./types"
import { AlertCircle, RefreshCw } from "lucide-react"

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = React.useState("overview")
  const [connections, setConnections] = React.useState<Connection[]>([])
  const [models, setModels] = React.useState<ModelItem[]>([])
  const [customModels, setCustomModels] = React.useState<CustomModel[]>([])
  const [combos, setCombos] = React.useState<Combo[]>([])
  const [apiKeys, setApiKeys] = React.useState<ApiKey[]>([])
  const [usageStats, setUsageStats] = React.useState<UsageStats | null>(null)

  const [requireLogin, setRequireLogin] = React.useState(false)
  const [authenticated, setAuthenticated] = React.useState(false)
  const [loginModalOpen, setLoginModalOpen] = React.useState(false)

  const [selectedConnection, setSelectedConnection] = React.useState<Connection | null>(null)
  const [providerFilter, setProviderFilter] = React.useState("all")
  const [playgroundModel, setPlaygroundModel] = React.useState("HERMES")

  const [isLoading, setIsLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const loadAllData = React.useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const [reqLoginRes, authRes] = await Promise.all([
        api.getRequireLogin(),
        api.getAuthStatus()
      ])

      const mustLogin = reqLoginRes.requireLogin && !authRes.authenticated
      setRequireLogin(reqLoginRes.requireLogin)
      setAuthenticated(authRes.authenticated)

      if (mustLogin) {
        setLoginModalOpen(true)
        setIsLoading(false)
        return
      }

      const [connsRes, modelsRes, customModelsRes, combosRes, keysRes, statsRes] = await Promise.allSettled([
        api.getConnections(),
        api.getModels(),
        api.getCustomModels(),
        api.getCombos(),
        api.getApiKeys(),
        api.getUsageStats()
      ])

      if (connsRes.status === "fulfilled") {
        setConnections(connsRes.value)
      } else if ((connsRes.reason as any)?.status === 401) {
        setRequireLogin(true)
        setAuthenticated(false)
        setLoginModalOpen(true)
        setIsLoading(false)
        return
      }

      if (modelsRes.status === "fulfilled") {
        setModels(modelsRes.value)
      }

      if (customModelsRes.status === "fulfilled") {
        setCustomModels(customModelsRes.value)
      }

      if (combosRes.status === "fulfilled") {
        setCombos(combosRes.value)
      }

      if (keysRes.status === "fulfilled") {
        setApiKeys(keysRes.value)
      }

      if (statsRes.status === "fulfilled") {
        setUsageStats(statsRes.value)
      }
    } catch (err: any) {
      setError(err?.message || "Gagal memuat data gateway")
    } finally {
      setIsLoading(false)
    }
  }, [])

  React.useEffect(() => {
    loadAllData()
  }, [loadAllData])

  const handleToggleConnection = async (id: string, nextState: boolean) => {
    const prevConnections = [...connections]
    setConnections((prev) =>
      prev.map((c) => (c.id === id ? { ...c, isActive: nextState } : c))
    )
    if (selectedConnection && selectedConnection.id === id) {
      setSelectedConnection({ ...selectedConnection, isActive: nextState })
    }

    try {
      const updated = await api.updateConnectionStatus(id, nextState)
      if (updated) {
        setConnections((prev) =>
          prev.map((c) => (c.id === id ? { ...c, ...updated } : c))
        )
      }
    } catch (err: any) {
      setConnections(prevConnections)
      alert(`Gagal mengubah status: ${err?.message || "Kesalahan server"}`)
    }
  }

  const handleLogin = async (password: string): Promise<boolean> => {
    const success = await api.login(password)
    if (success) {
      setAuthenticated(true)
      setLoginModalOpen(false)
      loadAllData()
      return true
    }
    return false
  }

  const handleLogout = async () => {
    await api.logout()
    setAuthenticated(false)
    if (requireLogin) {
      setLoginModalOpen(true)
    }
  }

  const handleSelectProviderFromOverview = (provider: string) => {
    setProviderFilter(provider)
    setActiveTab("connections")
  }

  const handleSelectModelForPlayground = (modelId: string) => {
    setPlaygroundModel(modelId)
    setActiveTab("playground")
  }

  const activeCount = connections.filter((c) => c.isActive).length

  return (
    <div className="min-h-screen bg-[#09090b] text-[#f4f4f5] flex flex-col font-sans">
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        totalConnections={connections.length}
        activeConnections={activeCount}
        totalModels={models.length}
        authenticated={authenticated}
        requireLogin={requireLogin}
        onOpenLogin={() => setLoginModalOpen(true)}
        onLogout={handleLogout}
        onRefresh={loadAllData}
        isLoading={isLoading}
      />

      <main className="flex-1 mx-auto w-full max-w-7xl px-4 py-6 sm:px-6">
        {error && (
          <div className="mb-6 flex items-center justify-between rounded-lg border border-rose-500/30 bg-rose-500/10 p-3.5 text-xs text-rose-300">
            <div className="flex items-center gap-2">
              <AlertCircle className="h-4 w-4 shrink-0 text-rose-400" />
              <span>{error}</span>
            </div>
            <button
              onClick={loadAllData}
              className="flex items-center gap-1 font-mono text-white underline hover:no-underline"
            >
              <RefreshCw className="h-3 w-3" />
              <span>Coba Lagi</span>
            </button>
          </div>
        )}

        {isLoading && connections.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <span className="h-8 w-8 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent mb-3" />
            <p className="text-xs font-mono text-[#a1a1aa]">Memuat koneksi gateway 9Router...</p>
          </div>
        ) : (
          <>
            {activeTab === "overview" && (
              <ProviderOverviewGrid
                connections={connections}
                onSelectProvider={handleSelectProviderFromOverview}
              />
            )}

            {activeTab === "connections" && (
              <ConnectionTable
                connections={connections}
                initialProviderFilter={providerFilter}
                onToggleActive={handleToggleConnection}
                onSelectConnection={(c) => setSelectedConnection(c)}
              />
            )}

            {activeTab === "models" && (
              <ModelsAndCombosView
                models={models}
                combos={combos}
                customModels={customModels}
                onSelectModelForPlayground={handleSelectModelForPlayground}
              />
            )}

            {activeTab === "setup" && (
              <SetupAndEndpointsView apiKeys={apiKeys} />
            )}

            {activeTab === "playground" && (
              <PlaygroundView
                models={models}
                selectedModel={playgroundModel}
                onSelectModel={setPlaygroundModel}
                apiKeys={apiKeys}
              />
            )}

            {activeTab === "usage" && (
              <UsageStatsView stats={usageStats} />
            )}
          </>
        )}
      </main>

      <footer className="border-t border-[#27272a] bg-[#0c0a09] py-4 text-center text-xs font-mono text-[#71717a]">
        <div className="mx-auto max-w-7xl px-4 flex flex-col sm:flex-row items-center justify-between gap-2">
          <span>Go 9Router • High Performance AI Gateway</span>
          <span>189 Provider Terdaftar • 108 Aktif • 94 Model</span>
        </div>
      </footer>

      <ProviderDrawer
        connection={selectedConnection}
        onClose={() => setSelectedConnection(null)}
        onToggleActive={handleToggleConnection}
      />

      <LoginModal
        isOpen={loginModalOpen}
        onClose={() => setLoginModalOpen(false)}
        onLogin={handleLogin}
        isDismissible={!requireLogin || authenticated}
      />
    </div>
  )
}
export default App

export interface Connection {
  id: string
  provider: string
  authType: string
  name?: string
  email?: string
  displayName?: string
  priority: number
  isActive: boolean
  createdAt: string
  updatedAt: string
  testStatus?: string
  lastError?: string
  errorCode?: number
  lastErrorAt?: string
  lastUsedAt?: string
  proxyPoolId?: string
  apiKey?: string
  rateLimitedUntil?: string
  backoffLevel?: number
  providerSpecificData?: Record<string, any>
  [key: string]: any
}

export interface ModelItem {
  id: string
  object: string
  owned_by?: string
  context_length?: number
  context_window?: number
}

export interface Combo {
  id: string
  name: string
  kind?: string
  models: string[]
  createdAt?: string
  updatedAt?: string
}

export interface CustomModel {
  id: string
  providerAlias?: string
  type?: string
  name?: string
  [key: string]: any
}

export interface ApiKey {
  id: string
  key: string
  name: string
  machineId?: string
  isActive: boolean
  createdAt: string
}

export interface UsageStats {
  summary?: {
    totalRequests: number
    promptTokens: number
    completionTokens: number
    cachedTokens?: number
    estimatedCost?: number
  }
  recentRequests?: Array<{
    timestamp: string
    model: string
    provider: string
    promptTokens: number
    completionTokens: number
    cachedTokens: number
    status: string
  }>
  byModel?: Record<string, any>
  byProvider?: Record<string, any>
}

export interface RequireLoginResponse {
  requireLogin: boolean
  tailscaleUrl?: string
  tunnelDashboardAccess?: boolean
  tunnelUrl?: string
}

export interface AuthStatusResponse {
  authenticated: boolean
  requireLogin?: boolean
}

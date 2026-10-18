import type {
  Connection,
  ModelItem,
  Combo,
  CustomModel,
  ApiKey,
  UsageStats,
  RequireLoginResponse,
  AuthStatusResponse
} from "../types"

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(path, {
    credentials: "same-origin",
    headers: {
      "Accept": "application/json",
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers
    },
    ...options
  })

  if (!res.ok) {
    let errorMsg = `HTTP ${res.status}`
    try {
      const errJson = await res.json()
      errorMsg = errJson.error || errJson.message || errorMsg
    } catch {}
    throw new ApiError(res.status, errorMsg)
  }

  return res.json() as Promise<T>
}

export const api = {
  async getRequireLogin(): Promise<RequireLoginResponse> {
    try {
      return await request<RequireLoginResponse>("/api/settings/require-login")
    } catch {
      return { requireLogin: false }
    }
  },

  async getAuthStatus(): Promise<AuthStatusResponse> {
    try {
      return await request<AuthStatusResponse>("/api/auth/status")
    } catch {
      return { authenticated: false }
    }
  },

  async login(password: string): Promise<boolean> {
    const res = await request<{ success: boolean }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ password })
    })
    return res.success
  },

  async logout(): Promise<boolean> {
    const res = await request<{ success: boolean }>("/api/auth/logout", {
      method: "POST"
    })
    return res.success
  },

  async getConnections(): Promise<Connection[]> {
    try {
      const res = await request<{ connections: Connection[] }>("/api/connections")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
    } catch (e: any) {
      if (e?.status === 401) {
        throw e
      }
    }

    try {
      const res = await request<{ connections: Connection[] }>("/api/providers")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
    } catch (e: any) {
      if (e?.status === 401) {
        throw e
      }
    }

    try {
      const res = await request<{ connections: Connection[] }>("/api/providers/client?provider=all&accountStatus=all&pageSize=500")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
    } catch (e: any) {
      if (e?.status === 401) {
        throw e
      }
    }

    return []
  },

  async updateConnectionStatus(id: string, isActive: boolean): Promise<Connection | null> {
    try {
      const res = await request<{ connection: Connection }>(`/api/connections/${id}`, {
        method: "PUT",
        body: JSON.stringify({ isActive })
      })
      return res.connection
    } catch {
      const res = await request<{ connection: Connection }>(`/api/providers/${id}`, {
        method: "PUT",
        body: JSON.stringify({ isActive })
      })
      return res.connection
    }
  },

  async getModels(): Promise<ModelItem[]> {
    try {
      const res = await request<{ data: ModelItem[] }>("/v1/models")
      return res.data || []
    } catch {
      try {
        const res = await request<{ data: ModelItem[] }>("/api/models")
        return res.data || []
      } catch {
        return []
      }
    }
  },

  async getCustomModels(): Promise<CustomModel[]> {
    try {
      const res = await request<{ models: CustomModel[] }>("/api/models/custom")
      return res.models || []
    } catch {
      return []
    }
  },

  async getCombos(): Promise<Combo[]> {
    try {
      const res = await request<{ combos: Combo[] }>("/api/combos")
      return res.combos || []
    } catch {
      return []
    }
  },

  async getApiKeys(): Promise<ApiKey[]> {
    try {
      const res = await request<{ keys: ApiKey[] }>("/api/keys")
      return res.keys || []
    } catch {
      return []
    }
  },

  async getUsageStats(): Promise<UsageStats | null> {
    try {
      return await request<UsageStats>("/api/usage/stats")
    } catch {
      return null
    }
  },

  async chatCompletion(params: {
    model: string
    prompt: string
    apiKey?: string
    temperature?: number
    maxTokens?: number
  }): Promise<{ content: string; latencyMs: number }> {
    const startTime = performance.now()
    const headers: Record<string, string> = {
      "Content-Type": "application/json"
    }
    if (params.apiKey) {
      headers["Authorization"] = `Bearer ${params.apiKey}`
    }

    const payload = {
      model: params.model,
      messages: [
        { role: "user", content: params.prompt }
      ],
      temperature: params.temperature ?? 0.7,
      max_tokens: params.maxTokens ?? 300
    }

    let endpoint = "/v1/chat/completions"
    if (!params.apiKey) {
      endpoint = "/api/dashboard/chat/completions"
    }

    const res = await fetch(endpoint, {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
      credentials: "same-origin"
    })

    const latencyMs = Math.round(performance.now() - startTime)

    if (!res.ok) {
      let errText = `HTTP ${res.status}`
      try {
        const errJson = await res.json()
        errText = errJson.error?.message || errJson.error || errText
      } catch {
        errText = await res.text()
      }
      throw new Error(errText || `Failed with HTTP ${res.status}`)
    }

    const data = await res.json()
    const content = data.choices?.[0]?.message?.content || "(empty response)"
    return { content, latencyMs }
  }
}

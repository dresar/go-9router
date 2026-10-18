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

const DEFAULT_API_KEY = "sk-7016dd129191d903-u8emvi-c0b721b9"

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    Accept: "application/json",
    Authorization: `Bearer ${DEFAULT_API_KEY}`,
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(options.headers as Record<string, string>)
  }

  const res = await fetch(path, {
    credentials: "same-origin",
    headers,
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
      const res = await request<any>("/api/connections")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
      if (Array.isArray(res)) {
        return res
      }
    } catch (e: any) {
      if (e?.status === 401) throw e
    }

    try {
      const res = await request<any>("/api/providers")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
      if (Array.isArray(res)) {
        return res
      }
    } catch (e: any) {
      if (e?.status === 401) throw e
    }

    try {
      const res = await request<any>("/api/providers/client?provider=all&accountStatus=all&pageSize=500")
      if (Array.isArray(res?.connections)) {
        return res.connections
      }
      if (Array.isArray(res)) {
        return res
      }
    } catch (e: any) {
      if (e?.status === 401) throw e
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
      try {
        const res = await request<{ connection: Connection }>(`/api/providers/${id}`, {
          method: "PUT",
          body: JSON.stringify({ isActive })
        })
        return res.connection
      } catch {
        const res = await request<{ connection: Connection }>(`/api/connections/${id}`, {
          method: "PATCH",
          body: JSON.stringify({ isActive })
        })
        return res.connection
      }
    }
  },

  async updateProviderBatchStatus(connectionIds: string[], isActive: boolean): Promise<boolean> {
    const promises = connectionIds.map((id) => this.updateConnectionStatus(id, isActive))
    await Promise.allSettled(promises)
    return true
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

  async chatCompletionStream(params: {
    model: string
    prompt: string
    apiKey?: string
    temperature?: number
    maxTokens?: number
    onChunk: (text: string) => void
    signal?: AbortSignal
  }): Promise<{ content: string; latencyMs: number }> {
    const startTime = performance.now()
    const effectiveKey = params.apiKey || DEFAULT_API_KEY
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${effectiveKey}`
    }

    const payload = {
      model: params.model,
      messages: [{ role: "user", content: params.prompt }],
      temperature: params.temperature ?? 0.7,
      max_tokens: params.maxTokens ?? 300,
      stream: true
    }

    const res = await fetch("/v1/chat/completions", {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
      credentials: "same-origin",
      signal: params.signal
    })

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

    if (!res.body) {
      const data = await res.json()
      const content = data.choices?.[0]?.message?.content || ""
      params.onChunk(content)
      const latencyMs = Math.round(performance.now() - startTime)
      return { content, latencyMs }
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let fullContent = ""
    let buffer = ""

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split("\n")
      buffer = lines.pop() || ""

      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed || trimmed.startsWith(":") || trimmed === "data: [DONE]") continue

        if (trimmed.startsWith("data: ")) {
          try {
            const parsed = JSON.parse(trimmed.slice(6))
            const delta = parsed.choices?.[0]?.delta?.content || ""
            if (delta) {
              fullContent += delta
              params.onChunk(fullContent)
            }
          } catch {}
        }
      }
    }

    if (buffer.trim().startsWith("data: ") && buffer.trim() !== "data: [DONE]") {
      try {
        const parsed = JSON.parse(buffer.trim().slice(6))
        const delta = parsed.choices?.[0]?.delta?.content || ""
        if (delta) {
          fullContent += delta
          params.onChunk(fullContent)
        }
      } catch {}
    }

    const latencyMs = Math.round(performance.now() - startTime)
    return { content: fullContent, latencyMs }
  }
}

"use client";

import { useState, useEffect, useRef, useMemo } from "react";
import {
  Card,
  Button,
  SegmentedControl,
  Badge,
  Toggle,
  ProviderIcon,
} from "@/shared/components";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";
import { getModelsByProviderId } from "@/shared/constants/models";
import PlaygroundDropdown from "./PlaygroundDropdown";

// Standard Fallback Provider Names
const PROVIDER_NAMES = {
  openai: "OpenAI",
  gemini: "Google Gemini",
  "gemini-cli": "Gemini CLI",
  claude: "Anthropic Claude",
  deepseek: "DeepSeek",
  groq: "Groq",
  openrouter: "OpenRouter",
  kimi: "Moonshot Kimi",
  kiro: "Kiro AI",
  cline: "Cline",
  clinepass: "Cline Pass",
  nvidia: "NVIDIA NIM",
  xai: "xAI Grok",
  chutes: "Chutes AI",
  "vercel-ai-gateway": "Vercel AI",
  bynara: "Bynara",
  cerebras: "Cerebras AI",
  sambanova: "SambaNova",
  siliconflow: "SiliconFlow",
  geraikita: "GeraiKita",
};

// Provider Presets for Chat Models (Curated free-tier and popular models)
const DEFAULT_PROVIDER_MODELS = {
  openrouter: [
    { id: "openrouter/qwen/qwen3.8-27b:free", name: "Qwen 3.8 27B (Free)", subtitle: "Alibaba latest free" },
    { id: "openrouter/nvidia/nemotron-3.5-lightning:free", name: "Nemotron 3.5 Lightning (Free)", subtitle: "NVIDIA fast reasoning" },
    { id: "openrouter/inclusionai/ling-3.0-flash-fin:free", name: "Ling 3.0 Flash (Free)", subtitle: "Financial & reasoning" },
    { id: "openrouter/deepseek/deepseek-r1:free", name: "DeepSeek R1 (Free)", subtitle: "Advanced reasoning" },
    { id: "openrouter/deepseek/deepseek-chat:free", name: "DeepSeek V3 (Free)", subtitle: "Fast dynamic chat" },
    { id: "openrouter/meta-llama/llama-3.3-70b-instruct:free", name: "Llama 3.3 70B (Free)", subtitle: "Meta instruction tuned" },
    { id: "openrouter/google/gemma-4-26b-a4b-it:free", name: "Gemma 4 26B (Free)", subtitle: "Google open model" },
  ],
  gemini: [
    { id: "gemini/gemini-2.5-flash", name: "Gemini 2.5 Flash", subtitle: "Fast multimodal" },
    { id: "gemini/gemini-2.5-pro", name: "Gemini 2.5 Pro", subtitle: "Complex reasoning" },
    { id: "gemini/gemini-1.5-flash", name: "Gemini 1.5 Flash", subtitle: "High efficiency" },
    { id: "gemini/gemini-1.5-pro", name: "Gemini 1.5 Pro", subtitle: "Extended window" },
  ],
  groq: [
    { id: "groq/openai/gpt-oss-120b", name: "GPT OSS 120B", subtitle: "Ultra-fast flagship reasoning" },
    { id: "groq/qwen/qwen3.8-27b", name: "Qwen 3.8 27B", subtitle: "Alibaba next-gen on LPU" },
    { id: "groq/openai/gpt-oss-20b", name: "GPT OSS 20B", subtitle: "Lightweight fast" },
    { id: "groq/allam-2-7b", name: "Allam 2 7B", subtitle: "Multilingual on LPU" },
  ],
  cerebras: [
    { id: "cerebras/gpt-oss-120b", name: "GPT OSS 120B (Cerebras)", subtitle: "2000+ tokens/sec on WSE" },
    { id: "cerebras/qwen-3.8-27b", name: "Qwen 3.8 27B (Cerebras)", subtitle: "Wafer-scale engine" },
  ],
  sambanova: [
    { id: "sambanova/DeepSeek-V3.2", name: "DeepSeek V3.2", subtitle: "SambaNova SN40L accelerated" },
    { id: "sambanova/Meta-Llama-3.3-70B-Instruct", name: "Llama 3.3 70B", subtitle: "Full context reasoning" },
    { id: "sambanova/MiniMax-M3", name: "MiniMax M3", subtitle: "High-speed MoE" },
    { id: "sambanova/gpt-oss-120b", name: "GPT OSS 120B", subtitle: "Reconfigurable dataflow" },
  ],
  bynara: [
    { id: "bynara/deepseek-v4-flash-alibaba", name: "DeepSeek V4 Flash", subtitle: "Bynara router accelerated" },
    { id: "bynara/claude-sonnet-5", name: "Claude Sonnet 5", subtitle: "Flagship intelligence" },
    { id: "bynara/agnes-2.5-flash", name: "Agnes 2.5 Flash", subtitle: "Multimodal via Bynara" },
  ],
  geraikita: [
    { id: "geraikita/claude-haiku-4.5", name: "Claude Haiku 4.5", subtitle: "1M context (Recommended)" },
    { id: "geraikita/gpt-5.6-sol", name: "GPT 5.6 Sol", subtitle: "1M context (Recommended)" },
    { id: "geraikita/gpt-5.6-terra", name: "GPT 5.6 Terra", subtitle: "1M context (Smooth)" },
    { id: "geraikita/gpt-5.6-luna", name: "GPT 5.6 Luna", subtitle: "1M context (Smooth)" },
    { id: "geraikita/glm-4.7-flash", name: "GLM 4.7 Flash", subtitle: "Fast & lightweight" },
    { id: "geraikita/glm-5", name: "GLM 5", subtitle: "Frontier model" },
    { id: "geraikita/glm-4.7", name: "GLM 4.7", subtitle: "Versatile model" },
    { id: "geraikita/kimi-k2.5", name: "Kimi K2.5", subtitle: "Agentic model" },
    { id: "geraikita/minimax-m2.5", name: "MiniMax M2.5", subtitle: "Reasoning model" },
    { id: "geraikita/minimax-m2.1", name: "MiniMax M2.1", subtitle: "Reasoning model" },
    { id: "geraikita/deepseek-v3.2", name: "DeepSeek V3.2", subtitle: "Versatile model" },
  ],
  kiro: [
    { id: "kiro/claude-3-7-sonnet", name: "Claude 3.7 Sonnet (Kiro)", subtitle: "Via Kiro proxy" },
    { id: "kiro/claude-3-5-sonnet", name: "Claude 3.5 Sonnet (Kiro)", subtitle: "Via Kiro proxy" },
    { id: "kiro/claude-3-5-haiku", name: "Claude 3.5 Haiku (Kiro)", subtitle: "Via Kiro proxy" },
  ],
  cline: [
    { id: "cline/claude-3-7-sonnet", name: "Claude 3.7 Sonnet (Cline)", subtitle: "Via Cline auth" },
    { id: "cline/claude-3-5-sonnet", name: "Claude 3.5 Sonnet (Cline)", subtitle: "Via Cline auth" },
    { id: "cline/claude-3-5-haiku", name: "Claude 3.5 Haiku (Cline)", subtitle: "Via Cline auth" },
  ],
  chutes: [
    { id: "chutes/deepseek-ai/DeepSeek-V3", name: "DeepSeek V3 (Chutes)", subtitle: "Fast open weights" },
    { id: "chutes/deepseek-ai/DeepSeek-R1", name: "DeepSeek R1 (Chutes)", subtitle: "Reasoning engine" },
  ],
};

// Presets for Image Generation (Free Tier only)
const IMAGE_PROVIDERS = [
  { id: "openrouter", name: "OpenRouter", providerId: "openrouter", subtitle: "FLUX & SD (Free Tier)" },
];

const IMAGE_MODELS_BY_PROVIDER = {
  openrouter: [
    { id: "openrouter/black-forest-labs/flux-1-schnell", name: "FLUX 1 Schnell", subtitle: "Free tier fast 4-step" },
    { id: "openrouter/black-forest-labs/flux-1-dev", name: "FLUX 1 Dev", subtitle: "Studio quality" },
  ],
};

const VISION_SAMPLE_PROMPTS = [
  "Deskripsikan objek utama dalam gambar ini.",
  "Ekstrak seluruh teks dalam gambar ini (OCR).",
  "Analisis komposisi warna dan suasana gambar ini.",
];

const IMAGE_SAMPLE_PROMPTS = [
  "Cyberpunk cat wearing neon sunglasses in Tokyo rain, vibrant 8k render",
  "Minimalist 3D isometric glowing AI router floating in dark space",
  "Warm cozy coffee shop corner on a rainy morning, watercolor illustration",
];

export default function PlaygroundClient() {
  const [activeTab, setActiveTab] = useState("chat"); // "chat" or "image"

  // Credentials & Keys
  const [keys, setKeys] = useState([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [showKeySecret, setShowKeySecret] = useState(false);
  const { copied, copy } = useCopyToClipboard(2000);

  // Raw Database Data from Gateway
  const [dbProviders, setDbProviders] = useState([]);
  const [dbCombos, setDbCombos] = useState([]);
  const [dbModels, setDbModels] = useState([]);
  const [dbCustomModels, setDbCustomModels] = useState([]);

  // Chat State
  const [chatProvider, setChatProvider] = useState("combo");
  const [chatModel, setChatModel] = useState("");
  const [chatCustomModel, setChatCustomModel] = useState("");
  const [chatPrompt, setChatPrompt] = useState("");
  const [systemPrompt, setSystemPrompt] = useState("You are an intelligent and helpful AI assistant.");
  const [chatTemperature, setChatTemperature] = useState(0.7);
  const [chatStream, setChatStream] = useState(true);
  const [chatImageDataUrl, setChatImageDataUrl] = useState("");
  const [chatSending, setChatSending] = useState(false);
  const [chatResponse, setChatResponse] = useState("");
  const [chatTelemetry, setChatTelemetry] = useState(null);
  const [chatRawInspector, setChatRawInspector] = useState(false);
  const [chatLastPayload, setChatLastPayload] = useState(null);
  const [chatError, setChatError] = useState("");

  // Image State (Free Tier default)
  const [imageProvider, setImageProvider] = useState("openrouter");
  const [imageModel, setImageModel] = useState("openrouter/black-forest-labs/flux-1-schnell");
  const [imageCustomModel, setImageCustomModel] = useState("");
  const [imagePrompt, setImagePrompt] = useState("");
  const [imageSize, setImageSize] = useState("1024x1024");
  const [imageQuality, setImageQuality] = useState("standard");
  const [imageStyle, setImageStyle] = useState("vivid");
  const [imageGenerating, setImageGenerating] = useState(false);
  const [imageTimer, setImageTimer] = useState(0);
  const [imageCurrentResult, setImageCurrentResult] = useState(null);
  const [imageGallery, setImageGallery] = useState([]);
  const [imageError, setImageError] = useState("");
  const [imageRawInspector, setImageRawInspector] = useState(false);
  const [imageLastPayload, setImageLastPayload] = useState(null);
  const [imageLastRawResponse, setImageLastRawResponse] = useState(null);
  const [lightboxImage, setLightboxImage] = useState(null);

  const fileInputRef = useRef(null);
  const timerRef = useRef(null);

  // 1. Initial Data Fetch: Keys, Providers, Combos, Models, Custom Models
  useEffect(() => {
    async function loadInitialData() {
      try {
        const [keysRes, providersRes, combosRes, modelsRes, customModelsRes] = await Promise.all([
          fetch("/api/keys").catch(() => null),
          fetch("/api/providers").catch(() => null),
          fetch("/api/combos").catch(() => null),
          fetch("/api/models").catch(() => null),
          fetch("/api/models/custom").catch(() => null),
        ]);

        if (keysRes && keysRes.ok) {
          const keysData = await keysRes.json();
          const activeKeys = keysData.keys || [];
          setKeys(activeKeys);
          const activeKey = activeKeys.find((k) => k.isActive) || activeKeys[0];
          if (activeKey) setSelectedKey(activeKey.key);
        }

        let loadedCombos = [];
        if (combosRes && combosRes.ok) {
          const combosData = await combosRes.json();
          loadedCombos = (combosData.combos || []).filter(
            (c) => !c.kind || c.kind === "llm"
          );
          setDbCombos(loadedCombos);
        }

        let activeProviderList = [];
        if (providersRes && providersRes.ok) {
          const providersData = await providersRes.json();
          const conns = providersData.connections || [];
          // Extract unique active providers
          const seen = new Set();
          for (const conn of conns) {
            if (conn.isActive && conn.provider && !seen.has(conn.provider)) {
              seen.add(conn.provider);
              activeProviderList.push({
                id: conn.provider,
                name: PROVIDER_NAMES[conn.provider] || conn.provider,
                providerId: conn.provider,
              });
            }
          }
          setDbProviders(activeProviderList);
        }

        if (modelsRes && modelsRes.ok) {
          const modelsData = await modelsRes.json();
          setDbModels(modelsData.data || []);
        }

        if (customModelsRes && customModelsRes.ok) {
          const customData = await customModelsRes.json();
          setDbCustomModels(customData.models || []);
        }

        // Set Default Provider & Model
        if (loadedCombos.length > 0) {
          setChatProvider("combo");
          setChatModel(loadedCombos[0].name);
        } else if (activeProviderList.length > 0) {
          const firstProv = activeProviderList[0].id;
          setChatProvider(firstProv);
        } else {
          setChatProvider("openrouter");
          setChatModel("openrouter/cohere/north-mini-code:free");
        }
      } catch (err) {
        console.error("Gagal memuat data playground:", err);
      }
    }

    loadInitialData();
  }, []);

  // Timer for Image Generation
  useEffect(() => {
    if (imageGenerating) {
      setImageTimer(0);
      timerRef.current = setInterval(() => {
        setImageTimer((t) => +(t + 0.1).toFixed(1));
      }, 100);
    } else {
      if (timerRef.current) clearInterval(timerRef.current);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [imageGenerating]);

  // Handle Vision Image Select
  const handleImageUpload = (file) => {
    if (!file || !file.type.startsWith("image/")) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      setChatImageDataUrl(e.target?.result || "");
    };
    reader.readAsDataURL(file);
  };

  const clearChatImage = () => {
    setChatImageDataUrl("");
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  // -------------------------------------------------------------
  // Dynamic Dropdown Options for Chat
  // -------------------------------------------------------------
  const chatProviderOptions = useMemo(() => {
    const list = [];

    // 1. Combo Router Option (if combos exist)
    if (dbCombos.length > 0) {
      list.push({
        id: "combo",
        name: "Combo Router",
        isCombo: true,
        badge: `${dbCombos.length} Combo`,
        subtitle: "Fallback multi-model cerdas",
      });
    }

    // 2. Active Providers configured in DB
    for (const p of dbProviders) {
      list.push({
        id: p.id,
        name: p.name,
        providerId: p.id,
        badge: "Aktif",
        subtitle: `Koneksi aktif (${p.id})`,
      });
    }

    // 3. Only if no active providers configured, fallback to verified free-tier providers
    if (list.length === 0) {
      const freeTierFallbacks = ["openrouter", "groq", "gemini"];
      for (const fb of freeTierFallbacks) {
        list.push({
          id: fb,
          name: PROVIDER_NAMES[fb] || fb,
          providerId: fb,
          subtitle: "Free Tier",
        });
      }
    }

    return list;
  }, [dbCombos, dbProviders]);

  const chatModelOptions = useMemo(() => {
    // A. Combos
    if (chatProvider === "combo") {
      return dbCombos.map((c) => ({
        id: c.name,
        name: c.name,
        isCombo: true,
        badge: "Combo",
        subtitle: `${c.models?.length || 0} model: ${(c.models || []).slice(0, 2).join(" → ")}${(c.models || []).length > 2 ? "..." : ""}`,
      }));
    }

    // B. Regular Provider Models
    const models = [];
    const seen = new Set();

    // 1. Curated Presets for this provider (e.g. Free Tier models & favorites)
    const presets = DEFAULT_PROVIDER_MODELS[chatProvider] || [];
    for (const p of presets) {
      if (!seen.has(p.id)) {
        seen.add(p.id);
        models.push({
          id: p.id,
          name: p.name,
          providerId: chatProvider,
          subtitle: p.subtitle,
        });
      }
    }

    // 2. Official models from registry (e.g. Qoder all 16 models, Groq, Gemini)
    const regModels = getModelsByProviderId(chatProvider) || [];
    for (const rm of regModels) {
      const fullId = rm.id.startsWith(`${chatProvider}/`) ? rm.id : `${chatProvider}/${rm.id}`;
      if (!seen.has(fullId)) {
        seen.add(fullId);
        models.push({
          id: fullId,
          name: rm.name || rm.id,
          providerId: chatProvider,
          subtitle: rm.name !== rm.id ? rm.id : "Official model",
        });
      }
    }

    // 3. Custom Models from DB matching this provider
    for (const cm of dbCustomModels) {
      if (cm.providerAlias === chatProvider || cm.provider === chatProvider) {
        const fullId = cm.id.startsWith(`${chatProvider}/`) ? cm.id : `${chatProvider}/${cm.id}`;
        if (!seen.has(fullId)) {
          seen.add(fullId);
          models.push({
            id: fullId,
            name: cm.name || cm.id,
            providerId: chatProvider,
            subtitle: "Custom model",
          });
        }
      }
    }

    // 4. Models from DB `/api/models` matching provider
    for (const m of dbModels) {
      const isOwned = m.owned_by === chatProvider;
      const isPrefixed = m.id.startsWith(`${chatProvider}/`);
      if (isOwned || isPrefixed) {
        if (!seen.has(m.id)) {
          seen.add(m.id);
          models.push({
            id: m.id,
            name: m.id,
            providerId: chatProvider,
            subtitle: "Gateway model",
          });
        }
      }
    }

    // 5. If empty, provide a sensible default
    if (models.length === 0) {
      models.push({
        id: `${chatProvider}/default`,
        name: `${PROVIDER_NAMES[chatProvider] || chatProvider} Default`,
        providerId: chatProvider,
        subtitle: "Rute default",
      });
    }

    // 6. Custom Model option
    models.push({
      id: "custom",
      name: "+ Model Kustom...",
      icon: "edit",
      subtitle: "Ketik nama model manual",
    });

    return models;
  }, [chatProvider, dbCombos, dbModels, dbCustomModels]);

  // Synchronize chatModel when chatProvider changes
  const handleChatProviderChange = (newProviderId) => {
    setChatProvider(newProviderId);
    setChatCustomModel("");

    if (newProviderId === "combo") {
      if (dbCombos.length > 0) {
        setChatModel(dbCombos[0].name);
      }
    } else {
      const presets = DEFAULT_PROVIDER_MODELS[newProviderId] || [];
      const regModels = getModelsByProviderId(newProviderId) || [];
      const customMatch = dbCustomModels.find(
        (m) => m.providerAlias === newProviderId || m.provider === newProviderId
      );
      const dbMatch = dbModels.find(
        (m) => m.owned_by === newProviderId || m.id.startsWith(`${newProviderId}/`)
      );

      if (presets.length > 0) {
        setChatModel(presets[0].id);
      } else if (regModels.length > 0) {
        const rm = regModels[0];
        setChatModel(rm.id.startsWith(`${newProviderId}/`) ? rm.id : `${newProviderId}/${rm.id}`);
      } else if (customMatch) {
        const cid = customMatch.id;
        setChatModel(cid.startsWith(`${newProviderId}/`) ? cid : `${newProviderId}/${cid}`);
      } else if (dbMatch) {
        setChatModel(dbMatch.id);
      } else {
        setChatModel(`${newProviderId}/default`);
      }
    }
  };

  // -------------------------------------------------------------
  // Dynamic Dropdown Options for Image Generation
  // -------------------------------------------------------------
  const imageModelOptions = useMemo(() => {
    const models = (IMAGE_MODELS_BY_PROVIDER[imageProvider] || []).map((m) => ({
      ...m,
      providerId: imageProvider,
    }));
    models.push({
      id: "custom",
      name: "+ Model Kustom...",
      icon: "edit",
      subtitle: "Ketik manual",
    });
    return models;
  }, [imageProvider]);

  const handleImageProviderChange = (newProviderId) => {
    setImageProvider(newProviderId);
    setImageCustomModel("");
    const presets = IMAGE_MODELS_BY_PROVIDER[newProviderId] || [];
    if (presets.length > 0) {
      setImageModel(presets[0].id);
    } else {
      setImageModel("custom");
    }
  };

  // -------------------------------------------------------------
  // ACTION: Execute Chat / Vision via Browser fetch()
  // -------------------------------------------------------------
  const handleSendChat = async () => {
    const promptText = chatPrompt.trim();
    if (!promptText && !chatImageDataUrl) return;

    const targetModel = chatModel === "custom" ? chatCustomModel.trim() : chatModel;
    if (!targetModel) {
      setChatError("Silakan tentukan model terlebih dahulu.");
      return;
    }

    setChatSending(true);
    setChatResponse("");
    setChatError("");
    setChatTelemetry(null);

    const startTime = performance.now();
    let chunkCount = 0;

    const userContent = [];
    if (promptText) {
      userContent.push({ type: "text", text: promptText });
    }
    if (chatImageDataUrl) {
      userContent.push({
        type: "image_url",
        image_url: { url: chatImageDataUrl },
      });
    }

    const messages = [];
    if (systemPrompt.trim()) {
      messages.push({ role: "system", content: systemPrompt.trim() });
    }
    messages.push({
      role: "user",
      content: userContent.length === 1 && userContent[0].type === "text" ? promptText : userContent,
    });

    const payload = {
      model: targetModel,
      messages,
      temperature: chatTemperature,
      stream: chatStream,
    };
    setChatLastPayload(payload);

    const headers = {
      "Content-Type": "application/json",
      "x-playground": "true",
      "x-9router-token-saver": "off",
    };
    if (selectedKey) {
      headers["Authorization"] = `Bearer ${selectedKey}`;
    }

    try {
      const response = await fetch("/v1/chat/completions", {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
      });

      const elapsed = Math.round(performance.now() - startTime);

      if (!response.ok) {
        const errJson = await response.json().catch(() => ({}));
        const msg = errJson.error?.message || errJson.error || errJson.message || `HTTP ${response.status}`;
        setChatError(typeof msg === "string" ? msg : JSON.stringify(msg));
        setChatTelemetry({ status: response.status, latencyMs: elapsed });
        return;
      }

      if (chatStream && response.body) {
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let fullText = "";
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split(/\r?\n/);
          buffer = lines.pop() || "";

          for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed.startsWith("data:")) continue;
            const dataStr = trimmed.slice(5).trim();
            if (dataStr === "[DONE]") continue;

            try {
              const parsed = JSON.parse(dataStr);
              chunkCount++;
              const delta = parsed.choices?.[0]?.delta?.content || "";
              if (delta) {
                fullText += delta;
                setChatResponse(fullText);
              }
            } catch {
              // Ignore partial chunks
            }
          }
        }

        if (!fullText && buffer.trim()) {
          try {
            const errObj = JSON.parse(buffer.trim());
            const msg = errObj.error?.message || errObj.message || buffer.trim();
            setChatError(typeof msg === "string" ? msg : JSON.stringify(msg));
          } catch {
            setChatResponse(buffer.trim());
          }
        }

        const totalElapsed = Math.round(performance.now() - startTime);
        setChatTelemetry({
          status: response.status,
          latencyMs: totalElapsed,
          chunks: chunkCount,
        });
      } else {
        const data = await response.json();
        const reply = data.choices?.[0]?.message?.content || JSON.stringify(data, null, 2);
        setChatResponse(reply);
        setChatTelemetry({
          status: response.status,
          latencyMs: elapsed,
          tokens: data.usage?.total_tokens,
        });
      }
    } catch (err) {
      setChatError(`Error: ${err.message}`);
    } finally {
      setChatSending(false);
    }
  };

  // -------------------------------------------------------------
  // ACTION: Execute Image Generation via Browser fetch()
  // -------------------------------------------------------------
  const handleGenerateImage = async () => {
    const promptText = imagePrompt.trim();
    if (!promptText) return;

    const targetModel = imageModel === "custom" ? imageCustomModel.trim() : imageModel;
    if (!targetModel) {
      setImageError("Silakan tentukan model gambar.");
      return;
    }

    setImageGenerating(true);
    setImageError("");
    setImageCurrentResult(null);

    const payload = {
      model: targetModel,
      prompt: promptText,
      n: 1,
      size: imageSize,
      quality: imageQuality,
      style: imageStyle,
      response_format: "url",
    };
    setImageLastPayload(payload);

    const headers = {
      "Content-Type": "application/json",
    };
    if (selectedKey) {
      headers["Authorization"] = `Bearer ${selectedKey}`;
    }

    try {
      const response = await fetch("/v1/images/generations", {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
      });

      const data = await response.json().catch(() => ({}));
      setImageLastRawResponse(data);

      if (!response.ok) {
        const errMsg = data.error?.message || data.error || data.message || `HTTP ${response.status}`;
        setImageError(typeof errMsg === "string" ? errMsg : JSON.stringify(errMsg));
        return;
      }

      const imageUrl = data.data?.[0]?.url || data.data?.[0]?.b64_json;
      const revisedPrompt = data.data?.[0]?.revised_prompt || promptText;

      if (!imageUrl) {
        setImageError("Gambar tidak ditemukan.");
        return;
      }

      const finalUrl = imageUrl.startsWith("data:") || imageUrl.startsWith("http")
        ? imageUrl
        : `data:image/png;base64,${imageUrl}`;

      const resultItem = {
        id: `img_${Date.now()}`,
        url: finalUrl,
        prompt: promptText,
        revisedPrompt,
        model: targetModel,
        size: imageSize,
        createdAt: new Date().toLocaleTimeString(),
      };

      setImageCurrentResult(resultItem);
      setImageGallery((prev) => [resultItem, ...prev]);
    } catch (err) {
      setImageError(`Error: ${err.message}`);
    } finally {
      setImageGenerating(false);
    }
  };

  const maskedKey = selectedKey
    ? `${selectedKey.slice(0, 7)}••••••••••••••••${selectedKey.slice(-6)}`
    : "Key default";

  return (
    <div className="flex min-w-0 flex-col gap-6 pb-12">
      {/* 1. Header Banner & Mode Switcher */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-text-main">
              Playground
            </h1>
            <Badge variant="primary" size="sm">
              Browser
            </Badge>
          </div>
          <p className="text-xs text-text-muted mt-1">
            Test models, fallback combos, and generate images locally.
          </p>
        </div>

        {/* Tab Switcher using 9Router's SegmentedControl */}
        <SegmentedControl
          options={[
            { value: "chat", label: "Chat & Vision", icon: "forum" },
            { value: "image", label: "Image", icon: "palette" },
          ]}
          value={activeTab}
          onChange={setActiveTab}
          size="md"
        />
      </div>

      {/* 2. Auto-Configured API Key Card */}
      <Card padding="xs" className="bg-surface border-border-subtle">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-2 py-1">
          <div className="flex items-center gap-3 min-w-0">
            <div className="p-2 rounded-[10px] bg-bg text-brand-500 shrink-0">
              <span className="material-symbols-outlined text-[18px]">key</span>
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-text-muted">API Key</span>
                <span className="text-[10px] bg-brand-500/10 text-brand-500 border border-brand-500/20 px-1.5 py-0.5 rounded font-mono">
                  Automatic
                </span>
              </div>
              <div className="font-mono text-xs text-text-main mt-0.5 flex items-center gap-2 truncate">
                <span>{showKeySecret ? selectedKey : maskedKey}</span>
                {selectedKey && (
                  <button
                    type="button"
                    onClick={() => setShowKeySecret(!showKeySecret)}
                    className="text-text-muted hover:text-text-main transition-colors cursor-pointer"
                    title={showKeySecret ? "Hide" : "Show"}
                  >
                    <span className="material-symbols-outlined text-[16px] align-middle">
                      {showKeySecret ? "visibility_off" : "visibility"}
                    </span>
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            {keys.length > 1 && (
              <select
                value={selectedKey}
                onChange={(e) => setSelectedKey(e.target.value)}
                className="bg-surface-2 text-xs text-text-main border border-border-subtle rounded-[10px] px-2.5 py-1.5 focus:outline-none focus:border-brand-500/50"
              >
                {keys.map((k) => (
                  <option key={k.id} value={k.key}>
                    {k.name}
                  </option>
                ))}
              </select>
            )}

            {selectedKey && (
              <Button
                variant="secondary"
                size="sm"
                icon={copied ? "check" : "content_copy"}
                onClick={() => copy(selectedKey)}
              >
                {copied ? "Copied!" : "Copy"}
              </Button>
            )}
          </div>
        </div>
      </Card>

      {/* ----------------------------------------------------------------- */}
      {/* TAB 1: CHATGPT & VISION (CHAT, COMBO ROUTING, MULTIMODAL)          */}
      {/* ----------------------------------------------------------------- */}
      {activeTab === "chat" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Controls Column */}
          <div className="lg:col-span-5 space-y-4">
            <Card padding="sm" className="space-y-4">
              <div className="flex items-center justify-between border-b border-border-subtle pb-2.5">
                <div className="flex items-center gap-2">
                  <span className="material-symbols-outlined text-brand-500 text-[20px]">
                    tune
                  </span>
                  <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
                    Parameter
                  </h2>
                </div>
                {chatProvider === "combo" && (
                  <Badge variant="primary" size="sm" icon="bolt">
                    Combo Mode
                  </Badge>
                )}
              </div>

              {/* DUAL DROPDOWNS: 1. PROVIDER & 2. MODEL/COMBO */}
              <div className="space-y-3">
                {/* Dropdown 1: Provider */}
                <PlaygroundDropdown
                  label="Provider"
                  value={chatProvider}
                  onChange={handleChatProviderChange}
                  options={chatProviderOptions}
                  placeholder="Select Provider"
                />

                {/* Dropdown 2: Model (Synchronized) */}
                <PlaygroundDropdown
                  label={chatProvider === "combo" ? "Combo Router" : "Model"}
                  value={chatModel}
                  onChange={(val) => {
                    setChatModel(val);
                    if (val !== "custom") setChatCustomModel("");
                  }}
                  options={chatModelOptions}
                  placeholder={chatProvider === "combo" ? "Select Combo..." : "Select Model..."}
                />

                {/* Custom Model Input when 'custom' selected */}
                {chatModel === "custom" && (
                  <div className="pt-1">
                    <input
                      type="text"
                      placeholder="Model"
                      value={chatCustomModel}
                      onChange={(e) => setChatCustomModel(e.target.value)}
                      className="w-full bg-surface-2 border border-border-subtle rounded-[10px] px-3 py-2 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50"
                    />
                  </div>
                )}
              </div>

              {/* System Instruction */}
              <div>
                <label className="block text-xs font-medium text-text-muted mb-1.5">
                  Instruction
                </label>
                <input
                  type="text"
                  value={systemPrompt}
                  onChange={(e) => setSystemPrompt(e.target.value)}
                  placeholder="Instruction..."
                  className="w-full bg-surface-2 border border-border-subtle rounded-[10px] px-3 py-2 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50"
                />
              </div>

              {/* Temperature & Stream Toggle */}
              <div className="grid grid-cols-2 gap-3 items-center">
                <div>
                  <div className="flex justify-between text-xs text-text-muted mb-1.5">
                    <span>Temperature</span>
                    <span className="font-mono text-brand-500">{chatTemperature}</span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={chatTemperature}
                    onChange={(e) => setChatTemperature(parseFloat(e.target.value))}
                    className="w-full accent-brand-500 cursor-pointer"
                  />
                </div>
                <div className="flex items-center justify-between p-2.5 bg-surface-2 rounded-[10px] border border-border-subtle">
                  <span className="text-xs text-text-main font-medium">Stream</span>
                  <input
                    type="checkbox"
                    checked={chatStream}
                    onChange={(e) => setChatStream(e.target.checked)}
                    className="w-4 h-4 accent-brand-500 rounded cursor-pointer"
                  />
                </div>
              </div>

              {/* Multimodal Vision Upload */}
              <div className="border-t border-border-subtle pt-3">
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-medium text-text-muted flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-brand-500 text-[16px]">
                      visibility
                    </span>
                    Vision
                  </label>
                  {chatImageDataUrl && (
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="text-[11px] text-red-400 hover:text-red-300 transition-colors cursor-pointer"
                    >
                      Remove
                    </button>
                  )}
                </div>

                {!chatImageDataUrl ? (
                  <div
                    onClick={() => fileInputRef.current?.click()}
                    className="border border-dashed border-border-subtle hover:border-brand-500/50 rounded-[10px] p-3 text-center cursor-pointer transition-colors bg-bg/50 group"
                  >
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => handleImageUpload(e.target.files?.[0])}
                    />
                    <span className="material-symbols-outlined text-xl text-text-muted group-hover:text-brand-500 transition-colors">
                      cloud_upload
                    </span>
                    <p className="text-xs text-text-muted mt-0.5">
                      Upload image
                    </p>
                  </div>
                ) : (
                  <div className="relative rounded-[10px] overflow-hidden border border-border-subtle bg-bg flex items-center justify-center max-h-44">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={chatImageDataUrl} alt="Vision" className="max-h-44 object-contain" />
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="absolute top-2 right-2 w-6 h-6 rounded-full bg-red-600/90 text-white flex items-center justify-center shadow-md hover:bg-red-500 cursor-pointer"
                    >
                      <span className="material-symbols-outlined text-xs">close</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Sample Vision Prompts */}
              {chatImageDataUrl && (
                <div>
                  <span className="text-[11px] font-medium text-text-muted block mb-1.5">
                    Sample Prompts:
                  </span>
                  <div className="flex flex-col gap-1">
                    {VISION_SAMPLE_PROMPTS.map((p, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setChatPrompt(p)}
                        className="text-left text-[11px] p-2 rounded-[8px] bg-surface-2 hover:bg-surface-3 text-text-main border border-border-subtle transition-colors line-clamp-1 cursor-pointer"
                      >
                        {p}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Chat Prompt Input */}
              <div>
                <label className="block text-xs font-medium text-text-muted mb-1.5">
                  Message
                </label>
                <textarea
                  rows={3}
                  value={chatPrompt}
                  onChange={(e) => setChatPrompt(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
                      handleSendChat();
                    }
                  }}
                  placeholder="Message..."
                  className="w-full bg-surface-2 border border-border-subtle rounded-[10px] p-2.5 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Send Button */}
              <Button
                variant="primary"
                fullWidth
                icon={chatSending ? "progress_activity" : "send"}
                loading={chatSending}
                disabled={chatSending || (!chatPrompt.trim() && !chatImageDataUrl)}
                onClick={handleSendChat}
              >
                {chatSending ? "Sending..." : "Send"}
              </Button>
            </Card>
          </div>

          {/* Response Column */}
          <div className="lg:col-span-7 space-y-4">
            <Card padding="sm" className="min-h-[480px] flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between border-b border-border-subtle pb-2.5 mb-3">
                  <div className="flex items-center gap-2">
                    <span className="material-symbols-outlined text-brand-500 text-[20px]">
                      smart_toy
                    </span>
                    <h3 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
                      Response
                    </h3>
                  </div>

                  {chatTelemetry && (
                    <div className="flex items-center gap-2 text-[11px] font-mono">
                      <span className="px-2 py-0.5 rounded-full bg-green-500/10 text-green-500 border border-green-500/20">
                        {chatTelemetry.status}
                      </span>
                      <span className="text-text-muted">{chatTelemetry.latencyMs}ms</span>
                      {chatTelemetry.chunks && (
                        <span className="text-text-muted">{chatTelemetry.chunks} chunks</span>
                      )}
                      {chatTelemetry.tokens && (
                        <span className="text-text-muted">{chatTelemetry.tokens} tok</span>
                      )}
                    </div>
                  )}
                </div>

                {/* Error Box */}
                {chatError && (
                  <div className="bg-red-500/10 border border-red-500/20 rounded-[10px] p-3 text-red-400 text-xs flex items-start gap-2 mb-3">
                    <span className="material-symbols-outlined text-red-500 text-base shrink-0 mt-0.5">
                      error
                    </span>
                    <div className="min-w-0">
                      <span className="font-semibold block">Failed</span>
                      <span className="font-mono text-[11px] break-all">{chatError}</span>
                    </div>
                  </div>
                )}

                {/* Response Text Display */}
                {chatResponse ? (
                  <div className="bg-bg p-4 rounded-[12px] border border-border-subtle text-xs text-text-main font-sans leading-relaxed whitespace-pre-wrap max-h-[500px] overflow-y-auto custom-scrollbar select-text">
                    {chatResponse}
                  </div>
                ) : chatSending ? (
                  <div className="flex flex-col items-center justify-center py-24 text-center">
                    <span className="material-symbols-outlined text-3xl text-brand-500 animate-spin mb-3">
                      progress_activity
                    </span>
                    <p className="text-xs text-text-muted">Streaming data...</p>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-24 text-center text-text-muted">
                    <span className="material-symbols-outlined text-3xl text-text-muted mb-2">
                      chat_bubble_outline
                    </span>
                    <h4 className="text-xs font-medium text-text-muted">No response yet</h4>
                  </div>
                )}
              </div>

              {/* Inspector & Copy Tools */}
              <div className="mt-4 pt-3 border-t border-border-subtle flex items-center justify-between text-xs">
                <Button
                  variant="ghost"
                  size="sm"
                  icon="terminal"
                  onClick={() => setChatRawInspector(!chatRawInspector)}
                >
                  {chatRawInspector ? "Close JSON" : "JSON"}
                </Button>

                {chatResponse && (
                  <Button
                    variant="secondary"
                    size="sm"
                    icon={copied ? "check" : "content_copy"}
                    onClick={() => copy(chatResponse)}
                  >
                    {copied ? "Copied!" : "Copy"}
                  </Button>
                )}
              </div>

              {/* Raw JSON Accordion */}
              {chatRawInspector && (
                <div className="mt-3 bg-bg p-3 rounded-[10px] border border-border-subtle overflow-x-auto text-[11px] font-mono">
                  <span className="text-brand-500 font-semibold block mb-1">Payload:</span>
                  <pre className="text-text-muted leading-tight">
                    {JSON.stringify(chatLastPayload, null, 2)}
                  </pre>
                </div>
              )}
            </Card>
          </div>
        </div>
      )}

      {/* ----------------------------------------------------------------- */}
      {/* TAB 2: IMAGE GENERATION (DALL-E & FLUX)                           */}
      {/* ----------------------------------------------------------------- */}
      {activeTab === "image" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Controls Column */}
          <div className="lg:col-span-5 space-y-4">
            <Card padding="sm" className="space-y-4">
              <div className="flex items-center gap-2 border-b border-border-subtle pb-2.5">
                <span className="material-symbols-outlined text-brand-500 text-[20px]">
                  brush
                </span>
                <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
                  Image Parameters
                </h2>
              </div>

              {/* DUAL DROPDOWNS: 1. PROVIDER & 2. MODEL */}
              <div className="space-y-3">
                <PlaygroundDropdown
                  label="Provider"
                  value={imageProvider}
                  onChange={handleImageProviderChange}
                  options={IMAGE_PROVIDERS}
                  placeholder="Select Provider"
                />

                <PlaygroundDropdown
                  label="Image Model"
                  value={imageModel}
                  onChange={(val) => {
                    setImageModel(val);
                    if (val !== "custom") setImageCustomModel("");
                  }}
                  options={imageModelOptions}
                  placeholder="Select Model"
                />

                {imageModel === "custom" && (
                  <div className="pt-1">
                    <input
                      type="text"
                      placeholder="Model"
                      value={imageCustomModel}
                      onChange={(e) => setImageCustomModel(e.target.value)}
                      className="w-full bg-surface-2 border border-border-subtle rounded-[10px] px-3 py-2 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50"
                    />
                  </div>
                )}
              </div>

              {/* Resolution & Quality */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-text-muted mb-1.5">
                    Resolution
                  </label>
                  <select
                    value={imageSize}
                    onChange={(e) => setImageSize(e.target.value)}
                    className="w-full bg-surface-2 border border-border-subtle rounded-[10px] px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-500/50"
                  >
                    <option value="1024x1024">1024x1024 (1:1)</option>
                    <option value="1024x1792">1024x1792 (9:16)</option>
                    <option value="1792x1024">1792x1024 (16:9)</option>
                    <option value="512x512">512x512</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-text-muted mb-1.5">
                    Quality
                  </label>
                  <select
                    value={imageQuality}
                    onChange={(e) => setImageQuality(e.target.value)}
                    className="w-full bg-surface-2 border border-border-subtle rounded-[10px] px-3 py-2 text-xs text-text-main focus:outline-none focus:border-brand-500/50"
                  >
                    <option value="standard">Standard</option>
                    <option value="hd">HD</option>
                  </select>
                </div>
              </div>

              {/* Style Selection */}
              <div>
                <label className="block text-xs font-medium text-text-muted mb-1.5">
                  Style
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setImageStyle("vivid")}
                    className={`px-3 py-2 text-xs font-medium rounded-[10px] border transition-all cursor-pointer ${
                      imageStyle === "vivid"
                        ? "bg-brand-500/10 border-brand-500 text-brand-500 font-semibold"
                        : "bg-surface-2 border-border-subtle text-text-muted hover:text-text-main"
                    }`}
                  >
                    Vivid
                  </button>
                  <button
                    type="button"
                    onClick={() => setImageStyle("natural")}
                    className={`px-3 py-2 text-xs font-medium rounded-[10px] border transition-all cursor-pointer ${
                      imageStyle === "natural"
                        ? "bg-brand-500/10 border-brand-500 text-brand-500 font-semibold"
                        : "bg-surface-2 border-border-subtle text-text-muted hover:text-text-main"
                    }`}
                  >
                    Natural
                  </button>
                </div>
              </div>

              {/* Prompt Textarea */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-medium text-text-muted">
                    Prompt
                  </label>
                  <span className="text-[10px] text-text-muted font-mono">
                    {imagePrompt.length} chars
                  </span>
                </div>
                <textarea
                  rows={4}
                  value={imagePrompt}
                  onChange={(e) => setImagePrompt(e.target.value)}
                  placeholder="Prompt..."
                  className="w-full bg-surface-2 border border-border-subtle rounded-[10px] p-3 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Sample Prompts */}
              <div>
                <span className="text-[11px] font-medium text-text-muted block mb-1.5">
                  Sample Prompts:
                </span>
                <div className="flex flex-col gap-1.5">
                  {IMAGE_SAMPLE_PROMPTS.map((p, idx) => (
                    <button
                      key={idx}
                      type="button"
                      onClick={() => setImagePrompt(p)}
                      className="text-left text-[11px] p-2 rounded-[8px] bg-surface-2 hover:bg-surface-3 text-text-main border border-border-subtle transition-colors line-clamp-1 cursor-pointer"
                    >
                      {p}
                    </button>
                  ))}
                </div>
              </div>

              {/* Generate Button */}
              <Button
                variant="primary"
                fullWidth
                icon={imageGenerating ? "progress_activity" : "palette"}
                loading={imageGenerating}
                disabled={imageGenerating || !imagePrompt.trim()}
                onClick={handleGenerateImage}
              >
                {imageGenerating ? `Processing (${imageTimer}s)...` : "Generate"}
              </Button>
            </Card>
          </div>

          {/* Results Column */}
          <div className="lg:col-span-7 space-y-4">
            <Card padding="sm" className="min-h-[480px] flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between border-b border-border-subtle pb-2.5 mb-3">
                  <div className="flex items-center gap-2">
                    <span className="material-symbols-outlined text-brand-500 text-[20px]">
                      image
                    </span>
                    <h3 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
                      Generated Image
                    </h3>
                  </div>

                  {imageCurrentResult && (
                    <Badge variant="primary" size="sm">
                      {imageCurrentResult.size}
                    </Badge>
                  )}
                </div>

                {/* Error Box */}
                {imageError && (
                  <div className="bg-red-500/10 border border-red-500/20 rounded-[10px] p-3 text-red-400 text-xs flex items-start gap-2 mb-3">
                    <span className="material-symbols-outlined text-red-500 text-base shrink-0 mt-0.5">
                      error
                    </span>
                    <div className="min-w-0">
                      <span className="font-semibold block">Failed</span>
                      <span className="font-mono text-[11px] break-all">{imageError}</span>
                    </div>
                  </div>
                )}

                {/* Generating Loading State */}
                {imageGenerating && (
                  <div className="flex flex-col items-center justify-center py-28 text-center">
                    <span className="material-symbols-outlined text-4xl text-brand-500 animate-spin mb-3">
                      palette
                    </span>
                    <h4 className="text-xs font-semibold text-text-main">
                      Processing...
                    </h4>
                    <p className="text-xs text-text-muted mt-1 font-mono">{imageTimer}s</p>
                  </div>
                )}

                {/* Image Display */}
                {!imageGenerating && imageCurrentResult && (
                  <div className="space-y-3">
                    <div
                      onClick={() => setLightboxImage(imageCurrentResult.url)}
                      className="group relative aspect-square max-h-[420px] mx-auto rounded-[12px] overflow-hidden border border-border-subtle bg-bg flex items-center justify-center cursor-pointer shadow-lg"
                    >
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img
                        src={imageCurrentResult.url}
                        alt={imageCurrentResult.prompt}
                        className="w-full h-full object-contain group-hover:scale-[1.02] transition-transform duration-300"
                      />
                      <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-3">
                        <span className="material-symbols-outlined text-white text-3xl">
                          zoom_in
                        </span>
                      </div>
                    </div>

                    <div className="bg-bg p-3 rounded-[10px] border border-border-subtle text-xs space-y-1">
                      <span className="text-[10px] font-semibold text-text-muted uppercase">
                        Revised Prompt:
                      </span>
                      <p className="text-text-main italic leading-relaxed">
                        {imageCurrentResult.revisedPrompt}
                      </p>
                    </div>
                  </div>
                )}

                {/* Empty State */}
                {!imageGenerating && !imageCurrentResult && !imageError && (
                  <div className="flex flex-col items-center justify-center py-28 text-center text-text-muted">
                    <span className="material-symbols-outlined text-3xl text-text-muted mb-2">
                      add_photo_alternate
                    </span>
                    <h4 className="text-xs font-medium text-text-muted">No results yet</h4>
                  </div>
                )}
              </div>

              {/* Inspector Footer */}
              <div className="mt-4 pt-3 border-t border-border-subtle flex items-center justify-between text-xs">
                <Button
                  variant="ghost"
                  size="sm"
                  icon="code"
                  onClick={() => setImageRawInspector(!imageRawInspector)}
                >
                  {imageRawInspector ? "Close JSON" : "JSON"}
                </Button>
              </div>

              {/* Raw JSON Accordion */}
              {imageRawInspector && (
                <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px] font-mono">
                  <div className="bg-bg p-2.5 rounded-[10px] border border-border-subtle overflow-x-auto">
                    <span className="text-brand-500 font-semibold block mb-1">Payload:</span>
                    <pre className="text-text-muted">
                      {JSON.stringify(imageLastPayload, null, 2)}
                    </pre>
                  </div>
                  <div className="bg-bg p-2.5 rounded-[10px] border border-border-subtle overflow-x-auto">
                    <span className="text-green-500 font-semibold block mb-1">Response:</span>
                    <pre className="text-text-muted">
                      {imageLastRawResponse ? JSON.stringify(imageLastRawResponse, null, 2) : "Empty"}
                    </pre>
                  </div>
                </div>
              )}
            </Card>

            {/* Gallery of Session Images */}
            {imageGallery.length > 1 && (
              <Card padding="sm">
                <h4 className="text-xs font-semibold text-text-muted mb-2.5 flex items-center gap-1.5">
                  <span className="material-symbols-outlined text-sm">photo_library</span>
                  History ({imageGallery.length})
                </h4>
                <div className="grid grid-cols-4 sm:grid-cols-6 gap-2">
                  {imageGallery.map((item) => (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => setImageCurrentResult(item)}
                      className={`relative aspect-square rounded-[8px] overflow-hidden border transition-all cursor-pointer ${
                        imageCurrentResult?.id === item.id
                          ? "border-brand-500 ring-2 ring-brand-500/30"
                          : "border-border-subtle opacity-75 hover:opacity-100"
                      }`}
                    >
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img src={item.url} alt={item.prompt} className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              </Card>
            )}
          </div>
        </div>
      )}

      {/* Lightbox Modal */}
      {lightboxImage && (
        <div
          onClick={() => setLightboxImage(null)}
          className="fixed inset-0 z-50 bg-black/90 backdrop-blur-sm flex items-center justify-center p-4 cursor-pointer"
        >
          <div className="relative max-w-4xl max-h-[90vh]">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={lightboxImage}
              alt="Preview"
              className="max-w-full max-h-[90vh] object-contain rounded-xl shadow-2xl"
            />
            <button
              type="button"
              onClick={() => setLightboxImage(null)}
              className="absolute top-3 right-3 w-8 h-8 rounded-full bg-black/80 text-white flex items-center justify-center shadow-lg hover:bg-zinc-800 cursor-pointer"
            >
              <span className="material-symbols-outlined text-sm">close</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

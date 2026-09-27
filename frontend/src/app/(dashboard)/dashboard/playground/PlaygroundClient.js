"use client";

import { useState, useEffect, useRef } from "react";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";

const CHAT_PRESETS = [
  { id: "openai/gpt-4o", name: "ChatGPT (GPT-4o)", provider: "openai" },
  { id: "openai/chatgpt-4o-latest", name: "ChatGPT 4o Latest", provider: "openai" },
  { id: "openai/gpt-4o-mini", name: "GPT-4o Mini (Fast)", provider: "openai" },
  { id: "claude-3-5-sonnet-20241022", name: "Claude 3.5 Sonnet", provider: "claude" },
  { id: "gemini/default", name: "Google Gemini 2.5 Flash", provider: "gemini" },
  { id: "deepseek/default", name: "DeepSeek Chat", provider: "deepseek" },
  { id: "grok-3", name: "xAI Grok 3", provider: "xai" },
  { id: "openrouter/default", name: "OpenRouter Default", provider: "openrouter" },
];

const IMAGE_PRESETS = [
  { id: "openai/dall-e-3", name: "ChatGPT DALL-E 3 (OpenAI)", provider: "openai" },
  { id: "openai/dall-e-2", name: "DALL-E 2 (OpenAI)", provider: "openai" },
  { id: "openrouter/black-forest-labs/flux-1-schnell", name: "FLUX 1 Schnell (Ultra Fast)", provider: "openrouter" },
  { id: "openrouter/black-forest-labs/flux-1-dev", name: "FLUX 1 Dev (High Quality)", provider: "openrouter" },
  { id: "grok-imagine", name: "xAI Grok Imagine", provider: "xai" },
];

const VISION_SAMPLE_PROMPTS = [
  "Deskripsikan gambar ini secara detail dan sebutkan elemen-elemen utamanya.",
  "Baca dan ekstrak seluruh teks yang tampak pada gambar ini (OCR).",
  "Analisis gaya artistik, warna dominan, dan suasana dari gambar ini.",
];

const IMAGE_SAMPLE_PROMPTS = [
  "A futuristic cyberpunk cat wearing neon sunglasses in Tokyo rain, vibrant cinematic lighting, photorealistic 8k",
  "Minimalist 3D isometric glowing AI router with purple and cyan optical fibers floating in dark space",
  "Warm cozy coffee shop corner on a rainy morning with steaming cup and bookshelf, watercolor illustration",
  "Hyper-detailed portrait of a cybernetic warrior with glowing holographic visor, cinematic depth of field",
];

export default function PlaygroundClient() {
  const [activeTab, setActiveTab] = useState("image"); // "image" or "chat"
  
  // API Key Management
  const [keys, setKeys] = useState([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [showKeySecret, setShowKeySecret] = useState(false);
  const [loadingKeys, setLoadingKeys] = useState(true);
  const { copied, copy } = useCopyToClipboard(2000);

  // Available Models from Gateway
  const [availableModels, setAvailableModels] = useState([]);

  // --- TAB 1: CHAT & VISION STATE ---
  const [chatModel, setChatModel] = useState("openai/gpt-4o");
  const [chatCustomModel, setChatCustomModel] = useState("");
  const [chatPrompt, setChatPrompt] = useState("");
  const [systemPrompt, setSystemPrompt] = useState("You are an intelligent and helpful AI assistant.");
  const [chatTemperature, setChatTemperature] = useState(0.7);
  const [chatStream, setChatStream] = useState(true);
  const [chatImageFile, setChatImageFile] = useState(null);
  const [chatImageDataUrl, setChatImageDataUrl] = useState("");
  const [chatSending, setChatSending] = useState(false);
  const [chatResponse, setChatResponse] = useState("");
  const [chatTelemetry, setChatTelemetry] = useState(null);
  const [chatRawInspector, setChatRawInspector] = useState(false);
  const [chatLastPayload, setChatLastPayload] = useState(null);
  const [chatError, setChatError] = useState("");

  // --- TAB 2: IMAGE GENERATION (DALL-E / FLUX) STATE ---
  const [imageModel, setImageModel] = useState("openai/dall-e-3");
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

  // 1. Auto-Fetch API Keys and Available Models
  useEffect(() => {
    async function initPlayground() {
      try {
        const [keysRes, modelsRes] = await Promise.all([
          fetch("/api/keys").catch(() => null),
          fetch("/api/models").catch(() => null),
        ]);

        if (keysRes && keysRes.ok) {
          const keysData = await keysRes.json();
          const activeKeys = keysData.keys || [];
          setKeys(activeKeys);
          const activeKey = activeKeys.find((k) => k.isActive) || activeKeys[0];
          if (activeKey) {
            setSelectedKey(activeKey.key);
          }
        }

        if (modelsRes && modelsRes.ok) {
          const modelsData = await modelsRes.json();
          const list = modelsData.data || [];
          setAvailableModels(list);
        }
      } catch (err) {
        console.error("Failed to load playground initial data:", err);
      } finally {
        setLoadingKeys(false);
      }
    }

    initPlayground();
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
    setChatImageFile(file);
    const reader = new FileReader();
    reader.onload = (e) => {
      setChatImageDataUrl(e.target?.result || "");
    };
    reader.readAsDataURL(file);
  };

  const clearChatImage = () => {
    setChatImageFile(null);
    setChatImageDataUrl("");
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  // -------------------------------------------------------------
  // ACTION: Execute Chat / Vision via Browser fetch()
  // -------------------------------------------------------------
  const handleSendChat = async () => {
    const promptText = chatPrompt.trim();
    if (!promptText && !chatImageDataUrl) return;

    const targetModel = chatCustomModel.trim() || chatModel;
    setChatSending(true);
    setChatResponse("");
    setChatError("");
    setChatTelemetry(null);

    const startTime = performance.now();
    let chunkCount = 0;

    // Build OpenAI-compliant Vision or Text message
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
        const msg = errJson.error?.message || errJson.error || errJson.message || `HTTP ${response.status} ${response.statusText}`;
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
              // Ignore partial chunk parsing
            }
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
      setChatError(`Network/Client Error: ${err.message}`);
    } finally {
      setChatSending(false);
    }
  };

  // -------------------------------------------------------------
  // ACTION: Execute Image Generation (DALL-E 3 / Flux) via Browser fetch()
  // -------------------------------------------------------------
  const handleGenerateImage = async () => {
    const promptText = imagePrompt.trim();
    if (!promptText) return;

    const targetModel = imageCustomModel.trim() || imageModel;
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
        const errMsg = data.error?.message || data.error || data.message || `HTTP ${response.status}: Failed to generate image`;
        setImageError(typeof errMsg === "string" ? errMsg : JSON.stringify(errMsg));
        return;
      }

      const imageUrl = data.data?.[0]?.url || data.data?.[0]?.b64_json;
      const revisedPrompt = data.data?.[0]?.revised_prompt || promptText;

      if (!imageUrl) {
        setImageError("No image URL or b64 data returned from provider.");
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
      setImageError(`Network/Client Error: ${err.message}`);
    } finally {
      setImageGenerating(false);
    }
  };

  const maskedKey = selectedKey
    ? `${selectedKey.slice(0, 7)}••••••••••••••••${selectedKey.slice(-6)}`
    : "No key configured (using gateway default)";

  return (
    <div className="w-full space-y-6 pb-12">
      {/* 1. Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-5">
        <div>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg shadow-indigo-500/20 text-white">
              <span className="material-symbols-outlined text-2xl">science</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                API Playground
                <span className="text-xs px-2.5 py-0.5 rounded-full font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  Client-Side • Direct Browser Fetch
                </span>
              </h1>
              <p className="text-xs text-zinc-400 mt-0.5">
                Uji coba langsung ChatGPT, Vision, dan Image Generation (DALL-E 3 / Flux) langsung dari browser Anda tanpa skrip backend.
              </p>
            </div>
          </div>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center bg-zinc-900 border border-zinc-800 p-1 rounded-xl shadow-inner">
          <button
            type="button"
            onClick={() => setActiveTab("image")}
            className={`flex items-center gap-2 px-4 py-2 text-xs font-semibold rounded-lg transition-all ${
              activeTab === "image"
                ? "bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md"
                : "text-zinc-400 hover:text-white"
            }`}
          >
            <span className="material-symbols-outlined text-base">palette</span>
            ChatGPT Gambar (DALL-E & Flux)
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("chat")}
            className={`flex items-center gap-2 px-4 py-2 text-xs font-semibold rounded-lg transition-all ${
              activeTab === "chat"
                ? "bg-gradient-to-r from-purple-600 to-indigo-600 text-white shadow-md"
                : "text-zinc-400 hover:text-white"
            }`}
          >
            <span className="material-symbols-outlined text-base">chat</span>
            ChatGPT & Vision (Chat / OCR)
          </button>
        </div>
      </div>

      {/* 2. Auto-Configured API Key Card */}
      <div className="bg-zinc-900/80 border border-zinc-800 rounded-xl p-4 shadow-sm">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <span className="material-symbols-outlined text-emerald-400 text-xl">key</span>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Gateway API Key</span>
                <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.2 rounded-full font-mono">
                  Otomatis Terdeteksi
                </span>
              </div>
              <div className="font-mono text-sm text-zinc-200 mt-0.5 flex items-center gap-2">
                <span>{showKeySecret ? selectedKey : maskedKey}</span>
                {selectedKey && (
                  <button
                    type="button"
                    onClick={() => setShowKeySecret(!showKeySecret)}
                    className="text-zinc-400 hover:text-zinc-200 transition-colors"
                    title={showKeySecret ? "Hide Key" : "Show Key"}
                  >
                    <span className="material-symbols-outlined text-base align-middle">
                      {showKeySecret ? "visibility_off" : "visibility"}
                    </span>
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {keys.length > 1 && (
              <select
                value={selectedKey}
                onChange={(e) => setSelectedKey(e.target.value)}
                className="bg-zinc-800 text-xs text-zinc-200 border border-zinc-700 rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                {keys.map((k) => (
                  <option key={k.id} value={k.key}>
                    {k.name} ({k.key.slice(0, 8)}...)
                  </option>
                ))}
              </select>
            )}

            {selectedKey && (
              <button
                type="button"
                onClick={() => copy(selectedKey)}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 transition-colors"
              >
                <span className="material-symbols-outlined text-sm">
                  {copied ? "check" : "content_copy"}
                </span>
                {copied ? "Tersalin!" : "Salin Key"}
              </button>
            )}
          </div>
        </div>
      </div>

      {/* ----------------------------------------------------------------- */}
      {/* TAB 1: CHATGPT GAMBAR / DALL-E & FLUX IMAGE GENERATION             */}
      {/* ----------------------------------------------------------------- */}
      {activeTab === "image" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Controls Column */}
          <div className="lg:col-span-5 space-y-4">
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2 border-b border-zinc-800/80 pb-3">
                <span className="material-symbols-outlined text-purple-400 text-lg">brush</span>
                Konfigurasi Gambar (Image Config)
              </h2>

              {/* Model Picker */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                  Model Gambar (AI Provider)
                </label>
                <select
                  value={imageModel}
                  onChange={(e) => {
                    setImageModel(e.target.value);
                    setImageCustomModel("");
                  }}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-purple-500"
                >
                  {IMAGE_PRESETS.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name}
                    </option>
                  ))}
                  <option value="custom">✏️ Custom Model ID...</option>
                </select>

                {imageModel === "custom" && (
                  <input
                    type="text"
                    placeholder="Contoh: openai/dall-e-3 atau flux-1-schnell"
                    value={imageCustomModel}
                    onChange={(e) => setImageCustomModel(e.target.value)}
                    className="mt-2 w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-purple-500"
                  />
                )}
              </div>

              {/* Resolution & Quality */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                    Resolusi / Aspek Rasio
                  </label>
                  <select
                    value={imageSize}
                    onChange={(e) => setImageSize(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-purple-500"
                  >
                    <option value="1024x1024">1024x1024 (Kotak 1:1)</option>
                    <option value="1024x1792">1024x1792 (Portrait 9:16)</option>
                    <option value="1792x1024">1792x1024 (Landscape 16:9)</option>
                    <option value="512x512">512x512 (DALL-E 2)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                    Kualitas Render
                  </label>
                  <select
                    value={imageQuality}
                    onChange={(e) => setImageQuality(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-purple-500"
                  >
                    <option value="standard">Standard (Cepat)</option>
                    <option value="hd">HD (High Detail)</option>
                  </select>
                </div>
              </div>

              {/* Style */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                  Gaya Gambar (Style)
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setImageStyle("vivid")}
                    className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-all ${
                      imageStyle === "vivid"
                        ? "bg-purple-600/20 border-purple-500 text-purple-300 font-semibold"
                        : "bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-white"
                    }`}
                  >
                    ✨ Vivid (Dramatis/Tajam)
                  </button>
                  <button
                    type="button"
                    onClick={() => setImageStyle("natural")}
                    className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-all ${
                      imageStyle === "natural"
                        ? "bg-purple-600/20 border-purple-500 text-purple-300 font-semibold"
                        : "bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-white"
                    }`}
                  >
                    🍃 Natural (Realistis/Alami)
                  </button>
                </div>
              </div>

              {/* Prompt Textarea */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-medium text-zinc-300">
                    Prompt Gambar (Deskripsi Visual)
                  </label>
                  <span className="text-[10px] text-zinc-500">{imagePrompt.length} karakter</span>
                </div>
                <textarea
                  rows={4}
                  value={imagePrompt}
                  onChange={(e) => setImagePrompt(e.target.value)}
                  placeholder="Ketik deskripsi gambar yang ingin dibuat... (Contoh: Kucing cyberpunk dengan kacamata neon di kota masa depan)"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Sample Prompts Chips */}
              <div>
                <span className="text-[11px] font-medium text-zinc-400 block mb-2">
                  💡 Contoh Prompt Cepat:
                </span>
                <div className="flex flex-col gap-1.5">
                  {IMAGE_SAMPLE_PROMPTS.map((p, idx) => (
                    <button
                      key={idx}
                      type="button"
                      onClick={() => setImagePrompt(p)}
                      className="text-left text-[11px] p-2 rounded-lg bg-zinc-950/60 hover:bg-zinc-800/80 text-zinc-300 hover:text-white border border-zinc-800/60 transition-colors line-clamp-1"
                    >
                      &quot;{p}&quot;
                    </button>
                  ))}
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="button"
                disabled={imageGenerating || !imagePrompt.trim()}
                onClick={handleGenerateImage}
                className="w-full py-2.5 px-4 rounded-lg bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 disabled:opacity-50 text-white font-medium text-xs shadow-lg shadow-purple-600/25 flex items-center justify-center gap-2 transition-all"
              >
                {imageGenerating ? (
                  <>
                    <span className="animate-spin material-symbols-outlined text-base">refresh</span>
                    <span>Menghasilkan Gambar ({imageTimer}s)...</span>
                  </>
                ) : (
                  <>
                    <span className="material-symbols-outlined text-base">auto_awesome</span>
                    <span>Generate Gambar (DALL-E 3)</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Results Column */}
          <div className="lg:col-span-7 space-y-4">
            {/* Main Result Display */}
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-5 min-h-[480px] flex flex-col justify-between">
              <div className="flex items-center justify-between border-b border-zinc-800/80 pb-3 mb-4">
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <span className="material-symbols-outlined text-indigo-400 text-lg">image</span>
                  Hasil Render Langsung (Live Preview)
                </h3>
                {imageCurrentResult && (
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        const link = document.createElement("a");
                        link.href = imageCurrentResult.url;
                        link.download = `9router-${Date.now()}.png`;
                        link.target = "_blank";
                        link.click();
                      }}
                      className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 flex items-center gap-1.5 transition-colors"
                    >
                      <span className="material-symbols-outlined text-sm">download</span>
                      Download
                    </button>
                    <button
                      type="button"
                      onClick={() => setLightboxImage(imageCurrentResult.url)}
                      className="px-3 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 flex items-center gap-1.5 transition-colors"
                    >
                      <span className="material-symbols-outlined text-sm">fullscreen</span>
                      Perbesar
                    </button>
                  </div>
                )}
              </div>

              {/* Error Display */}
              {imageError && (
                <div className="bg-rose-950/40 border border-rose-800/50 rounded-lg p-3 text-rose-300 text-xs flex items-start gap-2 mb-4">
                  <span className="material-symbols-outlined text-rose-400 text-base flex-shrink-0 mt-0.5">error</span>
                  <div>
                    <span className="font-semibold block">Gagal Membuat Gambar:</span>
                    <span className="text-rose-200/90 font-mono text-[11px] break-all">{imageError}</span>
                  </div>
                </div>
              )}

              {/* Generating Skeleton */}
              {imageGenerating && (
                <div className="flex-1 flex flex-col items-center justify-center py-16 text-center space-y-4">
                  <div className="relative">
                    <div className="w-20 h-20 rounded-2xl bg-purple-500/10 border border-purple-500/30 flex items-center justify-center animate-pulse">
                      <span className="material-symbols-outlined text-3xl text-purple-400 animate-spin">
                        progress_activity
                      </span>
                    </div>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-white">Memproses Permintaan Gambar...</h4>
                    <p className="text-xs text-zinc-400 mt-1 max-w-sm">
                      Model <span className="font-mono text-purple-300">{imageCustomModel || imageModel}</span> sedang merender gambar dengan resolusi {imageSize}.
                    </p>
                    <span className="inline-block mt-3 px-3 py-1 rounded-full bg-zinc-800 text-zinc-300 font-mono text-xs">
                      ⏱️ {imageTimer} detik
                    </span>
                  </div>
                </div>
              )}

              {/* Rendered Image */}
              {!imageGenerating && imageCurrentResult && (
                <div className="flex-1 flex flex-col items-center justify-center space-y-3">
                  <div
                    onClick={() => setLightboxImage(imageCurrentResult.url)}
                    className="relative group cursor-pointer max-w-lg rounded-xl overflow-hidden border border-zinc-800 shadow-2xl transition-transform hover:scale-[1.01]"
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={imageCurrentResult.url}
                      alt={imageCurrentResult.prompt}
                      className="w-full h-auto object-contain rounded-xl max-h-[460px] bg-black"
                    />
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                      <span className="px-3 py-1.5 bg-black/80 rounded-lg text-white text-xs font-medium flex items-center gap-1.5 shadow-lg">
                        <span className="material-symbols-outlined text-sm">zoom_in</span>
                        Klik untuk perbesar
                      </span>
                    </div>
                  </div>

                  <div className="w-full bg-zinc-950 p-3 rounded-lg border border-zinc-800/80 text-[11px] text-zinc-300 space-y-1">
                    <p>
                      <strong className="text-zinc-400">Prompt:</strong> {imageCurrentResult.prompt}
                    </p>
                    {imageCurrentResult.revisedPrompt && imageCurrentResult.revisedPrompt !== imageCurrentResult.prompt && (
                      <p className="text-zinc-500 italic">
                        <strong className="text-zinc-400 not-italic">DALL-E Revised:</strong> &quot;{imageCurrentResult.revisedPrompt}&quot;
                      </p>
                    )}
                  </div>
                </div>
              )}

              {/* Empty Placeholder */}
              {!imageGenerating && !imageCurrentResult && !imageError && (
                <div className="flex-1 flex flex-col items-center justify-center py-16 text-center text-zinc-500">
                  <div className="w-16 h-16 rounded-full bg-zinc-800/50 flex items-center justify-center mb-3">
                    <span className="material-symbols-outlined text-3xl text-zinc-400">add_photo_alternate</span>
                  </div>
                  <h4 className="text-sm font-medium text-zinc-300">Belum ada gambar yang digenerate</h4>
                  <p className="text-xs text-zinc-500 mt-1 max-w-sm">
                    Tuliskan prompt di sebelah kiri dan klik &quot;Generate Gambar&quot; untuk menguji endpoint DALL-E 3 atau Flux.
                  </p>
                </div>
              )}

              {/* Inspector Footer */}
              <div className="mt-4 pt-3 border-t border-zinc-800/80 flex items-center justify-between text-xs text-zinc-400">
                <button
                  type="button"
                  onClick={() => setImageRawInspector(!imageRawInspector)}
                  className="flex items-center gap-1.5 hover:text-white transition-colors"
                >
                  <span className="material-symbols-outlined text-sm">code</span>
                  <span>{imageRawInspector ? "Sembunyikan Raw JSON" : "Lihat Raw JSON Request & Response"}</span>
                </button>
              </div>

              {/* Raw JSON Accordion */}
              {imageRawInspector && (
                <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3 text-[11px] font-mono">
                  <div className="bg-zinc-950 p-3 rounded-lg border border-zinc-800 overflow-x-auto">
                    <span className="text-purple-400 font-bold block mb-1">POST /v1/images/generations Payload</span>
                    <pre className="text-zinc-300">{JSON.stringify(imageLastPayload, null, 2)}</pre>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-lg border border-zinc-800 overflow-x-auto">
                    <span className="text-emerald-400 font-bold block mb-1">Response JSON</span>
                    <pre className="text-zinc-300">
                      {imageLastRawResponse ? JSON.stringify(imageLastRawResponse, null, 2) : "Belum ada response"}
                    </pre>
                  </div>
                </div>
              )}
            </div>

            {/* Gallery of Session Images */}
            {imageGallery.length > 1 && (
              <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-4">
                <h4 className="text-xs font-semibold text-zinc-300 mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-base text-zinc-400">photo_library</span>
                  Riwayat Sesi Ini ({imageGallery.length} gambar)
                </h4>
                <div className="grid grid-cols-4 sm:grid-cols-6 gap-2">
                  {imageGallery.map((item) => (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => setImageCurrentResult(item)}
                      className={`relative aspect-square rounded-lg overflow-hidden border transition-all ${
                        imageCurrentResult?.id === item.id
                          ? "border-purple-500 ring-2 ring-purple-500/30"
                          : "border-zinc-800 hover:border-zinc-700 opacity-80 hover:opacity-100"
                      }`}
                    >
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img src={item.url} alt={item.prompt} className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ----------------------------------------------------------------- */}
      {/* TAB 2: CHATGPT & VISION (CHAT, MULTIMODAL, OCR)                    */}
      {/* ----------------------------------------------------------------- */}
      {activeTab === "chat" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Settings & Input Column */}
          <div className="lg:col-span-5 space-y-4">
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2 border-b border-zinc-800/80 pb-3">
                <span className="material-symbols-outlined text-indigo-400 text-lg">tune</span>
                Parameter Chat & Vision
              </h2>

              {/* Model Picker */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                  Pilih Model AI (Chat / Multimodal Vision)
                </label>
                <select
                  value={chatModel}
                  onChange={(e) => {
                    setChatModel(e.target.value);
                    setChatCustomModel("");
                  }}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                >
                  <optgroup label="Model Populer & ChatGPT">
                    {CHAT_PRESETS.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name} ({m.id})
                      </option>
                    ))}
                  </optgroup>
                  {availableModels.length > 0 && (
                    <optgroup label="Tersedia di Gateway 9Router">
                      {availableModels.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.id}
                        </option>
                      ))}
                    </optgroup>
                  )}
                  <option value="custom">✏️ Custom Model ID...</option>
                </select>

                {chatModel === "custom" && (
                  <input
                    type="text"
                    placeholder="Contoh: gpt-4o atau claude-3-5-sonnet"
                    value={chatCustomModel}
                    onChange={(e) => setChatCustomModel(e.target.value)}
                    className="mt-2 w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-indigo-500"
                  />
                )}
              </div>

              {/* System Prompt */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                  System Instructions (Opsional)
                </label>
                <input
                  type="text"
                  value={systemPrompt}
                  onChange={(e) => setSystemPrompt(e.target.value)}
                  placeholder="You are a helpful assistant..."
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              {/* Temperature & Stream Toggle */}
              <div className="grid grid-cols-2 gap-3 items-center">
                <div>
                  <div className="flex justify-between text-xs text-zinc-300 mb-1">
                    <span>Temperature</span>
                    <span className="font-mono text-indigo-400">{chatTemperature}</span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={chatTemperature}
                    onChange={(e) => setChatTemperature(parseFloat(e.target.value))}
                    className="w-full accent-indigo-500"
                  />
                </div>
                <div className="flex items-center justify-between p-2 bg-zinc-950 rounded-lg border border-zinc-800">
                  <span className="text-xs text-zinc-300 font-medium">Streaming (SSE)</span>
                  <input
                    type="checkbox"
                    checked={chatStream}
                    onChange={(e) => setChatStream(e.target.checked)}
                    className="w-4 h-4 accent-indigo-500 rounded cursor-pointer"
                  />
                </div>
              </div>

              {/* Multimodal Vision Upload */}
              <div className="border-t border-zinc-800/80 pt-4">
                <div className="flex items-center justify-between mb-2">
                  <label className="text-xs font-semibold text-white flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-emerald-400 text-base">visibility</span>
                    Uji ChatGPT Vision (Kirim Gambar)
                  </label>
                  {chatImageDataUrl && (
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="text-[11px] text-rose-400 hover:text-rose-300 transition-colors"
                    >
                      Hapus Gambar
                    </button>
                  )}
                </div>

                {!chatImageDataUrl ? (
                  <div
                    onClick={() => fileInputRef.current?.click()}
                    className="border-2 border-dashed border-zinc-800 hover:border-indigo-500/50 rounded-xl p-4 text-center cursor-pointer transition-colors bg-zinc-950/40 group"
                  >
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => handleImageUpload(e.target.files?.[0])}
                    />
                    <span className="material-symbols-outlined text-2xl text-zinc-500 group-hover:text-indigo-400 transition-colors">
                      cloud_upload
                    </span>
                    <p className="text-xs text-zinc-300 font-medium mt-1">
                      Klik untuk upload gambar atau drag file ke sini
                    </p>
                    <p className="text-[10px] text-zinc-500 mt-0.5">PNG, JPG, WEBP didukung untuk ChatGPT Vision</p>
                  </div>
                ) : (
                  <div className="relative rounded-lg overflow-hidden border border-zinc-800 bg-black flex items-center justify-center max-h-48">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={chatImageDataUrl} alt="Vision input" className="max-h-48 object-contain" />
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="absolute top-2 right-2 w-6 h-6 rounded-full bg-rose-600/90 text-white flex items-center justify-center shadow-md hover:bg-rose-500"
                    >
                      <span className="material-symbols-outlined text-sm">close</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Sample Vision Prompts */}
              {chatImageDataUrl && (
                <div>
                  <span className="text-[11px] font-medium text-zinc-400 block mb-1.5">
                    💡 Rekomendasi Pertanyaan Vision:
                  </span>
                  <div className="flex flex-col gap-1">
                    {VISION_SAMPLE_PROMPTS.map((p, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setChatPrompt(p)}
                        className="text-left text-[11px] p-1.5 rounded-lg bg-zinc-950/60 hover:bg-zinc-800/80 text-zinc-300 hover:text-white border border-zinc-800/60 transition-colors line-clamp-1"
                      >
                        &quot;{p}&quot;
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Chat Prompt Input */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1.5">
                  Pesan / Pertanyaan (Prompt)
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
                  placeholder="Ketik pertanyaan untuk model AI... (Tekan Ctrl+Enter untuk kirim)"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Send Button */}
              <button
                type="button"
                disabled={chatSending || (!chatPrompt.trim() && !chatImageDataUrl)}
                onClick={handleSendChat}
                className="w-full py-2.5 px-4 rounded-lg bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 disabled:opacity-50 text-white font-medium text-xs shadow-lg shadow-indigo-600/25 flex items-center justify-center gap-2 transition-all"
              >
                {chatSending ? (
                  <>
                    <span className="animate-spin material-symbols-outlined text-base">refresh</span>
                    <span>Menunggu Respon AI...</span>
                  </>
                ) : (
                  <>
                    <span className="material-symbols-outlined text-base">send</span>
                    <span>Kirim ke Model (POST /v1/chat/completions)</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Response Column */}
          <div className="lg:col-span-7 space-y-4">
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-5 min-h-[480px] flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between border-b border-zinc-800/80 pb-3 mb-4">
                  <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                    <span className="material-symbols-outlined text-emerald-400 text-lg">smart_toy</span>
                    Respon AI (Response Output)
                  </h3>

                  {chatTelemetry && (
                    <div className="flex items-center gap-2 text-[11px] font-mono">
                      <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        HTTP {chatTelemetry.status}
                      </span>
                      <span className="text-zinc-400">⚡ {chatTelemetry.latencyMs}ms</span>
                      {chatTelemetry.chunks && <span className="text-zinc-400">📦 {chatTelemetry.chunks} chunks</span>}
                      {chatTelemetry.tokens && <span className="text-zinc-400">🪙 {chatTelemetry.tokens} tok</span>}
                    </div>
                  )}
                </div>

                {/* Error Box */}
                {chatError && (
                  <div className="bg-rose-950/40 border border-rose-800/50 rounded-lg p-3 text-rose-300 text-xs flex items-start gap-2 mb-4">
                    <span className="material-symbols-outlined text-rose-400 text-base flex-shrink-0 mt-0.5">error</span>
                    <div>
                      <span className="font-semibold block">Error dari Gateway:</span>
                      <span className="text-rose-200/90 font-mono text-[11px] break-all">{chatError}</span>
                    </div>
                  </div>
                )}

                {/* Response Text Display */}
                {chatResponse ? (
                  <div className="bg-zinc-950 p-4 rounded-xl border border-zinc-800 text-xs text-zinc-100 font-sans leading-relaxed whitespace-pre-wrap max-h-[500px] overflow-y-auto custom-scrollbar select-text">
                    {chatResponse}
                  </div>
                ) : chatSending ? (
                  <div className="flex flex-col items-center justify-center py-24 text-center">
                    <span className="material-symbols-outlined text-3xl text-indigo-400 animate-spin mb-3">
                      progress_activity
                    </span>
                    <p className="text-xs text-zinc-400">Streaming respon dari gateway 9router...</p>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-24 text-center text-zinc-500">
                    <div className="w-16 h-16 rounded-full bg-zinc-800/50 flex items-center justify-center mb-3">
                      <span className="material-symbols-outlined text-3xl text-zinc-400">chat_bubble_outline</span>
                    </div>
                    <h4 className="text-sm font-medium text-zinc-300">Belum ada respon</h4>
                    <p className="text-xs text-zinc-500 mt-1 max-w-sm">
                      Ketik prompt di samping dan kirim untuk melihat respon streaming langsung di antarmuka browser.
                    </p>
                  </div>
                )}
              </div>

              {/* Inspector & Copy Tools */}
              <div className="mt-4 pt-3 border-t border-zinc-800/80 flex items-center justify-between text-xs text-zinc-400">
                <button
                  type="button"
                  onClick={() => setChatRawInspector(!chatRawInspector)}
                  className="flex items-center gap-1.5 hover:text-white transition-colors"
                >
                  <span className="material-symbols-outlined text-sm">terminal</span>
                  <span>{chatRawInspector ? "Tutup Inspector" : "Buka JSON Payload Inspector"}</span>
                </button>

                {chatResponse && (
                  <button
                    type="button"
                    onClick={() => copy(chatResponse)}
                    className="flex items-center gap-1.5 px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 transition-colors"
                  >
                    <span className="material-symbols-outlined text-sm">
                      {copied ? "check" : "content_copy"}
                    </span>
                    {copied ? "Tersalin!" : "Salin Jawaban"}
                  </button>
                )}
              </div>

              {/* Raw JSON Accordion */}
              {chatRawInspector && (
                <div className="mt-3 bg-zinc-950 p-3 rounded-lg border border-zinc-800 overflow-x-auto text-[11px] font-mono">
                  <span className="text-indigo-400 font-bold block mb-1">Live Request Body:</span>
                  <pre className="text-zinc-300">{JSON.stringify(chatLastPayload, null, 2)}</pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Lightbox Modal */}
      {lightboxImage && (
        <div
          onClick={() => setLightboxImage(null)}
          className="fixed inset-0 z-50 bg-black/90 backdrop-blur-md flex items-center justify-center p-4 cursor-pointer"
        >
          <div className="relative max-w-4xl max-h-[90vh]">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src={lightboxImage} alt="Full preview" className="max-w-full max-h-[90vh] object-contain rounded-xl shadow-2xl" />
            <button
              type="button"
              onClick={() => setLightboxImage(null)}
              className="absolute top-3 right-3 w-8 h-8 rounded-full bg-black/80 text-white flex items-center justify-center shadow-lg hover:bg-zinc-800"
            >
              <span className="material-symbols-outlined text-base">close</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

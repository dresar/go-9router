"use client";

import { useState, useEffect, useRef } from "react";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";

const CHAT_PRESETS = [
  { id: "openai/gpt-4o", name: "GPT-4o", provider: "openai" },
  { id: "openai/chatgpt-4o-latest", name: "ChatGPT 4o Latest", provider: "openai" },
  { id: "openai/gpt-4o-mini", name: "GPT-4o Mini", provider: "openai" },
  { id: "claude-3-5-sonnet-20241022", name: "Claude 3.5 Sonnet", provider: "claude" },
  { id: "gemini/default", name: "Gemini 2.5 Flash", provider: "gemini" },
  { id: "deepseek/default", name: "DeepSeek Chat", provider: "deepseek" },
  { id: "grok-3", name: "Grok 3", provider: "xai" },
  { id: "openrouter/default", name: "OpenRouter Default", provider: "openrouter" },
];

const IMAGE_PRESETS = [
  { id: "openai/dall-e-3", name: "DALL-E 3", provider: "openai" },
  { id: "openai/dall-e-2", name: "DALL-E 2", provider: "openai" },
  { id: "openrouter/black-forest-labs/flux-1-schnell", name: "FLUX 1 Schnell", provider: "openrouter" },
  { id: "openrouter/black-forest-labs/flux-1-dev", name: "FLUX 1 Dev", provider: "openrouter" },
  { id: "grok-imagine", name: "Grok Imagine", provider: "xai" },
];

const VISION_SAMPLE_PROMPTS = [
  "Deskripsikan objek utama dalam gambar ini.",
  "Ekstrak seluruh teks dalam gambar ini (OCR).",
  "Analisis komposisi warna dan suasana gambar ini.",
];

const IMAGE_SAMPLE_PROMPTS = [
  "Cyberpunk cat wearing neon sunglasses in Tokyo rain, vibrant 8k render",
  "Minimalist 3D isometric glowing AI router floating in space",
  "Warm cozy coffee shop corner on a rainy morning, watercolor illustration",
  "Hyper-detailed portrait of a cybernetic warrior with holographic visor",
];

export default function PlaygroundClient() {
  const [activeTab, setActiveTab] = useState("image"); // "image" or "chat"

  // API Key Management
  const [keys, setKeys] = useState([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [showKeySecret, setShowKeySecret] = useState(false);
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
        console.error("Failed to load initial data:", err);
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
    <div className="w-full space-y-6 pb-12">
      {/* 1. Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-4">
        <div>
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-600 text-white shadow-md">
              <span className="material-symbols-outlined text-xl">science</span>
            </div>
            <div>
              <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
                Playground
                <span className="text-[10px] px-2 py-0.5 rounded font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  Browser
                </span>
              </h1>
              <p className="text-xs text-zinc-400">
                Uji model dan gambar langsung di browser.
              </p>
            </div>
          </div>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center bg-zinc-900 border border-zinc-800 p-1 rounded-lg">
          <button
            type="button"
            onClick={() => setActiveTab("image")}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all ${
              activeTab === "image"
                ? "bg-purple-600 text-white shadow-sm"
                : "text-zinc-400 hover:text-white"
            }`}
          >
            <span className="material-symbols-outlined text-base">palette</span>
            Gambar
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("chat")}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all ${
              activeTab === "chat"
                ? "bg-purple-600 text-white shadow-sm"
                : "text-zinc-400 hover:text-white"
            }`}
          >
            <span className="material-symbols-outlined text-base">chat</span>
            Chat
          </button>
        </div>
      </div>

      {/* 2. Auto-Configured API Key Card */}
      <div className="bg-zinc-900/80 border border-zinc-800 rounded-xl p-3.5 shadow-sm">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <span className="material-symbols-outlined text-emerald-400 text-lg">key</span>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-zinc-400">API Key</span>
                <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-1.5 py-0.5 rounded font-mono">
                  Otomatis
                </span>
              </div>
              <div className="font-mono text-xs text-zinc-200 mt-0.5 flex items-center gap-2">
                <span>{showKeySecret ? selectedKey : maskedKey}</span>
                {selectedKey && (
                  <button
                    type="button"
                    onClick={() => setShowKeySecret(!showKeySecret)}
                    className="text-zinc-400 hover:text-zinc-200 transition-colors"
                  >
                    <span className="material-symbols-outlined text-sm align-middle">
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
                    {k.name}
                  </option>
                ))}
              </select>
            )}

            {selectedKey && (
              <button
                type="button"
                onClick={() => copy(selectedKey)}
                className="flex items-center gap-1 px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 transition-colors"
              >
                <span className="material-symbols-outlined text-sm">
                  {copied ? "check" : "content_copy"}
                </span>
                {copied ? "✓ Disalin!" : "Salin"}
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
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-4 space-y-4">
              <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5 border-b border-zinc-800/80 pb-2">
                <span className="material-symbols-outlined text-purple-400 text-base">brush</span>
                Parameter
              </h2>

              {/* Model Picker */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1">
                  Model
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
                  <option value="custom">Custom</option>
                </select>

                {imageModel === "custom" && (
                  <input
                    type="text"
                    placeholder="Model"
                    value={imageCustomModel}
                    onChange={(e) => setImageCustomModel(e.target.value)}
                    className="mt-2 w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-purple-500"
                  />
                )}
              </div>

              {/* Resolution & Quality */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-zinc-300 mb-1">
                    Resolusi
                  </label>
                  <select
                    value={imageSize}
                    onChange={(e) => setImageSize(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-purple-500"
                  >
                    <option value="1024x1024">1024x1024 (1:1)</option>
                    <option value="1024x1792">1024x1792 (9:16)</option>
                    <option value="1792x1024">1792x1024 (16:9)</option>
                    <option value="512x512">512x512</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-zinc-300 mb-1">
                    Kualitas
                  </label>
                  <select
                    value={imageQuality}
                    onChange={(e) => setImageQuality(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-purple-500"
                  >
                    <option value="standard">Standard</option>
                    <option value="hd">HD</option>
                  </select>
                </div>
              </div>

              {/* Style */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1">
                  Gaya
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
                    Vivid
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
                    Natural
                  </button>
                </div>
              </div>

              {/* Prompt Textarea */}
              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs font-medium text-zinc-300">
                    Prompt
                  </label>
                  <span className="text-[10px] text-zinc-500">{imagePrompt.length} karakter</span>
                </div>
                <textarea
                  rows={4}
                  value={imagePrompt}
                  onChange={(e) => setImagePrompt(e.target.value)}
                  placeholder="Prompt"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Sample Prompts Chips */}
              <div>
                <span className="text-[11px] font-medium text-zinc-400 block mb-1.5">
                  Contoh Prompt:
                </span>
                <div className="flex flex-col gap-1.5">
                  {IMAGE_SAMPLE_PROMPTS.map((p, idx) => (
                    <button
                      key={idx}
                      type="button"
                      onClick={() => setImagePrompt(p)}
                      className="text-left text-[11px] p-2 rounded-lg bg-zinc-950/60 hover:bg-zinc-800/80 text-zinc-300 hover:text-white border border-zinc-800/60 transition-colors line-clamp-1"
                    >
                      {p}
                    </button>
                  ))}
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="button"
                disabled={imageGenerating || !imagePrompt.trim()}
                onClick={handleGenerateImage}
                className="w-full py-2.5 px-4 rounded-lg bg-purple-600 hover:bg-purple-500 disabled:opacity-50 text-white font-medium text-xs shadow-md flex items-center justify-center gap-1.5 transition-all"
              >
                {imageGenerating ? (
                  <>
                    <span className="animate-spin material-symbols-outlined text-base">refresh</span>
                    <span>Merender ({imageTimer}s)...</span>
                  </>
                ) : (
                  <>
                    <span className="material-symbols-outlined text-base">auto_awesome</span>
                    <span>Generate</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Results Column */}
          <div className="lg:col-span-7 space-y-4">
            {/* Main Result Display */}
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-4 min-h-[460px] flex flex-col justify-between">
              <div className="flex items-center justify-between border-b border-zinc-800/80 pb-2 mb-3">
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5">
                  <span className="material-symbols-outlined text-indigo-400 text-base">image</span>
                  Hasil
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
                      className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 flex items-center gap-1 transition-colors"
                    >
                      <span className="material-symbols-outlined text-sm">download</span>
                      Download
                    </button>
                    <button
                      type="button"
                      onClick={() => setLightboxImage(imageCurrentResult.url)}
                      className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 flex items-center gap-1 transition-colors"
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
                    <span className="font-semibold block">Gagal:</span>
                    <span className="text-rose-200/90 font-mono text-[11px] break-all">{imageError}</span>
                  </div>
                </div>
              )}

              {/* Generating Skeleton */}
              {imageGenerating && (
                <div className="flex-1 flex flex-col items-center justify-center py-16 text-center space-y-3">
                  <div className="w-16 h-16 rounded-xl bg-purple-500/10 border border-purple-500/30 flex items-center justify-center animate-pulse">
                    <span className="material-symbols-outlined text-2xl text-purple-400 animate-spin">
                      progress_activity
                    </span>
                  </div>
                  <div>
                    <h4 className="text-xs font-semibold text-white">Merender...</h4>
                    <span className="inline-block mt-2 px-2.5 py-0.5 rounded bg-zinc-800 text-zinc-300 font-mono text-xs">
                      {imageTimer}s
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
                      className="w-full h-auto object-contain rounded-xl max-h-[440px] bg-black"
                    />
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                      <span className="px-3 py-1 bg-black/80 rounded-lg text-white text-xs font-medium flex items-center gap-1 shadow-lg">
                        <span className="material-symbols-outlined text-sm">zoom_in</span>
                        Perbesar
                      </span>
                    </div>
                  </div>

                  <div className="w-full bg-zinc-950 p-2.5 rounded-lg border border-zinc-800/80 text-[11px] text-zinc-300">
                    <p className="line-clamp-2">
                      <strong className="text-zinc-400">Prompt:</strong> {imageCurrentResult.prompt}
                    </p>
                  </div>
                </div>
              )}

              {/* Empty Placeholder */}
              {!imageGenerating && !imageCurrentResult && !imageError && (
                <div className="flex-1 flex flex-col items-center justify-center py-16 text-center text-zinc-500">
                  <span className="material-symbols-outlined text-3xl text-zinc-400 mb-2">add_photo_alternate</span>
                  <h4 className="text-xs font-medium text-zinc-300">Belum ada hasil</h4>
                </div>
              )}

              {/* Inspector Footer */}
              <div className="mt-3 pt-2 border-t border-zinc-800/80 flex items-center justify-between text-xs text-zinc-400">
                <button
                  type="button"
                  onClick={() => setImageRawInspector(!imageRawInspector)}
                  className="flex items-center gap-1 hover:text-white transition-colors"
                >
                  <span className="material-symbols-outlined text-sm">code</span>
                  <span>{imageRawInspector ? "Tutup JSON" : "JSON"}</span>
                </button>
              </div>

              {/* Raw JSON Accordion */}
              {imageRawInspector && (
                <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px] font-mono">
                  <div className="bg-zinc-950 p-2 rounded-lg border border-zinc-800 overflow-x-auto">
                    <span className="text-purple-400 font-bold block mb-1">Payload</span>
                    <pre className="text-zinc-300">{JSON.stringify(imageLastPayload, null, 2)}</pre>
                  </div>
                  <div className="bg-zinc-950 p-2 rounded-lg border border-zinc-800 overflow-x-auto">
                    <span className="text-emerald-400 font-bold block mb-1">Response</span>
                    <pre className="text-zinc-300">
                      {imageLastRawResponse ? JSON.stringify(imageLastRawResponse, null, 2) : "Kosong"}
                    </pre>
                  </div>
                </div>
              )}
            </div>

            {/* Gallery of Session Images */}
            {imageGallery.length > 1 && (
              <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-3.5">
                <h4 className="text-xs font-semibold text-zinc-300 mb-2 flex items-center gap-1.5">
                  <span className="material-symbols-outlined text-sm text-zinc-400">photo_library</span>
                  Riwayat ({imageGallery.length})
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
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-4 space-y-4">
              <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5 border-b border-zinc-800/80 pb-2">
                <span className="material-symbols-outlined text-indigo-400 text-base">tune</span>
                Parameter
              </h2>

              {/* Model Picker */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1">
                  Model
                </label>
                <select
                  value={chatModel}
                  onChange={(e) => {
                    setChatModel(e.target.value);
                    setChatCustomModel("");
                  }}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                >
                  <optgroup label="Model">
                    {CHAT_PRESETS.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name}
                      </option>
                    ))}
                  </optgroup>
                  {availableModels.length > 0 && (
                    <optgroup label="Gateway">
                      {availableModels.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.id}
                        </option>
                      ))}
                    </optgroup>
                  )}
                  <option value="custom">Custom</option>
                </select>

                {chatModel === "custom" && (
                  <input
                    type="text"
                    placeholder="Model"
                    value={chatCustomModel}
                    onChange={(e) => setChatCustomModel(e.target.value)}
                    className="mt-2 w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-indigo-500"
                  />
                )}
              </div>

              {/* System Prompt */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1">
                  Instruksi
                </label>
                <input
                  type="text"
                  value={systemPrompt}
                  onChange={(e) => setSystemPrompt(e.target.value)}
                  placeholder="Instruksi"
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
                  <span className="text-xs text-zinc-300 font-medium">Stream</span>
                  <input
                    type="checkbox"
                    checked={chatStream}
                    onChange={(e) => setChatStream(e.target.checked)}
                    className="w-4 h-4 accent-indigo-500 rounded cursor-pointer"
                  />
                </div>
              </div>

              {/* Multimodal Vision Upload */}
              <div className="border-t border-zinc-800/80 pt-3">
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-medium text-zinc-300 flex items-center gap-1">
                    <span className="material-symbols-outlined text-emerald-400 text-sm">visibility</span>
                    Vision
                  </label>
                  {chatImageDataUrl && (
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="text-[11px] text-rose-400 hover:text-rose-300 transition-colors"
                    >
                      Hapus
                    </button>
                  )}
                </div>

                {!chatImageDataUrl ? (
                  <div
                    onClick={() => fileInputRef.current?.click()}
                    className="border border-dashed border-zinc-800 hover:border-indigo-500/50 rounded-lg p-3 text-center cursor-pointer transition-colors bg-zinc-950/40 group"
                  >
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => handleImageUpload(e.target.files?.[0])}
                    />
                    <span className="material-symbols-outlined text-xl text-zinc-500 group-hover:text-indigo-400 transition-colors">
                      cloud_upload
                    </span>
                    <p className="text-xs text-zinc-300 font-medium mt-0.5">
                      Upload gambar
                    </p>
                  </div>
                ) : (
                  <div className="relative rounded-lg overflow-hidden border border-zinc-800 bg-black flex items-center justify-center max-h-44">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={chatImageDataUrl} alt="Vision" className="max-h-44 object-contain" />
                    <button
                      type="button"
                      onClick={clearChatImage}
                      className="absolute top-2 right-2 w-5 h-5 rounded-full bg-rose-600/90 text-white flex items-center justify-center shadow-md hover:bg-rose-500"
                    >
                      <span className="material-symbols-outlined text-xs">close</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Sample Vision Prompts */}
              {chatImageDataUrl && (
                <div>
                  <span className="text-[11px] font-medium text-zinc-400 block mb-1">
                    Contoh Pertanyaan:
                  </span>
                  <div className="flex flex-col gap-1">
                    {VISION_SAMPLE_PROMPTS.map((p, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setChatPrompt(p)}
                        className="text-left text-[11px] p-1.5 rounded-lg bg-zinc-950/60 hover:bg-zinc-800/80 text-zinc-300 hover:text-white border border-zinc-800/60 transition-colors line-clamp-1"
                      >
                        {p}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Chat Prompt Input */}
              <div>
                <label className="block text-xs font-medium text-zinc-300 mb-1">
                  Pesan
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
                  placeholder="Pesan"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500 leading-relaxed custom-scrollbar"
                />
              </div>

              {/* Send Button */}
              <button
                type="button"
                disabled={chatSending || (!chatPrompt.trim() && !chatImageDataUrl)}
                onClick={handleSendChat}
                className="w-full py-2 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium text-xs shadow-md flex items-center justify-center gap-1.5 transition-all"
              >
                {chatSending ? (
                  <>
                    <span className="animate-spin material-symbols-outlined text-base">refresh</span>
                    <span>Mengirim...</span>
                  </>
                ) : (
                  <>
                    <span className="material-symbols-outlined text-base">send</span>
                    <span>Kirim</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Response Column */}
          <div className="lg:col-span-7 space-y-4">
            <div className="bg-zinc-900/90 border border-zinc-800 rounded-xl p-4 min-h-[460px] flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between border-b border-zinc-800/80 pb-2 mb-3">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-emerald-400 text-base">smart_toy</span>
                    Respon
                  </h3>

                  {chatTelemetry && (
                    <div className="flex items-center gap-2 text-[11px] font-mono">
                      <span className="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        {chatTelemetry.status}
                      </span>
                      <span className="text-zinc-400">{chatTelemetry.latencyMs}ms</span>
                      {chatTelemetry.chunks && <span className="text-zinc-400">{chatTelemetry.chunks} chunks</span>}
                      {chatTelemetry.tokens && <span className="text-zinc-400">{chatTelemetry.tokens} tok</span>}
                    </div>
                  )}
                </div>

                {/* Error Box */}
                {chatError && (
                  <div className="bg-rose-950/40 border border-rose-800/50 rounded-lg p-2.5 text-rose-300 text-xs flex items-start gap-2 mb-3">
                    <span className="material-symbols-outlined text-rose-400 text-base flex-shrink-0 mt-0.5">error</span>
                    <div>
                      <span className="font-semibold block">Gagal:</span>
                      <span className="text-rose-200/90 font-mono text-[11px] break-all">{chatError}</span>
                    </div>
                  </div>
                )}

                {/* Response Text Display */}
                {chatResponse ? (
                  <div className="bg-zinc-950 p-3.5 rounded-xl border border-zinc-800 text-xs text-zinc-100 font-sans leading-relaxed whitespace-pre-wrap max-h-[480px] overflow-y-auto custom-scrollbar select-text">
                    {chatResponse}
                  </div>
                ) : chatSending ? (
                  <div className="flex flex-col items-center justify-center py-20 text-center">
                    <span className="material-symbols-outlined text-2xl text-indigo-400 animate-spin mb-2">
                      progress_activity
                    </span>
                    <p className="text-xs text-zinc-400">Mengalirkan data...</p>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-20 text-center text-zinc-500">
                    <span className="material-symbols-outlined text-2xl text-zinc-400 mb-2">chat_bubble_outline</span>
                    <h4 className="text-xs font-medium text-zinc-300">Belum ada pesan</h4>
                  </div>
                )}
              </div>

              {/* Inspector & Copy Tools */}
              <div className="mt-3 pt-2 border-t border-zinc-800/80 flex items-center justify-between text-xs text-zinc-400">
                <button
                  type="button"
                  onClick={() => setChatRawInspector(!chatRawInspector)}
                  className="flex items-center gap-1 hover:text-white transition-colors"
                >
                  <span className="material-symbols-outlined text-sm">terminal</span>
                  <span>{chatRawInspector ? "Tutup JSON" : "JSON"}</span>
                </button>

                {chatResponse && (
                  <button
                    type="button"
                    onClick={() => copy(chatResponse)}
                    className="flex items-center gap-1 px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium rounded-lg border border-zinc-700/60 transition-colors"
                  >
                    <span className="material-symbols-outlined text-sm">
                      {copied ? "check" : "content_copy"}
                    </span>
                    {copied ? "✓ Disalin!" : "Salin"}
                  </button>
                )}
              </div>

              {/* Raw JSON Accordion */}
              {chatRawInspector && (
                <div className="mt-2 bg-zinc-950 p-2.5 rounded-lg border border-zinc-800 overflow-x-auto text-[11px] font-mono">
                  <span className="text-indigo-400 font-bold block mb-1">Payload:</span>
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
            <img src={lightboxImage} alt="Preview" className="max-w-full max-h-[90vh] object-contain rounded-xl shadow-2xl" />
            <button
              type="button"
              onClick={() => setLightboxImage(null)}
              className="absolute top-3 right-3 w-7 h-7 rounded-full bg-black/80 text-white flex items-center justify-center shadow-lg hover:bg-zinc-800"
            >
              <span className="material-symbols-outlined text-sm">close</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

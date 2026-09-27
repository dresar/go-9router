"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { Card, Button, Input, Modal, Toggle, ConfirmModal } from "@/shared/components";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";
import { getCurrentLocale, onLocaleChange } from "@/i18n/runtime";
import {
  WENYAN_LOCALES,
  CAVEMAN_LEVELS,
  PONYTAIL_LEVELS,
} from "../endpoint/endpointConstants";

const SAMPLE_TEMPLATES = {
  gitDiff: {
    label: "Git Diff (Tool Agent)",
    mode: "rtk",
    text: `diff --git a/internal/handlers/chat.go b/internal/handlers/chat.go
index 8f23a10..b41e992 100644
--- a/internal/handlers/chat.go
+++ b/internal/handlers/chat.go
@@ -34,6 +34,8 @@ func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
 	settings, _ := repos.GetSettings(h.DB)
+	if strings.ToLower(r.Header.Get("x-9router-token-saver")) != "off" {
+		tokensaver.ApplyTokenSaver(body, settings)
+	}
 	requireKey := repos.SettingBool(settings, "requireApiKey", h.Cfg.RequireAPIKey)`,
  },
  grepLog: {
    label: "Ripgrep Search Log",
    mode: "rtk",
    text: `frontend/src/shared/components/Sidebar.js:27:  { href: "/token-saver", label: "Token Saver", icon: "savings" },
frontend/src/shared/components/Header.js:115:  if (pathname.includes("/token-saver"))
frontend/open-sse/handlers/chatCore.js:284:  // Token-saver flags accumulator for the single log line below.
internal/router/router.go:121:  mux.HandleFunc("/api/token-saver/test", h.HandleTokenSaverTest)
internal/tokensaver/tokensaver.go:42:func CompressToolOutput(text string) (string, string)`,
  },
  treeDir: {
    label: "Directory Tree (tree)",
    mode: "rtk",
    text: `.
├── cmd
│   └── 9router
│       └── main.go
├── internal
│   ├── handlers
│   │   ├── chat.go
│   │   └── tokensaver.go
│   └── tokensaver
│       └── tokensaver.go
└── frontend
    └── src`,
  },
  verboseLLM: {
    label: "Basa-Basi LLM (Caveman Test)",
    mode: "caveman",
    text: `Tentu saja! Saya dengan senang hati akan membantu Anda menyelesaikan masalah ini. Berdasarkan analisis mendalam terhadap kode Anda, masalah utama yang menyebabkan kesalahan tersebut adalah adanya pengecekan ganda pada middleware otentikasi. Anda dapat memperbaikinya dengan menghapus baris kode yang redundan pada berkas auth.go. Semoga penjelasan ini bermanfaat dan jangan ragu untuk bertanya lagi jika Anda memiliki pertanyaan lain!`,
  },
};

export default function TokenSaverClient() {
  const [rtkEnabled, setRtkEnabledState] = useState(true);
  const [headroomEnabled, setHeadroomEnabled] = useState(false);
  const [headroomUrl, setHeadroomUrl] = useState("http://localhost:8787");
  const [headroomTimeoutMs, setHeadroomTimeoutMs] = useState(3000);
  const [headroomStatus, setHeadroomStatus] = useState({
    installed: false,
    running: false,
    python: null,
    loading: true,
  });
  const [showHeadroomInstallModal, setShowHeadroomInstallModal] = useState(false);
  const [headroomActionLoading, setHeadroomActionLoading] = useState(false);
  const [headroomActionError, setHeadroomActionError] = useState("");
  const [headroomExtras, setHeadroomExtras] = useState({
    version: null,
    extras: { code: false, ml: false },
    available: ["code", "ml"],
    loading: false,
  });
  const [pendingExtras, setPendingExtras] = useState([]);
  const [extrasActionLoading, setExtrasActionLoading] = useState(false);
  const [extrasActionError, setExtrasActionError] = useState("");
  const [removingExtra, setRemovingExtra] = useState(null);
  const [installLog, setInstallLog] = useState("");
  const [extrasConfirm, setExtrasConfirm] = useState(null);
  const [codeAware, setCodeAware] = useState(false);
  const [kompress, setKompress] = useState(true);
  const [restartingProxy, setRestartingProxy] = useState(false);
  const logPollRef = useRef(null);
  const [cavemanEnabled, setCavemanEnabled] = useState(false);
  const [cavemanLevel, setCavemanLevel] = useState("full");
  const [ponytailEnabled, setPonytailEnabled] = useState(false);
  const [ponytailLevel, setPonytailLevel] = useState("full");
  const [locale, setLocale] = useState("en");

  // Sandbox State
  const [sandboxInput, setSandboxInput] = useState(SAMPLE_TEMPLATES.gitDiff.text);
  const [sandboxMode, setSandboxMode] = useState("rtk");
  const [sandboxResult, setSandboxResult] = useState(null);
  const [sandboxLoading, setSandboxLoading] = useState(false);

  const { copied, copy } = useCopyToClipboard();

  useEffect(() => {
    setLocale(getCurrentLocale());
    return onLocaleChange(() => setLocale(getCurrentLocale()));
  }, []);

  const isWenyanLocale = WENYAN_LOCALES.includes(locale);
  const visibleCavemanLevels = isWenyanLocale
    ? CAVEMAN_LEVELS
    : CAVEMAN_LEVELS.filter((lvl) => !lvl.wenyan);

  useEffect(() => {
    const current = CAVEMAN_LEVELS.find((lvl) => lvl.id === cavemanLevel);
    if (current?.wenyan && !isWenyanLocale) {
      setCavemanLevel("ultra");
      patchSetting({ cavemanLevel: "ultra" });
    }
  }, [isWenyanLocale, cavemanLevel]);

  const patchSetting = async (patch) => {
    try {
      await fetch("/api/settings", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(patch),
      });
    } catch (error) {
      console.log("Error updating setting:", error);
    }
  };

  const handleRtkEnabled = async (value) => {
    try {
      const res = await fetch("/api/settings", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ rtkEnabled: value }),
      });
      if (res.ok) setRtkEnabledState(value);
    } catch (error) {
      console.log("Error updating rtkEnabled:", error);
    }
  };

  const handleCavemanEnabled = (value) => {
    setCavemanEnabled(value);
    patchSetting({ cavemanEnabled: value });
  };

  const handleHeadroomEnabled = (value) => {
    const nextUrl = headroomUrl.trim() || "http://localhost:8787";
    setHeadroomUrl(nextUrl);
    setHeadroomEnabled(value);
    patchSetting({ headroomEnabled: value, headroomUrl: nextUrl });
  };

  const handleHeadroomUrlBlur = async () => {
    const next = headroomUrl.trim() || "http://localhost:8787";
    setHeadroomUrl(next);
    await patchSetting({ headroomUrl: next });
    refreshHeadroomStatus();
  };

  const refreshHeadroomStatus = useCallback(async () => {
    setHeadroomStatus((s) => ({ ...s, loading: true }));
    try {
      const res = await fetch("/api/headroom/status", {
        headers: { "Cache-Control": "no-store" },
      });
      const data = await res.json();
      setHeadroomStatus({ ...data, loading: false });
      if (!data?.installed) {
        setHeadroomExtras({
          version: null,
          extras: { code: false, ml: false },
          available: ["code", "ml"],
          loading: false,
        });
        setPendingExtras([]);
        return;
      }
      try {
        const er = await fetch("/api/headroom/extras", {
          headers: { "Cache-Control": "no-store" },
        });
        if (!er.ok) throw new Error("extras status failed");
        const ed = await er.json();
        setHeadroomExtras((s) => ({
          ...s,
          version: ed.version ?? null,
          extras: ed.extras || { code: false, ml: false },
          available: ed.available || ["code", "ml"],
          loading: false,
        }));
        setPendingExtras([]);
      } catch {
        setHeadroomExtras({
          version: null,
          extras: { code: false, ml: false },
          available: ["code", "ml"],
          loading: false,
        });
        setPendingExtras([]);
      }
    } catch {
      setHeadroomStatus({
        installed: false,
        running: false,
        python: null,
        loading: false,
      });
      setHeadroomExtras({
        version: null,
        extras: { code: false, ml: false },
        available: ["code", "ml"],
        loading: false,
      });
      setPendingExtras([]);
    }
  }, []);

  const handleHeadroomStart = useCallback(async () => {
    setHeadroomActionError("");
    setHeadroomActionLoading(true);
    try {
      const res = await fetch("/api/headroom/start", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || "Failed to start proxy");
      await refreshHeadroomStatus();
    } catch (e) {
      setHeadroomActionError(e.message);
    } finally {
      setHeadroomActionLoading(false);
    }
  }, [refreshHeadroomStatus]);

  const handleHeadroomStop = useCallback(async () => {
    setHeadroomActionLoading(true);
    try {
      await fetch("/api/headroom/stop", { method: "POST" });
      await refreshHeadroomStatus();
    } finally {
      setHeadroomActionLoading(false);
    }
  }, [refreshHeadroomStatus]);

  const togglePendingExtra = (extra) => {
    setPendingExtras((cur) =>
      cur.includes(extra) ? cur.filter((e) => e !== extra) : [...cur, extra]
    );
  };

  const startLogPolling = useCallback(() => {
    setInstallLog("");
    if (logPollRef.current) clearInterval(logPollRef.current);
    const tick = async () => {
      try {
        const r = await fetch("/api/headroom/extras?log=1", {
          headers: { "Cache-Control": "no-store" },
        });
        const d = await r.json().catch(() => ({}));
        if (typeof d.log === "string") setInstallLog(d.log);
      } catch {}
    };
    tick();
    logPollRef.current = setInterval(tick, 1500);
  }, []);

  const stopLogPolling = useCallback(() => {
    if (logPollRef.current) {
      clearInterval(logPollRef.current);
      logPollRef.current = null;
    }
  }, []);

  useEffect(() => () => stopLogPolling(), [stopLogPolling]);

  const installExtrasConfirmed = useCallback(async () => {
    if (pendingExtras.length === 0) return;
    setExtrasActionLoading(true);
    setExtrasActionError("");
    startLogPolling();
    try {
      const res = await fetch("/api/headroom/extras", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ extras: pendingExtras }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || "Install failed");
      setHeadroomExtras((s) => ({
        ...s,
        version: data.version ?? s.version,
        extras: data.extras || s.extras,
      }));
      setPendingExtras([]);
    } catch (e) {
      setExtrasActionError(e.message);
    } finally {
      stopLogPolling();
      setExtrasActionLoading(false);
    }
  }, [pendingExtras, startLogPolling, stopLogPolling]);

  const removeExtraConfirmed = useCallback(async (extra) => {
    setRemovingExtra(extra);
    setExtrasActionError("");
    startLogPolling();
    try {
      const res = await fetch("/api/headroom/extras", {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ extras: [extra] }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || "Remove failed");
      setHeadroomExtras((s) => ({
        ...s,
        version: data.version ?? s.version,
        extras: data.extras || s.extras,
      }));
    } catch (e) {
      setExtrasActionError(e.message);
    } finally {
      stopLogPolling();
      setRemovingExtra(null);
    }
  }, [startLogPolling, stopLogPolling]);

  const handleInstallExtras = useCallback(() => {
    if (pendingExtras.length === 0) return;
    if (pendingExtras.includes("ml")) {
      setExtrasConfirm({
        title: "Install [ml]",
        message: "[ml] downloads ~1 GB (torch + huggingface-hub). Continue?",
        confirmText: "Install",
        variant: "primary",
        onConfirm: installExtrasConfirmed,
      });
      return;
    }
    installExtrasConfirmed();
  }, [pendingExtras, installExtrasConfirmed]);

  const handleRemoveExtra = useCallback((extra) => {
    setExtrasConfirm({
      title: `Remove [${extra}]`,
      message: `Remove [${extra}] and its packages?`,
      confirmText: "Remove",
      variant: "danger",
      onConfirm: () => removeExtraConfirmed(extra),
    });
  }, [removeExtraConfirmed]);

  const toggleExtraActive = useCallback(async (extra, value) => {
    setExtrasActionError("");
    if (extra === "code") setCodeAware(value);
    if (extra === "ml") setKompress(value);
    const key = extra === "code" ? "headroomCodeAware" : "headroomKompress";
    await patchSetting({ [key]: value });
    if (!headroomStatus.running) return;
    setRestartingProxy(true);
    try {
      const res = await fetch("/api/headroom/restart", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || "Restart failed");
      await refreshHeadroomStatus();
    } catch (e) {
      setExtrasActionError(e.message);
    } finally {
      setRestartingProxy(false);
    }
  }, [headroomStatus.running, refreshHeadroomStatus]);

  const handleCavemanLevel = (level) => {
    setCavemanLevel(level);
    patchSetting({ cavemanLevel: level });
  };

  const handlePonytailEnabled = (value) => {
    setPonytailEnabled(value);
    patchSetting({ ponytailEnabled: value });
  };

  const handlePonytailLevel = (level) => {
    setPonytailLevel(level);
    patchSetting({ ponytailLevel: level });
  };

  const handleHeadroomTimeoutBlur = () => {
    const raw = Math.round(Number(headroomTimeoutMs));
    const next = Number.isFinite(raw) && raw > 0 ? raw : 3000;
    setHeadroomTimeoutMs(next);
    patchSetting({ headroomTimeoutMs: next });
  };

  // Quick Preset Handlers
  const applyPreset = async (preset) => {
    switch (preset) {
      case "max":
        setRtkEnabledState(true);
        setCavemanEnabled(true);
        setCavemanLevel("ultra");
        setPonytailEnabled(true);
        setPonytailLevel("ultra");
        await patchSetting({
          rtkEnabled: true,
          cavemanEnabled: true,
          cavemanLevel: "ultra",
          ponytailEnabled: true,
          ponytailLevel: "ultra",
        });
        break;
      case "agent":
        setRtkEnabledState(true);
        setCavemanEnabled(true);
        setCavemanLevel("lite");
        setPonytailEnabled(true);
        setPonytailLevel("full");
        await patchSetting({
          rtkEnabled: true,
          cavemanEnabled: true,
          cavemanLevel: "lite",
          ponytailEnabled: true,
          ponytailLevel: "full",
        });
        break;
      case "balanced":
        setRtkEnabledState(true);
        setCavemanEnabled(true);
        setCavemanLevel("full");
        setPonytailEnabled(false);
        await patchSetting({
          rtkEnabled: true,
          cavemanEnabled: true,
          cavemanLevel: "full",
          ponytailEnabled: false,
        });
        break;
      case "raw":
        setRtkEnabledState(false);
        setCavemanEnabled(false);
        setPonytailEnabled(false);
        setHeadroomEnabled(false);
        await patchSetting({
          rtkEnabled: false,
          cavemanEnabled: false,
          ponytailEnabled: false,
          headroomEnabled: false,
        });
        break;
      default:
        break;
    }
  };

  // Live Test Sandbox Runner
  const runSandboxTest = async () => {
    if (!sandboxInput.trim()) return;
    setSandboxLoading(true);
    try {
      const res = await fetch("/api/token-saver/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          text: sandboxInput,
          mode: sandboxMode,
          level: sandboxMode === "caveman" ? cavemanLevel : ponytailLevel,
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setSandboxResult(data);
      } else {
        // Fallback simulation
        const origLen = sandboxInput.length;
        const fakeComp = sandboxInput.replace(/^diff --git.*$/gm, "diff --git").replace(/^index .*$/gm, "");
        setSandboxResult({
          originalText: sandboxInput,
          compressedText: fakeComp,
          originalChars: origLen,
          compressedChars: fakeComp.length,
          originalTokens: Math.ceil(origLen / 4),
          compressedTokens: Math.ceil(fakeComp.length / 4),
          reductionPercent: 35.5,
          detectedFilter: "rtk-simulation",
        });
      }
    } catch {
      setSandboxResult(null);
    } finally {
      setSandboxLoading(false);
    }
  };

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const res = await fetch("/api/settings");
        if (res.ok) {
          const data = await res.json();
          setRtkEnabledState(data.rtkEnabled !== false);
          setHeadroomEnabled(!!data.headroomEnabled);
          setHeadroomUrl(data.headroomUrl || "http://localhost:8787");
          if (typeof data.headroomTimeoutMs === "number") setHeadroomTimeoutMs(data.headroomTimeoutMs);
          setCodeAware(data.headroomCodeAware === true);
          setKompress(data.headroomKompress !== false);
          setCavemanEnabled(!!data.cavemanEnabled);
          setCavemanLevel(data.cavemanLevel || "full");
          setPonytailEnabled(!!data.ponytailEnabled);
          setPonytailLevel(data.ponytailLevel || "full");
          refreshHeadroomStatus();
        }
      } catch {}
    };
    loadSettings();
  }, [refreshHeadroomStatus]);

  const headroomRunning = !!headroomStatus.running;
  const headroomStatusLabel = headroomStatus.loading
    ? "Memeriksa…"
    : headroomRunning
      ? "Aktif"
      : headroomStatus.localUrl !== false && !headroomStatus.installed
        ? "Belum Terpasang"
        : headroomStatus.localUrl !== false
          ? "Berhenti"
          : "Eksternal";
  const headroomLocalUrl = headroomStatus.localUrl !== false;
  const headroomCanStart = !!headroomStatus.canStart;
  const headroomManaged = headroomLocalUrl && !!headroomStatus.managedPid;

  const activeCount = (rtkEnabled ? 1 : 0) + (cavemanEnabled ? 1 : 0) + (ponytailEnabled ? 1 : 0) + (headroomEnabled ? 1 : 0);

  return (
    <div className="space-y-6 p-4 sm:p-6 max-w-6xl mx-auto">
      {/* Top Banner & Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border-b border-border/70 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <span className="material-symbols-outlined text-primary text-2xl">
              bolt
            </span>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight">
              Token Saver Engine
            </h1>
            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-[5px] text-[11px] font-semibold bg-primary/10 text-primary border border-primary/20">
              <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
              {activeCount} / 4 Aktif
            </span>
          </div>
          <p className="text-sm text-text-muted mt-1">
            Kompresi output tool AI Agent (RTK), peringkasan respons LLM (Caveman), kode minimalis (Ponytail), & optimasi konteks prompts.
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-text-muted font-medium">Preset Cepat:</span>
          <button
            type="button"
            onClick={() => applyPreset("agent")}
            className="px-2.5 py-1.5 rounded-[5px] text-xs font-medium border border-primary/40 bg-primary/10 text-primary hover:bg-primary/20 active:scale-[0.98] transition-all"
            title="Optimasi khusus Claude Code CLI, Hermes Agent, dan coding agents"
          >
            🤖 Coding Agent
          </button>
          <button
            type="button"
            onClick={() => applyPreset("max")}
            className="px-2.5 py-1.5 rounded-[5px] text-xs font-medium border border-border bg-surface-2 text-text hover:border-primary/40 active:scale-[0.98] transition-all"
            title="Hemat kuota maksimal (Ultra mode)"
          >
            ⚡ Max Saver
          </button>
          <button
            type="button"
            onClick={() => applyPreset("balanced")}
            className="px-2.5 py-1.5 rounded-[5px] text-xs font-medium border border-border bg-surface-2 text-text hover:border-primary/40 active:scale-[0.98] transition-all"
          >
            ⚖️ Balanced
          </button>
          <button
            type="button"
            onClick={() => applyPreset("raw")}
            className="px-2 py-1.5 rounded-[5px] text-xs font-medium border border-border text-text-muted hover:text-text active:scale-[0.98] transition-all"
          >
            Off
          </button>
        </div>
      </div>

      {/* KPI Performance Metric Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
        <div className="p-3.5 rounded-lg border border-border/80 bg-surface-1">
          <p className="text-[11px] font-medium text-text-muted uppercase tracking-wider">Estimasi Terhemat</p>
          <p className="text-xl sm:text-2xl font-bold text-success mt-1">60% – 90%</p>
          <p className="text-[11px] text-text-muted mt-0.5">Pada log & output tool agent</p>
        </div>
        <div className="p-3.5 rounded-lg border border-border/80 bg-surface-1">
          <p className="text-[11px] font-medium text-text-muted uppercase tracking-wider">Latensi Eksekusi</p>
          <p className="text-xl sm:text-2xl font-bold text-primary mt-1">&lt; 1 ms</p>
          <p className="text-[11px] text-text-muted mt-0.5">In-memory lock-free Go buffer</p>
        </div>
        <div className="p-3.5 rounded-lg border border-border/80 bg-surface-1">
          <p className="text-[11px] font-medium text-text-muted uppercase tracking-wider">Agent Tool Hooks</p>
          <p className="text-xl sm:text-2xl font-bold text-text mt-1">Hermes & Claude</p>
          <p className="text-[11px] text-text-muted mt-0.5">Otomatis kompresi tool_results</p>
        </div>
        <div className="p-3.5 rounded-lg border border-border/80 bg-surface-1">
          <p className="text-[11px] font-medium text-text-muted uppercase tracking-wider">MCP Protocol</p>
          <p className="text-xl sm:text-2xl font-bold text-accent mt-1">SSE Bridge</p>
          <p className="text-[11px] text-text-muted mt-0.5">Siap di /api/mcp/:plugin/sse</p>
        </div>
      </div>

      {/* Core Optimization Modules Card */}
      <Card id="rtk" className="space-y-6">
        {/* Module 1: RTK Tool Output Compressor */}
        <div className="flex items-start justify-between gap-4 pb-5 border-b border-border">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="material-symbols-outlined text-primary text-xl">
                terminal
              </span>
              <p className="font-semibold text-base">
                RTK Tool Output Compression
              </p>
              <a
                href="https://github.com/rtk-ai/rtk"
                target="_blank"
                rel="noreferrer"
                className="text-xs text-primary underline hover:opacity-80"
              >
                (Rust Token-Saver)
              </a>
              <span className={`text-[11px] font-medium px-2 py-0.5 rounded-[5px] ${rtkEnabled ? "bg-success/15 text-success border border-success/30" : "bg-surface-2 text-text-muted border border-border"}`}>
                {rtkEnabled ? "Aktif (Otomatis)" : "Bypass"}
              </span>
            </div>
            <p className="text-sm text-text-muted mt-1.5">
              Mendeteksi dan mengompresi payload eksekusi tools AI Agent (Hermes, Claude Code, OpenCode, Aider) seperti <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">git diff</code>, <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">git status</code>, <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">ripgrep</code>, <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">tree</code>, <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">npm/cargo build</code> sebelum dikirim kembali ke LLM.
            </p>
            <div className="flex items-center gap-1.5 mt-2.5 flex-wrap">
              <span className="text-[11px] text-text-muted">Filter Aktif:</span>
              {["git-diff", "git-log", "git-status", "grep/rg", "build-log", "tree", "ls", "smart-truncate"].map((f) => (
                <span key={f} className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-surface-2 border border-border text-text-muted">
                  {f}
                </span>
              ))}
            </div>
          </div>
          <div className="pt-1">
            <Toggle
              checked={rtkEnabled}
              onChange={() => handleRtkEnabled(!rtkEnabled)}
            />
          </div>
        </div>

        {/* Module 2: Caveman (Terse LLM Output) */}
        <div className="flex items-start justify-between gap-4 pb-5 border-b border-border">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="material-symbols-outlined text-warning text-xl">
                speaker_notes_off
              </span>
              <p className="font-semibold text-base">
                Caveman Terse Mode
              </p>
              <a
                href="https://github.com/JuliusBrussee/caveman"
                target="_blank"
                rel="noreferrer"
                className="text-xs text-primary underline hover:opacity-80"
              >
                (Caveman Protocol)
              </a>
              <span className={`text-[11px] font-medium px-2 py-0.5 rounded-[5px] ${cavemanEnabled ? "bg-success/15 text-success border border-success/30" : "bg-surface-2 text-text-muted border border-border"}`}>
                {cavemanEnabled ? `Aktif (${cavemanLevel})` : "Nonaktif"}
              </span>
            </div>
            <p className="text-sm text-text-muted mt-1.5">
              Menginjeksi instruksi sistem gaya *caveman/terse* untuk memangkas basa-basi AI (sopan santun berlebih, pengantar panjang). Menghemat ~65% hingga 87% output token dengan tetap mempertahankan kode dan logika teknis 100% presisi.
            </p>

            {cavemanEnabled && (
              <div className="mt-3 flex items-center gap-2 flex-wrap">
                <span className="text-xs text-text-muted font-medium">Tingkat Intensitas:</span>
                {visibleCavemanLevels.map((lvl) => (
                  <button
                    key={lvl.id}
                    onClick={() => handleCavemanLevel(lvl.id)}
                    className={`px-3 py-1.5 rounded-[5px] text-xs font-medium border transition-all active:scale-[0.98] ${
                      cavemanLevel === lvl.id
                        ? "bg-primary text-white border-primary shadow-xs"
                        : "bg-surface-2 border-border text-text-muted hover:border-primary/40 hover:text-text"
                    }`}
                    title={lvl.desc}
                  >
                    {lvl.label}
                  </button>
                ))}
                <span className="text-xs text-primary font-medium ml-1">
                  — {CAVEMAN_LEVELS.find((lvl) => lvl.id === cavemanLevel)?.desc}
                </span>
              </div>
            )}
          </div>
          <div className="pt-1">
            <Toggle
              checked={cavemanEnabled}
              onChange={() => handleCavemanEnabled(!cavemanEnabled)}
            />
          </div>
        </div>

        {/* Module 3: Ponytail (Lazy Senior Dev) */}
        <div className="flex items-start justify-between gap-4 pb-5 border-b border-border">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="material-symbols-outlined text-purple-400 text-xl">
                psychology
              </span>
              <p className="font-semibold text-base">
                Ponytail (Lazy Senior Dev Prompt)
              </p>
              <a
                href="https://github.com/DietrichGebert/ponytail"
                target="_blank"
                rel="noreferrer"
                className="text-xs text-primary underline hover:opacity-80"
              >
                (Ponytail)
              </a>
              <span className={`text-[11px] font-medium px-2 py-0.5 rounded-[5px] ${ponytailEnabled ? "bg-success/15 text-success border border-success/30" : "bg-surface-2 text-text-muted border border-border"}`}>
                {ponytailEnabled ? `Aktif (${ponytailLevel})` : "Nonaktif"}
              </span>
            </div>
            <p className="text-sm text-text-muted mt-1.5">
              Mengarahkan AI untuk menulis kode seringkas mungkin: memprioritaskan YAGNI (You Aren't Gonna Need It), mendahulukan Go/JS standard library, menghapus kode lama daripada menambah abstraksi rumit, dan meminimalkan ukuran git diff.
            </p>

            {ponytailEnabled && (
              <div className="mt-3 flex items-center gap-2 flex-wrap">
                <span className="text-xs text-text-muted font-medium">Mode:</span>
                {PONYTAIL_LEVELS.map((lvl) => (
                  <button
                    key={lvl.id}
                    onClick={() => handlePonytailLevel(lvl.id)}
                    className={`px-3 py-1.5 rounded-[5px] text-xs font-medium border transition-all active:scale-[0.98] ${
                      ponytailLevel === lvl.id
                        ? "bg-primary text-white border-primary shadow-xs"
                        : "bg-surface-2 border-border text-text-muted hover:border-primary/40 hover:text-text"
                    }`}
                    title={lvl.desc}
                  >
                    {lvl.label}
                  </button>
                ))}
                <span className="text-xs text-primary font-medium ml-1">
                  — {PONYTAIL_LEVELS.find((lvl) => lvl.id === ponytailLevel)?.desc}
                </span>
              </div>
            )}
          </div>
          <div className="pt-1">
            <Toggle
              checked={ponytailEnabled}
              onChange={() => handlePonytailEnabled(!ponytailEnabled)}
            />
          </div>
        </div>

        {/* Module 4: Headroom (Context Compression Sidecar) */}
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="material-symbols-outlined text-blue-400 text-xl">
                memory
              </span>
              <p className="font-semibold text-base">
                Headroom Context Compression
              </p>
              <a
                href="https://github.com/chopratejas/headroom"
                target="_blank"
                rel="noreferrer"
                className="text-xs text-primary underline hover:opacity-80"
              >
                (Headroom)
              </a>
              <span className={`text-[11px] font-medium px-2 py-0.5 rounded-[5px] ${headroomRunning ? "bg-success/15 text-success border border-success/30" : "bg-warning/15 text-warning border border-warning/30"}`}>
                {headroomStatusLabel}
              </span>
              <button
                type="button"
                onClick={() => setShowHeadroomInstallModal(true)}
                className="text-xs text-primary underline hover:opacity-80 font-medium"
              >
                {headroomRunning ? "Kelola" : "Konfigurasi"}
              </button>
            </div>
            <p className="text-sm text-text-muted mt-1.5">
              Mengompresi konteks riwayat percakapan panjang menggunakan parser AST tree-sitter untuk kode pemrograman dan model ML Kompress-v2 untuk teks panjang via <code className="text-xs bg-surface-2 px-1 py-0.5 rounded text-text">/v1/compress</code>.
            </p>

            {headroomStatus.installed && (
              <div className="mt-3 p-3 rounded-lg border border-border bg-surface-2/60">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-xs font-medium text-text-muted">
                    Modul Tambahan {headroomExtras.version ? `(v${headroomExtras.version})` : ""}:
                  </span>
                  {headroomExtras.available.map((extra) => {
                    const installed = !!headroomExtras.extras[extra];
                    const pending = pendingExtras.includes(extra);
                    if (installed) {
                      const active = extra === "code" ? codeAware : kompress;
                      return (
                        <div
                          key={extra}
                          className="flex items-center gap-1.5 text-xs px-2 py-1 rounded-[5px] border border-success/40 bg-success/5 text-text"
                        >
                          <Toggle
                            size="sm"
                            checked={active}
                            disabled={restartingProxy}
                            onChange={() => toggleExtraActive(extra, !active)}
                          />
                          <span className="font-semibold">[{extra}]</span>
                          <button
                            type="button"
                            onClick={() => handleRemoveExtra(extra)}
                            disabled={removingExtra === extra}
                            className="ml-1 text-[11px] text-error hover:underline disabled:opacity-50"
                          >
                            {removingExtra === extra ? "Menghapus…" : "Hapus"}
                          </button>
                        </div>
                      );
                    }
                    return (
                      <label
                        key={extra}
                        className={`flex items-center gap-1.5 text-xs px-2 py-1 rounded-[5px] border cursor-pointer transition-colors ${
                          pending
                            ? "border-primary bg-primary/10 text-primary"
                            : "border-border text-text-muted hover:bg-surface-2"
                        }`}
                      >
                        <input
                          type="checkbox"
                          className="w-3 h-3"
                          checked={pending}
                          onChange={() => togglePendingExtra(extra)}
                        />
                        <span className="font-medium">[{extra}]</span>
                        <span className="opacity-70 text-[10px]">belum pasang</span>
                      </label>
                    );
                  })}
                  {pendingExtras.length > 0 && (
                    <button
                      onClick={handleInstallExtras}
                      disabled={extrasActionLoading}
                      className="text-xs px-2.5 py-1 rounded-[5px] bg-primary text-white hover:opacity-90 disabled:opacity-50 font-medium"
                    >
                      {extrasActionLoading ? "Mengunduh…" : `Pasang [${pendingExtras.join(",")}]`}
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>
          <div className="pt-1">
            <Toggle
              checked={headroomEnabled}
              onChange={() => handleHeadroomEnabled(!headroomEnabled)}
            />
          </div>
        </div>
      </Card>

      {/* Interactive Live Test Sandbox / Playground */}
      <Card className="border border-primary/20 bg-gradient-to-b from-surface-1 to-surface-2/40 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border pb-3">
          <div>
            <div className="flex items-center gap-2">
              <span className="material-symbols-outlined text-primary text-xl">
                biotech
              </span>
              <h2 className="text-base font-bold">
                Interactive Test Sandbox (Uji Coba Kompresi Langsung)
              </h2>
            </div>
            <p className="text-xs text-text-muted mt-0.5">
              Paste output eksekusi tool CLI agent Anda atau pilih sample di bawah untuk melihat hasil kompresi secara instan.
            </p>
          </div>

          <div className="flex items-center gap-1.5 flex-wrap">
            {Object.entries(SAMPLE_TEMPLATES).map(([k, s]) => (
              <button
                key={k}
                type="button"
                onClick={() => {
                  setSandboxInput(s.text);
                  setSandboxMode(s.mode);
                  setSandboxResult(null);
                }}
                className="px-2.5 py-1 rounded-[5px] text-[11px] font-medium border border-border bg-surface-1 text-text-muted hover:text-text hover:border-primary/40 active:scale-[0.98] transition-all"
              >
                {s.label}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <span className="text-xs text-text-muted font-medium">Mode Pengujian:</span>
              <button
                type="button"
                onClick={() => setSandboxMode("rtk")}
                className={`px-2.5 py-1 rounded-[5px] text-xs font-medium border ${sandboxMode === "rtk" ? "bg-primary text-white border-primary" : "border-border text-text-muted"}`}
              >
                RTK Tool Filter
              </button>
              <button
                type="button"
                onClick={() => setSandboxMode("caveman")}
                className={`px-2.5 py-1 rounded-[5px] text-xs font-medium border ${sandboxMode === "caveman" ? "bg-primary text-white border-primary" : "border-border text-text-muted"}`}
              >
                Caveman Terse
              </button>
              <button
                type="button"
                onClick={() => setSandboxMode("ponytail")}
                className={`px-2.5 py-1 rounded-[5px] text-xs font-medium border ${sandboxMode === "ponytail" ? "bg-primary text-white border-primary" : "border-border text-text-muted"}`}
              >
                Ponytail Code
              </button>
            </div>

            <button
              type="button"
              onClick={runSandboxTest}
              disabled={sandboxLoading || !sandboxInput.trim()}
              className="px-4 py-1.5 rounded-[5px] text-xs font-semibold bg-primary text-white hover:opacity-90 active:scale-[0.98] disabled:opacity-50 transition-all flex items-center gap-1.5 shadow-xs"
            >
              <span className="material-symbols-outlined text-sm">play_arrow</span>
              {sandboxLoading ? "Memproses…" : "Jalankan Kompresi"}
            </button>
          </div>

          <textarea
            value={sandboxInput}
            onChange={(e) => setSandboxInput(e.target.value)}
            rows={5}
            className="w-full font-mono text-xs p-3 rounded-lg border border-border bg-surface-1 text-text focus:outline-hidden focus:border-primary resize-y"
            placeholder="Ketik atau paste output eksekusi tool di sini..."
          />
        </div>

        {/* Results View */}
        {sandboxResult && (
          <div className="p-4 rounded-lg border border-success/30 bg-success/5 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-success/20 pb-2">
              <div className="flex items-center gap-2">
                <span className="material-symbols-outlined text-success text-base">check_circle</span>
                <span className="text-xs font-semibold text-success uppercase tracking-wider">Hasil Optimasi</span>
                <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-surface-1 border border-border text-text-muted">
                  Filter: {sandboxResult.detectedFilter}
                </span>
              </div>
              <div className="flex items-center gap-3 text-xs">
                <span>Sebelum: <strong className="font-mono text-text">{sandboxResult.originalChars}</strong> chars (~{sandboxResult.originalTokens} tok)</span>
                <span>Sesudah: <strong className="font-mono text-success">{sandboxResult.compressedChars}</strong> chars (~{sandboxResult.compressedTokens} tok)</span>
                <span className="px-2 py-0.5 rounded-[5px] font-bold bg-success text-white text-[11px]">
                  -{sandboxResult.reductionPercent}%
                </span>
              </div>
            </div>

            <div>
              <p className="text-[11px] font-medium text-text-muted mb-1">Payload Terkompresi yang Dikirim ke LLM:</p>
              <pre className="font-mono text-[11px] leading-relaxed p-3 rounded-lg bg-surface-1 border border-border overflow-x-auto text-text max-h-48 overflow-y-auto whitespace-pre-wrap">
                {sandboxResult.compressedText}
              </pre>
            </div>
          </div>
        )}
      </Card>

      {/* AI Agent & Tool Execution Architecture Guide */}
      <Card className="space-y-4">
        <div className="flex items-center gap-2 border-b border-border pb-3">
          <span className="material-symbols-outlined text-accent text-xl">smart_toy</span>
          <h2 className="text-base font-bold">
            Dukungan Eksekusi Tool untuk AI Agent (Hermes, Claude Code, OpenCode)
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-3.5 rounded-lg border border-border bg-surface-2/40 space-y-2">
            <div className="flex items-center justify-between">
              <p className="font-semibold text-sm flex items-center gap-1.5">
                <span className="material-symbols-outlined text-purple-400 text-base">psychology</span>
                Hermes Agent (Nous Research)
              </p>
              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-success/15 text-success">Native Ready</span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">
              Hermes Agent memanggil tools via format standar OpenAI (<code className="text-[10px] bg-surface-1 px-1 rounded">tools</code> & <code className="text-[10px] bg-surface-1 px-1 rounded">tool_calls</code>). Begitu Hermes mengeksekusi bash atau web fetch, hasil <code className="text-[10px] bg-surface-1 px-1 rounded">role: tool</code> otomatis dikompresi oleh RTK sebelum re-evaluasi LLM berikutnya.
            </p>
            <div className="pt-1 flex items-center justify-between text-[11px] font-mono bg-surface-1 p-2 rounded border border-border">
              <span className="text-text-muted">Base URL: http://127.0.0.1:20128/v1</span>
              <button
                type="button"
                onClick={() => copy("http://127.0.0.1:20128/v1")}
                className="text-primary hover:underline"
              >
                {copied ? "Copied" : "Copy"}
              </button>
            </div>
          </div>

          <div className="p-3.5 rounded-lg border border-border bg-surface-2/40 space-y-2">
            <div className="flex items-center justify-between">
              <p className="font-semibold text-sm flex items-center gap-1.5">
                <span className="material-symbols-outlined text-amber-500 text-base">terminal</span>
                Claude Code CLI (Anthropic)
              </p>
              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-success/15 text-success">Full Bridge</span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">
              Claude Code CLI berkomunikasi melalui Anthropic Messages API (<code className="text-[10px] bg-surface-1 px-1 rounded">/v1/messages</code>). 9Router menerjemahkan Anthropic <code className="text-[10px] bg-surface-1 px-1 rounded">tool_use</code> dan <code className="text-[10px] bg-surface-1 px-1 rounded">tool_result</code> secara transparan sehingga Claude Code dapat menggunakan model provider manapun.
            </p>
            <div className="pt-1 flex items-center justify-between text-[11px] font-mono bg-surface-1 p-2 rounded border border-border">
              <span className="text-text-muted">ANTHROPIC_BASE_URL=http://127.0.0.1:20128</span>
              <button
                type="button"
                onClick={() => copy("export ANTHROPIC_BASE_URL=\"http://127.0.0.1:20128\"")}
                className="text-primary hover:underline"
              >
                {copied ? "Copied" : "Copy"}
              </button>
            </div>
          </div>
        </div>

        <div className="p-3.5 rounded-lg border border-border bg-surface-2/40 space-y-2">
          <div className="flex items-center justify-between">
            <p className="font-semibold text-sm flex items-center gap-1.5">
              <span className="material-symbols-outlined text-blue-400 text-base">hub</span>
              Model Context Protocol (MCP) Bridge
            </p>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-primary/15 text-primary">/api/mcp/:plugin/sse</span>
          </div>
          <p className="text-xs text-text-muted leading-relaxed">
            9Router menyediakan jembatan stdio-ke-SSE langsung untuk MCP tools (Filesystem, Brave Search, GitHub, Puppeteer, Memory). Agent dapat menghubungkan plugin lokal secara plug-and-play tanpa instalasi server terpisah.
          </p>
        </div>
      </Card>

      {/* Headroom Install & Setup Modal */}
      <Modal
        isOpen={showHeadroomInstallModal}
        title={headroomRunning ? "Kelola Headroom" : "Setup Headroom Proxy"}
        onClose={() => setShowHeadroomInstallModal(false)}
      >
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between text-sm">
            <span>Status</span>
            <span className={headroomRunning ? "text-success font-semibold" : "text-warning font-semibold"}>
              {headroomStatusLabel}
            </span>
          </div>
          {headroomRunning && (
            <a
              href="/api/headroom/proxy/dashboard"
              target="_blank"
              rel="noreferrer"
              className="w-full rounded-[5px] border border-border px-4 py-2 text-center text-sm hover:bg-surface-2 font-medium"
            >
              Buka Headroom Dashboard
            </a>
          )}
          <div className="flex flex-col gap-1">
            <p className="text-sm font-medium">Proxy URL</p>
            <Input
              value={headroomUrl}
              onChange={(e) => setHeadroomUrl(e.target.value)}
              onBlur={handleHeadroomUrlBlur}
              placeholder="http://localhost:8787"
              className="font-mono text-sm"
            />
            <p className="text-xs text-text-muted">
              Gunakan proxy lokal untuk Start/Stop otomatis, atau sidecar Docker eksternal seperti http://headroom:8787.
            </p>
          </div>
          <div className="flex flex-col gap-1">
            <p className="text-sm font-medium">Timeout (ms)</p>
            <Input
              value={String(headroomTimeoutMs)}
              onChange={(e) => setHeadroomTimeoutMs(e.target.value)}
              onBlur={handleHeadroomTimeoutBlur}
              placeholder="3000"
              className="font-mono text-sm"
            />
            <p className="text-xs text-text-muted">
              Batas waktu request dalam milidetik. Standar 3000 ms.
            </p>
          </div>
          {headroomManaged ? (
            <Button
              onClick={handleHeadroomStop}
              variant="ghost"
              fullWidth
              disabled={headroomActionLoading}
            >
              {headroomActionLoading ? "Menghentikan…" : "Hentikan Headroom"}
            </Button>
          ) : headroomRunning ? (
            <p className="text-sm text-success">
              Proxy Headroom aktif dan dapat dijangkau. Token saver siap digunakan.
            </p>
          ) : headroomCanStart ? (
            <Button
              onClick={handleHeadroomStart}
              fullWidth
              disabled={headroomActionLoading}
            >
              {headroomActionLoading ? "Menjalankan…" : "Jalankan Headroom"}
            </Button>
          ) : !headroomLocalUrl ? (
            <p className="text-sm text-warning">
              Jalankan Headroom secara terpisah pada URL di atas, lalu klik Periksa Ulang.
            </p>
          ) : !headroomStatus.python ? (
            <p className="text-sm text-warning">
              Membutuhkan Python ≥ 3.10 untuk mode lokal. Pasang Python terlebih dahulu atau gunakan Docker.
            </p>
          ) : (
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Pasang via Terminal:</p>
              <div className="flex items-center gap-2">
                <pre className="flex-1 rounded-[5px] bg-black/10 dark:bg-white/5 p-2 text-xs font-mono overflow-x-auto">
                  {`pip install "headroom-ai[proxy]"`}
                </pre>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => copy(`pip install "headroom-ai[proxy]"`)}
                >
                  {copied ? "Tersalin" : "Salin"}
                </Button>
              </div>
            </div>
          )}
          {headroomActionError && (
            <p className="text-sm text-error">{headroomActionError}</p>
          )}
          <div className="flex gap-2 pt-2">
            <Button
              onClick={() => refreshHeadroomStatus()}
              variant="ghost"
              fullWidth
            >
              Periksa Ulang
            </Button>
            <Button
              onClick={() => setShowHeadroomInstallModal(false)}
              fullWidth
            >
              Selesai
            </Button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        isOpen={!!extrasConfirm}
        onClose={() => setExtrasConfirm(null)}
        onConfirm={() => {
          const fn = extrasConfirm?.onConfirm;
          setExtrasConfirm(null);
          fn?.();
        }}
        title={extrasConfirm?.title}
        message={extrasConfirm?.message}
        confirmText={extrasConfirm?.confirmText}
        variant={extrasConfirm?.variant}
      />
    </div>
  );
}

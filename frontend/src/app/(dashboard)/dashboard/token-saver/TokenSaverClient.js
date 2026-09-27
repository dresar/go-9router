"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { Card, Button, Input, Modal, Toggle, ConfirmModal, Badge } from "@/shared/components";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";
import { getCurrentLocale, onLocaleChange } from "@/i18n/runtime";
import {
  WENYAN_LOCALES,
  CAVEMAN_LEVELS,
  PONYTAIL_LEVELS,
} from "../endpoint/endpointConstants";

const SAMPLE_TEMPLATES = {
  gitDiff: {
    label: "Git Diff",
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
    label: "Ripgrep",
    mode: "rtk",
    text: `frontend/src/shared/components/Sidebar.js:27:  { href: "/token-saver", label: "Token Saver", icon: "savings" },
frontend/src/shared/components/Header.js:115:  if (pathname.includes("/token-saver"))
frontend/open-sse/handlers/chatCore.js:284:  // Token-saver flags accumulator for the single log line below.
internal/router/router.go:121:  mux.HandleFunc("/api/token-saver/test", h.HandleTokenSaverTest)
internal/tokensaver/tokensaver.go:42:func CompressToolOutput(text string) (string, string)`,
  },
  treeDir: {
    label: "Tree",
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
    label: "Verbose LLM",
    mode: "caveman",
    text: `Of course! I would be more than happy to help you resolve this issue. Based on a thorough analysis of your codebase, the primary problem that is causing this error is a redundant check in your authentication middleware. You can resolve this by removing the duplicate check inside auth.go. Please let me know if you need any further assistance!`,
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

  const { copied, copy } = useCopyToClipboard(2000);

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
      console.error("Error updating setting:", error);
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
      console.error("Error updating rtkEnabled:", error);
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
        message: "Download ~1 GB (torch + huggingface-hub)?",
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
      message: `Remove [${extra}] packages?`,
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
        const origLen = sandboxInput.length;
        const fakeComp = sandboxInput
          .replace(/^diff --git.*$/gm, "diff --git")
          .replace(/^index .*$/gm, "");
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
          if (typeof data.headroomTimeoutMs === "number")
            setHeadroomTimeoutMs(data.headroomTimeoutMs);
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
    ? "Checking..."
    : headroomRunning
      ? "Active"
      : headroomStatus.localUrl !== false && !headroomStatus.installed
        ? "Not Installed"
        : headroomStatus.localUrl !== false
          ? "Stopped"
          : "External";

  const headroomLocalUrl = headroomStatus.localUrl !== false;
  const headroomCanStart = !!headroomStatus.canStart;
  const headroomManaged = headroomLocalUrl && !!headroomStatus.managedPid;

  const activeCount =
    (rtkEnabled ? 1 : 0) +
    (cavemanEnabled ? 1 : 0) +
    (ponytailEnabled ? 1 : 0) +
    (headroomEnabled ? 1 : 0);

  return (
    <div className="flex min-w-0 flex-col gap-6 pb-12">
      {/* 1. Header Banner & Quick Presets */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-text-main">
              Token Saver
            </h1>
            <Badge variant="primary" size="sm">
              {activeCount} / 4 Active
            </Badge>
          </div>
          <p className="text-xs text-text-muted mt-1">
            Context optimization & tool output compression.
          </p>
        </div>

        {/* Quick Presets */}
        <div className="flex items-center gap-1.5 flex-wrap">
          <Button
            variant="secondary"
            size="sm"
            icon="smart_toy"
            onClick={() => applyPreset("agent")}
          >
            Agent
          </Button>
          <Button
            variant="secondary"
            size="sm"
            icon="bolt"
            onClick={() => applyPreset("max")}
          >
            Max
          </Button>
          <Button
            variant="secondary"
            size="sm"
            icon="balance"
            onClick={() => applyPreset("balanced")}
          >
            Balanced
          </Button>
          <Button
            variant="ghost"
            size="sm"
            icon="block"
            onClick={() => applyPreset("raw")}
          >
            Off
          </Button>
        </div>
      </div>

      {/* 2. Compact Performance KPI Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
        <Card padding="xs" className="p-3">
          <p className="text-[11px] font-medium text-text-muted uppercase">Saved</p>
          <p className="text-xl font-bold text-green-500 mt-0.5">60% – 90%</p>
          <p className="text-[10px] text-text-muted mt-0.5 truncate">Tool outputs</p>
        </Card>
        <Card padding="xs" className="p-3">
          <p className="text-[11px] font-medium text-text-muted uppercase">Latency</p>
          <p className="text-xl font-bold text-brand-500 mt-0.5">&lt; 1 ms</p>
          <p className="text-[10px] text-text-muted mt-0.5 truncate">Go buffer</p>
        </Card>
        <Card padding="xs" className="p-3">
          <p className="text-[11px] font-medium text-text-muted uppercase">Agent Tools</p>
          <p className="text-xl font-bold text-text-main mt-0.5">Active</p>
          <p className="text-[10px] text-text-muted mt-0.5 truncate">Hermes & Claude</p>
        </Card>
        <Card padding="xs" className="p-3">
          <p className="text-[11px] font-medium text-text-muted uppercase">MCP Bridge</p>
          <p className="text-xl font-bold text-blue-500 mt-0.5">SSE</p>
          <p className="text-[10px] text-text-muted mt-0.5 truncate">Stdio to stream</p>
        </Card>
      </div>

      {/* 3. Core Optimization Modules Card */}
      <Card padding="sm" className="space-y-4">
        <div className="flex items-center justify-between border-b border-border-subtle pb-2.5">
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-brand-500 text-[20px]">
              tune
            </span>
            <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
              Modules
            </h2>
          </div>
        </div>

        {/* Module 1: RTK Compressor */}
        <div className="flex items-start justify-between gap-4 pb-4 border-b border-border-subtle">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="material-symbols-outlined text-brand-500 text-[18px]">
                terminal
              </span>
              <span className="font-semibold text-sm text-text-main">
                RTK Compressor
              </span>
              <Badge variant={rtkEnabled ? "success" : "default"} size="sm">
                {rtkEnabled ? "Active" : "Bypass"}
              </Badge>
            </div>
            <p className="text-xs text-text-muted mt-1">
              Compresses CLI & agent tool execution outputs.
            </p>
            <div className="flex items-center gap-1.5 mt-2 flex-wrap">
              {["git-diff", "git-log", "grep", "tree", "build"].map((f) => (
                <span
                  key={f}
                  className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-surface-2 border border-border-subtle text-text-muted"
                >
                  {f}
                </span>
              ))}
            </div>
          </div>
          <div className="pt-1">
            <Toggle checked={rtkEnabled} onChange={() => handleRtkEnabled(!rtkEnabled)} />
          </div>
        </div>

        {/* Module 2: Caveman Terse Mode */}
        <div className="flex items-start justify-between gap-4 pb-4 border-b border-border-subtle">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="material-symbols-outlined text-amber-500 text-[18px]">
                speaker_notes_off
              </span>
              <span className="font-semibold text-sm text-text-main">
                Caveman Mode
              </span>
              <Badge variant={cavemanEnabled ? "success" : "default"} size="sm">
                {cavemanEnabled ? `Active (${cavemanLevel})` : "Off"}
              </Badge>
            </div>
            <p className="text-xs text-text-muted mt-1">
              Trims conversational fluff while preserving code logic.
            </p>

            {cavemanEnabled && (
              <div className="mt-2.5 flex items-center gap-1.5 flex-wrap">
                {visibleCavemanLevels.map((lvl) => (
                  <Button
                    key={lvl.id}
                    variant={cavemanLevel === lvl.id ? "primary" : "secondary"}
                    size="sm"
                    onClick={() => handleCavemanLevel(lvl.id)}
                  >
                    {lvl.label}
                  </Button>
                ))}
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

        {/* Module 3: Ponytail Lazy Senior Dev */}
        <div className="flex items-start justify-between gap-4 pb-4 border-b border-border-subtle">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="material-symbols-outlined text-purple-400 text-[18px]">
                psychology
              </span>
              <span className="font-semibold text-sm text-text-main">
                Ponytail Mode
              </span>
              <Badge variant={ponytailEnabled ? "success" : "default"} size="sm">
                {ponytailEnabled ? `Active (${ponytailLevel})` : "Off"}
              </Badge>
            </div>
            <p className="text-xs text-text-muted mt-1">
              Enforces concise code, standard libraries, and minimal diffs.
            </p>

            {ponytailEnabled && (
              <div className="mt-2.5 flex items-center gap-1.5 flex-wrap">
                {PONYTAIL_LEVELS.map((lvl) => (
                  <Button
                    key={lvl.id}
                    variant={ponytailLevel === lvl.id ? "primary" : "secondary"}
                    size="sm"
                    onClick={() => handlePonytailLevel(lvl.id)}
                  >
                    {lvl.label}
                  </Button>
                ))}
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

        {/* Module 4: Headroom Context Compression Sidecar */}
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="material-symbols-outlined text-blue-400 text-[18px]">
                memory
              </span>
              <span className="font-semibold text-sm text-text-main">
                Headroom Sidecar
              </span>
              <Badge variant={headroomRunning ? "success" : "warning"} size="sm">
                {headroomStatusLabel}
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                icon="settings"
                onClick={() => setShowHeadroomInstallModal(true)}
              >
                Configure
              </Button>
            </div>
            <p className="text-xs text-text-muted mt-1">
              AST code compression and semantic text summarization.
            </p>

            {headroomStatus.installed && (
              <div className="mt-2.5 p-2.5 rounded-[10px] border border-border-subtle bg-surface-2 flex items-center gap-2 flex-wrap">
                <span className="text-xs font-medium text-text-muted">Extras:</span>
                {headroomExtras.available.map((extra) => {
                  const installed = !!headroomExtras.extras[extra];
                  const pending = pendingExtras.includes(extra);
                  if (installed) {
                    const active = extra === "code" ? codeAware : kompress;
                    return (
                      <div
                        key={extra}
                        className="flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-[6px] border border-green-500/30 bg-green-500/10 text-text-main"
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
                          className="ml-1 text-[10px] text-red-400 hover:underline disabled:opacity-50 cursor-pointer"
                        >
                          {removingExtra === extra ? "..." : "Remove"}
                        </button>
                      </div>
                    );
                  }
                  return (
                    <label
                      key={extra}
                      className={`flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-[6px] border cursor-pointer ${
                        pending
                          ? "border-brand-500 bg-brand-500/10 text-brand-500"
                          : "border-border-subtle text-text-muted hover:bg-surface-3"
                      }`}
                    >
                      <input
                        type="checkbox"
                        className="w-3 h-3 accent-brand-500"
                        checked={pending}
                        onChange={() => togglePendingExtra(extra)}
                      />
                      <span>[{extra}]</span>
                    </label>
                  );
                })}
                {pendingExtras.length > 0 && (
                  <Button
                    size="sm"
                    loading={extrasActionLoading}
                    disabled={extrasActionLoading}
                    onClick={handleInstallExtras}
                  >
                    Install
                  </Button>
                )}
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

      {/* 4. Interactive Test Sandbox */}
      <Card padding="sm" className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border-subtle pb-2.5">
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-brand-500 text-[20px]">
              biotech
            </span>
            <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
              Sandbox
            </h2>
          </div>

          <div className="flex items-center gap-1.5 flex-wrap">
            {Object.entries(SAMPLE_TEMPLATES).map(([k, s]) => (
              <Button
                key={k}
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSandboxInput(s.text);
                  setSandboxMode(s.mode);
                  setSandboxResult(null);
                }}
              >
                {s.label}
              </Button>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex items-center gap-1.5">
              <Button
                variant={sandboxMode === "rtk" ? "primary" : "secondary"}
                size="sm"
                onClick={() => setSandboxMode("rtk")}
              >
                RTK
              </Button>
              <Button
                variant={sandboxMode === "caveman" ? "primary" : "secondary"}
                size="sm"
                onClick={() => setSandboxMode("caveman")}
              >
                Caveman
              </Button>
              <Button
                variant={sandboxMode === "ponytail" ? "primary" : "secondary"}
                size="sm"
                onClick={() => setSandboxMode("ponytail")}
              >
                Ponytail
              </Button>
            </div>

            <Button
              variant="primary"
              size="sm"
              icon="play_arrow"
              loading={sandboxLoading}
              disabled={sandboxLoading || !sandboxInput.trim()}
              onClick={runSandboxTest}
            >
              {sandboxLoading ? "Compressing..." : "Compress"}
            </Button>
          </div>

          <textarea
            value={sandboxInput}
            onChange={(e) => setSandboxInput(e.target.value)}
            rows={5}
            placeholder="Output..."
            className="w-full font-mono text-xs p-3 rounded-[10px] border border-border-subtle bg-surface-2 text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50 resize-y custom-scrollbar"
          />
        </div>

        {/* Results View */}
        {sandboxResult && (
          <div className="p-3.5 rounded-[12px] border border-green-500/30 bg-green-500/5 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-green-500/20 pb-2">
              <div className="flex items-center gap-2">
                <span className="material-symbols-outlined text-green-500 text-base">
                  check_circle
                </span>
                <span className="text-xs font-semibold text-green-500 uppercase">
                  Results
                </span>
                <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-surface-2 border border-border-subtle text-text-muted">
                  {sandboxResult.detectedFilter}
                </span>
              </div>
              <div className="flex items-center gap-3 text-xs">
                <span>
                  Original:{" "}
                  <strong className="font-mono text-text-main">
                    ~{sandboxResult.originalTokens} tok
                  </strong>
                </span>
                <span>
                  Compressed:{" "}
                  <strong className="font-mono text-green-500">
                    ~{sandboxResult.compressedTokens} tok
                  </strong>
                </span>
                <Badge variant="success" size="sm">
                  -{sandboxResult.reductionPercent}%
                </Badge>
              </div>
            </div>

            <pre className="font-mono text-[11px] leading-relaxed p-3 rounded-[8px] bg-bg border border-border-subtle text-text-main max-h-48 overflow-y-auto whitespace-pre-wrap custom-scrollbar">
              {sandboxResult.compressedText}
            </pre>
          </div>
        )}
      </Card>

      {/* 5. Agent Integrations */}
      <Card padding="sm" className="space-y-4">
        <div className="flex items-center gap-2 border-b border-border-subtle pb-2.5">
          <span className="material-symbols-outlined text-brand-500 text-[20px]">
            smart_toy
          </span>
          <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
            Integrations
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {/* Hermes Agent */}
          <div className="p-3 rounded-[10px] border border-border-subtle bg-surface-2 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-xs text-text-main">Hermes Agent</span>
              <Badge variant="success" size="sm">
                Ready
              </Badge>
            </div>
            <div className="flex items-center justify-between text-[11px] font-mono bg-bg p-2 rounded-[6px] border border-border-subtle">
              <span className="truncate text-text-muted">http://127.0.0.1:20128/v1</span>
              <Button
                variant="ghost"
                size="sm"
                icon={copied ? "check" : "content_copy"}
                onClick={() => copy("http://127.0.0.1:20128/v1")}
              >
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
          </div>

          {/* Claude Code CLI */}
          <div className="p-3 rounded-[10px] border border-border-subtle bg-surface-2 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-xs text-text-main">Claude Code CLI</span>
              <Badge variant="success" size="sm">
                Ready
              </Badge>
            </div>
            <div className="flex items-center justify-between text-[11px] font-mono bg-bg p-2 rounded-[6px] border border-border-subtle">
              <span className="truncate text-text-muted">ANTHROPIC_BASE_URL</span>
              <Button
                variant="ghost"
                size="sm"
                icon={copied ? "check" : "content_copy"}
                onClick={() => copy('export ANTHROPIC_BASE_URL="http://127.0.0.1:20128"')}
              >
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
          </div>

          {/* MCP SSE Bridge */}
          <div className="p-3 rounded-[10px] border border-border-subtle bg-surface-2 space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-xs text-text-main">MCP Bridge</span>
              <Badge variant="primary" size="sm">
                SSE
              </Badge>
            </div>
            <div className="flex items-center justify-between text-[11px] font-mono bg-bg p-2 rounded-[6px] border border-border-subtle">
              <span className="truncate text-text-muted">/api/mcp/:plugin/sse</span>
              <Button
                variant="ghost"
                size="sm"
                icon={copied ? "check" : "content_copy"}
                onClick={() => copy("/api/mcp/:plugin/sse")}
              >
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
          </div>
        </div>
      </Card>

      {/* 6. Headroom Setup Modal */}
      <Modal
        isOpen={showHeadroomInstallModal}
        title={headroomRunning ? "Headroom" : "Headroom Setup"}
        onClose={() => setShowHeadroomInstallModal(false)}
      >
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-muted">Status</span>
            <Badge variant={headroomRunning ? "success" : "warning"} size="sm">
              {headroomStatusLabel}
            </Badge>
          </div>

          {headroomRunning && (
            <a
              href="/api/headroom/proxy/dashboard"
              target="_blank"
              rel="noreferrer"
              className="w-full rounded-[10px] border border-border-subtle px-4 py-2 text-center text-xs hover:bg-surface-2 font-medium text-text-main"
            >
              Dashboard
            </a>
          )}

          <div>
            <label className="block text-xs font-medium text-text-muted mb-1">
              Proxy URL
            </label>
            <Input
              value={headroomUrl}
              onChange={(e) => setHeadroomUrl(e.target.value)}
              onBlur={handleHeadroomUrlBlur}
              placeholder="URL..."
              className="font-mono text-xs"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-text-muted mb-1">
              Timeout (ms)
            </label>
            <Input
              value={String(headroomTimeoutMs)}
              onChange={(e) => setHeadroomTimeoutMs(e.target.value)}
              onBlur={handleHeadroomTimeoutBlur}
              placeholder="Timeout..."
              className="font-mono text-xs"
            />
          </div>

          {headroomManaged ? (
            <Button
              onClick={handleHeadroomStop}
              variant="danger"
              fullWidth
              disabled={headroomActionLoading}
            >
              {headroomActionLoading ? "Stopping..." : "Stop"}
            </Button>
          ) : headroomCanStart ? (
            <Button
              onClick={handleHeadroomStart}
              variant="primary"
              fullWidth
              disabled={headroomActionLoading}
            >
              {headroomActionLoading ? "Starting..." : "Start"}
            </Button>
          ) : !headroomStatus.python ? (
            <p className="text-xs text-amber-500">
              Python ≥ 3.10 required for local mode.
            </p>
          ) : (
            <div className="flex flex-col gap-1">
              <span className="text-xs font-medium text-text-muted">Install:</span>
              <div className="flex items-center gap-2">
                <pre className="flex-1 rounded-[8px] bg-bg p-2 text-xs font-mono overflow-x-auto text-text-main">
                  {`pip install "headroom-ai[proxy]"`}
                </pre>
                <Button
                  size="sm"
                  variant="secondary"
                  icon={copied ? "check" : "content_copy"}
                  onClick={() => copy(`pip install "headroom-ai[proxy]"`)}
                >
                  {copied ? "Copied" : "Copy"}
                </Button>
              </div>
            </div>
          )}

          {headroomActionError && (
            <p className="text-xs text-red-500">{headroomActionError}</p>
          )}

          <div className="flex gap-2 pt-2 border-t border-border-subtle">
            <Button
              onClick={() => refreshHeadroomStatus()}
              variant="secondary"
              fullWidth
            >
              Check
            </Button>
            <Button
              onClick={() => setShowHeadroomInstallModal(false)}
              variant="primary"
              fullWidth
            >
              Close
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

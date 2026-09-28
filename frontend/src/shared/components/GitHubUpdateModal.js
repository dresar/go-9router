"use client";

import { useState } from "react";
import Modal from "@/shared/components/Modal";
import Button from "@/shared/components/Button";
import Badge from "@/shared/components/Badge";
import { useGitHubUpdateStore } from "@/store/useGitHubUpdateStore";

export default function GitHubUpdateModal() {
  const {
    modalOpen,
    setModalOpen,
    updateInfo,
    loading,
    syncing,
    syncOutput,
    error,
    checkNow,
    syncNow
  } = useGitHubUpdateStore();

  const [copied, setCopied] = useState(false);

  if (!modalOpen) return null;

  const hasUpdate = Boolean(updateInfo?.hasUpdate);
  const currentSha = updateInfo?.currentCommit || "unknown";
  const latestSha = updateInfo?.latestCommit || currentSha;
  const commitMsg = updateInfo?.latestCommitMsg || "Tidak ada rincian pesan commit.";
  const commitDate = updateInfo?.latestCommitDate
    ? new Date(updateInfo.latestCommitDate).toLocaleString()
    : "-";
  const lastChecked = updateInfo?.lastChecked
    ? new Date(updateInfo.lastChecked).toLocaleTimeString()
    : "-";
  const repoUrl = updateInfo?.repoUrl || "https://github.com/dresar/go-9router";

  const handleCopySha = (sha) => {
    navigator.clipboard?.writeText(sha);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Modal
      open={modalOpen}
      onClose={() => setModalOpen(false)}
      title="Status Update & Sinkronisasi GitHub"
      size="md"
    >
      <div className="flex flex-col gap-4 text-text-main">
        {/* Top Status Banner */}
        <div
          className={`flex items-center gap-3 p-3 rounded-lg border ${
            hasUpdate
              ? "bg-amber-500/10 border-amber-500/30 text-amber-500"
              : "bg-emerald-500/10 border-emerald-500/30 text-emerald-500"
          }`}
        >
          <div
            className={`flex items-center justify-center size-8 rounded-md shrink-0 ${
              hasUpdate ? "bg-amber-500/20" : "bg-emerald-500/20"
            }`}
          >
            <span className="material-symbols-outlined text-[20px]">
              {hasUpdate ? "notifications_active" : "verified"}
            </span>
          </div>
          <div className="flex flex-col min-w-0 flex-1">
            <span className="text-xs font-semibold">
              {hasUpdate
                ? "Update Tersedia di GitHub!"
                : "Sistem Sudah Terkini (Up to Date)"}
            </span>
            <span className="text-[11px] text-text-muted truncate">
              {hasUpdate
                ? "Commit baru ditemukan di repositori GitHub remote."
                : "Versi lokal cocok dengan branch master di GitHub."}
            </span>
          </div>
          <Badge
            variant={hasUpdate ? "warning" : "success"}
            size="sm"
            dot
          >
            {hasUpdate ? "Ada Update" : "Terkini"}
          </Badge>
        </div>

        {/* Commit Details Card */}
        <div className="flex flex-col gap-2.5 p-3 rounded-lg border border-border-subtle bg-surface/50 text-xs">
          <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
            <span className="text-text-muted">Repositori GitHub</span>
            <a
              href={repoUrl}
              target="_blank"
              rel="noreferrer"
              className="text-primary hover:underline font-mono inline-flex items-center gap-1"
            >
              <span>dresar/go-9router</span>
              <span className="material-symbols-outlined text-[13px]">open_in_new</span>
            </a>
          </div>

          <div className="grid grid-cols-2 gap-2 py-1">
            <div className="flex flex-col gap-1 p-2 rounded bg-surface/80 border border-border-subtle">
              <span className="text-[11px] text-text-muted">Commit Lokal Saat Ini</span>
              <div className="flex items-center justify-between">
                <span className="font-mono font-medium text-text-main">{currentSha}</span>
                <button
                  type="button"
                  onClick={() => handleCopySha(currentSha)}
                  className="text-text-muted hover:text-text-main p-0.5 rounded"
                  title="Salin SHA"
                >
                  <span className="material-symbols-outlined text-[14px]">
                    {copied ? "check" : "content_copy"}
                  </span>
                </button>
              </div>
            </div>

            <div className="flex flex-col gap-1 p-2 rounded bg-surface/80 border border-border-subtle">
              <span className="text-[11px] text-text-muted">Commit Terkini di GitHub</span>
              <div className="flex items-center justify-between">
                <span className="font-mono font-medium text-amber-500">{latestSha}</span>
                <a
                  href={`${repoUrl}/commit/${latestSha}`}
                  target="_blank"
                  rel="noreferrer"
                  className="text-text-muted hover:text-primary p-0.5 rounded"
                  title="Lihat Commit di GitHub"
                >
                  <span className="material-symbols-outlined text-[14px]">open_in_new</span>
                </a>
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-1 pt-1">
            <span className="text-[11px] text-text-muted">Pesan Commit Terbaru</span>
            <div className="p-2 rounded bg-surface/60 border border-border-subtle font-mono text-[11px] leading-relaxed text-text-main break-words">
              {commitMsg}
            </div>
          </div>

          <div className="flex items-center justify-between pt-1 text-[11px] text-text-muted">
            <span>Waktu Commit: {commitDate}</span>
            <span>Diperiksa: {lastChecked}</span>
          </div>
        </div>

        {/* Sync Output or Error message */}
        {syncOutput && (
          <div className="flex flex-col gap-1 p-3 rounded-lg border border-emerald-500/30 bg-emerald-500/5 text-xs">
            <div className="flex items-center gap-1.5 font-semibold text-emerald-500">
              <span className="material-symbols-outlined text-[16px]">check_circle</span>
              <span>Hasil Sinkronisasi Git</span>
            </div>
            <pre className="font-mono text-[11px] whitespace-pre-wrap text-text-muted pt-1">
              {syncOutput}
            </pre>
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 p-3 rounded-lg border border-red-500/30 bg-red-500/10 text-red-500 text-xs">
            <span className="material-symbols-outlined text-[16px] shrink-0">error</span>
            <span>{error}</span>
          </div>
        )}

        {/* Periodic Schedule Note */}
        <div className="flex items-center gap-2 p-2.5 rounded-lg border border-border-subtle bg-surface/30 text-[11px] text-text-muted">
          <span className="material-symbols-outlined text-[15px] text-primary shrink-0">
            schedule
          </span>
          <span>
            Sinkronisasi otomatis berjalan di latar belakang setiap <strong>1 jam sekali</strong> (3600 detik).
          </span>
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-2 pt-2 border-t border-border-subtle">
          <Button
            variant="secondary"
            size="sm"
            onClick={checkNow}
            disabled={loading || syncing}
          >
            <span className={`material-symbols-outlined text-[15px] ${loading ? "animate-spin" : ""}`}>
              sync
            </span>
            <span>{loading ? "Memeriksa..." : "Cek Update Lagi"}</span>
          </Button>

          {hasUpdate && (
            <Button
              variant="primary"
              size="sm"
              onClick={syncNow}
              disabled={syncing || loading}
            >
              <span className={`material-symbols-outlined text-[15px] ${syncing ? "animate-spin" : ""}`}>
                download
              </span>
              <span>{syncing ? "Melakukan Sync..." : "Sync & Pull Sekarang"}</span>
            </Button>
          )}

          <Button
            variant="ghost"
            size="sm"
            onClick={() => setModalOpen(false)}
          >
            Tutup
          </Button>
        </div>
      </div>
    </Modal>
  );
}

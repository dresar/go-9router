"use client";

import { Card, Badge } from "@/shared/components";
import { useCopyToClipboard } from "@/shared/hooks/useCopyToClipboard";
import {
  SKILLS,
  SKILLS_REPO_URL,
  getSkillRawUrl,
  getSkillBlobUrl,
} from "@/shared/constants/skills";

function CopyButton({ value, label = "Copy link" }) {
  const { copied, copy } = useCopyToClipboard(2000);
  return (
    <button
      onClick={() => copy(value)}
      className="px-2.5 py-1.5 rounded-[5px] bg-primary text-white text-xs font-medium hover:opacity-90 active:scale-[0.98] transition-all cursor-pointer shrink-0 inline-flex items-center gap-1.5 shadow-xs"
      title={value}
    >
      <span className="material-symbols-outlined text-[14px]">
        {copied ? "check" : "content_copy"}
      </span>
      {copied ? "Tersalin!" : label}
    </button>
  );
}

function SkillRow({ skill }) {
  const rawUrl = getSkillRawUrl(skill.id);
  const blobUrl = getSkillBlobUrl(skill.id);

  return (
    <div
      className={`flex items-start gap-3.5 p-4 rounded-lg border transition-all ${
        skill.isEntry
          ? "border-primary/50 bg-primary/5 shadow-xs"
          : "border-border/80 bg-surface-1 hover:border-primary/40 hover:bg-surface-2"
      }`}
    >
      <div
        className={`size-9 rounded-[6px] flex items-center justify-center shrink-0 ${
          skill.isEntry ? "bg-primary text-white" : "bg-primary/10 text-primary"
        }`}
      >
        <span className="material-symbols-outlined text-[18px]">{skill.icon}</span>
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <h3 className="font-semibold text-sm text-text">{skill.name}</h3>
          {skill.isEntry && (
            <Badge variant="primary" size="sm">START HERE</Badge>
          )}
          {skill.endpoint && (
            <span className="px-1.5 py-0.5 rounded-[4px] text-[10px] font-mono bg-surface-2 border border-border text-text-muted">
              {skill.endpoint}
            </span>
          )}
        </div>
        <p className="text-xs text-text-muted mt-1 leading-relaxed">{skill.description}</p>
        <div className="mt-2 flex items-center gap-2 flex-wrap">
          <a
            href={blobUrl}
            target="_blank"
            rel="noreferrer"
            className="text-[11px] font-mono text-primary/80 hover:text-primary hover:underline inline-flex items-center gap-1 break-all"
          >
            {rawUrl}
            <span className="material-symbols-outlined text-[12px]">open_in_new</span>
          </a>
        </div>
      </div>

      <CopyButton value={rawUrl} />
    </div>
  );
}

export default function SkillsPage() {
  const entrySkillUrl = getSkillRawUrl("9router");
  const { copied, copy } = useCopyToClipboard(2000);

  return (
    <div className="max-w-4xl mx-auto space-y-6 p-4 sm:p-6">
      {/* Header & GitHub Repo Info */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 border-b border-border/70 pb-4">
        <div>
          <div className="flex items-center gap-2.5">
            <span className="material-symbols-outlined text-primary text-2xl">
              extension
            </span>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight">
              Agent Skills Hub
            </h1>
          </div>
          <p className="text-sm text-text-muted mt-1">
            Drop-in skills resmi untuk AI Agent (Claude Code CLI, Cursor, Hermes Agent, OpenCode, ChatGPT).
          </p>
        </div>

        <a
          href={SKILLS_REPO_URL}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[5px] text-xs font-medium border border-border bg-surface-2 hover:border-primary/40 text-text active:scale-[0.98] transition-all"
        >
          <span className="material-symbols-outlined text-sm text-primary">code</span>
          <span>github.com/dresar/go-9router</span>
          <span className="material-symbols-outlined text-[13px] text-text-muted">open_in_new</span>
        </a>
      </div>

      {/* Quick Prompt Copy Box */}
      <Card padding="md" className="border-primary/20 bg-primary/5">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="space-y-1">
            <div className="text-xs font-semibold text-primary uppercase tracking-wider">
              Prompt Siap Tempel ke AI Agent Anda:
            </div>
            <div className="font-mono text-xs text-text break-all">
              Read this skill and use it: {entrySkillUrl}
            </div>
          </div>
          <button
            onClick={() => copy(`Read this skill and use it: ${entrySkillUrl}`)}
            className="px-3 py-1.5 rounded-[5px] bg-primary text-white text-xs font-semibold hover:opacity-90 active:scale-[0.98] transition-all shrink-0 inline-flex items-center gap-1.5 shadow-xs"
          >
            <span className="material-symbols-outlined text-sm">
              {copied ? "check" : "content_copy"}
            </span>
            {copied ? "Tersalin!" : "Salin Prompt"}
          </button>
        </div>
      </Card>

      {/* Skills List */}
      <div className="space-y-2.5">
        {SKILLS.map((skill) => (
          <SkillRow key={skill.id} skill={skill} />
        ))}
      </div>

      {/* GitHub Repository Footer Card */}
      <Card padding="md">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <div>
            <h2 className="text-sm font-semibold text-text">Repository GitHub Resmi</h2>
            <p className="text-xs text-text-muted mt-0.5">
              Source code, dokumentasi markdown, dan seluruh agent skills tersimpan di <code className="text-[11px] font-mono text-primary">dresar/go-9router</code>.
            </p>
          </div>
          <a
            href={`${SKILLS_REPO_URL}/tree/master/skills`}
            target="_blank"
            rel="noreferrer"
            className="px-3.5 py-1.5 rounded-[5px] text-xs font-semibold border border-primary/30 bg-primary/10 text-primary hover:bg-primary/20 active:scale-[0.98] transition-all inline-flex items-center gap-1.5"
          >
            <span className="material-symbols-outlined text-[15px]">open_in_new</span>
            Buka Skills di GitHub (dresar/go-9router)
          </a>
        </div>
      </Card>
    </div>
  );
}

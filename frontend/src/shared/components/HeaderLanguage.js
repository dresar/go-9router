"use client";

import { useState, useEffect } from "react";
import { LOCALE_COOKIE, normalizeLocale } from "@/i18n/config";
import { reloadTranslations } from "@/i18n/runtime";

function getLocaleFromCookie() {
  if (typeof document === "undefined") return "id";
  const cookie = document.cookie
    .split(";")
    .find((c) => c.trim().startsWith(`${LOCALE_COOKIE}=`));
  const value = cookie ? decodeURIComponent(cookie.split("=")[1]) : "id";
  return normalizeLocale(value);
}

function FlagID() {
  return (
    <svg width="16" height="12" viewBox="0 0 16 12" fill="none" xmlns="http://www.w3.org/2000/svg" className="rounded-[2px] overflow-hidden shrink-0 shadow-[0_0_0_1px_rgba(0,0,0,0.18)]">
      <rect width="16" height="6" fill="#EF4444" />
      <rect y="6" width="16" height="6" fill="#FFFFFF" />
    </svg>
  );
}

function FlagEN() {
  return (
    <svg width="16" height="12" viewBox="0 0 16 12" fill="none" xmlns="http://www.w3.org/2000/svg" className="rounded-[2px] overflow-hidden shrink-0 shadow-[0_0_0_1px_rgba(0,0,0,0.18)]">
      <rect width="16" height="12" fill="#FFFFFF" />
      <rect width="16" height="1.85" fill="#DC2626" />
      <rect y="3.69" width="16" height="1.85" fill="#DC2626" />
      <rect y="7.38" width="16" height="1.85" fill="#DC2626" />
      <rect y="10.15" width="16" height="1.85" fill="#DC2626" />
      <rect width="7.5" height="6.5" fill="#1D4ED8" />
      <circle cx="2.2" cy="2" r="0.6" fill="#FFFFFF" />
      <circle cx="5.3" cy="2" r="0.6" fill="#FFFFFF" />
      <circle cx="3.75" cy="3.5" r="0.6" fill="#FFFFFF" />
      <circle cx="2.2" cy="5" r="0.6" fill="#FFFFFF" />
      <circle cx="5.3" cy="5" r="0.6" fill="#FFFFFF" />
    </svg>
  );
}

export default function HeaderLanguage() {
  const [locale, setLocale] = useState("id");
  const [isPending, setIsPending] = useState(false);

  useEffect(() => {
    setLocale(getLocaleFromCookie());
  }, []);

  const handleToggleLocale = async (targetLocale) => {
    const nextLocale = targetLocale || (locale === "id" ? "en" : "id");
    if (nextLocale === locale || isPending) return;

    setIsPending(true);
    try {
      await fetch("/api/locale", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ locale: nextLocale }),
      });
      await reloadTranslations();
      setLocale(nextLocale);
    } catch (err) {
      console.error("Failed to switch locale:", err);
    } finally {
      setIsPending(false);
    }
  };

  return (
    <div className="inline-flex items-center p-0.5 rounded-lg border border-black/10 dark:border-white/10 bg-black/5 dark:bg-white/5" data-i18n-skip="true">
      <button
        type="button"
        onClick={() => handleToggleLocale("id")}
        disabled={isPending}
        className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-[5px] text-xs font-semibold cursor-pointer transition-all duration-200 active:scale-[0.98] ${
          locale === "id"
            ? "bg-white dark:bg-zinc-800 text-text-main shadow-xs ring-1 ring-black/5 dark:ring-white/10"
            : "text-text-muted hover:text-text-main"
        } ${isPending ? "opacity-60 cursor-wait" : ""}`}
        title="Bahasa Indonesia"
      >
        <FlagID />
        <span>ID</span>
      </button>
      <button
        type="button"
        onClick={() => handleToggleLocale("en")}
        disabled={isPending}
        className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-[5px] text-xs font-semibold cursor-pointer transition-all duration-200 active:scale-[0.98] ${
          locale === "en"
            ? "bg-white dark:bg-zinc-800 text-text-main shadow-xs ring-1 ring-black/5 dark:ring-white/10"
            : "text-text-muted hover:text-text-main"
        } ${isPending ? "opacity-60 cursor-wait" : ""}`}
        title="English"
      >
        <FlagEN />
        <span>EN</span>
      </button>
    </div>
  );
}

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
    <div className="flex items-center rounded-lg border border-black/10 dark:border-white/10 p-0.5 bg-black/5 dark:bg-white/5" data-i18n-skip="true">
      <button
        onClick={() => handleToggleLocale("id")}
        disabled={isPending}
        className={`flex items-center gap-1 px-2 py-1 rounded-md text-xs font-semibold transition-all ${
          locale === "id"
            ? "bg-white dark:bg-zinc-800 text-text-main shadow-xs ring-1 ring-black/5 dark:ring-white/10"
            : "text-text-muted hover:text-text-main"
        } ${isPending ? "opacity-60 cursor-wait" : ""}`}
        title="Bahasa Indonesia"
      >
        <span className="text-sm">🇮🇩</span>
        <span>ID</span>
      </button>
      <button
        onClick={() => handleToggleLocale("en")}
        disabled={isPending}
        className={`flex items-center gap-1 px-2 py-1 rounded-md text-xs font-semibold transition-all ${
          locale === "en"
            ? "bg-white dark:bg-zinc-800 text-text-main shadow-xs ring-1 ring-black/5 dark:ring-white/10"
            : "text-text-muted hover:text-text-main"
        } ${isPending ? "opacity-60 cursor-wait" : ""}`}
        title="English"
      >
        <span className="text-sm">🇺🇸</span>
        <span>EN</span>
      </button>
    </div>
  );
}

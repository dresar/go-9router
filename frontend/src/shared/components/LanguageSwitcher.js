"use client";

import { useState, useEffect } from "react";
import { LOCALES, LOCALE_COOKIE, normalizeLocale } from "@/i18n/config";
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
    <svg width="18" height="13" viewBox="0 0 16 12" fill="none" xmlns="http://www.w3.org/2000/svg" className="rounded-[2px] overflow-hidden shrink-0 shadow-[0_0_0_1px_rgba(0,0,0,0.18)]">
      <rect width="16" height="6" fill="#EF4444" />
      <rect y="6" width="16" height="6" fill="#FFFFFF" />
    </svg>
  );
}

function FlagEN() {
  return (
    <svg width="18" height="13" viewBox="0 0 16 12" fill="none" xmlns="http://www.w3.org/2000/svg" className="rounded-[2px] overflow-hidden shrink-0 shadow-[0_0_0_1px_rgba(0,0,0,0.18)]">
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

export default function LanguageSwitcher({ className = "", isOpen, onClose, hideTrigger = false }) {
  const [locale, setLocale] = useState("id");
  const [isPending, setIsPending] = useState(false);

  useEffect(() => {
    setLocale(getLocaleFromCookie());
  }, []);

  const handleSetLocale = async (nextLocale) => {
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
      if (onClose) onClose(nextLocale);
    } catch (err) {
      console.error("Failed to set locale:", err);
    } finally {
      setIsPending(false);
    }
  };

  if (hideTrigger && !isOpen) return null;

  return (
    <div className={`flex items-center gap-1.5 ${className}`} data-i18n-skip="true">
      {LOCALES.map((item) => {
        const active = locale === item;
        const name = item === "id" ? "Bahasa Indonesia" : "English";
        return (
          <button
            key={item}
            type="button"
            onClick={() => handleSetLocale(item)}
            disabled={isPending}
            className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold cursor-pointer transition-all duration-200 active:scale-[0.98] ${
              active
                ? "bg-primary text-white shadow-xs"
                : "text-text-muted hover:text-text-main hover:bg-black/5 dark:hover:bg-white/5"
            } ${isPending ? "opacity-60 cursor-wait" : ""}`}
          >
            {item === "id" ? <FlagID /> : <FlagEN />}
            <span>{name}</span>
          </button>
        );
      })}
    </div>
  );
}

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

const getLocaleInfo = (locale) => {
  const locales = {
    "id": { name: "Bahasa Indonesia", flag: "🇮🇩" },
    "en": { name: "English", flag: "🇺🇸" },
  };
  return locales[locale] || { name: locale, flag: "🌐" };
};

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
    <div className={`flex items-center gap-1 ${className}`} data-i18n-skip="true">
      {LOCALES.map((item) => {
        const active = locale === item;
        const info = getLocaleInfo(item);
        return (
          <button
            key={item}
            onClick={() => handleSetLocale(item)}
            disabled={isPending}
            className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium transition-colors ${
              active
                ? "bg-primary text-white"
                : "text-text-muted hover:text-text-main hover:bg-black/5 dark:hover:bg-white/5"
            }`}
          >
            <span>{info.flag}</span>
            <span>{info.name}</span>
          </button>
        );
      })}
    </div>
  );
}

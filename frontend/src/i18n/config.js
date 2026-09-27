export const LOCALES = [
  "en",
  "id",
];
export const DEFAULT_LOCALE = "id";
export const LOCALE_COOKIE = "locale";

export const LOCALE_NAMES = {
  en: "English",
  id: "Bahasa Indonesia",
};

export function normalizeLocale(locale) {
  if (locale === "id" || locale === "in" || (typeof locale === "string" && locale.startsWith("id-"))) {
    return "id";
  }
  return "en";
}

export function isSupportedLocale(locale) {
  return LOCALES.includes(locale);
}

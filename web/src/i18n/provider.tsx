import * as React from "react";

// @ts-ignore: allow importing JSON files in TypeScript
import enUS from "../locales/en-US/common.json";
// @ts-ignore: allow importing JSON files in TypeScript
import ptBR from "../locales/pt-BR/common.json";

export type Locale = "pt-BR" | "en-US";

type Messages = typeof ptBR;

interface I18nContextValue {
  locale: Locale;
  t: (key: string, defaultValue?: string) => string;
  get: <T = unknown>(key: string, defaultValue?: T) => T;
  setLocale: (locale: Locale) => void;
}

const messagesMap: Record<Locale, Messages> = {
  "pt-BR": ptBR,
  "en-US": enUS,
};

const STORAGE_KEY = "kubex-locale";

const I18nContext = React.createContext<I18nContextValue | undefined>(undefined);

function getStoredLocale(): Locale | null {
  if (typeof window === "undefined") return null;
  const stored = window.localStorage.getItem(STORAGE_KEY) as Locale | null;
  if (stored === "pt-BR" || stored === "en-US") {
    return stored;
  }
  return null;
}

function getInitialLocale(): Locale {
  if (typeof window === "undefined") return "pt-BR";
  return getStoredLocale() ?? "pt-BR";
}

function resolveKey(messages: Messages, key: string): unknown {
  const parts = key.split(".");
  let current: unknown = messages;

  for (const part of parts) {
    if (current && typeof current === "object" && part in (current as Record<string, unknown>)) {
      current = (current as Record<string, unknown>)[part];
    } else {
      current = undefined;
      break;
    }
  }

  return current;
}

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = React.useState<Locale>(() => {
    const initial = getInitialLocale();
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("lang", initial);
    }
    return initial;
  });

  const setLocale = React.useCallback((next: Locale) => {
    setLocaleState(next);
    try {
      window.localStorage.setItem(STORAGE_KEY, next);
    } catch (error) {
      // ignore persistence issues
    }
  }, []);

  React.useEffect(() => {
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("lang", locale);
    }
  }, [locale]);

  const value = React.useMemo<I18nContextValue>(() => {
    const messages = messagesMap[locale] ?? messagesMap["pt-BR"];
    const t = (key: string, defaultValue?: string) => {
      const result = resolveKey(messages, key);
      if (typeof result === "string") {
        return result;
      }
      if (result == null && defaultValue) {
        return defaultValue;
      }
      return String(result ?? defaultValue ?? key);
    };
    const get = <T,>(key: string, defaultValue?: T) => {
      const result = resolveKey(messages, key);
      return (result as T | undefined) ?? defaultValue ?? (key as unknown as T);
    };
    return { locale, setLocale, t, get };
  }, [locale, setLocale]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const context = React.useContext(I18nContext);
  if (!context) {
    throw new Error("useI18n must be used within I18nProvider");
  }
  return context;
}

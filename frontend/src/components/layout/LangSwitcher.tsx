import * as React from "react";

import { Locale, useI18n } from "../../i18n/provider";

export function LangSwitcher() {
  const { locale, setLocale, t } = useI18n();

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const next = event.target.value as Locale;
    setLocale(next);
  };

  return (
    // <label className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.28em] text-slate-500 dark:text-slate-400">
    <label className="inline-flex items-center justify-center rounded-full px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary.hover dark:bg-cyan-500 dark:hover:bg-cyan-400">
      <span className="sr-only">{t("language.title", "Language")}</span>
      <select
        value={locale}
        onChange={handleChange}
        className="wh-full bg-transparent text-center text-sm font-semibold text-white outline-none *:ring-0"
        aria-label={t("language.title", "Language")}
      >
        <option value="pt-BR">{t("language.ptBR", "🇧🇷")}</option>
        <option value="en-US">{t("language.enUS", "🇺🇸")}</option>
      </select>
    </label>
  );
}

import * as React from "react";

import { Locale, useI18n } from "../../i18n/provider";

export function LangSwitcher() {
  const { locale, setLocale, t } = useI18n();

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const next = event.target.value as Locale;
    setLocale(next);
  };

  return (
    <label className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.28em] text-slate-500 dark:text-slate-400">
      {/* {t("language.title", "Lang")} */}
      <select
        value={locale}
        onChange={handleChange}
        className="rounded-full border border-slate-200 bg-white px-3 py-1 text-[0.7rem] font-semibold text-slate-700 shadow-sm transition hover:border-primary/50 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/30 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-cyan-400/50 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/30"
      >
        <option value="pt-BR">{t("language.ptBR", "Português (Brasil)")}</option>
        <option value="en-US">{t("language.enUS", "English (US)")}</option>
      </select>
    </label>
  );
}

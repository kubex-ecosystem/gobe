import * as React from "react";

import { useI18n } from "../../i18n/provider";
import { Link, useRouter } from "../../lib/router";
import { LangSwitcher } from "./LangSwitcher";
import { ThemeToggle } from "./ThemeToggle";

const NAV_ITEMS = [
  { key: "nav.home", label: "home", to: "/" },
  { key: "nav.manifesto", label: "manifesto", to: "/manifesto" },
  { key: "nav.access", label: "access", to: "/access" },
];

export function Navbar() {
  const { pathname } = useRouter();
  const { t } = useI18n();

  return (
    <header className="sticky top-0 z-20 border-b border-slate-200/60 bg-white/80 backdrop-blur dark:border-slate-800/70 dark:bg-slate-900/80">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
        <Link to="/" className="group flex items-center gap-3 text-slate-900 dark:text-slate-100">
          <span className="relative flex h-9 w-9 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-cyan-400/90 via-sky-400/80 to-fuchsia-500/80 shadow-sm">
            <span className="absolute inset-0 rounded-2xl bg-white/10" />
            <span className="relative text-sm font-semibold tracking-tight text-white">KX</span>
          </span>
          <div className="flex flex-col">
            <span className="text-sm font-semibold tracking-wide text-slate-900 dark:text-slate-100">Kubex GoBE</span>
            <span className="text-xs font-medium uppercase tracking-[0.3em] text-slate-500 dark:text-slate-400">Freedom Engineered</span>
          </div>
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-medium text-slate-700 dark:text-slate-300 md:flex">
          {NAV_ITEMS.map((item) => {
            const isActive = pathname === item.to;
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`rounded-full px-3 py-1 transition-colors ${isActive
                  ? "bg-primary.subtle text-primary.foreground dark:bg-slate-800 dark:text-primary"
                  : "hover:text-primary dark:hover:text-primary"
                  }`}
              >
                {t(item.key)}
              </Link>
            );
          })}
        </nav>

        <div className="flex items-center gap-3">
          <LangSwitcher />
          <ThemeToggle />
          <Link
            to="/access"
            className="hidden rounded-full border border-cyan-600/30 px-3 py-1 text-sm font-semibold text-cyan-700 shadow-sm transition hover:border-cyan-500/40 hover:text-cyan-600 dark:border-cyan-500/20 dark:text-cyan-300 dark:hover:border-cyan-400/50 dark:hover:text-cyan-200 sm:inline-flex"
          >
            {t("home.primaryCta", "Entrar")}
          </Link>
          <Link
            to="/access"
            className="inline-flex items-center justify-center rounded-full bg-primary px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary.hover dark:bg-cyan-500 dark:hover:bg-cyan-400"
          >
            {t("nav.signIn", "Entrar")}
          </Link>
        </div>
      </div>
    </header>
  );
}

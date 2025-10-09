import * as React from "react";
import { Key, Lock, LogIn } from "lucide-react";

import { Button } from "../ui/Button";
import { useI18n } from "../../i18n/provider";

const MODE_ORDER = ["kubex-id", "sso", "service"] as const;

type Mode = (typeof MODE_ORDER)[number];

interface AuthFormProps {
  onComplete?: (identifier: string) => void;
}

const baseTabStyles =
  "flex flex-col items-center gap-1 rounded-2xl border px-3 py-3 text-xs font-semibold transition";

export function AuthForm({ onComplete }: AuthFormProps) {
  const { t } = useI18n();
  const [mode, setMode] = React.useState<Mode>("kubex-id");
  const [status, setStatus] = React.useState<string | null>(null);

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const identifier = (form.get("email") || form.get("clientId") || "usuário").toString();
    setStatus(`Acesso preparado para ${identifier}.`);
    onComplete?.(identifier);
  };

  return (
    <div className="space-y-5">
      <div className="grid grid-cols-3 gap-2">
        {MODE_ORDER.map((tab) => {
          const isActive = mode === tab;
          const labelKey = tab === "kubex-id" ? "access.tabs.kubexId" : tab === "sso" ? "access.tabs.sso" : "access.tabs.token";
          return (
            <button
              key={tab}
              type="button"
              onClick={() => {
                setMode(tab);
                setStatus(null);
              }}
              className={`${baseTabStyles} ${
                isActive
                  ? "border-primary/60 bg-primary.subtle text-primary.foreground dark:border-cyan-500/40 dark:bg-slate-800/80 dark:text-cyan-200"
                  : "border-slate-200 bg-white text-slate-600 hover:border-primary/40 hover:text-primary dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:border-cyan-400/60 dark:hover:text-cyan-200"
              }`}
            >
              {t(labelKey)}
            </button>
          );
        })}
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {mode === "kubex-id" && (
          <>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.email")}
              <input
                required
                id="email"
                name="email"
                type="email"
                placeholder="coauthor@kubex.world"
                className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
              />
            </label>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.password")}
              <div className="relative mt-1">
                <input
                  required
                  id="password"
                  name="password"
                  type="password"
                  placeholder="********"
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
                />
                <Lock className="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300 dark:text-slate-500" />
              </div>
            </label>
          </>
        )}

        {mode === "sso" && (
          <>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.issuer")}
              <input
                required
                id="issuer"
                name="issuer"
                placeholder="https://login.corp.example/.well-known/openid-configuration"
                className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
              />
            </label>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.clientId")}
              <input
                required
                id="clientId"
                name="clientId"
                placeholder="kubex-gobe-client"
                className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
              />
            </label>
          </>
        )}

        {mode === "service" && (
          <>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.clientId")}
              <input
                required
                id="serviceClientId"
                name="clientId"
                placeholder="module-analyzer"
                className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
              />
            </label>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
              {t("access.form.clientSecret")}
              <div className="relative mt-1">
                <input
                  required
                  id="clientSecret"
                  name="clientSecret"
                  placeholder="••••••••••"
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
                />
                <Key className="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-300 dark:text-slate-500" />
              </div>
            </label>
          </>
        )}

        <Button type="submit" className="flex w-full items-center justify-center gap-2">
          <LogIn className="h-4 w-4" />
          {t("access.form.submit")}
        </Button>
        {status && <p className="text-sm font-semibold text-cyan-700 dark:text-cyan-200">{status}</p>}
      </form>
    </div>
  );
}

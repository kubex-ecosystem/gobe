import { Key, Lock, LogIn } from "lucide-react";
import * as React from "react";

import { useI18n } from "../../i18n/provider";
import { fetchJson } from "../../lib/api";
import { useRouter } from "../../lib/router";
import { useAuth } from "../../context/auth";
import { Button } from "../ui/Button";

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
  const [statusType, setStatusType] = React.useState<"success" | "error" | null>(null);
  const [submitting, setSubmitting] = React.useState(false);
  const { setAuth } = useAuth();
  const { navigate } = useRouter();

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const identifierField = form.get("identifier") || form.get("email") || form.get("clientId");
    const identifier = (identifierField || "usuário").toString();

    setSubmitting(true);
    setStatus(null);
    setStatusType(null);

    try {
      if (mode === "kubex-id") {
        const password = (form.get("password") || "").toString();
        if (!identifier || !password) {
          setStatusType("error");
          setStatus(t("access.feedback.error", "Authentication failed. Please check your credentials."));
          return;
        }

        const response = await fetchJson<AuthResponse>({
          path: "/api/v1/sign-in",
          method: "POST",
          body: {
            username: identifier,
            password,
          },
        });

        setAuth(response.access_token, response.user);
        if (typeof window !== "undefined") {
          window.sessionStorage.setItem("kubex:refreshToken", response.refresh_token);
        }

        setStatusType("success");
        setStatus(t("access.feedback.success", "You're authenticated. Tokens stored locally."));
        onComplete?.(identifier);
        navigate("/dashboard", { replace: true });
        return;
      }

      setStatusType("success");
      setStatus(
        t("access.feedback.placeholder", "Access prepared for {{identifier}}.", {
          identifier,
        })
      );
      onComplete?.(identifier);
    } catch (error) {
      setStatusType("error");
      if (error instanceof Error) {
        setStatus(error.message);
      } else {
        setStatus(t("access.feedback.error", "Authentication failed. Please check your credentials."));
      }
    } finally {
      setSubmitting(false);
    }
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
              className={`${baseTabStyles} ${isActive
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
              {t("access.form.identifier")}
              <input
                required
                id="identifier"
                name="identifier"
                type="text"
                placeholder="testUser ou coauthor@kubex.world"
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

        <Button
          type="submit"
          disabled={submitting}
          className="flex w-full items-center justify-center gap-2 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <LogIn className="h-4 w-4" />
          {t("access.form.submit")}
        </Button>
        {status && (
          <p
            className={`text-sm font-semibold ${
              statusType === "error"
                ? "text-rose-600 dark:text-rose-300"
                : "text-cyan-700 dark:text-cyan-200"
            }`}
          >
            {status}
          </p>
        )}
      </form>
    </div>
  );
}

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  refresh_expires_in: number;
  user: {
    id: string;
    username: string;
    email: string;
    name: string;
    role: string;
    active: boolean;
  };
}

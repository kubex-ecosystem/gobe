import { CheckCircle2, Key, Lock, LogIn, Shield } from "lucide-react";
import * as React from "react";

import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";
import { Card, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card";
import { useI18n } from "../../i18n/provider";
import { Link, useRouter } from "../../lib/router";

const modes = [
  { id: "kubex-id", label: "Kubex ID", description: "Email corporativo + MFA" },
  { id: "sso", label: "SSO", description: "Provedor OIDC / SAML" },
  { id: "service", label: "Token de serviço", description: "Client credentials" },
] as const;

type Mode = (typeof modes)[number]["id"];

export function AuthPage() {
  const [mode, setMode] = React.useState<Mode>("kubex-id");
  const [status, setStatus] = React.useState<string | null>(null);
  const { navigate } = useRouter();
  const { t } = useI18n();

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const identifier = (form.get("email") || form.get("clientId") || "usuário").toString();
    setStatus(`Acesso preparado para ${identifier}.`);
    window.setTimeout(() => navigate("/app/"), 900);
  };

  return (
    <div className="relative mx-auto flex max-w-5xl flex-col-reverse gap-14 px-4 pb-24 pt-12 text-slate-700 transition-colors duration-300 sm:px-6 md:flex-row md:items-start md:pb-28 md:pt-16 dark:text-slate-300">
      <div className="md:w-1/2">
        <Badge tone="cyan">{t("auth.badge", "Portal de Autenticação")}</Badge>
        <h1 className="mt-5 text-3xl font-semibold text-slate-900 dark:text-white">{t("auth.title", "Conceda acesso com governança Kubex")}</h1>
        <p className="mt-3 text-slate-600 dark:text-slate-300">
          {t("auth.description", "Escolha o fluxo adequado para humanos, agentes ou integrações. Cada tentativa é auditável e respeita as diretivas do manifesto Kubex.")}
        </p>

        <div className="mt-8 space-y-4">
          <div className="flex items-start gap-3 rounded-2xl border border-cyan-600/15 bg-cyan-50/70 p-4 text-sm text-cyan-700 dark:border-cyan-500/30 dark:bg-slate-900/70 dark:text-cyan-200">
            <Shield className="mt-0.5 h-5 w-5" />
            <p>
              {t("auth.tokensInfo", "Tokens e credenciais são armazenados no núcleo GoBE. Defina escopos, TTLs e políticas de rotação por módulo.")}
            </p>
          </div>
          <ul className="space-y-3 text-sm text-slate-600 dark:text-slate-300">
            <li className="flex items-start gap-2">
              <CheckCircle2 className="mt-0.5 h-4 w-4 text-cyan-600 dark:text-cyan-300" />
              {t("auth.mfaSupport", "Suporte a MFA via TOTP, WebAuthn e tokens físicos.")}
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="mt-0.5 h-4 w-4 text-cyan-600 dark:text-cyan-300" />
              {t("auth.enrichedLogs", "Logs enriquecidos com origem, dispositivo e agente responsável.")}
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle2 className="mt-0.5 h-4 w-4 text-cyan-600 dark:text-cyan-300" />
              {t("auth.automaticProvisioning", "Provisionamento automatizado via CLI ou APIs quando necessário.")}
            </li>
          </ul>

          <Link
            to="/app/about"
            className="inline-flex items-center gap-2 text-sm font-semibold text-cyan-700 transition hover:text-cyan-600 dark:text-cyan-200 dark:hover:text-cyan-100"
          >
            Manifesto Kubex
            <span aria-hidden>→</span>
          </Link>
        </div>
      </div>

      <Card className="md:w-1/2">
        <CardHeader>
          <Badge tone="neutral" className="self-start">{t("auth.flowSelection", "Escolha o fluxo")}</Badge>
          <CardTitle>{t("auth.coauthorAuthentication", "Autenticar coautores")}</CardTitle>
          <CardDescription>
            {t("auth.gobeValidation", "O GoBE valida as credenciais e aplica políticas definidas em")} <code>internal/module/module.go</code> {t("auth.cliRelated", "e CLI relacionadas.")}.
          </CardDescription>
        </CardHeader>

        <div className="grid grid-cols-3 gap-2">
          {modes.map(({ id, label }) => {
            const isActive = id === mode;
            return (
              <button
                key={id}
                type="button"
                onClick={() => setMode(id)}
                className={`flex flex-col items-center gap-1 rounded-2xl border px-3 py-3 text-xs font-semibold transition ${isActive
                  ? "border-primary/60 bg-primary.subtle text-primary.foreground dark:border-cyan-500/40 dark:bg-slate-800/80 dark:text-cyan-200"
                  : "border-slate-200 bg-white text-slate-600 hover:border-primary/40 hover:text-primary dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:border-cyan-400/60 dark:hover:text-cyan-200"
                  }`}
              >
                {label}
              </button>
            );
          })}
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-5">
          {mode === "kubex-id" && (
            <>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
                {t("auth.corporateEmail", "Email corporativo")}
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
                {t("auth.password", "Senha")}
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
                {t("auth.issuerUrl", "Issuer URL")}
                <input
                  required
                  id="issuer"
                  name="issuer"
                  placeholder="https://login.corp.example/.well-known/openid-configuration"
                  className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
                />
              </label>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
                {t("auth.clientId", "Client ID")}
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
                {t("auth.serviceClientId", "Client ID")}
                <input
                  required
                  id="serviceClientId"
                  name="clientId"
                  placeholder="module-analyzer"
                  className="mt-1 w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/40 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-cyan-400 dark:focus:ring-cyan-400/40"
                />
              </label>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-200">
                {t("auth.clientSecret", "Client Secret")}
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
            {t("auth.signIn", "Entrar")}
          </Button>
          {status && <p className="text-sm font-semibold text-cyan-700 dark:text-cyan-200">{status}</p>}
        </form>
      </Card>
    </div>
  );
}

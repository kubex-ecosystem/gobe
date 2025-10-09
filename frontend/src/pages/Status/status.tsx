import * as React from "react";
import { Activity, Database, RefreshCw, Server, ShieldCheck } from "lucide-react";

import { Card, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card";
import { KxBadge } from "../../components/ui/KxBadge";
import { useI18n } from "../../i18n/provider";
import { useHealthStatus } from "../../hooks/useHealthStatus";
import { useProvidersStatus } from "../../hooks/useProvidersStatus";
import { useScorecard } from "../../hooks/useScorecard";

const statusColorMap = {
  operational: "text-cyan-600 dark:text-cyan-300",
  degraded: "text-amber-600 dark:text-amber-300",
  down: "text-rose-600 dark:text-rose-300",
  unknown: "text-slate-500 dark:text-slate-400",
};

export function StatusPage() {
  const { t } = useI18n();
  const health = useHealthStatus(60000); // Poll every 60 seconds
  const providers = useProvidersStatus(60000); // Poll every 60 seconds
  const scorecard = useScorecard(4, 120000); // Poll every 2 minutes

  const serviceEntries = React.useMemo(() => {
    return Object.entries(health.data.services).map(([key, value]) => ({
      key,
      label: t(`status.services.${key}`, key),
      healthy: value.healthy,
      detail: value.detail,
    }));
  }, [health.data.services, t]);

  const providerList = providers.data?.providers ?? [];
  const scorecardItems = scorecard.data?.items ?? [];

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-12 px-4 pb-24 pt-12 text-slate-700 sm:px-6 md:gap-16 md:pt-16 dark:text-slate-300">
      <section className="space-y-4">
        <KxBadge tone="cyan">{t("status.badge", "GoBE status")}</KxBadge>
        <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div className="space-y-3">
            <h1 className="text-3xl font-semibold text-slate-900 dark:text-white">{t("status.title")}</h1>
            <p className="max-w-2xl text-slate-600 dark:text-slate-300">{t("status.subtitle")}</p>
          </div>
          <button
            type="button"
            onClick={() => {
              health.refresh();
              if (providerList.length) {
                providers.refetch();
              }
              if (scorecardItems.length) {
                scorecard.refetch();
              }
            }}
            className="inline-flex items-center gap-2 rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-primary/50 hover:text-primary dark:border-slate-700 dark:text-slate-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-200"
          >
            <RefreshCw className="h-4 w-4" />
            {t("status.refresh", "Atualizar")}
          </button>
        </div>
      </section>

      <section className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <KxBadge tone="neutral">{t("status.health.label", "Saúde geral")}</KxBadge>
            <CardTitle className="flex items-center gap-2">
              <ShieldCheck className="h-5 w-5 text-cyan-500" />
              {health.data.status}
            </CardTitle>
            <CardDescription>{t("status.health.description")}</CardDescription>
          </CardHeader>
          <div className="mt-4 space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-slate-600 dark:text-slate-300">{t("status.health.uptime", "Tempo de atividade")}</span>
              <span className="font-semibold text-slate-900 dark:text-white">{health.data.uptime ?? t("status.unavailable")}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600 dark:text-slate-300">{t("status.health.version", "Versão")}</span>
              <span className="font-semibold text-slate-900 dark:text-white">{health.data.version ?? "--"}</span>
            </div>
          </div>
        </Card>
        <Card>
          <CardHeader>
            <KxBadge tone="lilac">{t("status.providers.badge", "Provedores ativos")}</KxBadge>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5 text-fuchsia-500" />
              {providerList.length}
            </CardTitle>
            <CardDescription>
              {providers.loading
                ? t("status.providers.loading")
                : providerList.length
                  ? t("status.providers.description")
                  : t("status.providers.empty")}
            </CardDescription>
          </CardHeader>
        </Card>
      </section>

      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Activity className="h-5 w-5 text-cyan-500" />
          <h2 className="text-xl font-semibold text-slate-900 dark:text-white">{t("status.services.title")}</h2>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {serviceEntries.length === 0 && (
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>{t("status.unavailable")}</CardTitle>
                <CardDescription>{t("status.services.unavailable")}</CardDescription>
              </CardHeader>
            </Card>
          )}
          {serviceEntries.map((service) => (
            <Card key={service.key}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>{service.label}</CardTitle>
                  <span className={`text-xs font-semibold uppercase tracking-[0.18em] ${statusColorMap[service.healthy ? "operational" : "down"]}`}>
                    {service.healthy ? t("runtime.operational", "Operational") : t("runtime.down", "Down")}
                  </span>
                </div>
                {service.detail && (
                  <CardDescription>
                    {t("status.services.detail", {
                      total: service.detail.total ?? 0,
                      available: service.detail.available ?? 0,
                    })}
                  </CardDescription>
                )}
              </CardHeader>
            </Card>
          ))}
        </div>
      </section>

      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Database className="h-5 w-5 text-cyan-500" />
          <h2 className="text-xl font-semibold text-slate-900 dark:text-white">{t("status.providers.title")}</h2>
        </div>
        <div className="grid gap-3">
          {providerList.length === 0 && (
            <div className="rounded-2xl border border-slate-200 bg-white/80 px-4 py-3 text-sm text-slate-600 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-300">
              {providers.loading ? t("status.providers.loading") : t("status.providers.empty")}
            </div>
          )}
          {providerList.map((provider) => (
            <div
              key={`${provider.name}-${provider.defaultModel}`}
              className="flex flex-col gap-1 rounded-2xl border border-slate-200 bg-white/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-900/70"
            >
              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold text-slate-800 dark:text-slate-100">{provider.name}</span>
                <span className={`text-xs font-semibold ${statusColorMap[provider.available ? "operational" : "down"]}`}>
                  {provider.available ? t("runtime.operational", "Operational") : t("runtime.down", "Down")}
                </span>
              </div>
              <span className="text-xs text-slate-500 dark:text-slate-400">
                {provider.type} • {provider.defaultModel || t("status.providers.genericModel", "Modelo padrão")}
              </span>
            </div>
          ))}
        </div>
      </section>

      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <ShieldCheck className="h-5 w-5 text-cyan-500" />
          <h2 className="text-xl font-semibold text-slate-900 dark:text-white">{t("status.scorecard.title")}</h2>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {scorecardItems.length === 0 && (
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>{t("status.scorecard.emptyTitle", "Sem análises recentes")}</CardTitle>
                <CardDescription>{scorecard.loading ? t("status.scorecard.loading") : t("status.scorecard.empty")}</CardDescription>
              </CardHeader>
            </Card>
          )}
          {scorecardItems.map((item) => (
            <Card key={item.id ?? item.title}>
              <CardHeader>
                <CardTitle className="text-base">{item.title}</CardTitle>
                <CardDescription>{item.description}</CardDescription>
                {item.score !== undefined && (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white/80 px-3 py-1 text-xs font-semibold text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
                    <span>{t("status.scorecard.score", "Score")}</span>
                    <span>{(item.score * 100).toFixed(0)}%</span>
                  </div>
                )}
              </CardHeader>
            </Card>
          ))}
        </div>
      </section>
    </div>
  );
}

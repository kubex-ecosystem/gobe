import * as React from "react";

import { useI18n } from "../../i18n/provider";
import { useHealthStatus } from "../../hooks/useHealthStatus";

type RuntimeStatus = "operational" | "degraded" | "down" | "unknown";

const statusClasses: Record<RuntimeStatus, string> = {
  operational: "text-cyan-600 dark:text-cyan-300",
  degraded: "text-amber-600 dark:text-amber-300",
  down: "text-rose-600 dark:text-rose-300",
  unknown: "text-slate-500 dark:text-slate-400",
};

const SERVICE_MAPPING: Record<string, string> = {
  "Auth flows": "providers",
  "Event bus": "webhooks",
  "Modules health": "database",
};

function mapServiceStatus(healthy?: boolean): RuntimeStatus {
  if (healthy === undefined) return "unknown";
  return healthy ? "operational" : "down";
}

export function RuntimeOverview() {
  const { t, get } = useI18n();
  const items = get<string[]>("home.runtime.items", []);
  const { data: health, loading } = useHealthStatus();

  const overallStatus: RuntimeStatus = health.status === "ok"
    ? "operational"
    : health.status === "degraded"
      ? "degraded"
      : health.status === "down"
        ? "down"
        : "unknown";

  const rows = items.map((label) => {
    const serviceKey = SERVICE_MAPPING[label] ?? label.toLowerCase();
    const service = health.services[serviceKey];
    const status = mapServiceStatus(service?.healthy ?? health.status === "ok");
    return { label, status };
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <span className="text-xs font-semibold uppercase tracking-[0.25em] text-cyan-600/80 dark:text-cyan-300/80">
            {t("home.runtime.title")}
          </span>
          <h3 className="text-2xl font-semibold text-slate-900 dark:text-white">Runtime overview</h3>
        </div>
        <span className={`text-xs font-semibold uppercase tracking-[0.18em] ${statusClasses[overallStatus]}`}>
          {health.status}
        </span>
      </div>
      <div className="grid gap-4">
        {rows.map(({ label, status }) => (
          <div
            key={label}
            className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-900/70"
          >
            <span className="text-sm font-medium text-slate-600 dark:text-slate-300">{label}</span>
            <span className={`text-sm font-semibold ${statusClasses[status]}`}>
              {loading && status === "unknown" ? t("runtime.operational", "Operational") : t(`runtime.${status}`, status)}
            </span>
          </div>
        ))}
      </div>
      <div className="rounded-2xl border border-fuchsia-500/20 bg-fuchsia-50/70 p-4 text-sm text-fuchsia-700 dark:border-fuchsia-500/30 dark:bg-fuchsia-500/10 dark:text-fuchsia-200">
        {t("home.runtime.caption", "Freedom, Engineered — every module keeps autonomy yet speaks the same language.")}
      </div>
    </div>
  );
}

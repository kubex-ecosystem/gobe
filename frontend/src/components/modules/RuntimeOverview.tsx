import * as React from "react";

import { useI18n } from "../../i18n/provider";

interface RuntimeItem {
  name: string;
  status: "operational" | "degraded" | "down";
}

const STATUS_ORDER: RuntimeItem["status"][] = ["operational", "degraded", "down"];

function useRuntimeStub(baseItems: string[]) {
  const [tick, setTick] = React.useState(0);

  React.useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    const id = window.setInterval(() => {
      setTick((prev) => (prev + 1) % STATUS_ORDER.length);
    }, 8000);
    return () => window.clearInterval(id);
  }, []);

  return baseItems.map((name, index) => {
    const status = index === 0 ? "operational" : STATUS_ORDER[(tick + index) % STATUS_ORDER.length];
    return { name, status } satisfies RuntimeItem;
  });
}

const statusClasses: Record<RuntimeItem["status"], string> = {
  operational: "text-cyan-600 dark:text-cyan-300",
  degraded: "text-amber-600 dark:text-amber-300",
  down: "text-rose-600 dark:text-rose-300",
};

export function RuntimeOverview() {
  const { t, get } = useI18n();
  const items = get<string[]>("home.runtime.items", []);
  const statuses = useRuntimeStub(items);

  return (
    <div className="space-y-4">
      <div>
        <span className="text-xs font-semibold uppercase tracking-[0.25em] text-cyan-600/80 dark:text-cyan-300/80">
          {t("home.runtime.title")}
        </span>
        <h3 className="text-2xl font-semibold text-slate-900 dark:text-white">Runtime overview</h3>
      </div>
      <div className="grid gap-4">
        {statuses.map(({ name, status }) => (
          <div
            key={name}
            className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white/90 px-4 py-3 dark:border-slate-700 dark:bg-slate-900/70"
          >
            <span className="text-sm font-medium text-slate-600 dark:text-slate-300">{name}</span>
            <span className={`text-sm font-semibold ${statusClasses[status]}`}>
              {t(`runtime.${status}`, status)}
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

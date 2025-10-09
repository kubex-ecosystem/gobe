import * as React from "react";

import { ArrowRight, Layers, Sparkles, Users, Workflow } from "lucide-react";

import { HexBackground } from "../../components/layout/HexBackground";
import { Button } from "../../components/ui/Button";
import { Card, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card";
import { KxBadge } from "../../components/ui/KxBadge";
import { useI18n } from "../../i18n/provider";
import { Link } from "../../lib/router";

const pillarIcons = [Layers, Workflow, Sparkles, Users];

export default function AboutPage() {
  const { t, get } = useI18n();
  const pillars = get<{ title: string; description: string }[]>("about.pillars", []);
  const timeline = get<{ title: string; period: string; description: string }[]>("about.timeline", []);
  const values = get<{ title: string; description: string }[]>("about.values", []);

  return (
    <div className="relative mx-auto flex max-w-6xl flex-col gap-14 px-4 pb-24 pt-12 text-slate-700 sm:px-6 md:gap-20 md:pt-16 dark:text-slate-300">
      <HexBackground className="-z-10" />
      <section className="space-y-6">
        <KxBadge tone="cyan">{t("about.badge", "Kubex sobre nós")}</KxBadge>
        <div className="max-w-3xl space-y-4">
          <h1 className="font-display text-4xl font-extrabold tracking-tight text-slate-900 sm:text-5xl dark:text-white">
            {t("about.title", "Kubex é independência com governança elegante")}
          </h1>
          <p className="text-lg leading-relaxed text-slate-600 dark:text-slate-300">{t("about.subtitle")}</p>
        </div>
        <div className="flex flex-wrap gap-3">
          {(get<string[]>("about.badges", [])).map((badge) => (
            <KxBadge key={badge}>{badge}</KxBadge>
          ))}
        </div>
        <div className="flex flex-col gap-3 sm:flex-row">
          <Link
            to="/app/access"
            className="inline-flex items-center justify-center rounded-full bg-primary px-6 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-primary.hover dark:bg-cyan-500 dark:hover:bg-cyan-400"
          >
            {t("about.primaryCta", "Entrar no GoBE")}
          </Link>
          <Link
            to="/app/manifesto"
            className="inline-flex items-center justify-center rounded-full border border-slate-200 px-6 py-3 text-sm font-semibold text-slate-700 transition hover:border-primary/50 hover:text-primary dark:border-slate-700 dark:text-slate-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-200"
          >
            {t("about.secondaryCta", "Ler o manifesto")}
          </Link>
        </div>
      </section>

      <section className="grid gap-6 md:grid-cols-2">
        {values.map((item) => (
          <Card key={item.title} className="h-full">
            <CardHeader>
              <CardTitle>{item.title}</CardTitle>
              <CardDescription>{item.description}</CardDescription>
            </CardHeader>
          </Card>
        ))}
      </section>

      <section className="space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <KxBadge tone="neutral">{t("about.pillarsTitle", "Pilares Kubex")}</KxBadge>
            <h2 className="mt-3 text-2xl font-semibold text-slate-900 dark:text-white">
              {t("about.pillarsSubtitle", "Eixos que mantêm o GoBE coeso")}
            </h2>
          </div>
          <Link to="/app/status" className="inline-flex items-center gap-2 text-sm font-semibold text-primary hover:text-primary.hover">
            {t("about.viewStatus", "Ver status ao vivo")}
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
        <div className="grid gap-6 md:grid-cols-2">
          {pillars.map((pillar, index) => {
            const Icon = pillarIcons[index % pillarIcons.length];
            return (
              <Card key={pillar.title} className="flex h-full flex-col gap-3">
                <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-primary.subtle text-primary dark:bg-slate-800/80 dark:text-cyan-300">
                  <Icon className="h-5 w-5" />
                </span>
                <CardTitle>{pillar.title}</CardTitle>
                <CardDescription>{pillar.description}</CardDescription>
              </Card>
            );
          })}
        </div>
      </section>

      <section className="space-y-6">
        <div>
          <KxBadge tone="lilac">{t("about.timelineTitle", "Linha do tempo")}</KxBadge>
          <h2 className="mt-3 text-2xl font-semibold text-slate-900 dark:text-white">
            {t("about.timelineSubtitle", "Evolução contínua do GoBE")}
          </h2>
        </div>
        <div className="space-y-4">
          {timeline.map((entry) => (
            <div key={entry.title} className="rounded-3xl border border-slate-200 bg-white/90 p-5 shadow-sm dark:border-slate-700 dark:bg-slate-900/80">
              <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                <span className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-600/80 dark:text-cyan-300/80">
                  {entry.period}
                </span>
                <KxBadge tone="cyan">{entry.title}</KxBadge>
              </div>
              <p className="mt-3 text-sm text-slate-600 dark:text-slate-300">{entry.description}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

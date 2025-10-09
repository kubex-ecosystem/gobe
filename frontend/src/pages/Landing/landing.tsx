import { motion } from "framer-motion";
import { Layers, LucideIcon, Radar, ShieldCheck, Sparkle, Workflow } from "lucide-react";
import * as React from "react";

import { RuntimeOverview } from "../../components/modules/RuntimeOverview";
import { Button } from "../../components/ui/Button";
import { Card, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card";
import { KxBadge } from "../../components/ui/KxBadge";
import { useI18n } from "../../i18n/provider";
import { Link, useRouter } from "../../lib/router";

interface Feature {
  title: string;
  description: string;
  icon: LucideIcon;
}

const iconMap: Record<number, LucideIcon> = {
  0: ShieldCheck,
  1: Layers,
  2: Workflow,
  3: Radar,
};

const pillarIcon: Record<number, LucideIcon> = {
  0: Sparkle,
  1: Layers,
  2: Workflow,
};

export function LandingPage() {
  const { t, get } = useI18n();
  const { navigate } = useRouter();

  const badges = get<string[]>("home.badges", []);
  const featureData = get<{ title: string; description: string }[]>("home.features", []);
  const pillarData = get<{ title: string; description: string }[]>("home.sustainable.items", []);

  const features: Feature[] = featureData.map((item, index) => ({
    title: item.title,
    description: item.description,
    icon: iconMap[index as keyof typeof iconMap] ?? ShieldCheck,
  }));

  const pillars = pillarData.map((item, index) => ({
    title: item.title,
    description: item.description,
    icon: pillarIcon[index as keyof typeof pillarIcon] ?? Sparkle,
  }));

  return (
    <div className="relative mx-auto flex max-w-6xl flex-col gap-20 px-4 pb-24 pt-12 text-slate-700 transition-colors duration-300 sm:px-6 md:gap-24 md:pt-16 dark:text-slate-300">
      <section className="grid gap-12 md:grid-cols-[1.2fr,0.8fr] md:items-center">
        <div className="space-y-6">
          <div className="flex flex-wrap gap-2">
            {badges.map((label) => (
              <KxBadge key={label}>{label}</KxBadge>
            ))}
          </div>
          <div className="space-y-4">
            <motion.h1
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: "easeOut" }}
              className="text-balance text-4xl font-extrabold tracking-tight text-slate-900 sm:text-5xl lg:text-6xl dark:text-white"
            >
              {t("home.headline")}
            </motion.h1>
            <p className="max-w-xl text-lg leading-relaxed text-slate-600 dark:text-slate-300">
              {t("home.subtitle")}
            </p>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <Link
              to="/access"
              className="inline-flex items-center justify-center rounded-full bg-primary px-6 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-primary.hover dark:bg-cyan-500 dark:hover:bg-cyan-400"
            >
              {t("home.primaryCta")}
            </Link>
            <Link
              to="/manifesto"
              className="inline-flex items-center justify-center rounded-full border border-slate-200 px-6 py-3 text-sm font-semibold text-slate-700 transition hover:border-primary/50 hover:text-primary dark:border-slate-700 dark:text-slate-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-300"
            >
              {t("home.secondaryCta")}
            </Link>
          </div>
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6, delay: 0.1, ease: "easeOut" }}
          className="relative overflow-hidden rounded-3xl border border-slate-200/80 bg-white/70 p-6 shadow-soft-card backdrop-blur dark:border-slate-700 dark:bg-slate-900/70"
        >
          <RuntimeOverview />
          <div className="pointer-events-none absolute -right-10 -top-10 h-40 w-40 rounded-full bg-gradient-to-br from-cyan-400/30 via-sky-400/20 to-transparent blur-3xl" />
          <div className="pointer-events-none absolute -bottom-12 -left-12 h-48 w-48 rounded-full bg-gradient-to-br from-fuchsia-400/25 via-transparent to-transparent blur-3xl" />
        </motion.div>
      </section>

      <section className="space-y-12">
        <div className="mx-auto max-w-3xl text-center">
          <KxBadge tone="neutral" className="mx-auto">{t("home.matrix.badge", "Kubex matrix")}</KxBadge>
          <h2 className="mt-4 text-3xl font-semibold text-slate-900 dark:text-white">
            {t("home.matrix.heading", "Blocks that connect modules, agents, and teams")}
          </h2>
          <p className="mt-3 text-lg text-slate-600 dark:text-slate-300">
            {t(
              "home.matrix.copy",
              "Every feature is built to cooperate with CLI, APIs, and automations. Deliver consistent experiences even in hybrid environments."
            )}
          </p>
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {features.map(({ title, description, icon: Icon }) => (
            <Card key={title} className="flex flex-col gap-4">
              <CardHeader className="flex items-start gap-4">
                <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-primary.subtle text-primary dark:bg-slate-800/70 dark:text-cyan-300">
                  <Icon className="h-6 w-6" />
                </span>
                <CardTitle className="text-xl">{title}</CardTitle>
              </CardHeader>
              <CardDescription>{description}</CardDescription>
            </Card>
          ))}
        </div>
      </section>

      <section className="grid gap-8 rounded-3xl border border-slate-200 bg-white/95 p-10 shadow-soft-card backdrop-blur md:grid-cols-[1fr,1fr] dark:border-slate-700 dark:bg-slate-900/80">
        <div className="space-y-5">
          <KxBadge>Kubex</KxBadge>
          <h2 className="text-3xl font-semibold text-slate-900 dark:text-white">{t("home.sustainable.title")}</h2>
          <p className="text-slate-600 dark:text-slate-300">
            Do design system ao log streaming, o GoBE preserva a filosofia de interoperabilidade — cada decisão tem contexto, métricas e rastreabilidade.
          </p>
          <div className="flex flex-wrap gap-3">
            {pillars.map(({ title, description, icon: Icon }) => (
              <div
                key={title}
                className="flex items-start gap-3 rounded-2xl border border-slate-200 bg-white/90 p-4 dark:border-slate-700 dark:bg-slate-900/70"
              >
                <span className="rounded-xl bg-primary.subtle p-2 text-primary dark:bg-slate-800/80 dark:text-cyan-300">
                  <Icon className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">{title}</p>
                  <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">{description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
        <div className="flex flex-col justify-between gap-8 rounded-[28px] border border-cyan-500/20 bg-cyan-50/60 p-6 text-slate-700 dark:border-cyan-500/30 dark:bg-slate-900/70 dark:text-slate-200">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.3em] text-cyan-600/80">Pronto para uso</p>
            <h3 className="mt-3 text-2xl font-semibold text-slate-900 dark:text-white">{t("home.ready.title")}</h3>
            <p className="mt-3 text-sm leading-relaxed text-slate-600 dark:text-slate-300">{t("home.ready.description")}</p>
          </div>
          <div className="flex flex-col gap-3">
            <Button className="w-full" onClick={() => navigate("/access")}>{t("home.ready.primary")}</Button>
            <Link
              to="/access"
              className="w-full rounded-full border border-cyan-500/40 px-5 py-2 text-center text-sm font-semibold text-cyan-700 transition hover:border-cyan-500/60 hover:text-cyan-600 dark:border-cyan-500/30 dark:text-cyan-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-100"
            >
              {t("home.ready.secondary")}
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}

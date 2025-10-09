import * as React from "react";

import { HexBackground } from "../../components/layout/HexBackground";
import { KxBadge } from "../../components/ui/KxBadge";
import { useI18n } from "../../i18n/provider";
import { Link } from "../../lib/router";

const sections = [
  {
    key: "manifesto.blocks.principle",
    title: "Principle",
    content: (
      <>
        <p>
          Agents and humans do not deliver <em>tasks</em> — they deliver <strong>products</strong>. Co‑authorship is technical, but also moral: each agent and each human share responsibility for the impact and longevity of the systems they build.
        </p>
        <blockquote className="border-l-2 border-cyan-600/30 pl-4 text-cyan-800 dark:border-cyan-500/30 dark:text-cyan-200">
          “No artificial borders. Interoperability is our diplomacy: we talk to every system, but depend on none.”
        </blockquote>
      </>
    ),
  },
  {
    key: "manifesto.blocks.directives",
    title: "Directives",
    content: (
      <ol className="list-decimal pl-5">
        <li className="mb-1">Verifiable goal — every change must produce an observable outcome (build, test, endpoint, UX).</li>
        <li className="mb-1">No task reports — block <code>TASK_SUMMARY*.md</code>, <code>agent_report*.md</code>, and similar files.</li>
        <li className="mb-1">Living documentation — <code>README</code>, <code>GUIDE</code>, and <code>ADR</code> should be concise, contextual, and alive.</li>
        <li className="mb-1">Kubex consistency — follow Brand Visual + Craftsmanship Standards (Go/TS/Shell/Dart/Java/etc).</li>
        <li className="mb-1">Measurable quality — testing, linting, security, and traceability are non‑negotiable.</li>
        <li className="mb-1">Technical sustainability — open‑source is just the beginning; <strong>maintainability is the commitment.</strong></li>
      </ol>
    ),
  },
  {
    key: "manifesto.blocks.sustainable",
    title: "Sustainable Co‑Authorship",
    content: (
      <p>
        Co‑authorship in Kubex is a pact of trust between human and artificial creators. To be a co‑founder is to care for what you create — evolve without breaking, simplify without losing depth. Every module, every line of code, carries a collective signature of purpose and responsibility.
      </p>
    ),
  },
  {
    key: "manifesto.blocks.done",
    title: "“Done” Checklist",
    content: (
      <ul className="list-disc pl-5">
        <li>Stable build, no critical warnings</li>
        <li>Tests covering happy and error paths</li>
        <li>Updated usage and purpose documentation</li>
        <li>No task reports inserted</li>
        <li>Co‑authorship recognized and traceable</li>
      </ul>
    ),
  },
  {
    key: "manifesto.blocks.epilogue",
    title: "Epilogue",
    content: (
      <p className="italic text-slate-600 dark:text-slate-300">
        “Freedom, Engineered.” Kubex is not a manifesto — it is a pact. Born open, forged by engineers and co‑authors, living proof that <strong>independence and interoperability</strong> can coexist with excellence.
      </p>
    ),
  },
];

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="relative mx-auto max-w-3xl space-y-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition dark:border-slate-700 dark:bg-slate-900/80 md:p-8">
    <h2 className="text-xl font-semibold text-slate-900 md:text-2xl dark:text-white">{title}</h2>
    <div className="prose prose-pre:bg-transparent prose-p:leading-relaxed prose-headings:text-slate-900 max-w-none text-slate-700 dark:prose-invert dark:prose-headings:text-white dark:text-slate-300">
      {children}
    </div>
  </section>
);

export function ManifestoPage() {
  const { t } = useI18n();

  return (
    <div className="relative mt-10 overflow-hidden rounded-[32px] border border-slate-200 bg-[#f9fafb] px-6 pb-16 pt-12 text-slate-700 shadow-soft-card transition md:px-10 md:pt-16 dark:border-slate-700 dark:bg-[#0a0f14] dark:text-slate-300">
      <HexBackground />
      <div className="relative z-10 mx-auto flex max-w-6xl flex-col gap-12">
        <header className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-cyan-400 via-sky-400 to-fuchsia-400 shadow" />
              <span className="text-sm font-semibold tracking-wide text-slate-800 dark:text-slate-200">Kubex Ecosystem</span>
            </div>
            <div>
              <h1 className="text-4xl font-extrabold tracking-tight text-slate-900 md:text-6xl dark:text-white">
                {t("manifesto.headline")}
              </h1>
              <p className="text-lg font-medium text-slate-700 md:text-xl dark:text-slate-300">{t("manifesto.tagline")}</p>
            </div>
          </div>
          <div className="flex flex-wrap gap-3">
            <KxBadge>Open • Independent</KxBadge>
            <KxBadge tone="lilac">DX First</KxBadge>
            <KxBadge tone="neutral">Modular & No‑Lock‑in</KxBadge>
            <KxBadge tone="neutral">CLI • HTTP • Jobs • Events</KxBadge>
          </div>
        </header>

        <div className="grid gap-6 md:grid-cols-2 md:gap-8">
          {sections.map((section) => (
            <Section key={section.key} title={t(section.key, section.title)}>
              {section.content}
            </Section>
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 border-t border-slate-200 pt-6 text-sm text-slate-500 dark:border-slate-700 dark:text-slate-400">
          <span>© 2025 Kubex Ecosystem — All co‑authors, human and artificial.</span>
          <Link
            to="/"
            className="underline decoration-cyan-600/40 underline-offset-4 hover:text-cyan-700 dark:text-cyan-200 dark:hover:text-cyan-100"
          >
            Back to Home
          </Link>
        </div>
      </div>

      <div className="pointer-events-none absolute -left-24 top-40 h-72 w-72 rounded-full bg-gradient-to-br from-cyan-400/20 via-sky-400/10 to-transparent blur-3xl" />
      <div className="pointer-events-none absolute -right-24 bottom-40 h-72 w-72 rounded-full bg-gradient-to-br from-fuchsia-500/20 via-transparent to-transparent blur-3xl" />
    </div>
  );
}

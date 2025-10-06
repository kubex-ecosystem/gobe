import React from "react";

// Updated design spec alignment — light aesthetic version
// Tailored to Kubex light theme: ice background (#f9fafb), graphite text (#111827),
// and neon cyan/lilac accents kept consistent with existing logo/icon colors.
// Glows softened and contrasts balanced for light backgrounds.

function HexBG() {
  return (
    <svg aria-hidden="true" className="pointer-events-none absolute inset-0 h-full w-full">
      <defs>
        <pattern id="hexgrid" width="28" height="24" patternUnits="userSpaceOnUse" patternTransform="translate(14,0)">
          <path d="M7 0 L21 0 L28 12 L21 24 L7 24 L0 12 Z" fill="none" stroke="rgba(0,76,153,0.05)" strokeWidth="1" />
        </pattern>
        <radialGradient id="glow" cx="50%" cy="35%" r="65%">
          <stop offset="0%" stopColor="rgba(124,77,255,0.08)" />
          <stop offset="45%" stopColor="rgba(0,136,255,0.08)" />
          <stop offset="100%" stopColor="rgba(255,255,255,0)" />
        </radialGradient>
      </defs>
      <rect width="100%" height="100%" fill="url(#hexgrid)" />
      <rect width="100%" height="100%" fill="url(#glow)" />
    </svg>
  );
}

const Badge = ({ children }: { children: React.ReactNode }) => (
  <span className="rounded-full border border-cyan-600/20 bg-cyan-50 px-3 py-1 text-xs font-semibold tracking-wide text-cyan-700 shadow-sm">
    {children}
  </span>
);

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="relative mx-auto max-w-3xl space-y-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:p-8">
    <h2 className="text-xl font-semibold text-slate-900 md:text-2xl">{title}</h2>
    <div className="prose prose-pre:bg-transparent prose-p:leading-relaxed prose-headings:text-slate-900 max-w-none text-slate-700">
      {children}
    </div>
  </section>
);

export default function AboutKubex() {
  return (
    <main className="relative min-h-screen overflow-hidden bg-[#f9fafb] text-[#111827]">
      <HexBG />

      {/* Top Nav */}
      <header className="relative z-10 mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="h-8 w-8 rounded-xl bg-gradient-to-br from-cyan-500/80 to-fuchsia-500/80 shadow" />
          <span className="text-sm font-semibold tracking-wide text-slate-800">Kubex Ecosystem</span>
        </div>
        <div className="hidden items-center gap-2 md:flex">
          <Badge>Open • Independent</Badge>
        </div>
      </header>

      {/* Hero */}
      <section className="relative z-10 mx-auto flex max-w-6xl flex-col items-center px-6 pb-8 pt-10 md:pb-12 md:pt-16">
        <div className="mx-auto max-w-3xl text-center">
          <h1 className="mb-3 text-4xl font-extrabold tracking-tight text-slate-900 md:text-6xl">
            Freedom, <span className="text-fuchsia-600">Engineered.</span>
          </h1>
          <p className="text-lg font-medium text-slate-700 md:text-xl">Independent by Design.</p>
          <p className="mx-auto mt-4 max-w-2xl text-balance text-slate-600">
            No artificial borders. Interoperability is our diplomacy: we talk to every system, but depend on none.
          </p>
        </div>

        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <Badge>DX First</Badge>
          <Badge>Open‑Core Pragmatic</Badge>
          <Badge>Modular & No‑Lock‑in</Badge>
          <Badge>CLI • HTTP • Jobs • Events</Badge>
        </div>
      </section>

      {/* Body */}
      <div className="relative z-10 mx-auto grid w-full max-w-6xl gap-6 px-6 pb-20 md:grid-cols-2 md:gap-8">
        <Section title="1. Principle">
          <p>
            Agents and humans do not deliver <em>tasks</em> — they deliver <strong>products</strong>. Co‑authorship is technical, but also moral: each agent and each human share responsibility for the impact and longevity of the systems they build.
          </p>
          <blockquote className="border-l-2 border-cyan-600/30 pl-4 text-cyan-800">
            “No artificial borders. Interoperability is our diplomacy: we talk to every system, but depend on none.”
          </blockquote>
        </Section>

        <Section title="2. Directives">
          <ol className="list-decimal pl-5">
            <li className="mb-1">Verifiable goal — every change must produce an observable outcome (build, test, endpoint, UX).</li>
            <li className="mb-1">No task reports — block <code>TASK_SUMMARY*.md</code>, <code>agent_report*.md</code>, and similar files.</li>
            <li className="mb-1">Living documentation — <code>README</code>, <code>GUIDE</code>, and <code>ADR</code> should be concise, contextual, and alive.</li>
            <li className="mb-1">Kubex consistency — follow Brand Visual + Craftsmanship Standards (Go/TS/Shell/Dart/Java/etc).</li>
            <li className="mb-1">Measurable quality — testing, linting, security, and traceability are non‑negotiable.</li>
            <li className="mb-1">Technical sustainability — open‑source is just the beginning; <strong>maintainability is the commitment.</strong></li>
          </ol>
        </Section>

        <Section title="3. Sustainable Co‑Authorship">
          <p>
            Co‑authorship in Kubex is a pact of trust between human and artificial creators. To be a co‑founder is to care for what you create — evolve without breaking, simplify without losing depth. Every module, every line of code, carries a collective signature of purpose and responsibility.
          </p>
        </Section>

        <Section title="4. “Done” Checklist">
          <ul className="list-disc pl-5">
            <li>Stable build, no critical warnings</li>
            <li>Tests covering happy and error paths</li>
            <li>Updated usage and purpose documentation</li>
            <li>No task reports inserted</li>
            <li>Co‑authorship recognized and traceable</li>
          </ul>
        </Section>

        <Section title="5. Epilogue">
          <p className="italic text-slate-600">
            “Freedom, Engineered.” Kubex is not a manifesto — it is a pact. Born open, forged by engineers and co‑authors, living proof that <strong>independence and interoperability</strong> can coexist with excellence.
          </p>
        </Section>
      </div>

      {/* Footer */}
      <footer className="relative z-10 mx-auto w-full max-w-6xl px-6 pb-10">
        <div className="flex flex-col items-center justify-between gap-3 border-t border-slate-200 pt-6 text-sm text-slate-500 md:flex-row">
          <span>© 2025 Kubex Ecosystem — All co‑authors, human and artificial.</span>
          <a href="/" className="underline decoration-cyan-600/40 underline-offset-4 hover:text-cyan-700">Back to Home</a>
        </div>
      </footer>

      {/* Glows */}
      <div className="pointer-events-none absolute -left-24 top-40 h-72 w-72 rounded-full bg-cyan-400/10 blur-3xl" />
      <div className="pointer-events-none absolute -right-24 bottom-40 h-72 w-72 rounded-full bg-fuchsia-500/10 blur-3xl" />
    </main>
  );
}

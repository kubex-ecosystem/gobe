import * as React from "react";

import PageShell from "./components/layout/PageShell";
import { ThemeProvider } from "./context/theme";
import { I18nProvider } from "./i18n/provider";
import { Link, RouteObject, RouterProvider, useRouteMatch, useRouter } from "./lib/router";
import { AccessPage } from "./pages/Access";
import { LandingPage } from "./pages/Landing";
import { ManifestoPage } from "./pages/Manifesto";

const routes: RouteObject[] = [
  { path: "/", element: <PageShell><LandingPage /></PageShell> },
  { path: "/app/", element: <PageShell><LandingPage /></PageShell> },
  { path: "/app/access", element: <PageShell><AccessPage /></PageShell> },
  { path: "/app/auth", element: <PageShell><AccessPage /></PageShell> },
  { path: "/app/manifesto", element: <PageShell withBackgroundPattern={false}><ManifestoPage /></PageShell> },
  { path: "/app/about", element: <PageShell withBackgroundPattern={false}><ManifestoPage /></PageShell> },
];

function NotFound() {
  const { pathname } = useRouter();
  return (
    <PageShell>
      <div className="mx-auto flex min-h-[60vh] max-w-2xl flex-col items-center justify-center text-center">
        <span className="mb-3 text-xs font-semibold uppercase tracking-[0.3em] text-cyan-600/80">404</span>
        <h1 className="text-3xl font-semibold text-slate-900 dark:text-white">Page not found</h1>
        <p className="mt-4 text-slate-600">
          We could not find <span className="font-semibold">{pathname}</span>. Choose a destination below to continue building.
        </p>
        <div className="mt-8 flex gap-3">
          <Link
            to="/"
            className="rounded-full bg-primary px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary.hover"
          >
            Back to Home
          </Link>
          <Link
            to="/manifesto"
            className="rounded-full border border-slate-200 px-5 py-2 text-sm font-semibold text-slate-700 transition hover:border-primary/50 hover:text-primary dark:border-slate-700 dark:text-slate-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-200"
          >
            Manifesto Kubex
          </Link>
        </div>
      </div>
    </PageShell>
  );
}

function AppRoutes() {
  const element = useRouteMatch(routes);
  return element ?? <NotFound />;
}

export default function App() {
  return (
    <I18nProvider>
      <ThemeProvider>
        <RouterProvider>
          <AppRoutes />
        </RouterProvider>
      </ThemeProvider>
    </I18nProvider>
  );
}

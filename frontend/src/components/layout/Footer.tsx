import * as React from "react";

import { KxBadge } from "../ui/KxBadge";

export default function Footer() {
  return (
    <footer className="border-t border-slate-200/70 bg-white py-10 dark:border-slate-800 dark:bg-slate-950">
      <div className="mx-auto flex max-w-6xl flex-col gap-6 px-4 text-sm text-slate-500 dark:text-slate-400 sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div>
          <p className="font-semibold text-slate-600 dark:text-slate-200">© {new Date().getFullYear()} Kubex Ecosystem</p>
          <p className="mt-1 max-w-md text-slate-500 dark:text-slate-400">
            Co-authored by humans and agents. Independent by design, interoperable by choice.
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <KxBadge>Open • Independent</KxBadge>
          <KxBadge tone="lilac">DX First</KxBadge>
        </div>
      </div>
    </footer>
  );
}

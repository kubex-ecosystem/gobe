import * as React from "react";

import { cn } from "../../lib/cn";

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: "cyan" | "lilac" | "neutral";
}

const toneStyles: Record<NonNullable<BadgeProps["tone"]>, string> = {
  cyan: "border-cyan-600/20 bg-cyan-50 text-cyan-700 dark:border-cyan-500/30 dark:bg-cyan-500/10 dark:text-cyan-200",
  lilac: "border-fuchsia-500/20 bg-fuchsia-50 text-fuchsia-700 dark:border-fuchsia-500/30 dark:bg-fuchsia-500/10 dark:text-fuchsia-200",
  neutral: "border-slate-200 bg-white text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200",
};

export function Badge({ className, tone = "cyan", ...rest }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center justify-center rounded-full border px-3 py-1 text-xs font-semibold tracking-wide shadow-sm",
        toneStyles[tone],
        className
      )}
      {...rest}
    />
  );
}

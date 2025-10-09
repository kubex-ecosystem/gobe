import * as React from "react";

import { cn } from "../../lib/cn";

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Card({ className, ...rest }: CardProps) {
  return (
    <div
      className={cn(
        "rounded-3xl border border-slate-200 bg-white/95 p-6 shadow-sm backdrop-blur transition hover:-translate-y-1 hover:shadow-soft-card dark:border-slate-700 dark:bg-slate-900/80 dark:hover:shadow-[0_18px_40px_-28px_rgba(15,23,42,0.65)]",
        className
      )}
      {...rest}
    />
  );
}

interface CardHeaderProps extends React.HTMLAttributes<HTMLDivElement> {}

export function CardHeader({ className, ...rest }: CardHeaderProps) {
  return <div className={cn("mb-4 flex flex-col gap-2", className)} {...rest} />;
}

interface CardTitleProps extends React.HTMLAttributes<HTMLHeadingElement> {}

export function CardTitle({ className, ...rest }: CardTitleProps) {
  return <h3 className={cn("text-lg font-semibold text-slate-900 dark:text-slate-100", className)} {...rest} />;
}

interface CardDescriptionProps extends React.HTMLAttributes<HTMLParagraphElement> {}

export function CardDescription({ className, ...rest }: CardDescriptionProps) {
  return <p className={cn("text-sm leading-relaxed text-slate-600 dark:text-slate-300", className)} {...rest} />;
}

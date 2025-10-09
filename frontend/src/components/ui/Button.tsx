import * as React from "react";

import { cn } from "../../lib/cn";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost";
  size?: "md" | "lg" | "sm";
}

const baseStyles =
  "inline-flex items-center justify-center font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-cyan-400";

const variantStyles: Record<NonNullable<ButtonProps["variant"]>, string> = {
  primary: "rounded-full bg-primary px-5 py-2 text-white shadow-sm hover:bg-primary.hover dark:bg-cyan-500 dark:hover:bg-cyan-400",
  secondary:
    "rounded-full border border-slate-200 bg-white px-5 py-2 text-slate-700 shadow-sm hover:border-primary/50 hover:text-primary dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-cyan-400/60 dark:hover:text-cyan-300",
  ghost: "rounded-full px-4 py-2 text-slate-700 hover:text-primary dark:text-slate-300 dark:hover:text-cyan-300",
};

const sizeStyles: Record<NonNullable<ButtonProps["size"]>, string> = {
  sm: "text-xs",
  md: "text-sm",
  lg: "text-base",
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = "primary", size = "md", type = "button", ...rest },
  ref
) {
  return (
    <button
      ref={ref}
      type={type}
      className={cn(baseStyles, variantStyles[variant], sizeStyles[size], className)}
      {...rest}
    />
  );
});

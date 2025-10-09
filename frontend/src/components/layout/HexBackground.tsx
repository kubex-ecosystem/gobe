import * as React from "react";

import { cn } from "../../lib/cn";

interface HexBackgroundProps {
  className?: string;
  glow?: "cyan" | "fuchsia" | "both" | "none";
}

export function HexBackground({ className, glow = "both" }: HexBackgroundProps) {
  return (
    <div
      aria-hidden
      className={cn(
        "pointer-events-none absolute inset-0 -z-10 overflow-hidden",
        glow === "cyan" && "hex-bg hex-bg--cyan",
        glow === "fuchsia" && "hex-bg hex-bg--fuchsia",
        glow === "both" && "hex-bg hex-bg--both",
        glow === "none" && "hex-bg",
        className
      )}
    />
  );
}

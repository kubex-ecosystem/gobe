import * as React from "react";

import Footer from "./Footer";
import { HexBackground } from "./HexBackground";
import { Navbar } from "./Navbar";

interface PageShellProps {
  children: React.ReactNode;
  withBackgroundPattern?: boolean;
}

export default function PageShell({ children, withBackgroundPattern = true }: PageShellProps) {
  return (
    <div className="min-h-screen bg-background text-text-body transition-colors duration-300 dark:bg-[#0a0f14] dark:text-slate-200">
      <Navbar />
      <div className="relative isolate overflow-hidden">
        {withBackgroundPattern && <HexBackground />}
        {children}
      </div>
      <Footer />
    </div>
  );
}

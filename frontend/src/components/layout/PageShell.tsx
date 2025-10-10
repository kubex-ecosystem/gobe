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
    <div className="wh-max bg-background text-text-body transition-colors duration-300 dark:bg-[#0a0f14] dark:text-slate-200">
      <Navbar />
      <div className="relative isolate overflow-hidden h-fit min-h-[calc(90vh-7rem)] px-4 py-8 sm:px-6 lg:px-8">
        {withBackgroundPattern && <HexBackground />}
        {children}
      </div>
      <Footer />
    </div>
  );
}

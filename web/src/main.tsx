import * as React from "react";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import App from "./App";
// @ts-ignore: allow side-effect css import in TypeScript (handled by bundler)
import "./index.css";

const container = document.getElementById("root");

if (!container) {
  throw new Error("Root element '#root' not found");
}

const root = createRoot(container);

root.render(
  <StrictMode>
    <App />
  </StrictMode>
);

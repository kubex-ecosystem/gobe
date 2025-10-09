interface RuntimeConfig {
  apiBase: string;
  apiToken?: string;
  appBasePath: string;
}

declare global {
  interface Window {
    __KUBEX_CONFIG__?: Partial<RuntimeConfig>;
    __KUBEX_API_TOKEN__?: string;
    __KUBEX_API_BASE__?: string;
  }
}

const fallbackConfig: RuntimeConfig = {
  apiBase: import.meta.env.VITE_KUBEX_API_BASE || "",
  apiToken: import.meta.env.VITE_KUBEX_API_TOKEN || undefined,
  appBasePath: import.meta.env.VITE_KUBEX_APP_BASE || "/",
};

function normalizeBasePath(path: string) {
  if (!path) return "/";
  return path.endsWith("/") ? path.slice(0, -1) : path;
}

function resolveRuntimeConfig(): RuntimeConfig {
  if (typeof window === "undefined") {
    return {
      ...fallbackConfig,
      appBasePath: normalizeBasePath(fallbackConfig.appBasePath),
    };
  }

  const windowConfig = window.__KUBEX_CONFIG__ ?? {};

  const apiBase = window.__KUBEX_API_BASE__ ?? windowConfig.apiBase ?? fallbackConfig.apiBase ?? "";
  const apiToken = window.__KUBEX_API_TOKEN__ ?? windowConfig.apiToken ?? fallbackConfig.apiToken;
  const appBasePath = windowConfig.appBasePath ?? fallbackConfig.appBasePath ?? "/";

  return {
    apiBase,
    apiToken,
    appBasePath: normalizeBasePath(appBasePath),
  };
}

const runtimeConfig = resolveRuntimeConfig();

export function getApiBase() {
  return runtimeConfig.apiBase;
}

export function getAuthToken() {
  if (typeof window !== "undefined") {
    const stored = window.localStorage?.getItem("kubex:apiToken");
    if (stored) {
      return stored;
    }
  }
  return runtimeConfig.apiToken;
}

export function getAppBasePath() {
  return runtimeConfig.appBasePath;
}

export function buildApiUrl(path: string) {
  const base = getApiBase();
  if (!path.startsWith("/")) {
    path = `/${path}`;
  }
  if (!base) {
    return path;
  }
  return `${base.replace(/\/$/, "")}${path}`;
}

export function withAppPath(path = "/") {
  const base = getAppBasePath();
  const normalized = path.startsWith("/") ? path : `/${path}`;
  if (normalized === "/") {
    return `${base}/`;
  }
  return `${base}${normalized}`;
}

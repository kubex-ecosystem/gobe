import { buildApiUrl, getAuthToken } from "./config";
import { getAuthState } from "../state/authState";

export interface ApiRequestOptions {
  path: string;
  method?: string;
  body?: unknown;
  signal?: AbortSignal;
  headers?: Record<string, string>;
  auth?: boolean;
}

export async function fetchJson<T>({
  path,
  method = "GET",
  body,
  signal,
  headers = {},
  auth = false,
}: ApiRequestOptions): Promise<T> {
  const url = buildApiUrl(path);
  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };

  if (body && !(body instanceof FormData)) {
    finalHeaders["Content-Type"] = "application/json";
  }

  if (auth) {
    let token = getAuthState().accessToken;
    if (!token && typeof window !== "undefined") {
      token = window.sessionStorage.getItem("kubex:apiToken");
    }
    if (!token) {
      token = getAuthToken();
    }
    if (token) {
      finalHeaders["Authorization"] = `Bearer ${token}`;
    }
  }

  const response = await fetch(url, {
    method,
    headers: finalHeaders,
    body: body && !(body instanceof FormData) ? JSON.stringify(body) : (body as BodyInit | null | undefined),
    signal,
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error(`Request to ${url} failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return {} as T;
  }

  const contentType = response.headers.get("Content-Type") ?? "";
  if (contentType.includes("application/json")) {
    return (await response.json()) as T;
  }

  const text = await response.text();
  return text as unknown as T;
}

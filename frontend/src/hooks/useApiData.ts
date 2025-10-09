import * as React from "react";

import { fetchJson, ApiRequestOptions } from "../lib/api";

export interface UseApiDataOptions<T> extends Omit<ApiRequestOptions, "path"> {
  path: string;
  enabled?: boolean;
  transform?: (data: unknown) => T;
  fallbackData?: T;
  pollingInterval?: number; // milliseconds - 0 or undefined means no polling
}

export interface ApiDataState<T> {
  data?: T;
  error?: Error;
  loading: boolean;
  refetch: () => Promise<void>;
}

export function useApiData<T = unknown>({
  path,
  enabled = true,
  transform,
  fallbackData,
  auth,
  method,
  body,
  headers,
  pollingInterval,
}: UseApiDataOptions<T>): ApiDataState<T> {
  const [data, setData] = React.useState<T | undefined>(fallbackData);
  const [error, setError] = React.useState<Error | undefined>();
  const [loading, setLoading] = React.useState<boolean>(enabled);

  const fetchRef = React.useRef<AbortController | null>(null);
  const pollingRef = React.useRef<number | null>(null);
  const isMountedRef = React.useRef(true);

  // Store latest values in refs to avoid recreating functions
  const pathRef = React.useRef(path);
  const enabledRef = React.useRef(enabled);
  const transformRef = React.useRef(transform);
  const authRef = React.useRef(auth);
  const methodRef = React.useRef(method);
  const bodyRef = React.useRef(body);
  const headersRef = React.useRef(headers);

  // Update refs when values change (doesn't trigger re-renders)
  React.useEffect(() => {
    pathRef.current = path;
    enabledRef.current = enabled;
    transformRef.current = transform;
    authRef.current = auth;
    methodRef.current = method;
    bodyRef.current = body;
    headersRef.current = headers;
  });

  // Stable fetch function that doesn't change
  const runFetch = React.useCallback(async (isInitial = false) => {
    if (!enabledRef.current || !isMountedRef.current) {
      return;
    }

    // Cancel previous request
    if (fetchRef.current) {
      fetchRef.current.abort();
    }

    const controller = new AbortController();
    fetchRef.current = controller;

    if (isInitial) {
      setLoading(true);
    }
    setError(undefined);

    try {
      const raw = await fetchJson<unknown>({
        path: pathRef.current,
        auth: authRef.current,
        method: methodRef.current,
        body: bodyRef.current,
        headers: headersRef.current,
        signal: controller.signal,
      });

      if (!isMountedRef.current) return;

      const nextData = transformRef.current ? transformRef.current(raw) : (raw as T);
      setData(nextData);
    } catch (err) {
      if ((err as Error).name === "AbortError") {
        return;
      }
      if (isMountedRef.current) {
        setError(err as Error);
      }
    } finally {
      if (isMountedRef.current && isInitial) {
        setLoading(false);
      }
    }
  }, []); // NO DEPENDENCIES - stable function

  // Initial fetch - only runs when path, enabled, or pollingInterval change
  React.useEffect(() => {
    console.log('[useApiData] Effect triggered', { path, enabled, pollingInterval });

    if (!enabled) {
      // Clear any existing polling when disabled
      if (pollingRef.current) {
        console.log('[useApiData] Clearing interval (disabled)');
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
      return;
    }

    isMountedRef.current = true;

    // Initial fetch
    console.log('[useApiData] Initial fetch for:', path);
    runFetch(true);

    // Clear any existing polling before setting up new one
    if (pollingRef.current) {
      console.log('[useApiData] Clearing existing interval before new setup');
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }

    // Setup polling if interval is provided and > 0
    if (pollingInterval && pollingInterval > 0) {
      console.log('[useApiData] Setting up polling:', pollingInterval, 'ms for', path);
      pollingRef.current = window.setInterval(() => {
        console.log('[useApiData] Polling tick for:', path);
        runFetch(false);
      }, pollingInterval);
    }

    return () => {
      console.log('[useApiData] Cleanup for:', path);
      isMountedRef.current = false;
      fetchRef.current?.abort();
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, enabled, pollingInterval]); // runFetch is stable (empty deps), no need to include it

  // Manual refetch function
  const refetch = React.useCallback(async () => {
    await runFetch(true);
  }, [runFetch]);

  return {
    data,
    error,
    loading,
    refetch,
  };
}

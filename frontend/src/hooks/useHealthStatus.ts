import * as React from "react";

import { useApiData } from "./useApiData";
import { useAuthToken } from "./useAuthToken";

const _timeout = (ms: number, cb: () => void) => new Promise((resolve) => setTimeout(() => { cb(); resolve(true); }, ms));

interface GatewayServiceHealth {
  healthy: boolean;
  detail?: {
    total?: number;
    available?: number;
    unavailable?: number;
  };
}

interface GatewayHealthResponse {
  status: string;
  timestamp?: string;
  uptime?: string;
  started?: string;
  version?: string;
  services?: Record<string, GatewayServiceHealth>;
}

interface BasicHealthResponse {
  status: string;
  message?: string;
}

export interface HealthSummary {
  status: string;
  uptime?: string;
  version?: string;
  services: Record<string, GatewayServiceHealth>;
  source: "basic" | "detailed";
}

const DEFAULT_SUMMARY: HealthSummary = {
  status: "unknown",
  services: {},
  source: "basic",
};

function transformGatewayHealth(data: GatewayHealthResponse): HealthSummary {
  return {
    status: data.status ?? "unknown",
    uptime: data.uptime,
    version: data.version,
    services: data.services ?? {},
    source: "detailed",
  };
}

function transformBasicHealth(data: BasicHealthResponse): HealthSummary {
  return {
    status: data.status ?? data.message ?? "unknown",
    services: {},
    source: "basic",
  };
}

export function useHealthStatus(pollingInterval = 30000) {
  const token = useAuthToken();
  const wantsDetailed = Boolean(token);
  const detailed = useApiData<HealthSummary>({
    path: "/api/v1/health",
    enabled: wantsDetailed,
    auth: true,
    transform: (raw) => transformGatewayHealth(raw as GatewayHealthResponse),
    pollingInterval: wantsDetailed ? pollingInterval : 0,
  });

  const basic = useApiData<HealthSummary>({
    path: "/health",
    enabled: !wantsDetailed,
    transform: (raw) => transformBasicHealth(raw as BasicHealthResponse),
    pollingInterval: !wantsDetailed ? pollingInterval : 0,
  });

  const loading = detailed.loading || basic.loading;
  const error = detailed.error ?? basic.error;
  const data = detailed.data ?? basic.data ?? DEFAULT_SUMMARY;

  const refresh = React.useCallback(async () => {
    if (wantsDetailed) {
      await detailed.refetch();
    } else {
      await basic.refetch();
    }
  }, [basic, detailed, wantsDetailed]);

  return { data, loading, error, refresh };
}

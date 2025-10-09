import { useApiData } from "./useApiData";
import { useAuthToken } from "./useAuthToken";

export interface ProviderItem {
  name: string;
  type: string;
  org?: string;
  defaultModel?: string;
  available: boolean;
  lastError?: string;
  metadata?: Record<string, unknown>;
}

interface ProvidersResponse {
  providers: ProviderItem[];
  timestamp?: string;
}

export function useProvidersStatus(pollingInterval?: number) {
  const token = useAuthToken();
  const enabled = Boolean(token);

  const result = useApiData<ProvidersResponse>({
    path: "/providers",
    auth: true,
    enabled,
    pollingInterval,
  });

  return {
    data: result.data,
    loading: result.loading,
    error: result.error,
    refetch: result.refetch,
  };
}

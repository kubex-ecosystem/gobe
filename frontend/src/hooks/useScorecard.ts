import { useApiData } from "./useApiData";
import { useAuthToken } from "./useAuthToken";

export interface ScorecardEntry {
  id?: string;
  title: string;
  description: string;
  score?: number;
  status?: string;
  completedAt?: string;
  repoUrl?: string;
}

interface ScorecardResponse {
  items: ScorecardEntry[];
  total: number;
  version?: string;
}

export function useScorecard(limit = 6, pollingInterval?: number) {
  const token = useAuthToken();
  const enabled = Boolean(token);

  const result = useApiData<ScorecardResponse>({
    path: `/api/v1/scorecard?limit=${limit}`,
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

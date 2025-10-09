import * as React from "react";

import { getAuthToken } from "../lib/config";

export function useAuthToken() {
  const [token, setToken] = React.useState<string | undefined>(() => getAuthToken());

  React.useEffect(() => {
    const syncToken = () => {
      setToken(getAuthToken());
    };

    window.addEventListener("storage", syncToken);
    return () => window.removeEventListener("storage", syncToken);
  }, []);

  return token;
}

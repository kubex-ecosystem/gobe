import * as React from "react";

import { setAccessToken as setAccessTokenState } from "../state/authState";

interface AuthState {
  accessToken: string | null;
  user: UserSummary | null;
  setAuth: (token: string | null, user: UserSummary | null) => void;
  clearAuth: () => void;
}

interface UserSummary {
  id: string;
  username: string;
  email: string;
  name: string;
  role: string;
  active: boolean;
}

const AuthContext = React.createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [accessToken, setAccessTokenValue] = React.useState<string | null>(() => {
    if (typeof window === "undefined") return null;
    const token = window.sessionStorage.getItem("kubex:apiToken");
    if (token) {
      setAccessTokenState(token);
    }
    return token;
  });
  const [user, setUser] = React.useState<UserSummary | null>(() => {
    if (typeof window === "undefined") return null;
    const stored = window.sessionStorage.getItem("kubex:user");
    return stored ? (JSON.parse(stored) as UserSummary) : null;
  });

  const setAuth = React.useCallback((token: string | null, summary: UserSummary | null) => {
    setAccessTokenValue(token);
    setAccessTokenState(token);
    setUser(summary);
    if (typeof window !== "undefined") {
      if (token) {
        window.sessionStorage.setItem("kubex:apiToken", token);
        window.localStorage.removeItem("kubex:apiToken");
      } else {
        window.sessionStorage.removeItem("kubex:apiToken");
        window.localStorage.removeItem("kubex:apiToken");
      }
      if (summary) {
        window.sessionStorage.setItem("kubex:user", JSON.stringify(summary));
        window.localStorage.removeItem("kubex:user");
      } else {
        window.sessionStorage.removeItem("kubex:user");
        window.localStorage.removeItem("kubex:user");
      }
    }
  }, []);

  const clearAuth = React.useCallback(() => {
    setAuth(null, null);
    if (typeof window !== "undefined") {
      window.sessionStorage.removeItem("kubex:refreshToken");
      window.localStorage.removeItem("kubex:apiToken");
      window.localStorage.removeItem("kubex:user");
    }
  }, [setAuth]);

  const value = React.useMemo<AuthState>(() => ({ accessToken, user, setAuth, clearAuth }), [accessToken, user, setAuth, clearAuth]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

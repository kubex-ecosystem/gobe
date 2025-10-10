import * as React from "react";

import { useAuth } from "../../context/auth";
import { fetchJson } from "../../lib/api";
import { Button } from "../../components/ui/Button";

interface UserSummary {
  id: string;
  username: string;
  email: string;
  name: string;
  role: string;
  active: boolean;
}

interface UserListResponse {
  users: UserSummary[];
}

interface LogoutResponse {
  message: string;
}

export function DashboardPage() {
  const { user, clearAuth } = useAuth();
  const [users, setUsers] = React.useState<UserSummary[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const response = await fetchJson<UserListResponse>({
          path: "/users",
          auth: true,
        });
        if (mounted) {
          setUsers(response.users);
        }
      } catch (err) {
        if (mounted) {
          const message = err instanceof Error ? err.message : "Failed to load users";
          if (message.includes("status 401")) {
            clearAuth();
            if (typeof window !== "undefined") {
              window.location.href = "/app/access";
            }
            return;
          }
          setError(message);
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

  const handleLogout = React.useCallback(async () => {
    try {
      const refresh = typeof window !== "undefined" ? window.sessionStorage.getItem("kubex:refreshToken") : null;
      await fetchJson<LogoutResponse>(
        {
          path: "/api/v1/sign-out",
          method: "POST",
          auth: true,
          body: refresh ? { refresh_token: refresh } : undefined,
        }
      );
    } catch (err) {
      // ignore errors on sign-out
    } finally {
      if (typeof window !== "undefined") {
        window.sessionStorage.removeItem("kubex:refreshToken");
        window.sessionStorage.removeItem("kubex:apiToken");
        window.sessionStorage.removeItem("kubex:user");
      }
      clearAuth();
      if (typeof window !== "undefined") {
        window.location.href = "/app/access";
      }
    }
  }, [clearAuth]);

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.3em] text-cyan-600/80 dark:text-cyan-300/80">GoBE Dashboard</p>
          <h1 className="text-3xl font-semibold text-slate-900 dark:text-white">Welcome{user ? `, ${user.name}` : ""}</h1>
          <p className="text-sm text-slate-600 dark:text-slate-300">Manage Kubex identities and review active users.</p>
        </div>
        <Button onClick={handleLogout} variant="secondary" className="self-start md:self-auto">
          Sign out
        </Button>
      </header>

      {loading ? (
        <p className="text-sm text-slate-500 dark:text-slate-400">Loading users…</p>
      ) : error ? (
        <p className="text-sm text-rose-600 dark:text-rose-300">{error}</p>
      ) : (
        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-700 dark:bg-slate-900">
          <table className="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
            <thead className="bg-slate-50/80 dark:bg-slate-800/60">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-300">Username</th>
                <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-300">Name</th>
                <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-300">Email</th>
                <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-300">Role</th>
                <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-300">Active</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200 dark:divide-slate-800">
              {users.map((item) => (
                <tr key={item.id} className="hover:bg-slate-50/70 dark:hover:bg-slate-800/50">
                  <td className="px-4 py-3 font-medium text-slate-700 dark:text-slate-200">{item.username}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{item.name}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{item.email}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{item.role}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{item.active ? "Yes" : "No"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

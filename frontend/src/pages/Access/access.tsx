import { CheckCircle2, Shield } from "lucide-react";
import * as React from "react";

import { AuthForm } from "../../components/modules/AuthForm";
import { Card, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card";
import { KxBadge } from "../../components/ui/KxBadge";
import { useAuth } from "../../context/auth";
import { useI18n } from "../../i18n/provider";
import { Link, useRouter } from "../../lib/router";

export function AccessPage() {
  const { t, get } = useI18n();
  const { navigate } = useRouter();
  const { accessToken } = useAuth();
  const bullets = get<string[]>("access.bullets", []);
  const [primaryBullet, ...secondaryBullets] = bullets;

  const handleComplete = React.useCallback((_identifier: string) => {
    window.setTimeout(() => navigate("/dashboard", { replace: true }), 900);
  }, [navigate]);

  React.useEffect(() => {
    if (accessToken) {
      navigate("/dashboard", { replace: true });
    }
  }, [accessToken, navigate]);

  return (
    <div className="relative mx-auto flex max-w-6xl flex-col-reverse gap-14 px-4 pb-24 pt-12 text-slate-700 transition-colors duration-300 sm:px-6 md:flex-row md:items-start md:pb-28 md:pt-16 dark:text-slate-300">
      <div className="md:w-1/2">
        <KxBadge tone="cyan">{t("access.badge", "Portal de Autenticação")}</KxBadge>
        <h1 className="mt-5 text-3xl font-semibold text-slate-900 dark:text-white">{t("access.headline")}</h1>
        <p className="mt-3 text-slate-600 dark:text-slate-300">{t("access.subtitle")}</p>

        <div className="mt-8 space-y-4">
          {primaryBullet && (
            <div className="flex items-start gap-3 rounded-2xl border border-cyan-600/15 bg-cyan-50/70 p-4 text-sm text-cyan-700 dark:border-cyan-500/30 dark:bg-slate-900/70 dark:text-cyan-200">
              <Shield className="mt-0.5 h-5 w-5" />
              <p>{primaryBullet}</p>
            </div>
          )}
          <ul className="space-y-3 text-sm text-slate-600 dark:text-slate-300">
            {secondaryBullets.map((item) => (
              <li key={item} className="flex items-start gap-2">
                <CheckCircle2 className="mt-0.5 h-4 w-4 text-cyan-600 dark:text-cyan-300" />
                {item}
              </li>
            ))}
          </ul>

          <Link
            to="/manifesto"
            className="inline-flex items-center gap-2 text-sm font-semibold text-cyan-700 transition hover:text-cyan-600 dark:text-cyan-200 dark:hover:text-cyan-100"
          >
            Manifesto Kubex
            <span aria-hidden>→</span>
          </Link>
        </div>
      </div>

      <Card className="md:w-1/2">
        <CardHeader>
          <KxBadge tone="neutral" className="self-start">Escolha o fluxo</KxBadge>
          <CardTitle>Autenticar coautores</CardTitle>
          <CardDescription>
            O GoBE valida as credenciais e aplica políticas definidas em <code>internal/module/module.go</code> e CLI relacionadas.
          </CardDescription>
        </CardHeader>
        <AuthForm onComplete={handleComplete} />
      </Card>
    </div>
  );
}

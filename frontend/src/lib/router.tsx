import * as React from "react";

interface RouterState {
  pathname: string;
  navigate: (to: string, options?: { replace?: boolean }) => void;
}

const RouterContext = React.createContext<RouterState | undefined>(undefined);

function getPathname() {
  return window.location.pathname || "/";
}

export function RouterProvider({ children }: { children: React.ReactNode }) {
  const [pathname, setPathname] = React.useState<string>(() => getPathname());

  React.useEffect(() => {
    const handleChange = () => setPathname(getPathname());
    window.addEventListener("popstate", handleChange);
    window.addEventListener("hashchange", handleChange);
    return () => {
      window.removeEventListener("popstate", handleChange);
      window.removeEventListener("hashchange", handleChange);
    };
  }, []);

  const navigate = React.useCallback((to: string, options?: { replace?: boolean }) => {
    const normalized = to.startsWith("/app/") ? to : `/app/${to}`;
    if (options?.replace) {
      window.history.replaceState(null, "", normalized);
    } else {
      window.history.pushState(null, "", normalized);
    }
    setPathname(normalized);
  }, []);

  const value = React.useMemo(
    () => ({ pathname, navigate }),
    [pathname, navigate]
  );

  return <RouterContext.Provider value={value}>{children}</RouterContext.Provider>;
}

export function useRouter() {
  const context = React.useContext(RouterContext);
  if (!context) {
    throw new Error("useRouter must be used within a RouterProvider");
  }
  return context;
}

export interface RouteObject {
  path: string;
  element: React.ReactNode;
}

export function useRouteMatch(routes: RouteObject[]) {
  const { pathname } = useRouter();
  const match = React.useMemo(() => routes.find((route) => route.path === pathname), [routes, pathname]);
  return match?.element;
}

export interface LinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  to: string;
  replace?: boolean;
}

export const Link = React.forwardRef<HTMLAnchorElement, LinkProps>(function Link(
  { to, replace, onClick, target, rel, ...rest },
  ref
) {
  const { navigate } = useRouter();
  const isExternal = target === "_blank" || /^https?:/i.test(to);

  const handleClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
    onClick?.(event);
    if (event.defaultPrevented) return;
    if (event.button !== 0 || event.metaKey || event.altKey || event.ctrlKey || event.shiftKey) return;
    if (isExternal) return;
    event.preventDefault();
    navigate(to, { replace });
  };

  return (
    <a
      ref={ref}
      href={to}
      target={target}
      rel={target === "_blank" ? rel ?? "noopener noreferrer" : rel}
      onClick={handleClick}
      {...rest}
    />
  );
});

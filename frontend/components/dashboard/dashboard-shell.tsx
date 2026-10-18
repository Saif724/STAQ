"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ComponentProps,
  type ReactNode,
} from "react";
import { useRouter } from "next/navigation";
import { Menu } from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
// import NotificationBell from "@/components/notifications/notification-bell";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

type DashboardShellProps = {
  activePage: ComponentProps<typeof Sidebar>["activePage"];
  title: string;
  subtitle?: string;
  actions?: ReactNode;
  children: ReactNode;
};

const UserContext = createContext<CurrentUser | null>(null);

export function useShellUser() {
  return useContext(UserContext);
}

export default function DashboardShell({
  activePage,
  title,
  subtitle,
  actions,
  children,
}: DashboardShellProps) {
  const router = useRouter();

  const [user, setUser] = useState<CurrentUser | null>(null);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadUser() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }

      try {
        const response = await getCurrentUser();
        if (!cancelled) setUser(response.data);
      } catch (error) {
        if (cancelled) return;

        if (error instanceof ApiError && error.status === 401) {
          clearTokens();
          router.replace("/login");
        }
      }
    }

    void loadUser();

    return () => {
      cancelled = true;
    };
  }, [router]);

  async function handleLogout() {
    setLoggingOut(true);

    try {
      await logoutUser();
    } catch {
      // The local session is cleared below even if the API request fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  return (
    <UserContext.Provider value={user}>
      <div className="min-h-dvh bg-background text-foreground">
        <Sidebar
          mobileOpen={mobileOpen}
          onClose={() => setMobileOpen(false)}
          user={user}
          onLogout={handleLogout}
          loggingOut={loggingOut}
          activePage={activePage}
        />

        <div className="min-h-dvh lg:pl-66">
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between gap-3 border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:h-19 lg:px-8">
            <div className="flex min-w-0 items-center gap-2">
              <button
                type="button"
                aria-label="Open navigation"
                onClick={() => setMobileOpen(true)}
                className="-ml-2 flex size-10 shrink-0 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted lg:hidden"
              >
                <Menu className="size-5" />
              </button>

              <div className="min-w-0">
                <p className="truncate text-sm font-semibold">{title}</p>
                {subtitle && (
                  <p className="hidden truncate text-xs text-muted-foreground sm:block">
                    {subtitle}
                  </p>
                )}
              </div>
            </div>

            <div className="flex shrink-0 items-center gap-1 sm:gap-2">
              {actions}
              {/* <NotificationBell /> */}
            </div>
          </header>

          <main className="mx-auto w-full max-w-375 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
            {children}
          </main>
        </div>
      </div>
    </UserContext.Provider>
  );
}
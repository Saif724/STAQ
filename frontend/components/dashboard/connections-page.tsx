"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AlertCircle, CheckCircle2, ExternalLink, Link2, RefreshCw, ShieldCheck, Trash2 } from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { deleteConnection, getConnections, getGmailAuthorizationUrl, type Connection } from "@/lib/api/connections";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

function formatDate(value: string | null | undefined) {
  if (!value) return "Expiry not provided";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Expiry unavailable";
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(date);
}

export default function ConnectionsPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const gmailResult = searchParams.get("gmail");
  const gmailReason = searchParams.get("reason");
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshVersion, setRefreshVersion] = useState(0);
  const [connecting, setConnecting] = useState(false);
  const [removingId, setRemovingId] = useState("");
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const callbackNotice = gmailResult === "connected" ? "Gmail was connected successfully." : "";
  const callbackError = gmailResult === "error"
    ? gmailReason === "denied"
      ? "Google authorization was cancelled."
      : gmailReason === "invalid_callback"
        ? "The Google authorization response was incomplete. Please try again."
        : "Gmail could not be connected. Check your account and try again."
    : "";

  useEffect(() => {
    let cancelled = false;

    async function loadConnections() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setLoading(true);
      setError("");
      try {
        const [userResponse, connectionsResponse] = await Promise.all([
          getCurrentUser(),
          getConnections(),
        ]);
        if (cancelled) return;
        setUser(userResponse.data);
        setConnections(connectionsResponse.data);
      } catch (caught) {
        if (cancelled) return;
        if (caught instanceof ApiError && caught.status === 401) {
          clearTokens();
          router.replace("/login");
          return;
        }
        setError(caught instanceof Error ? caught.message : "Could not load connections.");
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadConnections();
    return () => {
      cancelled = true;
    };
  }, [refreshVersion, router]);

  async function handleConnect() {
    setConnecting(true);
    setError("");
    try {
      const response = await getGmailAuthorizationUrl();
      const authorizationUrl = new URL(response.data.authorization_url);
      if (authorizationUrl.protocol !== "https:" && authorizationUrl.protocol !== "http:") {
        throw new Error("The Gmail authorization URL is invalid.");
      }
      window.location.assign(authorizationUrl.toString());
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 401) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setError(caught instanceof Error ? caught.message : "Could not start Gmail authorization.");
      setConnecting(false);
    }
  }

  async function handleDisconnect(connection: Connection) {
    if (!window.confirm(`Disconnect ${connection.account_email} from STAQ?`)) return;
    setRemovingId(connection.id);
    setError("");
    setNotice("");
    try {
      await deleteConnection(connection.id);
      setConnections((current) => current.filter((item) => item.id !== connection.id));
      setNotice(`${connection.account_email} was disconnected.`);
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 401) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setError(caught instanceof Error ? caught.message : "Could not disconnect this account.");
    } finally {
      setRemovingId("");
    }
  }

  async function handleLogout() {
    setLoggingOut(true);
    try {
      await logoutUser();
    } catch {
      // Clear the local session even if the API request fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Sidebar mobileOpen={mobileOpen} onClose={() => setMobileOpen(false)} user={user} onLogout={handleLogout} loggingOut={loggingOut} activePage="connections" />
      <main className="min-h-screen lg:pl-66">
        <header className="sticky top-0 z-30 flex h-19 items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button type="button" aria-label="Open navigation" onClick={() => setMobileOpen(true)} className="rounded-lg border border-border p-2 text-muted-foreground hover:bg-muted lg:hidden"><Link2 className="size-4" /></button>
            <div><p className="text-sm font-semibold">Connections</p><p className="text-xs text-muted-foreground">External accounts used by your automations</p></div>
          </div>
          <Button type="button" variant="outline" size="sm" disabled={loading} onClick={() => setRefreshVersion((version) => version + 1)} className="gap-2"><RefreshCw className="size-3.5" />Refresh</Button>
        </header>

        <div className="mx-auto w-full max-w-375 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div><p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Integrations</p><h1 className="mt-2 text-2xl font-bold">Connected accounts</h1><p className="mt-2 text-sm text-muted-foreground">Authorize Gmail for email actions that send from your account.</p></div>
            <Button type="button" disabled={connecting} onClick={handleConnect} className="gap-2"><ExternalLink className="size-4" />{connecting ? "Connecting…" : connections.some((item) => item.provider === "GOOGLE") ? "Reconnect Gmail" : "Connect Gmail"}</Button>
          </div>

          {(notice || callbackNotice) && <div role="status" className="mt-6 flex items-center gap-2 rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-3 text-sm text-emerald-700 dark:text-emerald-300"><CheckCircle2 className="size-4" />{notice || callbackNotice}</div>}
          {(error || callbackError) && <div role="alert" className="mt-6 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive"><AlertCircle className="size-4" />{error || callbackError}</div>}

          <section className="mt-7 border-y border-border/70">
            <div className="flex items-center justify-between gap-3 py-4">
              <div><h2 className="font-semibold">Gmail</h2><p className="mt-1 text-xs text-muted-foreground">Send email actions using an authorized Google account.</p></div>
              <span className="text-xs text-muted-foreground">{loading ? "Checking status…" : `${connections.length} connected account${connections.length === 1 ? "" : "s"}`}</span>
            </div>

            {loading ? (
              <div className="space-y-3 pb-5">{[1, 2].map((item) => <div key={item} className="h-16 animate-pulse rounded-lg bg-muted" />)}</div>
            ) : connections.length === 0 ? (
              <div className="flex flex-col items-center border-t border-border/70 px-5 py-14 text-center">
                <div className="flex size-12 items-center justify-center rounded-xl bg-muted text-muted-foreground"><Link2 className="size-5" /></div>
                <h3 className="mt-4 font-semibold">No accounts connected</h3>
                <p className="mt-1 max-w-md text-sm text-muted-foreground">Connect Gmail to use the USER_GMAIL sender in email actions. STAQ will never display your stored tokens.</p>
              </div>
            ) : (
              <div className="divide-y divide-border/70 border-t border-border/70">
                {connections.map((connection) => (
                  <article key={connection.id} className="flex flex-col gap-4 py-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex min-w-0 items-center gap-3">
                      <span className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"><CheckCircle2 className="size-5" /></span>
                      <div className="min-w-0"><p className="truncate text-sm font-semibold">{connection.account_email}</p><p className="mt-1 text-xs text-muted-foreground">Google · connected {formatDate(connection.created_at)}</p><p className="mt-1 text-xs text-muted-foreground">{connection.token_expires_at ? `Token expires ${formatDate(connection.token_expires_at)}` : "Token expiry managed by Google"}</p></div>
                    </div>
                    <Button type="button" variant="outline" size="sm" disabled={removingId === connection.id} onClick={() => void handleDisconnect(connection)} className="gap-2 self-start text-destructive hover:text-destructive sm:self-auto"><Trash2 className="size-3.5" />{removingId === connection.id ? "Disconnecting…" : "Disconnect"}</Button>
                  </article>
                ))}
              </div>
            )}
          </section>

          <p className="mt-5 flex items-center gap-2 text-xs text-muted-foreground"><ShieldCheck className="size-3.5" />Only connection metadata is shown here. Credentials remain encrypted by the backend.</p>
        </div>
      </main>
    </div>
  );
}
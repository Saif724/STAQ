"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { AlertCircle, ArrowLeft, Clock3, RefreshCw, ShieldCheck } from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { getExecution, getExecutionLogs, type Execution, type ExecutionLog, type ExecutionStatus } from "@/lib/api/automation";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

function formatDate(value: string | null | undefined) {
  if (!value) return "Not available";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Date unavailable";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(date);
}

function statusTone(status: ExecutionStatus) {
  switch (status) {
    case "SUCCESS":
      return "border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
    case "FAILED":
      return "border-destructive/20 bg-destructive/10 text-destructive";
    case "TIMED_OUT":
      return "border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300";
    case "RUNNING":
      return "border-sky-500/20 bg-sky-500/10 text-sky-700 dark:text-sky-300";
    default:
      return "border-border bg-muted text-muted-foreground";
  }
}

export default function ExecutionDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const executionId = params.id;
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [execution, setExecution] = useState<Execution | null>(null);
  const [logs, setLogs] = useState<ExecutionLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [refreshVersion, setRefreshVersion] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function loadExecution() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setLoading(true);
      setError("");
      try {
        const [userResponse, executionResponse, logsResponse] = await Promise.all([
          getCurrentUser(),
          getExecution(executionId),
          getExecutionLogs(executionId),
        ]);
        if (cancelled) return;
        setUser(userResponse.data);
        setExecution(executionResponse.data);
        setLogs(logsResponse.data);
      } catch (caught) {
        if (cancelled) return;
        if (caught instanceof ApiError && caught.status === 401) {
          clearTokens();
          router.replace("/login");
          return;
        }
        setError(caught instanceof Error ? caught.message : "Could not load this execution.");
      } finally {
        if (!cancelled) {
          setLoading(false);
          setRefreshing(false);
        }
      }
    }

    void loadExecution();
    return () => {
      cancelled = true;
    };
  }, [executionId, refreshVersion, router]);

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

  function refresh() {
    setRefreshing(true);
    setRefreshVersion((version) => version + 1);
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Sidebar
        mobileOpen={mobileOpen}
        onClose={() => setMobileOpen(false)}
        user={user}
        onLogout={handleLogout}
        loggingOut={loggingOut}
        activePage="executions"
      />

      <main className="min-h-screen lg:pl-66">
        <header className="sticky top-0 z-30 flex h-19 items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button type="button" aria-label="Open navigation" onClick={() => setMobileOpen(true)} className="rounded-lg border border-border p-2 text-muted-foreground hover:bg-muted lg:hidden">
              <Clock3 className="size-4" />
            </button>
            <div>
              <p className="text-sm font-semibold">Execution details</p>
              <p className="text-xs text-muted-foreground">Run record and worker logs</p>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" disabled={loading || refreshing} onClick={refresh} className="gap-2">
            <RefreshCw className={`size-3.5 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </header>

        <div className="mx-auto w-full max-w-375 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <Link href="/executions" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
            <ArrowLeft className="size-4" /> Execution history
          </Link>

          {error && (
            <div role="alert" className="mt-5 flex items-center justify-between gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
              <span className="flex items-center gap-2"><AlertCircle className="size-4" />{error}</span>
              <Button type="button" variant="outline" size="sm" onClick={refresh}>Retry</Button>
            </div>
          )}

          {loading ? (
            <div className="mt-6 space-y-4 animate-pulse">
              <div className="h-28 rounded-xl bg-muted" />
              <div className="h-64 rounded-xl bg-muted" />
            </div>
          ) : execution ? (
            <>
              <section className="mt-6 rounded-xl border border-border bg-card p-5 sm:p-6">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Execution</p>
                    <h1 className="mt-2 break-all font-mono text-lg font-semibold">{execution.id}</h1>
                  </div>
                  <span className={`inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-semibold ${statusTone(execution.status)}`}>
                    <span className="size-1.5 rounded-full bg-current" />
                    {execution.status.replaceAll("_", " ")}
                  </span>
                </div>

                <dl className="mt-6 grid gap-x-6 gap-y-5 border-t border-border/70 pt-5 sm:grid-cols-2 xl:grid-cols-3">
                  <div><dt className="text-xs text-muted-foreground">Task ID</dt><dd className="mt-1 break-all font-mono text-sm">{execution.task_id}</dd></div>
                  <div><dt className="text-xs text-muted-foreground">Trigger ID</dt><dd className="mt-1 break-all font-mono text-sm">{execution.trigger_id}</dd></div>
                  <div><dt className="text-xs text-muted-foreground">Scheduled at</dt><dd className="mt-1 text-sm">{formatDate(execution.scheduled_at)}</dd></div>
                  <div><dt className="text-xs text-muted-foreground">Started at</dt><dd className="mt-1 text-sm">{formatDate(execution.started_at)}</dd></div>
                  <div><dt className="text-xs text-muted-foreground">Completed at</dt><dd className="mt-1 text-sm">{formatDate(execution.completed_at)}</dd></div>
                  <div><dt className="text-xs text-muted-foreground">Duration / retries</dt><dd className="mt-1 text-sm tabular-nums">{execution.duration_ms == null ? "In progress" : `${execution.duration_ms} ms`} · {execution.retry_count} retries</dd></div>
                </dl>

                {execution.error_message && (
                  <div className="mt-5 rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
                    {execution.error_message}
                  </div>
                )}
              </section>

              <section className="mt-6 overflow-hidden rounded-xl border border-border bg-card">
                <div className="flex items-center justify-between border-b border-border/70 px-5 py-4">
                  <div>
                    <h2 className="font-semibold">Execution logs</h2>
                    <p className="mt-1 text-xs text-muted-foreground">Worker messages in recorded order</p>
                  </div>
                  <span className="text-xs tabular-nums text-muted-foreground">{logs.length} entries</span>
                </div>
                {logs.length === 0 ? (
                  <p className="px-5 py-10 text-center text-sm text-muted-foreground">No execution logs were recorded.</p>
                ) : (
                  <ol className="divide-y divide-border/70">
                    {logs.map((log) => (
                      <li key={log.id} className="grid gap-2 px-5 py-4 sm:grid-cols-[170px_70px_minmax(0,1fr)] sm:items-start">
                        <time className="text-xs text-muted-foreground">{formatDate(log.created_at)}</time>
                        <span className={`text-[11px] font-semibold ${log.log_level === "ERROR" ? "text-destructive" : log.log_level === "WARNING" ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground"}`}>{log.log_level}</span>
                        <p className="min-w-0 break-words text-sm">{log.message}</p>
                      </li>
                    ))}
                  </ol>
                )}
              </section>
            </>
          ) : null}

          <footer className="mt-8 flex items-center gap-2 border-t border-border/70 py-5 text-xs text-muted-foreground">
            <ShieldCheck className="size-3.5" />
            Execution data is scoped to your account.
          </footer>
        </div>
      </main>
    </div>
  );
}
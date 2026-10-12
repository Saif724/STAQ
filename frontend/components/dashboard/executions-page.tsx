"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AlertCircle, ArrowDownRight, ArrowLeft, ArrowRight, Clock3, RefreshCw, Search } from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { getExecutions, type Execution, type ExecutionStatus } from "@/lib/api/automation";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

const pageSize = 20;

const statuses: { label: string; value: ExecutionStatus | "ALL" }[] = [
  { label: "All runs", value: "ALL" },
  { label: "Running", value: "RUNNING" },
  { label: "Successful", value: "SUCCESS" },
  { label: "Failed", value: "FAILED" },
  { label: "Timed out", value: "TIMED_OUT" },
  { label: "Cancelled", value: "CANCELLED" },
  { label: "Pending", value: "PENDING" },
];

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Date unavailable";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatDuration(duration: number | null | undefined) {
  if (duration == null) return "In progress";
  if (duration < 1000) return `${duration} ms`;
  return `${(duration / 1000).toFixed(1)} s`;
}

function StatusBadge({ status }: { status: ExecutionStatus }) {
  const styles: Record<ExecutionStatus, string> = {
    PENDING: "border-border bg-muted text-muted-foreground",
    RUNNING: "border-sky-500/20 bg-sky-500/10 text-sky-700 dark:text-sky-300",
    SUCCESS: "border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
    FAILED: "border-destructive/20 bg-destructive/10 text-destructive",
    CANCELLED: "border-border bg-muted text-muted-foreground",
    TIMED_OUT: "border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  };

  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium ${styles[status]}`}>
      <span className="size-1.5 rounded-full bg-current" />
      {status.replaceAll("_", " ")}
    </span>
  );
}

export default function ExecutionsPage() {
  const router = useRouter();
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<ExecutionStatus | "ALL">("ALL");
  const [taskInput, setTaskInput] = useState("");
  const [taskFilter, setTaskFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [refreshVersion, setRefreshVersion] = useState(0);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function loadPage() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }

      setLoading(true);
      setError("");
      try {
        const [userResponse, executionResponse] = await Promise.all([
          getCurrentUser(),
          getExecutions({
            page,
            pageSize,
            taskId: taskFilter || undefined,
            status: status === "ALL" ? undefined : status,
          }),
        ]);
        if (cancelled) return;
        setUser(userResponse.data);
        setExecutions(executionResponse.data.items);
        setTotal(executionResponse.data.total);
      } catch (caught) {
        if (cancelled) return;
        if (caught instanceof ApiError && caught.status === 401) {
          clearTokens();
          router.replace("/login");
          return;
        }
        setError(caught instanceof Error ? caught.message : "Could not load executions.");
      } finally {
        if (!cancelled) {
          setLoading(false);
          setRefreshing(false);
        }
      }
    }

    void loadPage();
    return () => {
      cancelled = true;
    };
  }, [page, refreshVersion, router, status, taskFilter]);

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
    setRefreshVersion((current) => current + 1);
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const firstResult = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const lastResult = Math.min(page * pageSize, total);

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
              <p className="text-sm font-semibold">Executions</p>
              <p className="text-xs text-muted-foreground">Run history and outcomes</p>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" disabled={loading || refreshing} onClick={refresh} className="gap-2">
            <RefreshCw className={`size-3.5 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </header>

        <div className="mx-auto w-full max-w-375 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Automation activity</p>
              <h1 className="mt-2 text-2xl font-bold">Execution history</h1>
              <p className="mt-2 text-sm text-muted-foreground">Inspect scheduled runs, outcomes, and recorded errors.</p>
            </div>
            <p className="text-sm tabular-nums text-muted-foreground">{loading ? "Loading runs…" : `${total} total runs`}</p>
          </div>

          {error && (
            <div role="alert" className="mt-6 flex items-center justify-between gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
              <span className="flex items-center gap-2"><AlertCircle className="size-4" />{error}</span>
              <Button type="button" variant="outline" size="sm" onClick={refresh}>Retry</Button>
            </div>
          )}

          <section className="mt-6 space-y-4 border-y border-border/70 py-4">
            <div className="flex flex-wrap gap-2" aria-label="Filter executions by status">
              {statuses.map((item) => (
                <button
                  key={item.value}
                  type="button"
                  aria-pressed={status === item.value}
                  onClick={() => {
                    setStatus(item.value);
                    setPage(1);
                  }}
                  className={`rounded-lg border px-3 py-2 text-xs font-medium ${status === item.value ? "border-primary/30 bg-primary/10 text-primary" : "border-border text-muted-foreground hover:bg-muted"}`}
                >
                  {item.label}
                </button>
              ))}
            </div>

            <form
              className="flex w-full max-w-xl gap-2"
              onSubmit={(event) => {
                event.preventDefault();
                setTaskFilter(taskInput.trim());
                setPage(1);
              }}
            >
              <label className="sr-only" htmlFor="task-filter">Filter by task ID</label>
              <div className="relative min-w-0 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input id="task-filter" value={taskInput} onChange={(event) => setTaskInput(event.target.value)} placeholder="Filter by task ID" className="h-10 w-full rounded-lg border border-input bg-background pl-9 pr-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" />
              </div>
              <Button type="submit" variant="outline" size="sm">Apply</Button>
            </form>
          </section>

          <section className="mt-5 overflow-hidden rounded-xl border border-border bg-card">
            <div className="overflow-x-auto">
              <table className="w-full min-w-195 text-left text-sm">
                <thead className="border-b border-border bg-muted/35 text-xs text-muted-foreground">
                  <tr>
                    <th className="px-4 py-3 font-medium">Status</th>
                    <th className="px-4 py-3 font-medium">Task</th>
                    <th className="px-4 py-3 font-medium">Scheduled</th>
                    <th className="px-4 py-3 font-medium">Started</th>
                    <th className="px-4 py-3 font-medium">Duration</th>
                    <th className="px-4 py-3 font-medium">Retries</th>
                    <th className="px-4 py-3"><span className="sr-only">Details</span></th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/70">
                  {loading ? (
                    Array.from({ length: 5 }, (_, index) => (
                      <tr key={index} className="animate-pulse">
                        {Array.from({ length: 7 }, (_, cell) => <td key={cell} className="px-4 py-4"><div className="h-4 w-24 rounded bg-muted" /></td>)}
                      </tr>
                    ))
                  ) : executions.length === 0 ? (
                    <tr><td colSpan={7} className="px-6 py-16 text-center">
                      <ArrowDownRight className="mx-auto size-6 text-muted-foreground" />
                      <p className="mt-3 font-medium">No executions found</p>
                      <p className="mt-1 text-sm text-muted-foreground">Runs will appear here after scheduled tasks are processed.</p>
                    </td></tr>
                  ) : executions.map((execution) => (
                    <tr key={execution.id} className="hover:bg-muted/25">
                      <td className="px-4 py-3"><StatusBadge status={execution.status} /></td>
                      <td className="px-4 py-3 font-mono text-xs" title={execution.task_id}>{execution.task_id.slice(0, 8)}</td>
                      <td className="px-4 py-3 text-muted-foreground">{formatDate(execution.scheduled_at)}</td>
                      <td className="px-4 py-3 text-muted-foreground">{formatDate(execution.started_at)}</td>
                      <td className="px-4 py-3 tabular-nums text-muted-foreground">{formatDuration(execution.duration_ms)}</td>
                      <td className="px-4 py-3 tabular-nums text-muted-foreground">{execution.retry_count}</td>
                      <td className="px-4 py-3 text-right">
                        <Link href={`/executions/${execution.id}`} className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
                          Details <ArrowRight className="size-3.5" />
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="flex flex-col gap-3 border-t border-border/70 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-xs text-muted-foreground">{loading ? " " : `Showing ${firstResult}–${lastResult} of ${total}`}</p>
              <div className="flex items-center gap-2">
                <span className="mr-2 text-xs tabular-nums text-muted-foreground">Page {page} of {totalPages}</span>
                <Button type="button" variant="outline" size="icon" aria-label="Previous page" title="Previous page" disabled={loading || page <= 1} onClick={() => setPage((current) => current - 1)}><ArrowLeft className="size-4" /></Button>
                <Button type="button" variant="outline" size="icon" aria-label="Next page" title="Next page" disabled={loading || page >= totalPages} onClick={() => setPage((current) => current + 1)}><ArrowRight className="size-4" /></Button>
              </div>
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}
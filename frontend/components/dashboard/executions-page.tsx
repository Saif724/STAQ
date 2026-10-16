"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  AlertCircle,
  ArrowDownRight,
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Clock3,
  RefreshCw,
  Search,
  XCircle,
  Timer,
  ListChecks,
} from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import {
  getExecutions,
  type Execution,
  type ExecutionStatus,
} from "@/lib/api/automation";
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

function formatDate(value: string | null | undefined) {
  if (!value) return "Not available";

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
    RUNNING:
      "border-sky-500/20 bg-sky-500/10 text-sky-700 dark:text-sky-300",
    SUCCESS:
      "border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
    FAILED: "border-destructive/20 bg-destructive/10 text-destructive",
    CANCELLED: "border-border bg-muted text-muted-foreground",
    TIMED_OUT:
      "border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  };

  const label = status.replaceAll("_", " ");

  return (
    <span
      className={`inline-flex max-w-full items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium ${styles[status]}`}
    >
      {status === "SUCCESS" ? (
        <CheckCircle2 className="size-3.5 shrink-0" />
      ) : status === "FAILED" || status === "TIMED_OUT" ? (
        <XCircle className="size-3.5 shrink-0" />
      ) : (
        <span className="size-1.5 shrink-0 rounded-full bg-current" />
      )}
      <span>{label}</span>
    </span>
  );
}

function ExecutionDetailsLink({ executionId }: { executionId: string }) {
  return (
    <Link
      href={`/executions/${executionId}`}
      className="inline-flex min-h-9 items-center justify-center gap-1.5 rounded-lg border border-primary/20 px-3 py-2 text-xs font-semibold text-primary transition hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      View details
      <ArrowRight className="size-3.5" />
    </Link>
  );
}

function EmptyExecutions() {
  return (
    <div className="px-5 py-14 text-center sm:py-16">
      <div className="mx-auto flex size-12 items-center justify-center rounded-2xl bg-muted">
        <ArrowDownRight className="size-6 text-muted-foreground" />
      </div>
      <h2 className="mt-4 text-sm font-semibold">No executions found</h2>
      <p className="mx-auto mt-1 max-w-sm text-sm leading-6 text-muted-foreground">
        No runs match the current filters. Try a different status or task ID.
      </p>
    </div>
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

        setError(
          caught instanceof Error
            ? caught.message
            : "Could not load executions.",
        );
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
    <div className="min-h-screen overflow-x-clip bg-background text-foreground">
      <Sidebar
        mobileOpen={mobileOpen}
        onClose={() => setMobileOpen(false)}
        user={user}
        onLogout={handleLogout}
        loggingOut={loggingOut}
        activePage="executions"
      />

      <main className="min-h-screen min-w-0 lg:pl-66">
        <header className="sticky top-0 z-30 flex min-h-19 items-center justify-between gap-3 border-b border-border/80 bg-background/90 px-4 py-3 backdrop-blur-xl sm:px-6 lg:px-8">
          <div className="flex min-w-0 items-center gap-3">
            <button
              type="button"
              aria-label="Open navigation"
              onClick={() => setMobileOpen(true)}
              className="shrink-0 rounded-lg border border-border p-2 text-muted-foreground transition hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring lg:hidden"
            >
              <ListChecks className="size-4" />
            </button>

            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">Executions</p>
              <p className="truncate text-xs text-muted-foreground">
                Run history and outcomes
              </p>
            </div>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={loading || refreshing}
            onClick={refresh}
            className="shrink-0 gap-2"
          >
            <RefreshCw
              className={`size-3.5 ${refreshing ? "animate-spin" : ""}`}
            />
            <span className="hidden xs:inline">Refresh</span>
            <span className="xs:hidden">Reload</span>
          </Button>
        </header>

        <div className="mx-auto w-full min-w-0 max-w-375 px-4 py-5 sm:px-6 sm:py-8 lg:px-8">
          <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">
                Automation activity
              </p>
              <h1 className="mt-2 text-2xl font-bold tracking-tight sm:text-3xl">
                Execution history
              </h1>
              <p className="mt-2 max-w-xl text-sm leading-6 text-muted-foreground">
                Inspect scheduled runs, outcomes, and recorded errors.
              </p>
            </div>

            <p className="text-sm tabular-nums text-muted-foreground sm:shrink-0">
              {loading ? "Loading runs…" : `${total} total runs`}
            </p>
          </div>

          {error && (
            <div
              role="alert"
              className="mt-6 flex min-w-0 flex-col gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive sm:flex-row sm:items-center sm:justify-between"
            >
              <span className="flex min-w-0 items-start gap-2 break-words">
                <AlertCircle className="mt-0.5 size-4 shrink-0" />
                <span>{error}</span>
              </span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={refresh}
                className="self-start sm:self-auto"
              >
                Retry
              </Button>
            </div>
          )}

          <section className="mt-6 min-w-0 space-y-4 rounded-2xl border border-border/80 bg-card p-4 sm:p-5">
            <div>
              <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Filter by status
              </p>

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
                    className={`min-h-9 rounded-lg border px-3 py-2 text-xs font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
                      status === item.value
                        ? "border-primary/30 bg-primary/10 text-primary"
                        : "border-border text-muted-foreground hover:bg-muted hover:text-foreground"
                    }`}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>

            <form
              className="flex w-full min-w-0 flex-col gap-2 sm:max-w-xl sm:flex-row"
              onSubmit={(event) => {
                event.preventDefault();
                setTaskFilter(taskInput.trim());
                setPage(1);
              }}
            >
              <label className="sr-only" htmlFor="task-filter">
                Filter by task ID
              </label>

              <div className="relative min-w-0 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  id="task-filter"
                  value={taskInput}
                  onChange={(event) => setTaskInput(event.target.value)}
                  placeholder="Filter by task ID"
                  className="h-10 w-full min-w-0 rounded-lg border border-input bg-background pl-9 pr-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
              </div>

              <Button type="submit" variant="outline" className="w-full sm:w-auto">
                Apply filters
              </Button>

              {taskFilter && (
                <Button
                  type="button"
                  variant="ghost"
                  className="w-full sm:w-auto"
                  onClick={() => {
                    setTaskInput("");
                    setTaskFilter("");
                    setPage(1);
                  }}
                >
                  Clear
                </Button>
              )}
            </form>
          </section>

          <section className="mt-5 min-w-0 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/5">
            {loading ? (
              <div className="space-y-3 p-4 sm:p-5">
                {Array.from({ length: 5 }, (_, index) => (
                  <div
                    key={index}
                    className="animate-pulse rounded-xl border border-border/70 p-4"
                  >
                    <div className="h-4 w-24 rounded bg-muted" />
                    <div className="mt-3 h-4 w-40 max-w-full rounded bg-muted" />
                    <div className="mt-3 h-3 w-56 max-w-full rounded bg-muted" />
                  </div>
                ))}
              </div>
            ) : executions.length === 0 ? (
              <EmptyExecutions />
            ) : (
              <>
                {/* Mobile layout: one readable card per execution. */}
                <div className="divide-y divide-border/70 md:hidden">
                  {executions.map((execution) => (
                    <article
                      key={execution.id}
                      className="min-w-0 p-4 transition-colors hover:bg-muted/20 sm:p-5"
                    >
                      <div className="flex min-w-0 items-start justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-xs text-muted-foreground">
                            Task ID
                          </p>
                          <p
                            className="mt-1 break-all font-mono text-sm font-medium"
                            title={execution.task_id}
                          >
                            {execution.task_id}
                          </p>
                        </div>

                        <div className="shrink-0">
                          <StatusBadge status={execution.status} />
                        </div>
                      </div>

                      <div className="mt-4 grid grid-cols-2 gap-x-3 gap-y-4">
                        <div className="min-w-0">
                          <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <Clock3 className="size-3.5 shrink-0" />
                            Scheduled
                          </p>
                          <p className="mt-1 break-words text-sm leading-5">
                            {formatDate(execution.scheduled_at)}
                          </p>
                        </div>

                        <div className="min-w-0">
                          <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <Timer className="size-3.5 shrink-0" />
                            Started
                          </p>
                          <p className="mt-1 break-words text-sm leading-5">
                            {formatDate(execution.started_at)}
                          </p>
                        </div>

                        <div className="min-w-0">
                          <p className="text-xs text-muted-foreground">
                            Duration
                          </p>
                          <p className="mt-1 text-sm font-medium tabular-nums">
                            {formatDuration(execution.duration_ms)}
                          </p>
                        </div>

                        <div className="min-w-0">
                          <p className="text-xs text-muted-foreground">
                            Retries
                          </p>
                          <p className="mt-1 text-sm font-medium tabular-nums">
                            {execution.retry_count}
                          </p>
                        </div>
                      </div>

                      <div className="mt-4 border-t border-border/70 pt-3">
                        <ExecutionDetailsLink executionId={execution.id} />
                      </div>
                    </article>
                  ))}
                </div>

                {/* Desktop layout: keep the dense table for wider screens. */}
                <div className="hidden w-full max-w-full overflow-x-auto overscroll-x-contain md:block">
                  <table className="w-full min-w-[850px] text-left text-sm">
                    <thead className="border-b border-border bg-muted/35 text-xs text-muted-foreground">
                      <tr>
                        <th className="whitespace-nowrap px-4 py-3 font-medium">
                          Status
                        </th>
                        <th className="px-4 py-3 font-medium">Task ID</th>
                        <th className="whitespace-nowrap px-4 py-3 font-medium">
                          Scheduled
                        </th>
                        <th className="whitespace-nowrap px-4 py-3 font-medium">
                          Started
                        </th>
                        <th className="whitespace-nowrap px-4 py-3 font-medium">
                          Duration
                        </th>
                        <th className="whitespace-nowrap px-4 py-3 font-medium">
                          Retries
                        </th>
                        <th className="px-4 py-3">
                          <span className="sr-only">Details</span>
                        </th>
                      </tr>
                    </thead>

                    <tbody className="divide-y divide-border/70">
                      {executions.map((execution) => (
                        <tr
                          key={execution.id}
                          className="transition-colors hover:bg-muted/25"
                        >
                          <td className="whitespace-nowrap px-4 py-3">
                            <StatusBadge status={execution.status} />
                          </td>
                          <td
                            className="px-4 py-3 font-mono text-xs"
                            title={execution.task_id}
                          >
                            <span className="block max-w-40 truncate">
                              {execution.task_id}
                            </span>
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                            {formatDate(execution.scheduled_at)}
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                            {formatDate(execution.started_at)}
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 tabular-nums text-muted-foreground">
                            {formatDuration(execution.duration_ms)}
                          </td>
                          <td className="px-4 py-3 tabular-nums text-muted-foreground">
                            {execution.retry_count}
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-right">
                            <Link
                              href={`/executions/${execution.id}`}
                              className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              Details
                              <ArrowRight className="size-3.5" />
                            </Link>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </>
            )}

            <div className="flex flex-col gap-3 border-t border-border/70 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
              <p className="text-xs text-muted-foreground">
                {loading
                  ? "Loading results…"
                  : total === 0
                    ? "No results to display"
                    : `Showing ${firstResult}–${lastResult} of ${total}`}
              </p>

              <div className="flex items-center justify-between gap-3 sm:justify-end">
                <span className="text-xs tabular-nums text-muted-foreground">
                  Page {page} of {totalPages}
                </span>

                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    aria-label="Previous page"
                    title="Previous page"
                    disabled={loading || page <= 1}
                    onClick={() => setPage((current) => current - 1)}
                  >
                    <ArrowLeft className="size-4" />
                  </Button>

                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    aria-label="Next page"
                    title="Next page"
                    disabled={loading || page >= totalPages}
                    onClick={() => setPage((current) => current + 1)}
                  >
                    <ArrowRight className="size-4" />
                  </Button>
                </div>
              </div>
            </div>
          </section>

          <footer className="mt-6 flex flex-col gap-2 border-t border-border/70 py-5 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <p>STAQ · Your automation workspace</p>
            <p className="flex items-center gap-1.5">
              <span className="size-1.5 rounded-full bg-emerald-500" />
              Authenticated workspace
            </p>
          </footer>
        </div>
      </main>
    </div>
  );
}

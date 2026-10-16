"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  AlertCircle,
  ArrowLeft,
  Clock3,
  RefreshCw,
  ShieldCheck,
  Timer,
  CalendarClock,
  Fingerprint,
  ListChecks,
  FileText,
  CheckCircle2,
  XCircle,
} from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
  getExecution,
  getExecutionLogs,
  type Execution,
  type ExecutionLog,
  type ExecutionStatus,
} from "@/lib/api/automation";
import { getCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

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
  return `${(duration / 1000).toFixed(2)} s`;
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

function StatusIcon({ status }: { status: ExecutionStatus }) {
  if (status === "SUCCESS") {
    return <CheckCircle2 className="size-4 shrink-0" />;
  }

  if (status === "FAILED" || status === "TIMED_OUT") {
    return <XCircle className="size-4 shrink-0" />;
  }

  return <span className="size-2 shrink-0 rounded-full bg-current" />;
}

function MetadataItem({
  label,
  value,
  icon: Icon,
  mono = false,
}: {
  label: string;
  value: string;
  icon: typeof Clock3;
  mono?: boolean;
}) {
  return (
    <div className="min-w-0 rounded-xl border border-border/70 bg-background/60 p-3.5 sm:p-4">
      <dt className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
        <Icon className="size-3.5 shrink-0" />
        {label}
      </dt>
      <dd
        className={`mt-2 break-words text-sm leading-6 ${
          mono ? "font-mono text-xs sm:text-sm [overflow-wrap:anywhere]" : ""
        }`}
      >
        {value}
      </dd>
    </div>
  );
}

function logLevelTone(level: string) {
  switch (level) {
    case "ERROR":
      return "border-destructive/20 bg-destructive/10 text-destructive";
    case "WARNING":
      return "border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300";
    case "INFO":
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
        const [userResponse, executionResponse, logsResponse] =
          await Promise.all([
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

        setError(
          caught instanceof Error
            ? caught.message
            : "Could not load this execution.",
        );
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
              <Clock3 className="size-4" />
            </button>

            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">
                Execution details
              </p>
              <p className="truncate text-xs text-muted-foreground">
                Run record and worker logs
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
            <span>Refresh</span>
          </Button>
        </header>

        <div className="mx-auto w-full min-w-0 max-w-375 px-4 py-5 sm:px-6 sm:py-8 lg:px-8">
          <Link
            href="/executions"
            className="inline-flex min-h-9 items-center gap-2 rounded-lg text-sm text-muted-foreground transition hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <ArrowLeft className="size-4" />
            Execution history
          </Link>

          {error && (
            <div
              role="alert"
              className="mt-5 flex min-w-0 flex-col gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive sm:flex-row sm:items-center sm:justify-between"
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

          {loading ? (
            <div className="mt-6 animate-pulse space-y-5">
              <div className="h-44 rounded-2xl border border-border bg-muted sm:h-52" />
              <div className="h-72 rounded-2xl border border-border bg-muted" />
            </div>
          ) : execution ? (
            <>
              <section className="mt-5 min-w-0 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/5 sm:mt-6">
                <div className="p-4 sm:p-6 lg:p-7">
                  <div className="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0 flex-1">
                      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">
                        Execution record
                      </p>

                      <h1 className="mt-2 break-words text-xl font-bold tracking-tight sm:text-2xl">
                        Execution details
                      </h1>

                      <p className="mt-2 break-all font-mono text-xs leading-5 text-muted-foreground [overflow-wrap:anywhere] sm:text-sm">
                        {execution.id}
                      </p>
                    </div>

                    <span
                      className={`inline-flex w-fit max-w-full shrink-0 items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-semibold ${statusTone(execution.status)}`}
                    >
                      <StatusIcon status={execution.status} />
                      <span className="break-words">
                        {execution.status.replaceAll("_", " ")}
                      </span>
                    </span>
                  </div>

                  <div className="mt-6 grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
                    <MetadataItem
                      label="Task ID"
                      value={execution.task_id}
                      icon={ListChecks}
                      mono
                    />

                    <MetadataItem
                      label="Trigger ID"
                      value={execution.trigger_id}
                      icon={Fingerprint}
                      mono
                    />

                    <MetadataItem
                      label="Scheduled at"
                      value={formatDate(execution.scheduled_at)}
                      icon={CalendarClock}
                    />

                    <MetadataItem
                      label="Started at"
                      value={formatDate(execution.started_at)}
                      icon={Clock3}
                    />

                    <MetadataItem
                      label="Completed at"
                      value={formatDate(execution.completed_at)}
                      icon={CheckCircle2}
                    />

                    <MetadataItem
                      label="Duration"
                      value={formatDuration(execution.duration_ms)}
                      icon={Timer}
                    />

                    <MetadataItem
                      label="Retries"
                      value={String(execution.retry_count)}
                      icon={RefreshCw}
                    />
                  </div>

                  {execution.error_message && (
                    <div className="mt-5 min-w-0 rounded-xl border border-destructive/20 bg-destructive/5 p-4">
                      <div className="flex items-center gap-2 text-sm font-semibold text-destructive">
                        <AlertCircle className="size-4 shrink-0" />
                        Execution error
                      </div>

                      <p className="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-destructive [overflow-wrap:anywhere]">
                        {execution.error_message}
                      </p>
                    </div>
                  )}
                </div>
              </section>

              <section className="mt-5 min-w-0 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/5 sm:mt-6">
                <div className="flex min-w-0 items-center justify-between gap-3 border-b border-border/70 px-4 py-4 sm:px-6">
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                      <FileText className="size-4.5" />
                    </div>

                    <div className="min-w-0">
                      <h2 className="text-sm font-semibold sm:text-base">
                        Execution logs
                      </h2>
                      <p className="mt-1 text-xs leading-5 text-muted-foreground">
                        Worker messages in recorded order
                      </p>
                    </div>
                  </div>

                  <span className="shrink-0 rounded-lg border border-border bg-muted/50 px-2.5 py-1.5 text-xs tabular-nums text-muted-foreground">
                    {logs.length} entries
                  </span>
                </div>

                {logs.length === 0 ? (
                  <div className="px-5 py-12 text-center">
                    <div className="mx-auto flex size-11 items-center justify-center rounded-xl bg-muted">
                      <FileText className="size-5 text-muted-foreground" />
                    </div>
                    <p className="mt-3 text-sm font-medium">
                      No logs recorded
                    </p>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Worker messages will appear here when available.
                    </p>
                  </div>
                ) : (
                  <ol className="divide-y divide-border/70">
                    {logs.map((log: ExecutionLog) => (
                      <li
                        key={log.id}
                        className="grid min-w-0 gap-3 p-4 sm:p-5 md:grid-cols-[minmax(140px,190px)_minmax(75px,90px)_minmax(0,1fr)] md:items-start"
                      >
                        <time className="break-words text-xs leading-5 text-muted-foreground">
                          {formatDate(log.created_at)}
                        </time>

                        <span
                          className={`inline-flex w-fit max-w-full items-center rounded-md border px-2 py-1 text-[10px] font-bold tracking-wide ${logLevelTone(log.log_level)}`}
                        >
                          {log.log_level}
                        </span>

                        <p className="min-w-0 whitespace-pre-wrap break-words text-sm leading-6 [overflow-wrap:anywhere]">
                          {log.message}
                        </p>
                      </li>
                    ))}
                  </ol>
                )}
              </section>
            </>
          ) : !error ? (
            <section className="mt-6 rounded-2xl border border-border bg-card px-5 py-12 text-center">
              <AlertCircle className="mx-auto size-6 text-muted-foreground" />
              <h1 className="mt-3 font-semibold">Execution unavailable</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                This execution could not be displayed.
              </p>
              <Link
                href="/executions"
                className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline"
              >
                <ArrowLeft className="size-4" />
                Back to execution history
              </Link>
            </section>
          ) : null}

          <footer className="mt-6 flex flex-col gap-2 border-t border-border/70 py-5 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <p className="flex items-center gap-2">
              <ShieldCheck className="size-3.5 shrink-0" />
              Execution data is scoped to your account.
            </p>
            <p>STAQ · Your automation workspace</p>
          </footer>
        </div>
      </main>
    </div>
  );
}

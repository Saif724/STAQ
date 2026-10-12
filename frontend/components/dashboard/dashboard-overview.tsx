"use client";

import { useCallback, useEffect, useState, useSyncExternalStore } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import Link from "next/link";
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  Clock3,
  Command,
  ListTodo,
  Menu,
  Moon,
  Plus,
  RefreshCw,
  ShieldCheck,
  Sun,
  Workflow,
} from "lucide-react";

import { ApiError } from "@/lib/api/client";
import {
  getDashboardSummary,
  getCurrentUser,
  getQueues,
  getTasks,
  type CurrentUser,
  type DashboardRecentExecution,
  type DashboardSummary,
  type Queue,
  type Task,
} from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";
import { Button } from "@/components/ui/button";

import Sidebar from "@/components/dashboard/sidebar";
import TaskCreationWizard from "@/components/tasks/task-creation-wizard";

type DashboardData = {
  user: CurrentUser | null;
  tasks: Task[];
  queues: Queue[];
  summary: DashboardSummary | null;
};

const subscribe = () => () => {};
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

function getInitials(name: string) {
  return (
    name
      .trim()
      .split(/\s+/)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? "")
      .join("") || "U"
  );
}

function formatDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Date unavailable";
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatDateInTimezone(value: string, timezone: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Date unavailable";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: timezone,
  }).format(date);
}

function ExecutionStatusBadge({ status }: { status: DashboardRecentExecution["status"] }) {
  const styles: Record<DashboardRecentExecution["status"], string> = {
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

function StatCard({
  label,
  value,
  description,
  icon: Icon,
  accent,
  suffix,
}: {
  label: string;
  value: number | null;
  description: string;
  icon: typeof ListTodo;
  accent: string;
  suffix?: string;
}) {
  return (
    <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/2 transition-shadow hover:shadow-md">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className="mt-3 text-3xl font-semibold tracking-tight tabular-nums">
            {value ?? "—"}{value !== null && suffix}
          </p>
        </div>

        <div
          className={`flex size-11 items-center justify-center rounded-xl ${accent}`}
        >
          <Icon className="size-5" />
        </div>
      </div>

      <p className="mt-4 text-xs text-muted-foreground">{description}</p>
    </div>
  );
}

export default function DashboardOverview() {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();

  const mounted = useSyncExternalStore(
    subscribe,
    getClientSnapshot,
    getServerSnapshot,
  );

  const [data, setData] = useState<DashboardData>({
    user: null,
    tasks: [],
    queues: [],
    summary: null,
  });
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  const [createTaskOpen, setCreateTaskOpen] = useState(false);

  const loadDashboard = useCallback(async () => {
    if (!getAccessToken()) {
      clearTokens();
      router.replace("/login");
      return;
    }

    setRefreshing(true);
    setErrors([]);

    try {
      const results = await Promise.allSettled([
        getCurrentUser(),
        getTasks(),
        getQueues(),
        getDashboardSummary(),
      ]);

      const authFailure = results.some(
        (result) =>
          result.status === "rejected" &&
          result.reason instanceof ApiError &&
          result.reason.status === 401,
      );

      if (authFailure) {
        clearTokens();
        router.replace("/login");
        return;
      }

      const nextErrors: string[] = [];
      const [userResult, taskResult, queueResult, summaryResult] = results;

      if (userResult.status === "fulfilled") {
        setData((previous) => ({
          ...previous,
          user: userResult.value.data,
        }));
      } else {
        nextErrors.push("Your profile could not be loaded.");
      }

      if (taskResult.status === "fulfilled") {
        setData((previous) => ({
          ...previous,
          tasks: taskResult.value.data,
        }));
      } else {
        nextErrors.push("Tasks could not be loaded.");
      }

      if (queueResult.status === "fulfilled") {
        setData((previous) => ({
          ...previous,
          queues: queueResult.value.data,
        }));
      } else {
        nextErrors.push("Queues could not be loaded.");
      }

      if (summaryResult.status === "fulfilled") {
        setData((previous) => ({
          ...previous,
          summary: summaryResult.value.data,
        }));
      } else {
        nextErrors.push("Execution summary could not be loaded.");
      }

      setErrors(nextErrors);
    } catch {
      setErrors(["Unable to load dashboard data. Please try again."]);
    } finally {
      setRefreshing(false);
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => void loadDashboard(), 0);
    return () => window.clearTimeout(timeoutId);
  }, [loadDashboard]);

  async function handleLogout() {
    setLoggingOut(true);

    try {
      await logoutUser();
    } catch {
      // The local session is cleared even if the API request fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  const activeTasks = data.tasks.filter(
    (task) => task.status === "Active",
  ).length;

  const pausedTasks = data.tasks.filter(
    (task) => task.status === "Paused",
  ).length;

  const archivedTasks = data.tasks.filter(
    (task) => task.status === "Archived",
  ).length;

  const failedRuns = data.summary
    ? data.summary.failed_count + data.summary.timed_out_count
    : null;
  const recentExecutions = data.summary?.recent_executions ?? [];
  const recentFailures = data.summary?.recent_failures ?? [];
  const upcomingExecutions = data.summary?.upcoming_executions ?? [];

  const firstName = data.user?.full_name?.trim().split(/\s+/)[0];

  return (
    <div className="min-h-screen bg-background">
      <Sidebar
        mobileOpen={mobileOpen}
        onClose={() => setMobileOpen(false)}
        user={data.user}
        onLogout={handleLogout}
        loggingOut={loggingOut}
        activePage="overview"
      />

      <div className="min-h-screen lg:pl-66">
        <header className="sticky top-0 z-30 flex h-19 items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-9">
          <div className="flex items-center gap-3">
            <button
              type="button"
              aria-label="Open navigation"
              onClick={() => setMobileOpen(true)}
              className="rounded-lg p-2 text-muted-foreground hover:bg-muted lg:hidden"
            >
              <Menu className="size-5" />
            </button>

            <div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>Workspace</span>
                <span>/</span>
                <span className="text-foreground">Overview</span>
              </div>
              <h1 className="mt-1 text-base font-semibold tracking-tight">
                Dashboard
              </h1>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className="hidden items-center gap-1.5 rounded-full border border-border/80 px-3 py-1.5 text-xs text-muted-foreground sm:flex">
              <span className="size-1.5 rounded-full bg-emerald-500" />
              Workspace
            </span>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              title="Refresh dashboard"
              aria-label="Refresh dashboard"
              onClick={() => void loadDashboard()}
            >
              <RefreshCw
                className={`size-4 ${refreshing ? "animate-spin" : ""}`}
              />
            </Button>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              title="Toggle theme"
              aria-label="Toggle theme"
              onClick={() =>
                setTheme(resolvedTheme === "dark" ? "light" : "dark")
              }
            >
              {!mounted ? (
                <span className="size-4" />
              ) : resolvedTheme === "dark" ? (
                <Sun className="size-4" />
              ) : (
                <Moon className="size-4" />
              )}
            </Button>

            <div className="ml-1 hidden h-8 w-px bg-border sm:block" />

            <div className="hidden items-center gap-2 sm:flex">
              <div className="flex size-8 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
                {getInitials(data.user?.full_name ?? "User")}
              </div>
              <span className="max-w-32 truncate text-sm font-medium">
                {data.user?.full_name ?? "Account"}
              </span>
              <ChevronDown className="size-3.5 text-muted-foreground" />
            </div>
          </div>
        </header>

        <main className="mx-auto max-w-360 px-4 py-7 sm:px-6 sm:py-9 lg:px-9">
          <section className="relative overflow-hidden rounded-3xl border border-primary/15 bg-linear-to-br from-primary/9 via-card to-card p-6 sm:p-8 lg:p-9">
            <div className="pointer-events-none absolute -right-10 -top-20 size-64 rounded-full bg-primary/10 blur-3xl" />
            <div className="pointer-events-none absolute -bottom-24 right-1/4 size-48 rounded-full bg-violet-400/10 blur-3xl" />

            <div className="relative flex flex-col justify-between gap-6 md:flex-row md:items-center">
              <div className="max-w-2xl">
                <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-background/70 px-3 py-1.5 text-xs font-medium text-primary">
                  <Command className="size-3.5" />
                  Your automation workspace
                </div>

                <h2 className="mt-5 text-3xl font-semibold tracking-tight sm:text-4xl">
                  {firstName
                    ? `Welcome back, ${firstName}.`
                    : "Welcome to STAQ."}
                </h2>

                <p className="mt-3 max-w-xl text-sm leading-6 text-muted-foreground sm:text-base">
                  Your workflows start here. Keep track of your tasks, organize
                  them into queues, and build a more reliable automation
                  routine.
                </p>

                <div className="mt-6 flex flex-wrap items-center gap-3">
                  <Button
                    type="button"
                    onClick={() => {
                      setCreateTaskOpen(true);
                    }}
                    className="gap-2"
                  >
                    <Plus className="size-4" />
                    Create a task
                  </Button>
                  <Link href="/executions" className="inline-flex h-10 items-center rounded-lg border border-border bg-background px-4 text-sm font-medium hover:bg-muted">
                    View executions
                  </Link>
                </div>
              </div>

              <div className="relative hidden shrink-0 md:block">
                <div className="flex size-32 items-center justify-center rounded-[2rem] border border-primary/15 bg-background/70 shadow-xl shadow-primary/5">
                  <div className="flex size-20 items-center justify-center rounded-3xl bg-primary text-primary-foreground">
                    <Workflow className="size-10" />
                  </div>
                  <span className="absolute right-1 top-2 flex size-7 items-center justify-center rounded-full border border-border bg-card text-emerald-500 shadow-sm">
                    <CheckCircle2 className="size-4" />
                  </span>
                  <span className="absolute bottom-2 left-1 flex size-7 items-center justify-center rounded-full border border-border bg-card text-primary shadow-sm">
                    <Clock3 className="size-4" />
                  </span>
                </div>
              </div>
            </div>
          </section>

          {errors.length > 0 && (
            <div
              role="status"
              className="mt-6 flex flex-col gap-3 rounded-2xl border border-amber-300/70 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="flex items-start gap-2.5">
                <AlertCircle className="mt-0.5 size-4 shrink-0" />
                <div>
                  <p className="font-semibold">
                    Some dashboard data is unavailable
                  </p>
                  <ul className="mt-1 list-inside list-disc text-xs leading-5">
                    {errors.map((error) => (
                      <li key={error}>{error}</li>
                    ))}
                  </ul>
                </div>
              </div>

              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => void loadDashboard()}
              >
                Try again
              </Button>
            </div>
          )}

          <section className="mt-8">
            <div className="mb-4 flex items-end justify-between gap-3">
              <div>
                <h3 className="text-base font-semibold tracking-tight">
                  Workspace overview
                </h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  A quick look at your automation setup.
                </p>
              </div>
              <span className="text-xs text-muted-foreground">
                {loading ? "Loading data…" : "Live API data"}
              </span>
            </div>

            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
              <StatCard
                label="Total tasks"
                value={loading ? null : data.tasks.length}
                description="All tasks in your workspace"
                icon={ListTodo}
                accent="bg-primary/10 text-primary"
              />

              <StatCard
                label="Active tasks"
                value={loading ? null : activeTasks}
                description="Tasks currently marked active"
                icon={Activity}
                accent="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              />

              <StatCard
                label="Successful runs"
                value={data.summary?.successful_count ?? null}
                description={`Completed successfully · last ${data.summary?.window_days ?? 30} days`}
                icon={CheckCircle2}
                accent="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              />

              <StatCard
                label="Failed or timed out"
                value={failedRuns}
                description={`Unsuccessful completed runs · last ${data.summary?.window_days ?? 30} days`}
                icon={AlertCircle}
                accent="bg-amber-500/10 text-amber-600 dark:text-amber-400"
              />

              <StatCard
                label="Success rate"
                value={data.summary ? Math.round(data.summary.success_rate) : null}
                description={`Completed runs · last ${data.summary?.window_days ?? 30} days`}
                icon={CheckCircle2}
                accent="bg-sky-500/10 text-sky-600 dark:text-sky-400"
                suffix="%"
              />
            </div>
          </section>

          <div className="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1.6fr)_minmax(300px,0.8fr)]">
            <section className="min-w-0 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/2">
              <div className="flex items-center justify-between border-b border-border/70 px-5 py-5 sm:px-6">
                <div>
                  <h3 className="font-semibold tracking-tight">Recent executions</h3>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Latest recorded runs across your tasks.
                  </p>
                </div>

                <Link href="/executions" className="text-xs font-medium text-primary hover:underline">View all</Link>
              </div>

              {loading ? (
                <div className="space-y-4 p-6">
                  {[1, 2, 3].map((item) => (
                    <div key={item} className="flex animate-pulse gap-3">
                      <div className="size-10 rounded-xl bg-muted" />
                      <div className="flex-1 space-y-2 py-1">
                        <div className="h-3 w-2/5 rounded bg-muted" />
                        <div className="h-3 w-1/3 rounded bg-muted" />
                      </div>
                    </div>
                  ))}
                </div>
              ) : !data.summary ? (
                <p className="px-6 py-12 text-center text-sm text-muted-foreground">Execution summary is unavailable.</p>
              ) : recentExecutions.length === 0 ? (
                <div className="flex flex-col items-center px-6 py-12 text-center">
                  <div className="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <Activity className="size-5" />
                  </div>
                  <h4 className="mt-4 text-sm font-semibold">No executions yet</h4>
                  <p className="mt-1 max-w-xs text-sm leading-6 text-muted-foreground">
                    Scheduled runs will appear here after the worker processes them.
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-border/70">
                  {recentExecutions.map((execution) => (
                    <Link key={execution.id} href={`/executions/${execution.id}`} className="flex items-center gap-3 px-5 py-4 transition-colors hover:bg-muted/30 sm:px-6">
                      <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/8 text-primary">
                        <Workflow className="size-4.5" />
                      </div>

                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{execution.task_name}</p>
                        <p className="mt-1 truncate text-xs text-muted-foreground">
                          Scheduled {formatDate(execution.scheduled_at)}
                        </p>
                      </div>

                      <ExecutionStatusBadge status={execution.status} />
                    </Link>
                  ))}
                </div>
              )}

              <div className="flex items-center justify-between border-t border-border/70 bg-muted/20 px-5 py-3.5 sm:px-6">
                <span className="text-xs text-muted-foreground">
                  {data.tasks.length} task{data.tasks.length === 1 ? "" : "s"}{" "}
                  total
                </span>
                <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <Clock3 className="size-3.5" />
                  Latest executions
                </span>
              </div>
            </section>

            <div className="space-y-6">
              <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/2 sm:p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-semibold tracking-tight">
                      Task distribution
                    </h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Current task statuses.
                    </p>
                  </div>
                  <Activity className="size-4 text-muted-foreground" />
                </div>

                <div className="mt-6 space-y-5">
                  {[
                    {
                      label: "Active",
                      value: activeTasks,
                      color: "bg-emerald-500",
                    },
                    {
                      label: "Paused",
                      value: pausedTasks,
                      color: "bg-amber-500",
                    },
                    {
                      label: "Archived",
                      value: archivedTasks,
                      color: "bg-slate-400",
                    },
                  ].map((item) => {
                    const percentage =
                      data.tasks.length === 0
                        ? 0
                        : (item.value / data.tasks.length) * 100;

                    return (
                      <div key={item.label}>
                        <div className="mb-2 flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">
                            {item.label}
                          </span>
                          <span className="font-medium tabular-nums">
                            {item.value}
                            <span className="ml-1.5 text-xs font-normal text-muted-foreground">
                              {Math.round(percentage)}%
                            </span>
                          </span>
                        </div>
                        <div className="h-2 overflow-hidden rounded-full bg-muted">
                          <div
                            className={`h-full rounded-full transition-all duration-500 ${item.color}`}
                            style={{ width: `${percentage}%` }}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              </section>

              <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/2 sm:p-6">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <h3 className="font-semibold tracking-tight">Recent failures</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Failed or timed-out runs that need attention.
                    </p>
                  </div>
                  <AlertCircle className="size-4 text-destructive" />
                </div>

                {loading && !data.summary ? (
                  <p className="mt-5 text-sm text-muted-foreground">Loading execution summary…</p>
                ) : !data.summary ? (
                  <p className="mt-5 rounded-lg bg-muted/40 p-3 text-sm text-muted-foreground">Execution summary is unavailable.</p>
                ) : recentFailures.length === 0 ? (
                  <p className="mt-5 rounded-lg bg-muted/40 p-3 text-sm text-muted-foreground">
                    No failed or timed-out executions were recorded.
                  </p>
                ) : (
                  <div className="mt-4 divide-y divide-border/70">
                    {recentFailures.slice(0, 3).map((execution) => (
                      <Link key={execution.id} href={`/executions/${execution.id}`} className="block py-3 first:pt-0 last:pb-0 hover:text-primary">
                        <div className="flex items-center justify-between gap-2">
                          <span className="truncate text-sm font-medium">{execution.task_name}</span>
                          <ExecutionStatusBadge status={execution.status} />
                        </div>
                        <p className="mt-1 truncate text-xs text-muted-foreground">
                          {execution.error_message || `Scheduled ${formatDate(execution.scheduled_at)}`}
                        </p>
                      </Link>
                    ))}
                  </div>
                )}

                <div className="mt-5 border-t border-border/70 pt-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <h4 className="text-sm font-semibold">Upcoming work</h4>
                      <p className="mt-1 text-xs text-muted-foreground">Next active scheduled triggers.</p>
                    </div>
                    <Clock3 className="size-4 text-muted-foreground" />
                  </div>
                  {loading && !data.summary ? (
                    <p className="mt-3 text-sm text-muted-foreground">Loading schedule…</p>
                  ) : !data.summary ? (
                    <p className="mt-3 text-sm text-muted-foreground">Schedule summary is unavailable.</p>
                  ) : upcomingExecutions.length === 0 ? (
                    <p className="mt-3 text-sm text-muted-foreground">No upcoming runs are scheduled.</p>
                  ) : (
                    <ul className="mt-3 space-y-2">
                      {upcomingExecutions.slice(0, 3).map((run) => (
                        <li key={run.trigger_id} className="flex items-start justify-between gap-3 text-sm">
                          <span className="min-w-0 truncate">{run.task_name}</span>
                          <time className="shrink-0 text-right text-xs text-muted-foreground">{formatDateInTimezone(run.next_run_at, run.timezone)}</time>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </section>
            </div>
          </div>

          <footer className="mt-9 flex flex-col gap-2 border-t border-border/70 py-5 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <p>STAQ · Your automation workspace</p>
            <p className="flex items-center gap-1.5">
              <ShieldCheck className="size-3.5" />
              Authenticated workspace
            </p>
          </footer>
          {createTaskOpen && (
            <TaskCreationWizard
              queues={data.queues}
              onClose={() => setCreateTaskOpen(false)}
              onCreated={() => {
                setCreateTaskOpen(false);
                void loadDashboard();
              }}
            />
          )}
        </main>
      </div>
    </div>
  );
}

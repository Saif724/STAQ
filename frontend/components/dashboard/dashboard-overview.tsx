"use client";

import { useCallback, useEffect, useState, useSyncExternalStore } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  CircleHelp,
  Clock3,
  Command,
  LayoutDashboard,
  Layers3,
  ListTodo,
  LogOut,
  Menu,
  Moon,
  Plus,
  RefreshCw,
  Settings2,
  ShieldCheck,
  Sun,
  Workflow,
  X,
} from "lucide-react";

import { ApiError } from "@/lib/api/client";
import {
  createTask,
  getCurrentUser,
  getQueues,
  getTasks,
  type CurrentUser,
  type Queue,
  type Task,
} from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";
import { Button } from "@/components/ui/button";

type DashboardData = {
  user: CurrentUser | null;
  tasks: Task[];
  queues: Queue[];
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
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(date);
}

function StatusBadge({ status }: { status: Task["status"] }) {
  const styles: Record<Task["status"], string> = {
    Active:
      "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/50 dark:text-emerald-300",
    Paused:
      "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900 dark:bg-amber-950/50 dark:text-amber-300",
    Archived: "border-border bg-muted text-muted-foreground",
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium ${styles[status]}`}
    >
      <span
        className={`size-1.5 rounded-full ${
          status === "Active"
            ? "bg-emerald-500"
            : status === "Paused"
              ? "bg-amber-500"
              : "bg-muted-foreground"
        }`}
      />
      {status}
    </span>
  );
}

function StatCard({
  label,
  value,
  description,
  icon: Icon,
  accent,
}: {
  label: string;
  value: number;
  description: string;
  icon: typeof ListTodo;
  accent: string;
}) {
  return (
    <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/[0.02] transition-shadow hover:shadow-md">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className="mt-3 text-3xl font-semibold tracking-tight">{value}</p>
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

function Sidebar({
  mobileOpen,
  onClose,
  user,
  onLogout,
  loggingOut,
}: {
  mobileOpen: boolean;
  onClose: () => void;
  user: CurrentUser | null;
  onLogout: () => void;
  loggingOut: boolean;
}) {
  const router = useRouter();

  return (
    <>
      {mobileOpen && (
        <button
          type="button"
          aria-label="Close navigation"
          onClick={onClose}
          className="fixed inset-0 z-40 bg-black/40 lg:hidden"
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-[264px] flex-col border-r border-border/80 bg-card transition-transform duration-200 lg:translate-x-0 ${
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex h-[76px] items-center justify-between border-b border-border/70 px-6">
          <button
            type="button"
            onClick={() => router.push("/dashboard")}
            className="flex items-center gap-3"
          >
            <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <Workflow className="size-5" />
            </span>

            <span className="text-left">
              <span className="block text-lg font-bold tracking-tight">
                STAQ
              </span>
              <span className="block text-[10px] font-medium uppercase tracking-[0.2em] text-muted-foreground">
                Automation workspace
              </span>
            </span>
          </button>

          <button
            type="button"
            aria-label="Close menu"
            onClick={onClose}
            className="rounded-lg p-2 text-muted-foreground hover:bg-muted lg:hidden"
          >
            <X className="size-4" />
          </button>
        </div>

        <div className="px-4 pt-6">
          <p className="mb-3 px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
            Workspace
          </p>

          <button
            type="button"
            onClick={() => {
              router.push("/dashboard");
              onClose();
            }}
            className="flex w-full items-center gap-3 rounded-xl bg-primary/10 px-3 py-2.5 text-sm font-semibold text-primary"
          >
            <LayoutDashboard className="size-[18px]" />
            Overview
            <span className="ml-auto size-1.5 rounded-full bg-primary" />
          </button>

          <div className="mt-1 space-y-1">
            {[
              { label: "Tasks", icon: ListTodo },
              { label: "Queues", icon: Layers3 },
              { label: "Connections", icon: ShieldCheck },
            ].map(({ label, icon: Icon }) => (
              <div
                key={label}
                title={`${label} page has not been implemented yet`}
                className="flex cursor-not-allowed items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground/70"
              >
                <Icon className="size-[18px]" />
                {label}
                <span className="ml-auto text-[10px] font-medium uppercase tracking-wide">
                  Soon
                </span>
              </div>
            ))}
          </div>
        </div>

        <div className="mx-4 mt-7 border-t border-border/70" />

        <div className="px-4 pt-5">
          <p className="mb-3 px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
            Preferences
          </p>

          <div className="flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground">
            <Settings2 className="size-[18px]" />
            Workspace settings
            <span className="ml-auto text-[10px]">Soon</span>
          </div>

          <div className="flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground">
            <CircleHelp className="size-[18px]" />
            Help &amp; support
          </div>
        </div>

        <div className="mt-auto p-4">
          <div className="rounded-2xl border border-border/80 bg-muted/40 p-3.5">
            <div className="flex items-center gap-2">
              <span className="flex size-8 items-center justify-center rounded-lg bg-background text-primary">
                <Activity className="size-4" />
              </span>
              <div>
                <p className="text-xs font-semibold">Your workspace</p>
                <p className="text-[11px] text-muted-foreground">
                  Personal automation
                </p>
              </div>
            </div>

            <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
              Build reliable workflows, one task at a time.
            </p>
          </div>

          <div className="mt-4 flex items-center gap-3 border-t border-border/70 pt-4">
            <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
              {getInitials(user?.full_name ?? "User")}
            </div>

            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-semibold">
                {user?.full_name ?? "Your account"}
              </p>
              <p className="truncate text-xs text-muted-foreground">
                {user?.email ?? ""}
              </p>
            </div>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              disabled={loggingOut}
              onClick={onLogout}
              aria-label="Log out"
              title="Log out"
            >
              <LogOut className="size-4" />
            </Button>
          </div>
        </div>
      </aside>
    </>
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
  });
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  const [createTaskOpen, setCreateTaskOpen] = useState(false);
  const [creatingTask, setCreatingTask] = useState(false);
  const [createTaskError, setCreateTaskError] = useState("");

  const [taskName, setTaskName] = useState("");
  const [taskDescription, setTaskDescription] = useState("");
  const [taskQueueId, setTaskQueueId] = useState("");
  const [taskTimeout, setTaskTimeout] = useState("30");
  const [taskRetries, setTaskRetries] = useState("2");

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
      const [userResult, taskResult, queueResult] = results;

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

      setErrors(nextErrors);
    } catch {
      setErrors(["Unable to load dashboard data. Please try again."]);
    } finally {
      setRefreshing(false);
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    let cancelled = false;

    async function initialLoad() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }

      try {
        const results = await Promise.allSettled([
          getCurrentUser(),
          getTasks(),
          getQueues(),
        ]);

        if (cancelled) return;

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
        const [userResult, taskResult, queueResult] = results;

        setData((previous) => {
          const next = { ...previous };

          if (userResult.status === "fulfilled") {
            next.user = userResult.value.data;
          } else {
            nextErrors.push("Your profile could not be loaded.");
          }

          if (taskResult.status === "fulfilled") {
            next.tasks = taskResult.value.data;
          } else {
            nextErrors.push("Tasks could not be loaded.");
          }

          if (queueResult.status === "fulfilled") {
            next.queues = queueResult.value.data;
          } else {
            nextErrors.push("Queues could not be loaded.");
          }

          return next;
        });

        setErrors(nextErrors);
      } catch {
        if (!cancelled) {
          setErrors(["Unable to load dashboard data. Please try again."]);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void initialLoad();

    return () => {
      cancelled = true;
    };
  }, [router]);

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

  const activeQueues = data.queues.filter((queue) => queue.is_active).length;

  const recentTasks = [...data.tasks]
    .sort(
      (a, b) =>
        new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
    )
    .slice(0, 5);

  const firstName = data.user?.full_name?.trim().split(/\s+/)[0];

  async function handleCreateTask(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCreateTaskError("");

    const name = taskName.trim();

    if (!name) {
      setCreateTaskError("Task name is required.");
      return;
    }

    if (!taskQueueId) {
      setCreateTaskError("Please select a queue.");
      return;
    }

    const timeout = Number(taskTimeout);
    const retries = Number(taskRetries);

    if (!Number.isInteger(timeout) || timeout < 1) {
      setCreateTaskError("Timeout must be a positive whole number.");
      return;
    }

    if (!Number.isInteger(retries) || retries < 0) {
      setCreateTaskError("Retries must be zero or a positive whole number.");
      return;
    }

    setCreatingTask(true);

    try {
      const response = await createTask({
        queue_id: taskQueueId,
        name,
        description: taskDescription.trim() || null,
        timeout_seconds: timeout,
        max_retries: retries,
      });

      setData((previous) => ({
        ...previous,
        tasks: [
          response.data,
          ...previous.tasks.filter((task) => task.id !== response.data.id),
        ],
      }));

      setTaskName("");
      setTaskDescription("");
      setTaskQueueId("");
      setTaskTimeout("30");
      setTaskRetries("2");
      setCreateTaskOpen(false);
      setCreateTaskError("");
    } catch (error) {
      setCreateTaskError(
        error instanceof Error
          ? error.message
          : "Could not create the task. Please try again.",
      );
    } finally {
      setCreatingTask(false);
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar
        mobileOpen={mobileOpen}
        onClose={() => setMobileOpen(false)}
        user={data.user}
        onLogout={handleLogout}
        loggingOut={loggingOut}
      />

      <div className="min-h-screen lg:pl-[264px]">
        <header className="sticky top-0 z-30 flex h-[76px] items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-9">
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

        <main className="mx-auto max-w-[1440px] px-4 py-7 sm:px-6 sm:py-9 lg:px-9">
          <section className="relative overflow-hidden rounded-3xl border border-primary/15 bg-gradient-to-br from-primary/[0.09] via-card to-card p-6 sm:p-8 lg:p-9">
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
                      setCreateTaskError("");
                      setCreateTaskOpen(true);
                    }}
                    className="gap-2"
                  >
                    <Plus className="size-4" />
                    Create a task
                  </Button>

                  <span className="text-xs text-muted-foreground">
                    Task creation UI is coming next.
                  </span>
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

            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
              <StatCard
                label="Total tasks"
                value={data.tasks.length}
                description="All tasks in your workspace"
                icon={ListTodo}
                accent="bg-primary/10 text-primary"
              />

              <StatCard
                label="Active tasks"
                value={activeTasks}
                description="Tasks currently marked active"
                icon={Activity}
                accent="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              />

              <StatCard
                label="Paused tasks"
                value={pausedTasks}
                description="Tasks currently on hold"
                icon={Clock3}
                accent="bg-amber-500/10 text-amber-600 dark:text-amber-400"
              />

              <StatCard
                label="Active queues"
                value={activeQueues}
                description={`${data.queues.length} total queues configured`}
                icon={Layers3}
                accent="bg-violet-500/10 text-violet-600 dark:text-violet-400"
              />
            </div>
          </section>

          <div className="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1.6fr)_minmax(300px,0.8fr)]">
            <section className="min-w-0 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/[0.02]">
              <div className="flex items-center justify-between border-b border-border/70 px-5 py-5 sm:px-6">
                <div>
                  <h3 className="font-semibold tracking-tight">
                    Recently updated tasks
                  </h3>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Your latest task changes.
                  </p>
                </div>

                <div className="flex size-9 items-center justify-center rounded-xl bg-muted">
                  <ListTodo className="size-4 text-muted-foreground" />
                </div>
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
              ) : recentTasks.length === 0 ? (
                <div className="flex flex-col items-center px-6 py-12 text-center">
                  <div className="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <ListTodo className="size-5" />
                  </div>
                  <h4 className="mt-4 text-sm font-semibold">No tasks yet</h4>
                  <p className="mt-1 max-w-xs text-sm leading-6 text-muted-foreground">
                    When you create tasks, their latest updates will appear
                    here.
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-border/70">
                  {recentTasks.map((task) => (
                    <div
                      key={task.id}
                      className="flex items-center gap-3 px-5 py-4 transition-colors hover:bg-muted/30 sm:px-6"
                    >
                      <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/8 text-primary">
                        <Workflow className="size-[18px]" />
                      </div>

                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">
                          {task.name}
                        </p>
                        <p className="mt-1 truncate text-xs text-muted-foreground">
                          Updated {formatDate(task.updated_at)}
                        </p>
                      </div>

                      <StatusBadge status={task.status} />
                    </div>
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
                  Sorted by last update
                </span>
              </div>
            </section>

            <div className="space-y-6">
              <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/[0.02] sm:p-6">
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

              <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/[0.02] sm:p-6">
                <div className="flex items-center gap-3">
                  <div className="flex size-10 items-center justify-center rounded-xl bg-violet-500/10 text-violet-600 dark:text-violet-400">
                    <Layers3 className="size-5" />
                  </div>
                  <div>
                    <h3 className="font-semibold tracking-tight">
                      Queue overview
                    </h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Your task organization.
                    </p>
                  </div>
                </div>

                <div className="mt-5 flex items-end justify-between">
                  <div>
                    <p className="text-3xl font-semibold tracking-tight">
                      {data.queues.length}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Configured queues
                    </p>
                  </div>

                  <div className="text-right">
                    <p className="text-lg font-semibold text-emerald-600 dark:text-emerald-400">
                      {activeQueues}
                    </p>
                    <p className="text-xs text-muted-foreground">Active</p>
                  </div>
                </div>

                {data.queues.length === 0 && !loading ? (
                  <p className="mt-4 rounded-xl bg-muted/50 p-3 text-xs leading-5 text-muted-foreground">
                    No queues are available yet. Queues organize the tasks that
                    belong to your automation workflows.
                  </p>
                ) : (
                  <div className="mt-4 space-y-2">
                    {data.queues.slice(0, 3).map((queue) => (
                      <div
                        key={queue.id}
                        className="flex items-center gap-2 rounded-xl border border-border/70 px-3 py-2.5"
                      >
                        <span
                          className={`size-2 rounded-full ${
                            queue.is_active
                              ? "bg-emerald-500"
                              : "bg-muted-foreground"
                          }`}
                        />
                        <span className="min-w-0 flex-1 truncate text-sm">
                          {queue.name}
                        </span>
                        <span className="text-[10px] text-muted-foreground">
                          {queue.is_active ? "Active" : "Inactive"}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
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
            <div
              className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-black/50 p-4 backdrop-blur-sm"
              onMouseDown={(event) => {
                if (event.target === event.currentTarget && !creatingTask) {
                  setCreateTaskOpen(false);
                }
              }}
            >
              <section
                role="dialog"
                aria-modal="true"
                aria-labelledby="create-task-title"
                className="my-auto w-full max-w-lg rounded-2xl border border-border bg-card p-6 shadow-2xl"
              >
                <div className="mb-6 flex items-start justify-between gap-4">
                  <div>
                    <h2
                      id="create-task-title"
                      className="text-xl font-semibold tracking-tight"
                    >
                      Create a task
                    </h2>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Configure a task for your automation queue.
                    </p>
                  </div>

                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    aria-label="Close dialog"
                    disabled={creatingTask}
                    onClick={() => setCreateTaskOpen(false)}
                  >
                    <X className="size-4" />
                  </Button>
                </div>

                <form onSubmit={handleCreateTask} className="space-y-5">
                  <div className="space-y-2">
                    <label htmlFor="task-name" className="text-sm font-medium">
                      Task name <span className="text-destructive">*</span>
                    </label>
                    <input
                      id="task-name"
                      autoFocus
                      required
                      maxLength={120}
                      value={taskName}
                      onChange={(event) => setTaskName(event.target.value)}
                      placeholder="e.g. Generate daily report"
                      className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                    />
                  </div>

                  <div className="space-y-2">
                    <label
                      htmlFor="task-description"
                      className="text-sm font-medium"
                    >
                      Description
                    </label>
                    <textarea
                      id="task-description"
                      rows={3}
                      maxLength={1000}
                      value={taskDescription}
                      onChange={(event) =>
                        setTaskDescription(event.target.value)
                      }
                      placeholder="What should this task accomplish?"
                      className="w-full resize-y rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                    />
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="task-queue" className="text-sm font-medium">
                      Queue <span className="text-destructive">*</span>
                    </label>
                    <select
                      id="task-queue"
                      required
                      value={taskQueueId}
                      onChange={(event) => setTaskQueueId(event.target.value)}
                      className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                    >
                      <option value="">Select a queue</option>
                      {data.queues
                        .filter((queue) => queue.is_active)
                        .map((queue) => (
                          <option key={queue.id} value={queue.id}>
                            {queue.name}
                          </option>
                        ))}
                    </select>

                    {data.queues.filter((queue) => queue.is_active).length ===
                      0 && (
                      <p className="text-xs text-amber-600 dark:text-amber-400">
                        No active queues are available. Create or activate a
                        queue before creating a task.
                      </p>
                    )}
                  </div>

                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <label
                        htmlFor="task-timeout"
                        className="text-sm font-medium"
                      >
                        Timeout (seconds)
                      </label>
                      <input
                        id="task-timeout"
                        type="number"
                        min={1}
                        step={1}
                        required
                        value={taskTimeout}
                        onChange={(event) => setTaskTimeout(event.target.value)}
                        className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                      />
                    </div>

                    <div className="space-y-2">
                      <label
                        htmlFor="task-retries"
                        className="text-sm font-medium"
                      >
                        Maximum retries
                      </label>
                      <input
                        id="task-retries"
                        type="number"
                        min={0}
                        step={1}
                        required
                        value={taskRetries}
                        onChange={(event) => setTaskRetries(event.target.value)}
                        className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                      />
                    </div>
                  </div>

                  {createTaskError && (
                    <div
                      role="alert"
                      className="rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive"
                    >
                      {createTaskError}
                    </div>
                  )}

                  <div className="flex justify-end gap-3 border-t border-border pt-4">
                    <Button
                      type="button"
                      variant="outline"
                      disabled={creatingTask}
                      onClick={() => setCreateTaskOpen(false)}
                    >
                      Cancel
                    </Button>

                    <Button
                      type="submit"
                      disabled={
                        creatingTask ||
                        data.queues.filter((queue) => queue.is_active)
                          .length === 0
                      }
                      className="gap-2"
                    >
                      {creatingTask && (
                        <RefreshCw className="size-4 animate-spin" />
                      )}
                      {creatingTask ? "Creating..." : "Create task"}
                    </Button>
                  </div>
                </form>
              </section>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

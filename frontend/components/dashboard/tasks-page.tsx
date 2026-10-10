"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  AlertCircle,
  Clock3,
  ListTodo,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Workflow,
  X,
} from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
  createTask,
  getCurrentUser,
  getQueues,
  getTasks,
  type CurrentUser,
  type Queue,
  type Task,
  type TaskStatus,
} from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

type StatusFilter = "All" | TaskStatus;

function formatDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) return "Unknown date";

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

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

function StatusBadge({ status }: { status: TaskStatus }) {
  const styles: Record<TaskStatus, string> = {
    Active:
      "border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
    Paused:
      "border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-400",
    Archived: "border-border bg-muted text-muted-foreground",
  };

  return (
    <span
      className={`inline-flex shrink-0 items-center rounded-full border px-2.5 py-1 text-xs font-medium ${styles[status]}`}
    >
      <span className="mr-1.5 size-1.5 rounded-full bg-current" />
      {status}
    </span>
  );
}

export default function TasksPage() {
  const router = useRouter();

  const [user, setUser] = useState<CurrentUser | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [queues, setQueues] = useState<Queue[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [error, setError] = useState("");

  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("All");

  const [createTaskOpen, setCreateTaskOpen] = useState(false);
  const [creatingTask, setCreatingTask] = useState(false);
  const [createTaskError, setCreateTaskError] = useState("");

  const [taskName, setTaskName] = useState("");
  const [taskDescription, setTaskDescription] = useState("");
  const [taskQueueId, setTaskQueueId] = useState("");
  const [taskTimeout, setTaskTimeout] = useState("30");
  const [taskRetries, setTaskRetries] = useState("2");

  const loadData = useCallback(
    async (showRefresh = false) => {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }

      if (showRefresh) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }

      setError("");

      try {
        const [userResult, tasksResult, queuesResult] =
          await Promise.allSettled([getCurrentUser(), getTasks(), getQueues()]);

        const results = [userResult, tasksResult, queuesResult];

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

        const failures: string[] = [];

        if (userResult.status === "fulfilled") {
          setUser(userResult.value.data);
        } else {
          failures.push("Your profile could not be loaded.");
        }

        if (tasksResult.status === "fulfilled") {
          setTasks(tasksResult.value.data);
        } else {
          failures.push("Tasks could not be loaded.");
        }

        if (queuesResult.status === "fulfilled") {
          setQueues(queuesResult.value.data);
        } else {
          failures.push("Queues could not be loaded.");
        }

        if (failures.length > 0) {
          setError(failures.join(" "));
        }
      } catch {
        setError("Unable to load your tasks. Please try again.");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [router],
  );

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void loadData();
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [loadData]);

  async function handleLogout() {
    setLoggingOut(true);

    try {
      await logoutUser();
    } catch {
      // Clear the local session even if the logout request fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  const filteredTasks = useMemo(() => {
    const query = search.trim().toLowerCase();

    return [...tasks]
      .filter((task) => {
        if (statusFilter !== "All" && task.status !== statusFilter) {
          return false;
        }

        if (!query) return true;

        return (
          task.name.toLowerCase().includes(query) ||
          (task.description ?? "").toLowerCase().includes(query)
        );
      })
      .sort(
        (a, b) =>
          new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
      );
  }, [tasks, search, statusFilter]);

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

      setTasks((previous) => [
        response.data,
        ...previous.filter((task) => task.id !== response.data.id),
      ]);

      setTaskName("");
      setTaskDescription("");
      setTaskQueueId("");
      setTaskTimeout("30");
      setTaskRetries("2");
      setCreateTaskOpen(false);
      setCreateTaskError("");
      setSearch("");
      setStatusFilter("All");
    } catch (err) {
      setCreateTaskError(
        err instanceof Error
          ? err.message
          : "Could not create the task. Please try again.",
      );
    } finally {
      setCreatingTask(false);
    }
  }

  const activeQueues = queues.filter((queue) => queue.is_active);
  const activeCount = tasks.filter((task) => task.status === "Active").length;
  const pausedCount = tasks.filter((task) => task.status === "Paused").length;
  const archivedCount = tasks.filter(
    (task) => task.status === "Archived",
  ).length;

  const filters: { label: StatusFilter; count: number }[] = [
    { label: "All", count: tasks.length },
    { label: "Active", count: activeCount },
    { label: "Paused", count: pausedCount },
    { label: "Archived", count: archivedCount },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Sidebar
        mobileOpen={mobileOpen}
        onClose={() => setMobileOpen(false)}
        user={user}
        onLogout={handleLogout}
        loggingOut={loggingOut}
        activePage="tasks"
      />

      <main className="min-h-screen lg:pl-[264px]">
        <header className="sticky top-0 z-30 flex h-[76px] items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setMobileOpen(true)}
              aria-label="Open navigation"
              className="rounded-lg border border-border p-2 text-muted-foreground hover:bg-muted lg:hidden"
            >
              <ListTodo className="size-4" />
            </button>

            <div>
              <p className="text-sm font-semibold">Tasks</p>
              <p className="text-xs text-muted-foreground">
                Manage your automation tasks
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => void loadData(true)}
              disabled={loading || refreshing}
              className="gap-2"
            >
              <RefreshCw
                className={`size-3.5 ${refreshing ? "animate-spin" : ""}`}
              />
              <span className="hidden sm:inline">Refresh</span>
            </Button>

            <Button
              type="button"
              size="sm"
              onClick={() => {
                setCreateTaskError("");
                setCreateTaskOpen(true);
              }}
              className="gap-2"
            >
              <Plus className="size-4" />
              <span className="hidden sm:inline">Create task</span>
              <span className="sm:hidden">Create</span>
            </Button>
          </div>
        </header>

        <div className="mx-auto w-full max-w-[1500px] px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/[0.02] sm:p-7">
            <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-center">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                  Workspace
                </p>
                <h1 className="mt-2 text-2xl font-bold tracking-tight sm:text-3xl">
                  Your tasks
                </h1>
                <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
                  Find, organize, and configure the tasks in your automation
                  workspace.
                </p>
              </div>

              <div className="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                <Workflow className="size-7" />
              </div>
            </div>

            <div className="mt-6 grid gap-3 sm:grid-cols-3">
              {[
                {
                  label: "Total tasks",
                  value: tasks.length,
                  icon: ListTodo,
                },
                {
                  label: "Active",
                  value: activeCount,
                  icon: Workflow,
                },
                {
                  label: "Paused",
                  value: pausedCount,
                  icon: Clock3,
                },
              ].map(({ label, value, icon: Icon }) => (
                <div
                  key={label}
                  className="flex items-center gap-3 rounded-xl border border-border/70 bg-background/70 p-4"
                >
                  <div className="flex size-10 items-center justify-center rounded-xl bg-muted text-muted-foreground">
                    <Icon className="size-5" />
                  </div>
                  <div>
                    <p className="text-xl font-semibold tabular-nums">
                      {loading ? "—" : value}
                    </p>
                    <p className="text-xs text-muted-foreground">{label}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {error && (
            <div
              role="alert"
              className="mt-6 flex flex-col gap-3 rounded-2xl border border-amber-300/70 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="flex items-start gap-2.5">
                <AlertCircle className="mt-0.5 size-4 shrink-0" />
                <p>{error}</p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => void loadData(true)}
                disabled={refreshing}
              >
                Try again
              </Button>
            </div>
          )}

          <section className="mt-7 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/[0.02]">
            <div className="border-b border-border/70 p-4 sm:p-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <h2 className="font-semibold tracking-tight">All tasks</h2>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {loading
                      ? "Loading your tasks…"
                      : `${filteredTasks.length} of ${tasks.length} tasks shown`}
                  </p>
                </div>

                <div className="relative w-full lg:max-w-sm">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <input
                    type="search"
                    aria-label="Search tasks"
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="Search by name or description…"
                    className="h-10 w-full rounded-xl border border-input bg-background pl-9 pr-3 text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
              </div>

              <div className="mt-4 flex flex-wrap gap-2">
                {filters.map((filter) => {
                  const selected = statusFilter === filter.label;

                  return (
                    <button
                      key={filter.label}
                      type="button"
                      onClick={() => setStatusFilter(filter.label)}
                      aria-pressed={selected}
                      className={`inline-flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-medium transition-colors ${
                        selected
                          ? "border-primary/30 bg-primary/10 text-primary"
                          : "border-border bg-background text-muted-foreground hover:bg-muted"
                      }`}
                    >
                      {filter.label}
                      <span
                        className={`rounded-md px-1.5 py-0.5 tabular-nums ${
                          selected ? "bg-primary/10" : "bg-muted"
                        }`}
                      >
                        {filter.count}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>

            {loading ? (
              <div className="space-y-4 p-6">
                {[1, 2, 3, 4].map((item) => (
                  <div key={item} className="flex animate-pulse gap-3">
                    <div className="size-10 rounded-xl bg-muted" />
                    <div className="flex-1 space-y-2 py-1">
                      <div className="h-3 w-2/5 rounded bg-muted" />
                      <div className="h-3 w-1/3 rounded bg-muted" />
                    </div>
                  </div>
                ))}
              </div>
            ) : filteredTasks.length === 0 ? (
              <div className="flex flex-col items-center px-6 py-14 text-center">
                <div className="flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                  {search || statusFilter !== "All" ? (
                    <Search className="size-6" />
                  ) : (
                    <ListTodo className="size-6" />
                  )}
                </div>

                <h3 className="mt-4 text-base font-semibold">
                  {search || statusFilter !== "All"
                    ? "No matching tasks"
                    : "No tasks yet"}
                </h3>

                <p className="mt-2 max-w-sm text-sm leading-6 text-muted-foreground">
                  {search || statusFilter !== "All"
                    ? "Try a different search term or status filter."
                    : "Create your first task to start organizing your automation workflow."}
                </p>

                {!search && statusFilter === "All" && (
                  <Button
                    type="button"
                    className="mt-5 gap-2"
                    onClick={() => setCreateTaskOpen(true)}
                  >
                    <Plus className="size-4" />
                    Create your first task
                  </Button>
                )}
              </div>
            ) : (
              <div className="divide-y divide-border/70">
                {filteredTasks.map((task) => {
                  const queue = queues.find(
                    (item) => item.id === task.queue_id,
                  );

                  return (
                    <article
                      key={task.id}
                      className="flex flex-col gap-4 px-4 py-5 transition-colors hover:bg-muted/20 sm:px-5 md:flex-row md:items-center md:justify-between"
                    >
                      <div className="flex min-w-0 items-start gap-3">
                        <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                          <Workflow className="size-[18px]" />
                        </div>

                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-2">
                            <h3 className="break-words text-sm font-semibold">
                              {task.name}
                            </h3>
                            <StatusBadge status={task.status} />
                          </div>

                          <p className="mt-1 break-words text-sm leading-5 text-muted-foreground">
                            {task.description || "No description provided."}
                          </p>

                          <div className="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-xs text-muted-foreground">
                            <span>
                              Queue:{" "}
                              <span className="font-medium text-foreground">
                                {queue?.name ?? "Unknown queue"}
                              </span>
                            </span>
                            <span>Timeout: {task.timeout_seconds}s</span>
                            <span>Retries: {task.max_retries}</span>
                          </div>

                          <p className="mt-2 text-xs text-muted-foreground">
                            Updated {formatDate(task.updated_at)}
                          </p>
                        </div>
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </section>

          <footer className="mt-8 flex flex-col gap-2 border-t border-border/70 py-5 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <p>STAQ · Your automation workspace</p>
            <p className="flex items-center gap-1.5">
              <ShieldCheck className="size-3.5" />
              Authenticated workspace
            </p>
          </footer>
        </div>
      </main>

      {createTaskOpen && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center overflow-y-auto bg-black/50 p-4 backdrop-blur-sm"
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
                  onChange={(event) => setTaskDescription(event.target.value)}
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
                  {activeQueues.map((queue) => (
                    <option key={queue.id} value={queue.id}>
                      {queue.name}
                    </option>
                  ))}
                </select>

                {activeQueues.length === 0 && (
                  <p className="text-xs text-amber-600 dark:text-amber-400">
                    No active queues are available. Create or activate a queue
                    before creating a task.
                  </p>
                )}
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label htmlFor="task-timeout" className="text-sm font-medium">
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
                  <label htmlFor="task-retries" className="text-sm font-medium">
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
                <p
                  role="alert"
                  className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive"
                >
                  {createTaskError}
                </p>
              )}

              <div className="flex flex-col-reverse gap-2 border-t border-border/70 pt-4 sm:flex-row sm:justify-end">
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
                  disabled={creatingTask || activeQueues.length === 0}
                  className="gap-2"
                >
                  <Plus className="size-4" />
                  {creatingTask ? "Creating…" : "Create task"}
                </Button>
              </div>
            </form>
          </section>
        </div>
      )}
    </div>
  );
}

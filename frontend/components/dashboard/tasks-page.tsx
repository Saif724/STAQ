"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type FormEvent,
} from "react";
import { useRouter } from "next/navigation";
import {
  AlertCircle,
  Archive,
  Check,
  Clock3,
  ListTodo,
  Pause,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Trash2,
  Workflow,
  CalendarClock,
  X,
} from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
  deleteTask,
  getCurrentUser,
  getQueues,
  getTasks,
  updateTask,
  type CurrentUser,
  type Queue,
  type Task,
  type TaskStatus,
} from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { getTaskTriggers, type Trigger } from "@/lib/api/automation";
import { logoutUser } from "@/lib/auth/session";
import TaskCreationWizard from "@/components/tasks/task-creation-wizard";

type StatusFilter = "All" | TaskStatus;

function formatDate(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) return "Unknown date";

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function getNextRun(triggers: Trigger[] | undefined, taskStatus: TaskStatus) {
  if (taskStatus !== "Active") return null;
  if (!triggers || triggers.length === 0) return null;

  const now = Date.now();
  const candidates = triggers
    .filter((trigger) => trigger.is_active)
    .map((trigger) => ({
      trigger,
      timestamp: new Date(trigger.next_run_at).getTime(),
    }))
    .filter(({ timestamp }) => Number.isFinite(timestamp) && timestamp > now)
    .sort((a, b) => a.timestamp - b.timestamp);

  return candidates[0]?.trigger ?? null;
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
  const [triggersByTask, setTriggersByTask] = useState<
    Record<string, Trigger[]>
  >({});
  const [triggerLoadFailed, setTriggerLoadFailed] = useState<
    Record<string, boolean>
  >({});
  const [queues, setQueues] = useState<Queue[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [error, setError] = useState("");

  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("All");

  const [createTaskOpen, setCreateTaskOpen] = useState(false);

  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [busyTaskId, setBusyTaskId] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<{
    task: Task;
    action: "delete" | "archive";
  } | null>(null);

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
          const loadedTasks = tasksResult.value.data;
          setTasks(loadedTasks);

          // Load schedules independently so a trigger API failure does not hide tasks.
          const triggerResults = await Promise.allSettled(
            loadedTasks.map(async (task) => ({
              taskId: task.id,
              result: await getTaskTriggers(task.id),
            })),
          );
          const nextTriggers: Record<string, Trigger[]> = {};
          const nextTriggerFailures: Record<string, boolean> = {};

          triggerResults.forEach((result, index) => {
            const taskId = loadedTasks[index].id;
            if (result.status === "fulfilled") {
              nextTriggers[result.value.taskId] = result.value.result.data;
            } else {
              nextTriggers[taskId] = [];
              nextTriggerFailures[taskId] = true;
            }
          });

          setTriggersByTask(nextTriggers);
          setTriggerLoadFailed(nextTriggerFailures);
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
      // Clear the local session even if logout fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  async function changeTaskStatus(task: Task, status: TaskStatus) {
    if (busyTaskId) return;

    setBusyTaskId(task.id);
    setError("");

    try {
      const result = await updateTask(task.id, {
        name: task.name,
        description: task.description ?? null,
        status,
        timeout_seconds: task.timeout_seconds,
        max_retries: task.max_retries,
      });

      setTasks((current) =>
        current.map((item) => (item.id === task.id ? result.data : item)),
      );
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not update the task.",
      );
    } finally {
      setBusyTaskId(null);
    }
  }

  function openEditTask(task: Task) {
    if (busyTaskId) return;

    setError("");
    setEditingTask(task);
  }

  function handleDeleteTask(task: Task) {
    if (busyTaskId) return;
    setPendingAction({ task, action: "delete" });
  }

  function handleArchiveTask(task: Task) {
    if (busyTaskId) return;
    setPendingAction({ task, action: "archive" });
  }

  async function confirmDeleteTask() {
    if (!pendingAction || pendingAction.action !== "delete") return;
    if (busyTaskId) return;

    const task = pendingAction.task;

    setBusyTaskId(task.id);
    setError("");

    try {
      await deleteTask(task.id);

      setTasks((current) => current.filter((item) => item.id !== task.id));

      if (editingTask?.id === task.id) {
        setEditingTask(null);
      }

      setPendingAction(null);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not delete the task.",
      );
    } finally {
      setBusyTaskId(null);
    }
  }

  const filteredTasks = useMemo(() => {
    const query = search.trim().toLowerCase();

    return [...tasks]
      .filter((task) => {
        // Keep archived tasks out of the default list; they remain available in Archived.
        if (statusFilter === "All" && task.status === "Archived") {
          return false;
        }

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

  const activeCount = tasks.filter((task) => task.status === "Active").length;

  const pausedCount = tasks.filter((task) => task.status === "Paused").length;

  const archivedCount = tasks.filter(
    (task) => task.status === "Archived",
  ).length;

  const filters: { label: StatusFilter; count: number }[] = [
    { label: "All", count: tasks.length - archivedCount },
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

      <main className="min-h-screen lg:pl-66">
        <header className="sticky top-0 z-30 flex h-19 items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
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

          <div className="flex shrink-0 items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => void loadData(true)}
              disabled={
                loading || refreshing || Boolean(busyTaskId)
              }
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
              onClick={() => setCreateTaskOpen(true)}
              className="gap-2"
            >
              <Plus className="size-4" />
              <span className="hidden sm:inline">Create task</span>
              <span className="sm:hidden">Create</span>
            </Button>
          </div>
        </header>

        <div className="mx-auto w-full max-w-375 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <section className="rounded-2xl border border-border/80 bg-card p-5 shadow-sm shadow-black/2 sm:p-7">
            <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-center">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                  Workspace
                </p>
                <h1 className="mt-2 text-2xl font-bold tracking-tight sm:text-3xl">
                  Your tasks
                </h1>
                <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
                  Manage task settings and control when your automations run.
                </p>
              </div>

              <div className="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                <Workflow className="size-7" />
              </div>
            </div>

            <div className="mt-6 grid gap-3 sm:grid-cols-3">
              {[
                { label: "Total tasks", value: tasks.length, icon: ListTodo },
                { label: "Active", value: activeCount, icon: Workflow },
                { label: "Paused", value: pausedCount, icon: Clock3 },
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

          <section className="mt-7 overflow-hidden rounded-2xl border border-border/80 bg-card shadow-sm shadow-black/2">
            <div className="border-b border-border/70 p-4 sm:p-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <h2 className="font-semibold tracking-tight">
                    {statusFilter === "All" ? "Tasks" : `${statusFilter} tasks`}
                  </h2>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {loading
                      ? "Loading your tasks…"
                      : `${filteredTasks.length} shown`}
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
                  const busy = busyTaskId === task.id;

                  return (
                    <article
                      key={task.id}
                      className="flex min-w-0 flex-col gap-4 px-4 py-5 transition-colors hover:bg-muted/20 sm:px-5 xl:flex-row xl:items-start xl:justify-between"
                    >
                      <div className="flex min-w-0 items-start gap-3">
                        <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                          <Workflow className="size-4.5" />
                        </div>

                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-2">
                            <h3 className="wrap-break-word text-sm font-semibold">
                              {task.name}
                            </h3>
                            <StatusBadge status={task.status} />
                            {busy && (
                              <span className="text-xs text-muted-foreground">
                                Updating…
                              </span>
                            )}
                          </div>

                          <p className="mt-1 text-sm leading-5 text-muted-foreground [overflow-wrap:anywhere]">
                            {task.description || "No description provided."}
                          </p>

                          <div className="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-xs text-muted-foreground">
                            <span>Timeout: {task.timeout_seconds}s</span>
                            <span>Retries: {task.max_retries}</span>
                          </div>

                          <p className="mt-2 text-xs text-muted-foreground">
                            Updated {formatDate(task.updated_at)}
                          </p>

                          <div className="mt-2 flex items-start gap-1.5 text-xs">
                            <CalendarClock className="mt-0.5 size-3.5 shrink-0 text-muted-foreground" />
                            <div className="min-w-0">
                              {task.status !== "Active" ? (
                                <span className="text-muted-foreground">
                                  Next run:{" "}
                                  {task.status === "Paused"
                                    ? "Paused"
                                    : "Archived"}
                                </span>
                              ) : triggerLoadFailed[task.id] ? (
                                <span className="text-muted-foreground">
                                  Next run: schedule could not be loaded
                                </span>
                              ) : !(task.id in triggersByTask) ? (
                                <span className="text-muted-foreground">
                                  Loading schedule…
                                </span>
                              ) : (
                                (() => {
                                  const nextTrigger = getNextRun(
                                    triggersByTask[task.id],
                                    task.status,
                                  );
                                  if (nextTrigger) {
                                    return (
                                      <span className="text-muted-foreground">
                                        Next run:{" "}
                                        <span className="font-medium text-foreground">
                                          {formatDate(nextTrigger.next_run_at)}
                                        </span>
                                        <span className="ml-1">
                                          ({nextTrigger.trigger_type})
                                        </span>
                                      </span>
                                    );
                                  }
                                  const hasActiveTrigger = triggersByTask[
                                    task.id
                                  ].some((trigger) => trigger.is_active);
                                  return (
                                    <span className="text-muted-foreground">
                                      Next run:{" "}
                                      {hasActiveTrigger
                                        ? "No upcoming run"
                                        : "No active schedule"}
                                    </span>
                                  );
                                })()
                              )}
                            </div>
                          </div>
                        </div>
                      </div>

                      <div className="flex min-w-0 flex-wrap items-center gap-2 xl:max-w-[45%] xl:justify-end">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => openEditTask(task)}
                          disabled={Boolean(busyTaskId)}
                          className="gap-1.5"
                        >
                          <Pencil className="size-3.5" />
                          Edit
                        </Button>

                        {task.status === "Archived" ? (
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={Boolean(busyTaskId)}
                            onClick={() =>
                              void changeTaskStatus(task, "Active")
                            }
                            className="gap-1.5"
                          >
                            <RefreshCw className="size-3.5" />
                            Restore
                          </Button>
                        ) : (
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={Boolean(busyTaskId)}
                            onClick={() =>
                              void changeTaskStatus(
                                task,
                                task.status === "Active" ? "Paused" : "Active",
                              )
                            }
                            className="gap-1.5"
                          >
                            {task.status === "Active" ? (
                              <Pause className="size-3.5" />
                            ) : (
                              <Play className="size-3.5" />
                            )}
                            {task.status === "Active" ? "Pause" : "Resume"}
                          </Button>
                        )}

                        {task.status !== "Archived" && (
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={Boolean(busyTaskId)}
                            onClick={() => handleArchiveTask(task)}
                            className="gap-1.5"
                          >
                            <Archive className="size-3.5" />
                            Archive
                          </Button>
                        )}

                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          disabled={Boolean(busyTaskId)}
                          onClick={() => void handleDeleteTask(task)}
                          className="gap-1.5 text-destructive hover:text-destructive"
                        >
                          <Trash2 className="size-3.5" />
                          Delete
                        </Button>
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
        <TaskCreationWizard
          queues={queues}
          onClose={() => setCreateTaskOpen(false)}
          onCreated={() => {
            setCreateTaskOpen(false);
            void loadData(true);
          }}
        />
      )}

      {editingTask && (
        <TaskCreationWizard
          queues={queues}
          task={editingTask}
          onClose={() => setEditingTask(null)}
          onCreated={() => {
            setEditingTask(null);
            void loadData(true);
          }}
        />
      )}

      {pendingAction && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-black/55 p-4 backdrop-blur-sm"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget && !busyTaskId) {
              setPendingAction(null);
            }
          }}
        >
          <section
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="task-confirm-title"
            aria-describedby="task-confirm-description"
            className="my-auto w-full max-w-md overflow-hidden rounded-2xl border border-border bg-card shadow-2xl"
          >
            <div className="p-5 sm:p-6">
              <div className="flex items-start gap-4">
                <div
                  className={`flex size-11 shrink-0 items-center justify-center rounded-xl ${
                    pendingAction.action === "delete"
                      ? "bg-destructive/10 text-destructive"
                      : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
                  }`}
                >
                  {pendingAction.action === "delete" ? (
                    <Trash2 className="size-5" />
                  ) : (
                    <Archive className="size-5" />
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <h2 id="task-confirm-title" className="text-lg font-semibold">
                    {pendingAction.action === "delete"
                      ? "Delete task?"
                      : "Archive task?"}
                  </h2>

                  <p
                    id="task-confirm-description"
                    className="mt-2 text-sm leading-6 text-muted-foreground"
                  >
                    {pendingAction.action === "delete"
                      ? "This permanently deletes the task. This action cannot be undone."
                      : "This archives the task, and the scheduler will no longer dispatch it."}
                  </p>

                  <div className="mt-4 rounded-xl border border-border bg-muted/40 p-3">
                    <p className="text-xs font-medium text-muted-foreground">
                      Selected task
                    </p>
                    <p className="mt-1 break-words text-sm font-semibold text-foreground">
                      {pendingAction.task.name}
                    </p>
                  </div>
                </div>

                <button
                  type="button"
                  aria-label="Close confirmation"
                  disabled={Boolean(busyTaskId)}
                  onClick={() => setPendingAction(null)}
                  className="rounded-lg p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-50"
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>

            <div className="flex flex-col-reverse gap-2 border-t border-border bg-muted/20 p-4 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                disabled={Boolean(busyTaskId)}
                onClick={() => setPendingAction(null)}
                className="w-full sm:w-auto"
              >
                Cancel
              </Button>

              <Button
                type="button"
                disabled={Boolean(busyTaskId)}
                onClick={() => {
                  if (pendingAction.action === "delete") {
                    void confirmDeleteTask();
                  } else {
                    const task = pendingAction.task;
                    setPendingAction(null);
                    void changeTaskStatus(task, "Archived");
                  }
                }}
                className={`w-full gap-2 sm:w-auto ${
                  pendingAction.action === "delete"
                    ? "bg-destructive text-destructive-foreground hover:bg-destructive/90"
                    : ""
                }`}
              >
                {pendingAction.action === "delete" ? (
                  <Trash2 className="size-4" />
                ) : (
                  <Archive className="size-4" />
                )}
                {busyTaskId === pendingAction.task.id
                  ? "Working…"
                  : pendingAction.action === "delete"
                    ? "Delete task"
                    : "Archive task"}
              </Button>
            </div>
          </section>
        </div>
      )}
    </div>
  );
}

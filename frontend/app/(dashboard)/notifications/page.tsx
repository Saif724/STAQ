"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AlertCircle, Bell, CheckCheck, RefreshCw } from "lucide-react";

import DashboardShell from "@/components/dashboard/dashboard-shell";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
  getNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  type AppNotification,
} from "@/lib/api/notifications";
import { clearTokens } from "@/lib/auth/storage";

type Filter = "all" | "unread";

function timeAgo(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const seconds = Math.round((Date.now() - date.getTime()) / 1000);
  if (seconds < 60) return "just now";

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} min ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} h ago`;

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function announceChange() {
  window.dispatchEvent(new Event("notifications:changed"));
}

export default function NotificationsPage() {
  const router = useRouter();

  const [items, setItems] = useState<AppNotification[]>([]);
  const [unread, setUnread] = useState(0);
  const [filter, setFilter] = useState<Filter>("all");
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [markingAll, setMarkingAll] = useState(false);
  const [busyId, setBusyId] = useState("");
  const [error, setError] = useState("");

  const handleError = useCallback(
    (caught: unknown, fallback: string) => {
      if (caught instanceof ApiError && caught.status === 401) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setError(caught instanceof Error ? caught.message : fallback);
    },
    [router],
  );

  const load = useCallback(
    async (showRefresh = false) => {
      if (showRefresh) setRefreshing(true);
      else setLoading(true);
      setError("");

      try {
        const response = await getNotifications(50);
        setItems(response.data.items);
        setUnread(response.data.unread_count);
      } catch (caught) {
        handleError(caught, "Could not load notifications.");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [handleError],
  );

  useEffect(() => {
    const timeoutId = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timeoutId);
  }, [load]);

  async function handleMarkRead(item: AppNotification) {
    if (item.read_at || busyId) return;

    setBusyId(item.id);
    setError("");

    try {
      await markNotificationRead(item.id);
      const now = new Date().toISOString();
      setItems((current) =>
        current.map((n) => (n.id === item.id ? { ...n, read_at: now } : n)),
      );
      setUnread((count) => Math.max(0, count - 1));
      announceChange();
    } catch (caught) {
      handleError(caught, "Could not mark this notification as read.");
    } finally {
      setBusyId("");
    }
  }

  async function handleMarkAll() {
    if (unread === 0 || markingAll) return;

    setMarkingAll(true);
    setError("");

    try {
      await markAllNotificationsRead();
      const now = new Date().toISOString();
      setItems((current) =>
        current.map((n) => (n.read_at ? n : { ...n, read_at: now })),
      );
      setUnread(0);
      announceChange();
    } catch (caught) {
      handleError(caught, "Could not mark notifications as read.");
    } finally {
      setMarkingAll(false);
    }
  }

  const visible =
    filter === "unread" ? items.filter((item) => !item.read_at) : items;

  return (
    <DashboardShell
      activePage="notifications"
      title="Notifications"
      subtitle="Reminders from your automations"
      actions={
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={loading || refreshing}
          onClick={() => void load(true)}
          className="gap-2"
        >
          <RefreshCw className={`size-3.5 ${refreshing ? "animate-spin" : ""}`} />
          <span className="hidden sm:inline">Refresh</span>
        </Button>
      }
    >
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
            Inbox
          </p>
          <h1 className="mt-2 text-2xl font-bold">Notifications</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Messages created by your reminder actions.
          </p>
        </div>

        <Button
          type="button"
          variant="outline"
          disabled={unread === 0 || markingAll}
          onClick={() => void handleMarkAll()}
          className="w-full gap-2 sm:w-auto"
        >
          <CheckCheck className="size-4" />
          {markingAll ? "Marking…" : "Mark all as read"}
        </Button>
      </div>

      {error && (
        <div
          role="alert"
          className="mt-6 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive"
        >
          <AlertCircle className="size-4 shrink-0" />
          {error}
        </div>
      )}

      <div className="mt-6 flex gap-2" role="group" aria-label="Filter notifications">
        {(["all", "unread"] as const).map((value) => (
          <button
            key={value}
            type="button"
            aria-pressed={filter === value}
            onClick={() => setFilter(value)}
            className={`rounded-lg border px-3 py-2 text-xs font-medium ${
              filter === value
                ? "border-primary/30 bg-primary/10 text-primary"
                : "border-border text-muted-foreground hover:bg-muted"
            }`}
          >
            {value === "all" ? "All" : `Unread (${unread})`}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="mt-5 space-y-3">
          {[1, 2, 3].map((n) => (
            <div key={n} className="h-20 animate-pulse rounded-xl bg-muted" />
          ))}
        </div>
      ) : visible.length === 0 ? (
        <div className="mt-5 flex flex-col items-center rounded-xl border border-border bg-card px-6 py-14 text-center">
          <div className="flex size-12 items-center justify-center rounded-xl bg-muted text-muted-foreground">
            <Bell className="size-5" />
          </div>
          <h2 className="mt-4 font-semibold">
            {filter === "unread" ? "You're all caught up" : "No notifications yet"}
          </h2>
          <p className="mt-1 max-w-sm text-sm text-muted-foreground">
            {filter === "unread"
              ? "There are no unread notifications."
              : "When a reminder action runs, its message will appear here."}
          </p>
        </div>
      ) : (
        <ul className="mt-5 divide-y divide-border/70 overflow-hidden rounded-xl border border-border bg-card">
          {visible.map((item) => {
            const isUnread = !item.read_at;

            return (
              <li
                key={item.id}
                className={`flex gap-3 px-4 py-4 sm:px-5 ${isUnread ? "bg-primary/5" : ""}`}
              >
                <span
                  aria-hidden="true"
                  className={`mt-1.5 size-2 shrink-0 rounded-full ${
                    isUnread ? "bg-primary" : "bg-transparent"
                  }`}
                />

                <div className="min-w-0 flex-1">
                  <p className={`wrap-break-word text-sm ${isUnread ? "font-semibold" : ""}`}>
                    {item.message}
                  </p>
                  <p
                    className="mt-1 text-xs text-muted-foreground"
                    title={new Date(item.created_at).toLocaleString()}
                  >
                    {timeAgo(item.created_at)}
                  </p>

                  <div className="mt-3 flex flex-wrap gap-2">
                    {item.execution_id && (
                      <Link
                        href={`/executions/${item.execution_id}`}
                        className="inline-flex h-8 items-center rounded-lg border border-border px-3 text-xs font-medium hover:bg-muted"
                      >
                        View run
                      </Link>
                    )}

                    {isUnread && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        disabled={busyId === item.id}
                        onClick={() => void handleMarkRead(item)}
                        className="h-8 text-xs"
                      >
                        {busyId === item.id ? "Saving…" : "Mark as read"}
                      </Button>
                    )}
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </DashboardShell>
  );
}
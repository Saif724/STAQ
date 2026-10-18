"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Bell } from "lucide-react";

import { getNotifications } from "@/lib/api/notifications";

const POLL_MS = 60_000;

export default function NotificationBell() {
  const [unread, setUnread] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function refresh() {
      if (document.visibilityState !== "visible") return;

      try {
        const response = await getNotifications(1);
        if (!cancelled) setUnread(response.data.unread_count);
      } catch {
        // The bell is decorative; page-level requests handle auth errors.
      }
    }

    void refresh();

    const timer = window.setInterval(() => void refresh(), POLL_MS);
    const onChange = () => void refresh();

    document.addEventListener("visibilitychange", onChange);
    window.addEventListener("notifications:changed", onChange);

    return () => {
      cancelled = true;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", onChange);
      window.removeEventListener("notifications:changed", onChange);
    };
  }, []);

  return (
    <Link
      href="/notifications"
      aria-label={unread > 0 ? `Notifications, ${unread} unread` : "Notifications"}
      className="relative flex size-10 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted"
    >
      <Bell className="size-5" />
      {unread > 0 && (
        <span className="absolute right-1 top-1 flex min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold leading-4 text-primary-foreground">
          {unread > 99 ? "99+" : unread}
        </span>
      )}
    </Link>
  );
}
"use client";

import { useRouter } from "next/navigation";
import {
  Activity,
  CircleHelp,
  LayoutDashboard,
  Layers3,
  ListTodo,
  LogOut,
  Settings2,
  ShieldCheck,
  Workflow,
  X,
} from "lucide-react";

import type { CurrentUser } from "@/lib/api/dashboard";
import { Button } from "@/components/ui/button";

type SidebarProps = {
  mobileOpen: boolean;
  onClose: () => void;
  user: CurrentUser | null;
  onLogout: () => void;
  loggingOut: boolean;
  activePage: "overview" | "tasks" | "queues";
};

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

export default function Sidebar({
  mobileOpen,
  onClose,
  user,
  onLogout,
  loggingOut,
  activePage,
}: SidebarProps) {
  const router = useRouter();

  const navigation = [
    { label: "Overview", path: "/dashboard", icon: LayoutDashboard, key: "overview" },
    { label: "Tasks", path: "/tasks", icon: ListTodo, key: "tasks" },
    { label: "Queues", path: "/queues", icon: Layers3, key: "queues" },
  ] as const;

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
            onClick={() => {
              router.push("/dashboard");
              onClose();
            }}
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

          <div className="space-y-1">
            {navigation.map(({ label, path, icon: Icon, key }) => {
              const active = activePage === key;

              return (
                <button
                  key={key}
                  type="button"
                  onClick={() => {
                    router.push(path);
                    onClose();
                  }}
                  className={`flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm ${
                    active
                      ? "bg-primary/10 font-semibold text-primary"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  }`}
                >
                  <Icon className="size-[18px]" />
                  {label}
                  {active && (
                    <span className="ml-auto size-1.5 rounded-full bg-primary" />
                  )}
                </button>
              );
            })}

            <div
              title="Connections page has not been implemented yet"
              className="flex cursor-not-allowed items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground/70"
            >
              <ShieldCheck className="size-[18px]" />
              Connections
              <span className="ml-auto text-[10px] font-medium uppercase tracking-wide">
                Soon
              </span>
            </div>
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

import Link from "next/link";
import {
  ArrowRight,
  Check,
  CheckCircle2,
  Clock3,
  Code2,
  GitBranch,
  Mail,
  Play,
  ShieldCheck,
  Terminal,
  Webhook,
  Zap,
} from "lucide-react";

import { SiteHeader } from "@/components/layout/site-header";

const features = [
  {
    icon: Clock3,
    title: "Schedule anything",
    description:
      "Run workflows once or on recurring schedules with timezone-aware triggers.",
  },
  {
    icon: GitBranch,
    title: "Chain actions",
    description:
      "Build workflows from ordered actions and control what happens when something fails.",
  },
  {
    icon: Zap,
    title: "Reliable execution",
    description:
      "Queues, workers, retries, and execution tracking keep your automation running.",
  },
  {
    icon: ShieldCheck,
    title: "Built for control",
    description:
      "Authentication, ownership checks, execution history, and structured task state.",
  },
];

const actions = [
  {
    icon: Webhook,
    title: "HTTP",
    description: "Call APIs and webhooks automatically.",
  },
  {
    icon: Mail,
    title: "Email",
    description: "Send automated emails through connected services.",
  },
  {
    icon: Terminal,
    title: "Shell",
    description: "Execute commands as part of your workflows.",
  },
  {
    icon: Clock3,
    title: "Reminder",
    description: "Schedule reminders exactly when you need them.",
  },
];

const executions = [
  {
    name: "Daily API sync",
    time: "2 minutes ago",
  },
  {
    name: "Weekly report",
    time: "18 minutes ago",
  },
  {
    name: "Email reminder",
    time: "42 minutes ago",
  },
];

export default function HomePage() {
  return (
    <main className="min-h-screen overflow-hidden bg-background text-foreground">
      <SiteHeader />

      {/* Hero */}
      <section className="relative isolate">
        <div className="pointer-events-none absolute inset-0 -z-10">
          <div className="absolute left-1/2 top-0 h-[600px] w-[900px] -translate-x-1/2 rounded-full bg-primary/10 blur-3xl animate-staq-pulse" />

          <div
            className="absolute inset-0 opacity-[0.035]"
            style={{
              backgroundImage:
                "linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)",
              backgroundSize: "48px 48px",
            }}
          />
        </div>

        <div className="mx-auto max-w-7xl px-6 pb-20 pt-32 sm:pt-36 lg:px-8 lg:pb-28 lg:pt-40">
          <div className="mx-auto max-w-4xl text-center">
            <div className="animate-staq-fade-up">
              <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-3.5 py-1.5 text-sm font-medium text-primary">
                <Zap className="size-3.5" />
                Smart task automation
              </div>
            </div>

            <h1
              className="mt-7 animate-staq-fade-up text-5xl font-bold tracking-[-0.04em] sm:text-6xl lg:text-7xl"
              style={{ animationDelay: "100ms" }}
            >
              Automate the work.
              <span className="block text-muted-foreground">
                Focus on what matters.
              </span>
            </h1>

            <p
              className="mx-auto mt-7 max-w-2xl animate-staq-fade-up text-lg leading-8 text-muted-foreground sm:text-xl"
              style={{ animationDelay: "180ms" }}
            >
              Create scheduled workflows, chain actions together, and track
              every execution from one reliable platform.
            </p>

            <div
              className="mt-9 flex animate-staq-fade-up flex-col items-center justify-center gap-3 sm:flex-row"
              style={{ animationDelay: "260ms" }}
            >
              <Link
                href="/register"
                className="group inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-primary px-6 text-sm font-semibold text-primary-foreground shadow-lg shadow-primary/20 transition-all duration-200 hover:-translate-y-0.5 hover:bg-primary/90 hover:shadow-xl hover:shadow-primary/25"
              >
                Start building

                <ArrowRight className="size-4 transition-transform duration-200 group-hover:translate-x-1" />
              </Link>

              <Link
                href="/login"
                className="inline-flex h-11 items-center justify-center rounded-lg border border-border bg-background/70 px-6 text-sm font-semibold backdrop-blur transition-all duration-200 hover:-translate-y-0.5 hover:bg-accent"
              >
                Sign in
              </Link>
            </div>
          </div>

          {/* Dashboard preview */}
          <div
            className="mx-auto mt-20 max-w-5xl animate-staq-fade-up"
            style={{ animationDelay: "350ms" }}
          >
            <div className="relative">
              <div className="absolute -inset-8 -z-10 rounded-[2rem] bg-primary/10 blur-3xl" />

              <div className="overflow-hidden rounded-2xl border border-border/70 bg-card shadow-2xl shadow-black/10">
                {/* Browser bar */}
                <div className="flex h-11 items-center gap-2 border-b border-border bg-muted/30 px-4">
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />

                  <div className="mx-auto hidden max-w-sm flex-1 rounded-md border border-border bg-background/70 px-3 py-1 text-center text-[10px] text-muted-foreground sm:block">
                    app.staq.dev/dashboard
                  </div>

                  <div className="w-16" />
                </div>

                <div className="grid min-h-[420px] grid-cols-1 md:grid-cols-[190px_1fr]">
                  {/* Sidebar */}
                  <aside className="hidden border-r border-border bg-muted/10 p-5 md:block">
                    <div className="mb-8 flex items-center gap-2 text-sm font-semibold">
                      <span className="flex size-6 items-center justify-center rounded-md bg-primary text-[10px] font-bold text-primary-foreground">
                        S
                      </span>

                      STAQ
                    </div>

                    <div className="space-y-1 text-xs">
                      <div className="rounded-lg bg-primary/10 px-3 py-2 font-medium text-primary">
                        Dashboard
                      </div>

                      <div className="px-3 py-2 text-muted-foreground">
                        Tasks
                      </div>

                      <div className="px-3 py-2 text-muted-foreground">
                        Executions
                      </div>

                      <div className="px-3 py-2 text-muted-foreground">
                        Connections
                      </div>
                    </div>
                  </aside>

                  {/* Dashboard */}
                  <div className="p-5 sm:p-7">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-xs text-muted-foreground">
                          Dashboard
                        </p>

                        <h2 className="mt-1 text-xl font-semibold tracking-tight">
                          Your automations
                        </h2>
                      </div>

                      <div className="hidden items-center gap-2 rounded-lg border border-border px-3 py-2 text-xs font-medium sm:flex">
                        <span className="size-2 rounded-full bg-emerald-500" />
                        All systems operational
                      </div>
                    </div>

                    <div className="mt-6 grid gap-3 sm:grid-cols-3">
                      {[
                        ["12", "Active tasks"],
                        ["248", "Executions"],
                        ["98.4%", "Success rate"],
                      ].map(([value, label]) => (
                        <div
                          key={label}
                          className="rounded-xl border border-border bg-background p-4"
                        >
                          <p className="text-xs text-muted-foreground">
                            {label}
                          </p>

                          <p className="mt-2 text-2xl font-semibold tracking-tight">
                            {value}
                          </p>
                        </div>
                      ))}
                    </div>

                    <div className="mt-5 rounded-xl border border-border">
                      <div className="flex items-center justify-between border-b border-border px-4 py-3">
                        <span className="text-xs font-semibold">
                          Recent executions
                        </span>

                        <span className="text-[11px] text-muted-foreground">
                          View all
                        </span>
                      </div>

                      {executions.map((execution) => (
                        <div
                          key={execution.name}
                          className="flex items-center justify-between border-b border-border px-4 py-3 last:border-0"
                        >
                          <div className="flex items-center gap-3">
                            <div className="flex size-7 items-center justify-center rounded-full bg-emerald-500/10">
                              <CheckCircle2 className="size-3.5 text-emerald-500" />
                            </div>

                            <div>
                              <p className="text-xs font-medium">
                                {execution.name}
                              </p>

                              <p className="text-[10px] text-muted-foreground">
                                {execution.time}
                              </p>
                            </div>
                          </div>

                          <span className="text-[10px] font-medium text-emerald-600 dark:text-emerald-400">
                            Completed
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Feature section */}
      <section
        id="features"
        className="scroll-mt-20 border-t border-border/70"
      >
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8 lg:py-28">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold uppercase tracking-wider text-primary">
              Why STAQ
            </p>

            <h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
              Automation without losing control.
            </h2>

            <p className="mt-4 text-base leading-7 text-muted-foreground">
              STAQ gives you the scheduling, execution, and visibility needed
              to turn repetitive processes into dependable workflows.
            </p>
          </div>

          <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((feature, index) => {
              const Icon = feature.icon;

              return (
                <div
                  key={feature.title}
                  className="group rounded-2xl border border-border bg-card p-6 transition-all duration-300 hover:-translate-y-1 hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5"
                  style={{
                    animation: "staq-fade-up 0.7s ease-out both",
                    animationDelay: `${index * 80}ms`,
                  }}
                >
                  <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary transition-transform duration-300 group-hover:scale-110">
                    <Icon className="size-5" />
                  </div>

                  <h3 className="mt-5 font-semibold">{feature.title}</h3>

                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    {feature.description}
                  </p>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Actions */}
      <section
        id="actions"
        className="scroll-mt-20 border-t border-border/70 bg-muted/20"
      >
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8 lg:py-28">
          <div className="grid gap-12 lg:grid-cols-[0.8fr_1.2fr] lg:items-center">
            <div>
              <p className="text-sm font-semibold uppercase tracking-wider text-primary">
                Action system
              </p>

              <h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
                One workflow.
                <span className="block text-muted-foreground">
                  Multiple actions.
                </span>
              </h2>

              <p className="mt-5 max-w-lg leading-7 text-muted-foreground">
                Build workflows from focused action types. Order them exactly
                how you want them to execute and decide how failures should be
                handled.
              </p>

              <div className="mt-7 space-y-3">
                {[
                  "Ordered action execution",
                  "Failure handling",
                  "Execution history",
                  "Retry support",
                ].map((item) => (
                  <div
                    key={item}
                    className="flex items-center gap-3 text-sm"
                  >
                    <span className="flex size-5 items-center justify-center rounded-full bg-primary/10 text-primary">
                      <Check className="size-3" />
                    </span>

                    {item}
                  </div>
                ))}
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              {actions.map((action) => {
                const Icon = action.icon;

                return (
                  <div
                    key={action.title}
                    className="group rounded-2xl border border-border bg-background p-6 transition-all duration-300 hover:-translate-y-1 hover:border-primary/30 hover:shadow-lg"
                  >
                    <div className="flex size-10 items-center justify-center rounded-xl bg-muted text-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
                      <Icon className="size-5" />
                    </div>

                    <h3 className="mt-5 font-semibold">{action.title}</h3>

                    <p className="mt-2 text-sm leading-6 text-muted-foreground">
                      {action.description}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-border/70">
        <div className="mx-auto max-w-3xl px-6 py-24 text-center lg:px-8 lg:py-28">
          <div className="mx-auto flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <Play className="size-5" />
          </div>

          <h2 className="mt-6 text-3xl font-bold tracking-tight sm:text-4xl">
            Start automating the repetitive work.
          </h2>

          <p className="mx-auto mt-4 max-w-xl leading-7 text-muted-foreground">
            Create your first workflow and let STAQ handle the scheduling,
            execution, and tracking.
          </p>

          <Link
            href="/register"
            className="group mt-8 inline-flex h-11 items-center gap-2 rounded-lg bg-primary px-6 text-sm font-semibold text-primary-foreground shadow-lg shadow-primary/20 transition-all duration-200 hover:-translate-y-0.5 hover:bg-primary/90"
          >
            Create your account

            <ArrowRight className="size-4 transition-transform duration-200 group-hover:translate-x-1" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border/70">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between lg:px-8">
          <div>
            <span className="font-semibold text-foreground">STAQ</span>

            <span className="ml-2">
              © {new Date().getFullYear()}
            </span>
          </div>

          <div className="flex items-center gap-5">
            <Link
              href="/privacy"
              className="transition-colors hover:text-foreground"
            >
              Privacy
            </Link>

            <Link
              href="/login"
              className="transition-colors hover:text-foreground"
            >
              Sign in
            </Link>

            <Link
              href="/register"
              className="transition-colors hover:text-foreground"
            >
              Get started
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}
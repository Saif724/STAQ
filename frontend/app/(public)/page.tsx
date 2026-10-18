import Link from "next/link";
import {
  ArrowRight,
  Check,
  Clock3,
  GitBranch,
  Mail,
  Play,
  ShieldCheck,
  Terminal,
  Webhook,
  Zap,
  History,
  RefreshCw,
} from "lucide-react";

import { SiteHeader } from "@/components/layout/site-header";

const features = [
  {
    icon: Clock3,
    title: "Flexible scheduling",
    description:
      "Schedule one-time or recurring workflows with timezone-aware triggers.",
  },
  {
    icon: GitBranch,
    title: "Ordered workflows",
    description:
      "Combine multiple actions in a defined order and configure how failures are handled.",
  },
  {
    icon: RefreshCw,
    title: "Retries and recovery",
    description:
      "Configure retries and track execution outcomes when actions encounter errors.",
  },
  {
    icon: History,
    title: "Execution visibility",
    description:
      "Review execution history, inspect failures, and understand what happened during a run.",
  },
];

const actions = [
  {
    icon: Webhook,
    title: "HTTP requests",
    description:
      "Send requests to APIs and webhooks as part of a scheduled workflow.",
  },
  {
    icon: Mail,
    title: "Email",
    description:
      "Send automated emails using the email integration configured for your workflow.",
  },
  {
    icon: Terminal,
    title: "Shell commands",
    description:
      "Run configured shell commands. Shell execution should be restricted to trusted, authorized workflows.",
  },
  {
    icon: Clock3,
    title: "Reminders",
    description:
      "Create reminder notifications associated with workflow executions.",
  },
];

const workflowSteps = [
  {
    number: "01",
    title: "Trigger",
    description: "A scheduled time arrives.",
    icon: Clock3,
  },
  {
    number: "02",
    title: "Execute",
    description: "The workflow runs its configured actions in order.",
    icon: Zap,
  },
  {
    number: "03",
    title: "Review",
    description: "Inspect the execution status and recorded outcome.",
    icon: History,
  },
];

export default function HomePage() {
  return (
    <main className="min-h-screen overflow-hidden bg-background text-foreground">
      <SiteHeader />

      {/* Hero */}
      <section className="relative isolate">
        <div className="pointer-events-none absolute inset-0 -z-10">
          <div className="animate-staq-pulse absolute left-1/2 top-0 h-150 w-225 -translate-x-1/2 rounded-full bg-primary/10 blur-3xl" />

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
            <div className="animate-staq-fade-up inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-3.5 py-1.5 text-sm font-medium text-primary">
              <Zap className="size-3.5" />
              Scheduled workflow automation
            </div>

            <h1
              className="animate-staq-fade-up mt-7 text-5xl font-bold tracking-[-0.04em] sm:text-6xl lg:text-7xl"
              style={{ animationDelay: "100ms" }}
            >
              Automate the work.
              <span className="block text-muted-foreground">
                Stay in control.
              </span>
            </h1>

            <p
              className="animate-staq-fade-up mx-auto mt-7 max-w-2xl text-lg leading-8 text-muted-foreground sm:text-xl"
              style={{ animationDelay: "180ms" }}
            >
              Schedule tasks, orchestrate ordered actions, and inspect every
              execution from one automation workspace.
            </p>

            <div
              className="animate-staq-fade-up mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row"
              style={{ animationDelay: "260ms" }}
            >
              <Link
                href="/register"
                className="group inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-primary px-6 text-sm font-semibold text-primary-foreground shadow-lg shadow-primary/20 transition-all duration-200 hover:-translate-y-0.5 hover:bg-primary/90 hover:shadow-xl hover:shadow-primary/25"
              >
                Get started
                <ArrowRight className="size-4 transition-transform duration-200 group-hover:translate-x-1" />
              </Link>

              <Link
                href="/login"
                className="inline-flex h-11 items-center justify-center rounded-lg border border-border bg-background/70 px-6 text-sm font-semibold backdrop-blur transition-all duration-200 hover:-translate-y-0.5 hover:bg-accent"
              >
                Sign in
              </Link>
            </div>

            <p className="mt-4 text-xs text-muted-foreground">
              Create workflows. Monitor outcomes. Refine your automation.
            </p>
          </div>

          {/* Illustrative dashboard preview */}
          <div
            className="animate-staq-fade-up mx-auto mt-20 max-w-5xl"
            style={{ animationDelay: "350ms" }}
          >
            <div className="relative">
              <div className="absolute -inset-8 -z-10 rounded-[2rem] bg-primary/10 blur-3xl" />

              <div className="overflow-hidden rounded-2xl border border-border/70 bg-card shadow-2xl shadow-black/10">
                <div className="flex h-11 items-center gap-2 border-b border-border bg-muted/30 px-4">
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />
                  <span className="size-2.5 rounded-full bg-muted-foreground/25" />

                  <div className="mx-auto hidden max-w-sm flex-1 rounded-md border border-border bg-background/70 px-3 py-1 text-center text-[10px] text-muted-foreground sm:block">
                    STAQ · Interface preview
                  </div>

                  <div className="w-16" />
                </div>

                <div className="grid min-h-105 grid-cols-1 md:grid-cols-[190px_1fr]">
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

                  <div className="p-5 sm:p-7">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <p className="text-xs text-muted-foreground">
                          Your workspace
                        </p>
                        <h2 className="mt-1 text-xl font-semibold tracking-tight">
                          Automation overview
                        </h2>
                      </div>
                      <span className="rounded-lg border border-border px-3 py-2 text-xs text-muted-foreground">
                        Preview
                      </span>
                    </div>

                    <div className="mt-6 grid gap-3 sm:grid-cols-3">
                      {["Tasks", "Recent runs", "Success rate"].map(
                        (label) => (
                          <div
                            key={label}
                            className="rounded-xl border border-border bg-background p-4"
                          >
                            <p className="text-xs text-muted-foreground">
                              {label}
                            </p>
                            <p className="mt-2 text-2xl font-semibold tracking-tight text-muted-foreground">
                              —
                            </p>
                          </div>
                        ),
                      )}
                    </div>

                    <div className="mt-5 rounded-xl border border-border">
                      <div className="flex items-center justify-between border-b border-border px-4 py-3">
                        <span className="text-xs font-semibold">
                          Execution history
                        </span>
                        <span className="text-[11px] text-muted-foreground">
                          Example layout
                        </span>
                      </div>

                      <div className="px-4 py-9 text-center">
                        <Clock3 className="mx-auto size-5 text-muted-foreground" />
                        <p className="mt-3 text-xs font-medium">
                          Your executions, in one place
                        </p>
                        <p className="mt-1 text-[11px] text-muted-foreground">
                          This illustration contains no live account data.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features */}
      <section
        id="features"
        className="scroll-mt-20 border-t border-border/70"
      >
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8 lg:py-28">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold uppercase tracking-wider text-primary">
              What STAQ does
            </p>

            <h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
              Automation you can inspect.
            </h2>

            <p className="mt-4 text-base leading-7 text-muted-foreground">
              Configure when your workflows run, define their actions, and
              review their recorded outcomes when you need to investigate.
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

      {/* Workflow example */}
      <section className="border-t border-border/70 bg-muted/20">
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8 lg:py-28">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-sm font-semibold uppercase tracking-wider text-primary">
              The workflow
            </p>
            <h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
              From schedule to outcome.
            </h2>
            <p className="mt-4 leading-7 text-muted-foreground">
              A simple view of how scheduled automation works in STAQ.
            </p>
          </div>

          <div className="mt-12 grid gap-4 md:grid-cols-3">
            {workflowSteps.map((step, index) => {
              const Icon = step.icon;

              return (
                <div
                  key={step.number}
                  className="relative rounded-2xl border border-border bg-card p-6"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold text-primary">
                      {step.number}
                    </span>
                    <Icon className="size-5 text-primary" />
                  </div>
                  <h3 className="mt-6 text-lg font-semibold">{step.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    {step.description}
                  </p>

                  {index < workflowSteps.length - 1 && (
                    <ArrowRight className="absolute -right-3 top-1/2 z-10 hidden size-6 -translate-y-1/2 rounded-full border border-border bg-background p-1 text-muted-foreground md:block" />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Action types */}
      <section
        id="actions"
        className="scroll-mt-20 border-t border-border/70"
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
                  Purpose-built actions.
                </span>
              </h2>

              <p className="mt-5 max-w-lg leading-7 text-muted-foreground">
                Build workflows from supported action types, control their
                order, and configure how execution errors are handled.
              </p>

              <div className="mt-7 space-y-3">
                {[
                  "Ordered action execution",
                  "Configurable failure handling",
                  "Execution history",
                  "Retry support",
                ].map((item) => (
                  <div key={item} className="flex items-center gap-3 text-sm">
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
            Put your recurring work on a schedule.
          </h2>

          <p className="mx-auto mt-4 max-w-xl leading-7 text-muted-foreground">
            Start with a task, configure its trigger and actions, and use
            execution history to follow the outcome.
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
            <span className="ml-2">© {new Date().getFullYear()}</span>
          </div>

          <div className="flex flex-wrap items-center gap-5">
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
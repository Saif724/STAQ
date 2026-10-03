import Link from "next/link";
import {
  ArrowRight,
  CheckCircle2,
  Clock3,
  Code2,
  GitBranch,
  Mail,
  ShieldCheck,
  Zap,
} from "lucide-react";

const features = [
  {
    icon: Clock3,
    title: "Schedule anything",
    description:
      "Run tasks once or on recurring schedules with timezone-aware triggers.",
  },
  {
    icon: GitBranch,
    title: "Chain actions",
    description:
      "Build workflows from ordered actions that execute exactly in the sequence you define.",
  },
  {
    icon: Zap,
    title: "Reliable execution",
    description:
      "STAQ uses queues, workers, retries, and execution tracking to process your tasks reliably.",
  },
  {
    icon: ShieldCheck,
    title: "Built for control",
    description:
      "Authentication, ownership checks, execution history, and structured task state keep automation manageable.",
  },
];

const actionTypes = [
  {
    icon: Code2,
    title: "HTTP",
    description: "Call APIs and webhooks automatically.",
  },
  {
    icon: Mail,
    title: "Email",
    description: "Send automated emails through connected services.",
  },
  {
    icon: Code2,
    title: "Shell",
    description: "Execute commands as part of your workflows.",
  },
  {
    icon: Clock3,
    title: "Reminder",
    description: "Create scheduled reminders for yourself or your workflow.",
  },
];

export default function HomePage() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      {/* Navigation */}
      <header className="border-b border-border/60">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6 lg:px-8">
          <Link href="/" className="text-xl font-bold tracking-tight">
            STAQ
          </Link>

          <nav className="flex items-center gap-3">
            <Link
              href="/privacy"
              className="hidden text-sm text-muted-foreground transition-colors hover:text-foreground sm:block"
            >
              Privacy
            </Link>

            <Link
              href="/login"
              className="rounded-md px-4 py-2 text-sm font-medium transition-colors hover:bg-muted"
            >
              Sign in
            </Link>

            <Link
              href="/register"
              className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              Get started
            </Link>
          </nav>
        </div>
      </header>

      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top,rgba(120,119,198,0.12),transparent_45%)]" />

        <div className="mx-auto max-w-7xl px-6 pb-24 pt-24 lg:px-8 lg:pb-32 lg:pt-32">
          <div className="mx-auto max-w-4xl text-center">
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-muted/40 px-3 py-1 text-sm text-muted-foreground">
              <Zap className="size-3.5" />
              Automate. Schedule. Execute.
            </div>

            <h1 className="text-5xl font-bold tracking-tight sm:text-6xl lg:text-7xl">
              Turn repetitive work into
              <span className="block text-muted-foreground">
                reliable automation.
              </span>
            </h1>

            <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
              STAQ lets you create scheduled workflows, chain actions together,
              and track every execution from one place.
            </p>

            <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Link
                href="/register"
                className="inline-flex h-11 items-center justify-center gap-2 rounded-md bg-primary px-6 text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90"
              >
                Start building
                <ArrowRight className="size-4" />
              </Link>

              <Link
                href="/login"
                className="inline-flex h-11 items-center justify-center rounded-md border border-border px-6 text-sm font-semibold transition-colors hover:bg-muted"
              >
                Sign in
              </Link>
            </div>
          </div>

          {/* Product preview */}
          <div className="mx-auto mt-20 max-w-5xl">
            <div className="overflow-hidden rounded-xl border border-border bg-card shadow-2xl">
              <div className="flex h-10 items-center gap-2 border-b border-border px-4">
                <span className="size-2.5 rounded-full bg-muted-foreground/30" />
                <span className="size-2.5 rounded-full bg-muted-foreground/30" />
                <span className="size-2.5 rounded-full bg-muted-foreground/30" />
              </div>

              <div className="grid min-h-[360px] grid-cols-1 md:grid-cols-[220px_1fr]">
                <aside className="hidden border-r border-border p-5 md:block">
                  <div className="mb-8 text-sm font-semibold">STAQ</div>

                  <div className="space-y-2 text-sm">
                    <div className="rounded-md bg-muted px-3 py-2 font-medium">
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

                <div className="p-6">
                  <div className="mb-6">
                    <p className="text-sm text-muted-foreground">
                      Dashboard
                    </p>
                    <h2 className="mt-1 text-2xl font-semibold">
                      Your automations
                    </h2>
                  </div>

                  <div className="grid gap-4 sm:grid-cols-3">
                    <div className="rounded-lg border border-border p-4">
                      <p className="text-sm text-muted-foreground">
                        Active tasks
                      </p>
                      <p className="mt-2 text-3xl font-semibold">12</p>
                    </div>

                    <div className="rounded-lg border border-border p-4">
                      <p className="text-sm text-muted-foreground">
                        Executions
                      </p>
                      <p className="mt-2 text-3xl font-semibold">248</p>
                    </div>

                    <div className="rounded-lg border border-border p-4">
                      <p className="text-sm text-muted-foreground">
                        Success rate
                      </p>
                      <p className="mt-2 text-3xl font-semibold">98.4%</p>
                    </div>
                  </div>

                  <div className="mt-6 rounded-lg border border-border">
                    <div className="border-b border-border px-4 py-3 text-sm font-medium">
                      Recent executions
                    </div>

                    {[
                      "Daily API sync",
                      "Weekly report",
                      "Email reminder",
                    ].map((task) => (
                      <div
                        key={task}
                        className="flex items-center justify-between border-b border-border px-4 py-3 last:border-0"
                      >
                        <span className="text-sm">{task}</span>

                        <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <CheckCircle2 className="size-3.5" />
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
      </section>

      {/* Features */}
      <section className="border-t border-border">
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Built for automation
            </p>

            <h2 className="mt-3 text-3xl font-bold tracking-tight sm:text-4xl">
              Everything you need to run dependable workflows.
            </h2>

            <p className="mt-4 text-muted-foreground">
              Define what should happen, when it should happen, and track what
              actually happened.
            </p>
          </div>

          <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((feature) => {
              const Icon = feature.icon;

              return (
                <div
                  key={feature.title}
                  className="rounded-xl border border-border p-6"
                >
                  <div className="mb-5 flex size-10 items-center justify-center rounded-lg bg-muted">
                    <Icon className="size-5" />
                  </div>

                  <h3 className="font-semibold">{feature.title}</h3>

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
      <section className="border-t border-border bg-muted/20">
        <div className="mx-auto max-w-7xl px-6 py-24 lg:px-8">
          <div className="grid gap-12 lg:grid-cols-[0.8fr_1.2fr] lg:items-start">
            <div>
              <p className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
                Flexible actions
              </p>

              <h2 className="mt-3 text-3xl font-bold tracking-tight">
                Build workflows from simple building blocks.
              </h2>

              <p className="mt-4 text-muted-foreground">
                Combine different action types into ordered workflows and
                decide how failures should affect the remaining actions.
              </p>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              {actionTypes.map((action) => {
                const Icon = action.icon;

                return (
                  <div
                    key={action.title}
                    className="rounded-xl border border-border bg-background p-5"
                  >
                    <Icon className="size-5" />

                    <h3 className="mt-4 font-semibold">{action.title}</h3>

                    <p className="mt-1 text-sm leading-6 text-muted-foreground">
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
      <section className="border-t border-border">
        <div className="mx-auto max-w-3xl px-6 py-24 text-center lg:px-8">
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Ready to automate the repetitive stuff?
          </h2>

          <p className="mx-auto mt-4 max-w-xl text-muted-foreground">
            Create your first workflow and let STAQ handle the scheduling and
            execution.
          </p>

          <Link
            href="/register"
            className="mt-8 inline-flex h-11 items-center gap-2 rounded-md bg-primary px-6 text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90"
          >
            Create your account
            <ArrowRight className="size-4" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border">
        <div className="mx-auto flex max-w-7xl flex-col gap-3 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between lg:px-8">
          <p>© {new Date().getFullYear()} STAQ. All rights reserved.</p>

          <Link
            href="/privacy"
            className="transition-colors hover:text-foreground"
          >
            Privacy Policy
          </Link>
        </div>
      </footer>
    </main>
  );
}
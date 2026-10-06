import Link from "next/link";
import {
  ArrowLeft,
  ShieldCheck,
  Zap,
} from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";

export const metadata = {
  title: "Privacy Policy",
  description: "Privacy Policy for STAQ.",
};

const sections = [
  {
    title: "Introduction",
    content: (
      <p>
        STAQ is an automation platform that allows users to create,
        schedule, and execute automated workflows. This Privacy Policy
        explains what information STAQ collects, how that information is
        used, and how it is protected.
      </p>
    ),
  },
  {
    title: "Information We Collect",
    content: (
      <>
        <p>
          When you create an account, we may collect information such as
          your name, email address, and authentication credentials.
        </p>

        <p>
          STAQ also stores information required to operate your workflows,
          including tasks, triggers, actions, execution records, and
          related configuration data.
        </p>
      </>
    ),
  },
  {
    title: "Google and Connected Services",
    content: (
      <>
        <p>
          If you connect a Google account or another supported service,
          STAQ may receive information necessary to provide the requested
          integration.
        </p>

        <p>
          We use connected-service data only to provide the functionality
          requested by the user. We do not sell users&apos; personal
          information or connected-service data.
        </p>

        <p>
          Access to connected services can be revoked by the user through
          the relevant service or through STAQ when supported.
        </p>
      </>
    ),
  },
  {
    title: "How We Use Information",
    content: (
      <>
        <p>
          Information is used to provide, maintain, secure, and improve
          STAQ and to execute the workflows configured by users.
        </p>

        <p>
          We may also use information to authenticate users, prevent abuse,
          troubleshoot failures, maintain system reliability, and
          communicate important service-related information.
        </p>
      </>
    ),
  },
  {
    title: "Data Security",
    content: (
      <>
        <p>
          STAQ uses reasonable technical and organizational measures to
          protect user information against unauthorized access,
          alteration, disclosure, or destruction.
        </p>

        <p>
          However, no internet-based service can guarantee absolute
          security.
        </p>
      </>
    ),
  },
  {
    title: "Data Retention",
    content: (
      <p>
        We retain account, workflow, and execution information for as
        long as necessary to provide the service and fulfill legitimate
        operational requirements.
      </p>
    ),
  },
  {
    title: "Your Choices",
    content: (
      <p>
        You may stop using STAQ at any time. Where supported, you may
        disconnect connected services and remove information associated
        with your account.
      </p>
    ),
  },
  {
    title: "Changes to This Policy",
    content: (
      <p>
        We may update this Privacy Policy from time to time. Changes will
        be reflected on this page with an updated revision date.
      </p>
    ),
  },
  {
    title: "Contact",
    content: (
      <p>
        If you have questions about this Privacy Policy or STAQ&apos;s
        handling of information, please contact the STAQ service
        administrator.
      </p>
    ),
  },
];

export default function PrivacyPage() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      {/* Background */}
      <div className="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
        <div className="absolute left-1/2 top-0 h-[500px] w-[700px] -translate-x-1/2 rounded-full bg-primary/5 blur-3xl" />

        <div
          className="absolute inset-0 opacity-[0.025]"
          style={{
            backgroundImage:
              "linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)",
            backgroundSize: "48px 48px",
          }}
        />
      </div>

      {/* Header */}
      <header className="sticky top-0 z-30 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-6 lg:px-8">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-lg font-bold tracking-tight"
          >
            <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-md shadow-primary/20">
              <Zap className="size-4 fill-current" />
            </span>

            STAQ
          </Link>

          <ThemeToggle />
        </div>
      </header>

      <div className="mx-auto max-w-3xl px-6 py-16 lg:px-8 lg:py-24">
        {/* Back link */}
        <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
          <Link
            href="/"
            className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="size-4 transition-transform group-hover:-translate-x-1" />
            Back to STAQ
          </Link>
        </div>

        {/* Heading */}
        <div className="mt-10 animate-in fade-in slide-in-from-bottom-3 duration-700">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border bg-muted/50 px-3 py-1.5 text-xs font-medium text-muted-foreground">
            <ShieldCheck className="size-3.5" />
            Your privacy matters
          </div>

          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
            Privacy Policy
          </h1>

          <p className="mt-4 text-sm text-muted-foreground">
            Last updated: October 5, 2026
          </p>
        </div>

        {/* Content */}
        <div className="mt-12 space-y-12">
          {sections.map((section, index) => (
            <section
              key={section.title}
              className="animate-in fade-in slide-in-from-bottom-3 duration-700"
              style={{
                animationDelay: `${Math.min(index * 70, 500)}ms`,
              }}
            >
              <h2 className="text-xl font-semibold tracking-tight">
                {index + 1}. {section.title}
              </h2>

              <div className="mt-4 space-y-4 text-sm leading-7 text-muted-foreground">
                {section.content}
              </div>
            </section>
          ))}
        </div>

        {/* Footer */}
        <div className="mt-16 border-t border-border pt-8">
          <Link
            href="/"
            className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="size-4 transition-transform group-hover:-translate-x-1" />
            Return to STAQ
          </Link>
        </div>
      </div>
    </main>
  );
}
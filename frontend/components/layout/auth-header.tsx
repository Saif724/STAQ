"use client";

import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";

type AuthHeaderProps = {
  backHref?: string;
  backLabel?: string;
};

export function AuthHeader({
  backHref = "/",
  backLabel = "Back to STAQ",
}: AuthHeaderProps) {
  return (
    <header className="absolute inset-x-0 top-0 z-50">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6 lg:px-8">
        <Link
          href="/"
          className="group flex items-center gap-2"
          aria-label="STAQ home"
        >
          <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground shadow-sm transition-transform duration-200 group-hover:scale-105">
            S
          </span>

          <span className="text-lg font-semibold tracking-tight">
            STAQ
          </span>
        </Link>

        <div className="flex items-center gap-2">
          <Link
            href={backHref}
            className="group hidden items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground sm:inline-flex"
          >
            <ArrowLeft className="size-4 transition-transform duration-200 group-hover:-translate-x-0.5" />

            {backLabel}
          </Link>

          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
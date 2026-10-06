"use client";

import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";

type PublicHeaderProps = {
  backHref?: string;
  backLabel?: string;
};

export function PublicHeader({
  backHref = "/",
  backLabel = "Back to STAQ",
}: PublicHeaderProps) {
  return (
    <header className="fixed inset-x-0 top-0 z-50">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6 lg:px-8">
        <Link
          href="/"
          className="group hidden items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors duration-200 hover:bg-accent hover:text-accent-foreground sm:flex"
        >
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground shadow-sm transition-transform duration-200 group-hover:scale-105">
            S
          </span>
          <span className="text-lg">STAQ</span>
        </Link>
        <div className="flex items-center gap-2">
          <Link
            href={backHref}
            className="hidden items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors duration-200 hover:bg-accent hover:text-accent-foreground sm:flex"
          >
            <ArrowLeft className="h-4 w-4 transition-transform duration-200 group-hover:-translate-x-0.5" />
            {backLabel}
          </Link>

          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}

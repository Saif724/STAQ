import type { Metadata } from "next";
import { Suspense } from "react";

import ConnectionsPage from "@/components/dashboard/connections-page";

export const metadata: Metadata = {
  title: "Connections",
  description: "Manage accounts connected to STAQ automations.",
};

export default function ConnectionsRoute() {
  return (
    <Suspense fallback={<main className="min-h-screen bg-background" />}>
      <ConnectionsPage />
    </Suspense>
  );
}
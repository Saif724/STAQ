import type { Metadata } from "next";
import DashboardOverview from "@/components/dashboard/dashboard-overview";

export const metadata: Metadata = {
  title: "Dashboard",
  description: "Manage your STAQ automation workspace.",
};

export default function DashboardPage() {
  return <DashboardOverview />;
}
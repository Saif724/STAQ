import type { Metadata } from "next";
import ExecutionDetailPage from "@/components/dashboard/execution-detail-page";

export const metadata: Metadata = {
  title: "Execution details",
  description: "Inspect an automation run and its execution logs.",
};

export default function ExecutionDetailRoute() {
  return <ExecutionDetailPage />;
}
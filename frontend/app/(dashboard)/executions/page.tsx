import type { Metadata } from "next";
import ExecutionsPage from "@/components/dashboard/executions-page";

export const metadata: Metadata = {
  title: "Executions",
  description: "Inspect automation execution history and outcomes.",
};

export default function ExecutionsRoute() {
  return <ExecutionsPage />;
}
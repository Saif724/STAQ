import type { Metadata } from "next";
import TasksPage from "@/components/dashboard/tasks-page";

export const metadata: Metadata = {
  title: "Tasks",
  description: "Create and manage your STAQ automation tasks.",
};

export default function TasksRoute() {
  return <TasksPage />;
}
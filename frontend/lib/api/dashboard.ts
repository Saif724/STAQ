import { apiRequest } from "@/lib/api/client";
import type { Execution } from "@/lib/api/automation";

export type CurrentUser = {
  id: string;
  full_name: string;
  email: string;
  email_verified: boolean;
  timezone: string;
};

export type DashboardRecentExecution = Execution & {
  task_name: string;
};

export type DashboardUpcomingExecution = {
  task_id: string;
  task_name: string;
  trigger_id: string;
  trigger_type: string;
  next_run_at: string;
  timezone: string;
};

export type DashboardSummary = {
  window_starts_at: string;
  window_days: number;
  successful_count: number;
  failed_count: number;
  timed_out_count: number;
  cancelled_count: number;
  pending_count: number;
  running_count: number;
  success_rate: number;
  recent_executions: DashboardRecentExecution[];
  recent_failures: DashboardRecentExecution[];
  upcoming_executions: DashboardUpcomingExecution[];
};

export type TaskStatus = "Active" | "Paused" | "Archived";

export type Task = {
  id: string;
  user_id: string;
  queue_id: string;
  name: string;
  description: string | null;
  status: TaskStatus;
  timeout_seconds: number;
  max_retries: number;
  created_at: string;
  updated_at: string;
};

export type Queue = {
  id: string;
  name: string;
  description: string | null;
  is_active: boolean;
  created_at: string;
};

type ResponseWrapper<T> = {
  data: T;
};

export function getCurrentUser() {
  return apiRequest<ResponseWrapper<CurrentUser>>("/users/me");
}

export function getTasks() {
  return apiRequest<ResponseWrapper<Task[]>>("/tasks");
}

export function getQueues() {
  return apiRequest<ResponseWrapper<Queue[]>>("/queues");
}

export function getDashboardSummary() {
  return apiRequest<ResponseWrapper<DashboardSummary>>("/dashboard/summary");
}

export type UpdateCurrentUserRequest = {
  full_name: string;
  timezone: string;
};

export function updateCurrentUser(payload: UpdateCurrentUserRequest) {
  return apiRequest<ResponseWrapper<CurrentUser>>("/users/me", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export type CreateTaskRequest = {
  name: string;
  description: string | null;
  timeout_seconds: number;
  max_retries: number;
};

export async function createTask(
  payload: CreateTaskRequest,
) {
  return apiRequest<{ data: Task }>("/tasks", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function deleteTask(taskId: string) {
  return apiRequest<void>(`/tasks/${taskId}`, {
    method: "DELETE",
  });
}
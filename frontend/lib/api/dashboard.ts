import { apiRequest } from "@/lib/api/client";

export type CurrentUser = {
  id: string;
  full_name: string;
  email: string;
  email_verified: boolean;
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

export type CreateTaskRequest = {
  queue_id: string;
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
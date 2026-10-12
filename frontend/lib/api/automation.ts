import { apiRequest } from "@/lib/api/client";

export type ApiResponse<T> = {
  data: T;
};

export type TriggerType =
  | "ONCE"
  | "DAILY"
  | "WEEKLY"
  | "MONTHLY"
  | "YEARLY"
  | "CRON";

export type Trigger = {
  id: string;
  task_id: string;
  trigger_type: TriggerType;
  cron_expression?: string | null;
  timezone: string;
  start_at: string;
  next_run_at: string;
  last_run_at?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type CreateTriggerRequest = {
  task_id: string;
  trigger_type: TriggerType;
  cron_expression?: string | null;
  timezone: string;
  start_at: string;
};

export type UpdateTriggerRequest = {
  trigger_type: TriggerType;
  cron_expression?: string | null;
  timezone: string;
  start_at: string;
  is_active: boolean;
};

export type ActionType = "REMINDER" | "EMAIL" | "HTTP" | "SHELL";

export type ActionConfiguration = Record<string, unknown>;

export type Action = {
  id: string;
  task_id: string;
  action_type: ActionType;
  execution_order: number;
  // The Go API returns stored JSON bytes as a Base64 string.
  configuration: string;
  continue_on_failure: boolean;
  created_at: string;
  updated_at: string;
};

export type CreateActionRequest = {
  task_id: string;
  action_type: ActionType;
  execution_order: number;
  configuration: ActionConfiguration;
  continue_on_failure: boolean;
};

export type UpdateActionRequest = {
  action_type: ActionType;
  execution_order: number;
  configuration: ActionConfiguration;
  continue_on_failure: boolean;
};

export type ExecutionStatus =
  | "PENDING"
  | "RUNNING"
  | "SUCCESS"
  | "FAILED"
  | "CANCELLED"
  | "TIMED_OUT";

export type Execution = {
  id: string;
  task_id: string;
  trigger_id: string;
  scheduled_at: string;
  status: ExecutionStatus;
  started_at: string;
  heartbeat_at?: string | null;
  completed_at?: string | null;
  duration_ms?: number | null;
  retry_count: number;
  error_message?: string | null;
  created_at: string;
};

export type ExecutionLog = {
  id: string;
  execution_id: string;
  log_level: "INFO" | "WARNING" | "ERROR";
  message: string;
  created_at: string;
};

export function getTaskTriggers(taskId: string) {
  return apiRequest<ApiResponse<Trigger[]>>(
    `/tasks/${taskId}/triggers`,
  );
}

export function createTrigger(payload: CreateTriggerRequest) {
  return apiRequest<ApiResponse<Trigger>>("/triggers", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateTrigger(
  triggerId: string,
  payload: UpdateTriggerRequest,
) {
  return apiRequest<ApiResponse<Trigger>>(`/triggers/${triggerId}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function deleteTrigger(triggerId: string) {
  return apiRequest<void>(`/triggers/${triggerId}`, {
    method: "DELETE",
  });
}

export function getTaskActions(taskId: string) {
  return apiRequest<ApiResponse<Action[]>>(
    `/tasks/${taskId}/actions`,
  );
}

export function createAction(payload: CreateActionRequest) {
  return apiRequest<ApiResponse<Action>>("/actions", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateAction(
  actionId: string,
  payload: UpdateActionRequest,
) {
  return apiRequest<ApiResponse<Action>>(`/actions/${actionId}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function deleteAction(actionId: string) {
  return apiRequest<void>(`/actions/${actionId}`, {
    method: "DELETE",
  });
}

export function getExecution(executionId: string) {
  return apiRequest<ApiResponse<Execution>>(
    `/executions/${executionId}`,
  );
}

export function getExecutionLogs(executionId: string) {
  return apiRequest<ApiResponse<ExecutionLog[]>>(
    `/executions/${executionId}/logs`,
  );
}
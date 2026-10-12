import { apiRequest } from "@/lib/api/client";
import type { Action, CreateActionRequest, CreateTriggerRequest, Trigger } from "@/lib/api/automation";
import type { CreateTaskRequest, Task } from "@/lib/api/dashboard";

export type CreateTaskSetupRequest = CreateTaskRequest &
  Omit<CreateTriggerRequest, "task_id"> & {
  actions: Array<Omit<CreateActionRequest, "task_id">>;
};

export type TaskSetup = {
  task: Task;
  trigger: Trigger;
  actions: Action[];
};

export function createTaskSetup(payload: CreateTaskSetupRequest) {
  return apiRequest<{ data: TaskSetup }>("/task-setups", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

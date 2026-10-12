"use client";

import { useState } from "react";
import { AlertCircle, ArrowLeft, ArrowRight, Plus, Trash2, Workflow, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  createTask,
  type Queue,
} from "@/lib/api/dashboard";
import {
  createAction,
  createTrigger,
  type ActionType,
  type TriggerType,
} from "@/lib/api/automation";

type Props = {
  queues: Queue[];
  onClose: () => void;
  onCreated: () => void;
};

type ActionDraft = {
  id: string;
  type: ActionType;
  continueOnFailure: boolean;
  message: string;
  sender: "STAQ" | "USER_GMAIL";
  to: string;
  subject: string;
  body: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  url: string;
  headers: string;
  command: string;
  args: string;
};

const steps = ["Task details", "Schedule", "Actions", "Review"] as const;

const inputClass =
  "h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring";

const textareaClass =
  "w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring";

function newAction(): ActionDraft {
  return {
    id: crypto.randomUUID(),
    type: "REMINDER",
    continueOnFailure: false,
    message: "",
    sender: "STAQ",
    to: "",
    subject: "",
    body: "",
    method: "POST",
    url: "",
    headers: '{\n  "Content-Type": "application/json"\n}',
    command: "echo",
    args: "Hello from STAQ",
  };
}

function localDateAfterOneHour() {
  const date = new Date(Date.now() + 60 * 60 * 1000);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 10);
}

function localTimeAfterOneHour() {
  const date = new Date(Date.now() + 60 * 60 * 1000);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(11, 16);
}

function configurationFor(action: ActionDraft): Record<string, unknown> {
  switch (action.type) {
    case "REMINDER":
      return { message: action.message.trim() };

    case "EMAIL":
      return {
        sender: action.sender,
        to: action.to.trim(),
        subject: action.subject.trim(),
        body: action.body,
      };

    case "HTTP": {
      let headers: unknown;

      try {
        headers = JSON.parse(action.headers);
      } catch {
        throw new Error(
          `Action ${action.id}: HTTP headers must be valid JSON.`,
        );
      }

      if (
        headers === null ||
        typeof headers !== "object" ||
        Array.isArray(headers)
      ) {
        throw new Error("HTTP headers must be a JSON object.");
      }

      return {
        method: action.method,
        url: action.url.trim(),
        headers,
        body: action.body,
      };
    }

    case "SHELL":
      return {
        command: action.command.trim(),
        args: action.args
          .split("\n")
          .map((value) => value.trim())
          .filter(Boolean),
      };
  }
}

function validateAction(action: ActionDraft): string | null {
  if (action.type === "REMINDER" && !action.message.trim()) {
    return "Enter a reminder message.";
  }

  if (action.type === "EMAIL") {
    if (!action.to.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(action.to)) {
      return "Enter a valid recipient email address.";
    }
    if (!action.subject.trim() || !action.body.trim()) {
      return "Enter an email subject and body.";
    }
  }

  if (action.type === "HTTP") {
    if (!/^https?:\/\/\S+$/i.test(action.url.trim())) {
      return "Enter a valid HTTP or HTTPS URL.";
    }
    try {
      const headers = JSON.parse(action.headers);
      if (
        headers === null ||
        typeof headers !== "object" ||
        Array.isArray(headers)
      ) {
        return "HTTP headers must be a JSON object.";
      }
    } catch {
      return "HTTP headers must be valid JSON.";
    }
  }

  if (action.type === "SHELL" && !action.command.trim()) {
    return "Enter a shell command.";
  }

  return null;
}

export default function TaskCreationWizard({
  queues,
  onClose,
  onCreated,
}: Props) {
  const [step, setStep] = useState(0);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [queueId, setQueueId] = useState("");
  const [timeoutSeconds, setTimeoutSeconds] = useState("60");
  const [maxRetries, setMaxRetries] = useState("3");

  const [triggerType, setTriggerType] = useState<TriggerType>("ONCE");
  const [startDate, setStartDate] = useState(localDateAfterOneHour);
  const [startTime, setStartTime] = useState(localTimeAfterOneHour);
  const [cronExpression, setCronExpression] = useState("0 9 * * *");
  const [timezone, setTimezone] = useState("Asia/Dhaka");

  const [actions, setActions] = useState<ActionDraft[]>([newAction()]);

  const activeQueues = queues.filter((queue) => queue.is_active);

  function updateAction(id: string, patch: Partial<ActionDraft>) {
    setActions((current) =>
      current.map((action) =>
        action.id === id ? { ...action, ...patch } : action,
      ),
    );
  }

  function validateStep(index: number): string | null {
    if (index === 0) {
      if (!name.trim()) return "Enter a task name.";
      if (!queueId) return "Choose an active queue.";

      const timeout = Number(timeoutSeconds);
      const retries = Number(maxRetries);

      if (!Number.isInteger(timeout) || timeout < 1) {
        return "Timeout must be a positive whole number.";
      }
      if (!Number.isInteger(retries) || retries < 0) {
        return "Maximum retries must be zero or greater.";
      }
    }

    if (index === 1) {
      if (!startDate || !startTime) {
        return "Choose a start date and time.";
      }

      const startAt = new Date(`${startDate}T${startTime}:00+06:00`);

      if (Number.isNaN(startAt.getTime())) {
        return "The selected start date or time is invalid.";
      }

      if (triggerType === "ONCE" && startAt.getTime() <= Date.now()) {
        return "A one-time trigger must be scheduled in the future.";
      }

      if (
        triggerType === "CRON" &&
        !cronExpression.trim()
      ) {
        return "Enter a cron expression.";
      }
    }

    if (index === 2) {
      if (actions.length === 0) return "Add at least one action.";

      for (let i = 0; i < actions.length; i += 1) {
        const issue = validateAction(actions[i]);
        if (issue) return `Action ${i + 1}: ${issue}`;
      }
    }

    return null;
  }

  function nextStep() {
    setError("");
    const issue = validateStep(step);

    if (issue) {
      setError(issue);
      return;
    }

    setStep((current) => Math.min(current + 1, steps.length - 1));
  }

  async function submitWorkflow() {
    setError("");

    for (let index = 0; index < 3; index += 1) {
      const issue = validateStep(index);
      if (issue) {
        setStep(index);
        setError(issue);
        return;
      }
    }

    setSaving(true);

    let createdTaskId: string | undefined;
    let createdTriggerId: string | undefined;
    const createdActionIds: string[] = [];

    try {
      const taskResponse = await createTask({
        name: name.trim(),
        description: description.trim() || null,
        timeout_seconds: Number(timeoutSeconds),
        max_retries: Number(maxRetries),
      });

      createdTaskId = taskResponse.data.id;

      const startAt = new Date(
        `${startDate}T${startTime}:00+06:00`,
      ).toISOString();

      const triggerResponse = await createTrigger({
        task_id: createdTaskId,
        trigger_type: triggerType,
        timezone: timezone.trim(),
        start_at: startAt,
        ...(triggerType === "CRON"
          ? { cron_expression: cronExpression.trim() }
          : {}),
      });

      createdTriggerId = triggerResponse.data.id;

      for (let index = 0; index < actions.length; index += 1) {
        const draft = actions[index];

        const actionResponse = await createAction({
          task_id: createdTaskId,
          action_type: draft.type,
          execution_order: index + 1,
          configuration: configurationFor(draft),
          continue_on_failure: draft.continueOnFailure,
        });

        createdActionIds.push(actionResponse.data.id);
      }

      onCreated();
    } catch (caught) {
      const message =
        caught instanceof Error
          ? caught.message
          : "The workflow could not be created.";

      if (createdTaskId) {
        const cleanupProblems: string[] = [];

        // Best-effort cleanup prevents leaving a partially configured task.
        for (const actionId of createdActionIds.reverse()) {
          try {
            const { deleteAction } = await import("@/lib/api/automation");
            await deleteAction(actionId);
          } catch {
            cleanupProblems.push("an action");
          }
        }

        if (createdTriggerId) {
          try {
            const { deleteTrigger } = await import("@/lib/api/automation");
            await deleteTrigger(createdTriggerId);
          } catch {
            cleanupProblems.push("the trigger");
          }
        }

        try {
          const { deleteTask } = await import("@/lib/api/dashboard");
          await deleteTask(createdTaskId);
        } catch {
          cleanupProblems.push("the task");
        }

        setError(
          cleanupProblems.length
            ? `Setup failed: ${message} Cleanup also failed for ${[
                ...new Set(cleanupProblems),
              ].join(", ")}. Refresh the task list and inspect the created task.`
            : `Setup failed: ${message} The incomplete task setup was cleaned up. Please review your inputs and try again.`,
        );
      } else {
        setError(message);
      }
    } finally {
      setSaving(false);
    }
  }

  function renderActionFields(action: ActionDraft) {
    if (action.type === "REMINDER") {
      return (
        <Field label="Reminder message">
          <textarea
            required
            rows={3}
            className={textareaClass}
            value={action.message}
            onChange={(event) =>
              updateAction(action.id, { message: event.target.value })
            }
            placeholder="Your report is ready."
          />
        </Field>
      );
    }

    if (action.type === "EMAIL") {
      return (
        <>
          <Field label="Sender">
            <select
              className={inputClass}
              value={action.sender}
              onChange={(event) =>
                updateAction(action.id, {
                  sender: event.target.value as ActionDraft["sender"],
                })
              }
            >
              <option value="STAQ">STAQ platform email</option>
              <option value="USER_GMAIL">Connected user Gmail</option>
            </select>
          </Field>

          {action.sender === "USER_GMAIL" && (
            <p className="text-xs text-muted-foreground">
              The backend requires a connection_id for USER_GMAIL. This wizard
              does not yet load Gmail connections, so choose STAQ unless you
              already have a connection ID available.
            </p>
          )}

          <Field label="Recipient email">
            <input
              className={inputClass}
              type="email"
              value={action.to}
              onChange={(event) =>
                updateAction(action.id, { to: event.target.value })
              }
              placeholder="recipient@example.com"
            />
          </Field>

          <Field label="Subject">
            <input
              className={inputClass}
              value={action.subject}
              onChange={(event) =>
                updateAction(action.id, { subject: event.target.value })
              }
              placeholder="Your report is ready"
            />
          </Field>

          <Field label="Email body">
            <textarea
              rows={4}
              className={textareaClass}
              value={action.body}
              onChange={(event) =>
                updateAction(action.id, { body: event.target.value })
              }
              placeholder="Write the email content..."
            />
          </Field>
        </>
      );
    }

    if (action.type === "HTTP") {
      return (
        <>
          <Field label="HTTP method">
            <select
              className={inputClass}
              value={action.method}
              onChange={(event) =>
                updateAction(action.id, {
                  method: event.target.value as ActionDraft["method"],
                })
              }
            >
              {["GET", "POST", "PUT", "PATCH", "DELETE"].map((method) => (
                <option key={method}>{method}</option>
              ))}
            </select>
          </Field>

          <Field label="Request URL">
            <input
              className={inputClass}
              type="url"
              value={action.url}
              onChange={(event) =>
                updateAction(action.id, { url: event.target.value })
              }
              placeholder="https://example.com/webhook"
            />
          </Field>

          <Field label="Headers (JSON object)">
            <textarea
              rows={4}
              className={`${textareaClass} font-mono text-xs`}
              value={action.headers}
              onChange={(event) =>
                updateAction(action.id, { headers: event.target.value })
              }
            />
          </Field>

          <Field label="Request body">
            <textarea
              rows={4}
              className={`${textareaClass} font-mono text-xs`}
              value={action.body}
              onChange={(event) =>
                updateAction(action.id, { body: event.target.value })
              }
              placeholder={'{"event":"task.completed"}'}
            />
          </Field>
        </>
      );
    }

    return (
      <>
        <Field label="Command">
          <input
            className={inputClass}
            value={action.command}
            onChange={(event) =>
              updateAction(action.id, { command: event.target.value })
            }
            placeholder="echo"
          />
        </Field>

        <Field label="Arguments (one per line)">
          <textarea
            rows={4}
            className={`${textareaClass} font-mono text-xs`}
            value={action.args}
            onChange={(event) =>
              updateAction(action.id, { args: event.target.value })
            }
            placeholder={"hello\nworld"}
          />
        </Field>

        <p className="text-xs text-muted-foreground">
          Shell execution can be dangerous. Only configure commands permitted
          by your backend worker&apos;s allowlist.
        </p>
      </>
    );
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center overflow-y-auto bg-black/50 p-3 backdrop-blur-sm sm:p-5"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !saving) onClose();
      }}
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="wizard-title"
        className="my-auto max-h-[94vh] w-full max-w-3xl overflow-y-auto rounded-2xl border border-border bg-card p-5 shadow-2xl sm:p-7"
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-primary">
              <Workflow className="size-5" />
              <span className="text-xs font-semibold uppercase tracking-[0.16em]">
                Workflow builder
              </span>
            </div>
            <h2 id="wizard-title" className="mt-2 text-xl font-semibold">
              Create an automation task
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Configure the task, its schedule, and the actions it should run.
            </p>
          </div>

          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Close wizard"
            disabled={saving}
            onClick={onClose}
          >
            <X className="size-4" />
          </Button>
        </div>

        <div className="mt-6 grid grid-cols-4 gap-2">
          {steps.map((label, index) => (
            <button
              type="button"
              key={label}
              disabled={saving || index > step}
              onClick={() => {
                setError("");
                setStep(index);
              }}
              className={`rounded-lg border px-2 py-3 text-left text-xs ${
                step === index
                  ? "border-primary/40 bg-primary/10 text-primary"
                  : index < step
                    ? "border-border bg-muted/50"
                    : "border-border text-muted-foreground"
              }`}
            >
              <span className="block font-semibold">0{index + 1}</span>
              <span className="mt-1 block">{label}</span>
            </button>
          ))}
        </div>

        <div className="mt-6 space-y-5">
          {step === 0 && (
            <>
              <Field label="Task name">
                <input
                  autoFocus
                  maxLength={120}
                  required
                  className={inputClass}
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="e.g. Generate daily report"
                />
              </Field>

              <Field label="Description">
                <textarea
                  rows={3}
                  maxLength={1000}
                  className={textareaClass}
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  placeholder="What should this workflow accomplish?"
                />
              </Field>

              <Field label="Execution queue">
                <select
                  required
                  className={inputClass}
                  value={queueId}
                  onChange={(event) => setQueueId(event.target.value)}
                >
                  <option value="">Select an active queue</option>
                  {activeQueues.map((queue) => (
                    <option key={queue.id} value={queue.id}>
                      {queue.name}
                    </option>
                  ))}
                </select>
                {activeQueues.length === 0 && (
                  <span className="text-xs text-destructive">
                    No active queues found. Create or activate a queue first.
                  </span>
                )}
              </Field>

              <div className="grid gap-4 sm:grid-cols-2">
                <Field label="Timeout (seconds)">
                  <input
                    type="number"
                    min={1}
                    step={1}
                    className={inputClass}
                    value={timeoutSeconds}
                    onChange={(event) => setTimeoutSeconds(event.target.value)}
                  />
                </Field>

                <Field label="Maximum retries">
                  <input
                    type="number"
                    min={0}
                    step={1}
                    className={inputClass}
                    value={maxRetries}
                    onChange={(event) => setMaxRetries(event.target.value)}
                  />
                </Field>
              </div>
            </>
          )}

          {step === 1 && (
            <>
              <Field label="Trigger type">
                <select
                  className={inputClass}
                  value={triggerType}
                  onChange={(event) =>
                    setTriggerType(event.target.value as TriggerType)
                  }
                >
                  <option value="ONCE">Once</option>
                  <option value="DAILY">Daily</option>
                  <option value="WEEKLY">Weekly</option>
                  <option value="MONTHLY">Monthly</option>
                  <option value="YEARLY">Yearly</option>
                  <option value="CRON">Custom cron schedule</option>
                </select>
              </Field>

              {triggerType === "CRON" && (
                <Field label="Cron expression">
                  <input
                    className={`${inputClass} font-mono`}
                    value={cronExpression}
                    onChange={(event) => setCronExpression(event.target.value)}
                    placeholder="0 9 * * *"
                  />
                  <span className="text-xs text-muted-foreground">
                    Example: 0 9 * * * means 09:00 every day.
                  </span>
                </Field>
              )}

              <div className="grid gap-4 sm:grid-cols-2">
                <Field label="Start date">
                  <input
                    type="date"
                    className={inputClass}
                    value={startDate}
                    onChange={(event) => setStartDate(event.target.value)}
                  />
                </Field>

                <Field label="Start time">
                  <input
                    type="time"
                    className={inputClass}
                    value={startTime}
                    onChange={(event) => setStartTime(event.target.value)}
                  />
                </Field>
              </div>

              <Field label="Time zone">
                <input
                  className={inputClass}
                  value={timezone}
                  onChange={(event) => setTimezone(event.target.value)}
                  placeholder="Asia/Dhaka"
                />
              </Field>

              <p className="rounded-lg border border-border bg-muted/30 p-3 text-xs leading-5 text-muted-foreground">
                The selected start date and time are interpreted as Bangladesh
                local time (UTC+06:00). Recurring schedules use the chosen time
                zone.
              </p>
            </>
          )}

          {step === 2 && (
            <>
              <div className="flex items-center justify-between gap-3">
                <div>
                  <h3 className="font-semibold">Workflow actions</h3>
                  <p className="text-xs text-muted-foreground">
                    Actions execute in the order shown.
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setActions((current) => [...current, newAction()])}
                >
                  <Plus className="mr-1 size-4" />
                  Add action
                </Button>
              </div>

              {actions.map((action, index) => (
                <section
                  key={action.id}
                  className="space-y-4 rounded-xl border border-border p-4"
                >
                  <div className="flex items-center justify-between gap-3">
                    <h4 className="text-sm font-semibold">
                      Step {index + 1}
                    </h4>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      disabled={actions.length === 1}
                      onClick={() =>
                        setActions((current) =>
                          current.filter((item) => item.id !== action.id),
                        )
                      }
                    >
                      <Trash2 className="mr-1 size-4" />
                      Remove
                    </Button>
                  </div>

                  <Field label="Action type">
                    <select
                      className={inputClass}
                      value={action.type}
                      onChange={(event) =>
                        updateAction(action.id, {
                          type: event.target.value as ActionType,
                        })
                      }
                    >
                      <option value="REMINDER">Reminder</option>
                      <option value="EMAIL">Email</option>
                      <option value="HTTP">HTTP request</option>
                      <option value="SHELL">Shell command</option>
                    </select>
                  </Field>

                  {renderActionFields(action)}

                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={action.continueOnFailure}
                      onChange={(event) =>
                        updateAction(action.id, {
                          continueOnFailure: event.target.checked,
                        })
                      }
                    />
                    Continue if this action fails
                  </label>
                </section>
              ))}
            </>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <ReviewRow label="Task" value={name || "Untitled task"} />
              <ReviewRow
                label="Queue"
                value={
                  activeQueues.find((queue) => queue.id === queueId)?.name ??
                  "No queue selected"
                }
              />
              <ReviewRow
                label="Timeout / retries"
                value={`${timeoutSeconds}s / ${maxRetries} retries`}
              />
              <ReviewRow
                label="Schedule"
                value={`${triggerType} · ${startDate} ${startTime} · ${timezone}`}
              />
              {triggerType === "CRON" && (
                <ReviewRow label="Cron" value={cronExpression} />
              )}

              <div className="rounded-xl border border-border p-4">
                <h3 className="font-semibold">Actions in execution order</h3>
                <ol className="mt-3 space-y-2">
                  {actions.map((action, index) => (
                    <li
                      key={action.id}
                      className="flex items-center gap-3 text-sm"
                    >
                      <span className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-xs font-semibold text-primary">
                        {index + 1}
                      </span>
                      <span>
                        {{
                          REMINDER: "Reminder",
                          EMAIL: "Email",
                          HTTP: "HTTP request",
                          SHELL: "Shell command",
                        }[action.type]}
                      </span>
                      {action.continueOnFailure && (
                        <span className="text-xs text-muted-foreground">
                          · continue on failure
                        </span>
                      )}
                    </li>
                  ))}
                </ol>
              </div>

              <p className="text-xs leading-5 text-muted-foreground">
                Review your configuration before creating it. The task,
                trigger, and actions are saved through separate API requests.
              </p>
            </div>
          )}

          {error && (
            <div
              role="alert"
              className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive"
            >
              <AlertCircle className="mt-0.5 size-4 shrink-0" />
              <p>{error}</p>
            </div>
          )}

          <div className="flex flex-col-reverse justify-between gap-2 border-t border-border pt-4 sm:flex-row">
            <Button
              type="button"
              variant="outline"
              disabled={saving}
              onClick={() => {
                if (step === 0) onClose();
                else {
                  setError("");
                  setStep((current) => current - 1);
                }
              }}
            >
              {step === 0 ? (
                "Cancel"
              ) : (
                <>
                  <ArrowLeft className="mr-1 size-4" />
                  Back
                </>
              )}
            </Button>

            {step < steps.length - 1 ? (
              <Button type="button" disabled={saving} onClick={nextStep}>
                Continue
                <ArrowRight className="ml-1 size-4" />
              </Button>
            ) : (
              <Button
                type="button"
                disabled={saving || activeQueues.length === 0}
                onClick={() => void submitWorkflow()}
              >
                {saving ? "Creating workflow…" : "Create workflow"}
              </Button>
            )}
          </div>
        </div>
      </section>
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium">{label}</label>
      {children}
    </div>
  );
}

function ReviewRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1 border-b border-border/60 pb-3 sm:flex-row sm:justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="break-words text-sm font-medium sm:text-right">
        {value}
      </span>
    </div>
  );
}

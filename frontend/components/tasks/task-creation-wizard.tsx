"use client";

import { useEffect, useId, useState, useSyncExternalStore } from "react";
import {
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  Plus,
  Trash2,
  Workflow,
  X,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { createTaskSetup } from "@/lib/api/task-setups";

import {
  createAction,
  deleteAction,
  getTaskActions,
  getTaskTriggers,
  updateAction,
  updateTrigger,
  type ActionType,
  type Trigger,
  type TriggerType,
} from "@/lib/api/automation";

import { updateTask, type Queue, type Task } from "@/lib/api/dashboard";

type Props = {
  queues: Queue[];
  onClose: () => void;
  onCreated: () => void;
  task?: Task;
};

type ActionDraft = {
  id: string;
  persistedId?: string;
  connectionId?: string;
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
const subscribe = () => () => {};
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

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

function getBrowserTimezone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
}

function getTimezoneOptions(preferredTimezone: string) {
  const supportedTimezones =
    typeof Intl.supportedValuesOf === "function"
      ? Intl.supportedValuesOf("timeZone")
      : [
          "Africa/Cairo",
          "America/Chicago",
          "America/Denver",
          "America/Los_Angeles",
          "America/New_York",
          "Asia/Dhaka",
          "Asia/Dubai",
          "Asia/Kolkata",
          "Asia/Singapore",
          "Asia/Tokyo",
          "Australia/Sydney",
          "Europe/Berlin",
          "Europe/London",
          "Pacific/Auckland",
        ];

  return Array.from(
    new Set(["UTC", preferredTimezone, ...supportedTimezones]),
  ).sort((first, second) => first.localeCompare(second));
}

function formatTimezoneOffset(timeZone: string, date: Date) {
  try {
    const offset = new Intl.DateTimeFormat("en", {
      timeZone,
      timeZoneName: "shortOffset",
    })
      .formatToParts(date)
      .find((part) => part.type === "timeZoneName")?.value;

    if (!offset || offset === "GMT") return "UTC+00:00";

    const match = /^GMT([+-])(\d{1,2})(?::(\d{2}))?$/.exec(offset);
    if (!match) return offset;

    const [, sign, hours, minutes = "00"] = match;
    return `UTC${sign}${hours.padStart(2, "0")}:${minutes}`;
  } catch {
    return "UTC offset unavailable";
  }
}

function TimezonePicker({
  value,
  timezones,
  referenceDate,
  onChange,
}: {
  value: string;
  timezones: string[];
  referenceDate: Date;
  onChange: (timeZone: string) => void;
}) {
  const listboxId = useId();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const normalizedQuery = query.trim().toLowerCase();

  const filteredTimezones = normalizedQuery
    ? timezones
        .filter((timeZone) => {
          const searchText = `${timeZone} ${timeZone
            .split("/")
            .at(-1)
            ?.replaceAll("_", " ")}`.toLowerCase();
          return searchText.includes(normalizedQuery);
        })
        .slice(0, 80)
    : [];

  function choose(timeZone: string) {
    onChange(timeZone);
    setQuery("");
    setOpen(false);
  }

  return (
    <div className="relative">
      <input
        role="combobox"
        aria-label="Search time zones by city or region"
        aria-autocomplete="list"
        aria-expanded={open}
        aria-controls={listboxId}
        autoComplete="off"
        className={inputClass}
        value={
          open
            ? query
            : `${formatTimezoneOffset(value, referenceDate)} · ${value}`
        }
        placeholder="Search city or time zone"
        onFocus={() => {
          setQuery("");
          setActiveIndex(0);
          setOpen(true);
        }}
        onBlur={() => setOpen(false)}
        onChange={(event) => {
          setQuery(event.target.value);
          setActiveIndex(0);
          setOpen(true);
        }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown") {
            event.preventDefault();
            setOpen(true);
            setActiveIndex((index) =>
              Math.min(index + 1, filteredTimezones.length - 1),
            );
          } else if (event.key === "ArrowUp") {
            event.preventDefault();
            setActiveIndex((index) => Math.max(index - 1, 0));
          } else if (
            event.key === "Enter" &&
            open &&
            filteredTimezones[activeIndex]
          ) {
            event.preventDefault();
            choose(filteredTimezones[activeIndex]);
          } else if (event.key === "Escape") {
            setOpen(false);
          }
        }}
      />

      {open && (
        <div
          id={listboxId}
          role="listbox"
          className="absolute z-20 mt-1 max-h-64 w-full overflow-y-auto rounded-lg border border-border bg-popover p-1 text-popover-foreground shadow-lg"
        >
          {filteredTimezones.length > 0 ? (
            filteredTimezones.map((timeZone, index) => (
              <button
                key={timeZone}
                type="button"
                role="option"
                aria-selected={timeZone === value}
                className={`flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm ${
                  index === activeIndex
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-muted"
                }`}
                onMouseDown={(event) => event.preventDefault()}
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => choose(timeZone)}
              >
                <span className="truncate">{timeZone}</span>
                <span className="shrink-0 text-xs text-muted-foreground">
                  {formatTimezoneOffset(timeZone, referenceDate)}
                </span>
              </button>
            ))
          ) : normalizedQuery ? (
            <p className="px-3 py-2 text-sm text-muted-foreground">
              No matching time zones.
            </p>
          ) : (
            <p className="px-3 py-2 text-sm text-muted-foreground">
              Type a city or region to search time zones.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function dateTimePartsInTimezone(date: Date, timeZone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  })
    .formatToParts(date)
    .reduce<Record<string, string>>((parts, part) => {
      if (part.type !== "literal") parts[part.type] = part.value;
      return parts;
    }, {});
}

function dateTimeAfterOneHour(timeZone: string, timestamp: number) {
  const parts = dateTimePartsInTimezone(new Date(timestamp), timeZone);

  return {
    date: `${parts.year}-${parts.month}-${parts.day}`,
    time: `${parts.hour}:${parts.minute}`,
  };
}

function localDateTimeInTimezone(
  date: string,
  time: string,
  timeZone: string,
): Date | null {
  const dateMatch = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date);
  const timeMatch = /^(\d{2}):(\d{2})$/.exec(time);
  if (!dateMatch || !timeMatch) return null;

  const [, yearText, monthText, dayText] = dateMatch;
  const [, hourText, minuteText] = timeMatch;
  const year = Number(yearText);
  const month = Number(monthText);
  const day = Number(dayText);
  const hour = Number(hourText);
  const minute = Number(minuteText);

  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > 31 ||
    hour > 23 ||
    minute > 59
  ) {
    return null;
  }

  const requestedTime = new Date(0);
  requestedTime.setUTCFullYear(year, month - 1, day);
  requestedTime.setUTCHours(hour, minute, 0, 0);
  if (
    requestedTime.getUTCFullYear() !== year ||
    requestedTime.getUTCMonth() !== month - 1 ||
    requestedTime.getUTCDate() !== day
  ) {
    return null;
  }

  const requestedTimestamp = requestedTime.getTime();
  let candidateTimestamp = requestedTimestamp;

  try {
    for (let attempt = 0; attempt < 4; attempt += 1) {
      const parts = dateTimePartsInTimezone(
        new Date(candidateTimestamp),
        timeZone,
      );
      const representedTime = new Date(0);
      representedTime.setUTCFullYear(
        Number(parts.year),
        Number(parts.month) - 1,
        Number(parts.day),
      );
      representedTime.setUTCHours(
        Number(parts.hour),
        Number(parts.minute),
        0,
        0,
      );

      const offset = representedTime.getTime() - candidateTimestamp;
      const nextCandidate = requestedTimestamp - offset;
      if (nextCandidate === candidateTimestamp) break;
      candidateTimestamp = nextCandidate;
    }

    const result = new Date(candidateTimestamp);
    const resultParts = dateTimePartsInTimezone(result, timeZone);
    if (
      resultParts.year !== yearText ||
      resultParts.month !== monthText ||
      resultParts.day !== dayText ||
      resultParts.hour !== hourText ||
      resultParts.minute !== minuteText
    ) {
      return null;
    }

    return result;
  } catch {
    return null;
  }
}

function configurationFor(action: ActionDraft): Record<string, unknown> {
  switch (action.type) {
    case "REMINDER":
      return { message: action.message.trim() };

    case "EMAIL":
      return {
        sender: action.sender,
        ...(action.sender === "USER_GMAIL" && action.connectionId
          ? { connection_id: action.connectionId }
          : {}),
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
  task,
}: Props) {
  const mounted = useSyncExternalStore(
    subscribe,
    getClientSnapshot,
    getServerSnapshot,
  );
  const browserTimezone = mounted ? getBrowserTimezone() : "UTC";
  const timezoneOptions = mounted
    ? getTimezoneOptions(browserTimezone)
    : ["UTC"];
  const [defaultStartTimestamp] = useState(() => Date.now() + 60 * 60 * 1000);
  const [step, setStep] = useState(0);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [timeoutSeconds, setTimeoutSeconds] = useState("60");
  const [maxRetries, setMaxRetries] = useState("3");

  const [triggerType, setTriggerType] = useState<TriggerType>("ONCE");
  const [startDate, setStartDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [cronExpression, setCronExpression] = useState("0 9 * * *");
  const [timezoneOverride, setTimezoneOverride] = useState<string | null>(null);
  const timezone = timezoneOverride ?? browserTimezone;

  const [actions, setActions] = useState<ActionDraft[]>([newAction()]);

  const [persistedTrigger, setPersistedTrigger] = useState<Trigger | null>(
    null,
  );
  const [originalActionIds, setOriginalActionIds] = useState<string[]>([]);

  useEffect(() => {
    if (!task) return;

    const editingTask = task;

    let cancelled = false;

    async function loadTaskForEditing() {
      setSaving(true);
      setError("");

      try {
        const [triggerResponse, actionResponse] = await Promise.all([
          getTaskTriggers(editingTask.id),
          getTaskActions(editingTask.id),
        ]);

        if (cancelled) return;

        const triggers = triggerResponse.data;
        const existingActions = actionResponse.data;

        if (triggers.length === 0) {
          throw new Error(
            "This task has no trigger. Editing its schedule is not supported yet.",
          );
        }

        if (triggers.length > 1) {
          throw new Error(
            "This task has multiple triggers. Editing is blocked to avoid accidentally changing the wrong schedule.",
          );
        }

        const trigger = triggers[0];

        setPersistedTrigger(trigger);
        setOriginalActionIds(existingActions.map((action) => action.id));

        setName(editingTask.name);
        setDescription(editingTask.description ?? "");
        setTimeoutSeconds(String(editingTask.timeout_seconds));
        setMaxRetries(String(editingTask.max_retries));

        setTriggerType(trigger.trigger_type);
        setTimezoneOverride(trigger.timezone || "UTC");
        setCronExpression(trigger.cron_expression ?? "0 9 * * *");

        const scheduledAt = new Date(trigger.start_at);

        if (Number.isNaN(scheduledAt.getTime())) {
          throw new Error("The task has an invalid scheduled date.");
        }

        const scheduleParts = dateTimePartsInTimezone(
          scheduledAt,
          trigger.timezone || "UTC",
        );

        setStartDate(
          `${scheduleParts.year}-${scheduleParts.month}-${scheduleParts.day}`,
        );
        setStartTime(`${scheduleParts.hour}:${scheduleParts.minute}`);

        const decodeConfiguration = (
          encoded: string,
        ): Record<string, unknown> => {
          try {
            const binary = atob(encoded);
            const bytes = Uint8Array.from(binary, (character) =>
              character.charCodeAt(0),
            );
            const decoded = new TextDecoder().decode(bytes);
            const parsed: unknown = JSON.parse(decoded);

            if (
              parsed === null ||
              typeof parsed !== "object" ||
              Array.isArray(parsed)
            ) {
              throw new Error("Expected a JSON object.");
            }

            return parsed as Record<string, unknown>;
          } catch {
            throw new Error(
              "Could not read an action's saved configuration. No changes have been saved.",
            );
          }
        };

        const drafts: ActionDraft[] = existingActions
          .slice()
          .sort((a, b) => a.execution_order - b.execution_order)
          .map((existingAction) => {
            const config = decodeConfiguration(existingAction.configuration);
            const draft = newAction();

            draft.persistedId = existingAction.id;
            draft.type = existingAction.action_type;
            draft.continueOnFailure = existingAction.continue_on_failure;

            if (draft.type === "REMINDER") {
              draft.message =
                typeof config.message === "string" ? config.message : "";
            }

            if (draft.type === "EMAIL") {
              draft.sender =
                config.sender === "USER_GMAIL" ? "USER_GMAIL" : "STAQ";
              draft.connectionId =
                typeof config.connection_id === "string"
                  ? config.connection_id
                  : undefined;
              draft.to = typeof config.to === "string" ? config.to : "";
              draft.subject =
                typeof config.subject === "string" ? config.subject : "";
              draft.body = typeof config.body === "string" ? config.body : "";
            }

            if (draft.type === "HTTP") {
              const method = config.method;
              draft.method =
                method === "GET" ||
                method === "POST" ||
                method === "PUT" ||
                method === "PATCH" ||
                method === "DELETE"
                  ? method
                  : "POST";

              draft.url = typeof config.url === "string" ? config.url : "";
              draft.body = typeof config.body === "string" ? config.body : "";
              draft.headers = JSON.stringify(
                config.headers &&
                  typeof config.headers === "object" &&
                  !Array.isArray(config.headers)
                  ? config.headers
                  : {},
                null,
                2,
              );
            }

            if (draft.type === "SHELL") {
              draft.command =
                typeof config.command === "string" ? config.command : "echo";
              draft.args = Array.isArray(config.args)
                ? config.args
                    .filter(
                      (value): value is string => typeof value === "string",
                    )
                    .join("\n")
                : "";
            }

            return draft;
          });

        setActions(drafts.length > 0 ? drafts : [newAction()]);
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof Error
              ? err.message
              : "Failed to load the task for editing.",
          );
        }
      } finally {
        if (!cancelled) setSaving(false);
      }
    }

    void loadTaskForEditing();

    return () => {
      cancelled = true;
    };
  }, [task]);

  const defaultQueue = queues.find(
    (queue) => queue.name === "default" && queue.is_active,
  );
  const defaultSchedule = mounted
    ? dateTimeAfterOneHour(browserTimezone, defaultStartTimestamp)
    : { date: "", time: "" };
  const scheduleDate = startDate || defaultSchedule.date;
  const scheduleTime = startTime || defaultSchedule.time;
  const timezoneReferenceDate =
    localDateTimeInTimezone(scheduleDate, scheduleTime, timezone) ??
    new Date(defaultStartTimestamp);

  function updateActionDraft(id: string, patch: Partial<ActionDraft>) {
    setActions((current) =>
      current.map((action) =>
        action.id === id ? { ...action, ...patch } : action,
      ),
    );
}

  function validateStep(index: number): string | null {
    if (index === 0) {
      if (!name.trim()) return "Enter a task name.";
      if (!defaultQueue) return "The active default queue is unavailable.";

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
      if (!scheduleDate || !scheduleTime) {
        return "Choose a start date and time.";
      }

      const startAt = localDateTimeInTimezone(
        scheduleDate,
        scheduleTime,
        timezone.trim(),
      );

      if (!startAt) return "Choose a valid time and IANA time zone.";

      if (triggerType === "ONCE" && startAt.getTime() <= Date.now()) {
        return "A one-time trigger must be scheduled in the future.";
      }

      if (triggerType === "CRON" && !cronExpression.trim()) {
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

    if (task && !persistedTrigger) {
      setError(
        "The existing trigger has not loaded. Please close and reopen the editor.",
      );
      return;
    }

    setSaving(true);

    try {
      const scheduledStart = localDateTimeInTimezone(
        scheduleDate,
        scheduleTime,
        timezone.trim(),
      );

      if (!scheduledStart) {
        throw new Error(
          "The selected local time is not valid in this time zone.",
        );
      }

      const startAt = scheduledStart.toISOString();
      const timeout = Number(timeoutSeconds);
      const retries = Number(maxRetries);

      if (!task) {
        // CREATE MODE: keep the existing transactional setup endpoint.
        await createTaskSetup({
          name: name.trim(),
          description: description.trim() || null,
          timeout_seconds: timeout,
          max_retries: retries,
          trigger_type: triggerType,
          timezone: timezone.trim(),
          start_at: startAt,
          ...(triggerType === "CRON"
            ? { cron_expression: cronExpression.trim() }
            : {}),
          actions: actions.map((draft, index) => ({
            action_type: draft.type,
            execution_order: index + 1,
            configuration: configurationFor(draft),
            continue_on_failure: draft.continueOnFailure,
          })),
        });
      } else {
        // EDIT MODE: update existing records rather than creating a new task.
        const trigger = persistedTrigger!;

        await updateTask(task.id, {
          name: name.trim(),
          description: description.trim() || null,
          status: task.status,
          timeout_seconds: timeout,
          max_retries: retries,
        });

        await updateTrigger(trigger.id, {
          trigger_type: triggerType,
          ...(triggerType === "CRON"
            ? { cron_expression: cronExpression.trim() }
            : { cron_expression: null }),
          timezone: timezone.trim(),
          start_at: startAt,
          is_active: trigger.is_active,
        });

        const currentPersistedIds = new Set(
          actions
            .map((action) => action.persistedId)
            .filter((id): id is string => Boolean(id)),
        );

        // Delete original actions removed from the editor.
        for (const originalId of originalActionIds) {
          if (!currentPersistedIds.has(originalId)) {
            await deleteAction(originalId);
          }
        }

        // Update existing actions and create newly added actions.
        for (let index = 0; index < actions.length; index += 1) {
          const draft = actions[index];

          const payload = {
            action_type: draft.type,
            execution_order: index + 1,
            configuration: configurationFor(draft),
            continue_on_failure: draft.continueOnFailure,
          };

          if (draft.persistedId) {
            await updateAction(draft.persistedId, payload);
          } else {
            await createAction({
              task_id: task.id,
              ...payload,
            });
          }
        }
      }

      onCreated();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not save the workflow.",
      );
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
              updateActionDraft(action.id, { message: event.target.value })
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
                updateActionDraft(action.id, {
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
                updateActionDraft(action.id, { to: event.target.value })
              }
              placeholder="recipient@example.com"
            />
          </Field>

          <Field label="Subject">
            <input
              className={inputClass}
              value={action.subject}
              onChange={(event) =>
                updateActionDraft(action.id, { subject: event.target.value })
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
                updateActionDraft(action.id, { body: event.target.value })
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
                updateActionDraft(action.id, {
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
                updateActionDraft(action.id, { url: event.target.value })
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
                updateActionDraft(action.id, { headers: event.target.value })
              }
            />
          </Field>

          <Field label="Request body">
            <textarea
              rows={4}
              className={`${textareaClass} font-mono text-xs`}
              value={action.body}
              onChange={(event) =>
                updateActionDraft(action.id, { body: event.target.value })
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
              updateActionDraft(action.id, { command: event.target.value })
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
              updateActionDraft(action.id, { args: event.target.value })
            }
            placeholder={"hello\nworld"}
          />
        </Field>

        <p className="text-xs text-muted-foreground">
          Shell execution can be dangerous. Only configure commands permitted by
          your backend worker&apos;s allowlist.
        </p>
      </>
    );
  }

  return (
    <div
      className="fixed inset-0 z-60 flex items-center justify-center overflow-y-auto bg-black/50 p-3 backdrop-blur-sm sm:p-5"
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
              {task ? "Edit automation task" : "Create an automation task"}
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
                <div
                  className="flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-input bg-muted/40 px-3 text-sm font-medium text-foreground"
                  aria-readonly="true"
                >
                  {defaultQueue ? (
                    <>
                      <span className="size-2 rounded-full bg-emerald-500" />
                      <span>Default queue</span>
                    </>
                  ) : (
                    <span className="text-destructive">
                      Default queue unavailable
                    </span>
                  )}
                </div>
                {!defaultQueue && (
                  <span className="text-xs text-destructive">
                    The backend requires its active default queue.
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
                    value={scheduleDate}
                    onChange={(event) => setStartDate(event.target.value)}
                  />
                </Field>

                <Field label="Start time">
                  <input
                    type="time"
                    className={inputClass}
                    value={scheduleTime}
                    onChange={(event) => setStartTime(event.target.value)}
                  />
                </Field>
              </div>

              <Field label="Time zone">
                <TimezonePicker
                  value={timezone}
                  timezones={timezoneOptions}
                  referenceDate={timezoneReferenceDate}
                  onChange={setTimezoneOverride}
                />
              </Field>

              <p className="rounded-lg border border-border bg-muted/30 p-3 text-xs leading-5 text-muted-foreground">
                The selected date and time use this IANA time zone. Daylight
                saving gaps are rejected. A recurring CRON trigger starts at its
                next matching time after this date and time.
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
                  onClick={() =>
                    setActions((current) => [...current, newAction()])
                  }
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
                    <h4 className="text-sm font-semibold">Step {index + 1}</h4>
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
                        updateActionDraft(action.id, {
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
                        updateActionDraft(action.id, {
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
                value={defaultQueue?.name ?? "Default queue unavailable"}
              />
              <ReviewRow
                label="Timeout / retries"
                value={`${timeoutSeconds}s / ${maxRetries} retries`}
              />
              <ReviewRow
                label="Schedule"
                value={`${triggerType} · ${scheduleDate} ${scheduleTime} · ${timezone}`}
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
                        {
                          {
                            REMINDER: "Reminder",
                            EMAIL: "Email",
                            HTTP: "HTTP request",
                            SHELL: "Shell command",
                          }[action.type]
                        }
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
                Review your configuration before creating it. The task, trigger,
                and actions are saved through separate API requests.
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
                disabled={saving || !defaultQueue}
                onClick={() => void submitWorkflow()}
              >
                {saving
                  ? "Saving..."
                  : task
                    ? "Save changes"
                    : "Create workflow"}
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
      <span className="wrap-break-word text-sm font-medium sm:text-right">
        {value}
      </span>
    </div>
  );
}

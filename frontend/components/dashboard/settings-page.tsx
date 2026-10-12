"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AlertCircle, CheckCircle2, RefreshCw, Save, ShieldCheck } from "lucide-react";

import Sidebar from "@/components/dashboard/sidebar";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, updateCurrentUser, type CurrentUser } from "@/lib/api/dashboard";
import { clearTokens, getAccessToken } from "@/lib/auth/storage";
import { logoutUser } from "@/lib/auth/session";

const inputClass = "h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring";

function timezoneOptions(browserTimezone: string) {
  const supported = typeof Intl.supportedValuesOf === "function"
    ? Intl.supportedValuesOf("timeZone")
    : ["Asia/Dhaka", "America/Los_Angeles", "America/New_York", "Europe/London", "UTC"];
  return Array.from(new Set(["UTC", browserTimezone, ...supported])).sort((a, b) => a.localeCompare(b));
}

export default function SettingsPage() {
  const router = useRouter();
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [fullName, setFullName] = useState("");
  const [timezone, setTimezone] = useState("UTC");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [refreshVersion, setRefreshVersion] = useState(0);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const timezones = timezoneOptions(timezone || "UTC");

  useEffect(() => {
    let cancelled = false;
    async function loadProfile() {
      if (!getAccessToken()) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setLoading(true);
      setError("");
      try {
        const response = await getCurrentUser();
        if (cancelled) return;
        setUser(response.data);
        setFullName(response.data.full_name);
        setTimezone(response.data.timezone || "UTC");
      } catch (caught) {
        if (cancelled) return;
        if (caught instanceof ApiError && caught.status === 401) {
          clearTokens();
          router.replace("/login");
          return;
        }
        setError(caught instanceof Error ? caught.message : "Could not load your settings.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void loadProfile();
    return () => {
      cancelled = true;
    };
  }, [refreshVersion, router]);

  async function handleSave(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setNotice("");
    setSaving(true);
    try {
      const response = await updateCurrentUser({
        full_name: fullName.trim(),
        timezone,
      });
      setUser(response.data);
      setFullName(response.data.full_name);
      setTimezone(response.data.timezone);
      setNotice("Profile and timezone preference saved.");
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 401) {
        clearTokens();
        router.replace("/login");
        return;
      }
      setError(caught instanceof Error ? caught.message : "Could not save your settings.");
    } finally {
      setSaving(false);
    }
  }

  async function handleLogout() {
    setLoggingOut(true);
    try {
      await logoutUser();
    } catch {
      // Clear the local session even if the API request fails.
    } finally {
      clearTokens();
      router.replace("/login");
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <Sidebar mobileOpen={mobileOpen} onClose={() => setMobileOpen(false)} user={user} onLogout={handleLogout} loggingOut={loggingOut} activePage="settings" />
      <main className="min-h-screen lg:pl-66">
        <header className="sticky top-0 z-30 flex h-19 items-center justify-between border-b border-border/80 bg-background/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button type="button" aria-label="Open navigation" onClick={() => setMobileOpen(true)} className="rounded-lg border border-border p-2 text-muted-foreground hover:bg-muted lg:hidden"><ShieldCheck className="size-4" /></button>
            <div><p className="text-sm font-semibold">Settings</p><p className="text-xs text-muted-foreground">Profile and scheduling preference</p></div>
          </div>
          <Button type="button" variant="outline" size="sm" disabled={loading} onClick={() => setRefreshVersion((version) => version + 1)} className="gap-2"><RefreshCw className="size-3.5" />Refresh</Button>
        </header>

        <div className="mx-auto w-full max-w-360 px-4 py-7 sm:px-6 sm:py-9 lg:px-9">
          <div><p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Account</p><h1 className="mt-2 text-2xl font-bold">Profile settings</h1><p className="mt-2 text-sm text-muted-foreground">Manage the profile details and timezone used by scheduled workflows.</p></div>

          {error && <div role="alert" className="mt-6 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive"><AlertCircle className="size-4" />{error}</div>}
          {notice && <div role="status" className="mt-6 flex items-center gap-2 rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-3 text-sm text-emerald-700 dark:text-emerald-300"><CheckCircle2 className="size-4" />{notice}</div>}

          <form onSubmit={(event) => void handleSave(event)} className="mt-7 max-w-2xl space-y-6 border-y border-border/70 py-6">
            <div className="space-y-2">
              <label htmlFor="settings-name" className="text-sm font-medium">Full name</label>
              <input id="settings-name" className={inputClass} minLength={2} maxLength={100} required value={fullName} disabled={loading || saving} onChange={(event) => setFullName(event.target.value)} />
            </div>

            <div className="space-y-2">
              <label htmlFor="settings-email" className="text-sm font-medium">Email</label>
              <input id="settings-email" className={`${inputClass} bg-muted/40 text-muted-foreground`} value={user?.email ?? ""} readOnly aria-readonly="true" />
              <p className="text-xs text-muted-foreground">Email changes are not supported by the current account API.</p>
            </div>

            <div className="space-y-2">
              <label htmlFor="settings-timezone" className="text-sm font-medium">Timezone</label>
              <select id="settings-timezone" className={inputClass} value={timezone} disabled={loading || saving} onChange={(event) => setTimezone(event.target.value)}>
                {timezones.map((zone) => <option key={zone} value={zone}>{zone.replaceAll("_", " ")}</option>)}
              </select>
              <p className="text-xs text-muted-foreground">New schedules use this timezone by default. Existing triggers keep their own saved timezone.</p>
            </div>

            <div className="flex items-center justify-between gap-4 border-t border-border/70 pt-5">
              <p className="text-xs text-muted-foreground">Only fields supported by the account API can be changed here.</p>
              <Button type="submit" disabled={loading || saving || !fullName.trim()} className="gap-2"><Save className="size-4" />{saving ? "Saving…" : "Save changes"}</Button>
            </div>
          </form>
        </div>
      </main>
    </div>
  );
}
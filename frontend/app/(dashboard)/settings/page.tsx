import type { Metadata } from "next";
import SettingsPage from "@/components/dashboard/settings-page";

export const metadata: Metadata = {
  title: "Settings",
  description: "Manage your STAQ profile and timezone preference.",
};

export default function SettingsRoute() {
  return <SettingsPage />;
}
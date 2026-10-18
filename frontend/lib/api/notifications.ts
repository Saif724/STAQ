import { apiRequest } from "@/lib/api/client";

export type AppNotification = {
  id: string;
  task_id?: string | null;
  execution_id?: string | null;
  type: string;
  message: string;
  read_at?: string | null;
  created_at: string;
};

export type NotificationList = {
  items: AppNotification[];
  unread_count: number;
};

export function getNotifications(limit = 30) {
  return apiRequest<{ data: NotificationList }>(`/notifications?limit=${limit}`);
}

export function markNotificationRead(id: string) {
  return apiRequest<{ data: { read: boolean } }>(`/notifications/${id}/read`, {
    method: "POST",
  });
}

export function markAllNotificationsRead() {
  return apiRequest<{ data: { updated: boolean } }>(`/notifications/read-all`, {
    method: "POST",
  });
}
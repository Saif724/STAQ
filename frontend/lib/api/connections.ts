import { apiRequest } from "@/lib/api/client";

export type Connection = {
  id: string;
  provider: string;
  account_email: string;
  scopes: string[];
  token_expires_at?: string | null;
  created_at: string;
  updated_at: string;
};

type ResponseWrapper<T> = {
  data: T;
};

export function getConnections() {
  return apiRequest<ResponseWrapper<Connection[]>>("/connections");
}

export function deleteConnection(connectionId: string) {
  return apiRequest<void>(`/connections/${connectionId}`, {
    method: "DELETE",
  });
}

export function getGmailAuthorizationUrl() {
  return apiRequest<ResponseWrapper<{ authorization_url: string }>>(
    "/connections/google/authorization",
  );
}
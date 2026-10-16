import { config } from "@/lib/config";
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  setTokens,
} from "@/lib/auth/storage";

export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

type ApiRequestOptions = RequestInit & {
  auth?: boolean;
};

type ErrorResponse = {
  error?: {
    code?: string;
    message?: string;
  };
};

type TokenResponse = {
  data?: {
    access_token?: string;
    refresh_token?: string;
  };
  error?: {
    code?: string;
    message?: string;
  };
};

let refreshPromise: Promise<string> | null = null;

function buildHeaders(
  headers: HeadersInit | undefined,
  token?: string | null,
): Headers {
  const result = new Headers(headers);

  if (!result.has("Content-Type")) {
    result.set("Content-Type", "application/json");
  }

  if (token) {
    result.set("Authorization", `Bearer ${token}`);
  } else {
    result.delete("Authorization");
  }

  return result;
}

async function parseError(response: Response): Promise<ApiError> {
  let errorData: ErrorResponse | null = null;

  try {
    errorData = (await response.json()) as ErrorResponse;
  } catch {
    // The server may return an empty or non-JSON error response.
  }

  return new ApiError(
    errorData?.error?.message ?? "An unexpected error occurred.",
    response.status,
    errorData?.error?.code,
  );
}

async function refreshAccessTokenAutomatically(): Promise<string> {
  const refreshToken = getRefreshToken();

  if (!refreshToken) {
    clearTokens();
    throw new ApiError("Your session has expired. Please log in again.", 401);
  }

  const response = await fetch(`${config.apiUrl}/auth/refresh`, {
    method: "POST",
    headers: buildHeaders(undefined),
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok) {
    const error = await parseError(response);

    // Only clear the session when the server rejects the refresh
    // credentials. Keep tokens if the server has a temporary failure.
    if (
      response.status === 400 ||
      response.status === 401 ||
      response.status === 403
    ) {
      clearTokens();
    }

    throw error;
  }

  const result = (await response.json()) as TokenResponse;
  const newAccessToken = result.data?.access_token;
  const newRefreshToken = result.data?.refresh_token;

  if (!newAccessToken) {
    throw new ApiError(
      "The refresh endpoint did not return an access token.",
      502,
    );
  }

  // Preserve the old refresh token if the backend does not rotate it.
  // If the backend returns a replacement, store that replacement.
  setTokens(newAccessToken, newRefreshToken || refreshToken);

  return newAccessToken;
}

function getRefreshedAccessToken(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = refreshAccessTokenAutomatically().finally(() => {
      refreshPromise = null;
    });
  }

  return refreshPromise;
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<T> {
  const { auth = true, headers, ...requestInit } = options;

  // Keep the original request options so the request can be retried once.
  const makeRequest = (token: string | null) =>
    fetch(`${config.apiUrl}${path}`, {
      ...requestInit,
      headers: buildHeaders(headers, auth ? token : null),
    });

  let response = await makeRequest(auth ? getAccessToken() : null);

  // Public requests and non-auth failures must not trigger token refresh.
  if (auth && response.status === 401) {
    // Consume the response body before retrying; the first response is
    // no longer needed once we decide to refresh.
    try {
      await response.body?.cancel();
    } catch {
      // Ignore errors while disposing of the old response body.
    }

    const newAccessToken = await getRefreshedAccessToken();
    response = await makeRequest(newAccessToken);
  }

  if (!response.ok) {
    throw await parseError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

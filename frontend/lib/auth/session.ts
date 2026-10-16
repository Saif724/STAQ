import {
  login,
  register,
  logout,
  refreshAccessToken,
} from "@/lib/api/auth";

import {
  clearTokens,
  getRefreshToken,
  setTokens,
} from "@/lib/auth/storage";

import type {
  LoginRequest,
  RegisterRequest,
} from "@/types/auth";

export async function loginUser(data: LoginRequest): Promise<void> {
  const response = await login(data);

  setTokens(
    response.data.access_token,
    response.data.refresh_token,
  );
}

export async function registerUser(data: RegisterRequest) {
  return register(data);
}

export async function refreshSession(): Promise<void> {
  const refreshToken = getRefreshToken();

  if (!refreshToken) {
    clearTokens();
    throw new Error("No refresh token available. Please log in again.");
  }

  const response = await refreshAccessToken({
    refresh_token: refreshToken,
  });

  const data = response.data as typeof response.data & {
    refresh_token?: unknown;
  };

  const newAccessToken = data.access_token;
  const newRefreshToken =
    typeof data.refresh_token === "string" && data.refresh_token.length > 0
      ? data.refresh_token
      : refreshToken;

  setTokens(newAccessToken, newRefreshToken);
}

export async function logoutUser(): Promise<void> {
  const refreshToken = getRefreshToken();

  try {
    if (refreshToken) {
      await logout({
        refresh_token: refreshToken,
      });
    }
  } finally {
    clearTokens();
  }
}

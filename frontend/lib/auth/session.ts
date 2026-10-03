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

export async function loginUser(
    data: LoginRequest,
): Promise<void> {
    const response = await login(data);

    setTokens(
        response.data.access_token,
        response.data.refresh_token,
    );
}

export async function registerUser(
    data: RegisterRequest,
) {
    return register(data);
}

export async function refreshSession(): Promise<void> {
    const refreshToken = getRefreshToken();

    if (!refreshToken) {
        throw new Error("No refresh token available");
    }

    const response = await refreshAccessToken({
        refresh_token: refreshToken,
    });

    setTokens(
        response.data.access_token,
        refreshToken,
    );
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
import {apiRequest} from "@/lib/api/client";
import type {
    LoginRequest,
    LoginResponseWrapper,
    LogoutRequest,
    LogoutResponseWrapper,
    RefreshRequest,
    RefreshResponseWrapper,
    RegisterRequest,
    RegisterResponseWrapper,
} from "@/types/auth";

export async function login(
    data: LoginRequest,
): Promise<LoginResponseWrapper> {
    return apiRequest<LoginResponseWrapper>("/auth/login", {
        method: "POST",
        auth: false,
        body: JSON.stringify(data),
    });
}

export async function register(
    data: RegisterRequest,
): Promise<RegisterResponseWrapper> {
    return apiRequest<RegisterResponseWrapper>("/auth/register", {
        method: "POST",
        auth: false,
        body: JSON.stringify(data),
    });
}

export async function refreshAccessToken(
    data: RefreshRequest,
): Promise<RefreshResponseWrapper> {
    return apiRequest<RefreshResponseWrapper>("/auth/refresh", {
        method: "POST",
        auth: false,
        body: JSON.stringify(data),
    });
}

export async function logout(
    data: LogoutRequest,
): Promise<LogoutResponseWrapper> {
    return apiRequest<LogoutResponseWrapper>("/auth/logout", {
        method: "POST",
        auth: false,
        body: JSON.stringify(data),
    });
}
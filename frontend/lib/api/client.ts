import {config} from "@/lib/config";
import {getAccessToken} from "@/lib/auth/storage";

export class ApiError extends Error {
    readonly status: number;
    readonly code?: string;

    constructor(
        message: string,
        status: number,
        code?: string,
    ){
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

export async function apiRequest<T>(
    path: string,
    options: ApiRequestOptions = {},
): Promise<T> {
    const { auth = true, headers, ...requestInit } = options;
    
    const requestHeaders = new Headers(headers);

    requestHeaders.set("Content-Type", "application/json");

    if (auth) {
        const token = getAccessToken();

        if (token) {
            requestHeaders.set("Authorization", `Bearer ${token}`);
        }
    }

    const response = await fetch(`${config.apiUrl}${path}`, {
        ...requestInit,
        headers: requestHeaders,
    });

    if (!response.ok) {
        let errorData: ErrorResponse | null = null;

        try {
            errorData = (await response.json()) as ErrorResponse;
        } catch {

        }

        throw new ApiError(
            errorData?.error?.message ?? "An unexpected error occured",
            response.status,
            errorData?.error?.code,
        );
    }

    if (response.status === 204) {
        return undefined as T;
    }

    return (await response.json()) as T;
}
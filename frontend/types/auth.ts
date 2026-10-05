export type RegisterRequest = {
    full_name: string;
    email: string;
    password: string;
};

export type RegisterResponse = {
    id: string;
    full_name: string;
    email: string;
    email_verified: boolean;
};

export type RegisterResponseWrapper = {
    data: RegisterResponse;
}

export type LoginRequest = {
    email: string;
    password: string;
};

export type LoginResponse = {
    access_token: string;
    refresh_token: string;
};

export type LoginResponseWrapper = {
    data: LoginResponse;
};

export type RefreshRequest = {
    refresh_token: string;
};

export type RefreshResponse = {
    access_token: string;
};

export type RefreshResponseWrapper = {
    data: RefreshResponse;
};

export type LogoutRequest = {
    refresh_token: string;
};

export type LogoutResponseWrapper = {
    data: string;
};

export type VerifyEmailRequest = {
    email: string;
    code: string;
};

export type ResendVerificationRequest = {
    email: string;
};

export type MessageResponse = {
    message: string;
};

export type MessageResponseWrapper = {
    data: MessageResponse;
};

export type VerifyEmailResponseWrapper = MessageResponseWrapper;

export type ResendVerificationResponseWrapper = MessageResponseWrapper;
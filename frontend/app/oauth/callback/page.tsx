"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { LoaderCircle, Zap } from "lucide-react";

import { exchangeGoogleLoginCode } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import { setTokens } from "@/lib/auth/storage";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";

function GoogleCallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();

    const code = searchParams.get("code");

    const [error, setError] = useState("");

    useEffect(() => {
        if (!code) return;

        let cancelled = false;

        async function completeLogin() {
            try {
                const response = await exchangeGoogleLoginCode(code!);

                if (cancelled) return;

                setTokens(
                    response.data.access_token,
                    response.data.refresh_token,
                );

                router.replace("/dashboard");
            } catch (err) {
                if (cancelled) return;

                if (err instanceof ApiError) {
                    setError(
                        err.status === 401
                            ? "Your Google sign-in session expired. Please try again."
                            : err.message,
                    );
                } else {
                    setError(
                        "Unable to complete Google sign-in. Please try again.",
                    );
                }
            }
        }

        void completeLogin();

        return () => {
            cancelled = true;
        };
    }, [code, router]);

    if (!code && !error) {
        return (
            <CallbackLayout
                title="Sign-in failed"
                description="The Google sign-in code is missing."
            >
                <p className="text-center text-sm text-destructive">
                    Please start the sign-in process again.
                </p>

                <Button
                    className="w-full"
                    onClick={() => router.replace("/login")}
                >
                    Return to login
                </Button>
            </CallbackLayout>
        );
    }

    return (
        <CallbackLayout
            title={error ? "Sign-in failed" : "Signing you in"}
            description={
                error
                    ? "Google authentication could not be completed."
                    : "Please wait while we complete your sign-in."
            }
        >
            {error ? (
                <>
                    <p
                        role="alert"
                        className="text-center text-sm text-destructive"
                    >
                        {error}
                    </p>

                    <Button
                        className="w-full"
                        onClick={() => router.replace("/login")}
                    >
                        Return to login
                    </Button>
                </>
            ) : (
                <LoaderCircle className="size-7 animate-spin text-primary" />
            )}
        </CallbackLayout>
    );
}

function CallbackLayout({
    title,
    description,
    children,
}: {
    title: string;
    description: string;
    children: React.ReactNode;
}) {
    return (
        <main className="flex min-h-screen items-center justify-center bg-background px-5">
            <Card className="w-full max-w-md">
                <CardHeader className="items-center text-center">
                    <div className="mb-3 flex size-12 items-center justify-center rounded-xl bg-primary text-primary-foreground">
                        <Zap className="size-6 fill-current" />
                    </div>

                    <CardTitle className="text-2xl">{title}</CardTitle>

                    <CardDescription>{description}</CardDescription>
                </CardHeader>

                <CardContent className="flex flex-col items-center gap-4">
                    {children}
                </CardContent>
            </Card>
        </main>
    );
}

export default function GoogleCallbackPage() {
    return (
        <Suspense
            fallback={
                <main className="flex min-h-screen items-center justify-center">
                    <LoaderCircle className="size-7 animate-spin" />
                </main>
            }
        >
            <GoogleCallbackContent />
        </Suspense>
    );
}
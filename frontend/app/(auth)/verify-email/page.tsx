"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, Loader2, MailCheck } from "lucide-react";

import { ApiError } from "@/lib/api/client";
import {
  resendVerification,
  verifyEmail,
} from "@/lib/api/auth";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const verifySchema = z.object({
  code: z
    .string()
    .trim()
    .length(6, "Verification code must be 6 digits")
    .regex(/^\d{6}$/, "Verification code must contain only digits"),
});

type VerifyFormData = z.infer<typeof verifySchema>;

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const email = searchParams.get("email") ?? "";

  const [verifyError, setVerifyError] = useState("");
  const [resendError, setResendError] = useState("");
  const [resendMessage, setResendMessage] = useState("");
  const [verified, setVerified] = useState(false);
  const [isResending, setIsResending] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<VerifyFormData>({
    resolver: zodResolver(verifySchema),
  });

  async function onSubmit(data: VerifyFormData) {
    setVerifyError("");

    if (!email) {
      setVerifyError(
        "Your email address is missing. Please register again.",
      );
      return;
    }

    try {
      await verifyEmail({
        email,
        code: data.code,
      });

      setVerified(true);
    } catch (error) {
      if (error instanceof ApiError) {
        setVerifyError(error.message);
      } else {
        setVerifyError("Unable to verify your email. Please try again.");
      }
    }
  }

  async function handleResend() {
    setResendError("");
    setResendMessage("");

    if (!email) {
      setResendError(
        "Your email address is missing. Please register again.",
      );
      return;
    }

    setIsResending(true);

    try {
      const response = await resendVerification({
        email,
      });

      setResendMessage(response.data.message);
    } catch (error) {
      if (error instanceof ApiError) {
        setResendError(error.message);
      } else {
        setResendError(
          "Unable to resend the verification email. Please try again.",
        );
      }
    } finally {
      setIsResending(false);
    }
  }

  if (verified) {
    return (
      <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-6 py-12">
        <div className="absolute inset-0 -z-10 bg-[linear-gradient(to_right,hsl(var(--border))_1px,transparent_1px),linear-gradient(to_bottom,hsl(var(--border))_1px,transparent_1px)] bg-[size:48px_48px] opacity-30" />

        <div className="absolute left-1/2 top-1/3 -z-10 h-80 w-80 -translate-x-1/2 rounded-full bg-primary/10 blur-3xl" />

        <Card className="w-full max-w-md border-border/60 shadow-xl">
          <CardHeader className="items-center text-center">
            <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <MailCheck className="h-7 w-7 text-primary" />
            </div>

            <CardTitle className="text-2xl">
              Email verified
            </CardTitle>

            <CardDescription>
              Your STAQ account has been successfully verified.
            </CardDescription>
          </CardHeader>

          <CardContent>
            <Button
              className="w-full"
              onClick={() => router.push("/login")}
            >
              Continue to login
            </Button>
          </CardContent>
        </Card>
      </main>
    );
  }

  return (
    <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-6 py-12">
      <div className="absolute inset-0 -z-10 bg-[linear-gradient(to_right,hsl(var(--border))_1px,transparent_1px),linear-gradient(to_bottom,hsl(var(--border))_1px,transparent_1px)] bg-[size:48px_48px] opacity-30" />

      <div className="absolute left-1/2 top-1/3 -z-10 h-80 w-80 -translate-x-1/2 rounded-full bg-primary/10 blur-3xl" />

      <Card className="w-full max-w-md border-border/60 shadow-xl">
        <CardHeader className="space-y-4 text-center">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
            <MailCheck className="h-7 w-7 text-primary" />
          </div>

          <div>
            <CardTitle className="text-2xl">
              Verify your email
            </CardTitle>

            <CardDescription className="mt-2">
              We sent a 6-digit verification code to
            </CardDescription>

            <p className="mt-1 break-all text-sm font-medium">
              {email || "your email address"}
            </p>
          </div>
        </CardHeader>

        <CardContent className="space-y-6">
          <form
            onSubmit={handleSubmit(onSubmit)}
            className="space-y-5"
          >
            <div className="space-y-2">
              <Label htmlFor="code">
                Verification code
              </Label>

              <Input
                id="code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                placeholder="123456"
                className="text-center text-lg tracking-[0.35em]"
                {...register("code")}
              />

              {errors.code && (
                <p className="text-sm text-destructive">
                  {errors.code.message}
                </p>
              )}
            </div>

            {verifyError && (
              <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {verifyError}
              </div>
            )}

            <Button
              type="submit"
              className="w-full"
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Verifying...
                </>
              ) : (
                "Verify email"
              )}
            </Button>
          </form>

          <div className="space-y-3 border-t pt-6 text-center">
            <p className="text-sm text-muted-foreground">
              Didn&apos;t receive the code?
            </p>

            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={handleResend}
              disabled={isResending}
            >
              {isResending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Sending...
                </>
              ) : (
                "Resend verification email"
              )}
            </Button>

            {resendMessage && (
              <p className="text-sm text-green-600">
                {resendMessage}
              </p>
            )}

            {resendError && (
              <p className="text-sm text-destructive">
                {resendError}
              </p>
            )}
          </div>

          <div className="text-center text-sm text-muted-foreground">
            <Link
              href="/login"
              className="font-medium text-foreground underline-offset-4 hover:underline"
            >
              Back to login
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import {
  ArrowLeft,
  Eye,
  EyeOff,
  UserPlus,
  Zap,
} from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

import { ApiError } from "@/lib/api/client";
import { registerUser } from "@/lib/auth/session";

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
import { ThemeToggle } from "@/components/theme-toggle";

const registerSchema = z.object({
  full_name: z
    .string()
    .trim()
    .min(2, "Full name must be at least 2 characters")
    .max(100, "Full name must not exceed 100 characters"),

  email: z
    .string()
    .trim()
    .email("Enter a valid email address")
    .max(255, "Email must not exceed 255 characters"),

  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(120, "Password must not exceed 120 characters"),
});

type RegisterFormData = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();

  const [serverError, setServerError] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
    },
  });

  async function onSubmit(data: RegisterFormData) {
    setServerError("");

    try {
      await registerUser(data);

      router.push(
        `/verify-email?email=${encodeURIComponent(data.email)}`,
      );
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.code === "EMAIL_EXISTS") {
          setServerError(
            "An account with this email already exists.",
          );
        } else {
          setServerError(error.message);
        }
      } else {
        setServerError(
          "Unable to create your account. Please try again.",
        );
      }
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-background">
      {/* Background decoration */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-32 -top-32 h-96 w-96 animate-pulse rounded-full bg-primary/10 blur-3xl" />

        <div
          className="absolute inset-0 opacity-[0.035]"
          style={{
            backgroundImage:
              "linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)",
            backgroundSize: "48px 48px",
          }}
        />

        <div className="absolute -bottom-32 -right-32 h-96 w-96 rounded-full bg-primary/10 blur-3xl" />
      </div>

      {/* Top navigation */}
      <div className="absolute left-0 right-0 top-0 z-20">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 sm:px-8 lg:px-10">
          <Link
            href="/"
            className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="size-4 transition-transform group-hover:-translate-x-1" />
            Back to STAQ
          </Link>

          <ThemeToggle />
        </div>
      </div>

      <div className="relative grid min-h-screen lg:grid-cols-2">
        {/* Left marketing panel */}
        <section className="hidden flex-col justify-between border-r bg-muted/30 p-10 pt-24 lg:flex">
          <div>
            <Link
              href="/"
              className="inline-flex items-center gap-2 text-xl font-bold tracking-tight"
            >
              <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg shadow-primary/20">
                <Zap className="size-5 fill-current" />
              </span>

              STAQ
            </Link>
          </div>

          <div className="max-w-lg animate-in fade-in slide-in-from-left-4 duration-700">
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border bg-background/70 px-3 py-1.5 text-sm text-muted-foreground shadow-sm backdrop-blur">
              <span className="size-2 animate-pulse rounded-full bg-emerald-500" />
              Smart task automation
            </div>

            <h1 className="text-5xl font-semibold tracking-tight xl:text-6xl">
              Build once.
              <br />

              <span className="text-muted-foreground">
                Automate repeatedly.
              </span>
            </h1>

            <p className="mt-6 max-w-md text-lg leading-8 text-muted-foreground">
              Create automated workflows that run on your schedule,
              execute ordered actions, and keep a complete execution
              history.
            </p>

            <div className="mt-10 space-y-3">
              <div className="flex items-center gap-3 rounded-xl border bg-background/60 p-4 shadow-sm backdrop-blur transition-transform duration-300 hover:-translate-y-1">
                <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <Zap className="size-4 text-primary" />
                </div>

                <div>
                  <p className="text-sm font-medium">
                    Scheduled automation
                  </p>

                  <p className="text-xs text-muted-foreground">
                    Daily, weekly, monthly, yearly and cron triggers.
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded-xl border bg-background/60 p-4 shadow-sm backdrop-blur transition-transform duration-300 hover:-translate-y-1">
                <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <UserPlus className="size-4 text-primary" />
                </div>

                <div>
                  <p className="text-sm font-medium">
                    Multiple actions
                  </p>

                  <p className="text-xs text-muted-foreground">
                    Chain HTTP, email, shell and reminder actions.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <p className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} STAQ
          </p>
        </section>

        {/* Register section */}
        <section className="flex items-center justify-center px-5 pb-12 pt-24 sm:px-8 lg:pt-16">
          <div className="w-full max-w-md animate-in fade-in slide-in-from-bottom-4 duration-700">
            {/* Mobile logo */}
            <div className="mb-8 flex justify-center lg:hidden">
              <Link
                href="/"
                className="inline-flex items-center gap-2 text-xl font-bold tracking-tight"
              >
                <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg shadow-primary/20">
                  <Zap className="size-5 fill-current" />
                </span>

                STAQ
              </Link>
            </div>

            <Card className="border-border/60 bg-card/95 shadow-xl shadow-black/5 backdrop-blur">
              <CardHeader className="space-y-4 pb-6">
                <div className="flex size-11 items-center justify-center rounded-xl border bg-muted/50">
                  <UserPlus className="size-5 text-muted-foreground" />
                </div>

                <div>
                  <CardTitle className="text-2xl">
                    Create your account
                  </CardTitle>

                  <CardDescription className="mt-2">
                    Start building automated workflows with STAQ.
                  </CardDescription>
                </div>
              </CardHeader>

              <CardContent>
                <form
                  onSubmit={handleSubmit(onSubmit)}
                  className="space-y-5"
                >
                  {/* Full name */}
                  <div className="space-y-2">
                    <Label htmlFor="full_name">
                      Full name
                    </Label>

                    <Input
                      id="full_name"
                      type="text"
                      placeholder="Ahsan Ahmed"
                      autoComplete="name"
                      className="h-11"
                      {...register("full_name")}
                    />

                    {errors.full_name && (
                      <p className="text-sm text-destructive">
                        {errors.full_name.message}
                      </p>
                    )}
                  </div>

                  {/* Email */}
                  <div className="space-y-2">
                    <Label htmlFor="email">
                      Email
                    </Label>

                    <Input
                      id="email"
                      type="email"
                      placeholder="you@example.com"
                      autoComplete="email"
                      className="h-11"
                      {...register("email")}
                    />

                    {errors.email && (
                      <p className="text-sm text-destructive">
                        {errors.email.message}
                      </p>
                    )}
                  </div>

                  {/* Password */}
                  <div className="space-y-2">
                    <Label htmlFor="password">
                      Password
                    </Label>

                    <div className="relative">
                      <Input
                        id="password"
                        type={showPassword ? "text" : "password"}
                        placeholder="At least 8 characters"
                        autoComplete="new-password"
                        className="h-11 pr-11"
                        {...register("password")}
                      />

                      <button
                        type="button"
                        onClick={() =>
                          setShowPassword((value) => !value)
                        }
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                        aria-label={
                          showPassword
                            ? "Hide password"
                            : "Show password"
                        }
                      >
                        {showPassword ? (
                          <EyeOff className="size-4" />
                        ) : (
                          <Eye className="size-4" />
                        )}
                      </button>
                    </div>

                    {errors.password && (
                      <p className="text-sm text-destructive">
                        {errors.password.message}
                      </p>
                    )}

                    <p className="text-xs text-muted-foreground">
                      Use between 8 and 120 characters.
                    </p>
                  </div>

                  {/* Server error */}
                  {serverError && (
                    <div
                      role="alert"
                      className="rounded-lg border border-destructive/20 bg-destructive/10 px-3 py-2.5 text-sm text-destructive"
                    >
                      {serverError}
                    </div>
                  )}

                  {/* Submit */}
                  <Button
                    type="submit"
                    className="h-11 w-full transition-transform active:scale-[0.98]"
                    disabled={isSubmitting}
                  >
                    {isSubmitting
                      ? "Creating account..."
                      : "Create account"}
                  </Button>
                </form>

                {/* Login */}
                <p className="mt-6 text-center text-sm text-muted-foreground">
                  Already have an account?{" "}
                  <Link
                    href="/login"
                    className="font-medium text-primary underline-offset-4 hover:underline"
                  >
                    Sign in
                  </Link>
                </p>
              </CardContent>
            </Card>

            <p className="mt-6 text-center text-xs text-muted-foreground">
              You&apos;ll need to verify your email before signing in.
            </p>
          </div>
        </section>
      </div>
    </main>
  );
}
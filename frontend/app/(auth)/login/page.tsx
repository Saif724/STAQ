"use client";

import Link from "next/link";
import {useRouter} from "next/navigation";
import {useState} from "react";
import {Eye, EyeOff, LockKeyhole, Zap} from "lucide-react";
import {useForm} from "react-hook-form";
import {z} from "zod";
import {zodResolver} from "@hookform/resolvers/zod";

import {loginUser} from "@/lib/auth/session";
import {ApiError} from "@/lib/api/client";

import {Button} from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";

const loginSchema = z.object({
    email: z
        .string()
        .email("Please enter a valid email address"),
    password: z
        .string()
        .min(1, "Password is required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
    const router = useRouter();

    const [serverError, setServerError] = useState("");
    const [showPassword, setShowPassword] = useState(false);

    const {
        register,
        handleSubmit,
        formState: {errors, isSubmitting},
    } = useForm<LoginFormData>({
        resolver: zodResolver(loginSchema),
        defaultValues: {
            email: "",
            password: "",
        },
    });

    async function onSubmit(data: LoginFormData) {
        setServerError("");

        try {
            await loginUser(data);
            router.push("/dashboard"); 
        } catch (error) {
            if (error instanceof ApiError) {
                setServerError(error.message);
            } else {
                setServerError("Unable to log in. Please try again.");
            }
        }
    }

    return (
        <main className="relative min-h-screen overflow-hidden bg-background">

            <div className="pointer-events-none absolute inset-0">
                <div className="absolute -left-32 -top-32 h-96 w-96 rounded-full bg-primary/10 blur-3xl"/>
                <div className="absolute -bottom-32 -right-32 h-96 w-96 rounded-full bg-primary/10 blur-3xl"/>

                <div
                    className="absolute inset-0 opacity-[0.035]"
                    style={{
                        backgroundImage: "linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)",
                        backgroundSize: "48px 48px",
                    }}
                />
            </div>

            <div className="relative grid min-h-screen lg:grid-cols-2">
                <section className="hidden flex-col justify-between border-r bg-muted/30 p-10 lg:flex">
                    <div>
                        <Link
                            href="/"
                            className="inline-flex items-center gap-2 text-xl font-bold tracking-tight"
                        >
                            <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground">
                                <Zap className="size-5 fill-current"/>
                            </span>

                            STAQ
                        </Link>
                    </div>

                    <div className="max-w-lg">
                        <div className="mb-6 inline-flex items-center gap-2 rounded-full border bg-background/70 px-3 py-1.5 text-sm text-muted-foreground backdrop-blur">
                            <span className="size-2 rounded-full bg-emerald-500"/>
                            Smart task automation
                        </div>

                        <h1 className="text-5xl font-semibold tracking-tight xl:text-6xl">
                            Automate the work.
                            <br />
                            <span className="text-muted-foreground">
                                Focus on what matters.
                            </span>
                        </h1>

                        <p className="mt-6 max-w-md text-lg leading-8 text-muted-foreground">
                            Schedule tasks, trigger actions, and let STAQ handle repetitive
                            work automatically.
                        </p>

                        <div className="mt-10 grid grid-cols-3 gap-3">
                            <div className="rounded-xl border bg-background/60 p-4 backdrop-blur">
                                <p className="text-2xl font-semibold">24/7</p>
                                <p className="mt-1 text-xs text-muted-foreground">
                                    Automation
                                </p>
                            </div>

                            <div className="rounded-xl border bg-background/60 p-4 backdrop-blur-2xl">
                                <p className="text-2xl font-semibold">Multi</p>
                                <p className="mt-1 text-xs text-muted-foreground">
                                    Action workflows
                                </p>
                            </div>

                            <div className="rounded-xl border bg-background/60 p-4 backdrop-blur-2xl">
                                <p className="text-2xl font-semibold">Live</p>
                                <p className="mt-1 text-xs text-muted-foreground">
                                    Execution tracking
                                </p>
                            </div>
                        </div>
                    </div>

                    <p className="text-sm text-muted-foreground">
                        © {new Date().getFullYear()} STAQ
                    </p>
                </section>

                <section className="flex items-center justify-center px-5 py-12 sm:px-8">
                    <div className="w-full max-w-md">
                        <div className="mb-10 flex justify-center lg:hidden">
                            <Link
                                href="/"
                                className="inline-flex items-center gap-2 text-xl font-bold tracking-tight"
                            >
                                <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground">
                                    <Zap className="size-5 fill-current" />
                                </span>

                                STAQ
                            </Link>
                        </div>

                         <Card className="border-border/60 bg-card/95 shadow-xl shadow-black/5 backdrop-blur">
                            <CardHeader className="space-y-3 pb-6">
                                <div className="flex size-11 items-center justify-center rounded-xl border bg-muted/50">
                                    <LockKeyhole className="size-5 text-muted-foreground" />
                                </div>
                                <div>
                                    <CardTitle className="text-2xl">
                                        Welcome back
                                    </CardTitle>
                                    <CardDescription className="mt-2">
                                        Sign in to continue to your STAQ workspace.
                                    </CardDescription>
                                </div>
                            </CardHeader>

                            <CardContent>
                                <form
                                    onSubmit={handleSubmit(onSubmit)}
                                    className="space-y-5"
                                >
                                    <div className="space-y-2">
                                        <Label htmlFor="email">Email</Label>

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

                                    <div className="space-y-2">
                                        <div className="flex items-center justify-between">
                                            <Label htmlFor="password">Password</Label>

                                            <span className="text-xs text-muted-foreground">
                                                Forgot password?
                                            </span>
                                        </div>
                                        
                                        <div className="relative">
                                            <Input
                                                id="password"
                                                type={showPassword ? "text" : "password"}
                                                placeholder="Enter your password"
                                                autoComplete="current-password"
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

                                    </div>

                                    {serverError && (
                                        <div
                                            role="alert"
                                            className="rounded-lg border border-destructive/20 bg-destructive/10 px-3 py-2.5 text-sm text-destructive"
                                        >
                                            {serverError}
                                        </div>
                                    )}

                                    <Button
                                        type="submit"
                                        className="h-11 w-full"
                                        disabled={isSubmitting}
                                    >
                                        {isSubmitting ? "Signing in..." : "Sign in"}
                                    </Button>
                                </form>

                                <div className="my-6 flex items-center gap-3">
                                    <div className="h-px flex-1 bg-border" />
                                    <span className="text-xs text-muted-foreground">
                                        OR
                                    </span>
                                    <div className="h-px flex-1 bg-border" />
                                </div>

                                <Button
                                    type="button"
                                    variant="outline"
                                    className="h-11 w-full"
                                    disabled
                                >
                                    Continue with Google
                                </Button>

                                <p className="mt-6 text-center text-sm text-muted-foreground">
                                    Don&apos;t have an account?{" "}
                                    <Link
                                        href="/register"
                                        className="font-medium text-foreground underline-offset-4 hover:underline"
                                    >
                                        Create one
                                    </Link>
                                </p>
                            </CardContent>
                        </Card>

                        <p className="mt-6 text-center text-xs text-muted-foreground">
                            By continuing, you agree to STAQ&apos;s terms and
                            privacy policy.
                        </p>
                    </div>
                </section>
            </div>
           
        </main>
    );
}
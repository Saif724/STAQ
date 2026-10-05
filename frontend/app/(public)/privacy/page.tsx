import Link from "next/link";

export const metadata = {
  title: "Privacy Policy",
  description: "Privacy Policy for STAQ.",
};

export default function PrivacyPage() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <div className="mx-auto max-w-3xl px-6 py-16 lg:px-8 lg:py-24">
        <div className="mb-12">
          <Link
            href="/"
            className="text-sm font-semibold tracking-tight hover:opacity-70"
          >
            ← Back to STAQ
          </Link>

          <h1 className="mt-8 text-4xl font-bold tracking-tight">
            Privacy Policy
          </h1>

          <p className="mt-3 text-sm text-muted-foreground">
            Last updated: October 5, 2026
          </p>
        </div>

        <div className="space-y-10 text-sm leading-7 text-muted-foreground">
          <section>
            <h2 className="text-xl font-semibold text-foreground">
              1. Introduction
            </h2>

            <p className="mt-3">
              STAQ is an automation platform that allows users to create,
              schedule, and execute automated workflows. This Privacy Policy
              explains what information STAQ collects, how that information is
              used, and how it is protected.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              2. Information We Collect
            </h2>

            <p className="mt-3">
              When you create an account, we may collect information such as
              your name, email address, and authentication credentials.
            </p>

            <p className="mt-3">
              STAQ also stores information required to operate your workflows,
              including tasks, triggers, actions, execution records, and
              related configuration data.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              3. Google and Connected Services
            </h2>

            <p className="mt-3">
              If you connect a Google account or another supported service,
              STAQ may receive information necessary to provide the requested
              integration.
            </p>

            <p className="mt-3">
              We use connected-service data only to provide the functionality
              requested by the user. We do not sell users&apos; personal
              information or connected-service data.
            </p>

            <p className="mt-3">
              Access to connected services can be revoked by the user through
              the relevant service or through STAQ when supported.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              4. How We Use Information
            </h2>

            <p className="mt-3">
              Information is used to provide, maintain, secure, and improve
              STAQ and to execute the workflows configured by users.
            </p>

            <p className="mt-3">
              We may also use information to authenticate users, prevent abuse,
              troubleshoot failures, maintain system reliability, and
              communicate important service-related information.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              5. Data Security
            </h2>

            <p className="mt-3">
              STAQ uses reasonable technical and organizational measures to
              protect user information against unauthorized access,
              alteration, disclosure, or destruction.
            </p>

            <p className="mt-3">
              However, no internet-based service can guarantee absolute
              security.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              6. Data Retention
            </h2>

            <p className="mt-3">
              We retain account, workflow, and execution information for as
              long as necessary to provide the service and fulfill legitimate
              operational requirements.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              7. Your Choices
            </h2>

            <p className="mt-3">
              You may stop using STAQ at any time. Where supported, you may
              disconnect connected services and remove information associated
              with your account.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              8. Changes to This Policy
            </h2>

            <p className="mt-3">
              We may update this Privacy Policy from time to time. Changes will
              be reflected on this page with an updated revision date.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold text-foreground">
              9. Contact
            </h2>

            <p className="mt-3">
              If you have questions about this Privacy Policy or STAQ&apos;s
              handling of information, please contact the STAQ service
              administrator.
            </p>
          </section>
        </div>

        <div className="mt-16 border-t border-border pt-6">
          <Link
            href="/"
            className="text-sm font-medium text-foreground hover:opacity-70"
          >
            ← Return to STAQ
          </Link>
        </div>
      </div>
    </main>
  );
}

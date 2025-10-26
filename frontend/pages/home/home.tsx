import * as preact from "preact";
import * as vlens from "vlens";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./home-styles";

type Data = {};

type CodeEntryForm = {
  code: string;
  error: string;
  loading: boolean;
  autoValidated: boolean; // Track if we've already auto-validated
  rateLimitedUntil: Date | null; // When rate limit expires
  retryCountdown: number; // Seconds until can retry
};

const useCodeForm = vlens.declareHook(
  (routeKey: string): CodeEntryForm => ({
    code: "",
    error: "",
    loading: false,
    autoValidated: false,
    rateLimitedUntil: null,
    retryCountdown: 0,
  }),
);

export async function fetch(route: string, prefix: string) {
  return rpc.ok<Data>({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Extract code from URL if present (e.g., /12345)
  const urlCode = extractCodeFromRoute(route, prefix);

  // Use route as cache key to maintain state across renders
  const form = useCodeForm(route);

  // Auto-validate if code is in URL and we haven't validated yet
  if (urlCode && !form.autoValidated && !form.loading) {
    form.code = urlCode;
    form.autoValidated = true;
    // Trigger validation asynchronously
    setTimeout(() => autoValidateCode(form, urlCode), 0);
  }

  // Update countdown every second if rate limited
  if (form.rateLimitedUntil) {
    const now = Date.now();
    const targetTime = form.rateLimitedUntil.getTime();
    const remainingSeconds = Math.ceil((targetTime - now) / 1000);

    if (remainingSeconds > 0) {
      form.retryCountdown = remainingSeconds;
      // Schedule next update in 1 second
      setTimeout(() => vlens.scheduleRedraw(), 1000);
    } else {
      // Rate limit expired
      form.rateLimitedUntil = null;
      form.retryCountdown = 0;
      vlens.scheduleRedraw();
    }
  }

  const isRateLimited =
    form.rateLimitedUntil !== null && form.retryCountdown > 0;

  return (
    <div>
      <Header />
      <main className="landing-container">
        <div className="landing-card">
          <h1 className="landing-title">Join Stream</h1>
          <p className="landing-description">
            Enter the 5-digit access code to join the stream.
          </p>

          <form
            onSubmit={(e) => onSubmitCode(form, e)}
            className="landing-form"
          >
            <div className="landing-input-group">
              <input
                type="text"
                className="landing-input"
                placeholder="Enter code"
                value={form.code}
                maxLength={5}
                pattern="[0-9]*"
                inputMode="numeric"
                disabled={form.loading || isRateLimited}
                onInput={(e) => {
                  const target = e.target as HTMLInputElement;
                  // Only allow digits
                  form.code = target.value.replace(/[^0-9]/g, "");
                  form.error = "";
                  vlens.scheduleRedraw();
                }}
                autoFocus
              />
            </div>

            {form.error && <div className="landing-error">{form.error}</div>}

            {isRateLimited && (
              <div className="landing-rate-limit">
                Too many failed attempts. Try again in {form.retryCountdown}{" "}
                second{form.retryCountdown !== 1 ? "s" : ""}.
              </div>
            )}

            <button
              type="submit"
              className="landing-submit-btn"
              disabled={form.loading || form.code.length !== 5 || isRateLimited}
            >
              {form.loading
                ? "Validating..."
                : isRateLimited
                  ? `Locked (${form.retryCountdown}s)`
                  : "Join Stream"}
            </button>
          </form>

          <div className="landing-help">
            <p>Don't have an access code? Contact the stream owner.</p>
          </div>

          <div className="landing-auth-links">
            <span>Have an account? </span>
            <a href="/login">Sign in</a>
            <span> | </span>
            <a href="/create-account">Create account</a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

function extractCodeFromRoute(route: string, prefix: string): string {
  // Extract code from URL like /12345
  const afterPrefix = route.substring(prefix.length);
  const parts = afterPrefix.split("/").filter((p) => p.length > 0);

  if (parts.length > 0) {
    const code = parts[0];
    // Validate it's a 5-digit code
    if (/^\d{5}$/.test(code)) {
      return code;
    }
  }

  return "";
}

async function autoValidateCode(form: CodeEntryForm, code: string) {
  form.loading = true;
  form.error = "";
  vlens.scheduleRedraw();

  await validateCode(form, code);
}

async function validateCode(form: CodeEntryForm, code: string) {
  const nativeFetch = window.fetch.bind(window);
  try {
    const res = await nativeFetch("/api/validate-access-code", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        code: code,
      }),
    });

    const data = await res.json();

    if (data.redirectTo) {
      // Successfully validated - redirect to stream
      core.setRoute(data.redirectTo);
    } else if (data.rateLimited) {
      // Rate limited - set lockout timer
      const retryAfterSeconds = data.retryAfterSeconds || 60;
      form.rateLimitedUntil = new Date(Date.now() + retryAfterSeconds * 1000);
      form.retryCountdown = retryAfterSeconds;
      form.error = ""; // Clear error, show rate limit message instead
      form.loading = false;
      vlens.scheduleRedraw();
    } else {
      // Validation failed - show error
      form.error = data.error || "Invalid access code";
      form.rateLimitedUntil = null;
      form.retryCountdown = 0;
      form.loading = false;
      vlens.scheduleRedraw();
    }
  } catch (err) {
    form.error = "Failed to validate code. Please try again.";
    form.loading = false;
    vlens.scheduleRedraw();
  }
}

async function onSubmitCode(form: CodeEntryForm, event: Event) {
  event.preventDefault();

  if (form.code.length !== 5) {
    form.error = "Please enter a 5-digit code";
    vlens.scheduleRedraw();
    return;
  }

  form.loading = true;
  form.error = "";
  vlens.scheduleRedraw();

  await validateCode(form, form.code);
}

import * as preact from "preact";
import * as vlens from "vlens";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./watch-styles";

type Data = {};

type CodeEntryForm = {
  code: string;
  error: string;
  loading: boolean;
  autoValidated: boolean; // Track if we've already auto-validated
};

const useCodeForm = vlens.declareHook(
  (routeKey: string): CodeEntryForm => ({
    code: "",
    error: "",
    loading: false,
    autoValidated: false,
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
  // Extract code from URL if present (e.g., /watch/12345)
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

  return (
    <div>
      <Header />
      <main className="watch-container">
        <div className="watch-card">
          <h1 className="watch-title">Join Stream</h1>
          <p className="watch-description">
            Enter the 5-digit access code to join the stream.
          </p>

          <form onSubmit={(e) => onSubmitCode(form, e)} className="watch-form">
            <div className="watch-input-group">
              <input
                type="text"
                className="watch-input"
                placeholder="Enter code"
                value={form.code}
                maxLength={5}
                pattern="[0-9]*"
                inputMode="numeric"
                disabled={form.loading}
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

            {form.error && <div className="watch-error">{form.error}</div>}

            <button
              type="submit"
              className="watch-submit-btn"
              disabled={form.loading || form.code.length !== 5}
            >
              {form.loading ? "Validating..." : "Join Stream"}
            </button>
          </form>

          <div className="watch-help">
            <p>Don't have an access code? Contact the stream owner.</p>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

function extractCodeFromRoute(route: string, prefix: string): string {
  // Extract code from URL like /watch/12345
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

    if (data.success) {
      // Successfully validated - redirect to stream
      core.setRoute(data.redirectTo);
    } else {
      // Validation failed - show error
      form.error = data.error || "Invalid access code";
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

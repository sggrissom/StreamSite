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
};

const useCodeForm = vlens.declareHook(
  (): CodeEntryForm => ({
    code: "",
    error: "",
    loading: false,
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
  const form = useCodeForm();

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

  const nativeFetch = window.fetch.bind(window);
  try {
    const res = await nativeFetch("/api/validate-access-code", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        code: form.code,
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

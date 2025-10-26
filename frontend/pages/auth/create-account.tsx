import * as preact from "preact";
import * as vlens from "vlens";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./create-account-styles";

type Data = {
  authId: number;
};

type CreateAccountForm = {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
  error: string;
  loading: boolean;
};

const useCreateAccountForm = vlens.declareHook(
  (): CreateAccountForm => ({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
    error: "",
    loading: false,
  }),
);

export async function fetch(route: string, prefix: string) {
  // Check if user is already authenticated
  let [authResp, authErr] = await server.GetAuthContext({});

  return rpc.ok<Data>({
    authId: authResp?.id || 0,
  });
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Redirect to dashboard if already authenticated
  if (data.authId > 0) {
    core.setRoute("/dashboard");
    return <div></div>;
  }

  const form = useCreateAccountForm();

  return (
    <div>
      <Header />
      <main className="create-account-container">
        <CreateAccountPage form={form} />
      </main>
      <Footer />
    </div>
  );
}

async function onCreateAccountClicked(form: CreateAccountForm, event: Event) {
  event.preventDefault();
  form.loading = true;
  form.error = "";
  vlens.scheduleRedraw();

  // Frontend validation
  if (form.password !== form.confirmPassword) {
    form.loading = false;
    form.error = "Passwords do not match";
    vlens.scheduleRedraw();
    return;
  }

  if (form.password.length < 8) {
    form.loading = false;
    form.error = "Password must be at least 8 characters long";
    vlens.scheduleRedraw();
    return;
  }

  // Call backend CreateAccount procedure
  let [resp, err] = await server.CreateAccount({
    name: form.name,
    email: form.email,
    password: form.password,
    confirmPassword: form.confirmPassword,
  });

  form.loading = false;

  if (err) {
    form.error = err || "Failed to create account";
  } else {
    // Clear form
    form.name = "";
    form.email = "";
    form.password = "";
    form.confirmPassword = "";
    form.error = "";
    // Redirect to dashboard
    window.location.href = "/dashboard";
  }
  vlens.scheduleRedraw();

  // Scroll to error message if there's an error
  if (form.error) {
    setTimeout(() => {
      const errorElement = document.querySelector(".error-message");
      if (errorElement) {
        errorElement.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 100);
  }
}

function onGoogleSignup() {
  // Redirect to Google OAuth endpoint (same as login)
  window.location.href = "/api/login/google";
}

interface CreateAccountPageProps {
  form: CreateAccountForm;
}

const CreateAccountPage = ({ form }: CreateAccountPageProps) => (
  <div className="create-account-page">
    <div className="auth-card">
      <div className="auth-header">
        <h1>Create Account</h1>
        <p>Get started with your stream account</p>
      </div>

      {form.error && <div className="error-message">{form.error}</div>}

      <div className="auth-methods">
        <button
          className="btn btn-google"
          disabled={form.loading}
          onClick={onGoogleSignup}
        >
          <GoogleIcon />
          Continue with Google
        </button>

        <div className="auth-divider">
          <span>or</span>
        </div>

        <form
          className="auth-form"
          onSubmit={vlens.cachePartial(onCreateAccountClicked, form)}
        >
          <div className="form-group">
            <label htmlFor="name">Full Name</label>
            <input
              type="text"
              id="name"
              placeholder="Enter your full name"
              {...vlens.attrsBindInput(vlens.ref(form, "name"))}
              required
              disabled={form.loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="email">Email Address</label>
            <input
              type="email"
              id="email"
              placeholder="Enter your email"
              {...vlens.attrsBindInput(vlens.ref(form, "email"))}
              required
              disabled={form.loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              placeholder="Create a password"
              {...vlens.attrsBindInput(vlens.ref(form, "password"))}
              required
              disabled={form.loading}
            />
            <small className="form-hint">Must be at least 8 characters</small>
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm Password</label>
            <input
              type="password"
              id="confirmPassword"
              placeholder="Confirm your password"
              {...vlens.attrsBindInput(vlens.ref(form, "confirmPassword"))}
              required
              disabled={form.loading}
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-large auth-submit"
            disabled={form.loading}
          >
            {form.loading ? "Creating..." : "Create Account"}
          </button>
        </form>
      </div>

      <div className="auth-footer">
        <p>
          Already have an account?
          <a href="/login" className="auth-link">
            Sign in
          </a>
        </p>
      </div>
    </div>
  </div>
);

const GoogleIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
    <path
      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      fill="#4285F4"
    />
    <path
      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      fill="#34A853"
    />
    <path
      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      fill="#FBBC05"
    />
    <path
      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      fill="#EA4335"
    />
  </svg>
);

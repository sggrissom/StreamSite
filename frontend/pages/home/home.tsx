import * as preact from "preact";
import * as vlens from "vlens";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./home-styles";

type Data = {
  authId: number;
};

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
  return (
    <div>
      <Header />
      <main className="landing-container">
        <div className="landing-content">
          <h1 className="landing-title">Welcome to Stream</h1>
          <p className="landing-description">
            Sign in to access your stream or create a new account to get started.
          </p>
          <div className="landing-actions">
            <a href="/login" className="btn btn-primary btn-large">
              Sign In
            </a>
            <a href="/create-account" className="btn btn-large">
              Create Account
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

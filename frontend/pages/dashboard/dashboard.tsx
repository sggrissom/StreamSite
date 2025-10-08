import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./dashboard-styles";

type Data = {
  authId: number;
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated
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
  // Redirect to site root if not authenticated
  if (data.authId === 0) {
    core.setRoute("/");
    return <div></div>;
  }

  return (
    <div>
      <Header />
      <main className="dashboard-container">
        <div className="dashboard-content">
          <h1 className="dashboard-title">Welcome to Stream</h1>
          <p className="dashboard-description">
            You're all set! Ready to watch the stream?
          </p>
          <div className="dashboard-actions">
            <a href="/stream" className="btn btn-primary btn-large">
              Watch Live Stream
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

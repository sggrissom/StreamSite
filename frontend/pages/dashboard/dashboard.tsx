import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./dashboard-styles";

type Data = {
  authId: number;
  streamStatus: server.GetStreamStatusResponse | null;
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated
  let [authResp, authErr] = await server.GetAuthContext({});

  // Check stream status
  let [statusResp, statusErr] = await server.GetStreamStatus({});

  return rpc.ok<Data>({
    authId: authResp?.id || 0,
    streamStatus: statusResp || null,
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

  const isLive = data.streamStatus?.isLive || false;

  return (
    <div>
      <Header />
      <main className="dashboard-container">
        <div className="dashboard-content">
          <h1 className="dashboard-title">Welcome to Stream</h1>
          <p className="dashboard-description">
            You're all set! Ready to watch the stream?
          </p>

          {/* Stream status badge */}
          {data.streamStatus && (
            <div className="stream-status-badge">
              <span
                className={`stream-status-dot ${isLive ? "status-live" : "status-offline"}`}
              ></span>
              <span
                className={`stream-status-text ${isLive ? "status-live" : "status-offline"}`}
              >
                {isLive ? "LIVE NOW" : "OFFLINE"}
              </span>
            </div>
          )}

          <div className="dashboard-actions">
            {isLive ? (
              <a href="/stream" className="btn btn-primary btn-large">
                Watch Live Stream
              </a>
            ) : (
              <button className="btn btn-primary btn-large" disabled>
                Stream Offline
              </button>
            )}
            <a href="/studios" className="btn btn-secondary btn-large">
              My Studios
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

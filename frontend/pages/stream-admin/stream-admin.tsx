import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./stream-admin-styles";

type Data = {
  auth: server.AuthResponse | null;
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated and has stream admin role
  let [authResp, authErr] = await server.GetAuthContext({});

  return rpc.ok<Data>({
    auth: authResp || null,
  });
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Redirect to login if not authenticated
  if (!data.auth || data.auth.id === 0) {
    core.setRoute("/login");
    return <div></div>;
  }

  // Redirect to regular dashboard if not stream admin
  if (!data.auth.isStreamAdmin) {
    core.setRoute("/dashboard");
    return <div></div>;
  }

  return (
    <div>
      <Header />
      <main className="stream-admin-container">
        <div className="stream-admin-content">
          <h1 className="stream-admin-title">Stream Admin Dashboard</h1>
          <p className="stream-admin-description">
            Welcome, {data.auth.name}! You have Stream Admin access.
          </p>

          <div className="admin-sections">
            <div className="admin-section">
              <h2>Stream Management</h2>
              <p>Manage live streams, schedules, and broadcasting settings.</p>
              <div className="section-actions">
                <button className="btn btn-primary">Manage Streams</button>
              </div>
            </div>

            <div className="admin-section">
              <h2>Class Schedule</h2>
              <p>
                Create and manage class schedules and instructor assignments.
              </p>
              <div className="section-actions">
                <button className="btn btn-primary">Manage Schedule</button>
              </div>
            </div>

            <div className="admin-section">
              <h2>Content Library</h2>
              <p>Manage recorded sessions and on-demand content.</p>
              <div className="section-actions">
                <button className="btn btn-primary">Manage Content</button>
              </div>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

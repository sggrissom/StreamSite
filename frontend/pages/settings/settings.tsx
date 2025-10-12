import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./settings-styles";

type Data = {
  auth: server.AuthResponse | null;
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated and has site admin role
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

  // Redirect to dashboard if not site admin
  if (!data.auth.isSiteAdmin) {
    core.setRoute("/dashboard");
    return <div></div>;
  }

  return (
    <div>
      <Header />
      <main className="settings-container">
        <div className="settings-content">
          <div className="settings-header">
            <h1 className="settings-title">Site Settings</h1>
            <p className="settings-description">
              Configure site-wide settings and preferences.
            </p>
          </div>

          <div className="settings-sections">
            <div className="settings-section">
              <h2>Site Configuration</h2>
              <p>General site settings and configuration options.</p>
              <div className="settings-placeholder">
                <span className="placeholder-text">
                  Settings will be added here
                </span>
              </div>
            </div>

            <div className="settings-section">
              <h2>User Defaults</h2>
              <p>Default settings for new user accounts.</p>
              <div className="settings-placeholder">
                <span className="placeholder-text">
                  Settings will be added here
                </span>
              </div>
            </div>

            <div className="settings-section">
              <h2>Security Settings</h2>
              <p>Authentication and security configuration.</p>
              <div className="settings-placeholder">
                <span className="placeholder-text">
                  Settings will be added here
                </span>
              </div>
            </div>

            <div className="settings-section">
              <h2>System Information</h2>
              <p>View system status and version information.</p>
              <div className="settings-placeholder">
                <span className="placeholder-text">
                  Settings will be added here
                </span>
              </div>
            </div>
          </div>

          <div className="settings-actions">
            <a href="/site-admin" className="btn btn-secondary">
              Back to Admin Dashboard
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

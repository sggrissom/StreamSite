import * as preact from "preact";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./stream-admin-styles";

type Data = server.ListMyStudiosResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListMyStudios({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const studios = data?.studios || [];

  return (
    <div>
      <Header />
      <main className="stream-admin-container">
        <div className="stream-admin-content">
          <h1 className="stream-admin-title">Stream Admin Dashboard</h1>
          <p className="stream-admin-description">
            Manage your studios and streaming setup from this central hub.
          </p>

          <div className="admin-overview">
            <div className="overview-card">
              <div className="overview-icon">📺</div>
              <div className="overview-content">
                <h3>Your Studios</h3>
                <p className="overview-stat">{studios.length}</p>
                <p className="overview-label">
                  {studios.length === 1 ? "Studio" : "Studios"}
                </p>
              </div>
            </div>

            <div className="overview-card">
              <div className="overview-icon">🎬</div>
              <div className="overview-content">
                <h3>Quick Actions</h3>
                <div className="quick-actions">
                  <a href="/studios" className="btn btn-primary btn-sm">
                    Manage All Studios
                  </a>
                  <a href="/stream" className="btn btn-secondary btn-sm">
                    View Live Stream
                  </a>
                </div>
              </div>
            </div>
          </div>

          <div className="studios-section">
            <div className="section-header">
              <h2>Your Studios</h2>
              <a href="/studios" className="btn btn-primary">
                Manage All Studios
              </a>
            </div>

            {studios.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">📺</div>
                <h3>No Studios Yet</h3>
                <p>
                  Create your first studio to start managing streaming rooms and
                  members.
                </p>
                <a href="/studios" className="btn btn-primary">
                  Create Your First Studio
                </a>
              </div>
            ) : (
              <div className="studios-list">
                {studios.map((studio) => (
                  <div key={studio.id} className="studio-item">
                    <div className="studio-item-header">
                      <div>
                        <h3 className="studio-item-name">{studio.name}</h3>
                        {studio.description && (
                          <p className="studio-item-description">
                            {studio.description}
                          </p>
                        )}
                      </div>
                      <span className={`studio-role role-${studio.myRole}`}>
                        {studio.myRoleName}
                      </span>
                    </div>

                    <div className="studio-item-footer">
                      <div className="studio-item-meta">
                        <span className="meta-badge">
                          Max Rooms: {studio.maxRooms}
                        </span>
                      </div>
                      <a
                        href={`/studio/${studio.id}`}
                        className="btn btn-primary btn-sm"
                      >
                        Open Studio →
                      </a>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="admin-help">
            <h3>Stream Admin Guide</h3>
            <div className="help-grid">
              <div className="help-card">
                <h4>🏢 Studios</h4>
                <p>
                  Organizational units that contain multiple streaming rooms.
                  Each studio has its own members and permissions.
                </p>
              </div>
              <div className="help-card">
                <h4>🚪 Rooms</h4>
                <p>
                  Individual streaming endpoints within a studio. Each room has
                  its own stream key for OBS/streaming software.
                </p>
              </div>
              <div className="help-card">
                <h4>👥 Members</h4>
                <p>
                  Add users to your studio with different roles: Viewer, Member,
                  Admin, or Owner with varying permission levels.
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

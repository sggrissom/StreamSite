import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./studio-styles";

type Data = server.GetStudioDashboardResponse;

export async function fetch(route: string, prefix: string) {
  // Extract studio ID from route (e.g., "/studio/123" -> "123")
  const studioIdStr = route
    .replace(prefix, "")
    .replace(/^\//, "")
    .split("/")[0];
  const studioId = parseInt(studioIdStr, 10);

  return server.GetStudioDashboard({ studioId });
}

function getRoleBadgeClass(role: number): string {
  return `studio-role role-${role}`;
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Handle errors or missing data
  if (!data || !data.success) {
    return (
      <div>
        <Header />
        <main className="studio-container">
          <div className="studio-content">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Studio Not Found</h2>
              <p>
                {data?.error ||
                  "The studio you're looking for doesn't exist or you don't have permission to view it."}
              </p>
              <a href="/studios" className="btn btn-primary">
                Back to Studios
              </a>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const studio = data.studio;
  const rooms = data.rooms || [];
  const myRole = data.myRole;
  const myRoleName = data.myRoleName;

  // Check if user can manage rooms (Admin or Owner)
  const canManageRooms = myRole >= 2; // Admin or Owner

  return (
    <div>
      <Header />
      <main className="studio-container">
        <div className="studio-content">
          {/* Breadcrumb */}
          <div className="breadcrumb">
            <a href="/studios">Studios</a>
            <span className="breadcrumb-separator">/</span>
            <span className="breadcrumb-current">{studio.name}</span>
          </div>

          {/* Studio Header */}
          <div className="studio-header">
            <div className="studio-header-main">
              <h1 className="studio-title">{studio.name}</h1>
              <span className={getRoleBadgeClass(myRole)}>{myRoleName}</span>
            </div>
            {studio.description && (
              <p className="studio-description">{studio.description}</p>
            )}
          </div>

          {/* Studio Metadata */}
          <div className="studio-metadata">
            <div className="metadata-card">
              <div className="metadata-label">Max Rooms</div>
              <div className="metadata-value">{studio.maxRooms}</div>
            </div>
            <div className="metadata-card">
              <div className="metadata-label">Total Rooms</div>
              <div className="metadata-value">{rooms.length}</div>
            </div>
            <div className="metadata-card">
              <div className="metadata-label">Active Rooms</div>
              <div className="metadata-value">
                {rooms.filter((r) => r.isActive).length}
              </div>
            </div>
          </div>

          {/* Action Buttons Placeholder */}
          {canManageRooms && (
            <div className="studio-actions">
              <button className="btn btn-secondary" disabled>
                Edit Studio
              </button>
              <button className="btn btn-secondary" disabled>
                Manage Members
              </button>
            </div>
          )}

          {/* Rooms Section */}
          <div className="rooms-section">
            <div className="rooms-header">
              <h2 className="section-title">Rooms</h2>
              {canManageRooms && rooms.length < studio.maxRooms && (
                <button className="btn btn-primary btn-sm" disabled>
                  Create Room
                </button>
              )}
            </div>

            {rooms.length === 0 ? (
              <div className="rooms-empty">
                <div className="empty-icon">🎬</div>
                <h3>No Rooms Yet</h3>
                <p>
                  {canManageRooms
                    ? "Create your first room to start streaming."
                    : "This studio doesn't have any rooms yet."}
                </p>
                {canManageRooms && (
                  <button className="btn btn-primary" disabled>
                    Create First Room
                  </button>
                )}
              </div>
            ) : (
              <div className="rooms-grid">
                {rooms.map((room) => (
                  <div key={room.id} className="room-card">
                    <div className="room-header">
                      <div className="room-number">Room {room.roomNumber}</div>
                      {room.isActive && (
                        <span className="room-status active">🔴 Live</span>
                      )}
                    </div>

                    <h3 className="room-name">{room.name}</h3>

                    <div className="room-meta">
                      <span className="meta-item">
                        <span className="meta-label">Created:</span>
                        <span className="meta-value">
                          {new Date(room.creation).toLocaleDateString()}
                        </span>
                      </span>
                    </div>

                    <div className="room-actions">
                      {canManageRooms && (
                        <>
                          <button className="btn btn-secondary btn-sm" disabled>
                            View Stream Key
                          </button>
                          <button className="btn btn-secondary btn-sm" disabled>
                            Edit
                          </button>
                        </>
                      )}
                      {!canManageRooms && room.isActive && (
                        <button className="btn btn-primary btn-sm" disabled>
                          Watch Stream
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {canManageRooms && rooms.length >= studio.maxRooms && (
              <div className="rooms-limit-notice">
                <p>
                  ⚠️ You've reached the maximum number of rooms (
                  {studio.maxRooms}
                  ). To create more rooms, increase the limit in studio
                  settings.
                </p>
              </div>
            )}
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./dashboard-styles";

type Data = {
  authId: number;
  rooms: server.ListMyAccessibleRoomsResponse | null;
  auth: server.AuthResponse | null;
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated
  let [authResp, authErr] = await server.GetAuthContext({});

  // Get accessible rooms
  let [roomsResp, roomsErr] = await server.ListMyAccessibleRooms({});

  return rpc.ok<Data>({
    authId: authResp?.id || 0,
    rooms: roomsResp || null,
    auth: authResp || null,
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

  const auth = data.auth;
  const canManageStudios = auth?.canManageStudios || false;

  // Redirect Studio Members to /studios
  // Dashboard is only for viewer-only users
  if (canManageStudios) {
    core.setRoute("/studios");
    return <div></div>;
  }

  const rooms = data.rooms?.rooms || [];
  const liveRooms = rooms.filter((r) => r.isActive);

  return (
    <div>
      <Header />
      <main className="dashboard-container">
        <div className="dashboard-content">
          <h1 className="dashboard-title">My Streams</h1>
          <p className="dashboard-description">
            All streaming rooms you have access to. Watch live streams or check
            back when they go live.
          </p>

          {rooms.length === 0 ? (
            <div className="rooms-empty">
              <div className="empty-icon">🎬</div>
              <h3>No Streams Available</h3>
              <p>
                You don't have access to any streaming rooms yet. Contact your
                administrator to be added to a studio.
              </p>
            </div>
          ) : (
            <div className="dashboard-rooms-section">
              {liveRooms.length > 0 && (
                <div className="live-indicator">
                  <span className="live-badge">
                    🔴 {liveRooms.length} Live Now
                  </span>
                </div>
              )}

              <div className="rooms-grid">
                {rooms.map((room) => (
                  <div
                    key={room.id}
                    className={`room-card ${room.isActive ? "room-live" : ""}`}
                  >
                    <div className="room-header">
                      <div className="room-studio">{room.studioName}</div>
                      {room.isActive && (
                        <span className="room-status active">🔴 Live</span>
                      )}
                    </div>

                    <h3 className="room-name">{room.name}</h3>

                    <div className="room-meta">
                      <span className="meta-item">Room #{room.roomNumber}</span>
                    </div>

                    <div className="room-actions">
                      <a
                        href={`/stream/${room.id}`}
                        className={`btn btn-sm ${room.isActive ? "btn-primary" : "btn-secondary"}`}
                      >
                        {room.isActive ? "Watch Stream" : "View Stream"}
                      </a>
                      <a
                        href={`/studio/${room.studioId}`}
                        className="btn btn-secondary btn-sm"
                      >
                        Studio
                      </a>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </div>
  );
}

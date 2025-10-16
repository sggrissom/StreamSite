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
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated
  let [authResp, authErr] = await server.GetAuthContext({});

  // Get accessible rooms
  let [roomsResp, roomsErr] = await server.ListMyAccessibleRooms({});

  return rpc.ok<Data>({
    authId: authResp?.id || 0,
    rooms: roomsResp || null,
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

  const rooms = data.rooms?.rooms || [];
  const liveRooms = rooms.filter((r) => r.isActive);

  return (
    <div>
      <Header />
      <main className="dashboard-container">
        <div className="dashboard-content">
          <h1 className="dashboard-title">Welcome to Stream</h1>
          <p className="dashboard-description">
            Access your streaming rooms and watch live content.
          </p>

          <div className="dashboard-actions">
            <a href="/studios" className="btn btn-primary btn-large">
              My Studios
            </a>
          </div>

          {/* Rooms section */}
          <div className="dashboard-rooms-section">
            <h2 className="section-title">
              My Rooms
              {liveRooms.length > 0 && (
                <span className="live-count"> ({liveRooms.length} live)</span>
              )}
            </h2>

            {rooms.length === 0 ? (
              <div className="rooms-empty">
                <div className="empty-icon">🎬</div>
                <h3>No Rooms Available</h3>
                <p>
                  You don't have access to any rooms yet. Join a studio to get
                  started.
                </p>
                <a href="/studios" className="btn btn-primary">
                  Browse Studios
                </a>
              </div>
            ) : (
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
                        View Studio
                      </a>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

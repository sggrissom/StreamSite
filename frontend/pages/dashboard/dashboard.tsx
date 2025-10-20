import * as preact from "preact";
import * as vlens from "vlens";
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

// Countdown timer for code expiration
type CountdownState = {
  timeRemaining: number; // seconds
  formattedTime: string;
  intervalId: number | null;
  cleanup: () => void;
};

const useCountdown = vlens.declareHook((expiresAt: string): CountdownState => {
  const state: CountdownState = {
    timeRemaining: 0,
    formattedTime: "",
    intervalId: null,
    cleanup: () => {
      if (state.intervalId !== null) {
        clearInterval(state.intervalId);
        state.intervalId = null;
      }
    },
  };

  // Calculate time remaining
  const updateTimeRemaining = () => {
    const now = Date.now();
    const expiryTime = new Date(expiresAt).getTime();
    const secondsLeft = Math.max(0, Math.floor((expiryTime - now) / 1000));
    state.timeRemaining = secondsLeft;

    // Check if this is a "never expires" code (>50 years away)
    const fiftyYearsInSeconds = 50 * 365.25 * 24 * 60 * 60;
    if (secondsLeft > fiftyYearsInSeconds) {
      state.formattedTime = "Never expires";
      state.cleanup(); // No need to keep updating
      vlens.scheduleRedraw();
      return;
    }

    // Format based on time remaining
    if (secondsLeft <= 0) {
      state.formattedTime = "Expired";
      state.cleanup();
    } else if (secondsLeft < 600) {
      // Under 10 minutes: show MM:SS
      const minutes = Math.floor(secondsLeft / 60);
      const seconds = secondsLeft % 60;
      state.formattedTime = `${minutes}:${seconds.toString().padStart(2, "0")}`;
    } else if (secondsLeft < 3600) {
      // Under 1 hour: show minutes
      const minutes = Math.ceil(secondsLeft / 60);
      state.formattedTime = `${minutes} minute${minutes !== 1 ? "s" : ""}`;
    } else if (secondsLeft < 86400) {
      // Under 24 hours: show hours
      const hours = Math.ceil(secondsLeft / 3600);
      state.formattedTime = `${hours} hour${hours !== 1 ? "s" : ""}`;
    } else {
      // 24+ hours: show days
      const days = Math.ceil(secondsLeft / 86400);
      state.formattedTime = `${days} day${days !== 1 ? "s" : ""}`;
    }

    vlens.scheduleRedraw();
  };

  // Initial calculation
  updateTimeRemaining();

  // Set up interval if not expired
  if (state.timeRemaining > 0 && state.intervalId === null) {
    state.intervalId = window.setInterval(updateTimeRemaining, 1000);
  }

  return state;
});

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
  const isCodeSession = data.authId === -1;

  // Redirect Studio Members to /studios (but not code sessions)
  // Dashboard is only for viewer-only users and code sessions
  if (canManageStudios && !isCodeSession) {
    core.setRoute("/studios");
    return <div></div>;
  }

  const rooms = data.rooms?.rooms || [];
  const liveRooms = rooms.filter((r) => r.isActive);
  const codeExpiresAt = data.rooms?.codeExpiresAt;

  // Set up countdown timer if we have an expiration (code session)
  let countdown: CountdownState | null = null;
  if (codeExpiresAt) {
    countdown = useCountdown(codeExpiresAt);
  }

  return (
    <div>
      <Header />
      <main className="dashboard-container">
        {/* Expiration banner for code sessions */}
        {countdown &&
          countdown.timeRemaining > 0 &&
          countdown.formattedTime !== "Never expires" && (
            <div className="expiration-banner">
              <div className="expiration-content">
                <span className="expiration-icon">⏱️</span>
                <span className="expiration-text">
                  Access expires in {countdown.formattedTime}
                </span>
              </div>
            </div>
          )}

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

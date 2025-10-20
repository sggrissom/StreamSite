import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "./stream-styles";

type Data = server.GetRoomDetailsResponse;

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

type StreamState = {
  videoElement: HTMLVideoElement | null;
  hlsInstance: any;
  streamUrl: string;
  isStreamLive: boolean;
  roomId: number;
  eventSource: EventSource | null;
  retryCount: number;
  onVideoRef: (el: HTMLVideoElement | null) => void;
  setStreamUrl: (url: string) => void;
  setStreamLive: (live: boolean) => void;
  connectSSE: () => void;
  disconnectSSE: () => void;
};

const useStreamPlayer = vlens.declareHook((): StreamState => {
  const state: StreamState = {
    videoElement: null,
    hlsInstance: null,
    streamUrl: "",
    isStreamLive: false,
    roomId: 0,
    eventSource: null,
    retryCount: 0,
    onVideoRef: (el: HTMLVideoElement | null) => {
      initializePlayer(state, el);
    },
    setStreamUrl: (url: string) => {
      if (state.streamUrl !== url) {
        state.streamUrl = url;
        // Reinitialize player with new URL if already attached
        if (state.videoElement && state.isStreamLive) {
          initializePlayer(state, state.videoElement);
        }
      }
    },
    setStreamLive: (live: boolean) => {
      if (state.isStreamLive === live) return;

      state.isStreamLive = live;

      if (live) {
        // Stream came online - initialize player
        if (state.videoElement && state.streamUrl) {
          initializePlayer(state, state.videoElement);
        }
      } else {
        // Stream went offline - cleanup player
        cleanupPlayer(state);
      }

      vlens.scheduleRedraw();
    },
    connectSSE: () => {
      if (state.roomId === 0) return;
      if (state.eventSource) return; // Already connected

      const url = `/api/room/events?roomId=${state.roomId}`;
      state.eventSource = new EventSource(url);

      state.eventSource.addEventListener("status", (e) => {
        const data = JSON.parse(e.data);
        state.setStreamLive(data.isActive);
      });

      state.eventSource.onerror = (err) => {
        console.warn("SSE error, will auto-reconnect:", err);
        // Browser auto-reconnects, no action needed
      };
    },
    disconnectSSE: () => {
      if (state.eventSource) {
        state.eventSource.close();
        state.eventSource = null;
      }
    },
  };
  return state;
});

// Helper to cleanup video player
function cleanupPlayer(state: StreamState) {
  if (state.hlsInstance?.destroy) {
    try {
      state.hlsInstance.destroy();
    } catch {}
    state.hlsInstance = null;
  }
  if (state.videoElement) {
    try {
      state.videoElement.removeAttribute("src");
      state.videoElement.load?.();
    } catch {}
  }
}

// lazy-load hls.js exactly once
function loadHlsOnce(): Promise<any> {
  const w = window as any;
  if (w.__hlsPromise) return w.__hlsPromise;
  w.__hlsPromise = new Promise((resolve, reject) => {
    if (w.Hls) return resolve(w.Hls);
    const s = document.createElement("script");
    s.src = "https://cdn.jsdelivr.net/npm/hls.js@latest";
    s.async = true;
    s.onload = () => resolve((window as any).Hls);
    s.onerror = reject;
    document.head.appendChild(s);
  });
  return w.__hlsPromise;
}

function initializePlayer(state: StreamState, el: HTMLVideoElement | null) {
  const url = state.streamUrl;

  // Don't initialize if we don't have a URL yet
  if (!url) return;

  // if the ref points to the same element, do nothing
  if (el === state.videoElement) return;

  // cleanup old attachment if ref changed or unmounted
  if (state.videoElement) {
    try {
      state.videoElement.removeAttribute("src");
      state.videoElement.load?.();
    } catch {}
  }
  if (state.hlsInstance?.destroy) {
    try {
      state.hlsInstance.destroy();
    } catch {}
    state.hlsInstance = null;
  }
  state.videoElement = el;

  // if unmounting, we're done
  if (!el) return;

  // init exactly once for this element
  if (el.canPlayType("application/vnd.apple.mpegurl")) {
    // Safari native HLS
    el.src = url;
    el.play().catch((e) => console.warn("Autoplay blocked:", e));
    return;
  }

  // Other browsers: HLS.js
  loadHlsOnce()
    .then((Hls) => {
      // element might have been replaced/unmounted meanwhile
      if (state.videoElement !== el) return;
      if (!Hls || !Hls.isSupported()) return;

      state.hlsInstance = new Hls({ lowLatencyMode: true });
      state.hlsInstance.attachMedia(el);

      // Auto-play when manifest is loaded
      state.hlsInstance.on(Hls.Events.MANIFEST_PARSED, () => {
        // Reset retry counter on successful load
        state.retryCount = 0;

        el.play().catch((e) => {
          // Autoplay might be blocked by browser policy
          console.warn("Autoplay blocked:", e);
        });
      });

      state.hlsInstance.loadSource(url);

      // optional: mild error recovery (prevents rapid reload storms)
      state.hlsInstance.on(Hls.Events.ERROR, (_e: any, data: any) => {
        if (!data?.fatal) return;
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            // If manifest load failed (stream just started but m3u8 not ready yet)
            // retry with exponential backoff
            if (data.details === Hls.ErrorDetails.MANIFEST_LOAD_ERROR) {
              const retryDelay = Math.min(
                1000 * Math.pow(2, state.retryCount),
                4000,
              );
              state.retryCount++;

              console.log(
                `Manifest load failed, retrying in ${retryDelay}ms (attempt ${state.retryCount})`,
              );

              setTimeout(() => {
                if (
                  state.hlsInstance &&
                  state.isStreamLive &&
                  state.streamUrl
                ) {
                  state.hlsInstance.loadSource(state.streamUrl);
                }
              }, retryDelay);
            } else {
              // Other network errors - retry immediately
              state.hlsInstance.startLoad();
            }
            break;
          case Hls.ErrorTypes.MEDIA_ERROR:
            state.hlsInstance.recoverMediaError();
            break;
          default:
            // Fatal error - cleanup but don't change live status
            // (SSE will notify us when stream actually ends)
            state.hlsInstance.destroy();
            state.hlsInstance = null;
        }
      });
    })
    .catch((e) => console.warn("Failed to load hls.js", e));
}

function extractRoomIdFromRoute(route: string): number | null {
  // Route format: /stream/:roomId (roomId is required)
  const parts = route.split("/").filter((p) => p);
  if (parts.length >= 2 && parts[0] === "stream") {
    const roomId = parseInt(parts[1]);
    return isNaN(roomId) ? null : roomId;
  }
  return null;
}

export async function fetch(route: string, prefix: string) {
  const roomId = extractRoomIdFromRoute(route);

  // If no roomId, pass 0 which will fail validation and return error
  return server.GetRoomDetails({ roomId: roomId || 0 });
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const state = useStreamPlayer();

  // Check if we have valid room details
  const hasValidRoom = data && data.room && data.success;

  // If no valid room, show error
  if (!hasValidRoom) {
    return (
      <div>
        <Header />
        <main className="stream-container">
          <div className="error-container">
            <h1>Room Not Found</h1>
            <p>
              The stream you're trying to access doesn't exist or you don't have
              permission to view it.
            </p>
            <a href="/" className="btn btn-primary">
              ← Back to Home
            </a>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  // Initialize room ID and connect SSE
  if (state.roomId !== data.room.id) {
    // Disconnect from old room if switching
    state.disconnectSSE();

    // Connect to new room
    state.roomId = data.room.id;

    // Sync initial state from backend BEFORE connecting SSE
    // After SSE connects, it becomes the source of truth
    state.isStreamLive = data.room.isActive;

    state.connectSSE();
  }

  // Build stream URL from room data
  // Pattern: /streams/room/{roomId}.m3u8
  // Backend will rewrite to: /streams/live/{streamKey}.m3u8
  const roomStreamUrl = `/streams/room/${data.room.id}.m3u8`;
  state.setStreamUrl(roomStreamUrl);

  return (
    <div>
      <Header />
      <main className="stream-container">
        <div className="stream-context">
          <div className="context-header">
            <div className="context-info">
              <h2 className="context-studio">{data.studioName}</h2>
              <h1 className="context-room">{data.room.name}</h1>
            </div>
            <div className="context-meta">
              <span className={`role-badge role-${data.myRole}`}>
                {data.myRoleName}
              </span>
              <span className="room-number">Room #{data.room.roomNumber}</span>
            </div>
          </div>
        </div>

        {data.isCodeAuth && (
          <div className="code-session-banner">
            <span className="banner-icon">🔑</span>
            <span className="banner-text">Watching via access code</span>
            {data.codeExpiresAt &&
              (() => {
                const countdown = useCountdown(data.codeExpiresAt);
                return (
                  <span className="banner-countdown">
                    {countdown.timeRemaining > 0
                      ? `Expires in ${countdown.formattedTime}`
                      : "Code expired"}
                  </span>
                );
              })()}
          </div>
        )}

        {state.isStreamLive ? (
          <div className="video-container">
            <video
              ref={state.onVideoRef}
              controls
              autoPlay
              muted
              playsInline
              preload="auto"
              className="video-player"
            />
          </div>
        ) : (
          <div className="stream-offline">
            <div className="offline-icon">⏸️</div>
            <h2>Stream Offline</h2>
            <p>This stream is currently not broadcasting.</p>
            <p className="offline-hint">
              The page will automatically refresh when the stream starts.
            </p>
          </div>
        )}

        <div className="stream-actions">
          <a
            href={`/studio/${data.room.studioId}`}
            className="btn btn-secondary"
          >
            ← Back to Studio
          </a>
        </div>
      </main>
      <Footer />
    </div>
  );
}

import * as preact from "preact";
import * as rpc from "vlens/rpc";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "./stream-styles";

type Data = server.GetRoomDetailsResponse | {};

type StreamState = {
  videoElement: HTMLVideoElement | null;
  hlsInstance: any;
  onVideoRef: (el: HTMLVideoElement | null) => void;
};

const useStreamPlayer = vlens.declareHook((): StreamState => {
  const state: StreamState = {
    videoElement: null,
    hlsInstance: null,
    onVideoRef: (el: HTMLVideoElement | null) => {
      initializePlayer(state, el);
    },
  };
  return state;
});

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
  const url = "/streams/live/stream.m3u8";

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
      state.hlsInstance.loadSource(url);

      // optional: mild error recovery (prevents rapid reload storms)
      state.hlsInstance.on(Hls.Events.ERROR, (_e: any, data: any) => {
        if (!data?.fatal) return;
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            state.hlsInstance.startLoad();
            break;
          case Hls.ErrorTypes.MEDIA_ERROR:
            state.hlsInstance.recoverMediaError();
            break;
          default:
            state.hlsInstance.destroy();
            state.hlsInstance = null;
        }
      });
    })
    .catch((e) => console.warn("Failed to load hls.js", e));
}

function extractRoomIdFromRoute(route: string): number | null {
  // Route format: /stream or /stream/:roomId
  const parts = route.split("/").filter((p) => p);
  if (parts.length >= 2 && parts[0] === "stream") {
    const roomId = parseInt(parts[1]);
    return isNaN(roomId) ? null : roomId;
  }
  return null;
}

export async function fetch(route: string, prefix: string) {
  const roomId = extractRoomIdFromRoute(route);

  if (roomId) {
    return server.GetRoomDetails({ roomId });
  }

  // No roomId - return empty response for generic stream
  return rpc.ok({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const state = useStreamPlayer();

  // Check if we have room details (vs empty object)
  const hasRoomDetails = data && "room" in data && data.success;
  const roomDetails = hasRoomDetails
    ? (data as server.GetRoomDetailsResponse)
    : null;

  return (
    <div>
      <Header />
      <main className="stream-container">
        {roomDetails && roomDetails.room ? (
          <div className="stream-context">
            <div className="context-header">
              <div className="context-info">
                <h2 className="context-studio">{roomDetails.studioName}</h2>
                <h1 className="context-room">{roomDetails.room.name}</h1>
              </div>
              <div className="context-meta">
                <span className={`role-badge role-${roomDetails.myRole}`}>
                  {roomDetails.myRoleName}
                </span>
                <span className="room-number">
                  Room #{roomDetails.room.roomNumber}
                </span>
              </div>
            </div>
          </div>
        ) : (
          <h1 className="stream-title">Live Stream</h1>
        )}

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

        {roomDetails && roomDetails.room && (
          <div className="stream-actions">
            <a
              href={`/studio/${roomDetails.room.studioId}`}
              className="btn btn-secondary"
            >
              ← Back to Studio
            </a>
          </div>
        )}
      </main>
      <Footer />
    </div>
  );
}

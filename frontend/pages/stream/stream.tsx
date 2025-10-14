import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "./stream-styles";

type Data = server.GetRoomDetailsResponse;

type StreamState = {
  videoElement: HTMLVideoElement | null;
  hlsInstance: any;
  streamUrl: string;
  onVideoRef: (el: HTMLVideoElement | null) => void;
  setStreamUrl: (url: string) => void;
};

const useStreamPlayer = vlens.declareHook((): StreamState => {
  const state: StreamState = {
    videoElement: null,
    hlsInstance: null,
    streamUrl: "",
    onVideoRef: (el: HTMLVideoElement | null) => {
      initializePlayer(state, el);
    },
    setStreamUrl: (url: string) => {
      if (state.streamUrl !== url) {
        state.streamUrl = url;
        // Reinitialize player with new URL if already attached
        if (state.videoElement) {
          initializePlayer(state, state.videoElement);
        }
      }
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

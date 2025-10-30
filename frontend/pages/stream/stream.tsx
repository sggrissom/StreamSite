import * as preact from "preact";
import * as vlens from "vlens";
import { registerCleanupFunction } from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { VideoControls } from "./VideoControls";
import { EmotePicker } from "./EmotePicker";
import "./stream-styles";

type Data = server.GetRoomDetailsResponse;

// Emote animation for floating effects
type EmoteAnimation = {
  id: string; // Unique ID for key
  emote: string; // Emoji character
  x: number; // Horizontal position (percentage)
  timestamp: number; // When it was created
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

type StreamState = {
  videoElement: HTMLVideoElement | null;
  containerElement: HTMLDivElement | null;
  hlsInstance: any;
  streamUrl: string;
  isStreamLive: boolean;
  isPlaying: boolean;
  isBehindLive: boolean;
  secondsBehindLive: number;
  roomId: number;
  eventSource: EventSource | null;
  retryCount: number;
  showRevokedModal: boolean;
  inGracePeriod: boolean;
  gracePeriodUntil: string | null;
  liveEdgeCheckInterval: number | null;
  offlineGraceTimer: number | null;
  isInOfflineGrace: boolean;
  viewerCount: number;
  activeEmotes: EmoteAnimation[];
  emoteCounter: number; // Counter for unique emote IDs
  controlsVisible: boolean;
  controlsAutoHideTimer: number | null;
  onVideoRef: (el: HTMLVideoElement | null) => void;
  onContainerRef: (el: HTMLDivElement | null) => void;
  setStreamUrl: (url: string) => void;
  setStreamLive: (live: boolean) => void;
  setViewerCount: (count: number) => void;
  addEmote: (emote: string) => void;
  removeEmote: (id: string) => void;
  sendEmote: (emote: string) => void;
  showControls: () => void;
  hideControls: () => void;
  connectSSE: () => void;
  disconnectSSE: () => void;
  handleCodeRevoked: () => void;
  handleCodeExpiredGrace: () => void;
  jumpToLive: () => void;
  checkLiveEdge: () => void;
  startLiveEdgeChecking: () => void;
  stopLiveEdgeChecking: () => void;
};

type OrientationState = {
  isPortrait: boolean;
  isMobile: boolean;
  showHint: boolean;
  dismissHint: () => void;
};

const useOrientation = vlens.declareHook((): OrientationState => {
  const state: OrientationState = {
    isPortrait: false,
    isMobile: false,
    showHint: false,
    dismissHint: () => {
      state.showHint = false;
      vlens.scheduleRedraw();
    },
  };

  // Check if mobile device
  const checkMobile = () => {
    state.isMobile = window.innerWidth <= 768;
  };

  // Check orientation
  const checkOrientation = () => {
    const isPortraitNow = window.matchMedia("(orientation: portrait)").matches;
    state.isPortrait = isPortraitNow;
    state.showHint = state.isMobile && state.isPortrait;
    vlens.scheduleRedraw();
  };

  checkMobile();
  checkOrientation();

  // Listen for orientation changes
  const orientationQuery = window.matchMedia("(orientation: portrait)");
  const handleOrientationChange = () => {
    checkOrientation();
  };

  if (orientationQuery.addEventListener) {
    orientationQuery.addEventListener("change", handleOrientationChange);
  } else {
    // Fallback for older browsers
    orientationQuery.addListener(handleOrientationChange);
  }

  // Listen for resize to detect mobile
  const handleResize = () => {
    checkMobile();
    checkOrientation();
  };
  window.addEventListener("resize", handleResize);

  return state;
});

// Module-level reference to stream state for cleanup during navigation
// Updated by useStreamPlayer hook on each render
let streamStateRef: StreamState | null = null;

// Register cleanup function with vlens
// This will be called automatically during navigation
registerCleanupFunction(() => {
  if (streamStateRef && streamStateRef.eventSource) {
    streamStateRef.disconnectSSE();
    cleanupPlayer(streamStateRef);
    streamStateRef.roomId = 0;
  }
});

const useStreamPlayer = vlens.declareHook((): StreamState => {
  const state: StreamState = {
    videoElement: null,
    containerElement: null,
    hlsInstance: null,
    streamUrl: "",
    isStreamLive: false,
    isPlaying: false,
    isBehindLive: false,
    secondsBehindLive: 0,
    roomId: 0,
    eventSource: null,
    retryCount: 0,
    showRevokedModal: false,
    inGracePeriod: false,
    gracePeriodUntil: null,
    liveEdgeCheckInterval: null,
    offlineGraceTimer: null,
    isInOfflineGrace: false,
    viewerCount: 0,
    activeEmotes: [],
    emoteCounter: 0,
    controlsVisible: true,
    controlsAutoHideTimer: null,
    onVideoRef: (el: HTMLVideoElement | null) => {
      initializePlayer(state, el);

      // Set up video click handler
      if (el) {
        const handleVideoClick = (e: MouseEvent) => {
          // Only toggle if clicking directly on video, not controls
          if ((e.target as HTMLElement).tagName === "VIDEO") {
            if (state.controlsVisible) {
              state.hideControls();
            } else {
              state.showControls();
            }
          }
        };
        el.addEventListener("click", handleVideoClick);
      }
    },
    onContainerRef: (el: HTMLDivElement | null) => {
      state.containerElement = el;
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
        // Stream came online - cancel grace period and initialize player
        if (state.offlineGraceTimer !== null) {
          clearTimeout(state.offlineGraceTimer);
          state.offlineGraceTimer = null;
        }
        state.isInOfflineGrace = false;

        if (state.videoElement && state.streamUrl) {
          initializePlayer(state, state.videoElement);
        }
        // Start checking live edge
        state.startLiveEdgeChecking();
      } else {
        // Stream went offline - start grace period (30 seconds)
        state.isInOfflineGrace = true;
        state.offlineGraceTimer = window.setTimeout(() => {
          // Grace period expired - cleanup player and show offline message
          state.isInOfflineGrace = false;
          state.offlineGraceTimer = null;
          cleanupPlayer(state);
          vlens.scheduleRedraw();
        }, 30000); // 30 seconds
      }

      vlens.scheduleRedraw();
    },
    setViewerCount: (count: number) => {
      state.viewerCount = count;
      vlens.scheduleRedraw();
    },
    addEmote: (emote: string) => {
      // Generate unique ID
      state.emoteCounter++;
      const id = `emote-${state.emoteCounter}-${Date.now()}`;

      // Random horizontal position (20% - 80%)
      const x = 20 + Math.random() * 60;

      // Create animation object
      const animation: EmoteAnimation = {
        id,
        emote,
        x,
        timestamp: Date.now(),
      };

      // Add to active emotes (limit to 20 for performance)
      state.activeEmotes = [...state.activeEmotes, animation].slice(-20);
      vlens.scheduleRedraw();

      // Auto-remove after 3 seconds (animation duration)
      setTimeout(() => {
        state.removeEmote(id);
      }, 3000);
    },
    removeEmote: (id: string) => {
      state.activeEmotes = state.activeEmotes.filter((e) => e.id !== id);
      vlens.scheduleRedraw();
    },
    sendEmote: (emote: string) => {
      // Call API to broadcast emote
      server
        .SendEmote({
          roomId: state.roomId,
          emote: emote,
        })
        .catch((err: any) => {
          console.warn("Failed to send emote:", err);
        });
    },
    showControls: () => {
      state.controlsVisible = true;
      vlens.scheduleRedraw();

      // Clear existing timer
      if (state.controlsAutoHideTimer !== null) {
        clearTimeout(state.controlsAutoHideTimer);
      }

      // Auto-hide after 3 seconds
      state.controlsAutoHideTimer = window.setTimeout(() => {
        state.hideControls();
      }, 3000);
    },
    hideControls: () => {
      state.controlsVisible = false;
      vlens.scheduleRedraw();
    },
    handleCodeRevoked: () => {
      state.showRevokedModal = true;
      cleanupPlayer(state);
      state.disconnectSSE();
      vlens.scheduleRedraw();
    },
    handleCodeExpiredGrace: () => {
      // Set grace period to 15 minutes from now
      const gracePeriodEnd = new Date(Date.now() + 15 * 60 * 1000);
      state.inGracePeriod = true;
      state.gracePeriodUntil = gracePeriodEnd.toISOString();
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

      state.eventSource.addEventListener("code_revoked", (e) => {
        state.handleCodeRevoked();
      });

      state.eventSource.addEventListener("code_expired_grace", (e) => {
        state.handleCodeExpiredGrace();
      });

      state.eventSource.addEventListener("viewer_count", (e) => {
        const data = JSON.parse(e.data);
        state.setViewerCount(data.count);
      });

      state.eventSource.addEventListener("emote", (e) => {
        const data = JSON.parse(e.data);
        state.addEmote(data.emote);
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
    jumpToLive: () => {
      if (!state.videoElement) return;

      try {
        if (state.hlsInstance) {
          // HLS.js: Use liveSyncPosition or seek to end
          const liveSyncPos = state.hlsInstance.liveSyncPosition;
          if (liveSyncPos !== undefined && liveSyncPos !== null) {
            state.videoElement.currentTime = liveSyncPos;
          } else {
            // Fallback: seek to the end of the seekable range
            const seekable = state.videoElement.seekable;
            if (seekable.length > 0) {
              state.videoElement.currentTime = seekable.end(
                seekable.length - 1,
              );
            }
          }
        } else {
          // Safari native HLS: seek to end of seekable range
          const seekable = state.videoElement.seekable;
          if (seekable.length > 0) {
            state.videoElement.currentTime = seekable.end(seekable.length - 1);
          }
        }

        // Resume playback if paused
        if (state.videoElement.paused) {
          state.videoElement
            .play()
            .catch((e) => console.warn("Play failed:", e));
        }

        // Immediately update state
        state.isBehindLive = false;
        state.secondsBehindLive = 0;
        vlens.scheduleRedraw();
      } catch (e) {
        console.warn("Failed to jump to live:", e);
      }
    },
    checkLiveEdge: () => {
      if (!state.videoElement || !state.isStreamLive) {
        state.isBehindLive = false;
        state.secondsBehindLive = 0;
        return;
      }

      try {
        let liveEdge = 0;
        let currentTime = state.videoElement.currentTime;

        if (state.hlsInstance) {
          // HLS.js: Use liveSyncPosition
          const liveSyncPos = state.hlsInstance.liveSyncPosition;
          if (liveSyncPos !== undefined && liveSyncPos !== null) {
            liveEdge = liveSyncPos;
          } else {
            // Fallback to seekable range
            const seekable = state.videoElement.seekable;
            if (seekable.length > 0) {
              liveEdge = seekable.end(seekable.length - 1);
            }
          }
        } else {
          // Safari native HLS: Use seekable range
          const seekable = state.videoElement.seekable;
          if (seekable.length > 0) {
            liveEdge = seekable.end(seekable.length - 1);
          }
        }

        // Calculate distance from live edge
        const distance = liveEdge - currentTime;
        state.secondsBehindLive = Math.max(0, distance);

        // Consider "behind" if more than 3 seconds from live edge OR if paused
        const threshold = 3;
        const wasBehind = state.isBehindLive;
        state.isBehindLive =
          state.secondsBehindLive > threshold || state.videoElement.paused;

        // Redraw if state changed
        if (wasBehind !== state.isBehindLive) {
          vlens.scheduleRedraw();
        }
      } catch (e) {
        console.warn("Error checking live edge:", e);
      }
    },
    startLiveEdgeChecking: () => {
      if (state.liveEdgeCheckInterval !== null) return; // Already running

      // Check immediately
      state.checkLiveEdge();

      // Check every 2 seconds
      state.liveEdgeCheckInterval = window.setInterval(() => {
        state.checkLiveEdge();
      }, 2000);
    },
    stopLiveEdgeChecking: () => {
      if (state.liveEdgeCheckInterval !== null) {
        clearInterval(state.liveEdgeCheckInterval);
        state.liveEdgeCheckInterval = null;
      }
      state.isBehindLive = false;
      state.secondsBehindLive = 0;
    },
  };

  // Update module-level reference for cleanup during navigation
  streamStateRef = state;

  return state;
});

// Helper to cleanup video player
function cleanupPlayer(state: StreamState) {
  // Stop live edge checking
  state.stopLiveEdgeChecking();

  // Clear offline grace timer if active
  if (state.offlineGraceTimer !== null) {
    clearTimeout(state.offlineGraceTimer);
    state.offlineGraceTimer = null;
  }
  state.isInOfflineGrace = false;

  // Clear controls auto-hide timer if active
  if (state.controlsAutoHideTimer !== null) {
    clearTimeout(state.controlsAutoHideTimer);
    state.controlsAutoHideTimer = null;
  }

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

  // Add play/pause event listeners to track playing state
  const handlePlay = () => {
    state.isPlaying = true;
    vlens.scheduleRedraw();
  };
  const handlePause = () => {
    state.isPlaying = false;
    vlens.scheduleRedraw();
  };
  const handleEnded = () => {
    // If video ended during grace period, immediately show offline message
    if (state.isInOfflineGrace) {
      if (state.offlineGraceTimer !== null) {
        clearTimeout(state.offlineGraceTimer);
        state.offlineGraceTimer = null;
      }
      state.isInOfflineGrace = false;
      cleanupPlayer(state);
      vlens.scheduleRedraw();
    }
  };
  el.addEventListener("play", handlePlay);
  el.addEventListener("pause", handlePause);
  el.addEventListener("playing", handlePlay);
  el.addEventListener("waiting", handlePause);
  el.addEventListener("ended", handleEnded);

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
  const orientation = useOrientation();

  // Check if we have valid room details
  const hasValidRoom = data && data.room;

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
    state.setStreamLive(data.room.isActive);

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
      <main className={`stream-container`}>
        <div className="stream-context">
          <div className="context-header">
            <div className="context-info">
              <h2 className="context-studio">{data.studioName}</h2>
              <h1 className="context-room">{data.room.name}</h1>
            </div>
            <div className="context-meta">
              <span className="viewer-count">
                👁 {state.viewerCount} watching
              </span>
              <span className={`role-badge role-${data.myRole}`}>
                {data.myRoleName}
              </span>
              <span className="room-number">Room #{data.room.roomNumber}</span>
            </div>
          </div>
        </div>

        {data.isCodeAuth && (
          <div
            className={`code-session-banner ${state.inGracePeriod || (data.codeExpiresAt && new Date(data.codeExpiresAt) < new Date()) ? "grace-period" : ""}`}
          >
            <span className="banner-icon">
              {state.inGracePeriod ||
              (data.codeExpiresAt && new Date(data.codeExpiresAt) < new Date())
                ? "⚠️"
                : "🔑"}
            </span>
            {state.inGracePeriod ||
            (data.codeExpiresAt &&
              new Date(data.codeExpiresAt) < new Date()) ? (
              <>
                <span className="banner-text">
                  Code expired • Grace period:
                </span>
                {state.gracePeriodUntil &&
                  (() => {
                    const countdown = useCountdown(state.gracePeriodUntil);
                    return (
                      <span className="banner-countdown">
                        {countdown.timeRemaining > 0
                          ? countdown.formattedTime
                          : "Grace period ended"}
                      </span>
                    );
                  })()}
              </>
            ) : (
              <>
                <span className="banner-text">Watching via access code</span>
                {data.codeExpiresAt &&
                  (() => {
                    const countdown = useCountdown(data.codeExpiresAt);

                    let displayText = "Code expired";
                    if (countdown.timeRemaining > 0) {
                      if (countdown.formattedTime === "Never expires") {
                        displayText = "Never expires";
                      } else {
                        displayText = `Expires in ${countdown.formattedTime}`;
                      }
                    }

                    return (
                      <span className="banner-countdown">{displayText}</span>
                    );
                  })()}
              </>
            )}
          </div>
        )}

        {state.showRevokedModal && (
          <div className="revoked-modal-overlay">
            <div className="revoked-modal">
              <h2>Access Revoked</h2>
              <p>
                The access code for this stream has been revoked by the stream
                owner.
              </p>
              <a href="/watch" className="btn btn-primary">
                Enter New Code
              </a>
            </div>
          </div>
        )}

        {state.isStreamLive || state.isInOfflineGrace ? (
          <div className="video-container" ref={state.onContainerRef}>
            <video
              ref={state.onVideoRef}
              autoPlay
              muted
              playsInline
              preload="auto"
              className="video-player"
            />

            {/* Floating emote overlay */}
            <div className="emote-overlay">
              {state.activeEmotes.map((emote) => (
                <div
                  key={emote.id}
                  className="floating-emote"
                  style={{ left: `${emote.x}%` }}
                >
                  {emote.emote}
                </div>
              ))}
            </div>

            <VideoControls
              id={`video-controls-${data.room?.id || 0}`}
              videoElement={state.videoElement}
              containerElement={state.containerElement}
              isPlaying={state.isPlaying}
              isBehindLive={state.isBehindLive}
              secondsBehindLive={state.secondsBehindLive}
              viewerCount={state.viewerCount}
              visible={state.controlsVisible}
              onJumpToLive={state.jumpToLive}
              onShowControls={state.showControls}
              onHideControls={state.hideControls}
            />

            {/* Emote picker */}
            <EmotePicker
              id={`emote-picker-${data.room?.id || 0}`}
              controlsVisible={state.controlsVisible}
              onLocalEmote={state.addEmote}
              onSendEmote={state.sendEmote}
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
          {!data.isCodeAuth && data.myRole >= server.StudioRoleMember ? (
            <a
              href={`/studio/${data.room.studioId}`}
              className="btn btn-secondary"
            >
              ← Back to Studio
            </a>
          ) : (
            <a href="/dashboard" className="btn btn-secondary">
              ← Back to My Streams
            </a>
          )}
        </div>
      </main>
      <Footer />
    </div>
  );
}

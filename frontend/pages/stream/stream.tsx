import * as preact from "preact";
import * as vlens from "vlens";
import { registerCleanupFunction } from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { VideoControls } from "./VideoControls";
import { EmotePicker } from "./EmotePicker";
import { StatsOverlay, type StreamStats } from "./StatsOverlay";
import { ChatSidebar } from "./ChatSidebar";
import "./stream-styles";
import "./chat-styles";

type Data = server.GetRoomDetailsResponse;
type ChatMessage = server.ChatMessage;

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
  isHlsReady: boolean;
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
  hlsReadyTimeout: number | null;
  hlsReadyTimedOut: boolean;
  viewerCount: number;
  activeEmotes: EmoteAnimation[];
  emoteCounter: number; // Counter for unique emote IDs
  ignoreNextEmote: string | null; // Ignore next SSE emote matching this (our own emote coming back)
  controlsVisible: boolean;
  controlsAutoHideTimer: number | null;
  statsVisible: boolean;
  streamStats: StreamStats | null;
  statsUpdateInterval: number | null;
  connectionStartTime: number | null;
  streamStartTime: number | null;
  keydownHandler: ((e: KeyboardEvent) => void) | null;

  // Performance Metrics Tracking
  metricsPlayAttemptStart: number | null; // When play() was called
  metricsTimeToFirstFrame: number; // Milliseconds to first frame (-1 if failed)
  metricsStartupSucceeded: boolean; // Whether playback started
  metricsRebufferEvents: number; // Count of buffering interruptions
  metricsLastBufferStart: number | null; // When current buffer started
  metricsTotalRebufferMs: number; // Total time spent buffering
  metricsWatchStartTime: number | null; // When viewing started
  metricsQualitySeconds: { [key: string]: number }; // Seconds at each quality
  metricsLastQualitySwitch: number | null; // When last quality change happened
  metricsCurrentQuality: string; // Current quality level (480p/720p/1080p)
  metricsNetworkErrors: number; // Network/loading errors
  metricsMediaErrors: number; // Decoding/format errors
  metricsReportInterval: number | null; // Periodic metrics reporting timer
  metricsReported: boolean; // Whether metrics have been sent
  onVideoRef: (el: HTMLVideoElement | null) => void;
  onContainerRef: (el: HTMLDivElement | null) => void;
  setStreamUrl: (url: string) => void;
  setStreamLive: (live: boolean) => void;
  setHlsReady: (ready: boolean) => void;
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
  toggleStats: () => void;
  updateStats: () => void;
  startStatsUpdating: () => void;
  stopStatsUpdating: () => void;

  // Chat state
  chatMessages: ChatMessage[];
  isChatVisible: boolean;
  addChatMessage: (msg: ChatMessage) => void;
  sendChatMessage: (text: string) => void;
  toggleChat: () => void;
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
  // Helper function to try player initialization
  // Called whenever state changes that might enable initialization
  const tryInitializePlayer = (state: StreamState) => {
    // Only initialize if ALL conditions are met and player isn't already initialized
    if (
      state.videoElement &&
      state.streamUrl &&
      state.isStreamLive &&
      state.isHlsReady &&
      !state.hlsInstance
    ) {
      initializePlayer(state, state.videoElement);
    }
  };

  const state: StreamState = {
    videoElement: null,
    containerElement: null,
    hlsInstance: null,
    streamUrl: "",
    isStreamLive: false,
    isHlsReady: false,
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
    hlsReadyTimeout: null,
    hlsReadyTimedOut: false,
    viewerCount: 0,
    activeEmotes: [],
    emoteCounter: 0,
    ignoreNextEmote: null,
    controlsVisible: true,
    controlsAutoHideTimer: null,
    statsVisible: false,
    streamStats: null,
    statsUpdateInterval: null,
    connectionStartTime: null,
    streamStartTime: null,
    keydownHandler: null,

    // Performance Metrics Tracking - Initialize
    metricsPlayAttemptStart: null,
    metricsTimeToFirstFrame: -1,
    metricsStartupSucceeded: false,
    metricsRebufferEvents: 0,
    metricsLastBufferStart: null,
    metricsTotalRebufferMs: 0,
    metricsWatchStartTime: null,
    metricsQualitySeconds: {},
    metricsLastQualitySwitch: null,
    metricsCurrentQuality: "unknown",
    metricsNetworkErrors: 0,
    metricsMediaErrors: 0,
    metricsReportInterval: null,
    metricsReported: false,

    // Chat state - Initialize
    chatMessages: [],
    isChatVisible: true, // Default visible on desktop

    onVideoRef: (el: HTMLVideoElement | null) => {
      // If same element, do nothing (avoid re-processing during re-renders)
      if (el === state.videoElement) return;

      // Cleanup old attachment if element is changing or unmounting
      if (state.videoElement && state.videoElement !== el) {
        try {
          state.videoElement.removeAttribute("src");
          state.videoElement.load?.();
        } catch {}

        if (state.hlsInstance?.destroy) {
          try {
            state.hlsInstance.destroy();
          } catch {}
          state.hlsInstance = null;
        }
      }

      // Store new element reference
      state.videoElement = el;

      // Try to initialize player if all conditions are now met
      tryInitializePlayer(state);

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
        // Try to initialize player now that URL is set
        // This provides a safety net for page reload scenarios where video element
        // might mount before URL is set, ensuring initialization happens either way
        tryInitializePlayer(state);
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

        // Track stream start time for stats
        if (!state.streamStartTime) {
          state.streamStartTime = Date.now();
        }

        // Track connection start time
        if (!state.connectionStartTime) {
          state.connectionStartTime = Date.now();
        }

        // Start HLS ready timeout (30 seconds)
        if (!state.isHlsReady && state.hlsReadyTimeout === null) {
          state.hlsReadyTimedOut = false;
          state.hlsReadyTimeout = window.setTimeout(() => {
            state.hlsReadyTimedOut = true;
            state.hlsReadyTimeout = null;
            vlens.scheduleRedraw();
          }, 30000);
        }

        // Try to initialize player now that stream is live
        tryInitializePlayer(state);
        // Start checking live edge
        state.startLiveEdgeChecking();
      } else {
        // Stream went offline - start grace period (30 seconds)
        state.isInOfflineGrace = true;
        state.offlineGraceTimer = window.setTimeout(() => {
          // Grace period expired - cleanup player and show offline message
          state.isInOfflineGrace = false;
          state.offlineGraceTimer = null;
          state.streamStartTime = null; // Reset stream start time
          cleanupPlayer(state);
          vlens.scheduleRedraw();
        }, 30000); // 30 seconds
      }

      vlens.scheduleRedraw();
    },
    setHlsReady: (ready: boolean) => {
      if (state.isHlsReady === ready) return;

      state.isHlsReady = ready;

      if (ready) {
        // HLS became ready - clear timeout
        if (state.hlsReadyTimeout !== null) {
          clearTimeout(state.hlsReadyTimeout);
          state.hlsReadyTimeout = null;
        }
        state.hlsReadyTimedOut = false;

        // Try to initialize player now that HLS is ready
        tryInitializePlayer(state);
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
      // Mark this emote to be ignored when it comes back via SSE
      state.ignoreNextEmote = emote;

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
    addChatMessage: (msg: ChatMessage) => {
      state.chatMessages.push(msg);
      // Keep only last 200 messages in memory
      if (state.chatMessages.length > 200) {
        state.chatMessages = state.chatMessages.slice(-200);
      }
      vlens.scheduleRedraw();
    },
    sendChatMessage: (text: string) => {
      // Call API to send message
      // Don't add optimistically - SSE will broadcast it back to us
      server
        .SendChatMessage({
          roomId: state.roomId,
          text: text,
        })
        .then(([resp, err]) => {
          if (err) {
            console.warn("Failed to send chat message:", err);
          }
        })
        .catch((err: any) => {
          console.warn("Failed to send chat message:", err);
        });
    },
    toggleChat: () => {
      state.isChatVisible = !state.isChatVisible;
      vlens.scheduleRedraw();
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
        if (data.isHlsReady !== undefined) {
          state.setHlsReady(data.isHlsReady);
        }
      });

      state.eventSource.addEventListener("stream_ready", (e) => {
        state.setHlsReady(true);
        state.retryCount = 0; // Reset retry count now that HLS is ready
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
        const emote = data.emote;

        // Check if this is our own emote coming back
        if (emote === state.ignoreNextEmote) {
          state.ignoreNextEmote = null; // Clear the flag
          return; // Skip display (already shown locally)
        }

        // This is an emote from another viewer - display it
        state.addEmote(emote);
      });

      state.eventSource.addEventListener("chat_message", (e) => {
        const message: ChatMessage = JSON.parse(e.data);

        // Check if we already have this message (deduplicate)
        const exists = state.chatMessages.some((msg) => msg.id === message.id);
        if (!exists) {
          state.addChatMessage(message);
        }
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
    toggleStats: () => {
      state.statsVisible = !state.statsVisible;
      if (state.statsVisible) {
        state.startStatsUpdating();
      } else {
        state.stopStatsUpdating();
      }
      vlens.scheduleRedraw();
    },
    updateStats: () => {
      if (!state.videoElement) return;

      // Helper to format time duration
      const formatDuration = (seconds: number): string => {
        if (seconds < 60) return `${Math.floor(seconds)}s`;
        const mins = Math.floor(seconds / 60);
        if (mins < 60) return `${mins}m ${Math.floor(seconds % 60)}s`;
        const hours = Math.floor(mins / 60);
        return `${hours}h ${mins % 60}m`;
      };

      // Default stats
      let currentResolution = "N/A";
      let currentBitrate = "N/A";
      let bandwidthEstimate = "N/A";
      let bufferHealth = "N/A";
      let isAutoQuality = true;
      let droppedFrames = "N/A";

      // HLS.js stats
      if (state.hlsInstance) {
        try {
          const currentLevel = state.hlsInstance.currentLevel;
          const levels = state.hlsInstance.levels;
          const autoLevelEnabled = state.hlsInstance.autoLevelEnabled;

          if (currentLevel >= 0 && levels && levels[currentLevel]) {
            // Read actual resolution from level metadata instead of assuming order
            const level = levels[currentLevel];
            const height = level.height;
            if (height) {
              currentResolution = `${height}p`;
            } else {
              currentResolution = `Level ${currentLevel}`;
            }
            const bitrateKbps = Math.round(level.bitrate / 1000);
            currentBitrate = `${bitrateKbps} Kbps`;
          }

          isAutoQuality = autoLevelEnabled !== false;

          const bandwidth = state.hlsInstance.bandwidthEstimate;
          if (bandwidth) {
            const bandwidthMbps = (bandwidth / 1000000).toFixed(2);
            bandwidthEstimate = `${bandwidthMbps} Mbps`;
          }
        } catch (e) {
          // Ignore errors reading HLS.js state
        }
      }

      // Buffer health
      try {
        const buffered = state.videoElement.buffered;
        if (buffered.length > 0) {
          const bufferSeconds =
            buffered.end(buffered.length - 1) - state.videoElement.currentTime;
          bufferHealth = `${bufferSeconds.toFixed(1)}s`;
        }
      } catch (e) {
        // Ignore errors
      }

      // Dropped frames (browser-specific)
      try {
        const videoElement = state.videoElement as any;
        if (videoElement.getVideoPlaybackQuality) {
          const quality = videoElement.getVideoPlaybackQuality();
          const dropped = quality.droppedVideoFrames || 0;
          const total = quality.totalVideoFrames || 0;
          if (total > 0) {
            const dropRate = ((dropped / total) * 100).toFixed(2);
            droppedFrames = `${dropped} (${dropRate}%)`;
          }
        }
      } catch (e) {
        // Not supported in this browser
      }

      // Video resolution
      const videoResolution = `${state.videoElement.videoWidth}x${state.videoElement.videoHeight}`;

      // Connection time
      let connectionTime = "N/A";
      if (state.connectionStartTime) {
        const elapsedSeconds = (Date.now() - state.connectionStartTime) / 1000;
        connectionTime = formatDuration(elapsedSeconds);
      }

      // Stream uptime
      let streamUptime = "N/A";
      if (state.streamStartTime) {
        const elapsedSeconds = (Date.now() - state.streamStartTime) / 1000;
        streamUptime = formatDuration(elapsedSeconds);
      }

      state.streamStats = {
        currentResolution,
        currentBitrate,
        bandwidthEstimate,
        bufferHealth,
        isAutoQuality,
        droppedFrames,
        videoResolution,
        connectionTime,
        streamUptime,
      };

      vlens.scheduleRedraw();
    },
    startStatsUpdating: () => {
      if (state.statsUpdateInterval !== null) return; // Already running

      // Update immediately
      state.updateStats();

      // Update every 1 second
      state.statsUpdateInterval = window.setInterval(() => {
        state.updateStats();
      }, 1000);
    },
    stopStatsUpdating: () => {
      if (state.statsUpdateInterval !== null) {
        clearInterval(state.statsUpdateInterval);
        state.statsUpdateInterval = null;
      }
    },
  };

  // Set up keyboard listener for 'i' key to toggle stats
  state.keydownHandler = (e: KeyboardEvent) => {
    if (e.key === "i" || e.key === "I") {
      state.toggleStats();
    }
  };
  document.addEventListener("keydown", state.keydownHandler);

  // Update module-level reference for cleanup during navigation
  streamStateRef = state;

  return state;
});

// Helper to cleanup video player
function cleanupPlayer(state: StreamState) {
  // Stop live edge checking
  state.stopLiveEdgeChecking();

  // Stop stats updating
  state.stopStatsUpdating();

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

  // Remove keyboard listener
  if (state.keydownHandler) {
    document.removeEventListener("keydown", state.keydownHandler);
    state.keydownHandler = null;
  }

  // Report final metrics before cleanup
  reportMetrics(state);

  // Clear metrics reporting interval
  if (state.metricsReportInterval !== null) {
    clearInterval(state.metricsReportInterval);
    state.metricsReportInterval = null;
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

// Helper: Update quality tracking (accumulate seconds at each quality level)
function updateQualityTracking(state: StreamState) {
  if (
    state.metricsLastQualitySwitch !== null &&
    state.metricsCurrentQuality !== "unknown"
  ) {
    const now = Date.now();
    const secondsAtQuality = (now - state.metricsLastQualitySwitch) / 1000;
    const qualityKey = state.metricsCurrentQuality;
    state.metricsQualitySeconds[qualityKey] =
      (state.metricsQualitySeconds[qualityKey] || 0) + secondsAtQuality;
  }
  state.metricsLastQualitySwitch = Date.now();
}

// Helper: Map HLS level height to quality string
function getQualityString(height: number): string {
  if (height >= 1080) return "1080p";
  if (height >= 720) return "720p";
  if (height >= 480) return "480p";
  return "unknown";
}

// Helper: Report metrics to backend
function reportMetrics(state: StreamState) {
  if (state.metricsReported || state.roomId === 0) return;

  // Update final quality tracking
  updateQualityTracking(state);

  // Calculate total watch time
  const watchSeconds = state.metricsWatchStartTime
    ? Math.floor((Date.now() - state.metricsWatchStartTime) / 1000)
    : 0;

  // Calculate rebuffer time in seconds
  const rebufferSeconds = Math.floor(state.metricsTotalRebufferMs / 1000);

  // Calculate average bitrate from quality distribution
  const totalQualitySeconds =
    (state.metricsQualitySeconds["480p"] || 0) +
    (state.metricsQualitySeconds["720p"] || 0) +
    (state.metricsQualitySeconds["1080p"] || 0);

  let avgBitrate = 0;
  if (totalQualitySeconds > 0) {
    // Weighted average based on nominal bitrates
    avgBitrate =
      ((state.metricsQualitySeconds["480p"] || 0) * 1.2 +
        (state.metricsQualitySeconds["720p"] || 0) * 2.5 +
        (state.metricsQualitySeconds["1080p"] || 0) * 5.0) /
      totalQualitySeconds;
  }

  const metrics = {
    roomId: state.roomId,
    timeToFirstFrame: state.metricsTimeToFirstFrame,
    startupSucceeded: state.metricsStartupSucceeded,
    rebufferEvents: state.metricsRebufferEvents,
    rebufferSeconds: rebufferSeconds,
    watchSeconds: watchSeconds,
    seconds480p: Math.floor(state.metricsQualitySeconds["480p"] || 0),
    seconds720p: Math.floor(state.metricsQualitySeconds["720p"] || 0),
    seconds1080p: Math.floor(state.metricsQualitySeconds["1080p"] || 0),
    avgBitrate: avgBitrate,
    networkErrors: state.metricsNetworkErrors,
    mediaErrors: state.metricsMediaErrors,
  };

  // Only report if we have meaningful data
  if (watchSeconds > 5 || state.metricsStartupSucceeded) {
    server
      .ReportStreamMetrics({ metrics })
      .then(() => {
        console.log("Metrics reported successfully", metrics);
        state.metricsReported = true;
      })
      .catch((err: any) => {
        console.warn("Failed to report metrics:", err);
      });
  }
}

function initializePlayer(state: StreamState, el: HTMLVideoElement | null) {
  const url = state.streamUrl;

  // Don't initialize if we don't have a URL yet
  if (!url) return;

  // Don't initialize if stream is live but HLS is not ready yet
  if (state.isStreamLive && !state.isHlsReady) return;

  // Check if we're switching elements or reinitializing the same element
  const isSameElement = el === state.videoElement;

  // cleanup old attachment if ref changed or unmounted (but not if same element)
  if (state.videoElement && !isSameElement) {
    try {
      state.videoElement.removeAttribute("src");
      state.videoElement.load?.();
    } catch {}
  }

  // Always destroy HLS instance before reinitializing
  if (state.hlsInstance?.destroy) {
    try {
      state.hlsInstance.destroy();
    } catch {}
    state.hlsInstance = null;
  }

  state.videoElement = el;

  // if unmounting, we're done
  if (!el) return;

  // Add play/pause event listeners to track playing state and metrics
  const handlePlay = () => {
    state.isPlaying = true;
    vlens.scheduleRedraw();
  };
  const handlePlaying = () => {
    state.isPlaying = true;

    // Track Time To First Frame (TTFF)
    if (
      state.metricsPlayAttemptStart !== null &&
      !state.metricsStartupSucceeded
    ) {
      const ttff = Date.now() - state.metricsPlayAttemptStart;
      state.metricsTimeToFirstFrame = ttff;
      state.metricsStartupSucceeded = true;
      state.metricsWatchStartTime = Date.now();
      console.log(`TTFF: ${ttff}ms`);
    }

    // Track end of buffering event
    if (state.metricsLastBufferStart !== null) {
      const bufferDuration = Date.now() - state.metricsLastBufferStart;
      state.metricsTotalRebufferMs += bufferDuration;
      state.metricsLastBufferStart = null;
    }

    vlens.scheduleRedraw();
  };
  const handleWaiting = () => {
    state.isPlaying = false;

    // Track start of buffering event (only if playback has started)
    if (
      state.metricsStartupSucceeded &&
      state.metricsLastBufferStart === null
    ) {
      state.metricsLastBufferStart = Date.now();
      state.metricsRebufferEvents++;
    }

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
  el.addEventListener("playing", handlePlaying);
  el.addEventListener("waiting", handleWaiting);
  el.addEventListener("ended", handleEnded);

  // Start muted for autoplay compatibility - preference restored after playback starts
  el.muted = true;

  // init exactly once for this element
  if (el.canPlayType("application/vnd.apple.mpegurl")) {
    // Safari native HLS
    el.src = url;

    // Track play attempt start (for TTFF measurement)
    state.metricsPlayAttemptStart = Date.now();

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

        // Track play attempt start (for TTFF measurement)
        state.metricsPlayAttemptStart = Date.now();

        el.play().catch((e) => {
          // Autoplay might be blocked by browser policy
          console.warn("Autoplay blocked:", e);
        });
      });

      // Track quality level switches
      state.hlsInstance.on(Hls.Events.LEVEL_SWITCHED, (_e: any, data: any) => {
        if (data.level >= 0 && state.hlsInstance.levels[data.level]) {
          const level = state.hlsInstance.levels[data.level];
          const quality = getQualityString(level.height);

          // Update tracking for previous quality
          if (state.metricsCurrentQuality !== "unknown") {
            updateQualityTracking(state);
          }

          state.metricsCurrentQuality = quality;
          state.metricsLastQualitySwitch = Date.now();

          console.log(`Quality switched to ${quality} (${level.height}p)`);
        }
      });

      state.hlsInstance.loadSource(url);

      // optional: mild error recovery (prevents rapid reload storms)
      state.hlsInstance.on(Hls.Events.ERROR, (_e: any, data: any) => {
        // Track errors for metrics
        if (data?.fatal) {
          if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
            state.metricsNetworkErrors++;
          } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
            state.metricsMediaErrors++;
          }
        }

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
                  state.isHlsReady &&
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

// Format class end time as "Ends at 3:00 PM" or "Ends in 15 mins"
function formatEndTime(isoTime: string): string {
  const endTime = new Date(isoTime);
  const now = new Date();
  const minsUntilEnd = Math.floor(
    (endTime.getTime() - now.getTime()) / (1000 * 60),
  );

  if (minsUntilEnd <= 0) {
    return "Class ended";
  } else if (minsUntilEnd < 60) {
    return `Ends in ${minsUntilEnd} min${minsUntilEnd !== 1 ? "s" : ""}`;
  } else {
    const timeStr = endTime.toLocaleTimeString("en-US", {
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    });
    return `Ends at ${timeStr}`;
  }
}

// Format class start time as "Starts in 15 mins" or "Starts at 3:00 PM"
function formatStartTime(isoTime: string): string {
  const startTime = new Date(isoTime);
  const now = new Date();
  const minsUntilStart = Math.floor(
    (startTime.getTime() - now.getTime()) / (1000 * 60),
  );

  if (minsUntilStart <= 0) {
    return "Starting now";
  } else if (minsUntilStart < 60) {
    return `Starts in ${minsUntilStart} min${minsUntilStart !== 1 ? "s" : ""}`;
  } else if (minsUntilStart < 1440) {
    // Less than 24 hours
    const hours = Math.floor(minsUntilStart / 60);
    return `Starts in ${hours} hour${hours !== 1 ? "s" : ""}`;
  } else {
    const timeStr = startTime.toLocaleTimeString("en-US", {
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    });
    const dateStr = startTime.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    });
    return `Starts ${dateStr} at ${timeStr}`;
  }
}

// Upcoming Schedule Component - Collapsible view of upcoming classes
type UpcomingScheduleState = {
  isExpanded: boolean;
};

const useUpcomingSchedule = vlens.declareHook(
  (): UpcomingScheduleState => ({
    isExpanded: false,
  }),
);

interface UpcomingScheduleProps {
  classes: server.ClassScheduleWithInstance[];
}

function UpcomingSchedule({ classes }: UpcomingScheduleProps) {
  const state = useUpcomingSchedule();

  // Show first 3 by default, rest when expanded
  const visibleClasses = state.isExpanded ? classes : classes.slice(0, 3);
  const hasMore = classes.length > 3;

  return (
    <div className="upcoming-schedule">
      <div
        className="schedule-header"
        onClick={() => {
          state.isExpanded = !state.isExpanded;
          vlens.scheduleRedraw();
        }}
      >
        <span className="schedule-title">
          📅 Upcoming Schedule ({classes.length})
        </span>
        {hasMore && (
          <span className="schedule-toggle">
            {state.isExpanded ? "▲ Show less" : "▼ View all"}
          </span>
        )}
      </div>
      <div className="schedule-list">
        {visibleClasses.map((classItem, index) => {
          const startTime = new Date(classItem.instanceStart);
          const timeStr = startTime.toLocaleTimeString("en-US", {
            hour: "numeric",
            minute: "2-digit",
            hour12: true,
          });
          const dateStr = startTime.toLocaleDateString("en-US", {
            weekday: "short",
            month: "short",
            day: "numeric",
          });

          return (
            <div
              key={`${classItem.schedule.id}-${classItem.instanceStart}-${index}`}
              className="schedule-item"
            >
              <div className="schedule-item-time">
                <div className="time">{timeStr}</div>
                <div className="date">{dateStr}</div>
              </div>
              <div className="schedule-item-details">
                <div className="class-name">{classItem.schedule.name}</div>
                {classItem.schedule.description && (
                  <div className="class-description">
                    {classItem.schedule.description}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
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
    state.setHlsReady(data.room.isHlsReady || false);

    // Load chat history
    server
      .GetChatHistory({
        roomId: data.room.id,
        limit: 100,
      })
      .then(([resp, err]) => {
        if (!err && resp?.messages) {
          state.chatMessages = resp.messages;
          vlens.scheduleRedraw();
        }
      })
      .catch((err) => {
        console.warn("Failed to load chat history:", err);
      });

    state.connectSSE();

    // Setup metrics reporting
    // Report metrics when user leaves page
    const handleBeforeUnload = () => {
      reportMetrics(state);
    };
    window.addEventListener("beforeunload", handleBeforeUnload);

    // Report metrics periodically (every 5 minutes) for long sessions
    if (state.metricsReportInterval === null) {
      state.metricsReportInterval = window.setInterval(
        () => {
          reportMetrics(state);
        },
        5 * 60 * 1000,
      ); // 5 minutes
    }
  }

  // Build stream URL from room data
  // ABR HLS pattern: /hls/{roomId}/master.m3u8
  // Backend serves adaptive bitrate HLS with 720p and 480p variants
  // HLS.js and native iOS HLS will automatically switch quality
  const roomStreamUrl = `/hls/${data.room.id}/master.m3u8`;
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

          {/* Class schedule info */}
          {data.currentClass && (
            <div className="class-info-banner active">
              <span className="class-icon">📚</span>
              <span className="class-name">
                {data.currentClass.schedule.name}
              </span>
              <span className="class-time">
                {formatEndTime(data.currentClass.instanceEnd)}
              </span>
            </div>
          )}

          {!data.currentClass && data.nextClass && !state.isStreamLive && (
            <div className="class-info-banner upcoming">
              <span className="class-icon">📅</span>
              <span className="class-label">Next class:</span>
              <span className="class-name">{data.nextClass.schedule.name}</span>
              <span className="class-time">
                {formatStartTime(data.nextClass.instanceStart)}
              </span>
            </div>
          )}

          {/* Upcoming classes schedule */}
          {data.upcomingClasses && data.upcomingClasses.length > 0 && (
            <UpcomingSchedule classes={data.upcomingClasses} />
          )}
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
          <div className="stream-player-wrapper">
            <div className="video-container" ref={state.onContainerRef}>
              <video
                ref={state.onVideoRef}
                autoPlay
                muted
                playsInline
                preload="auto"
                className="video-player"
              />

              {/* Loading overlay when stream is starting */}
              {state.isStreamLive && !state.isHlsReady && (
                <div className="stream-loading-overlay">
                  {!state.hlsReadyTimedOut ? (
                    <>
                      <div className="loading-spinner"></div>
                      <div className="loading-message">Preparing stream...</div>
                    </>
                  ) : (
                    <>
                      <div className="loading-error-icon">⚠️</div>
                      <div className="loading-message">
                        Stream is taking longer than expected to start
                      </div>
                      <div className="loading-hint">
                        Please check your streaming software and connection
                      </div>
                    </>
                  )}
                </div>
              )}

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
                isChatVisible={state.isChatVisible}
                onJumpToLive={state.jumpToLive}
                onToggleChat={state.toggleChat}
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

              {/* Stats overlay */}
              <StatsOverlay
                stats={state.streamStats}
                visible={state.statsVisible}
              />
            </div>

            {/* Chat sidebar */}
            {state.isChatVisible && (
              <ChatSidebar
                id={`chat-${data.room?.id || 0}`}
                roomId={data.room?.id || 0}
                messages={state.chatMessages}
                onSendMessage={state.sendChatMessage}
                onClose={state.toggleChat}
              />
            )}
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

import * as preact from "preact";
import * as vlens from "vlens";

// Detect iOS devices (all browsers on iOS use WebKit)
function isIOS(): boolean {
  return /iPhone|iPad|iPod/.test(navigator.userAgent);
}

type VideoControlsProps = {
  id: string;
  videoElement: HTMLVideoElement | null;
  containerElement: HTMLDivElement | null;
  isPlaying: boolean;
  isBehindLive: boolean;
  secondsBehindLive: number;
  viewerCount: number;
  visible: boolean;
  isChatVisible: boolean;
  onJumpToLive: () => void;
  onToggleChat: () => void;
  onShowControls: () => void;
  onHideControls: () => void;
};

type ControlsState = {
  isPiPSupported: boolean;
  isPiPActive: boolean;
  isFullscreen: boolean;
  isIOS: boolean;
  canUseNativeFullscreen: boolean;
  originalScrollY: number;
  wakeLock: any | null;
  volume: number;
  isMuted: boolean;
  isVolumePopupVisible: boolean;
  isMobile: boolean;
  videoClickHandler: ((e: MouseEvent) => void) | null;
  fullscreenChangeHandler: (() => void) | null;
  keydownHandler: ((e: KeyboardEvent) => void) | null;
  onControlsClick: (e: Event) => void;
  onPlayPauseClick: () => void;
  onFullscreenClick: () => void;
  onPiPClick: () => void;
  onMuteClick: () => void;
  onVolumeChange: (value: number) => void;
  toggleVolumePopup: () => void;
  enableWakeLock: () => Promise<void>;
  disableWakeLock: () => void;
  checkPiPSupport: () => void;
  checkFullscreenState: () => void;
  cleanup: () => void;
};

const useVideoControls = vlens.declareHook(
  (
    id: string,
    videoElement: HTMLVideoElement | null,
    containerElement: HTMLDivElement | null,
    onShowControls: () => void,
    onHideControls: () => void,
  ): ControlsState => {
    // Load volume preferences from localStorage
    const loadVolumePrefs = () => {
      try {
        const saved = localStorage.getItem("streamsite_volume");
        if (saved) {
          const prefs = JSON.parse(saved);
          return {
            volume: prefs.volume || 100,
            isMuted: prefs.isMuted || false,
          };
        }
      } catch (e) {
        console.warn("Failed to load volume preferences:", e);
      }
      return { volume: 100, isMuted: false };
    };

    const volumePrefs = loadVolumePrefs();

    const iosDevice = isIOS();
    const doc = document as any;
    const canFullscreen = !!(
      doc.fullscreenEnabled || doc.webkitFullscreenEnabled
    );
    const isMobileDevice = window.innerWidth <= 768;

    const state: ControlsState = {
      isPiPSupported: false,
      isPiPActive: false,
      isFullscreen: false,
      isIOS: iosDevice,
      canUseNativeFullscreen: canFullscreen && !iosDevice,
      originalScrollY: 0,
      wakeLock: null,
      volume: volumePrefs.volume,
      isMuted: volumePrefs.isMuted,
      isVolumePopupVisible: false,
      isMobile: isMobileDevice,
      videoClickHandler: null,
      fullscreenChangeHandler: null,
      keydownHandler: null,
      onControlsClick: (e: Event) => {
        // Prevent click from bubbling to video element
        e.stopPropagation();
      },
      onPlayPauseClick: () => {
        if (!videoElement) return;

        if (videoElement.paused) {
          videoElement.play().catch((e) => console.warn("Play failed:", e));
        } else {
          videoElement.pause();
          onShowControls();
        }
      },
      onFullscreenClick: () => {
        if (!containerElement) return;

        // iOS: Use CSS theater mode
        if (state.isIOS) {
          if (state.isFullscreen) {
            // Exit theater mode
            containerElement.classList.remove("video-container-fullscreen-ios");
            state.isFullscreen = false;

            // Restore scroll position without jump
            const scrollY = state.originalScrollY;
            document.body.style.position = "";
            document.body.style.top = "";
            window.scrollTo(0, scrollY);

            // Disable wake lock
            state.disableWakeLock();
          } else {
            // Enter theater mode
            // Store current scroll position
            state.originalScrollY = window.scrollY;

            // Lock scroll without jump
            document.body.style.position = "fixed";
            document.body.style.top = `-${state.originalScrollY}px`;
            document.body.style.width = "100%";

            containerElement.classList.add("video-container-fullscreen-ios");
            state.isFullscreen = true;

            // Enable wake lock
            state.enableWakeLock();
          }
          vlens.scheduleRedraw();
          onShowControls();
          return;
        }

        // Desktop: Use native Fullscreen API
        const doc = document as any;

        if (state.isFullscreen) {
          // Exit fullscreen
          if (doc.exitFullscreen) {
            doc.exitFullscreen();
          } else if (doc.webkitExitFullscreen) {
            doc.webkitExitFullscreen();
          }
        } else {
          // Enter fullscreen on container (includes video + controls)
          const elem = containerElement as any;
          if (elem.requestFullscreen) {
            elem.requestFullscreen();
          } else if (elem.webkitRequestFullscreen) {
            elem.webkitRequestFullscreen();
          }
        }
        onShowControls();
      },
      onPiPClick: async () => {
        if (!videoElement) return;
        const doc = document as any;
        const elem = videoElement as any;

        try {
          if (state.isPiPActive) {
            // Exit PiP
            if (doc.exitPictureInPicture) {
              await doc.exitPictureInPicture();
            }
          } else {
            // Enter PiP
            if (elem.requestPictureInPicture) {
              await elem.requestPictureInPicture();
            } else if (elem.webkitSetPresentationMode) {
              // iOS Safari
              elem.webkitSetPresentationMode("picture-in-picture");
            }
          }
        } catch (e) {
          console.warn("PiP failed:", e);
        }
        onShowControls();
      },
      onMuteClick: () => {
        if (!videoElement) return;

        // On mobile, toggle volume popup instead of just muting
        if (state.isMobile) {
          state.toggleVolumePopup();
          return;
        }

        state.isMuted = !state.isMuted;
        videoElement.muted = state.isMuted;

        // Save to localStorage
        try {
          localStorage.setItem(
            "streamsite_volume",
            JSON.stringify({
              volume: state.volume,
              isMuted: state.isMuted,
            }),
          );
        } catch (e) {
          console.warn("Failed to save volume preferences:", e);
        }

        vlens.scheduleRedraw();
        onShowControls();
      },
      toggleVolumePopup: () => {
        state.isVolumePopupVisible = !state.isVolumePopupVisible;
        vlens.scheduleRedraw();
        onShowControls();
      },
      onVolumeChange: (value: number) => {
        if (!videoElement) return;
        state.volume = value;
        videoElement.volume = value / 100; // Convert 0-100 to 0.0-1.0

        // Unmute if user adjusts volume
        if (state.isMuted && value > 0) {
          state.isMuted = false;
          videoElement.muted = false;
        }

        // Save to localStorage
        try {
          localStorage.setItem(
            "streamsite_volume",
            JSON.stringify({
              volume: state.volume,
              isMuted: state.isMuted,
            }),
          );
        } catch (e) {
          console.warn("Failed to save volume preferences:", e);
        }

        vlens.scheduleRedraw();
        onShowControls();
      },
      enableWakeLock: async () => {
        if (!("wakeLock" in navigator)) return;
        try {
          const nav = navigator as any;
          state.wakeLock = await nav.wakeLock.request("screen");
        } catch (e) {
          console.warn("Wake lock request failed:", e);
        }
      },
      disableWakeLock: () => {
        if (!state.wakeLock) return;
        try {
          state.wakeLock.release();
          state.wakeLock = null;
        } catch (e) {
          console.warn("Wake lock release failed:", e);
        }
      },
      checkPiPSupport: () => {
        const doc = document as any;
        const elem = videoElement as any;
        state.isPiPSupported = !!(
          doc.pictureInPictureEnabled || elem?.webkitSetPresentationMode
        );
        vlens.scheduleRedraw();
      },
      checkFullscreenState: () => {
        const doc = document as any;
        const isFS = !!(
          doc.fullscreenElement ||
          doc.webkitFullscreenElement ||
          doc.webkitCurrentFullScreenElement
        );
        if (state.isFullscreen !== isFS) {
          state.isFullscreen = isFS;
          vlens.scheduleRedraw();
        }
      },
      cleanup: () => {
        // Clean up iOS theater mode if active
        if (state.isIOS && state.isFullscreen && containerElement) {
          containerElement.classList.remove("video-container-fullscreen-ios");
          const scrollY = state.originalScrollY;
          document.body.style.position = "";
          document.body.style.top = "";
          document.body.style.width = "";
          window.scrollTo(0, scrollY);
        }

        // Release wake lock
        state.disableWakeLock();

        // Remove fullscreen change listeners
        if (state.fullscreenChangeHandler) {
          document.removeEventListener(
            "fullscreenchange",
            state.fullscreenChangeHandler,
          );
          document.removeEventListener(
            "webkitfullscreenchange",
            state.fullscreenChangeHandler,
          );
          state.fullscreenChangeHandler = null;
        }

        // Remove keydown listener
        if (state.keydownHandler) {
          document.removeEventListener("keydown", state.keydownHandler);
          state.keydownHandler = null;
        }

        // Remove video click listener
        if (videoElement && state.videoClickHandler) {
          videoElement.removeEventListener(
            "click",
            state.videoClickHandler as any,
          );
          state.videoClickHandler = null;
        }
      },
    };

    // Video click handler is set up in parent (stream.tsx)
    // Parent manages visibility state and auto-hide timer

    // Check PiP support when video element changes
    if (videoElement) {
      state.checkPiPSupport();

      // Apply volume settings from state
      videoElement.volume = state.volume / 100;
      videoElement.muted = state.isMuted;

      // Listen for PiP events
      const handleEnterPiP = () => {
        state.isPiPActive = true;
        vlens.scheduleRedraw();
      };
      const handleLeavePiP = () => {
        state.isPiPActive = false;
        vlens.scheduleRedraw();
      };

      videoElement.addEventListener("enterpictureinpicture", handleEnterPiP);
      videoElement.addEventListener("leavepictureinpicture", handleLeavePiP);

      // iOS Safari PiP events
      const handleWebkitPresentationModeChanged = () => {
        const elem = videoElement as any;
        state.isPiPActive =
          elem.webkitPresentationMode === "picture-in-picture";
        vlens.scheduleRedraw();
      };
      videoElement.addEventListener(
        "webkitpresentationmodechanged",
        handleWebkitPresentationModeChanged,
      );

      // Listen for fullscreen changes (only for desktop fullscreen API)
      if (!state.isIOS) {
        state.fullscreenChangeHandler = () => {
          state.checkFullscreenState();
        };
        document.addEventListener(
          "fullscreenchange",
          state.fullscreenChangeHandler,
        );
        document.addEventListener(
          "webkitfullscreenchange",
          state.fullscreenChangeHandler,
        );
      }

      // Handle escape key for CSS fullscreen on iOS (for iPad with keyboard)
      if (state.isIOS) {
        state.keydownHandler = (e: KeyboardEvent) => {
          if (state.isFullscreen && e.key === "Escape") {
            if (containerElement) {
              containerElement.classList.remove(
                "video-container-fullscreen-ios",
              );
              state.isFullscreen = false;
              const scrollY = state.originalScrollY;
              document.body.style.position = "";
              document.body.style.top = "";
              window.scrollTo(0, scrollY);
              state.disableWakeLock();
              vlens.scheduleRedraw();
            }
          }
        };
        document.addEventListener("keydown", state.keydownHandler);
      }
    }

    return state;
  },
);

export function VideoControls(props: VideoControlsProps) {
  const state = useVideoControls(
    props.id,
    props.videoElement,
    props.containerElement,
    props.onShowControls,
    props.onHideControls,
  );

  if (!props.videoElement) return null;

  return (
    <div
      className={`video-controls-container ${props.visible ? "visible" : "hidden"}`}
    >
      <div
        className="video-controls-overlay"
        onClick={(e) => state.onControlsClick(e)}
      >
        {/* iOS Theater Mode Exit Button */}
        {state.isIOS && state.isFullscreen && (
          <button
            className="ios-theater-exit-btn"
            onClick={state.onFullscreenClick}
            aria-label="Exit theater mode"
          >
            ✕
          </button>
        )}

        {/* Center play/pause button */}
        <div className="control-center">
          <button
            className="control-btn control-play-pause"
            onClick={state.onPlayPauseClick}
            aria-label={props.isPlaying ? "Pause" : "Play"}
          >
            {props.isPlaying ? (
              <div className="pause-icon" />
            ) : (
              <div className="play-icon" />
            )}
          </button>
        </div>

        {/* Bottom controls bar */}
        <div className="control-bar">
          {/* Viewer count badge */}
          <div className="control-viewer-badge">
            <span className="viewer-icon">👁</span>
            <span className="viewer-text">{props.viewerCount}</span>
          </div>

          {/* Live indicator / Go Live button */}
          {props.isBehindLive ? (
            <button
              className="control-btn control-go-live"
              onClick={props.onJumpToLive}
              aria-label="Jump to live"
            >
              <span className="live-icon">●</span>
              <span className="live-text">GO LIVE</span>
              {props.secondsBehindLive > 0 && props.secondsBehindLive < 60 && (
                <span className="live-time">
                  -{Math.floor(props.secondsBehindLive)}s
                </span>
              )}
            </button>
          ) : (
            <div className="control-live-badge">
              <span className="live-icon pulsing">●</span>
              <span className="live-text">LIVE</span>
            </div>
          )}

          <div className="control-spacer"></div>

          {/* Volume control container */}
          <div style={{ position: "relative" }}>
            {/* Mute/unmute button */}
            <button
              className="control-btn control-mute"
              onClick={state.onMuteClick}
              aria-label={state.isMuted ? "Unmute" : "Mute"}
            >
              {state.isMuted ? "🔇" : "🔊"}
            </button>

            {/* Volume popup overlay (mobile only) */}
            {state.isMobile && state.isVolumePopupVisible && (
              <div className="volume-popup-overlay">
                <span className="volume-popup-label">{state.volume}%</span>
                <input
                  type="range"
                  className="volume-popup-slider"
                  min="0"
                  max="100"
                  value={state.volume}
                  onInput={(e) =>
                    state.onVolumeChange(parseInt(e.currentTarget.value))
                  }
                  aria-label="Volume"
                />
              </div>
            )}
          </div>

          {/* Volume slider (desktop only) */}
          <input
            type="range"
            className="volume-slider"
            min="0"
            max="100"
            value={state.volume}
            onInput={(e) =>
              state.onVolumeChange(parseInt(e.currentTarget.value))
            }
            aria-label="Volume"
          />

          {/* Chat toggle button (mobile only) */}
          <button
            className="control-btn control-chat mobile-only"
            onClick={props.onToggleChat}
            aria-label={props.isChatVisible ? "Hide chat" : "Show chat"}
          >
            💬
          </button>

          {/* Picture-in-Picture button (only show if supported) */}
          {state.isPiPSupported && (
            <button
              className={`control-btn control-pip ${state.isPiPActive ? "active" : ""}`}
              onClick={state.onPiPClick}
              aria-label={
                state.isPiPActive
                  ? "Exit picture-in-picture"
                  : "Picture-in-picture"
              }
            >
              ⧉
            </button>
          )}

          {/* Fullscreen button */}
          <button
            className="control-btn control-fullscreen"
            onClick={state.onFullscreenClick}
            aria-label={state.isFullscreen ? "Exit fullscreen" : "Fullscreen"}
          >
            {state.isFullscreen ? "⛶" : "⛶"}
          </button>
        </div>
      </div>
    </div>
  );
}

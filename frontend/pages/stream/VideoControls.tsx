import * as preact from "preact";
import * as vlens from "vlens";

type VideoControlsProps = {
  id: string;
  videoElement: HTMLVideoElement | null;
  containerElement: HTMLDivElement | null;
  isPlaying: boolean;
  isBehindLive: boolean;
  secondsBehindLive: number;
  viewerCount: number;
  visible: boolean;
  onJumpToLive: () => void;
  onShowControls: () => void;
  onHideControls: () => void;
};

type ControlsState = {
  isPiPSupported: boolean;
  isPiPActive: boolean;
  isFullscreen: boolean;
  volume: number;
  isMuted: boolean;
  videoClickHandler: ((e: MouseEvent) => void) | null;
  onControlsClick: (e: Event) => void;
  onPlayPauseClick: () => void;
  onFullscreenClick: () => void;
  onPiPClick: () => void;
  onMuteClick: () => void;
  onVolumeChange: (value: number) => void;
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

    const state: ControlsState = {
      isPiPSupported: false,
      isPiPActive: false,
      isFullscreen: false,
      volume: volumePrefs.volume,
      isMuted: volumePrefs.isMuted,
      videoClickHandler: null,
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
        }
        onShowControls();
      },
      onFullscreenClick: () => {
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
          if (containerElement) {
            const elem = containerElement as any;
            if (elem.requestFullscreen) {
              elem.requestFullscreen();
            } else if (elem.webkitRequestFullscreen) {
              elem.webkitRequestFullscreen();
            }
          } else if (videoElement) {
            // Fallback to video-only fullscreen if container not available
            const elem = videoElement as any;
            if (elem.requestFullscreen) {
              elem.requestFullscreen();
            } else if (elem.webkitRequestFullscreen) {
              elem.webkitRequestFullscreen();
            } else if (elem.webkitEnterFullscreen) {
              // iOS Safari
              elem.webkitEnterFullscreen();
            }
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

      // Listen for fullscreen changes
      const handleFullscreenChange = () => {
        state.checkFullscreenState();
      };
      document.addEventListener("fullscreenchange", handleFullscreenChange);
      document.addEventListener(
        "webkitfullscreenchange",
        handleFullscreenChange,
      );
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
        {/* Center play/pause button */}
        <div className="control-center">
          <button
            className="control-btn control-play-pause"
            onClick={state.onPlayPauseClick}
            aria-label={props.isPlaying ? "Pause" : "Play"}
          >
            {props.isPlaying ? "⏸" : "▶"}
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

          {/* Mute/unmute button */}
          <button
            className="control-btn control-mute"
            onClick={state.onMuteClick}
            aria-label={state.isMuted ? "Unmute" : "Mute"}
          >
            {state.isMuted ? "🔇" : "🔊"}
          </button>

          {/* Volume slider */}
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

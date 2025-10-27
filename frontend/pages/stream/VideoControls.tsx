import * as preact from "preact";
import * as vlens from "vlens";

type VideoControlsProps = {
  id: string;
  videoElement: HTMLVideoElement | null;
  isPlaying: boolean;
  isBehindLive: boolean;
  secondsBehindLive: number;
  onJumpToLive: () => void;
};

type ControlsState = {
  visible: boolean;
  isPiPSupported: boolean;
  isPiPActive: boolean;
  isFullscreen: boolean;
  autoHideTimer: number | null;
  videoClickHandler: ((e: MouseEvent) => void) | null;
  onControlsClick: (e: Event) => void;
  onPlayPauseClick: () => void;
  onFullscreenClick: () => void;
  onPiPClick: () => void;
  showControls: () => void;
  hideControls: () => void;
  checkPiPSupport: () => void;
  checkFullscreenState: () => void;
  cleanup: () => void;
};

const useVideoControls = vlens.declareHook(
  (id: string, videoElement: HTMLVideoElement | null): ControlsState => {
    const state: ControlsState = {
      visible: true,
      isPiPSupported: false,
      isPiPActive: false,
      isFullscreen: false,
      autoHideTimer: null,
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
        state.showControls();
      },
      onFullscreenClick: () => {
        if (!videoElement) return;

        const doc = document as any;
        const elem = videoElement as any;

        if (state.isFullscreen) {
          // Exit fullscreen
          if (doc.exitFullscreen) {
            doc.exitFullscreen();
          } else if (doc.webkitExitFullscreen) {
            doc.webkitExitFullscreen();
          } else if (elem.webkitExitFullscreen) {
            elem.webkitExitFullscreen();
          }
        } else {
          // Enter fullscreen
          if (elem.requestFullscreen) {
            elem.requestFullscreen();
          } else if (elem.webkitRequestFullscreen) {
            elem.webkitRequestFullscreen();
          } else if (elem.webkitEnterFullscreen) {
            // iOS Safari
            elem.webkitEnterFullscreen();
          }
        }
        state.showControls();
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
        state.showControls();
      },
      showControls: () => {
        state.visible = true;
        vlens.scheduleRedraw();

        // Clear existing timer
        if (state.autoHideTimer !== null) {
          clearTimeout(state.autoHideTimer);
        }

        // Auto-hide after 3 seconds
        state.autoHideTimer = window.setTimeout(() => {
          state.hideControls();
        }, 3000);
      },
      hideControls: () => {
        state.visible = false;
        vlens.scheduleRedraw();
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
        // Clear auto-hide timer
        if (state.autoHideTimer !== null) {
          clearTimeout(state.autoHideTimer);
          state.autoHideTimer = null;
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

    // Set up video click handler to toggle controls
    const handleVideoClick = (e: MouseEvent) => {
      // Only toggle if clicking directly on video, not controls
      if ((e.target as HTMLElement).tagName === "VIDEO") {
        if (state.visible) {
          state.hideControls();
        } else {
          state.showControls();
        }
      }
    };

    state.videoClickHandler = handleVideoClick;

    // Check PiP support when video element changes
    if (videoElement) {
      state.checkPiPSupport();

      // Attach click handler to video element
      videoElement.addEventListener("click", state.videoClickHandler as any);

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
  const state = useVideoControls(props.id, props.videoElement);

  if (!props.videoElement) return null;

  return (
    <div
      className={`video-controls-container ${state.visible ? "visible" : "hidden"}`}
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

          {/* Fullscreen button */}
          <button
            className="control-btn control-fullscreen"
            onClick={state.onFullscreenClick}
            aria-label={state.isFullscreen ? "Exit fullscreen" : "Fullscreen"}
          >
            {state.isFullscreen ? "⛶" : "⛶"}
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
        </div>
      </div>
    </div>
  );
}

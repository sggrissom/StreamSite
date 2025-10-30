import * as preact from "preact";
import * as vlens from "vlens";

type EmotePickerProps = {
  id: string; // Required for vlens caching
  controlsVisible: boolean; // Show/hide with video controls
  onLocalEmote: (emote: string) => void; // Immediate local display
  onSendEmote: (emote: string) => void; // Broadcast to all viewers
};

// Allowed emotes matching backend
const EMOTES = ["❤️", "👏", "🔥", "😂", "😮", "👍"];

type EmotePickerState = {
  // Track if any emote was recently clicked (for visual feedback)
  lastClickedEmote: string | null;

  // Cooldown tracking
  lastSentTime: number; // Timestamp of last sent emote
  cooldownEndsAt: number; // Timestamp when cooldown ends
  queuedEmote: string | null; // Emote queued during cooldown
  cooldownTimer: number | null; // Interval timer for cooldown updates
  cooldownProgress: number; // 0-100, percentage of cooldown remaining

  onEmoteClick: (emote: string) => void;
  updateCooldown: () => void;
  clearCooldownTimer: () => void;
};

const COOLDOWN_DURATION_MS = 2000; // 2 seconds

const useEmotePicker = vlens.declareHook(
  (
    id: string,
    onLocalEmote: (emote: string) => void,
    onSendEmote: (emote: string) => void,
  ): EmotePickerState => {
    const state: EmotePickerState = {
      lastClickedEmote: null,
      lastSentTime: 0,
      cooldownEndsAt: 0,
      queuedEmote: null,
      cooldownTimer: null,
      cooldownProgress: 0,

      updateCooldown: () => {
        const now = Date.now();
        const remaining = Math.max(0, state.cooldownEndsAt - now);

        if (remaining === 0) {
          // Cooldown expired
          state.cooldownProgress = 0;
          state.clearCooldownTimer();

          // Send queued emote if any
          if (state.queuedEmote) {
            const emoteToSend = state.queuedEmote;
            state.queuedEmote = null;

            // Send to server (broadcast)
            onSendEmote(emoteToSend);

            // Update cooldown timing
            state.lastSentTime = now;
            state.cooldownEndsAt = now + COOLDOWN_DURATION_MS;

            // Restart cooldown timer
            state.cooldownTimer = window.setInterval(state.updateCooldown, 100);
          }
        } else {
          // Update progress (0-100)
          state.cooldownProgress = (remaining / COOLDOWN_DURATION_MS) * 100;
        }

        vlens.scheduleRedraw();
      },

      clearCooldownTimer: () => {
        if (state.cooldownTimer !== null) {
          clearInterval(state.cooldownTimer);
          state.cooldownTimer = null;
        }
      },

      onEmoteClick: (emote: string) => {
        const now = Date.now();
        const timeSinceLastSend = now - state.lastSentTime;

        // Visual feedback for click
        state.lastClickedEmote = emote;
        setTimeout(() => {
          state.lastClickedEmote = null;
          vlens.scheduleRedraw();
        }, 300);

        // Always show locally immediately
        onLocalEmote(emote);

        // Check if we can send to server
        if (timeSinceLastSend >= COOLDOWN_DURATION_MS) {
          // Not in cooldown - send immediately
          onSendEmote(emote);

          // Start cooldown
          state.lastSentTime = now;
          state.cooldownEndsAt = now + COOLDOWN_DURATION_MS;
          state.queuedEmote = null;

          // Start cooldown timer
          state.clearCooldownTimer();
          state.cooldownTimer = window.setInterval(state.updateCooldown, 100);
          state.updateCooldown(); // Update immediately
        } else {
          // In cooldown - queue this emote
          state.queuedEmote = emote;
        }

        vlens.scheduleRedraw();
      },
    };

    return state;
  },
);

export function EmotePicker(props: EmotePickerProps) {
  const state = useEmotePicker(props.id, props.onLocalEmote, props.onSendEmote);

  const isInCooldown = state.cooldownProgress > 0;

  return (
    <div
      className={`emote-picker ${props.controlsVisible ? "visible" : "hidden"}`}
    >
      {/* Cooldown indicator */}
      {isInCooldown && (
        <div className="emote-cooldown-bar">
          <div
            className="emote-cooldown-progress"
            style={{ width: `${state.cooldownProgress}%` }}
          />
        </div>
      )}

      <div className="emote-buttons">
        {EMOTES.map((emote) => (
          <button
            key={emote}
            className={`emote-btn ${state.lastClickedEmote === emote ? "emote-btn-clicked" : ""} ${isInCooldown ? "emote-btn-cooldown" : ""}`}
            onClick={() => state.onEmoteClick(emote)}
            title={
              isInCooldown
                ? `${emote} (queued for broadcast)`
                : `Send ${emote} reaction`
            }
          >
            {emote}
          </button>
        ))}
      </div>

      {/* Queue indicator */}
      {state.queuedEmote && (
        <div className="emote-queue-indicator">Queued: {state.queuedEmote}</div>
      )}
    </div>
  );
}

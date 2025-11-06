import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";

type ChatMessage = server.ChatMessage;

type ChatSidebarProps = {
  id: string;
  roomId: number;
  messages: ChatMessage[];
  onSendMessage: (text: string) => void;
  onClose?: () => void; // For mobile
};

type ChatState = {
  messageText: string;
  scrollContainerRef: HTMLDivElement | null;
  onScrollContainerRef: (el: HTMLDivElement | null) => void;
  isAtBottom: boolean;
  shouldAutoScroll: boolean;
  onScroll: (e: Event) => void;
  handleSend: () => void;
  handleKeyDown: (
    e: KeyboardEvent,
    onSendMessage: (text: string) => void,
  ) => void;
};

const useChatState = vlens.declareHook((): ChatState => {
  const state: ChatState = {
    messageText: "",
    scrollContainerRef: null,
    isAtBottom: true,
    shouldAutoScroll: true,

    onScrollContainerRef: (el: HTMLDivElement | null) => {
      state.scrollContainerRef = el;
      if (el) {
        // Auto-scroll to bottom on mount
        setTimeout(() => {
          el.scrollTop = el.scrollHeight;
        }, 0);
      }
    },

    onScroll: (e: Event) => {
      const container = e.target as HTMLDivElement;
      const threshold = 50; // pixels from bottom
      const distanceFromBottom =
        container.scrollHeight - container.scrollTop - container.clientHeight;
      state.isAtBottom = distanceFromBottom < threshold;
      state.shouldAutoScroll = state.isAtBottom;
    },

    handleSend: () => {
      // Just clear the input and scroll
      state.messageText = "";
      vlens.scheduleRedraw();

      // Scroll to bottom after sending
      state.shouldAutoScroll = true;
      if (state.scrollContainerRef) {
        setTimeout(() => {
          if (state.scrollContainerRef) {
            state.scrollContainerRef.scrollTop =
              state.scrollContainerRef.scrollHeight;
          }
        }, 50);
      }
    },

    handleKeyDown: (
      e: KeyboardEvent,
      onSendMessage: (text: string) => void,
    ) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        const text = state.messageText.trim();
        if (text.length > 0 && text.length <= 500) {
          onSendMessage(text);
          state.handleSend();
        }
      }
    },
  };

  return state;
});

// Helper function to format timestamp
function formatTimestamp(timestamp: string | Date): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`;
  return date.toLocaleDateString();
}

export function ChatSidebar(props: ChatSidebarProps) {
  const state = useChatState();

  // Auto-scroll to bottom when new messages arrive (if user was already at bottom)
  if (state.scrollContainerRef && state.shouldAutoScroll) {
    const container = state.scrollContainerRef;
    // Use requestAnimationFrame to ensure DOM has updated
    requestAnimationFrame(() => {
      container.scrollTop = container.scrollHeight;
    });
  }

  const charCount = state.messageText.length;
  const charLimit = 500;
  const isOverLimit = charCount > charLimit;

  return (
    <div className="chat-sidebar">
      {/* Header */}
      <div className="chat-header">
        <span className="chat-title">Chat</span>
        {props.onClose && (
          <button
            className="chat-close-btn"
            onClick={props.onClose}
            aria-label="Close chat"
          >
            ✕
          </button>
        )}
      </div>

      {/* Messages */}
      <div
        className="chat-messages"
        ref={state.onScrollContainerRef}
        onScroll={state.onScroll}
      >
        {props.messages.length === 0 ? (
          <div className="chat-empty">
            <p>No messages yet</p>
            <p className="chat-empty-hint">Be the first to say something!</p>
          </div>
        ) : (
          props.messages.map((msg) => (
            <div key={msg.id} className="chat-message">
              <div className="chat-message-header">
                <span className="chat-username">{msg.userName}</span>
                <span className="chat-timestamp">
                  {formatTimestamp(msg.timestamp)}
                </span>
              </div>
              <div className="chat-message-text">{msg.text}</div>
            </div>
          ))
        )}
      </div>

      {/* Input */}
      <div className="chat-input-container">
        <div className="chat-input-wrapper">
          <textarea
            className="chat-input"
            placeholder="Send a message..."
            maxLength={charLimit}
            rows={2}
            {...vlens.attrsBindInput(vlens.ref(state, "messageText"))}
            onKeyDown={(e) => state.handleKeyDown(e, props.onSendMessage)}
          />
          <div className={`chat-char-count ${isOverLimit ? "over-limit" : ""}`}>
            {charCount}/{charLimit}
          </div>
        </div>
        <button
          className="chat-send-btn"
          onClick={() => {
            const text = state.messageText.trim();
            if (text.length > 0 && text.length <= 500) {
              props.onSendMessage(text);
              state.handleSend();
            }
          }}
          disabled={state.messageText.trim().length === 0 || isOverLimit}
        >
          Send
        </button>
      </div>
    </div>
  );
}

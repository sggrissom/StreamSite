import { block } from "vlens/css";

// Chat sidebar container
block(`
.chat-sidebar {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  height: 100%;
  overflow: hidden;
}
`);

// Header
block(`
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border);
  background: var(--background);
}
`);

block(`
.chat-title {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text);
}
`);

block(`
.chat-close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 0;
  width: 32px;
  height: 32px;
  display: none;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}
`);

block(`
.chat-close-btn:hover {
  background: var(--border);
}
`);

// Messages container
block(`
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
`);

// Individual message
block(`
.chat-message {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
`);

block(`
.chat-message-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
}
`);

block(`
.chat-username {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text);
}
`);

block(`
.chat-timestamp {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}
`);

block(`
.chat-message-text {
  font-size: 0.875rem;
  color: var(--text);
  word-wrap: break-word;
  white-space: pre-wrap;
  line-height: 1.4;
}
`);

// Empty state
block(`
.chat-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 0.5rem;
  color: var(--text-secondary);
  text-align: center;
}
`);

block(`
.chat-empty p {
  margin: 0;
}
`);

block(`
.chat-empty-hint {
  font-size: 0.875rem;
  color: var(--text-tertiary);
}
`);

// Input container
block(`
.chat-input-container {
  border-top: 1px solid var(--border);
  padding: 1rem;
  background: var(--background);
  display: flex;
  gap: 0.75rem;
}
`);

block(`
.chat-input-wrapper {
  flex: 1;
  position: relative;
}
`);

block(`
.chat-input {
  width: 100%;
  padding: 0.75rem;
  padding-bottom: 2rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  font-size: 0.875rem;
  font-family: inherit;
  resize: none;
  background: var(--surface);
  color: var(--text);
  transition: border-color 0.2s;
}
`);

block(`
.chat-input:focus {
  outline: none;
  border-color: #3b82f6;
}
`);

block(`
.chat-char-count {
  position: absolute;
  bottom: 0.5rem;
  right: 0.75rem;
  font-size: 0.75rem;
  color: var(--text-tertiary);
}
`);

block(`
.chat-char-count.over-limit {
  color: #ef4444;
  font-weight: 600;
}
`);

block(`
.chat-send-btn {
  padding: 0.75rem 1.5rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
  align-self: flex-end;
}
`);

block(`
.chat-send-btn:hover:not(:disabled) {
  background: #2563eb;
}
`);

block(`
.chat-send-btn:disabled {
  background: var(--border);
  cursor: not-allowed;
  opacity: 0.5;
}
`);

block(`
.chat-readonly-notice {
  flex: 1;
  padding: 0.75rem;
  text-align: center;
  font-size: 0.875rem;
  color: var(--text-secondary);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 4px;
}
`);

// Scrollbar styling
block(`
.chat-messages::-webkit-scrollbar {
  width: 8px;
}
`);

block(`
.chat-messages::-webkit-scrollbar-track {
  background: var(--background);
}
`);

block(`
.chat-messages::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 4px;
}
`);

block(`
.chat-messages::-webkit-scrollbar-thumb:hover {
  background: var(--text-tertiary);
}
`);

// Mobile responsive styles
block(`
@media (max-width: 768px) {
  .chat-sidebar {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    width: 100%;
    height: 60vh;
    border-radius: 16px 16px 0 0;
    z-index: 100;
    box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  }

  .chat-close-btn {
    display: flex;
  }

  .chat-input {
    font-size: 16px;
  }
}
`);

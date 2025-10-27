import { block } from "vlens/css";

block(`
.stream-container {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
`);

block(`
.stream-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 2rem;
}
`);

block(`
.stream-context {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}
`);

block(`
.context-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1.5rem;
}
`);

block(`
.context-info {
  flex: 1;
}
`);

block(`
.context-studio {
  font-size: 0.875rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin: 0 0 0.5rem 0;
}
`);

block(`
.context-room {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}
`);

block(`
.context-meta {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}
`);

block(`
.role-badge {
  display: inline-block;
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: var(--background);
  border: 1px solid var(--border);
  color: var(--text-secondary);
}
`);

block(`
.role-badge.role-0 {
  background: #e3f2fd;
  border-color: #90caf9;
  color: #1565c0;
}
`);

block(`
.role-badge.role-1 {
  background: #e8f5e9;
  border-color: #81c784;
  color: #2e7d32;
}
`);

block(`
.role-badge.role-2 {
  background: #fff3e0;
  border-color: #ffb74d;
  color: #e65100;
}
`);

block(`
.role-badge.role-3 {
  background: #f3e5f5;
  border-color: #ba68c8;
  color: #6a1b9a;
}
`);

block(`
.room-number {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  padding: 0.25rem 0.75rem;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: 12px;
}
`);

block(`
.code-session-banner {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 8px;
  padding: 1rem 1.5rem;
  margin-top: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}
`);

block(`
.code-session-banner .banner-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}
`);

block(`
.code-session-banner .banner-text {
  font-size: 0.9375rem;
  font-weight: 600;
  color: white;
  letter-spacing: 0.01em;
}
`);

block(`
.code-session-banner .banner-countdown {
  margin-left: auto;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  background: rgba(255, 255, 255, 0.15);
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  white-space: nowrap;
}
`);

block(`
.code-session-banner.grace-period {
  background: linear-gradient(135deg, #ff9800 0%, #f57c00 100%);
  border-color: rgba(255, 152, 0, 0.3);
  box-shadow: 0 2px 8px rgba(255, 152, 0, 0.15);
}
`);

block(`
.revoked-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
`);

block(`
.revoked-modal {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 2rem;
  max-width: 500px;
  width: 90%;
  text-align: center;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}
`);

block(`
.revoked-modal h2 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 1rem 0;
}
`);

block(`
.revoked-modal p {
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0 0 2rem 0;
}
`);

block(`
.video-container {
  margin-top: 2rem;
  position: relative;
}
`);

block(`
.video-player {
  width: 100%;
  max-width: 1200px;
  background-color: #000;
  border-radius: 8px;
}
`);

block(`
.video-controls-container {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;
  z-index: 10;
}
`);

block(`
.video-controls-container.hidden {
  opacity: 0;
}
`);

block(`
.video-controls-container.visible {
  opacity: 1;
}
`);

block(`
.video-controls-overlay {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  background: linear-gradient(
    to bottom,
    rgba(0, 0, 0, 0.3) 0%,
    transparent 20%,
    transparent 80%,
    rgba(0, 0, 0, 0.5) 100%
  );
  pointer-events: none;
}
`);

block(`
.video-controls-container.visible .video-controls-overlay {
  pointer-events: auto;
}
`);

block(`
.control-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
`);

block(`
.control-btn {
  background: rgba(0, 0, 0, 0.6);
  border: 2px solid rgba(255, 255, 255, 0.8);
  color: white;
  cursor: pointer;
  font-size: 1.5rem;
  padding: 0;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}
`);

block(`
.control-btn:hover {
  background: rgba(0, 0, 0, 0.8);
  border-color: white;
  transform: scale(1.05);
}
`);

block(`
.control-btn:active {
  transform: scale(0.95);
}
`);

block(`
.control-play-pause {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  font-size: 2rem;
}
`);

block(`
.control-bar {
  display: flex;
  gap: 0.75rem;
  padding: 1rem;
  justify-content: flex-end;
  align-items: center;
}
`);

block(`
.control-fullscreen,
.control-pip {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  font-size: 1.25rem;
}
`);

block(`
.control-pip.active {
  background: rgba(var(--accent), 0.8);
  border-color: var(--accent);
}
`);

block(`
.orientation-hint {
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
  border: 1px solid rgba(255, 107, 107, 0.3);
  border-radius: 8px;
  padding: 0.875rem 1.25rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 2px 8px rgba(255, 107, 107, 0.2);
}
`);

block(`
.orientation-hint .hint-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}
`);

block(`
.orientation-hint .hint-text {
  font-size: 0.9375rem;
  font-weight: 500;
  color: white;
  line-height: 1.4;
}
`);

block(`
.stream-actions {
  margin-top: 1.5rem;
  display: flex;
  gap: 1rem;
}
`);

block(`
.stream-offline {
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 12px;
  padding: 4rem 2rem;
  text-align: center;
  margin-top: 2rem;
}
`);

block(`
.offline-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}
`);

block(`
.stream-offline h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 1rem 0;
}
`);

block(`
.stream-offline p {
  color: var(--text-secondary);
  font-size: 1rem;
  margin: 0.5rem 0;
  line-height: 1.5;
}
`);

block(`
.offline-hint {
  font-size: 0.875rem;
  font-style: italic;
  color: var(--text-tertiary);
  margin-top: 1.5rem !important;
}
`);

block(`
@media (max-width: 768px) {
  .stream-container {
    padding: 1rem;
  }

  .stream-title {
    font-size: 1.5rem;
  }

  .context-header {
    flex-direction: column;
  }

  .context-room {
    font-size: 1.5rem;
  }

  .context-meta {
    flex-direction: row;
    align-items: center;
  }

  .stream-actions {
    flex-direction: column;
  }

  .stream-offline {
    padding: 3rem 1.5rem;
  }

  .offline-icon {
    font-size: 3rem;
  }

  .stream-offline h2 {
    font-size: 1.25rem;
  }

  .code-session-banner {
    padding: 0.875rem 1rem;
    margin-top: 1rem;
    flex-wrap: wrap;
  }

  .code-session-banner .banner-text {
    font-size: 0.875rem;
  }

  .code-session-banner .banner-countdown {
    font-size: 0.8125rem;
    padding: 0.25rem 0.5rem;
  }

  .control-play-pause {
    width: 70px;
    height: 70px;
    font-size: 1.75rem;
  }

  .control-fullscreen,
  .control-pip {
    width: 48px;
    height: 48px;
    font-size: 1.375rem;
  }

  .control-bar {
    padding: 0.875rem;
  }

  .orientation-hint {
    padding: 0.75rem 1rem;
    margin-bottom: 0.75rem;
  }

  .orientation-hint .hint-icon {
    font-size: 1.25rem;
  }

  .orientation-hint .hint-text {
    font-size: 0.875rem;
  }
}
`);

block(`
@media (orientation: landscape) and (max-width: 896px) {
  .video-container {
    margin-top: 0;
  }

  .video-player {
    border-radius: 0;
    max-height: 100vh;
  }

  .control-bar {
    padding: 0.5rem 1rem;
  }
}
`);

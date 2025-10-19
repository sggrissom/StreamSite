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
.video-container {
  margin-top: 2rem;
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
  }

  .code-session-banner .banner-text {
    font-size: 0.875rem;
  }
}
`);

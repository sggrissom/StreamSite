import { block } from "vlens/css";

block(`
.dashboard-container {
  min-height: calc(100vh - 200px);
  padding: 40px 20px;
}
`);

block(`
.expiration-banner {
  background: linear-gradient(135deg, var(--primary-accent), var(--accent));
  color: white;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  margin-bottom: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  max-width: 1200px;
  margin-left: auto;
  margin-right: auto;
}
`);

block(`
.expiration-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 0.95rem;
  font-weight: 500;
}
`);

block(`
.expiration-icon {
  font-size: 1.25rem;
}
`);

block(`
.dashboard-content {
  text-align: center;
  max-width: 1200px;
  margin: 0 auto;
}
`);

block(`
.dashboard-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 1rem;
}
`);

block(`
.dashboard-description {
  font-size: 1.1rem;
  color: var(--muted);
  margin: 0 0 2rem;
  line-height: 1.6;
}
`);

block(`
.dashboard-actions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}
`);

block(`
.dashboard-actions .btn {
  width: 100%;
  max-width: 280px;
}
`);

block(`
.stream-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  letter-spacing: 0.05em;
  margin-bottom: 1.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
}
`);

block(`
.stream-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
`);

block(`
.stream-status-dot.status-live {
  background: #ef4444;
  animation: pulse-dot 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
`);

block(`
.stream-status-dot.status-offline {
  background: #9ca3af;
}
`);

block(`
.stream-status-text.status-live {
  color: #dc2626;
}
`);

block(`
.stream-status-text.status-offline {
  color: var(--muted);
}
`);

block(`
@keyframes pulse-dot {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
`);

block(`
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
`);

block(`
.dashboard-rooms-section {
  margin-top: 3rem;
  text-align: left;
}
`);

block(`
.dashboard-rooms-section .section-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 1.5rem;
  text-align: center;
}
`);

block(`
.dashboard-rooms-section .live-count {
  color: #dc2626;
  font-size: 1.25rem;
}
`);

block(`
.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-top: 1.5rem;
}
`);

block(`
.room-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.5rem;
  transition: all 0.2s;
}
`);

block(`
.room-card:hover {
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
`);

block(`
.room-card.room-live {
  border-color: #dc2626;
  background: linear-gradient(135deg, var(--surface) 0%, rgba(220, 38, 38, 0.05) 100%);
}
`);

block(`
.room-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
`);

block(`
.room-studio {
  font-size: 0.875rem;
  color: var(--muted);
  font-weight: 500;
}
`);

block(`
.room-status.active {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.05em;
}
`);

block(`
.room-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 1rem;
}
`);

block(`
.room-meta {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  color: var(--muted);
  font-size: 0.875rem;
}
`);

block(`
.room-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
`);

block(`
.rooms-empty {
  text-align: center;
  padding: 3rem 1rem;
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 12px;
  margin-top: 1.5rem;
}
`);

block(`
.rooms-empty .empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}
`);

block(`
.rooms-empty h3 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.5rem;
}
`);

block(`
.rooms-empty p {
  color: var(--muted);
  margin: 0 0 1.5rem;
}
`);

block(`
@media (max-width: 768px) {
  .dashboard-container {
    padding: 20px 16px;
    align-items: flex-start;
  }

  .dashboard-content {
    max-width: 100%;
  }

  .dashboard-title {
    font-size: 2rem;
  }

  .dashboard-description {
    font-size: 1rem;
  }

  .rooms-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-rooms-section .section-title {
    font-size: 1.5rem;
  }
}
`);

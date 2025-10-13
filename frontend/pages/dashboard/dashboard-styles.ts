import { block } from "vlens/css";

block(`
.dashboard-container {
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}
`);

block(`
.dashboard-content {
  text-align: center;
  max-width: 600px;
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
@media (max-width: 768px) {
  .dashboard-container {
    padding: 20px 16px;
  }

  .dashboard-title {
    font-size: 2rem;
  }

  .dashboard-description {
    font-size: 1rem;
  }
}
`);

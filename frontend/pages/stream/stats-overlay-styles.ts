import { block } from "vlens/css";

block(`
.stats-overlay {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(10px);
  border-radius: 8px;
  padding: 1rem;
  min-width: 280px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.875rem;
  color: #fff;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
}
`);

block(`
.stats-overlay-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}
`);

block(`
.stats-overlay-title {
  font-weight: 600;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #fff;
}
`);

block(`
.stats-overlay-hint {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.5);
  font-style: italic;
}
`);

block(`
.stats-overlay-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
`);

block(`
.stats-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
`);

block(`
.stats-label {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.8125rem;
}
`);

block(`
.stats-value {
  color: #fff;
  font-weight: 500;
  text-align: right;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
`);

block(`
.stats-badge {
  display: inline-block;
  background: #4caf50;
  color: #fff;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
`);

block(`
@media (max-width: 768px) {
  .stats-overlay {
    top: 0.5rem;
    right: 0.5rem;
    min-width: 240px;
    font-size: 0.8125rem;
    padding: 0.75rem;
  }

  .stats-row {
    gap: 0.75rem;
  }

  .stats-label {
    font-size: 0.75rem;
  }

  .stats-value {
    font-size: 0.8125rem;
  }
}
`);

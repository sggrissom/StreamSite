import { block } from "vlens/css";

// Modal content
block(`
.execution-logs-modal {
  padding: 1.5rem;
  max-width: 900px;
  min-width: 700px;
}
`);

// Error and loading messages
block(`
.execution-logs-modal .error-message {
  padding: 0.75rem 1rem;
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 4px;
  color: var(--danger, #dc3545);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}
`);

block(`
.execution-logs-modal .loading-message {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary, #666);
  font-size: 0.875rem;
}
`);

block(`
.execution-logs-modal .empty-message {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary, #666);
  font-size: 0.875rem;
  background: var(--bg-secondary, #f8f9fa);
  border-radius: 4px;
  line-height: 1.6;
}
`);

// Logs header
block(`
.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
}
`);

block(`
.logs-count {
  font-size: 0.875rem;
  color: var(--text-secondary, #666);
  font-weight: 500;
}
`);

// Table layout
block(`
.logs-table {
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  overflow: hidden;
  font-size: 0.875rem;
}
`);

block(`
.logs-table-header {
  display: grid;
  grid-template-columns: 160px 160px 120px 1fr;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--bg-secondary, #f8f9fa);
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  border-bottom: 2px solid var(--border-color, #e0e0e0);
}
`);

block(`
.logs-table-body {
  max-height: 500px;
  overflow-y: auto;
}
`);

block(`
.log-row {
  display: grid;
  grid-template-columns: 160px 160px 120px 1fr;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  transition: background 0.2s ease;
}
`);

block(`
.log-row:hover {
  background: var(--bg-hover, #f8f9fa);
}
`);

block(`
.log-row:last-child {
  border-bottom: none;
}
`);

block(`
.log-row.error {
  background: #fff5f5;
}
`);

block(`
.log-row.error:hover {
  background: #ffe8e8;
}
`);

// Column styles
block(`
.col-timestamp {
  color: var(--text-secondary, #666);
  font-size: 0.8125rem;
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
}
`);

block(`
.col-action {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
`);

block(`
.action-icon {
  font-size: 1rem;
  line-height: 1;
}
`);

block(`
.action-label {
  color: var(--text-primary, #1a1a1a);
  font-weight: 500;
}
`);

block(`
.col-status {
  display: flex;
  align-items: center;
}
`);

block(`
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.8125rem;
  font-weight: 500;
  line-height: 1;
}
`);

block(`
.status-badge.success {
  background: #d4edda;
  color: #155724;
}
`);

block(`
.status-badge.error {
  background: #f8d7da;
  color: #721c24;
}
`);

block(`
.col-error {
  color: var(--text-secondary, #666);
  font-size: 0.8125rem;
  line-height: 1.4;
  word-break: break-word;
}
`);

block(`
.log-row.error .col-error {
  color: var(--danger, #dc3545);
  font-weight: 500;
}
`);

// Pagination
block(`
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color, #e0e0e0);
}
`);

block(`
.pagination-btn {
  padding: 0.5rem 1rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
`);

block(`
.pagination-btn:hover:not(:disabled) {
  background: var(--bg-hover, #f8f9fa);
  border-color: var(--text-secondary, #666);
}
`);

block(`
.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  color: var(--text-secondary, #666);
}
`);

block(`
.pagination-info {
  font-size: 0.875rem;
  color: var(--text-secondary, #666);
  font-weight: 500;
}
`);

// Responsive design
block(`
@media (max-width: 800px) {
  .execution-logs-modal {
    min-width: unset;
    padding: 1rem;
  }

  .logs-table-header,
  .log-row {
    grid-template-columns: 120px 140px 100px 1fr;
    gap: 0.5rem;
    padding: 0.625rem 0.75rem;
    font-size: 0.8125rem;
  }

  .col-timestamp {
    font-size: 0.75rem;
  }

  .action-label {
    font-size: 0.8125rem;
  }

  .pagination {
    flex-direction: column;
    gap: 0.75rem;
  }

  .pagination-info {
    order: -1;
  }

  .pagination-btn {
    width: 100%;
  }
}
`);

block(`
@media (max-width: 600px) {
  .logs-table-header,
  .log-row {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }

  .logs-table-header {
    display: none;
  }

  .log-row {
    padding: 1rem;
  }

  .col-timestamp::before {
    content: '🕐 ';
  }

  .col-status {
    margin-top: 0.5rem;
  }

  .col-error {
    margin-top: 0.5rem;
    padding-left: 1.5rem;
  }

  .col-error::before {
    content: '→ ';
  }
}
`);

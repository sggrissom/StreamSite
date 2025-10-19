import { block } from "vlens/css";

// Access Codes Section
block(`
.access-codes-section {
  margin-top: 3rem;
}
`);

block(`
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
`);

block(`
.section-header h2 {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--hero);
  margin: 0;
}
`);

block(`
.section-actions {
  display: flex;
  gap: 0.75rem;
}
`);

// Empty and Loading States
block(`
.empty-state {
  padding: 3rem 2rem;
  text-align: center;
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 12px;
  color: var(--muted);
}
`);

block(`
.empty-state p {
  margin: 0;
  font-size: 1rem;
}
`);

block(`
.loading-state {
  padding: 3rem 2rem;
  text-align: center;
  color: var(--muted);
  font-size: 1rem;
}
`);

// Table Container
block(`
.codes-table-container {
  overflow-x: auto;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
}
`);

// Table Styling
block(`
.codes-table {
  width: 100%;
  border-collapse: collapse;
}
`);

block(`
.codes-table thead {
  background: var(--bg);
  border-bottom: 2px solid var(--border);
}
`);

block(`
.codes-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.codes-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text);
}
`);

block(`
.codes-table tbody tr:last-child td {
  border-bottom: none;
}
`);

block(`
.codes-table tbody tr:hover {
  background: var(--bg);
}
`);

// Table Cell Specific Styling
block(`
.code-cell {
  font-family: 'Courier New', monospace;
  font-size: 1.1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
`);

block(`
.code-display {
  color: var(--hero);
  letter-spacing: 0.1em;
}
`);

block(`
.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  font-size: 1rem;
  opacity: 0.6;
  transition: opacity 0.2s ease, transform 0.1s ease;
}
`);

block(`
.btn-icon:hover {
  opacity: 1;
  transform: scale(1.1);
}
`);

block(`
.label-cell {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
`);

block(`
.label-cell em {
  color: var(--muted);
  font-style: italic;
}
`);

block(`
.scope-cell {
  font-size: 0.9rem;
}
`);

block(`
.scope-studio {
  display: inline-block;
  padding: 0.3rem 0.6rem;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.85rem;
}
`);

block(`
.scope-room {
  color: var(--text);
  font-weight: 500;
}
`);

block(`
.status-cell {
  text-align: center;
}
`);

block(`
.expires-cell {
  font-size: 0.95rem;
}
`);

block(`
.countdown {
  font-weight: 500;
  color: var(--text);
}
`);

block(`
.expired-text {
  color: var(--muted);
  font-style: italic;
}
`);

block(`
.viewers-cell {
  text-align: center;
  font-weight: 600;
  color: var(--hero);
}
`);

block(`
.total-cell {
  text-align: center;
  color: var(--muted);
}
`);

block(`
.actions-cell {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
`);

// Status Badges
block(`
.status-badge {
  display: inline-block;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
}
`);

block(`
.status-badge.status-active {
  background: #d1fae5;
  color: #059669;
  border: 1px solid #6ee7b7;
}
`);

block(`
.status-badge.status-expired {
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #d1d5db;
}
`);

block(`
.status-badge.status-revoked {
  background: #fee;
  color: #c00;
  border: 1px solid #fcc;
}
`);

// Row Status Classes
block(`
.codes-table tbody tr.status-revoked {
  opacity: 0.7;
}
`);

block(`
.codes-table tbody tr.status-expired {
  opacity: 0.8;
}
`);

// Revoke Modal Specific Styles
block(`
.revoke-details {
  margin: 1.5rem 0;
  padding: 1rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
}
`);

block(`
.detail-row {
  margin: 0.5rem 0;
  font-size: 0.95rem;
  color: var(--text);
}
`);

block(`
.detail-row strong {
  color: var(--muted);
  margin-right: 0.5rem;
}
`);

block(`
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}
`);

// Responsive Design
block(`
@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .section-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .codes-table th,
  .codes-table td {
    padding: 0.75rem;
    font-size: 0.85rem;
  }

  .code-cell {
    font-size: 0.95rem;
    flex-direction: column;
    align-items: flex-start;
  }

  .label-cell {
    max-width: 150px;
  }

  .actions-cell {
    flex-direction: column;
  }

  .modal-actions {
    flex-direction: column-reverse;
  }

  .modal-actions .btn {
    width: 100%;
  }
}
`);

block(`
@media (max-width: 480px) {
  .section-header h2 {
    font-size: 1.5rem;
  }

  .section-actions {
    flex-direction: column;
    width: 100%;
  }

  .section-actions .btn {
    width: 100%;
  }

  .codes-table {
    font-size: 0.8rem;
  }

  .code-display {
    font-size: 1rem;
  }

  .status-badge {
    font-size: 0.75rem;
    padding: 0.3rem 0.6rem;
  }
}
`);

import { block } from "vlens/css";

// Modal content
block(`
.class-permissions-modal {
  padding: 1.5rem;
  max-width: 700px;
  min-width: 600px;
}
`);

// Section titles
block(`
.permissions-section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 1rem;
}
`);

// Add section
block(`
.permissions-add-section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
}
`);

block(`
.permissions-add-form {
  background: var(--bg-secondary, #f8f9fa);
  padding: 1rem;
  border-radius: 4px;
}
`);

block(`
.permissions-form-row {
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;
}
`);

block(`
.permissions-form-group {
  flex: 1;
  min-width: 0;
}
`);

block(`
.permissions-form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 0.5rem;
}
`);

block(`
.permissions-form-select {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  cursor: pointer;
  box-sizing: border-box;
}
`);

block(`
.permissions-form-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: var(--bg-disabled, #e9ecef);
}
`);

block(`
.btn-permissions-grant {
  padding: 0.625rem 1.25rem;
  background: var(--primary, #007bff);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease;
  white-space: nowrap;
  height: fit-content;
}
`);

block(`
.btn-permissions-grant:hover:not(:disabled) {
  background: var(--primary-dark, #0056b3);
}
`);

block(`
.btn-permissions-grant:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

// Error message
block(`
.permissions-error {
  padding: 0.75rem 1rem;
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 4px;
  color: var(--danger, #dc3545);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}
`);

// List section
block(`
.permissions-list-section {
  margin-bottom: 1.5rem;
}
`);

block(`
.permissions-loading {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary, #666);
  font-size: 0.875rem;
}
`);

block(`
.permissions-empty {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary, #666);
  font-size: 0.875rem;
  background: var(--bg-secondary, #f8f9fa);
  border-radius: 4px;
}
`);

// Table
block(`
.permissions-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}
`);

block(`
.permissions-table thead {
  background: var(--bg-secondary, #f8f9fa);
}
`);

block(`
.permissions-table th {
  padding: 0.75rem;
  text-align: left;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  border-bottom: 2px solid var(--border-color, #e0e0e0);
}
`);

block(`
.permissions-table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  color: var(--text-primary, #1a1a1a);
}
`);

block(`
.permissions-table tbody tr:hover {
  background: var(--bg-hover, #f8f9fa);
}
`);

block(`
.permissions-table-email {
  color: var(--text-secondary, #666);
  font-size: 0.8125rem;
}
`);

block(`
.btn-permissions-revoke {
  padding: 0.375rem 0.75rem;
  background: var(--danger, #dc3545);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease;
}
`);

block(`
.btn-permissions-revoke:hover:not(:disabled) {
  background: var(--danger-dark, #c82333);
}
`);

block(`
.btn-permissions-revoke:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

// Modal actions
block(`
.permissions-modal-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color, #e0e0e0);
}
`);

block(`
.btn-permissions-close {
  padding: 0.625rem 1.5rem;
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
.btn-permissions-close:hover {
  background: var(--bg-hover, #f8f9fa);
  border-color: var(--text-secondary, #666);
}
`);

// Responsive design
block(`
@media (max-width: 700px) {
  .class-permissions-modal {
    min-width: unset;
    padding: 1rem;
  }

  .permissions-form-row {
    flex-direction: column;
    align-items: stretch;
  }

  .btn-permissions-grant {
    width: 100%;
  }

  .permissions-table {
    font-size: 0.8125rem;
  }

  .permissions-table th,
  .permissions-table td {
    padding: 0.5rem;
  }
}
`);

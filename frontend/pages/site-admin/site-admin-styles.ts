import { block } from "vlens/css";

block(`
.site-admin-container {
  min-height: 100vh;
  padding: 2rem;
  background: var(--background);
}
`);

block(`
.site-admin-content {
  max-width: 1400px;
  margin: 0 auto;
}
`);

block(`
.site-admin-header {
  margin-bottom: 2rem;
}
`);

block(`
.site-admin-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text);
}
`);

block(`
.site-admin-description {
  font-size: 1.125rem;
  color: var(--text-secondary);
  margin: 0;
}
`);

block(`
.admin-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
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
.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  color: var(--text);
}
`);

block(`
.admin-table {
  width: 100%;
  border-collapse: collapse;
}
`);

block(`
.admin-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  background: var(--background);
  border-bottom: 2px solid var(--border);
  font-weight: 600;
  color: var(--text);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.admin-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text);
  vertical-align: top;
}
`);

block(`
.admin-table tr:hover {
  background: var(--background);
}
`);

block(`
.studio-name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
`);

block(`
.studio-description-preview {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
}
`);

block(`
.owner-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
`);

block(`
.email-text {
  font-size: 0.875rem;
  color: var(--text-secondary);
}
`);

block(`
.actions-cell {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
`);

block(`
.role-badge {
  display: inline-block;
  padding: 0.35rem 0.75rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 600;
  white-space: nowrap;
}
`);

block(`
.role-badge.role-0 {
  background: #e3f2fd;
  color: #1565c0;
}
`);

block(`
.role-badge.role-1 {
  background: #fff3e0;
  color: #e65100;
}
`);

block(`
.role-badge.role-2 {
  background: #fce4ec;
  color: #ad1457;
}
`);

block(`
.role-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--background);
  color: var(--text);
  font-size: 0.875rem;
  cursor: pointer;
}
`);

block(`
.role-select:hover {
  border-color: var(--primary);
}
`);

block(`
.role-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
`);

block(`
.text-muted {
  color: var(--text-secondary);
  font-style: italic;
  font-size: 0.875rem;
}
`);

block(`
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-secondary);
}
`);

block(`
.loading-state {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--text-secondary);
}
`);

block(`
.error-state {
  text-align: center;
  padding: 3rem 1rem;
}
`);

block(`
.error-state .error-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}
`);

block(`
.error-state h2 {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--text);
}
`);

block(`
.error-state p {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  line-height: 1.6;
}
`);

block(`
.error-message {
  background: #ffebee;
  color: #c62828;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  border-left: 4px solid #c62828;
}
`);

block(`
.delete-confirmation {
  padding: 0.5rem 0;
}
`);

block(`
.confirmation-text {
  font-size: 1rem;
  color: var(--text);
  margin-bottom: 1rem;
  line-height: 1.6;
}
`);

block(`
.studio-info {
  background: var(--background);
  padding: 0.75rem;
  border-radius: 4px;
  border-left: 3px solid var(--primary);
}
`);

block(`
.studios-table-container {
  overflow-x: auto;
}
`);

block(`
.users-table-container {
  overflow-x: auto;
}
`);

block(`
.section-nav {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 2rem;
  border-bottom: 2px solid var(--border);
}
`);

block(`
.nav-tab {
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 3px solid transparent;
  color: var(--text-secondary);
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
`);

block(`
.nav-tab:hover {
  color: var(--text);
  background: var(--background);
}
`);

block(`
.nav-tab.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
}
`);

block(`
.log-filters {
  background: var(--background);
  padding: 1rem;
  border-radius: 6px;
  margin-bottom: 1.5rem;
  border: 1px solid var(--border);
}
`);

block(`
.filter-row {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
  flex-wrap: wrap;
}
`);

block(`
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 120px;
}
`);

block(`
.filter-search-group {
  flex: 1;
  min-width: 200px;
}
`);

block(`
.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text);
}
`);

block(`
.filter-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-size: 0.875rem;
  cursor: pointer;
}
`);

block(`
.filter-select:hover {
  border-color: var(--primary);
}
`);

block(`
.filter-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-size: 0.875rem;
  width: 100%;
}
`);

block(`
.filter-input:focus {
  outline: none;
  border-color: var(--primary);
}
`);

block(`
.log-count {
  margin-top: 0.75rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}
`);

block(`
.logs-table-container {
  overflow-x: auto;
}
`);

block(`
.logs-table {
  font-size: 0.875rem;
}
`);

block(`
.log-row:hover {
  background: var(--background);
}
`);

block(`
.log-timestamp {
  white-space: nowrap;
  font-family: monospace;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}
`);

block(`
.log-level-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.log-level-info {
  background: #e3f2fd;
  color: #1565c0;
}
`);

block(`
.log-level-warn {
  background: #fff3e0;
  color: #e65100;
}
`);

block(`
.log-level-error {
  background: #ffebee;
  color: #c62828;
}
`);

block(`
.log-level-debug {
  background: #f5f5f5;
  color: #616161;
}
`);

block(`
.log-level-unknown {
  background: #e0e0e0;
  color: #424242;
}
`);

block(`
.log-category {
  font-weight: 500;
  color: var(--text);
}
`);

block(`
.log-message {
  max-width: 500px;
  word-break: break-word;
}
`);

block(`
.log-details-row {
  background: var(--background);
}
`);

block(`
.log-details-row:hover {
  background: var(--background);
}
`);

block(`
.log-details {
  padding: 1rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--surface);
}
`);

block(`
.log-detail-item {
  margin-bottom: 0.75rem;
}
`);

block(`
.log-detail-item:last-child {
  margin-bottom: 0;
}
`);

block(`
.log-detail-item strong {
  color: var(--text);
  display: block;
  margin-bottom: 0.25rem;
}
`);

block(`
.log-data-json {
  background: var(--background);
  padding: 0.75rem;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 0.8125rem;
  color: var(--text);
  margin-top: 0.5rem;
  border: 1px solid var(--border);
}
`);

block(`
@media (max-width: 768px) {
  .site-admin-container {
    padding: 1rem;
  }

  .site-admin-title {
    font-size: 1.5rem;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .admin-table {
    font-size: 0.875rem;
  }

  .admin-table th,
  .admin-table td {
    padding: 0.5rem;
  }

  .actions-cell {
    flex-direction: column;
  }

  .actions-cell .btn {
    width: 100%;
  }

  .section-nav {
    overflow-x: auto;
    flex-wrap: nowrap;
  }

  .nav-tab {
    white-space: nowrap;
    padding: 0.75rem 1rem;
  }

  .filter-row {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    min-width: auto;
  }

  .filter-search-group {
    min-width: auto;
  }

  .log-message {
    max-width: 100%;
  }

  .logs-table {
    font-size: 0.8125rem;
  }

  .log-timestamp {
    font-size: 0.75rem;
  }
}
`);

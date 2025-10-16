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
}
`);

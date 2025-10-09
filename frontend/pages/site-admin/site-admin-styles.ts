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
  margin-bottom: 2rem;
}
`);

block(`
.admin-sections {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
}
`);

block(`
.admin-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
}
`);

block(`
.admin-section.full-width {
  grid-column: 1 / -1;
}
`);

block(`
.admin-section h2 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text);
}
`);

block(`
.admin-section p {
  color: var(--text-secondary);
  margin-bottom: 1rem;
  line-height: 1.6;
}
`);

block(`
.section-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
`);

block(`
.users-table {
  overflow-x: auto;
  margin-top: 1rem;
}
`);

block(`
.users-table table {
  width: 100%;
  border-collapse: collapse;
}
`);

block(`
.users-table th {
  text-align: left;
  padding: 0.75rem;
  background: var(--background);
  border-bottom: 2px solid var(--border);
  font-weight: 600;
  color: var(--text);
}
`);

block(`
.users-table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--border);
  color: var(--text);
}
`);

block(`
.users-table select {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--background);
  color: var(--text);
  font-size: 0.875rem;
}
`);

block(`
.role-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
}
`);

block(`
.role-badge.role-0 {
  background: #e3f2fd;
  color: #1976d2;
}
`);

block(`
.role-badge.role-1 {
  background: #fff3e0;
  color: #f57c00;
}
`);

block(`
.role-badge.role-2 {
  background: #fce4ec;
  color: #c2185b;
}
`);

block(`
.text-muted {
  color: var(--text-secondary);
  font-style: italic;
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
@media (max-width: 768px) {
  .site-admin-container {
    padding: 1rem;
  }

  .site-admin-title {
    font-size: 1.5rem;
  }

  .admin-sections {
    grid-template-columns: 1fr;
  }

  .users-table {
    font-size: 0.875rem;
  }

  .users-table th,
  .users-table td {
    padding: 0.5rem;
  }
}
`);

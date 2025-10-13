import { block } from "vlens/css";

block(`
.stream-admin-container {
  min-height: 100vh;
  padding: 2rem;
  background: var(--background);
}
`);

block(`
.stream-admin-content {
  max-width: 1200px;
  margin: 0 auto;
}
`);

block(`
.stream-admin-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text);
}
`);

block(`
.stream-admin-description {
  font-size: 1.125rem;
  color: var(--text-secondary);
  margin-bottom: 2rem;
}
`);

block(`
.admin-overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}
`);

block(`
.overview-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  gap: 1rem;
  align-items: center;
}
`);

block(`
.overview-icon {
  font-size: 2.5rem;
  flex-shrink: 0;
}
`);

block(`
.overview-content {
  flex: 1;
}
`);

block(`
.overview-content h3 {
  font-size: 0.875rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}
`);

block(`
.overview-stat {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
  line-height: 1;
}
`);

block(`
.overview-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0.25rem 0 0 0;
}
`);

block(`
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
`);

block(`
.studios-section {
  margin-bottom: 3rem;
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
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
`);

block(`
.empty-state {
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 8px;
  padding: 3rem 2rem;
  text-align: center;
}
`);

block(`
.empty-state .empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}
`);

block(`
.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.5rem;
}
`);

block(`
.empty-state p {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}
`);

block(`
.studios-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
`);

block(`
.studio-item {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  transition: box-shadow 0.2s;
}
`);

block(`
.studio-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
`);

block(`
.studio-item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1rem;
}
`);

block(`
.studio-item-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.25rem 0;
}
`);

block(`
.studio-item-description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}
`);

block(`
.studio-item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
`);

block(`
.studio-item-meta {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
`);

block(`
.meta-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
}
`);

block(`
.admin-help {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 2rem;
}
`);

block(`
.admin-help h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 1.5rem 0;
}
`);

block(`
.help-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}
`);

block(`
.help-card h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.5rem 0;
}
`);

block(`
.help-card p {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}
`);

block(`
@media (max-width: 768px) {
  .stream-admin-container {
    padding: 1rem;
  }

  .stream-admin-title {
    font-size: 1.5rem;
  }

  .admin-overview {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .studio-item-header {
    flex-direction: column;
  }

  .studio-item-footer {
    flex-direction: column;
    align-items: flex-start;
  }

  .help-grid {
    grid-template-columns: 1fr;
  }
}
`);

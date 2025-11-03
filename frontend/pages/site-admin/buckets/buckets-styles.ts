import { block } from "vlens/css";

block(`
.bucket-viewer-container {
  padding: 2rem;
  max-width: 1800px;
  margin: 0 auto;
  min-height: calc(100vh - 200px);
}
`);

block(`
.bucket-viewer-header {
  margin-bottom: 2rem;
}
`);

block(`
.bucket-viewer-header h1 {
  margin: 0 0 0.5rem 0;
  color: var(--text);
}
`);

block(`
.bucket-viewer-subtitle {
  margin: 0;
  color: var(--text-secondary);
}
`);

block(`
.bucket-viewer-content {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 2rem;
  align-items: start;
}
`);

block(`
.bucket-list-sidebar {
  background: var(--surface);
  border-radius: 8px;
  padding: 1.5rem;
  border: 1px solid var(--border);
  position: sticky;
  top: 2rem;
  max-height: calc(100vh - 4rem);
  overflow-y: auto;
}
`);

block(`
.bucket-list-sidebar h2 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
}
`);

block(`
.bucket-list-sidebar h2:not(:first-child) {
  margin-top: 2rem;
}
`);

block(`
.bucket-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
`);

block(`
.bucket-list-item {
  padding: 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  border: 1px solid transparent;
}
`);

block(`
.bucket-list-item:hover {
  background: var(--surface-hover);
}
`);

block(`
.bucket-list-item.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}
`);

block(`
.bucket-list-item.active .bucket-description,
.bucket-list-item.active .bucket-meta {
  color: rgba(255, 255, 255, 0.9);
}
`);

block(`
.bucket-name {
  font-weight: 600;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
}
`);

block(`
.bucket-description {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 0.25rem;
}
`);

block(`
.bucket-meta {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  font-family: 'Courier New', monospace;
}
`);

block(`
.bucket-data-area {
  background: var(--surface);
  border-radius: 8px;
  padding: 2rem;
  border: 1px solid var(--border);
  min-height: 400px;
}
`);

block(`
.bucket-data-empty,
.bucket-data-loading,
.bucket-data-error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: var(--text-secondary);
}
`);

block(`
.bucket-data-error {
  color: var(--error);
}
`);

block(`
.bucket-data-header {
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border);
}
`);

block(`
.bucket-data-header h2 {
  margin: 0 0 0.5rem 0;
  font-family: 'Courier New', monospace;
  color: var(--text);
}
`);

block(`
.bucket-data-count {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}
`);

block(`
.bucket-data-table-container {
  overflow-x: auto;
}
`);

block(`
.bucket-data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
`);

block(`
.bucket-data-table thead {
  background: var(--surface-hover);
}
`);

block(`
.bucket-data-table th {
  text-align: left;
  padding: 0.75rem;
  font-weight: 600;
  color: var(--text);
  border-bottom: 2px solid var(--border);
}
`);

block(`
.bucket-data-table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--border);
  vertical-align: top;
}
`);

block(`
.bucket-data-table tbody tr:hover {
  background: var(--surface-hover);
}
`);

block(`
.bucket-data-key {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: var(--primary);
  width: 200px;
  white-space: nowrap;
}
`);

block(`
.bucket-data-value {
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
}
`);

block(`
.bucket-data-value pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--text);
}
`);

block(`
.bucket-viewer-error {
  text-align: center;
  padding: 3rem;
}
`);

block(`
.bucket-viewer-error h1 {
  color: var(--error);
  margin-bottom: 1rem;
}
`);

// Responsive styles
block(`
@media (max-width: 1200px) {
  .bucket-viewer-content {
    grid-template-columns: 250px 1fr;
    gap: 1rem;
  }
}
`);

block(`
@media (max-width: 768px) {
  .bucket-viewer-container {
    padding: 1rem;
  }

  .bucket-viewer-content {
    grid-template-columns: 1fr;
  }

  .bucket-list-sidebar {
    position: static;
    max-height: none;
  }

  .bucket-data-key {
    width: auto;
    white-space: normal;
  }
}
`);

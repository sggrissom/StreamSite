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
@media (max-width: 768px) {
  .stream-admin-container {
    padding: 1rem;
  }

  .stream-admin-title {
    font-size: 1.5rem;
  }

  .admin-sections {
    grid-template-columns: 1fr;
  }
}
`);

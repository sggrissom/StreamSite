import { block } from "vlens/css";

block(`
.settings-container {
  min-height: 100vh;
  padding: 2rem;
  background: var(--background);
}
`);

block(`
.settings-content {
  max-width: 1200px;
  margin: 0 auto;
}
`);

block(`
.settings-header {
  margin-bottom: 2rem;
}
`);

block(`
.settings-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text);
}
`);

block(`
.settings-description {
  font-size: 1.125rem;
  color: var(--text-secondary);
  margin: 0;
}
`);

block(`
.settings-sections {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}
`);

block(`
.settings-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
}
`);

block(`
.settings-section h2 {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: var(--text);
}
`);

block(`
.settings-section p {
  color: var(--text-secondary);
  margin: 0 0 1rem 0;
  line-height: 1.6;
}
`);

block(`
.settings-placeholder {
  background: var(--background);
  border: 2px dashed var(--border);
  border-radius: 6px;
  padding: 2rem;
  text-align: center;
}
`);

block(`
.placeholder-text {
  color: var(--text-secondary);
  font-style: italic;
  font-size: 0.95rem;
}
`);

block(`
.settings-actions {
  display: flex;
  gap: 1rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border);
}
`);

block(`
.btn-secondary {
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
}
`);

block(`
.btn-secondary:hover {
  background: var(--hover-bg);
  border-color: var(--accent);
}
`);

block(`
@media (max-width: 768px) {
  .settings-container {
    padding: 1rem;
  }

  .settings-title {
    font-size: 1.5rem;
  }

  .settings-sections {
    grid-template-columns: 1fr;
  }

  .settings-section {
    padding: 1.25rem;
  }

  .settings-placeholder {
    padding: 1.5rem;
  }
}
`);

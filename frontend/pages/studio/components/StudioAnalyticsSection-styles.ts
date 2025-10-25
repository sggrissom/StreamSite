import { block } from "vlens/css";

block(`
.studio-analytics-section {
  background: var(--surface);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  border: 1px solid var(--border);
}
`);

block(`
.studio-analytics-section .section-header {
  margin-bottom: 1.5rem;
}
`);

block(`
.studio-analytics-section .section-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
`);

block(`
.analytics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}
`);

block(`
.analytics-card {
  background: var(--background);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.25rem;
  transition: transform 0.2s, box-shadow 0.2s;
}
`);

block(`
.analytics-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
`);

block(`
.analytics-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
`);

block(`
.analytics-icon {
  font-size: 1.5rem;
  line-height: 1;
}
`);

block(`
.analytics-card-header h3 {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
`);

block(`
.analytics-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
  margin-bottom: 0.5rem;
}
`);

block(`
.analytics-value-primary {
  color: var(--primary);
}
`);

block(`
.analytics-unit {
  font-size: 1.25rem;
  font-weight: 500;
  color: var(--text-secondary);
}
`);

block(`
.analytics-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
}
`);

block(`
.analytics-sublabel {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-top: 0.25rem;
}
`);

block(`
.analytics-loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}
`);

block(`
.analytics-error {
  text-align: center;
  padding: 2rem;
  color: var(--danger);
  background: rgba(220, 38, 38, 0.1);
  border-radius: 8px;
}
`);

block(`
.analytics-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}
`);

block(`
.analytics-footer {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}
`);

block(`
.analytics-refresh-note {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  text-align: center;
  margin: 0;
}
`);

// Responsive styles
block(`
@media (max-width: 768px) {
  .analytics-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .analytics-value {
    font-size: 1.75rem;
  }

  .analytics-card {
    padding: 1rem;
  }
}
`);

block(`
@media (max-width: 480px) {
  .analytics-grid {
    grid-template-columns: 1fr;
  }

  .studio-analytics-section {
    padding: 1rem;
  }

  .analytics-value {
    font-size: 1.5rem;
  }
}
`);

import { block } from "vlens/css";

block(`
.site-header {
  padding: 1rem;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
}
`);

block(`
.site-nav {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
`);

block(`
.site-logo {
  font-size: 1.25rem;
  font-weight: bold;
  color: var(--text);
  text-decoration: none;
  transition: color var(--transition-speed) ease;
}
`);

block(`
.site-logo:hover {
  color: var(--accent);
}
`);

block(`
.nav-links {
  display: flex;
  align-items: center;
  gap: 1rem;
}
`);

block(`
.nav-link {
  padding: 0.5rem 1rem;
  color: var(--text);
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  border-radius: 8px;
  transition: all var(--transition-speed) ease;
}
`);

block(`
.nav-link:hover {
  background: var(--hover-bg);
  color: var(--accent);
}
`);

block(`
.logout-button {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-speed) ease;
}
`);

block(`
.logout-button:hover {
  background: var(--hover-bg);
  border-color: #dc2626;
  color: #dc2626;
}
`);

block(`
.site-footer {
  padding: 1rem;
  text-align: center;
  color: var(--muted);
  border-top: 1px solid var(--border);
  background: var(--surface);
}
`);

block(`
.site-footer p {
  margin: 0;
}
`);

block(`
.auth-links {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.9rem;
}
`);

block(`
.auth-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}
`);

block(`
.auth-link:hover {
  color: var(--primary);
}
`);

block(`
.auth-separator {
  color: var(--text-tertiary);
}
`);

block(`
@media (max-width: 768px) {
  .nav-links {
    gap: 0.5rem;
  }

  .nav-link,
  .logout-button {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }
}
`);

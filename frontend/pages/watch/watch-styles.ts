import { block } from "vlens/css";

block(`
.watch-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 200px);
  padding: 2rem;
}
`);

block(`
.watch-card {
  background: var(--surface);
  border-radius: 12px;
  padding: 3rem;
  max-width: 500px;
  width: 100%;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
`);

block(`
.watch-title {
  font-size: 2rem;
  font-weight: 600;
  margin: 0 0 1rem 0;
  color: var(--text);
  text-align: center;
}
`);

block(`
.watch-description {
  color: var(--text-secondary);
  text-align: center;
  margin: 0 0 2rem 0;
  line-height: 1.5;
}
`);

block(`
.watch-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
`);

block(`
.watch-input-group {
  display: flex;
  justify-content: center;
}
`);

block(`
.watch-input {
  width: 100%;
  padding: 1rem;
  font-size: 2rem;
  font-weight: 600;
  text-align: center;
  letter-spacing: 0.5rem;
  border: 2px solid var(--border);
  border-radius: 8px;
  background: var(--background);
  color: var(--text);
  font-family: 'Courier New', monospace;
  transition: border-color 0.2s;
}
`);

block(`
.watch-input:focus {
  outline: none;
  border-color: var(--primary);
}
`);

block(`
.watch-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.watch-error {
  padding: 1rem;
  background: rgba(220, 38, 38, 0.1);
  border: 1px solid rgba(220, 38, 38, 0.3);
  border-radius: 8px;
  color: #dc2626;
  text-align: center;
  font-size: 0.9rem;
}
`);

block(`
.watch-submit-btn {
  padding: 1rem 2rem;
  font-size: 1.1rem;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  background: linear-gradient(90deg, var(--accent), var(--primary-accent));
  color: var(--button-text);
  cursor: pointer;
  transition: all 0.2s;
}
`);

block(`
.watch-submit-btn:hover:not(:disabled) {
  background: linear-gradient(90deg, var(--accent-hover), var(--primary-accent-hover));
  filter: brightness(1.05);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
`);

block(`
.watch-submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.watch-help {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--border);
  text-align: center;
}
`);

block(`
.watch-help p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin: 0;
}
`);

// Mobile responsive design
block(`
@media (max-width: 600px) {
  .watch-card {
    padding: 2rem 1.5rem;
  }

  .watch-title {
    font-size: 1.75rem;
  }

  .watch-input {
    font-size: 1.75rem;
    letter-spacing: 0.4rem;
    padding: 0.875rem;
  }

  .watch-submit-btn {
    padding: 0.875rem 1.5rem;
    font-size: 1rem;
  }
}
`);

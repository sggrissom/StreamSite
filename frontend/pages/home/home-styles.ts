import { block } from "vlens/css";

block(`
.landing-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 200px);
  padding: 2rem;
}
`);

block(`
.landing-card {
  background: var(--surface);
  border-radius: 12px;
  padding: 3rem;
  max-width: 500px;
  width: 100%;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
`);

block(`
.landing-title {
  font-size: 2rem;
  font-weight: 600;
  margin: 0 0 1rem 0;
  color: var(--text);
  text-align: center;
}
`);

block(`
.landing-description {
  color: var(--text-secondary);
  text-align: center;
  margin: 0 0 2rem 0;
  line-height: 1.5;
}
`);

block(`
.landing-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
`);

block(`
.landing-input-group {
  display: flex;
  justify-content: center;
}
`);

block(`
.landing-input {
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
.landing-input:focus {
  outline: none;
  border-color: var(--primary);
}
`);

block(`
.landing-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.landing-error {
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
.landing-rate-limit {
  padding: 1rem;
  background: rgba(234, 179, 8, 0.1);
  border: 1px solid rgba(234, 179, 8, 0.3);
  border-radius: 8px;
  color: #ca8a04;
  text-align: center;
  font-size: 0.95rem;
  font-weight: 500;
}
`);

block(`
.landing-submit-btn {
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
.landing-submit-btn:hover:not(:disabled) {
  background: linear-gradient(90deg, var(--accent-hover), var(--primary-accent-hover));
  filter: brightness(1.05);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
`);

block(`
.landing-submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.landing-help {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--border);
  text-align: center;
}
`);

block(`
.landing-help p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin: 0;
}
`);

block(`
.landing-auth-links {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.9rem;
  color: var(--text-secondary);
}
`);

block(`
.landing-auth-links a {
  color: var(--primary);
  text-decoration: none;
  font-weight: 500;
}
`);

block(`
.landing-auth-links a:hover {
  text-decoration: underline;
}
`);

// Mobile responsive design
block(`
@media (max-width: 600px) {
  .landing-card {
    padding: 2rem 1.5rem;
  }

  .landing-title {
    font-size: 1.75rem;
  }

  .landing-input {
    font-size: 1.75rem;
    letter-spacing: 0.4rem;
    padding: 0.875rem;
  }

  .landing-submit-btn {
    padding: 0.875rem 1.5rem;
    font-size: 1rem;
  }
}
`);

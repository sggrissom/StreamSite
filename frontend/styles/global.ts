import { block } from "vlens/css";

// Global resets
block(`
* {
  box-sizing: border-box;
}
`);

block(`
html,
body {
  height: 100%;
}
`);

// Button Styles
block(`
.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-radius: 8px;
  text-decoration: none;
  font-size: 1rem;
  font-weight: 600;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  transition: all var(--transition-speed) ease;
  min-height: 48px;
}
`);

block(`
.btn:hover {
  background: var(--hover-bg);
  transform: translateY(-1px);
}
`);

block(`
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}
`);

block(`
.btn-primary {
  background: linear-gradient(90deg, var(--accent), var(--primary-accent));
  color: var(--button-text);
  border: none;
  font-weight: 700;
}
`);

block(`
.btn-primary:hover {
  background: linear-gradient(
    90deg,
    var(--accent-hover),
    var(--primary-accent-hover)
  );
  filter: brightness(1.05);
}
`);

block(`
.btn-large {
  padding: 14px 24px;
  font-size: 1.05rem;
}
`);

// Form Styles
block(`
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
`);

block(`
.form-group label {
  font-weight: 600;
  color: var(--text);
  font-size: 0.9rem;
}
`);

block(`
.form-group input,
.form-group select,
.form-group textarea {
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
  color: var(--text);
  font-size: 1rem;
  font-family: inherit;
  transition: border-color var(--transition-speed) ease;
}
`);

block(`
.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}
`);

block(`
.form-group input::placeholder {
  color: var(--muted);
}
`);

block(`
.form-group input:disabled,
.form-group select:disabled,
.form-group textarea:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.form-hint {
  color: var(--muted);
  font-size: 0.85rem;
  margin-top: 4px;
}
`);

// Message Styles
block(`
.error-message {
  background: #fee2e2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 0.9rem;
}
`);

block(`
.success-message {
  background: #d1fae5;
  border: 1px solid #a7f3d0;
  color: #065f46;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 0.9rem;
}
`);

// Auth Card Styles (shared between login and create-account)
block(`
.auth-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  width: 100%;
}
`);

block(`
.auth-header {
  text-align: center;
  margin-bottom: 32px;
}
`);

block(`
.auth-header h1 {
  font-size: 2rem;
  margin: 0 0 8px;
  color: var(--text);
}
`);

block(`
.auth-header p {
  color: var(--muted);
  margin: 0;
  font-size: 1rem;
}
`);

block(`
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
`);

block(`
.auth-submit {
  margin-top: 8px;
  width: 100%;
  justify-content: center;
}
`);

block(`
.auth-footer {
  margin-top: 24px;
  text-align: center;
  padding-top: 24px;
  border-top: 1px solid var(--border);
}
`);

block(`
.auth-footer p {
  color: var(--muted);
  margin: 0;
}
`);

block(`
.auth-link {
  color: var(--accent);
  text-decoration: none;
  font-weight: 600;
  margin-left: 8px;
  transition: color var(--transition-speed) ease;
}
`);

block(`
.auth-link:hover {
  color: var(--primary-accent);
  text-decoration: underline;
}
`);

// Mobile responsive styles
block(`
@media (max-width: 480px) {
  .auth-card {
    padding: 32px 24px;
  }

  .auth-header h1 {
    font-size: 1.8rem;
  }

  .btn-large {
    padding: 12px 20px;
    font-size: 1rem;
  }
}
`);

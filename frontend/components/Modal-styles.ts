import { block } from "vlens/css";

// Modal Overlay
block(`
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}
`);

// Modal Content
block(`
.modal-content {
  background: var(--surface);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}
`);

block(`
.modal-content-large {
  max-width: 700px;
}
`);

// Modal Header
block(`
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border);
}
`);

block(`
.modal-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--hero);
  margin: 0;
}
`);

block(`
.modal-close {
  background: none;
  border: none;
  font-size: 2rem;
  color: var(--muted);
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  transition: color 0.2s ease;
}
`);

block(`
.modal-close:hover {
  color: var(--text);
}
`);

// Modal Body
block(`
.modal-body {
  padding: 1.5rem;
}
`);

// Modal Footer
block(`
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid var(--border);
}
`);

// Form Elements
block(`
.form-group {
  margin-bottom: 1.5rem;
}
`);

block(`
.form-group:last-child {
  margin-bottom: 0;
}
`);

block(`
.form-group label {
  display: block;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 0.5rem;
}
`);

block(`
.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 1rem;
  color: var(--text);
  background: var(--bg);
  transition: border-color 0.2s ease;
  font-family: inherit;
}
`);

block(`
.form-input:focus {
  outline: none;
  border-color: var(--primary-accent);
}
`);

block(`
.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.form-help {
  display: block;
  font-size: 0.85rem;
  color: var(--muted);
  margin-top: 0.5rem;
}
`);

// Error Message
block(`
.error-message {
  padding: 0.75rem 1rem;
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 6px;
  color: #c00;
  font-size: 0.95rem;
  margin-bottom: 1rem;
}
`);

// Button Styles
block(`
.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
  text-align: center;
  transition: all 0.2s ease;
}
`);

block(`
.btn-primary {
  background: var(--primary-accent);
  color: var(--button-text);
}
`);

block(`
.btn-primary:hover:not(:disabled) {
  background: var(--primary-accent-hover);
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
.btn-secondary:hover:not(:disabled) {
  background: var(--hover-bg);
}
`);

block(`
.btn-danger {
  background: #dc2626;
  color: white;
}
`);

block(`
.btn-danger:hover:not(:disabled) {
  background: #b91c1c;
}
`);

block(`
.btn-sm {
  padding: 0.5rem 1rem;
  font-size: 0.9rem;
}
`);

block(`
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
`);

// Responsive Design
block(`
@media (max-width: 768px) {
  .modal-content {
    max-width: 95%;
  }

  .modal-content-large {
    max-width: 95%;
  }

  .modal-footer {
    flex-direction: column-reverse;
  }

  .modal-footer .btn {
    width: 100%;
  }
}
`);

block(`
@media (max-width: 480px) {
  .modal-header {
    padding: 1rem;
  }

  .modal-body {
    padding: 1rem;
  }

  .modal-footer {
    padding: 1rem;
  }

  .modal-title {
    font-size: 1.25rem;
  }
}
`);

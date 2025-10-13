import { block } from "vlens/css";

// Main container
block(`
.studios-container {
  min-height: calc(100vh - 140px);
  padding: 2rem 1rem;
  background: var(--background);
}
`);

block(`
.studios-content {
  max-width: 1200px;
  margin: 0 auto;
}
`);

// Header section
block(`
.studios-header {
  margin-bottom: 2rem;
}
`);

block(`
.studios-title {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.5rem 0;
}
`);

block(`
.studios-description {
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0;
}
`);

// Empty state
block(`
.studios-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
}
`);

block(`
.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}
`);

block(`
.studios-empty h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.5rem 0;
}
`);

block(`
.studios-empty p {
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem 0;
  max-width: 500px;
}
`);

// Studios grid
block(`
.studios-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}
`);

// Studio card
block(`
.studio-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: box-shadow 0.2s ease;
}
`);

block(`
.studio-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
`);

// Studio card header
block(`
.studio-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}
`);

block(`
.studio-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
  flex: 1;
  word-break: break-word;
}
`);

// Role badges
block(`
.studio-role {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
}
`);

block(`
.role-0 {
  background: #e3f2fd;
  color: #1976d2;
}
`);

block(`
.role-1 {
  background: #f3e5f5;
  color: #7b1fa2;
}
`);

block(`
.role-2 {
  background: #fff3e0;
  color: #f57c00;
}
`);

block(`
.role-3 {
  background: #e8f5e9;
  color: #388e3c;
}
`);

// Studio description
block(`
.studio-description {
  font-size: 0.95rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}
`);

// Studio metadata
block(`
.studio-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border);
}
`);

block(`
.meta-item {
  display: flex;
  gap: 0.5rem;
  font-size: 0.9rem;
}
`);

block(`
.meta-label {
  color: var(--text-secondary);
  font-weight: 500;
}
`);

block(`
.meta-value {
  color: var(--text);
  font-weight: 600;
}
`);

// Studio actions
block(`
.studio-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
`);

// Footer with create button
block(`
.studios-footer {
  margin-top: 2rem;
  display: flex;
  justify-content: center;
}
`);

// Responsive adjustments
block(`
@media (max-width: 768px) {
  .studios-container {
    padding: 1.5rem 1rem;
  }

  .studios-title {
    font-size: 1.75rem;
  }

  .studios-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .studio-card {
    padding: 1.25rem;
  }

  .studios-empty {
    padding: 3rem 1.5rem;
  }

  .empty-icon {
    font-size: 3rem;
  }
}
`);

block(`
@media (max-width: 480px) {
  .studios-container {
    padding: 1rem 0.75rem;
  }

  .studios-title {
    font-size: 1.5rem;
  }

  .studio-name {
    font-size: 1.1rem;
  }

  .studio-card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .studio-role {
    align-self: flex-start;
  }

  .btn {
    width: 100%;
  }

  .studio-actions {
    flex-direction: column;
  }
}
`);

// Modal overlay
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

// Modal content
block(`
.modal-content {
  background: var(--surface);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}
`);

// Modal header
block(`
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border);
}
`);

block(`
.modal-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
`);

block(`
.modal-close {
  background: none;
  border: none;
  font-size: 2rem;
  line-height: 1;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}
`);

block(`
.modal-close:hover {
  background: var(--border);
  color: var(--text);
}
`);

// Modal body
block(`
.modal-body {
  padding: 1.5rem;
}
`);

// Form elements
block(`
.form-group {
  margin-bottom: 1.25rem;
}
`);

block(`
.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text);
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
  background: var(--background);
  transition: all 0.2s ease;
}
`);

block(`
.form-input:focus {
  outline: none;
  border-color: var(--primary-accent);
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
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
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}
`);

block(`
textarea.form-input {
  resize: vertical;
  min-height: 80px;
}
`);

// Error message
block(`
.error-message {
  background: #fee;
  border: 1px solid #fcc;
  color: #c33;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}
`);

// Modal footer
block(`
.modal-footer {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 1.5rem;
  border-top: 1px solid var(--border);
}
`);

// Modal responsive
block(`
@media (max-width: 480px) {
  .modal-content {
    max-width: 100%;
    border-radius: 0;
    max-height: 100vh;
  }

  .modal-overlay {
    padding: 0;
  }

  .modal-footer {
    flex-direction: column-reverse;
  }

  .modal-footer .btn {
    width: 100%;
  }
}
`);

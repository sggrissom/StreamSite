import { block } from "vlens/css";

// Modal content
block(`
.create-schedule-modal {
  padding: 1.5rem;
  max-width: 600px;
}
`);

// Form groups
block(`
.schedule-form-group {
  margin-bottom: 1.25rem;
}
`);

block(`
.schedule-form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 0.5rem;
}
`);

block(`
.schedule-form-label-required::after {
  content: " *";
  color: var(--danger, #dc3545);
}
`);

block(`
.schedule-form-input {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  box-sizing: border-box;
}
`);

block(`
.schedule-form-input:focus {
  outline: none;
  border-color: var(--primary, #007bff);
  box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
}
`);

block(`
.schedule-form-textarea {
  min-height: 80px;
  resize: vertical;
  font-family: inherit;
}
`);

block(`
.schedule-form-select {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  cursor: pointer;
}
`);

// Schedule type toggle
block(`
.schedule-type-toggle {
  display: flex;
  gap: 1rem;
  padding: 0.75rem 0;
}
`);

block(`
.schedule-type-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}
`);

block(`
.schedule-type-radio {
  width: 18px;
  height: 18px;
  cursor: pointer;
}
`);

block(`
.schedule-type-label {
  font-size: 0.875rem;
  cursor: pointer;
  user-select: none;
}
`);

// Weekday selector
block(`
.weekday-selector {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
`);

block(`
.weekday-button {
  width: 40px;
  height: 40px;
  border: 2px solid var(--border-color, #e0e0e0);
  background: var(--surface, #fff);
  color: var(--text-secondary, #666);
  border-radius: 50%;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}
`);

block(`
.weekday-button:hover {
  border-color: var(--primary, #007bff);
  background: var(--bg-hover, #f8f9fa);
}
`);

block(`
.weekday-button.selected {
  background: var(--primary, #007bff);
  color: white;
  border-color: var(--primary, #007bff);
}
`);

// Time inputs row
block(`
.time-inputs-row {
  display: flex;
  gap: 1rem;
  align-items: center;
}
`);

block(`
.time-input-group {
  flex: 1;
}
`);

block(`
.time-separator {
  font-size: 0.875rem;
  color: var(--text-secondary, #666);
  padding-top: 1.75rem;
}
`);

// Checkbox toggle
block(`
.checkbox-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0;
}
`);

block(`
.checkbox-toggle input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}
`);

block(`
.checkbox-toggle label {
  font-size: 0.875rem;
  cursor: pointer;
  user-select: none;
  color: var(--text-primary, #1a1a1a);
}
`);

// Number input (for pre-roll/post-roll)
block(`
.number-input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
`);

block(`
.number-input-small {
  width: 80px;
  padding: 0.5rem;
  text-align: center;
}
`);

block(`
.number-input-suffix {
  font-size: 0.875rem;
  color: var(--text-secondary, #666);
}
`);

// Camera automation section
block(`
.camera-automation-section {
  padding: 1rem;
  background: var(--bg-secondary, #f8f9fa);
  border-radius: 4px;
  margin-top: 1rem;
}
`);

block(`
.camera-automation-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 0.75rem;
}
`);

// Error message
block(`
.schedule-form-error {
  color: var(--danger, #dc3545);
  font-size: 0.8125rem;
  margin-top: 0.25rem;
}
`);

// Form actions
block(`
.schedule-form-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color, #e0e0e0);
}
`);

block(`
.btn-schedule-submit {
  padding: 0.625rem 1.5rem;
  background: var(--primary, #007bff);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease;
}
`);

block(`
.btn-schedule-submit:hover:not(:disabled) {
  background: var(--primary-dark, #0056b3);
}
`);

block(`
.btn-schedule-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.btn-schedule-cancel {
  padding: 0.625rem 1.5rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
`);

block(`
.btn-schedule-cancel:hover {
  background: var(--bg-hover, #f8f9fa);
  border-color: var(--text-secondary, #666);
}
`);

// Helper text
block(`
.form-helper-text {
  font-size: 0.8125rem;
  color: var(--text-secondary, #666);
  margin-top: 0.25rem;
}
`);

// Responsive design
block(`
@media (max-width: 600px) {
  .create-schedule-modal {
    padding: 1rem;
  }

  .time-inputs-row {
    flex-direction: column;
    align-items: stretch;
  }

  .time-separator {
    padding-top: 0;
    text-align: center;
  }

  .weekday-selector {
    justify-content: space-between;
  }

  .weekday-button {
    flex: 1;
    max-width: 45px;
  }

  .schedule-form-actions {
    flex-direction: column-reverse;
  }

  .btn-schedule-submit,
  .btn-schedule-cancel {
    width: 100%;
  }
}
`);

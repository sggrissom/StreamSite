import { block } from "vlens/css";

// Camera status indicator
block(`
.camera-status {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}
`);

block(`
.camera-status.running {
  background: var(--success-bg, #d4edda);
  color: var(--success-text, #155724);
  border: 1px solid var(--success-border, #c3e6cb);
}
`);

// Camera controls container
block(`
.camera-controls {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color, #e0e0e0);
}
`);

// Camera control buttons
block(`
.btn-camera {
  background: var(--primary, #007bff);
  color: white;
  border: none;
  transition: background 0.2s ease;
}
`);

block(`
.btn-camera:hover:not(:disabled) {
  background: var(--primary-dark, #0056b3);
}
`);

block(`
.btn-camera:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

block(`
.btn-camera-stop {
  background: var(--danger, #dc3545);
  color: white;
  border: none;
  transition: background 0.2s ease;
}
`);

block(`
.btn-camera-stop:hover:not(:disabled) {
  background: var(--danger-dark, #c82333);
}
`);

block(`
.btn-camera-stop:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
`);

// Camera error message
block(`
.camera-error {
  padding: 0.5rem;
  background: var(--danger-bg, #f8d7da);
  color: var(--danger-text, #721c24);
  border: 1px solid var(--danger-border, #f5c6cb);
  border-radius: 4px;
  font-size: 0.875rem;
}
`);

// Responsive adjustments
block(`
@media (max-width: 600px) {
  .camera-controls {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
  }

  .camera-status {
    font-size: 0.7rem;
    padding: 0.2rem 0.4rem;
  }

  .camera-error {
    font-size: 0.8rem;
    padding: 0.4rem;
  }
}
`);

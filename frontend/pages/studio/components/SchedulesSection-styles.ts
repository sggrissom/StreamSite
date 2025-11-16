import { block } from "vlens/css";

// Schedules section container
block(`
.schedules-section {
  margin-top: 2rem;
  padding: 1.5rem;
  background: var(--surface, #fff);
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
}
`);

// Section header
block(`
.schedules-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
`);

block(`
.schedules-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0;
}
`);

// Filter controls
block(`
.schedules-filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  align-items: center;
}
`);

block(`
.schedules-filter-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary, #666);
}
`);

block(`
.schedules-filter-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 4px;
  font-size: 0.875rem;
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  cursor: pointer;
}
`);

// Schedules table
block(`
.schedules-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
}
`);

block(`
.schedules-table th {
  text-align: left;
  padding: 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary, #666);
  border-bottom: 2px solid var(--border-color, #e0e0e0);
  background: var(--bg-secondary, #f8f9fa);
}
`);

block(`
.schedules-table td {
  padding: 1rem 0.75rem;
  font-size: 0.875rem;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  vertical-align: top;
}
`);

block(`
.schedules-table tr:hover {
  background: var(--bg-hover, #f8f9fa);
}
`);

// Schedule name cell
block(`
.schedule-name {
  font-weight: 500;
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 0.25rem;
}
`);

block(`
.schedule-description {
  font-size: 0.8125rem;
  color: var(--text-secondary, #666);
  margin: 0;
}
`);

// Schedule pattern text
block(`
.schedule-pattern {
  color: var(--text-primary, #1a1a1a);
  margin-bottom: 0.25rem;
}
`);

block(`
.schedule-time {
  font-size: 0.8125rem;
  color: var(--text-secondary, #666);
  margin: 0;
}
`);

// Status badges
block(`
.schedule-status {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
}
`);

block(`
.schedule-status.live {
  background: var(--success-bg, #d4edda);
  color: var(--success-text, #155724);
  border: 1px solid var(--success-border, #c3e6cb);
}
`);

block(`
.schedule-status.upcoming {
  background: var(--warning-bg, #fff3cd);
  color: var(--warning-text, #856404);
  border: 1px solid var(--warning-border, #ffeaa7);
}
`);

block(`
.schedule-status.idle {
  background: var(--bg-secondary, #f8f9fa);
  color: var(--text-secondary, #666);
  border: 1px solid var(--border-color, #e0e0e0);
}
`);

// Action buttons
block(`
.schedule-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
`);

block(`
.btn-schedule {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  border-radius: 4px;
  border: 1px solid var(--border-color, #e0e0e0);
  background: var(--surface, #fff);
  color: var(--text-primary, #1a1a1a);
  cursor: pointer;
  transition: all 0.2s ease;
}
`);

block(`
.btn-schedule:hover {
  background: var(--bg-hover, #f8f9fa);
  border-color: var(--primary, #007bff);
}
`);

block(`
.btn-schedule-delete {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  border-radius: 4px;
  border: 1px solid var(--danger, #dc3545);
  background: var(--surface, #fff);
  color: var(--danger, #dc3545);
  cursor: pointer;
  transition: all 0.2s ease;
}
`);

block(`
.btn-schedule-delete:hover {
  background: var(--danger, #dc3545);
  color: white;
}
`);

// Empty state
block(`
.schedules-empty {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-secondary, #666);
}
`);

block(`
.schedules-empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}
`);

block(`
.schedules-empty-text {
  font-size: 1rem;
  margin-bottom: 1rem;
}
`);

// Loading state
block(`
.schedules-loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary, #666);
}
`);

// Responsive design
block(`
@media (max-width: 768px) {
  .schedules-table {
    font-size: 0.8125rem;
  }

  .schedules-table th,
  .schedules-table td {
    padding: 0.5rem 0.375rem;
  }

  .schedule-actions {
    flex-direction: column;
    gap: 0.25rem;
  }

  .btn-schedule,
  .btn-schedule-delete {
    width: 100%;
    text-align: center;
  }
}
`);

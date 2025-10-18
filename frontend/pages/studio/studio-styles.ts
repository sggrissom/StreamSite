import { block } from "vlens/css";

// Main container
block(`
.studio-container {
  min-height: calc(100vh - 140px);
  padding: 2rem 1rem;
  background: var(--bg);
}
`);

block(`
.studio-content {
  max-width: 1200px;
  margin: 0 auto;
}
`);

// Breadcrumb
block(`
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}
`);

block(`
.breadcrumb a {
  color: var(--primary-accent);
  text-decoration: none;
}
`);

block(`
.breadcrumb a:hover {
  text-decoration: underline;
}
`);

block(`
.breadcrumb-separator {
  color: var(--muted);
}
`);

block(`
.breadcrumb-current {
  color: var(--text);
  font-weight: 500;
}
`);

// Studio Header
block(`
.studio-header {
  margin-bottom: 2rem;
}
`);

block(`
.studio-header-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.75rem;
}
`);

block(`
.studio-title {
  font-size: 2.25rem;
  font-weight: 700;
  color: var(--hero);
  margin: 0;
}
`);

block(`
.studio-badges {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}
`);

block(`
.active-badge {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  background: #fee;
  color: #c00;
  white-space: nowrap;
  animation: pulse 2s ease-in-out infinite;
}
`);

block(`
.studio-description {
  font-size: 1.1rem;
  color: var(--muted);
  margin: 0.75rem 0;
  line-height: 1.6;
}
`);

block(`
.studio-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
  font-size: 0.95rem;
  color: var(--muted);
}
`);

block(`
.studio-stats-text {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
`);

block(`
.stat-item {
  font-weight: 500;
}
`);

block(`
.stat-separator {
  color: var(--border);
}
`);

// Role badge (enhanced with visual hierarchy)
block(`
.studio-role {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  white-space: nowrap;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}
`);

block(`
.role-0 {
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #d1d5db;
}
`);

block(`
.role-0::before {
  content: '👁️';
  font-size: 1rem;
}
`);

block(`
.role-1 {
  background: #d1fae5;
  color: #059669;
  border: 1px solid #6ee7b7;
}
`);

block(`
.role-1::before {
  content: '👤';
  font-size: 1rem;
}
`);

block(`
.role-2 {
  background: #dbeafe;
  color: #2563eb;
  border: 1px solid #93c5fd;
}
`);

block(`
.role-2::before {
  content: '⚙️';
  font-size: 1rem;
}
`);

block(`
.role-3 {
  background: #fef3c7;
  color: #d97706;
  border: 1px solid #fcd34d;
}
`);

block(`
.role-3::before {
  content: '👑';
  font-size: 1rem;
}
`);

// Role Guide
block(`
.role-guide {
  margin-bottom: 1.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}
`);

block(`
.role-guide-toggle {
  width: 100%;
  padding: 0.875rem 1rem;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text);
  text-align: left;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background 0.2s ease;
}
`);

block(`
.role-guide-toggle:hover {
  background: var(--bg);
}
`);

block(`
.role-guide-content {
  padding: 1rem;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
`);

block(`
.role-capability {
  display: flex;
  align-items: center;
  gap: 1rem;
}
`);

block(`
.capability-desc {
  color: var(--muted);
  font-size: 0.95rem;
}
`);

// Studio Metadata Cards
block(`
.studio-metadata {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}
`);

block(`
.metadata-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  text-align: center;
}
`);

block(`
.metadata-label {
  font-size: 0.875rem;
  color: var(--muted);
  font-weight: 500;
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.metadata-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--hero);
}
`);

// Rooms Section
block(`
.rooms-section {
  margin-top: 3rem;
}
`);

block(`
.rooms-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
`);

block(`
.section-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--hero);
  margin: 0;
}
`);

// Rooms Empty State
block(`
.rooms-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 12px;
}
`);

block(`
.rooms-empty .empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}
`);

block(`
.rooms-empty h3 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.5rem 0;
}
`);

block(`
.rooms-empty p {
  font-size: 1rem;
  color: var(--muted);
  margin: 0 0 1.5rem 0;
  max-width: 500px;
}
`);

// Rooms Grid
block(`
.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}
`);

// Room Card
block(`
.room-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: all 0.2s ease;
}
`);

block(`
.room-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: var(--primary-accent);
}
`);

// Room Header
block(`
.room-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
`);

block(`
.room-number {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.room-status {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
}
`);

block(`
.room-status.active {
  background: #fee;
  color: #c00;
  animation: pulse 2s ease-in-out infinite;
}
`);

block(`
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}
`);

// Room Name
block(`
.room-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--hero);
  margin: 0;
  word-break: break-word;
}
`);

// Room Meta
block(`
.room-meta {
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
  color: var(--muted);
  font-weight: 500;
}
`);

block(`
.meta-value {
  color: var(--text);
  font-weight: 600;
}
`);

// Room Actions
block(`
.room-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
`);

// Rooms Limit Notice
block(`
.rooms-limit-notice {
  margin-top: 1.5rem;
  padding: 1rem 1.5rem;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 6px;
}
`);

block(`
.rooms-limit-notice p {
  margin: 0;
  color: #92400e;
  font-size: 0.95rem;
}
`);

// Members Section
block(`
.members-section {
  margin-top: 3rem;
}
`);

block(`
.members-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
`);

block(`
.members-empty {
  padding: 2rem;
  text-align: center;
  background: var(--surface);
  border: 2px dashed var(--border);
  border-radius: 12px;
  color: var(--muted);
}
`);

block(`
.members-table-wrapper {
  overflow-x: auto;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
}
`);

block(`
.members-table {
  width: 100%;
  border-collapse: collapse;
}
`);

block(`
.members-table thead {
  background: var(--bg);
  border-bottom: 2px solid var(--border);
}
`);

block(`
.members-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
`);

block(`
.members-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text);
}
`);

block(`
.members-table tbody tr:last-child td {
  border-bottom: none;
}
`);

block(`
.members-table tbody tr:hover {
  background: var(--bg);
}
`);

block(`
.table-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
`);

block(`
.table-owner-label {
  font-size: 0.9rem;
  color: var(--muted);
  font-style: italic;
}
`);

// Dropdown Component
block(`
.dropdown {
  position: relative;
  display: inline-block;
}
`);

block(`
.dropdown-trigger {
  cursor: pointer;
}
`);

block(`
.dropdown-menu {
  position: absolute;
  top: 100%;
  margin-top: 0.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  z-index: 100;
  overflow: hidden;
}
`);

block(`
.dropdown-menu-left {
  left: 0;
}
`);

block(`
.dropdown-menu-right {
  right: 0;
}
`);

block(`
.dropdown-item {
  width: 100%;
  padding: 0.75rem 1rem;
  text-align: left;
  border: none;
  background: transparent;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease;
  display: block;
}
`);

block(`
.dropdown-item:hover {
  background: var(--bg);
}
`);

block(`
.dropdown-item-danger {
  color: #dc2626;
}
`);

block(`
.dropdown-item-danger:hover {
  background: #fee;
}
`);

// Error State
block(`
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  margin: 2rem auto;
  max-width: 600px;
}
`);

block(`
.error-state .error-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}
`);

block(`
.error-state h2 {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text);
  margin: 0 0 0.75rem 0;
}
`);

block(`
.error-state p {
  font-size: 1rem;
  color: var(--muted);
  margin: 0 0 2rem 0;
  max-width: 500px;
}
`);

// Modal Styles
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

block(`
.modal-body {
  padding: 1.5rem;
}
`);

block(`
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid var(--border);
}
`);

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

// Stream Key Styles
block(`
.stream-key-room-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--hero);
  padding: 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
}
`);

block(`
.stream-key-display {
  font-family: 'Courier New', monospace;
  font-size: 0.95rem;
  padding: 1rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  word-break: break-all;
  color: var(--text);
  user-select: all;
}
`);

block(`
.stream-key-loading {
  text-align: center;
  padding: 2rem;
  color: var(--muted);
  font-size: 1rem;
}
`);

block(`
.stream-key-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
`);

block(`
.confirmation-dialog {
  margin-top: 1.5rem;
  padding: 1rem;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 6px;
}
`);

block(`
.confirmation-text {
  margin: 0 0 1rem 0;
  color: #92400e;
  font-size: 0.95rem;
  line-height: 1.5;
}
`);

block(`
.confirmation-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}
`);

// Delete Warning/Confirmation Styles
block(`
.delete-warning {
  padding: 1rem;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 6px;
}
`);

block(`
.delete-confirmation {
  padding: 1rem;
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 6px;
}
`);

block(`
.warning-text {
  margin: 0 0 1rem 0;
  color: #92400e;
  font-size: 0.95rem;
  line-height: 1.5;
}
`);

block(`
.delete-confirmation .confirmation-text {
  margin: 0 0 1rem 0;
  color: #c00;
  font-size: 0.95rem;
  line-height: 1.5;
}
`);

block(`
.room-info {
  padding: 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.95rem;
}
`);

block(`
.room-info strong {
  color: var(--muted);
  margin-right: 0.5rem;
}
`);

// Responsive Design
block(`
@media (max-width: 768px) {
  .studio-container {
    padding: 1.5rem 1rem;
  }

  .studio-title {
    font-size: 1.75rem;
  }

  .studio-header-main {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .studio-badges {
    width: 100%;
    justify-content: flex-start;
  }

  .studio-stats {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .rooms-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .rooms-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .section-title {
    font-size: 1.5rem;
  }

  .members-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .members-table-wrapper {
    overflow-x: auto;
  }

  .members-table th,
  .members-table td {
    padding: 0.75rem;
    font-size: 0.85rem;
  }

  .table-actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .modal-content {
    max-width: 95%;
  }

  .modal-footer {
    flex-direction: column-reverse;
  }

  .modal-footer .btn {
    width: 100%;
  }

  .stream-key-actions {
    flex-direction: column;
  }

  .stream-key-actions .btn {
    width: 100%;
  }

  .confirmation-actions {
    flex-direction: column-reverse;
  }

  .confirmation-actions .btn {
    width: 100%;
  }
}
`);

block(`
@media (max-width: 480px) {
  .studio-container {
    padding: 1rem 0.75rem;
  }

  .studio-title {
    font-size: 1.5rem;
  }

  .breadcrumb {
    font-size: 0.8rem;
  }

  .room-card {
    padding: 1.25rem;
  }

  .room-name {
    font-size: 1.1rem;
  }

  .btn {
    width: 100%;
  }

  .room-actions {
    flex-direction: column;
  }

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

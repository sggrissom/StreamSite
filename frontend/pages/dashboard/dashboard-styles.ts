import { block } from "vlens/css";

block(`
.dashboard-container {
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}
`);

block(`
.dashboard-content {
  text-align: center;
  max-width: 600px;
}
`);

block(`
.dashboard-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 1rem;
}
`);

block(`
.dashboard-description {
  font-size: 1.1rem;
  color: var(--muted);
  margin: 0 0 2rem;
  line-height: 1.6;
}
`);

block(`
.dashboard-actions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}
`);

block(`
.dashboard-actions .btn {
  width: 100%;
  max-width: 280px;
}
`);

block(`
@media (max-width: 768px) {
  .dashboard-container {
    padding: 20px 16px;
  }

  .dashboard-title {
    font-size: 2rem;
  }

  .dashboard-description {
    font-size: 1rem;
  }
}
`);

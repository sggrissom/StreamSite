import { block } from "vlens/css";

block(`
.home-container {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
`);

block(`
.home-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 1rem;
}
`);

block(`
.home-description {
  font-size: 1.1rem;
  color: var(--muted);
  margin: 0;
}
`);

block(`
.stream-link-container {
  margin-top: 2rem;
}
`);

block(`
.stream-link {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background-color: var(--accent);
  color: white;
  text-decoration: none;
  border-radius: 8px;
  font-weight: bold;
  transition: all 0.2s ease;
}
`);

block(`
.stream-link:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  text-decoration: none;
}
`);

block(`
@media (max-width: 768px) {
  .home-container {
    padding: 1rem;
  }

  .home-title {
    font-size: 1.5rem;
  }

  .home-description {
    font-size: 1rem;
  }
}
`);

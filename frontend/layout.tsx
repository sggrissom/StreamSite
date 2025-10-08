import * as preact from "preact";

export const Header = () => {
  return (
    <header
      style={{
        padding: "1rem",
        background: "var(--surface)",
        borderBottom: "1px solid var(--border)",
      }}
    >
      <nav>
        <a
          href="/"
          style={{
            fontSize: "1.25rem",
            fontWeight: "bold",
            color: "var(--text)",
            textDecoration: "none",
          }}
        >
          Stream
        </a>
      </nav>
    </header>
  );
};

export const Footer = () => (
  <footer
    style={{ padding: "1rem", textAlign: "center", color: "var(--muted)" }}
  >
    <p>© 2025 Stream. All rights reserved.</p>
  </footer>
);

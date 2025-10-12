import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "./server";
import "./layout-styles";

const useAuthCheck = vlens.declareHook(() => {
  const state = {
    auth: null as server.AuthResponse | null,
    isAuthenticated: false,
    isLoading: true,
  };

  // Check auth on mount
  server.GetAuthContext({}).then(([authResp, authErr]) => {
    if (authResp && authResp.id > 0) {
      state.auth = authResp;
      state.isAuthenticated = true;
    }
    state.isLoading = false;
    vlens.scheduleRedraw();
  });

  return state;
});

async function handleLogout(event: Event) {
  event.preventDefault();

  const nativeFetch = window.fetch.bind(window);
  try {
    const res = await nativeFetch("/api/logout", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const result = await res.json();
    if (result.success) {
      // Redirect to home page after logout
      window.location.href = "/";
    }
  } catch (error) {
    console.error("Logout failed:", error);
    // Still redirect on error
    window.location.href = "/";
  }
}

export const Header = () => {
  const authState = useAuthCheck();

  return (
    <header className="site-header">
      <nav className="site-nav">
        <a href="/" className="site-logo">
          Stream
        </a>
        {authState.isAuthenticated && authState.auth && (
          <div className="nav-links">
            <a href="/dashboard" className="nav-link">
              Dashboard
            </a>
            <a href="/studios" className="nav-link">
              Studios
            </a>
            {authState.auth.isStreamAdmin && (
              <a href="/stream-admin" className="nav-link">
                Stream Admin
              </a>
            )}
            {authState.auth.isSiteAdmin && (
              <a href="/site-admin" className="nav-link">
                Site Admin
              </a>
            )}
            <button onClick={handleLogout} className="logout-button">
              Logout
            </button>
          </div>
        )}
      </nav>
    </header>
  );
};

export const Footer = () => (
  <footer className="site-footer">
    <p>© 2025 Stream. All rights reserved.</p>
  </footer>
);

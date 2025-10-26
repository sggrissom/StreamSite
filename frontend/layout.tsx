import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "./server";
import "./layout-styles";

type AuthState = {
  auth: server.AuthResponse | null;
  isAuthenticated: boolean;
  isLoading: boolean;
};

const useAuthCheck = vlens.declareHook((): AuthState => {
  const state: AuthState = {
    auth: null,
    isAuthenticated: false,
    isLoading: true,
  };

  // Check auth on mount
  server.GetAuthContext({}).then(([authResp, authErr]) => {
    if (authResp) {
      state.auth = authResp;
      // Only set isAuthenticated for real users (id > 0)
      if (authResp.id > 0) {
        state.isAuthenticated = true;
      }
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

// Helper function to determine which header type to render
function getHeaderType(authState: AuthState): string {
  if (authState.isLoading) return "guest";
  if (!authState.auth || authState.auth.id === 0) return "guest";
  if (authState.auth.id === -1) return "anonymous-code";
  if (authState.auth.isSiteAdmin) return "site-admin";
  if (authState.auth.canManageStudios) return "studio-manager";
  return "viewer";
}

// Header navigation components for each auth type

const GuestHeaderNav = (): preact.ComponentChild => null;

const AnonymousCodeHeaderNav = (): preact.ComponentChild => (
  <div className="auth-links">
    <a href="/login" className="auth-link">
      Sign in
    </a>
    <span className="auth-separator">|</span>
    <a href="/create-account" className="auth-link">
      Create account
    </a>
  </div>
);

const ViewerHeaderNav = (): preact.ComponentChild => (
  <div className="nav-links">
    <a href="/dashboard" className="nav-link">
      Streams
    </a>
    <button onClick={handleLogout} className="logout-button">
      Logout
    </button>
  </div>
);

const StudioManagerHeaderNav = (): preact.ComponentChild => (
  <div className="nav-links">
    <a href="/studios" className="nav-link">
      Studios
    </a>
    <button onClick={handleLogout} className="logout-button">
      Logout
    </button>
  </div>
);

const SiteAdminHeaderNav = (): preact.ComponentChild => (
  <div className="nav-links">
    <a href="/studios" className="nav-link">
      Studios
    </a>
    <a href="/site-admin" className="nav-link">
      Admin
    </a>
    <button onClick={handleLogout} className="logout-button">
      Logout
    </button>
  </div>
);

// Component map for dynamic header rendering
const HEADER_COMPONENTS: Record<string, () => preact.ComponentChild> = {
  guest: GuestHeaderNav,
  "anonymous-code": AnonymousCodeHeaderNav,
  viewer: ViewerHeaderNav,
  "studio-manager": StudioManagerHeaderNav,
  "site-admin": SiteAdminHeaderNav,
};

export const Header = () => {
  const authState = useAuthCheck();
  const headerType = getHeaderType(authState);
  const HeaderNavComponent = HEADER_COMPONENTS[headerType];

  return (
    <header className="site-header">
      <nav className="site-nav">
        <a href="/" className="site-logo">
          Stream
        </a>
        <HeaderNavComponent />
      </nav>
    </header>
  );
};

export const Footer = () => (
  <footer className="site-footer">
    <p>© 2025 Stream. All rights reserved.</p>
  </footer>
);
